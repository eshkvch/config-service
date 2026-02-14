package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Connection interface {
	GetDB() *sql.DB
	Close() error
}

type postgresConnection struct {
	db *sql.DB
}

func NewPostgresConnection(dsn string) (Connection, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &postgresConnection{db: db}, nil
}

func (c *postgresConnection) GetDB() *sql.DB {
	return c.db
}

func (c *postgresConnection) Close() error {
	return c.db.Close()
}
