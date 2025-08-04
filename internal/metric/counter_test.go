package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkCounter_Update(b *testing.B) {
	c := MustNewCounter("counter1", 0)

	for i := 0; i < b.N; i++ {
		c.Update(1)
	}
}

func TestNewCounter(t *testing.T) {
	c := NewCounter("requests_total")
	require.NotNil(t, c)
	assert.Equal(t, "requests_total", c.Name)
	assert.Equal(t, int64(0), c.Value)
}

func TestMustNewCounter(t *testing.T) {
	c := MustNewCounter("errors_total", 42)
	require.NotNil(t, c)
	assert.Equal(t, "errors_total", c.Name)
	assert.Equal(t, int64(42), c.Value)
}

func TestCounter_Getters(t *testing.T) {
	c := MustNewCounter("my_metric", 123)

	assert.Equal(t, "my_metric", c.GetName())
	assert.Equal(t, MetricTypeCounter, c.GetType())

	val, ok := c.GetValue().(int64)
	require.True(t, ok)
	assert.Equal(t, int64(123), val)
}

func TestCounter_Update_WithInt64(t *testing.T) {
	c := MustNewCounter("counter1", 10)

	err := c.Update(int64(5))
	require.NoError(t, err)
	assert.Equal(t, int64(15), c.Value)
}

func TestCounter_Update_WithString(t *testing.T) {
	c := MustNewCounter("counter2", 10)

	err := c.Update("7")
	require.NoError(t, err)
	assert.Equal(t, int64(17), c.Value)
}

func TestCounter_Update_InvalidString(t *testing.T) {
	c := MustNewCounter("counter3", 0)

	err := c.Update("not_a_number")
	require.ErrorIs(t, err, ErrorInvalidMetricValue)
	assert.Equal(t, int64(0), c.Value)
}

func TestCounter_Update_InvalidType(t *testing.T) {
	c := MustNewCounter("counter4", 0)

	err := c.Update(3.14)
	require.ErrorIs(t, err, ErrorInvalidMetricValue)
	assert.Equal(t, int64(0), c.Value)
}

func TestCounter_tryParseInt64Value(t *testing.T) {
	c := &Counter{}

	tests := []struct {
		name     string
		input    interface{}
		expected int64
		wantErr  bool
	}{
		{"string ok", "123", 123, false},
		{"string bad", "abc", -1, true},
		{"int64 ok", int64(42), 42, false},
		{"float bad", 3.14, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := c.tryParseInt64Value(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrorInvalidMetricValue)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}
