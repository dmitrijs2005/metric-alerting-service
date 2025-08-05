package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkGauge_Update(b *testing.B) {
	c := MustNewGauge("gauge1", 0)

	for i := 0; i < b.N; i++ {
		c.Update(i)
	}
}

func TestNewGauge(t *testing.T) {
	g := NewGauge("cpu_usage")
	require.NotNil(t, g)
	assert.Equal(t, "cpu_usage", g.Name)
	assert.Equal(t, 0.0, g.Value)
}

func TestMustNewGauge(t *testing.T) {
	g := MustNewGauge("temperature", 36.6)
	require.NotNil(t, g)
	assert.Equal(t, "temperature", g.Name)
	assert.Equal(t, 36.6, g.Value)
}

func TestGauge_Getters(t *testing.T) {
	g := MustNewGauge("metric_x", 12.34)

	assert.Equal(t, "metric_x", g.GetName())
	assert.Equal(t, MetricTypeGauge, g.GetType())

	val, ok := g.GetValue().(float64)
	require.True(t, ok)
	assert.Equal(t, 12.34, val)
}

func TestGauge_Update_WithFloat64(t *testing.T) {
	g := MustNewGauge("test", 1.1)

	err := g.Update(2.2)
	require.NoError(t, err)
	assert.Equal(t, 2.2, g.Value)
}

func TestGauge_Update_WithString(t *testing.T) {
	g := MustNewGauge("test", 0.0)

	err := g.Update("3.14")
	require.NoError(t, err)
	assert.Equal(t, 3.14, g.Value)
}

func TestGauge_Update_InvalidString(t *testing.T) {
	g := MustNewGauge("test", 0.0)

	err := g.Update("not-a-number")
	require.ErrorIs(t, err, ErrorInvalidMetricValue)
	assert.Equal(t, 0.0, g.Value)
}

func TestGauge_Update_InvalidType(t *testing.T) {
	g := MustNewGauge("test", 0.0)

	err := g.Update(42) // int, not float64
	require.ErrorIs(t, err, ErrorInvalidMetricValue)
	assert.Equal(t, 0.0, g.Value)
}

func TestGauge_tryParseFloat64Value(t *testing.T) {
	g := &Gauge{}

	tests := []struct {
		name     string
		input    interface{}
		expected float64
		wantErr  bool
	}{
		{"valid string", "2.71", 2.71, false},
		{"invalid string", "abc", -1, true},
		{"valid float64", 5.5, 5.5, false},
		{"invalid int", 7, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := g.tryParseFloat64Value(tt.input)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrorInvalidMetricValue)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
		})
	}
}
