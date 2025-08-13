// Package server initializes and runs the main application server.
// It configures storage backends, handles graceful shutdown, restores and saves metric dumps,
// and starts the HTTP server for metric collection.
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
}

func NewApp(logger logger.Logger) (*App, error) {

	config := config.LoadConfig()
	return &App{config: config, logger: logger}, nil
}

func (app *App) initSignalHandler(cancelFunc context.CancelFunc) {
	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancelFunc()
	}()
}

func (app *App) initDumpSyncAgent(s storage.Storage) (*file.FileSaver, error) {
	return file.NewFileSaver(app.config.FileStoragePath, s), nil
}

func (app *App) initStorage(ctx context.Context) (storage.Storage, error) {

	var s storage.Storage

	if app.config.DatabaseDSN == "" {

		s = memory.NewMemStorage()
	} else {

		var err error

		pgClient, err := db.NewPostgresClient(app.config.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		if err := pgClient.RunMigrations(ctx); err != nil {
			return nil, err
		}

		s = pgClient

	}

	return s, nil

}

func (app *App) closeDBIfNeeded(s storage.Storage) (bool, error) {

	db, ok := s.(storage.DBStorage)
	if ok {
		err := db.Close()
		return true, err
	}

	return false, nil

}

func (app *App) restoreDumpIfNeeded(ctx context.Context, a *file.FileSaver, s storage.Storage) (bool, error) {

	if !app.config.Restore {
		return false, nil
	}

	_, ok := s.(storage.DBStorage)
	if ok {
		return false, nil
	}

	err := a.RestoreDump(ctx)
	if err != nil {
		return false, err
	}

	return true, nil

}

func (app *App) startHTTPServer(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, s storage.Storage) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := httpserver.NewHTTPServer(app.config.EndpointAddr, app.config.Key, s, app.logger, app.config.CryptoKey)
		if err != nil {
			app.logger.Error(err)
			cancelFunc()
		}
		if err := s.Run(ctx); err != nil {
			app.logger.Error(err)
			cancelFunc()
		}
	}()
}

func (app *App) saveDump(ctx context.Context, a *file.FileSaver) {

	err := a.SaveDump(ctx)

	if err != nil {
		app.logger.Error(err)
	} else {
		app.logger.Info("Dump saved successfully")
	}

}

func (app *App) initPeriodicDumpSaveIfNeeded(ctx context.Context, s storage.Storage, a *file.FileSaver, wg *sync.WaitGroup) {

	_, ok := s.(storage.DBStorage)
	if ok {
		return
	}

	if app.config.StoreInterval == 0 {
		return
	}

	//if store interval is not 0, launching regular dump saving task
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
				app.saveDump(ctx, a)
			}
		}
	}()
}

func (app *App) saveDumpIfNeeded(ctx context.Context, s storage.Storage, a *file.FileSaver) {

	_, ok := s.(storage.DBStorage)
	if ok {
		return
	}

	// значение 0 делает запись синхронной
	if app.config.StoreInterval != 0 {
		return
	}

	app.saveDump(ctx, a)
}

func (app *App) Run() {

	ctx, cancelFunc := context.WithCancel(context.Background())

	app.logger.Infow("Starting app",
		"restore", app.config.Restore,
		"store_interval", app.config.StoreInterval,
		"file_storage_path", app.config.FileStoragePath,
		"database_dsn", app.config.DatabaseDSN,
	)

	app.initSignalHandler(cancelFunc)

	s, err := app.initStorage(ctx)
	if err != nil {
		app.logger.Errorw("Storage initialization error", "err", err)
		cancelFunc()
		return
	}

	a, err := app.initDumpSyncAgent(s)
	if err != nil {
		app.logger.Errorw("Dump sync agent initialization error", "err", err)
		cancelFunc()
		return
	}

	restored, err := app.restoreDumpIfNeeded(ctx, a, s)
	if err != nil {
		app.logger.Errorw("Dump restore error", "err", err)
	}

	if restored {
		app.logger.Infow("Dump restored successfully")
	}

	defer func() {
		closed, err := app.closeDBIfNeeded(s)
		if err != nil {
			app.logger.Errorw("Error closing database connection:", "err", err)
		} else {
			if closed {
				app.logger.Infow("Database closed")
			}
		}
	}()

	var wg sync.WaitGroup

	app.startHTTPServer(ctx, cancelFunc, &wg, s)
	app.initPeriodicDumpSaveIfNeeded(ctx, s, a, &wg)

	wg.Wait()

	app.saveDumpIfNeeded(ctx, s, a)

}
