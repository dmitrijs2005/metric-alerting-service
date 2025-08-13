// Package sender is responsible for preparing and sending metrics to the monitoring server.
// It supports batching, gzip compression, optional AES-256 signature, and concurrent workers.
package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
)

// Sender handles sending metrics to the monitoring server.
// It supports compression, signature, batching, and concurrent processing.
type Sender struct {
	Jobs           chan metric.Metric
	GzipWriterPool *sync.Pool
	BufferPool     *sync.Pool
	Data           *sync.Map
	Key            string
	ServerURL      string
	ReportInterval time.Duration
	SendRateLimit  int
	PubKey         *rsa.PublicKey
}

// NewSender creates and returns a new Sender instance.
//
// Parameters:
//   - data: pointer to a sync.Map containing metrics.
//   - reportInterval: how often metrics are sent.
//   - serverURL: base URL of the monitoring server.
//   - key: optional secret key for signing payloads.
//   - sendRateLimit: number of concurrent workers.
//
// Returns:
//   - *Sender: a new Sender instance.
func NewSender(data *sync.Map, reportInterval time.Duration, serverURL string, key string, sendRateLimit int, cryptoKey string) (*Sender, error) {

	var pubKey *rsa.PublicKey
	var err error

	if cryptoKey != "" {
		pubKey, err = secure.LoadRSAPublicKeyFromPEM(cryptoKey)
		if err != nil {
			return nil, err
		}
	}

	return &Sender{
		ReportInterval: reportInterval,
		Data:           data,
		ServerURL:      serverURL,
		Key:            key,
		SendRateLimit:  sendRateLimit,
		Jobs:           make(chan metric.Metric),
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
		PubKey: pubKey,
	}, nil
}

func (s *Sender) worker(ind int, jobs <-chan metric.Metric) {

	label := fmt.Sprintf("Worker #%d", ind+1)

	common.WriteToConsole(fmt.Sprintf("%s started", label))
	defer common.WriteToConsole(fmt.Sprintf("%s exited", label))

	for j := range jobs {
		err := s.SendMetric(j)
		if err != nil {
			common.WriteToConsole(fmt.Sprintf("Error sending metric %v", err))
		} else {
			common.WriteToConsole(fmt.Sprintf("%s sent metric %s", label, j.GetName()))
		}
	}

}

// MetricToDto converts a metric.Metric into a DTO (Data Transfer Object) for JSON serialization.
//
// Parameters:
//   - m: the metric to convert.
//
// Returns:
//   - *dto.Metrics: the converted DTO.
//   - error: if the value type is invalid for its metric type.
func (s *Sender) MetricToDto(m metric.Metric) (*dto.Metrics, error) {
	data := &dto.Metrics{ID: m.GetName(), MType: string(m.GetType())}

	if m.GetType() == metric.MetricTypeCounter {
		v, ok := m.GetValue().(int64)
		if ok {
			data.Delta = &v
		} else {
			return nil, common.ErrorTypeConversion
		}
	} else if m.GetType() == metric.MetricTypeGauge {
		v, ok := m.GetValue().(float64)
		if ok {
			data.Value = &v
		} else {
			return nil, common.ErrorTypeConversion
		}
	}
	return data, nil
}

// SendMetric sends a single metric to the /update/ endpoint with gzip compression.
//
// Parameters:
//   - m: the metric to be sent.
//
// Returns:
//   - error: if conversion, compression, or network transmission fails.
func (s *Sender) SendMetric(m metric.Metric) error {

	url := s.ServerURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	data, err := s.MetricToDto(m)
	if err != nil {
		return err
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return common.ErrorMarshallingJSON
	}

	if s.PubKey != nil {
		encryptedData, err := secure.EncryptRSAOAEPChunked(jsonData, s.PubKey)
		if err != nil {
			return common.NewWrappedError("Error sending request", err)
		}
		jsonData = []byte(encryptedData)
	}

	buf := s.BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	zb := s.GzipWriterPool.Get().(*gzip.Writer)
	zb.Reset(buf)

	defer func() {
		s.GzipWriterPool.Put(zb)
		s.BufferPool.Put(buf)
	}()

	_, err = zb.Write(jsonData)
	if err != nil {
		return err
	}

	err = zb.Close()
	if err != nil {
		return err
	}

	url = fmt.Sprintf("%s/update/", url)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return err
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// SendAllMetricsInOneBatch sends all metrics in a single batch request to /updates/ endpoint.
// It uses gzip compression and optionally adds AES-256 signature.
//
// Returns:
//   - error: in case of serialization, compression, or network issues.
func (s *Sender) SendAllMetricsInOneBatch() error {

	url := s.ServerURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	data := make([]*dto.Metrics, 0)

	// Concurrent reading (safe)
	s.Data.Range(func(key, val interface{}) bool {

		// Convert interface{} to *metric.Counter and update value
		if m, ok := val.(metric.Metric); ok {
			item, err := s.MetricToDto(m)

			if err != nil {
				common.WriteToConsole(fmt.Sprintf("Error converting metric to DTO: %v", err))
			}

			data = append(data, item)
		}

		return true
	})

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return common.ErrorMarshallingJSON
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)

	_, err = zb.Write(jsonData)
	if err != nil {
		return common.NewWrappedError("Error writing to buffer request", err)
	}

	err = zb.Close()
	if err != nil {
		return common.NewWrappedError("Error closing buffer", err)
	}

	url = fmt.Sprintf("%s/updates/", url)

	common.WriteToConsole("sending...")

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return common.NewWrappedError("Error creating request", err)
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// signing if key is specified
	if s.Key != "" {
		var sign []byte
		sign, err = secure.CreateAes256Signature(jsonData, s.Key)
		if err != nil {
			return common.NewWrappedError("Error signing request", err)
		}
		req.Header.Set("HashSHA256", base64.RawStdEncoding.EncodeToString(sign))
	}

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return common.NewWrappedError("Error sending request", err)
	}
	defer resp.Body.Close()

	common.WriteToConsole("reply received...")

	return nil

}

// SendAllMetrics pushes all collected metrics to the jobs channel for worker processing.
//
// Returns:
//   - error: always nil (included for interface compatibility).
func (s *Sender) SendAllMetrics() error {
	// Concurrent reading (safe)
	s.Data.Range(func(key, val interface{}) bool {
		// Convert interface{} to *metric.Counter and update value
		if m, ok := val.(metric.Metric); ok {
			s.Jobs <- m
		}
		return true
	})
	return nil
}

// Run launches the sender loop that periodically sends all metrics.
// It starts worker goroutines and listens for context cancellation.
//
// Parameters:
//   - ctx: context for graceful shutdown.
//   - wg: WaitGroup to signal when sender has stopped.
func (s *Sender) Run(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	var workerWg sync.WaitGroup
	for i := 0; i < s.SendRateLimit; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			s.worker(i, s.Jobs)
		}()
	}

loop:
	for {
		select {
		case <-time.After(s.ReportInterval):
			{
				_, err := common.RetryWithResult(ctx, func() (interface{}, error) {
					err := s.SendAllMetrics()
					return nil, err
				})

				if err != nil {
					common.WriteToConsole(err.Error())
				}
			}
		case <-ctx.Done():
			break loop
		}
	}

	close(s.Jobs)
	workerWg.Wait()

}
