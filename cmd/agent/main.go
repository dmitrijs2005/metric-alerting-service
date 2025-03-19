package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
)

func main() {

	cfg := config.LoadConfig()
	a := agent.NewMetricAgent(cfg)
	a.Run()
}
