package metric

// MetricType defines the type of a metric, such as "gauge" or "counter".
type MetricType string

const (
	// MetricTypeGauge represents a float64 metric that can increase or decrease.
	MetricTypeGauge MetricType = "gauge"

	// MetricTypeCounter represents an int64 metric that only increases.
	MetricTypeCounter MetricType = "counter"
)

// Metric represents a single named metric of a specific type.
// It supports getting its type, name, current value, and updating the value.
type Metric interface {
	GetType() MetricType
	GetName() string
	GetValue() interface{}
	Update(interface{}) error
}
