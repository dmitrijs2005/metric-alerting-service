package main

import (
	"os"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/buildinfo"
)

func main() {

	buildinfo.PrintBuildData(os.Stdout)

	cfg := config.LoadConfig()
	a := agent.NewMetricAgent(cfg)
	a.Run()
}
