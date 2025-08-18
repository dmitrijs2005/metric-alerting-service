package agent

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
	"github.com/stretchr/testify/require"
)

func TestNewMetricAgent_Success(t *testing.T) {
	cfg := &config.Config{
		PollInterval:   10 * time.Millisecond,
		ReportInterval: 10 * time.Millisecond,
		EndpointAddr:   "localhost:9999",
		Key:            "",
		SendRateLimit:  1,
		CryptoKey:      "",
	}

	a, err := NewMetricAgent(cfg)
	require.NoError(t, err)
	require.NotNil(t, a)
	require.NotNil(t, a.collector)
	require.NotNil(t, a.sender)
}

func TestSignalHandler_CancelsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	a := &MetricAgent{}
	a.initSignalHandler(cancel)

	// send SIGTERM into process
	p, _ := os.FindProcess(os.Getpid())
	require.NoError(t, p.Signal(syscall.SIGTERM))

	// context should cancel quickly
	select {
	case <-ctx.Done():
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context not canceled after SIGTERM")
	}
}

func TestRun_ShutsDownOnSignal(t *testing.T) {
	cfg := &config.Config{
		PollInterval:   20 * time.Millisecond,
		ReportInterval: 20 * time.Millisecond,
		EndpointAddr:   "localhost:9999",
		Key:            "",
		SendRateLimit:  1,
		CryptoKey:      "",
	}

	a, err := NewMetricAgent(cfg)
	require.NoError(t, err)

	done := make(chan struct{})

	go func() {
		a.Run()
		close(done)
	}()

	// allow agent to start
	time.Sleep(50 * time.Millisecond)

	// send SIGINT to trigger shutdown
	p, _ := os.FindProcess(os.Getpid())
	require.NoError(t, p.Signal(syscall.SIGINT))

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("agent did not exit after SIGINT")
	}
}
