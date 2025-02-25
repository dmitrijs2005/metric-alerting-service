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
	Data           map[string]metric.Metric
}

func NewSender(reportInterval time.Duration, data map[string]metric.Metric, serverURL string) *Sender {
	return &Sender{
		ReportInterval: reportInterval,
		Data:           data,
		ServerURL:      serverURL,
	}
}

func (s *Sender) SendMetric(m metric.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	//v := fmt.Sprintf("%v", m.GetValue())

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

func (s *Sender) Run(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		//fmt.Println("Sending metrics...")
		sendWg := sync.WaitGroup{}
		for _, v := range s.Data {
			sendWg.Add(1)
			go s.SendMetric(v, &sendWg)
		}
		sendWg.Wait()

		time.Sleep(s.ReportInterval)

	}

}
