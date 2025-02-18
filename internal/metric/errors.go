package metric

import "errors"

var (
	ErrorInvalidMetricType  = errors.New("invalid metric type")
	ErrorInvalidMetricName  = errors.New("invalid metric name")
	ErrorInvalidMetricValue = errors.New("invalid metric value")
)
