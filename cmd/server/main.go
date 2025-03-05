package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server"
)

func main() {

	log := logger.GetLogger()
	defer logger.Sync()

	app, err := server.NewApp(log)
	if err != nil {
		log.Errorw(err.Error())
		return
	}

	app.Run()
}
