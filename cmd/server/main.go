package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func main() {

	//adress := ":8080"
	stor := storage.NewMemStorage()

	cfg := config.LoadConfig()

	s := httpserver.NewHTTPServer(cfg.EndpointAddr, stor)
	s.Run()
}
