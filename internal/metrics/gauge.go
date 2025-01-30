package metrics

import (
	"errors"
	"strconv"
)

type Gauge struct {
	Name  string
	Value float64
}

func (c *Gauge) GetType() MetricType {
	return MetricTypeGauge
}

func (c *Gauge) GetName() string {
	return c.Name
}

func (c *Gauge) GetValue() interface{} {
	return c.Value
}

func (c *Gauge) tryParseFloat64Value(value interface{}) (float64, error) {

	// if value is string
	stringVal, ok := value.(string)

	if ok {
		val, err := strconv.ParseFloat(stringVal, 64)
		if err != nil {
			return -1, errors.New(ErrorInvalidMetricValue)
		}
		return val, nil
	}

	val, ok := value.(float64)

	if !ok {
		return -1, errors.New(ErrorInvalidMetricValue)
	}
	return val, nil

}

func (c *Gauge) Update(value interface{}) error {

	val, err := c.tryParseFloat64Value(value)

	if err != nil {
		return err
	}

	c.Value = val
	return nil
}

func NewGauge(name string) *Gauge {
	return &Gauge{Name: name}
}
