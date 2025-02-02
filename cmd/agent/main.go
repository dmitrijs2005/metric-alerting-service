package main

import "github.com/dmitrijs2005/metric-alerting-service/internal/agent"

func main() {

	parseFlags()
	a := agent.NewMetricAgent(options.PollInterval, options.ReportInterval, options.EndpointAddr)
	a.Run()
}
