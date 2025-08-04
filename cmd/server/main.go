package main

import (
	"os"

	"github.com/dmitrijs2005/metric-alerting-service/internal/buildinfo"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server"
)

func main() {

	buildinfo.PrintBuildData(os.Stdout)

	log := logger.GetLogger()
	defer logger.Sync()

	app, err := server.NewApp(log)
	if err != nil {
		log.Errorw(err.Error())
		return
	}

	app.Run()
}
