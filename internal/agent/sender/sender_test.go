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

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/require"
)

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

func TestSender_GracefulShutdown_WaitsForInFlightMetrics(t *testing.T) {
	var mu sync.Mutex
	var received []string

	// fake server with artificial delay
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// simulate slow server processing
		time.Sleep(300 * time.Millisecond)

		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Errorf("failed to create gzip reader: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer gr.Close()

		var m dto.Metrics
		if err := json.NewDecoder(gr).Decode(&m); err != nil {
			t.Errorf("failed to decode json: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		received = append(received, m.ID)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// prepare sync.Map with a single metric
	data := &sync.Map{}
	data.Store("counter1", metric.NewCounter("counter1"))

	// create Sender with short report interval
	s, err := NewSender(data, 100*time.Millisecond, srv.URL, "", 1, "")
	if err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go s.Run(ctx, &wg)

	// wait until first batch is definitely in-flight
	time.Sleep(150 * time.Millisecond)

	// initiate shutdown while request is still processing
	cancel()
	wg.Wait()

	// check that the in-flight request was completed
	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 metric, got %d: %v", len(received), received)
	}
}
