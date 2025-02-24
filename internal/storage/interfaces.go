package storage

import "github.com/dmitrijs2005/metric-alerting-service/internal/metric"

const (
	MetricDoesNotExist  = "metric does not exist"
	MetricAlreadyExists = "metric already exists"
)

type Storage interface {
	Add(m metric.Metric) error
	Update(m metric.Metric, v interface{}) error
	Retrieve(m metric.MetricType, n string) (metric.Metric, error)
	RetrieveAll() ([]metric.Metric, error)
	SaveDump() error
	RestoreDump() error
}

type DumpSaver interface {
	SaveDump(s Storage) error
	RestoreDump(s Storage) error
}
