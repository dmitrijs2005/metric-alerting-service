// Package storage defines abstract interfaces for working with metric storage backends.
//
// The Storage interface provides basic CRUD operations for metrics,
// and is implemented by various backends such as in-memory storage or PostgreSQL.
//
// The DBStorage interface extends Storage with database-specific lifecycle methods,
// such as Ping and RunMigrations, useful for health checks and migrations in SQL-based systems.
package storage
