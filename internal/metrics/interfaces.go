package metrics

type MetricType string

const (
	MetricTypeGauge         MetricType = "gauge"
	MetricTypeCounter       MetricType = "counter"
	ErrorInvalidMetricType  string     = "invalid metric type"
	ErrorInvalidMetricName  string     = "invalid metric name"
	ErrorInvalidMetricValue string     = "invalid metric value"
)

type Metric interface {
	GetType() MetricType
	GetName() string
	GetValue() interface{}
	Update(interface{}) error
}
