package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

func NewJWT(issuer, clientId, email string, seconds int) *jwt.Token {
	token := jwt.New(jwt.SigningMethodRS256)
	now := time.Now()
	token.Claims["aud"] = clientId
	token.Claims["exp"] = now.Add(time.Second * time.Duration(seconds)).Unix()
	token.Claims["iat"] = now.Unix()
	token.Claims["iss"] = issuer
	token.Claims["sub"] = email

	return token
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	State        string `json:"-"`
}
