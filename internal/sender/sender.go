package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

type Sender struct {
	ReportInterval time.Duration
	ServerURL      string
	Data           *sync.Map
}

func NewSender(reportInterval time.Duration, data *sync.Map, serverURL string) *Sender {
	return &Sender{
		ReportInterval: reportInterval,
		Data:           data,
		ServerURL:      serverURL,
	}
}

func (s *Sender) SendMetric(m metric.Metric, wg *sync.WaitGroup) {
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
			panic(ErrorTypeConversion)
		}
	} else if m.GetType() == metric.MetricTypeGauge {
		v, ok := m.GetValue().(float64)
		if ok {
			data.Value = &v
		} else {
			panic(ErrorTypeConversion)
		}
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(ErrorMarshallingJSON)
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)

	_, err = zb.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to buffer request:", err)
		return
	}

	err = zb.Close()
	if err != nil {
		fmt.Println("Error closing buffer:", err)
		return
	}

	url = fmt.Sprintf("%s/update/", url)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

}

func (s *Sender) SendMetrics() error {
	//defer wg.Done()

	//v := fmt.Sprintf("%v", m.GetValue())

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

	// for m := s.Data.Range() {

	// }

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ErrorMarshallingJSON
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)

	_, err = zb.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to buffer request:", err)
		return err
	}

	err = zb.Close()
	if err != nil {
		fmt.Println("Error closing buffer:", err)
		return err
	}

	url = fmt.Sprintf("%s/updates/", url)

	fmt.Println("sending...")

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set the content type to application/json
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Println("received...")

	return nil

}

func (s *Sender) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		err := s.SendMetrics()
		if err != nil {
			fmt.Println("Error sending metrics:", err)
		}

		time.Sleep(s.ReportInterval)

	}

}
