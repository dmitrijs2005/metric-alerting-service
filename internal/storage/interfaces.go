package storage

import "github.com/dmitrijs2005/metric-alerting-service/internal/metrics"

type Storage interface {
	Update(*metrics.Metric) error
	Retrieve(metricType string, metricName string) (interface{}, error)
}
