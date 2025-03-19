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
}

func NewSender(reportInterval time.Duration, data *sync.Map, serverURL string, key string) *Sender {
	return &Sender{
		ReportInterval: reportInterval,
		Data:           data,
		ServerURL:      serverURL,
		Key:            key,
	}
}

func (s *Sender) SendMetric(m metric.Metric, wg *sync.WaitGroup) error {
	defer wg.Done()

	url := s.ServerURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	data := &dto.Metrics{ID: m.GetName(), MType: string(m.GetType())}

	if m.GetType() == metric.MetricTypeCounter {
		v, ok := m.GetValue().(int64)
		if ok {
			data.Delta = &v
		} else {
			return common.ErrorTypeConversion
		}
	} else if m.GetType() == metric.MetricTypeGauge {
		v, ok := m.GetValue().(float64)
		if ok {
			data.Value = &v
		} else {
			return common.ErrorTypeConversion
		}
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

func (s *Sender) WriteToConsole(msg string) {
	fmt.Printf("%v %s \n", time.Now(), msg)
}

// func (s *Sender) CreateSignature(body []byte) ([]byte, error) {
// 	h := hmac.New(sha256.New, []byte(s.Key))
// 	_, err := h.Write(body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return h.Sum(nil), nil
// }

func (s *Sender) SendMetrics() error {

	url := s.ServerURL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	data := make([]*dto.Metrics, 0)

	// Concurrent reading (safe)
	s.Data.Range(func(key, val interface{}) bool {

		// Convert interface{} to *metric.Counter and update value
		if m, ok := val.(metric.Metric); ok {
			item := &dto.Metrics{ID: m.GetName(), MType: string(m.GetType())}

			if m.GetType() == metric.MetricTypeCounter {
				v, ok := m.GetValue().(int64)
				if ok {
					item.Delta = &v
				} else {
					return false
				}
			} else if m.GetType() == metric.MetricTypeGauge {
				v, ok := m.GetValue().(float64)
				if ok {
					item.Value = &v
				} else {
					return false
				}
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

	s.WriteToConsole("sending...")

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

	s.WriteToConsole("reply received...")

	return nil

}

func (s *Sender) Run(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		_, err := common.RetryWithResult(ctx, func() (interface{}, error) {
			err := s.SendMetrics()
			return nil, err
		})

		if err != nil {
			s.WriteToConsole(err.Error())
		}

		time.Sleep(s.ReportInterval)

	}

}
