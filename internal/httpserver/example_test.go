package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
)

// Example_UpdateHandler demonstrates how to use UpdateHandler with an in-memory HTTP server.
func ExampleHTTPServer_UpdateHandler() {
	// Creating dependencies
	address := "localhost:8080"
	basePath := "/"
	storage := memory.NewMemStorage()
	log := logger.GetLogger()

	// Initializing server
	srv := NewHTTPServer(address, basePath, storage, log)

	// Echo instance
	e := echo.New()
	e.POST("/update/:type/:name/:value", srv.UpdateHandler)

	// HTTP test server
	ts := httptest.NewServer(e)
	defer ts.Close()

	// Performing HTTP-request
	resp, err := http.Post(fmt.Sprintf("%s/update/counter/requests/5", ts.URL), "text/plain", nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))

	// Output:
	// 200 OK
}

func ExampleHTTPServer_ValueHandler() {
	storage := memory.NewMemStorage()
	log := logger.GetLogger()

	// Initializing server
	srv := NewHTTPServer("localhost:8080", "/", storage, log)

	// Saving metric
	err := storage.Add(context.Background(), &metric.Counter{Name: "requests", Value: 42})
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.GET("/value/:type/:name", srv.ValueHandler)

	ts := httptest.NewServer(e)
	defer ts.Close()

	// Performing HTTP-request
	resp, err := http.Get(fmt.Sprintf("%s/value/counter/requests", ts.URL))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))

	// Output:
	// 200 42
}

func ExampleHTTPServer_UpdateJSONHandler() {
	storage := memory.NewMemStorage()
	log := logger.GetLogger()

	// Initializing server
	srv := NewHTTPServer("localhost:8080", "/", storage, log)

	e := echo.New()
	e.POST("/update/", srv.UpdateJSONHandler)

	ts := httptest.NewServer(e)
	defer ts.Close()

	// Test data
	jsonData := `{"id":"temperature","type":"gauge","value":36.6}`

	// Performing HTTP-request
	resp, err := http.Post(fmt.Sprintf("%s/update/", ts.URL), "application/json", bytes.NewBufferString(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))

	// Output:
	// 200 {"value":36.6,"id":"temperature","type":"gauge"}
}

func ExampleHTTPServer_UpdatesJSONHandler() {
	storage := memory.NewMemStorage()
	log := logger.GetLogger()

	// Initializing server
	srv := NewHTTPServer("localhost:8080", "/", storage, log)

	e := echo.New()
	e.POST("/updates/", srv.UpdatesJSONHandler)

	ts := httptest.NewServer(e)
	defer ts.Close()

	_ = storage.Add(context.Background(), &metric.Counter{Name: "requests", Value: 1})
	_ = storage.Add(context.Background(), &metric.Gauge{Name: "load", Value: 1})

	// Test data
	jsonData := `[{"id":"requests","type":"counter","delta":5},{"id":"load","type":"gauge","value":0.83}]`

	// Performing HTTP-request
	resp, err := http.Post(fmt.Sprintf("%s/updates/", ts.URL), "application/json", bytes.NewBufferString(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))

	// Output:
	// 200 [{"delta":6,"id":"requests","type":"counter"},{"value":0.83,"id":"load","type":"gauge"}]
}

func ExampleHTTPServer_ValueJSONHandler() {
	ctx := context.Background()
	storage := memory.NewMemStorage()
	log := logger.GetLogger()

	// Initializing server
	srv := NewHTTPServer("localhost:8080", "/", storage, log)

	_ = storage.Add(ctx, &metric.Gauge{Name: "temperature", Value: 36.6})

	e := echo.New()
	e.POST("/value/", srv.ValueJSONHandler)

	ts := httptest.NewServer(e)
	defer ts.Close()

	// Test data
	jsonData := `{"id":"temperature","type":"gauge"}`

	// Performing HTTP-request
	resp, err := http.Post(fmt.Sprintf("%s/value/", ts.URL), "application/json", bytes.NewBufferString(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))

	// Output:
	// 200 {"value":36.6,"id":"temperature","type":"gauge"}
}
