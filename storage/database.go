package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func init() {
	sqlx.NameMapper = toSnake
}

type Database interface {
	Close() error
	Begin() (Tx, error)

	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
}

type SqlxDatabase struct {
	*sqlx.DB
}

func NewDatabase(db *sqlx.DB) Database {
	return &SqlxDatabase{db}
}

func (db *SqlxDatabase) Begin() (Tx, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}

	return &SqlxTx{tx}, nil
}

type Tx interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	Select(dest interface{}, query string, args ...interface{}) error
	Rollback() error
}

type SqlxTx struct {
	*sqlx.Tx
}
