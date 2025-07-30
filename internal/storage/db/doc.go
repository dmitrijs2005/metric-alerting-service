// Package db defines interfaces and abstractions for database access.
//
// It includes the DBExecutor interface, which abstracts over *sql.DB and *sql.Tx,
// allowing for writing generic database code that can operate within or outside transactions.
//
// This abstraction is useful for dependency injection and testability of database-related logic.
// It defines the PostgresClient type which implements methods for persisting,
// retrieving, and updating metrics using a relational database.
// This package also includes database migration support via goose,
// and provides abstractions for executing queries within or outside transactions.
//
// Example usage:
//
//	func saveUser(ctx context.Context, db db.DBExecutor, name string) error {
//	    _, err := db.ExecContext(ctx, "INSERT INTO users(name) VALUES(?)", name)
//	    return err
//	}
package db
