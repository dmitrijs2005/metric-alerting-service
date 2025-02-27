package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server"
)

func main() {

	log := logger.GetLogger()
	defer logger.Sync()

	app := server.NewApp(log)
	app.Run()
}
