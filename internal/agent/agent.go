package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/collector"
	"github.com/dmitrijs2005/metric-alerting-service/internal/sender"
)

type MetricAgent struct {
	collector *collector.Collector
	sender    *sender.Sender
}

func NewMetricAgent(pollInterval time.Duration, reportInterval time.Duration, serverURL string) *MetricAgent {

	collector := collector.NewCollector(pollInterval)
	sender := sender.NewSender(reportInterval, &collector.Data, serverURL)

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
