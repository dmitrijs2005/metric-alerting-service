// Package agent implements an agent for collecting and sending metrics.
// It defines the structures and functions required for initializing and running
// metric collectors (collector) and senders (sender), as well as handling shutdown signals.
package agent

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/collector"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/sender"
	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
)

type MetricAgent struct {
	collector *collector.Collector
	sender    *sender.Sender
}

func NewMetricAgent(cfg *config.Config) (*MetricAgent, error) {

	collector := collector.NewCollector(cfg.PollInterval)
	sender, err := sender.NewSender(&collector.Data, cfg.ReportInterval, cfg.EndpointAddr, cfg.Key, cfg.SendRateLimit, cfg.CryptoKey)

	if err != nil {
		return nil, err
	}

	return &MetricAgent{
		collector: collector,
		sender:    sender,
	}, nil
}

func (a *MetricAgent) initSignalHandler(cancelFunc context.CancelFunc) {
	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelFunc()
	}()
}

func (a *MetricAgent) Run() {

	ctx, cancelFunc := context.WithCancel(context.Background())
	a.initSignalHandler(cancelFunc)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go a.collector.RunStatUpdater(ctx, &wg)

	wg.Add(1)
	go a.collector.RunPSUtilMetricsUpdater(ctx, &wg)

	wg.Add(1)
	go a.sender.Run(ctx, &wg)

	common.WriteToConsole("Agent started...")

	wg.Wait()

	common.WriteToConsole("Agent finished...")

}
