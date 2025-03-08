package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/httpserver"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/db"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/file"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
)

type App struct {
	config *config.Config
	logger logger.Logger
	//storage storage.Storage
	//saver   file.DumpSaver
}

func NewApp(logger logger.Logger) (*App, error) {

	config := config.LoadConfig()

	//, storage: storage, saver: saver

	return &App{config: config, logger: logger}, nil
}

func (app *App) Run() {

	ctx, cancelFunc := context.WithCancel(context.Background())

	var s storage.Storage

	useDB := app.config.DatabaseDSN != ""

	if !useDB {
		s = memory.NewMemStorage()
	} else {
		var err error
		s, err = db.NewPostgresClient(app.config.DatabaseDSN)
		if err != nil {
			app.logger.Errorw("Error", "err", err)
			cancelFunc()
		}
	}

	db, ok := s.(storage.DBStorage)

	if ok {
		db.RunMigrations(ctx)
		defer func() {
			if err := db.Close(); err != nil {
				app.logger.Errorw("Error closing database connection:", "err", err)
			} else {
				app.logger.Infow("Database closed")
			}
		}()
	}

	saver := file.NewFileSaver(app.config.FileStoragePath, s)

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

	// restoring data from dump
	if !useDB && app.config.Restore {
		err := saver.RestoreDump(ctx)
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
		s := httpserver.NewHTTPServer(ctx, app.config.EndpointAddr, s, app.logger)
		if err := s.Run(); err != nil {
			cancelFunc()
		}
	}()

	//if store interval is not 0, launching regular dump saving task
	if !useDB && app.config.StoreInterval > 0 {
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
					saver.SaveDump(ctx)
				}
			}
		}()
	}

	wg.Wait()

	// значение 0 делает запись синхронной
	if !useDB && app.config.StoreInterval == 0 {
		err := saver.SaveDump(ctx)

		if err != nil {
			app.logger.Error(err)
		} else {
			app.logger.Info("Dump saved successfully")
		}
	}

}
