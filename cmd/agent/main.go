// Command agent is the entry point for the metric agent binary.
// It prints build information and runs the metric collection agent.
package main

import (
	"os"

	"github.com/dmitrijs2005/metric-alerting-service/internal/agent"
	"github.com/dmitrijs2005/metric-alerting-service/internal/agent/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/buildinfo"
)

// main initializes the application:
// - prints build metadata to stdout,
// - loads agent configuration,
// - creates and runs the metric agent.
func main() {

	buildinfo.PrintBuildData(os.Stdout)

	cfg := config.LoadConfig()
	a := agent.NewMetricAgent(cfg)
	a.Run()
}
