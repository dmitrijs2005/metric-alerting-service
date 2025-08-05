// Command server is the entry point for the metrics server binary.
// It initializes logging, prints build info, and runs the HTTP server.
package main

import (
	"os"

	"github.com/dmitrijs2005/metric-alerting-service/internal/buildinfo"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server"
)

// main initializes the server application:
// - prints build metadata,
// - configures the logger,
// - creates and runs the HTTP server.
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
