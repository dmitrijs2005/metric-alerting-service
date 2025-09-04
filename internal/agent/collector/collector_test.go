package collector

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewCollector_SetsPollInterval(t *testing.T) {
	c := NewCollector(123 * time.Millisecond)
	assert.Equal(t, 123*time.Millisecond, c.PollInterval)
}

func TestGetIndexedMetricNameHelpers(t *testing.T) {
	assert.Equal(t, "CPUutilization1", GetIndexedMetricNameSprintf("CPUutilization", 1))
	assert.Equal(t, "CPUutilization10", GetIndexedMetricNameItoa("CPUutilization", 10))
}

func TestUpdateGaugeAndCounter_StoreTypes(t *testing.T) {
	c := &Collector{}
	// Gauge path
	c.updateGauge("Alloc", 42.5)
	v, ok := c.Data.Load("Alloc")
	require.True(t, ok, "Alloc should be stored")
	_, isGauge := v.(*metric.Gauge)
	assert.True(t, isGauge, "Alloc should be a *metric.Gauge")

	// Counter path
	c.updateCounter("PollCount", 1)
	v, ok = c.Data.Load("PollCount")
	require.True(t, ok, "PollCount should be stored")
	_, isCounter := v.(*metric.Counter)
	assert.True(t, isCounter, "PollCount should be a *metric.Counter")

}

func TestRunStatUpdater_PopulatesMetricsAndCancels(t *testing.T) {
	c := NewCollector(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go c.RunStatUpdater(ctx, &wg)

	// Give it a little time to tick and populate
	time.Sleep(30 * time.Millisecond)
	cancel()
	wg.Wait()

	// A few representative runtime metrics should exist
	keys := []string{
		"Alloc",
		"HeapAlloc",
		"NumGC",
		"Sys",
		"RandomValue",
		"PollCount",
	}
	for _, k := range keys {
		_, ok := c.Data.Load(k)
		assert.Truef(t, ok, "expected metric %q to be present", k)
	}
}

func TestUpdatePSUtilsMemoryMetrics_PopulatesTotals(t *testing.T) {
	c := NewCollector(0)
	// Call directly; should be quick and not block
	c.updatePSUtilsMemoryMetrics(context.Background())

	// Presence checks
	_, ok := c.Data.Load("TotalMemory")
	assert.True(t, ok, "TotalMemory should be present")

	_, ok = c.Data.Load("FreeMemory")
	assert.True(t, ok, "FreeMemory should be present")
}

func TestUpdatePSUtilsCPUMetrics(t *testing.T) {
	oldFunc := cpuPercentFunc // экспортируемая переменная
	defer func() { cpuPercentFunc = oldFunc }()

	cpuPercentFunc = func(ctx context.Context, interval time.Duration, percpu bool) ([]float64, error) {
		return []float64{10.5, 20.0}, nil
	}

	c := NewCollector(0)
	ctx := context.Background()

	c.updatePSUtilsCPUMetrics(ctx)

	v1, ok := c.Data.Load("CPUutilization1")
	require.True(t, ok)
	require.InDelta(t, 10.5, v1.(*metric.Gauge).GetValue(), 0.0001)

	v2, ok := c.Data.Load("CPUutilization2")
	require.True(t, ok)
	require.InDelta(t, 20.0, v2.(*metric.Gauge).GetValue(), 0.0001)
}

func TestCollector_RunPSUtilMetricsUpdater(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	c := &Collector{PollInterval: 50 * time.Millisecond}
	var wg sync.WaitGroup
	wg.Add(1)

	go c.RunPSUtilMetricsUpdater(ctx, &wg)

	wg.Wait()
}
