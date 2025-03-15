package storage

import (
	"context"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

type Storage interface {
	Add(ctx context.Context, m metric.Metric) error
	Update(ctx context.Context, m metric.Metric, v interface{}) error
	Retrieve(ctx context.Context, m metric.MetricType, n string) (metric.Metric, error)
	RetrieveAll(ctx context.Context) ([]metric.Metric, error)
	UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error
}

type DBStorage interface {
	Storage
	Close() error
	RunMigrations(ctx context.Context) error
	Ping(ctx context.Context) error
}
