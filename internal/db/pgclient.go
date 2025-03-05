package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresClient struct {
	db *sql.DB
}

func NewPostgresClient(dsn string) (*PostgresClient, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{db}, nil
}

func (c *PostgresClient) Close() error {
	return c.db.Close()
}

func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
