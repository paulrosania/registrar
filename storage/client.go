package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Client struct {
	db Database

	Apps   ApplicationsService
	Tokens TokensService
	Users  UsersService
}

func NewClient(cfg *Config) (*Client, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s connect_timeout=10", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, cfg.Database.Sslmode)
	pg, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed opening Postgres connection:", err)
	}

	pg.SetMaxOpenConns(cfg.Database.Pool) // 0 = unlimited

	// Validate DSN data
	err = pg.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed opening Postgres connection:", err)
	}

	db := NewDatabase(pg)

	c := &Client{db: db}
	c.Apps = &LocalApplicationsService{c}
	c.Tokens = &LocalTokensService{c}
	c.Users = &LocalUsersService{c}

	return c, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}
