package collector

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

type Collector struct {
	PollInterval time.Duration
	Data         map[string]metric.Metric
}

func NewCollector(pollInterval time.Duration) *Collector {
	return &Collector{
		PollInterval: pollInterval,
		Data:         make(map[string]metric.Metric),
	}
}

func (c *Collector) updateGauge(metricName string, metricValue float64) {
	m, exists := c.Data[metricName]

	if !exists {
		m = metric.NewGauge(metricName)
		c.Data[metricName] = m
	}

	m.Update(metricValue)
}

func (c *Collector) updateCounter(metricName string, metricValue int64) {
	m, exists := c.Data[metricName]

	if !exists {
		m = metric.NewCounter(metricName)
		c.Data[metricName] = m
	}

	m.Update(metricValue)

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

func (c *Collector) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		ms := c.collectMemStats()
		c.updateMemStats(ms)
		c.updateAdditionalMetrics()
		time.Sleep(c.PollInterval)
	}
}
