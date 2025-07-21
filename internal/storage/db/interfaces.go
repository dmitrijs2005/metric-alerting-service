package db

import (
	"context"
	"database/sql"
)

// DBExecutor abstracts the sql.DB and sql.Tx types for executing queries.
//
// This interface allows writing reusable code that works with either a database connection
// or a transaction, making it easier to manage transactional boundaries and testing.
//
// Both *sql.DB and *sql.Tx implement this interface.
type DBExecutor interface {
	// ExecContext executes a query without returning any rows.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// QueryRowContext executes a query that is expected to return at most one row.
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
