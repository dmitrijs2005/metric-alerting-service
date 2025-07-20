package sender

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/collector"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricAgent_SendMetric(t *testing.T) {

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1}

	tests := []struct {
		name   string
		metric metric.Metric
	}{
		{"Test Counter", metric1},
		{"Test Gauge", metric2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
				assert.Equal(t, r.URL.Path, "/update/", "Unexpected URL path")
				w.WriteHeader(http.StatusOK)
			}))
			defer mockServer.Close()

			agent := &Sender{
				ServerURL:      mockServer.URL,
				ReportInterval: 10 * time.Second,
				GzipWriterPool: &sync.Pool{
					New: func() interface{} {
						w, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
						if err != nil {
							panic(fmt.Sprintf("gzip.NewWriterLevel failed: %v", err))
						}
						return w
					},
				},
				BufferPool: &sync.Pool{
					New: func() interface{} {
						return new(bytes.Buffer)
					},
				},
			}

			agent.SendMetric(tt.metric)

		})
	}
}

func TestMetricAgent_SendMetrics(t *testing.T) {

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
		assert.Equal(t, r.URL.Path, "/updates/", "Unexpected URL path")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	collector := collector.NewCollector(1)

	agent := &Sender{
		ServerURL:      mockServer.URL,
		ReportInterval: 10 * time.Second,
	}

	agent.Data = &collector.Data

	agent.Data.Store(metric1.GetName(), metric1)
	agent.Data.Store(metric2.GetName(), metric2)

	agent.SendAllMetricsInOneBatch()

}
