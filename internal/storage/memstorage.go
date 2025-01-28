package storage

import (
	"errors"
	"fmt"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
)

type MemStorage struct {
	Data map[string]metrics.Metric
}

func getKey(metricType metrics.MetricType, metricName string) string {
	return fmt.Sprintf("%s|%s", metricType, metricName)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Data: make(map[string]metrics.Metric)}
}

func (s *MemStorage) Retrieve(metricType metrics.MetricType, metricName string) (metrics.Metric, error) {
	key := getKey(metricType, metricName)
	if value, exists := s.Data[key]; exists {
		return value, nil
	} else {
		return nil, errors.New(MetricDoesNotExist)
	}
}

func (s *MemStorage) RetrieveAll() ([]metrics.Metric, error) {
	result := []metrics.Metric{}
	for _, metric := range s.Data {
		result = append(result, metric)
	}
	return result, nil
}

func (s *MemStorage) Add(metric metrics.Metric) error {
	key := getKey(metric.GetType(), metric.GetName())
	_, exists := s.Data[key]
	if exists {
		return errors.New(MetricAlreadyExists)
	}
	s.Data[key] = metric
	return nil
}

func (s *MemStorage) Update(metric metrics.Metric, value interface{}) error {
	return metric.Update(value)
}
