package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/cmd/agent/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent"
)

func main() {

	cfg := config.LoadConfig()
	a := agent.NewMetricAgent(cfg.PollInterval, cfg.ReportInterval, cfg.EndpointAddr)
	a.Run()
}
