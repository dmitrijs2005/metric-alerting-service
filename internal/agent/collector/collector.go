// Package collector implements a metrics collector from runtime and system sources.
// It updates metric values such as memory statistics, CPU usage, and additional custom data.
package collector

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Collector struct {
	Data         sync.Map
	PollInterval time.Duration
}

func NewCollector(pollInterval time.Duration) *Collector {
	return &Collector{
		PollInterval: pollInterval,
	}
}

func (c *Collector) updateGauge(metricName string, metricValue float64) {
	val, exists := c.Data.Load(metricName)

	if !exists {
		val = metric.NewGauge(metricName)
	}

	if gauge, ok := val.(*metric.Gauge); ok {
		gauge.Update(metricValue)
	}
	c.Data.Store(metricName, val)
}

func (c *Collector) updateCounter(metricName string, metricValue int64) {
	val, exists := c.Data.Load(metricName)

	if !exists {
		val = metric.NewCounter(metricName)
	}

	if counter, ok := val.(*metric.Counter); ok {
		counter.Update(metricValue)
	}
	c.Data.Store(metricName, val)

}

// collectMemStats собирает статистику памяти.
func (c *Collector) collectMemStats() *runtime.MemStats {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	return ms
}

// updateMemStats обновляет метрики, связанные с runtime.MemStats.
func (c *Collector) updateMemStats(ms *runtime.MemStats) {
	c.updateGauge("Alloc", float64(ms.Alloc))
	c.updateGauge("BuckHashSys", float64(ms.BuckHashSys))
	c.updateGauge("Frees", float64(ms.Frees))
	c.updateGauge("GCCPUFraction", float64(ms.GCCPUFraction))
	c.updateGauge("GCSys", float64(ms.GCSys))
	c.updateGauge("HeapAlloc", float64(ms.HeapAlloc))
	c.updateGauge("HeapIdle", float64(ms.HeapIdle))
	c.updateGauge("HeapInuse", float64(ms.HeapInuse))
	c.updateGauge("HeapObjects", float64(ms.HeapObjects))
	c.updateGauge("HeapReleased", float64(ms.HeapReleased))
	c.updateGauge("HeapSys", float64(ms.HeapSys))
	c.updateGauge("LastGC", float64(ms.LastGC))
	c.updateGauge("Lookups", float64(ms.Lookups))
	c.updateGauge("MCacheInuse", float64(ms.MCacheInuse))
	c.updateGauge("MCacheSys", float64(ms.MCacheSys))
	c.updateGauge("MSpanInuse", float64(ms.MSpanInuse))
	c.updateGauge("MSpanSys", float64(ms.MSpanSys))
	c.updateGauge("Mallocs", float64(ms.Mallocs))
	c.updateGauge("NextGC", float64(ms.NextGC))
	c.updateGauge("NumForcedGC", float64(ms.NumForcedGC))
	c.updateGauge("NumGC", float64(ms.NumGC))
	c.updateGauge("OtherSys", float64(ms.OtherSys))
	c.updateGauge("PauseTotalNs", float64(ms.PauseTotalNs))
	c.updateGauge("StackInuse", float64(ms.StackInuse))
	c.updateGauge("StackSys", float64(ms.StackSys))
	c.updateGauge("Sys", float64(ms.Sys))
	c.updateGauge("TotalAlloc", float64(ms.TotalAlloc))
}

// updateAdditionalMetrics обновляет дополнительные метрики.
func (c *Collector) updateAdditionalMetrics() {
	c.updateGauge("RandomValue", rand.Float64())
	c.updateCounter("PollCount", 1)
}

func (c *Collector) RunStatUpdater(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		select {
		case <-time.After(c.PollInterval):
			ms := c.collectMemStats()
			c.updateMemStats(ms)
			c.updateAdditionalMetrics()
		case <-ctx.Done():
			return
		}
	}

}

func (c *Collector) RunPSUtilMetricsUpdater(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		select {
		case <-time.After(c.PollInterval):
			c.updatePSUtilsMemoryMetrics(ctx)
			c.updatePSUtilsCPUMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}

}

func (c *Collector) updatePSUtilsMemoryMetrics(ctx context.Context) {

	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		common.WriteToConsole("Error reading GOPSUTIL memory data")
		return
	}
	c.updateGauge("TotalMemory", float64(v.Total))
	c.updateGauge("FreeMemory", float64(v.Free))

}

func (c *Collector) updatePSUtilsCPUMetrics(ctx context.Context) {

	percentages, err := cpu.PercentWithContext(ctx, time.Second, true)
	if err != nil {
		common.WriteToConsole("Error reading GOPSUTIL CPU data")
		return
	}

	for i, perc := range percentages {
		metricName := GetIndexedMetricNameItoa("CPUutilization", i+1)
		c.updateGauge(metricName, perc)
	}

}

func GetIndexedMetricNameSprintf(name string, index int) string {
	return fmt.Sprintf("%s%d", name, index)
}

func GetIndexedMetricNameItoa(name string, index int) string {
	return name + strconv.Itoa(index)
}
