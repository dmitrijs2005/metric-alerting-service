package storage

import "errors"

const (
	MetricDoesNotExist  = "metric does not exist"
	MetricAlreadyExists = "metric already exists"
)

var ErrorMetricDoesNotExist = errors.New("metric does not exist")
