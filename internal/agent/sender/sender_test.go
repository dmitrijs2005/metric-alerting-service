package sender

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/require"
)

// func TestMetricAgent_SendMetric(t *testing.T) {

// 	metric1 := &metric.Counter{Name: "counter1", Value: 1}
// 	metric2 := &metric.Gauge{Name: "gauge1", Value: 1}

// 	tests := []struct {
// 		metric metric.Metric
// 		name   string
// 	}{
// 		{name: "Test Counter", metric: metric1},
// 		{name: "Test Gauge", metric: metric2},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
// 				assert.Equal(t, r.URL.Path, "/update/", "Unexpected URL path")
// 				w.WriteHeader(http.StatusOK)
// 			}))
// 			defer mockServer.Close()

// 			agent := &Sender{
// 				ServerURL:      mockServer.URL,
// 				ReportInterval: 10 * time.Second,
// 				GzipWriterPool: &sync.Pool{
// 					New: func() interface{} {
// 						w, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
// 						if err != nil {
// 							panic(fmt.Sprintf("gzip.NewWriterLevel failed: %v", err))
// 						}
// 						return w
// 					},
// 				},
// 				BufferPool: &sync.Pool{
// 					New: func() interface{} {
// 						return new(bytes.Buffer)
// 					},
// 				},
// 			}

// 			agent.SendMetric(tt.metric)

// 		})
// 	}
// }

// func TestMetricAgent_SendMetrics(t *testing.T) {

// 	metric1 := &metric.Counter{Name: "counter1", Value: 1}
// 	metric2 := &metric.Gauge{Name: "gauge1", Value: 1}

// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
// 		assert.Equal(t, r.URL.Path, "/updates/", "Unexpected URL path")
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer mockServer.Close()

// 	collector := collector.NewCollector(1)

// 	agent := &Sender{
// 		ServerURL:      mockServer.URL,
// 		ReportInterval: 10 * time.Second,
// 	}

// 	agent.Data = &collector.Data

// 	agent.Data.Store(metric1.GetName(), metric1)
// 	agent.Data.Store(metric2.GetName(), metric2)

// 	agent.SendAllMetricsInOneBatch()

// }

func TestMetricToDto_ValidGauge(t *testing.T) {
	data := &sync.Map{}
	s, err := NewSender(data, time.Second, "http://localhost", "", 1, "")
	require.NoError(t, err)

	m := metric.NewGauge("cpu_load")
	m.Update(0.42)

	dto, err := s.MetricToDto(m)
	require.NoError(t, err)
	require.Equal(t, "cpu_load", dto.ID)
	require.Equal(t, "gauge", dto.MType)
	require.NotNil(t, dto.Value)
	require.Equal(t, 0.42, *dto.Value)
}

func TestSendMetric_Success(t *testing.T) {
	received := make(chan []byte, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		gr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, _ := io.ReadAll(gr)
		received <- body

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	s, _ := NewSender(data, time.Second, ts.URL, "", 1, "")

	m := metric.NewGauge("cpu_load")
	m.Update(1.23)

	err := s.SendMetric(m)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(<-received, &got))
	require.Equal(t, "cpu_load", got["id"])
	require.Equal(t, "gauge", got["type"])
}

func TestSendAllMetricsInOneBatch_Success(t *testing.T) {
	received := make(chan []byte, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gr, _ := gzip.NewReader(r.Body)
		defer gr.Close()
		body, _ := io.ReadAll(gr)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	g := metric.NewGauge("temp")
	g.Update(99.9)
	data.Store("temp", g)

	s, _ := NewSender(data, time.Second, ts.URL, "", 1, "")

	err := s.SendAllMetricsInOneBatch()
	require.NoError(t, err)

	var arr []map[string]interface{}
	require.NoError(t, json.Unmarshal(<-received, &arr))
	require.Equal(t, "temp", arr[0]["id"])
	require.Equal(t, "gauge", arr[0]["type"])
}

func TestRun_SendsMetrics(t *testing.T) {
	var mu sync.Mutex
	count := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	g := metric.NewGauge("load")
	g.Update(0.99)
	data.Store("load", g)

	s, _ := NewSender(data, 50*time.Millisecond, ts.URL, "", 1, "")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go s.Run(ctx, &wg)

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, count, 1, "expected at least one batch sent")
}
