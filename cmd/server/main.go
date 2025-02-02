package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func main() {

	//adress := ":8080"
	stor := storage.NewMemStorage()

	parseFlags()

	s := httpserver.NewHTTPServer(flagEndpointAddr, stor)
	s.Run()
}
