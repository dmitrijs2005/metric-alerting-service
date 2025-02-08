package metric

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metric interface {
	GetType() MetricType
	GetName() string
	GetValue() interface{}
	Update(interface{}) error
}
