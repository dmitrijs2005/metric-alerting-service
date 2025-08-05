// Package dto defines data transfer objects used for communication between the agent and the server.
// It includes representations of metrics in JSON format for both gauge and counter types.
package dto

// Metrics represents a metric data transfer object.
type Metrics struct {
	// ID is the name of the metric.
	ID string `json:"id"`

	// MType indicates the metric type: "gauge" or "counter".
	MType string `json:"type"`

	// Delta is the value for a "counter" metric. Can be nil.
	Delta *int64 `json:"delta,omitempty"`

	// Value is the value for a "gauge" metric. Can be nil.
	Value *float64 `json:"value,omitempty"`
}
