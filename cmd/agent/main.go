package main

import "github.com/dmitrijs2005/metric-alerting-service/internal/agent"

func main() {

	pollInterval := 2
	reportInterval := 10
	serverURL := "http://localhost:8080"

	a := agent.NewMetricAgent(pollInterval, reportInterval, serverURL)
	a.Run()
}
