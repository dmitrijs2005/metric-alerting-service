package metrics

import (
	"errors"
	"regexp"
)

func IsValidMetricType(t string) bool {
	return MetricType(t) == MetricTypeGauge || MetricType(t) == MetricTypeCounter
}

// Metric names and labels (from Prometheus docs)
// Every time series is uniquely identified by its metric name and optional key-value pairs called labels.

// Metric names:

// Specify the general feature of a system that is measured (e.g. http_requests_total - the total number of HTTP requests received).
// Metric names may contain ASCII letters, digits, underscores, and colons. It must match the regex [a-zA-Z_:][a-zA-Z0-9_:]*.

func IsMetricNameValid(n string) bool {
	pattern := `^[a-zA-Z_:][a-zA-Z0-9_:]*$`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	re.MatchString(n)

	return re.MatchString(n)
}

func NewMetric(metricType MetricType, metricName string) (Metric, error) {

	if !IsMetricNameValid(metricName) {
		return nil, errors.New(ErrorInvalidMetricName)
	}

	switch metricType {
	case MetricTypeGauge:
		return NewGauge(metricName), nil
	case MetricTypeCounter:
		return NewCounter(metricName), nil
	default:
		return nil, errors.New(ErrorInvalidMetricType)
	}

}
