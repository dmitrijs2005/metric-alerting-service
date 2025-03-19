package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/collector"
	"github.com/dmitrijs2005/metric-alerting-service/internal/sender"
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

func (a *MetricAgent) Run() {

	ctx := context.Background()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go a.collector.Run(&wg)

	wg.Add(1)
	go a.sender.Run(ctx, &wg)

	wg.Wait()

	fmt.Println("Finished...")

}
