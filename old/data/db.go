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
