// Package memory provides an in-memory implementation of the Storage interface
// for storing and retrieving metric values without persistent storage.
package memory

import (
	"fmt"
	"sync"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"golang.org/x/net/context"
)

type MemStorage struct {
	Data map[string]metric.Metric
	mu   sync.Mutex
}

func getKey(metricType metric.MetricType, metricName string) string {
	return fmt.Sprintf("%s|%s", metricType, metricName)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Data: make(map[string]metric.Metric)}
}

func (s *MemStorage) Retrieve(ctx context.Context, metricType metric.MetricType, metricName string) (metric.Metric, error) {
	key := getKey(metricType, metricName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if value, exists := s.Data[key]; exists {
		return value, nil
	} else {
		return nil, common.ErrorMetricDoesNotExist
	}
}

func (s *MemStorage) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {

	result := []metric.Metric{}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range s.Data {
		result = append(result, metric)
	}

	return result, nil
}

func (s *MemStorage) Add(ctx context.Context, metric metric.Metric) error {
	key := getKey(metric.GetType(), metric.GetName())

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.Data[key]
	if exists {
		return common.ErrorMetricAlreadyExists
	}
	s.Data[key] = metric
	return nil
}

func (s *MemStorage) Update(ctx context.Context, metric metric.Metric, value interface{}) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getKey(metric.GetType(), metric.GetName())
	m, exists := s.Data[key]
	if exists {
		return m.Update(value)
	}
	return common.ErrorMetricDoesNotExist

}

func (s *MemStorage) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range *metrics {
		key := getKey(metric.GetType(), metric.GetName())
		m, exists := s.Data[key]
		if exists {
			err := m.Update(metric.GetValue())
			if err != nil {
				return fmt.Errorf("error updating %s", metric.GetName())
			}
		}
	}

	return nil

}
