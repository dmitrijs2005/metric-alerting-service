package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dumpsaver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

type App struct {
	ctx        context.Context
	config     *config.Config
	cancelFunc context.CancelFunc
	logger     logger.Logger
	storage    storage.Storage
	saver      dumpsaver.DumpSaver
}

func NewApp(logger logger.Logger) *App {
	ctx, cancelFunc := context.WithCancel(context.Background())

	config := config.LoadConfig()

	storage := storage.NewMemStorage()
	saver := dumpsaver.NewFileSaver(config.FileStoragePath, storage)

	return &App{ctx: ctx, config: config, cancelFunc: cancelFunc, logger: logger, storage: storage, saver: saver}
}

func (app *App) Run() {

	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		app.cancelFunc()
	}()

	app.logger.Infow("Starting app",
		"restore", app.config.Restore,
		"store_interval", app.config.StoreInterval,
		"file_storage_path", app.config.FileStoragePath)

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
		s := httpserver.NewHTTPServer(app.ctx, app.config.EndpointAddr, app.storage, app.logger)
		if err := s.Run(); err != nil {
			app.cancelFunc()
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
				case <-app.ctx.Done():
					app.logger.Info("Background saving task received cancellation signal. Exiting...")
					return
				case <-ticker.C:
					// Place your periodic task logic here.
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
