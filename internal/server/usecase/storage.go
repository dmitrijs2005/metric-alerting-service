package usecase

import (
	"context"
	"errors"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func RetrieveMetric(ctx context.Context, storage storage.Storage, metricType string, metricName string) (metric.Metric, error) {
	return storage.Retrieve(ctx, metric.MetricType(metricType), metricName)
}

func UpdateMetric(ctx context.Context, storage storage.Storage, m metric.Metric, metricValue any) error {
	x := storage.Update(ctx, m, metricValue)
	return x
}

func AddNewMetric(ctx context.Context, storage storage.Storage, metricType string, metricName string, metricValue any) (metric.Metric, error) {
	m, err := NewMetricWithValue(metricType, metricName, metricValue)
	if err != nil {
		return nil, err
	}
	err = storage.Add(ctx, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewMetricWithValue(metricType string, metricName string, metricValue any) (metric.Metric, error) {
	m, err := metric.NewMetric(metric.MetricType(metricType), metricName)
	if err != nil {
		return nil, err
	}

	if gauge, ok := m.(*metric.Gauge); ok {
		if err := gauge.Update(metricValue); err != nil {
			return nil, err
		}
	} else if counter, ok := m.(*metric.Counter); ok {
		if err := counter.Update(metricValue); err != nil {
			return nil, err
		}
	} else {
		return nil, metric.ErrorInvalidMetricType
	}

	return m, nil
}

func UpdateMetricByValue(ctx context.Context, storage storage.Storage, metricType string, metricName string, metricValue interface{}) (metric.Metric, error) {

	m, err := RetrieveMetric(ctx, storage, metricType, metricName)

	if err != nil {
		if !errors.Is(err, common.ErrorMetricDoesNotExist) {
			return nil, err
		} else {
			m, err = AddNewMetric(ctx, storage, metricType, metricName, metricValue)
			if err != nil {
				return nil, err
			}
		}
	} else {
		err = UpdateMetric(ctx, storage, m, metricValue)
		if err != nil {
			return nil, err
		}
	}

	return m, nil

}

// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"g22","type":"gauge","value":123.12}'
// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"c33","type":"counter","delta":3}'

func fillValue(m metric.Metric, r *dto.Metrics) error {
	switch m.GetType() {
	case metric.MetricTypeCounter:
		int64Val, ok := m.GetValue().(int64)
		if ok {
			r.Delta = &int64Val
		} else {
			return common.ErrorTypeConversion
		}
	case metric.MetricTypeGauge:
		float64Val, ok := m.GetValue().(float64)
		if ok {
			r.Value = &float64Val
		} else {
			return common.ErrorTypeConversion
		}
	default:
		return metric.ErrorInvalidMetricType
	}
	return nil
}
