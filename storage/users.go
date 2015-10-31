package storage

import (
	"errors"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type UserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UsersService interface {
	FindByCredentials(email, password string) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByToken(token string) (*User, error)
	FindByRefreshToken(clientId int64, token string) (*User, error)
	New(params *UserParams) (*User, error)
	Authorize(userId, clientId int64, scope string, refresh bool) (string, error)
}

type LocalUsersService struct {
	client *Client
}

const DefaultPasswordCost = 12

const findUserByTokenSql = `SELECT u.id, u.email FROM users u
JOIN oauth_tokens t ON t.user_id = u.id
WHERE t.token = $1 AND t.type = 'access_token' AND t.expires_at > NOW() GROUP BY u.id LIMIT 1`

func (s *LocalUsersService) FindByToken(token string) (u *User, err error) {
	var id int64
	var email string
	err = s.client.db.QueryRow(findUserByTokenSql, token).Scan(&id, &email)
	if err != nil {
		log.Println("Users.FindByToken:", err)
		return
	}

	return &User{id, email}, nil
}

const findUserByRefreshTokenSql = `SELECT u.id, u.email FROM users u
JOIN oauth_tokens t ON t.user_id = u.id
WHERE t.client_id = $1 AND t.token = $2 AND t.type = 'refresh_token'
  AND (t.expires_at IS NULL OR t.expires_at > NOW())
GROUP BY u.id LIMIT 1`

func (s *LocalUsersService) FindByRefreshToken(clientId int64, token string) (u *User, err error) {
	var id int64
	var email string
	err = s.client.db.QueryRow(findUserByRefreshTokenSql, clientId, token).Scan(&id, &email)
	if err != nil {
		log.Println("db.FindUserByRefreshToken:", err)
		return
	}

	return &User{id, email}, nil
}

const findUserByEmailSql = `SELECT u.id FROM users u WHERE u.email = $1 LIMIT 1`

func (s *LocalUsersService) FindByEmail(email string) (u *User, err error) {
	var id int64
	err = s.client.db.QueryRow(findUserByEmailSql, email).Scan(&id)
	if err != nil {
		log.Println("db.FindUserByEmail:", err)
		return
	}

	return &User{id, email}, nil
}

const findUserByEmailWithPasswordSql = `SELECT u.id, u.password FROM users u
WHERE u.email = $1 LIMIT 1`

func (s *LocalUsersService) FindByCredentials(email, password string) (u *User, err error) {
	var id int64
	var hashed string
	err = s.client.db.QueryRow(findUserByEmailWithPasswordSql, email).Scan(&id, &hashed)
	if err != nil {
		log.Println("db.FindUserByCredentials:", err)
		return
	}

	// NOTE: Timing attacks will reveal existence of users, since we only hash
	// when we find a real user record in the database.
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		log.Println("db.FindUserByCredentials:", err)
		return
	}

	return &User{id, email}, nil
}

const createUserSql = `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`

var (
	ErrNotUnique = errors.New("duplicate key value violates unique key constraint")
)

func translateError(err error) error {
	switch err := err.(type) {
	case nil:
		return nil
	case *pq.Error:
		switch err.Code {
		case "23505": // unique_violation
			return ErrNotUnique
		default:
			log.Printf("returning raw PQ error (code %s)", err.Code)
			return err
		}
	default:
		return err
	}
}

func (s *LocalUsersService) New(params *UserParams) (*User, error) {
	crypted, err := bcrypt.GenerateFromPassword([]byte(params.Password), DefaultPasswordCost)
	if err != nil {
		return nil, translateError(err)
	}

	var id int64
	err = s.client.db.QueryRow(createUserSql, params.Email, string(crypted)).Scan(&id)
	if err != nil {
		return nil, translateError(err)
	}

	return &User{id, params.Email}, nil
}

const createTokenSql = `INSERT INTO oauth_tokens
(client_id, user_id, type, token, expires_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id`
const attachScopesSql = `INSERT INTO authorized_scopes (oauth_token_id, scope_id)
SELECT ?, s.id FROM scopes s INNER JOIN permitted_scopes ps ON ps.scope_id = s.id
WHERE s.name IN (?) AND ps.client_id = ?`

func (s *LocalUsersService) Authorize(userId, clientId int64, scope string, refresh bool) (string, error) {
	if !refresh {
		return "", nil
	}

	tx, err := s.client.db.Begin()
	if err != nil {
		log.Println("User.Authorize:", err)
		return "", err
	}

	refreshToken, err := RandomToken()
	if err != nil {
		log.Println("User.Authorize: failed generating token:", err)
		return "", err
	}

	var tokenId int
	err = tx.QueryRow(createTokenSql, clientId, userId, "refresh_token", refreshToken, nil).Scan(&tokenId)
	if err != nil {
		tx.Rollback()
		log.Println("User.Authorize: failed inserting token:", err)
		return "", err
	}

	scopes := strings.Split(scope, " ")
	query, args, err := sqlx.In(attachScopesSql, tokenId, scopes, clientId)
	query = s.client.db.Rebind(query)
	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		log.Println("User.Authorize: failed attaching scopes:", err)
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		log.Println("User.Authorize: failed committing transaction:", err)
		return "", err
	}

	return refreshToken, nil
}
