package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
)

type MetricAgent struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Data           map[string]metric.Metric
	ServerURL      string
	HTTPClient     *http.Client
}

func NewMetricAgent(pollInterval time.Duration, reportInterval time.Duration, serverURL string) *MetricAgent {

	return &MetricAgent{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Data:           make(map[string]metric.Metric),
		ServerURL:      serverURL,
		HTTPClient:     &http.Client{},
	}
}

func (a *MetricAgent) updateGauge(metricName string, metricValue float64) {
	m, exists := a.Data[metricName]

	if !exists {
		m = metric.NewGauge(metricName)
		a.Data[metricName] = m
	}

	m.Update(metricValue)
}

func (a *MetricAgent) updateCounter(metricName string, metricValue int64) {
	m, exists := a.Data[metricName]

	if !exists {
		m = metric.NewCounter(metricName)
		a.Data[metricName] = m
	}

	m.Update(metricValue)

}

// Метрики нужно отправлять по протоколу HTTP методом POST:
// Формат данных — http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>.
// Адрес сервера — http://localhost:8080.
// Заголовок — Content-Type: text/plain.
// Пример запроса к серверу:

// POST /update/counter/someMetric/527 HTTP/1.1
// Host: localhost:8080
// Content-Length: 0
// Content-Type: text/plain
// Пример ответа от сервера:

func (a *MetricAgent) SendMetric(m metric.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	//v := fmt.Sprintf("%v", m.GetValue())

	url := a.ServerURL
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

func (a *MetricAgent) RunReport(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		//fmt.Println("Sending metrics...")
		sendWg := sync.WaitGroup{}
		for _, v := range a.Data {
			sendWg.Add(1)
			go a.SendMetric(v, &sendWg)
		}
		sendWg.Wait()

		time.Sleep(a.ReportInterval)

	}

}

func (a *MetricAgent) RunPoll(wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		//fmt.Println("Updating metrics...")

		ms := &runtime.MemStats{}
		runtime.ReadMemStats(ms)

		a.updateGauge("Alloc", float64(ms.Alloc))
		a.updateGauge("BuckHashSys", float64(ms.BuckHashSys))
		a.updateGauge("Frees", float64(ms.Frees))
		a.updateGauge("GCCPUFraction", float64(ms.GCCPUFraction))
		a.updateGauge("GCSys", float64(ms.GCSys))
		a.updateGauge("HeapAlloc", float64(ms.HeapAlloc))
		a.updateGauge("HeapIdle", float64(ms.HeapIdle))
		a.updateGauge("HeapInuse", float64(ms.HeapInuse))
		a.updateGauge("HeapObjects", float64(ms.HeapObjects))
		a.updateGauge("HeapReleased", float64(ms.HeapReleased))
		a.updateGauge("HeapSys", float64(ms.HeapSys))
		a.updateGauge("LastGC", float64(ms.LastGC))
		a.updateGauge("Lookups", float64(ms.Lookups))
		a.updateGauge("MCacheInuse", float64(ms.MCacheInuse))
		a.updateGauge("MCacheSys", float64(ms.MCacheSys))
		a.updateGauge("MSpanInuse", float64(ms.MSpanInuse))
		a.updateGauge("MSpanSys", float64(ms.MSpanSys))
		a.updateGauge("Mallocs", float64(ms.Mallocs))
		a.updateGauge("NextGC", float64(ms.NextGC))
		a.updateGauge("NumForcedGC", float64(ms.NumForcedGC))
		a.updateGauge("NumGC", float64(ms.NumGC))
		a.updateGauge("OtherSys", float64(ms.OtherSys))
		a.updateGauge("PauseTotalNs", float64(ms.PauseTotalNs))
		a.updateGauge("StackInuse", float64(ms.StackInuse))
		a.updateGauge("StackSys", float64(ms.StackSys))
		a.updateGauge("Sys", float64(ms.Sys))
		a.updateGauge("TotalAlloc", float64(ms.TotalAlloc))

		a.updateGauge("RandomValue", rand.Float64())
		a.updateCounter("PollCount", 1)

		time.Sleep(a.PollInterval)
	}
}

func (a *MetricAgent) Run() {

	wg := sync.WaitGroup{}

	wg.Add(1)
	go a.RunPoll(&wg)

	wg.Add(1)
	go a.RunReport(&wg)

	wg.Wait()

	fmt.Println("Finished...")

}
