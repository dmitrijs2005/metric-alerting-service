package storage

import "github.com/dmitrijs2005/metric-alerting-service/internal/metrics"

const (
	MetricDoesNotExist  = "metric does not exist"
	MetricAlreadyExists = "metric already exists"
)

type Storage interface {
	Add(m metrics.Metric) error
	Update(m metrics.Metric, v interface{}) error
	Retrieve(m metrics.MetricType, n string) (metrics.Metric, error)
	RetrieveAll() ([]metrics.Metric, error)
}
