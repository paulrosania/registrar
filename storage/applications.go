package storage

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

const DefaultClientSecretCost = 12

type Application struct {
	Id                 int64  `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Website            string `json:"website"`
	Logo               string `json:"logo"`
	ClientType         string `json:"client_type"`
	ClientId           string `json:"client_id"`
	ClientSecret       string `json:"client_secret,omitempty"`
	HashedClientSecret string `json:"-"`
}

type ApplicationParams struct {
}

type ApplicationsService interface {
	Authorize(a *Application, scope string) (*Token, error)
	ExchangeAuthCode(a *Application, code, redirectUri string) (*Token, error)
	FindByCredentials(email, password string) (*Application, error)
	New(name, description, website, logo, clientType string) (*Application, error)
}

type LocalApplicationsService struct {
	client *Client
}

const defaultApplicationFields = `a.id, a.name, a.description, a.website, a.logo,
a.client_type, a.client_id, a.client_secret as hashed_client_secret`

const findApplicationByClientIdSql = "SELECT " + defaultApplicationFields + ` FROM applications a
WHERE a.client_id = $1 GROUP BY a.id LIMIT 1`

func (s *LocalApplicationsService) FindByClientId(id string) (*Application, error) {
	a := &Application{}
	err := s.client.db.Get(a, findApplicationByClientIdSql, id)
	if err != nil {
		log.Println("Apps.FindByClientId:", err)
		return nil, err
	}

	return a, nil
}

func (s *LocalApplicationsService) FindByCredentials(id, secret string) (*Application, error) {
	a, err := s.FindByClientId(id)
	if err != nil {
		log.Println("Apps.FindByCredentials:", err)
		return nil, err
	}

	// NOTE: Timing attacks will reveal existence of applications, since we only hash
	// when we find a real client record in the database.
	err = bcrypt.CompareHashAndPassword([]byte(a.HashedClientSecret), []byte(secret))
	if err != nil {
		log.Println("Apps.FindByCredentials:", err)
		return nil, err
	}

	return a, nil
}

const createApplicationSql = `INSERT INTO applications
(name, description, website, logo, client_type, client_id, client_secret)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

func (s *LocalApplicationsService) New(name, description, website, logo, clientType string) (*Application, error) {
	a := &Application{Name: name, Description: description, Website: website, Logo: logo, ClientType: clientType}

	var err error
	a.ClientId, err = RandomToken()
	if err != nil {
		return nil, err
	}

	a.ClientSecret, err = RandomToken()
	if err != nil {
		return nil, err
	}

	crypted, err := bcrypt.GenerateFromPassword([]byte(a.ClientSecret), DefaultClientSecretCost)
	if err != nil {
		return nil, err
	}
	a.HashedClientSecret = string(crypted)

	var id int64
	err = s.client.db.QueryRow(createApplicationSql, a.Name, a.Description, a.Website, a.Logo, a.ClientType, a.ClientId, a.HashedClientSecret).Scan(&id)
	if err != nil {
		return nil, err
	}
	a.Id = id

	return a, nil
}

func (s *LocalApplicationsService) ExchangeAuthCode(a *Application, code, redirectUri string) (*Token, error) {
	// Verify we found a record

	// In a transaction:
	// 1. Create an auth token
	// 2. Consume the auth code (delete/mark as used)

	// Respond with access token
	return nil, nil
}

func (s *LocalApplicationsService) Authorize(a *Application, scope string) (*Token, error) {
	// Assume client is already authorized, and generate an access token
	return nil, nil
}
