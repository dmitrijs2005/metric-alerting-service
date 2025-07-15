package metric

import (
	"strconv"
)

// Gauge represents a floating-point metric that can go up or down.
type Gauge struct {
	Name  string  // Name is the unique name of the metric.
	Value float64 // Value holds the current float64 value of the metric.
}

// GetType returns the type of the metric ("gauge").
func (c *Gauge) GetType() MetricType {
	return MetricTypeGauge
}

// GetName returns the name of the gauge metric.
func (c *Gauge) GetName() string {
	return c.Name
}

// GetValue returns the current value of the gauge as interface{}.
func (c *Gauge) GetValue() interface{} {
	return c.Value
}

// tryParseFloat64Value attempts to convert the input value into float64.
// It supports both float64 and string input formats.
func (c *Gauge) tryParseFloat64Value(value interface{}) (float64, error) {

	// if value is string
	stringVal, ok := value.(string)

	if ok {
		val, err := strconv.ParseFloat(stringVal, 64)
		if err != nil {
			return -1, ErrorInvalidMetricValue
		}
		return val, nil
	}

	val, ok := value.(float64)

	if !ok {
		return -1, ErrorInvalidMetricValue
	}
	return val, nil

}

// Update parses and sets the gauge's value.
// Accepts either a float64 or a string that can be parsed as float64.
func (c *Gauge) Update(value interface{}) error {

	val, err := c.tryParseFloat64Value(value)

	if err != nil {
		return err
	}

	c.Value = val
	return nil
}

// NewGauge creates a new Gauge with the specified name and an initial value of 0.
func NewGauge(name string) *Gauge {
	return &Gauge{Name: name}
}

// MustNewGauge creates a new Gauge with the specified name and value.
// It is intended for convenient creation when the value is known and trusted.
func MustNewGauge(name string, val float64) *Gauge {
	m := NewGauge(name)
	m.Value = val
	return m
}
