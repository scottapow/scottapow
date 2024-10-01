package db

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func Connect() (*Store, error) {
	connStr := os.Getenv("DB_CONN_STR")
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}
	defer db.Close()

	return &Store{db: db}, nil
}

func (s *Store) Setup() error {
	ctx, cancel := context.WithCancel(context.Background())
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		cancel()
		return err
	}

	_, err = tx.Query(`
		CREATE TABLE IF NOT EXISTS users(
			id                  SERIAL PRIMARY KEY,
			username            VARCHAR(255) UNIQUE NOT NULL,
			email               VARCHAR(255) UNIQUE NOT NULL,
			password            VARCHAR(255),
			created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			oauth_provider      VARCHAR(255),
			oauth_provider_id   VARCHAR(255),
			oauth_access_token  VARCHAR(255),
			oauth_refresh_token VARCHAR(255),
			oauth_expires_at    TIMESTAMP
		)
	`)

	if err != nil {
		cancel()
		return err
	}
	cancel()
	return nil
}
