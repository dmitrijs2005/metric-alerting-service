package storage

import (
	"context"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

// Storage defines a generic interface for storing and managing metrics.
//
// Implementations may store metrics in-memory, in a file, or in a database.
// This interface abstracts metric operations such as add, update, retrieve, and batch update.
type Storage interface {
	// Add inserts a new metric into the storage.
	Add(ctx context.Context, m metric.Metric) error

	// Update modifies the value of an existing metric.
	Update(ctx context.Context, m metric.Metric, v interface{}) error

	// Retrieve fetches a single metric by type and name.
	Retrieve(ctx context.Context, m metric.MetricType, n string) (metric.Metric, error)

	// RetrieveAll returns all stored metrics.
	RetrieveAll(ctx context.Context) ([]metric.Metric, error)

	// UpdateBatch updates or inserts multiple metrics atomically if supported.
	UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error
}

// DBStorage extends the Storage interface with database-specific functionality.
//
// It is used by database-backed implementations (e.g., PostgreSQL) to support
// lifecycle operations like connection health checks and running schema migrations.
type DBStorage interface {
	Storage

	// Close gracefully closes the underlying storage connection.
	Close() error

	// RunMigrations applies schema changes to the database.
	RunMigrations(ctx context.Context) error

	// Ping checks if the storage backend is reachable.
	Ping(ctx context.Context) error
}
