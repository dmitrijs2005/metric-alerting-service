package db

import (
	"context"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/mock"
)

type MockDBClient struct {
	mock.Mock
}

func (c *MockDBClient) Ping(ctx context.Context) error {
	return nil
}

func (c *MockDBClient) Add(ctx context.Context, m metric.Metric) error {
	return nil
}

func (c *MockDBClient) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return nil
}

func (c *MockDBClient) Retrieve(ctx context.Context, m metric.MetricType, n string) (metric.Metric, error) {
	return nil, nil
}

func (c *MockDBClient) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, nil
}

func NewMockDBClient() *MockDBClient {
	return &MockDBClient{}
}

func (c *MockDBClient) Close() error {
	return nil
}

func (c *MockDBClient) RunMigrations(ctx context.Context) error {
	return nil
}

func (s *MockDBClient) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return nil
}
