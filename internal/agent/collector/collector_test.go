package collector

import (
	"runtime"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricAgent_updateGauge(t *testing.T) {
	a := &Collector{}
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

			val, ok := a.Data.Load(tt.args.metricName)
			assert.True(t, ok)

			m, ok := val.(metric.Metric)
			assert.True(t, ok)

			assert.Equal(t, m.GetType(), metric.MetricTypeGauge)
			assert.Equal(t, m.GetValue(), tt.args.metricValue)
		})
	}
}

func TestMetricAgent_updateCounter(t *testing.T) {
	a := &Collector{}
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

			val, ok := a.Data.Load(tt.args.metricName)
			assert.True(t, ok)

			m, ok := val.(metric.Metric)
			assert.True(t, ok)

			assert.Equal(t, m.GetType(), metric.MetricTypeCounter)
			assert.Equal(t, m.GetValue(), tt.args.metricValue)
		})
	}
}

func TestMetricAgent_updateAdditionalMetrics(t *testing.T) {
	a := &Collector{}
	t.Run("Test", func(t *testing.T) {
		a.updateAdditionalMetrics()

		names := []string{"RandomValue", "PollCount"}

		for _, name := range names {
			val, ok := a.Data.Load(name)
			assert.True(t, ok)

			_, ok = val.(metric.Metric)
			assert.True(t, ok)
		}
	})
}

func TestMetricAgent_updateMemStats(t *testing.T) {
	a := &Collector{}
	t.Run("Test", func(t *testing.T) {

		ms := &runtime.MemStats{}
		a.updateMemStats(ms)

		names := []string{
			"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
			"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
			"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

		for _, name := range names {
			val, ok := a.Data.Load(name)
			assert.True(t, ok)

			_, ok = val.(metric.Metric)
			assert.True(t, ok)
		}
	})
}

func BenchmarkIndexedMetricName(b *testing.B) {
	b.Run("sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetIndexedMetricNameSprintf("CPUutilization", i)
		}
	})

	b.Run("itoa", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetIndexedMetricNameItoa("CPUutilization", i)
		}
	})
}
