package storage

import (
	"io"
	"time"

	"crypto/rand"
)

const maxTokenSize = 32

func RandomToken() (string, error) {
	rawToken := make([]byte, maxTokenSize)
	_, err := io.ReadFull(rand.Reader, rawToken)
	if err != nil {
		return "", err
	}

	return hexEncode(string(rawToken)), nil
}

type Token struct {
	Id        int
	ClientId  int
	UserId    int
	Type      string
	Token     string
	ExpiresAt time.Time
}

type TokenParams struct {
}

type TokensService interface {
	FindById(userId, labelId int64) (*Token, error)
	New(userId int64, params *TokenParams) (*Token, error)
}

type LocalTokensService struct {
	client *Client
}

const (
	labelsQuery                   = `SELECT * FROM labels WHERE user_id = $1`
	findLabelByUserIdAndUidQuery  = `SELECT * FROM labels WHERE user_id = $1 AND uid = $2`
	findLabelByUserIdAndNameQuery = `SELECT * FROM labels WHERE user_id = $1 AND name = $2`
)

func (s *LocalTokensService) FindById(userId, labelId int64) (*Token, error) {
	tx, err := s.client.db.Begin()
	if err != nil {
		return nil, err
	}

	t := new(Token)
	err = tx.Get(t, findLabelByUserIdAndUidQuery, userId, labelId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return t, nil
}

func (s *LocalTokensService) New(userId int64, params *TokenParams) (*Token, error) {
	return nil, nil
}
