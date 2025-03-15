package db

import (
	"context"
	"database/sql"
)

// DBExecutor abstracts sql.DB and sql.Tx for executing queries
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
