package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/db"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dumpsaver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

type App struct {
	config  *config.Config
	logger  logger.Logger
	storage storage.Storage
	saver   dumpsaver.DumpSaver
}

func NewApp(logger logger.Logger) (*App, error) {

	config := config.LoadConfig()

	storage := storage.NewMemStorage()

	saver := dumpsaver.NewFileSaver(config.FileStoragePath, storage)

	return &App{config: config, logger: logger, storage: storage, saver: saver}, nil
}

func (app *App) Run() {

	ctx, cancelFunc := context.WithCancel(context.Background())

	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelFunc()
	}()

	app.logger.Infow("Starting app",
		"restore", app.config.Restore,
		"store_interval", app.config.StoreInterval,
		"file_storage_path", app.config.FileStoragePath,
		"database_dsn", app.config.DatabaseDSN,
	)

	//s := "host=localhost user=postgres password=mysecretpassword sslmode=disable"
	dbClient, err := db.NewPostgresClient(app.config.DatabaseDSN)
	defer func() {
		if err := dbClient.Close(); err != nil {
			app.logger.Errorw("Error closing database connection:", "err", err)
		} else {
			app.logger.Infow("Database closed")
		}

	}()

	if err != nil {
		app.logger.Errorw("Error", "err", err)
		return
	}

	// restoring data from dump
	if app.config.Restore {
		err := app.saver.RestoreDump()
		if err != nil {
			app.logger.Error(err.Error())
		} else {
			app.logger.Info("Dump restored successfully!!!")
		}
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s := httpserver.NewHTTPServer(ctx, app.config.EndpointAddr, app.storage, dbClient, app.logger)
		if err := s.Run(); err != nil {
			cancelFunc()
		}
	}()

	// if store interval is not 0, launching regular dump saving task
	if app.config.StoreInterval > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(app.config.StoreInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					app.logger.Info("Background saving task received cancellation signal. Exiting...")
					return
				case <-ticker.C:
					// Periodic task logic
					app.logger.Info("Performing regular saving task")
					app.saver.SaveDump()
				}
			}
		}()
	}

	wg.Wait()

	// значение 0 делает запись синхронной
	if app.config.StoreInterval == 0 {
		err := app.saver.SaveDump()

		if err != nil {
			app.logger.Error(err)
		} else {
			app.logger.Info("Dump saved successfully")
		}
	}

}
