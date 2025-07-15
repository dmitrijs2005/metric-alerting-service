package metric

import (
	"strconv"
)

// Counter represents a 64-bit integer metric that can only increase.
// It is commonly used to track things like the number of requests, events, or errors.
type Counter struct {
	Name  string // Name is the unique name of the counter metric.
	Value int64  // Value holds the current counter value.
}

// GetType returns the metric type ("counter").
func (c *Counter) GetType() MetricType {
	return MetricTypeCounter
}

// GetName returns the name of the counter metric.
func (c *Counter) GetName() string {
	return c.Name
}

// GetValue returns the current value of the counter as interface{}.
func (c *Counter) GetValue() interface{} {
	return c.Value
}

// tryParseInt64Value attempts to convert the input value into int64.
// It supports both int64 and string input formats.
// Returns an error if the value is invalid or not convertible.
func (c *Counter) tryParseInt64Value(value interface{}) (int64, error) {

	// if value is string
	stringVal, ok := value.(string)

	if ok {
		val, err := strconv.ParseInt(stringVal, 10, 64)
		if err != nil {
			return -1, ErrorInvalidMetricValue
		}
		return val, nil
	}

	val, ok := value.(int64)

	if !ok {
		return -1, ErrorInvalidMetricValue
	}
	return val, nil

}

// Update adds the given value to the current counter value.
// It accepts an int64 or a string that can be parsed to int64.
// Returns an error if the value cannot be parsed.
func (c *Counter) Update(value interface{}) error {

	val, err := c.tryParseInt64Value(value)

	if err != nil {
		return err
	}

	c.Value += val
	return nil
}

// NewCounter creates a new Counter with the given name and a value of 0.
func NewCounter(name string) *Counter {
	return &Counter{Name: name}
}

// MustNewCounter creates a new Counter with the given name and initial value.
// It is intended for use when the initial value is known and trusted.
func MustNewCounter(name string, val int64) *Counter {
	m := NewCounter(name)
	m.Value = val
	return m
}
