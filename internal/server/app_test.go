package server

import (
	"context"
	"errors"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/config"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApp_initStorage_Memory(t *testing.T) {
	app := &App{config: &config.Config{DatabaseDSN: ""}}
	st, err := app.initStorage(context.Background())
	require.NoError(t, err)
	require.IsType(t, &memory.MemStorage{}, st)
}

type fakeDBStorage struct {
	storage.Storage
	closed bool
}

func (f *fakeDBStorage) Close() error {
	f.closed = true
	return nil
}
func (f *fakeDBStorage) RunMigrations(ctx context.Context) error { return nil }
func (f *fakeDBStorage) Ping(ctx context.Context) error          { return nil }

func TestApp_closeDBIfNeeded(t *testing.T) {
	app := &App{}
	dbs := &fakeDBStorage{}
	closed, err := app.closeDBIfNeeded(dbs)
	require.NoError(t, err)
	require.True(t, closed)
	require.True(t, dbs.closed)
}

type fakeFileSaver struct {
	called bool
}

func (f *fakeFileSaver) RestoreDump(ctx context.Context) error {
	f.called = true
	return nil
}
func (f *fakeFileSaver) SaveDump(ctx context.Context) error {
	f.called = true
	return nil
}

func TestApp_restoreDumpIfNeeded(t *testing.T) {
	app := &App{config: &config.Config{Restore: true}}
	saver := &fakeFileSaver{}
	st := memory.NewMemStorage()

	ok, err := app.restoreDumpIfNeeded(context.Background(), saver, st)
	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, saver.called)
}

func TestApp_saveDumpIfNeeded(t *testing.T) {
	app := &App{config: &config.Config{StoreInterval: 0}, logger: logger.GetLogger()}
	saver := &fakeFileSaver{}
	st := memory.NewMemStorage()

	app.saveDumpIfNeeded(context.Background(), st, saver)
	require.True(t, saver.called)
}

func TestApp_startHTTPServer(t *testing.T) {
	app := &App{config: &config.Config{
		EndpointAddr: ":0",
	}, logger: logger.GetLogger()}
	st := memory.NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	app.startHTTPServer(ctx, cancel, &wg, st)
	cancel()
	wg.Wait()
}

func TestApp_Run_CancelsGracefully(t *testing.T) {
	app := &App{config: &config.Config{
		EndpointAddr:     ":0",
		GRPCEndpointAddr: ":0",
		StoreInterval:    time.Millisecond * 50,
	}}
	app.logger = logger.GetLogger() // если есть no-op логгер

	go func() {
		time.Sleep(200 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	app.Run()
}

type MockDBClient struct {
	mock.Mock
}

func (c *MockDBClient) Ping(ctx context.Context) error {
	return nil
}

func (c *MockDBClient) Add(ctx context.Context, m metric.Metric) error {
	return nil
}

func (c *MockDBClient) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return nil
}

func (c *MockDBClient) Retrieve(ctx context.Context, m metric.MetricType, n string) (metric.Metric, error) {
	return nil, nil
}

func (c *MockDBClient) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, nil
}

func NewMockDBClient() *MockDBClient {
	return &MockDBClient{}
}

func (c *MockDBClient) Close() error {
	return nil
}

func (c *MockDBClient) RunMigrations(ctx context.Context) error {
	return nil
}

func (c *MockDBClient) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return nil
}

type mockPostgresClient struct {
	storage.Storage
	runMigrationsFunc func(ctx context.Context) error
	closeFunc         func() error
}

func (m *mockPostgresClient) RunMigrations(ctx context.Context) error {
	if m.runMigrationsFunc != nil {
		return m.runMigrationsFunc(ctx)
	}
	return nil
}
func (m *mockPostgresClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
func (m *mockPostgresClient) Ping(ctx context.Context) error { return nil }

func TestApp_initStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("memory storage when DSN is empty", func(t *testing.T) {
		app := &App{config: &config.Config{DatabaseDSN: ""}}
		st, err := app.initStorage(ctx)
		require.NoError(t, err)
		require.IsType(t, &memory.MemStorage{}, st)
	})

	t.Run("postgres client ok", func(t *testing.T) {
		mockClient := &mockPostgresClient{
			runMigrationsFunc: func(ctx context.Context) error { return nil },
		}

		oldNew := newPostgresClient
		newPostgresClient = func(_ string) (storage.DBStorage, error) {
			return mockClient, nil
		}
		defer func() { newPostgresClient = oldNew }()

		app := &App{config: &config.Config{DatabaseDSN: "mock-dsn"}}
		st, err := app.initStorage(ctx)
		require.NoError(t, err)
		require.Equal(t, mockClient, st)
	})

	t.Run("postgres client migration error", func(t *testing.T) {
		mockClient := &mockPostgresClient{
			runMigrationsFunc: func(ctx context.Context) error { return errors.New("migration failed") },
		}

		oldNew := newPostgresClient
		newPostgresClient = func(_ string) (storage.DBStorage, error) {
			return mockClient, nil
		}
		defer func() { newPostgresClient = oldNew }()

		app := &App{config: &config.Config{DatabaseDSN: "mock-dsn"}}
		st, err := app.initStorage(ctx)
		require.Error(t, err)
		require.Nil(t, st)
	})

	t.Run("postgres client constructor error", func(t *testing.T) {
		oldNew := newPostgresClient
		newPostgresClient = func(_ string) (storage.DBStorage, error) {
			return nil, errors.New("cannot connect")
		}
		defer func() { newPostgresClient = oldNew }()

		app := &App{config: &config.Config{DatabaseDSN: "mock-dsn"}}
		st, err := app.initStorage(ctx)
		require.Error(t, err)
		require.Nil(t, st)
	})
}
