package metrics

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

func GetAllowedMetricTypes() []string {
	return []string{MetricTypeGauge, MetricTypeCounter}
}

type Metric struct {
	Type  string
	Name  string
	Value interface{}
}

func parseMetricValue(metricType string, metricValue string) (interface{}, error) {
	if metricType == MetricTypeGauge {
		float64Value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return nil, errors.New("incorrect float64 value")
		}
		return float64Value, nil
	} else if metricType == MetricTypeCounter {
		int64Value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return nil, errors.New("incorrect int64 value")
		}
		return int64Value, nil
	} else {
		return nil, errors.New("unknown metric type")
	}
}

func NewMetricFromRequest(metricType string, metricName string, metricValue string) (*Metric, error) {

	if !slices.Contains(GetAllowedMetricTypes(), metricType) {
		return nil, errors.New("invalid metric type")
	}

	fmt.Println(IsMetricNameValid(metricName))

	if !IsMetricNameValid(metricName) {
		return nil, errors.New("invalid metric name")
	}

	val, err := parseMetricValue(metricType, metricValue)

	if err != nil {
		return nil, err
	}

	return &Metric{Type: metricType, Name: metricName, Value: val}, nil
}
