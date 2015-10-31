package main

import (
	"github.com/paulrosania/registrar/storage"
)

type Config struct {
	Server struct {
		Bind    string
		BaseUrl string `toml:"base-url"`
	}

	OpenID struct {
		Issuer string `toml:"open-id"`
	}

	JWT struct {
		PublicKey  string `toml:"public-key"`  // openssl genrsa -out registrar.rsa <key-size>
		PrivateKey string `toml:"private-key"` // openssl rsa -in registrar.rsa -pubout > registrar.rsa.pub
	}

	Log struct {
		Path string
	}

	Database storage.DatabaseConfig
}
