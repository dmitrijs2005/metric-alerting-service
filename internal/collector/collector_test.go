package collector

import (
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricAgent_updateGauge(t *testing.T) {
	a := &Collector{
		Data: make(map[string]metric.Metric),
	}
	type args struct {
		metricName  string
		metricValue float64
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test 1", args{"gauge1", 1.234}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.updateGauge(tt.args.metricName, tt.args.metricValue)
			assert.Equal(t, a.Data[tt.args.metricName].GetType(), metric.MetricTypeGauge)
			assert.Equal(t, a.Data[tt.args.metricName].GetValue(), tt.args.metricValue)
		})
	}
}

func TestMetricAgent_updateCounter(t *testing.T) {
	a := &Collector{
		Data: make(map[string]metric.Metric),
	}
	type args struct {
		metricName  string
		metricValue int64
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test 1", args{"counter1", 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.updateCounter(tt.args.metricName, tt.args.metricValue)
			assert.Equal(t, a.Data[tt.args.metricName].GetType(), metric.MetricTypeCounter)
			assert.Equal(t, a.Data[tt.args.metricName].GetValue(), tt.args.metricValue)
		})
	}
}
