package storage

import (
	"fmt"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
)

type MemStorage struct {
	Data map[string]interface{}
}

func getKey(metricType string, metricName string) string {
	return fmt.Sprintf("%s|%s", metricType, metricName)
}

func (s *MemStorage) updateGauge(m *metrics.Metric) error {
	key := getKey(m.Type, m.Name)

	float64Val, ok := m.Value.(float64)

	if !ok {
		return fmt.Errorf("error converting value (float64)")
	}

	s.Data[key] = float64Val
	return nil
}

func (s *MemStorage) updateCounter(m *metrics.Metric) error {

	oldVal, err := s.Retrieve(m.Type, m.Name)

	if err != nil {
		return fmt.Errorf("error retrieving data: %s", err.Error())
	}

	incrementVal, ok := m.Value.(int64)

	if !ok {
		return fmt.Errorf("error converting increment value (int64)")
	}

	var newValue int64

	if oldVal == nil {
		newValue = incrementVal
	} else {
		oldInt64Val, ok := oldVal.(int64)

		if !ok {
			return fmt.Errorf("error converting old value (int64)")
		}
		newValue = oldInt64Val + incrementVal
	}

	key := getKey(m.Type, m.Name)
	s.Data[key] = newValue
	return nil
}

func (s *MemStorage) Update(m *metrics.Metric) error {

	if m.Type == metrics.MetricTypeGauge {
		if err := s.updateGauge(m); err != nil {
			return err
		}
	} else if m.Type == metrics.MetricTypeCounter {

		if err := s.updateCounter(m); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("invalid metric type")
	}

	return nil
}

func (s *MemStorage) Retrieve(metricType string, metricName string) (interface{}, error) {
	key := getKey(metricType, metricName)
	return s.Data[key], nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Data: make(map[string]interface{})}
}
