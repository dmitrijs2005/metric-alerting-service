package sender

import (
	"bytes"
	"compress/gzip"
	"context"
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
)

type Sender struct {
	ReportInterval time.Duration
	ServerURL      string
	Data           *sync.Map
	Key            string
	SendRateLimit  int
	Jobs           chan metric.Metric
}

func NewSender(data *sync.Map, reportInterval time.Duration, serverURL string, key string, sendRateLimit int) *Sender {
	return &Sender{
		ReportInterval: reportInterval,
		Data:           data,
		ServerURL:      serverURL,
		Key:            key,
		SendRateLimit:  sendRateLimit,
		Jobs:           make(chan metric.Metric),
	}
}

func (s *Sender) worker(ind int, jobs <-chan metric.Metric) {

	label := fmt.Sprintf("Worker #%d", ind+1)

	common.WriteToConsole(fmt.Sprintf("%s started", label))
	defer common.WriteToConsole(fmt.Sprintf("%s exited", label))

	for j := range jobs {
		s.SendMetric(j)
		common.WriteToConsole(fmt.Sprintf("%s sent metric %s", label, j.GetName()))
	}

}

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

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)

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

	return nil
}

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
		sign, err := common.CreateAes256Signature(jsonData, s.Key)
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
