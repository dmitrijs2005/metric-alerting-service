package metrics

import (
	"errors"
	"strconv"
)

type Counter struct {
	name  string
	value int64
}

func (c *Counter) GetType() MetricType {
	return MetricTypeCounter
}

func (c *Counter) GetName() string {
	return c.name
}

func (c *Counter) GetValue() interface{} {
	return c.value
}

func (c *Counter) tryParseInt64Value(value interface{}) (int64, error) {

	// if value is string
	stringVal, ok := value.(string)

	if ok {
		val, err := strconv.ParseInt(stringVal, 10, 64)
		if err != nil {
			return -1, errors.New(ErrorInvalidMetricValue)
		}
		return val, nil
	}

	val, ok := value.(int64)

	if !ok {
		return -1, errors.New(ErrorInvalidMetricValue)
	}
	return val, nil

}

func (c *Counter) Update(value interface{}) error {

	val, err := c.tryParseInt64Value(value)

	if err != nil {
		return err
	}

	c.value += val
	return nil
}

func NewCounter(name string) *Counter {
	return &Counter{name: name}
}
