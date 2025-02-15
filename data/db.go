package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

func Connect() (*Store, error) {
	connStr := os.Getenv("DB_CONN_STR")
	bkg := context.Background()
	db, err := pgxpool.New(bkg, connStr)

	if err != nil {
		return nil, err
	}
	// defer db.Close()

	return &Store{DB: db}, nil
}

func (s *Store) Setup() error {
	bkg := context.Background()

	// types
	_, err := s.DB.Exec(bkg, `
		CREATE TYPE permission AS ENUM ('calories_read', 'calories_write');
	`)

	_, err = s.DB.Exec(bkg, `
		CREATE TABLE IF NOT EXISTS users(
			id                  UUID PRIMARY KEY,
			email               VARCHAR(255) UNIQUE NOT NULL,
			password            VARCHAR(255),
			created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at          TIMESTAMP,
			login_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			firstname           VARCHAR(255),
			surname             VARCHAR(255),
			avatar_url          VARCHAR(255),
			oauth_provider      VARCHAR(255),
			oauth_provider_id   VARCHAR(255),
			oauth_access_token  VARCHAR(255),
			oauth_refresh_token VARCHAR(255),
			oauth_expires_at    TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS permissions(
			id          UUID PRIMARY KEY,
            user_id     UUID REFERENCES users(id),
            permission  permission NOT NULL,
            created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS calories(
			id          UUID PRIMARY KEY,
			user_id     UUID REFERENCES users(id),
			created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at  TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS calorie_entries(
			id           UUID PRIMARY KEY,
			calories_id  UUID REFERENCES calories(id),
			amount       SMALLINT,
			consumed_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err != nil {
		return err
	}

	return nil
}
