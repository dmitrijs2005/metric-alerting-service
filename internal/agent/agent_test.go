package agent

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricAgent_updateGauge(t *testing.T) {
	a := &MetricAgent{
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
	a := &MetricAgent{
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

			agent := &MetricAgent{
				ServerURL:  mockServer.URL,
				HTTPClient: &http.Client{},
			}

			var wg sync.WaitGroup
			wg.Add(1)

			agent.SendMetric(tt.metric, &wg)

			wg.Wait()
		})
	}

}
