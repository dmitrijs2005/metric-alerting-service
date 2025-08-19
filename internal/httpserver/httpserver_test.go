package httpserver

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
)

func TestEchoGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.GetLogger()
	stor := memory.NewMemStorage()

	s := &HTTPServer{Address: ":8080", Storage: stor, logger: log}

	// test handler
	e := echo.New()
	e.GET("/slow", func(c echo.Context) error {

		time.Sleep(500 * time.Millisecond) // slow request
		return c.String(http.StatusOK, "OK")
	})

	// starting the server
	go func() {
		if err := s.Run(ctx, e); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	// wait a little for server to start
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{}
	doneCh := make(chan string, 1)

	// starting slow request
	go func() {
		resp, err := client.Get("http://localhost:8080/slow")
		if err != nil {
			t.Errorf("request failed: %v", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		doneCh <- string(body)
	}()

	// waiting for request processing to start
	time.Sleep(100 * time.Millisecond)
	cancel() // shutdown

	// checking if request is processed
	select {
	case res := <-doneCh:
		if res != "OK" {
			t.Errorf("expected 'OK', got %q", res)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("request did not complete gracefully")
	}
}
