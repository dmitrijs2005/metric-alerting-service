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

func NewMetricAgent(cfg *config.Config) *MetricAgent {

	collector := collector.NewCollector(cfg.PollInterval)
	sender := sender.NewSender(cfg.ReportInterval, &collector.Data, cfg.EndpointAddr, cfg.Key)

	return &MetricAgent{
		collector: collector,
		sender:    sender,
	}
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
	go a.collector.Run(ctx, &wg)

	wg.Add(1)
	go a.sender.Run(ctx, &wg)

	common.WriteToConsole("Agent started...")

	wg.Wait()

	common.WriteToConsole("Agent finished...")

}
