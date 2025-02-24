package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func main() {

	cfg := config.LoadConfig()

	stor := storage.NewMemStorage()
	saver := storage.NewFileSaver(cfg.FileStoragePath)

	log := logger.GetLogger()
	defer logger.Sync()

	s := httpserver.NewHTTPServer(cfg.EndpointAddr, cfg.StoreInterval, cfg.Restore, stor, saver, log)
	s.Run()
}
