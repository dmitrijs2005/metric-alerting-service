package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndRestoreDump(t *testing.T) {
	ctx := context.Background()
	tmpFile := "test_metrics_dmp.txt"
	defer os.Remove(tmpFile)

	stor := memory.NewMemStorage()

	metric1 := &metric.Counter{Name: "counter1", Value: int64(123)}
	metric2 := &metric.Gauge{Name: "gauge1", Value: float64(1.234)}

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	fs := NewFileSaver(tmpFile, stor)

	// save dump
	err := fs.SaveDump(ctx)
	require.NoError(t, err, "SaveDump failed")

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err, "failed to read file")

	content := string(data)
	assert.Contains(t, content, "counter1:counter:123")
	assert.Contains(t, content, "gauge1:gauge:1.234")

	// restore dump
	stor2 := memory.NewMemStorage()
	fs2 := NewFileSaver(tmpFile, stor2)

	err = fs2.RestoreDump(ctx)
	require.NoError(t, err, "RestoreDump failed")

	assert.Len(t, stor2.Data, len(stor.Data), "expected 2 added metrics")

	m, err := stor2.Retrieve(ctx, metric.MetricTypeCounter, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, m.GetValue(), int64(123))

	m2, err := stor2.Retrieve(ctx, metric.MetricTypeGauge, "gauge1")
	assert.NoError(t, err)
	assert.Equal(t, m2.GetValue(), float64(1.234))
}

func BenchmarkFileSaver_SaveDump(b *testing.B) {
	ctx := context.Background()
	tmpFile := "test_save.txt"
	defer os.Remove(tmpFile)

	stor := memory.NewMemStorage()

	for i := 0; i < 1000; i++ {
		stor.Add(ctx, metric.MustNewCounter(fmt.Sprintf("counter%d", i), int64(i)))
	}

	fs := NewFileSaver(tmpFile, stor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.SaveDump(ctx)
	}
}

func BenchmarkFileSaver_RestoreDump(b *testing.B) {
	ctx := context.Background()
	tmpFile := "test_restore.txt"
	defer os.Remove(tmpFile)

	// preparing data file
	f, _ := os.Create(tmpFile)
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(f, "m%d:counter:%d\n", i, i)
	}
	f.Close()

	stor := memory.NewMemStorage()
	fs := NewFileSaver(tmpFile, stor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.RestoreDump(ctx)
	}
}

type faultyStorage struct{}

func (f faultyStorage) Add(ctx context.Context, m metric.Metric) error {
	return errors.New("forced error in Add")
}
func (f faultyStorage) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return errors.New("forced error in Update")
}
func (f faultyStorage) Retrieve(ctx context.Context, t metric.MetricType, n string) (metric.Metric, error) {
	return nil, errors.New("forced error in Retrieve")
}
func (f faultyStorage) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, errors.New("forced error in RetrieveAll")
}
func (f faultyStorage) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return errors.New("forced error in UpdateBatch")
}

func TestFileSaver_SaveDump_RetrieveAllError(t *testing.T) {
	fs := &FileSaver{
		FileStoragePath: "ignored",
		Storage:         &faultyStorage{},
	}

	err := fs.SaveDump(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "forced error in RetrieveAll")
}

func TestFileSaver_SaveDump_FileOpenError(t *testing.T) {
	tmp := t.TempDir()
	path := tmp

	fs := &FileSaver{
		FileStoragePath: path,
		Storage:         memory.NewMemStorage(),
	}

	err := fs.SaveDump(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "error opening file")
}

func TestFileSaver_RestoreDump_InvalidLine(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/dump.txt"

	require.NoError(t, os.WriteFile(path, []byte("badline\n"), 0644))

	fs := &FileSaver{FileStoragePath: path, Storage: memory.NewMemStorage()}
	err := fs.RestoreDump(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid dump line")
}

func TestFileSaver_RestoreDump_NewMetricError(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/dump.txt"

	require.NoError(t, os.WriteFile(path, []byte("name:badtype:123\n"), 0644))

	fs := &FileSaver{FileStoragePath: path, Storage: memory.NewMemStorage()}
	err := fs.RestoreDump(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid metric type")
}

type brokenFile struct{}

func (b *brokenFile) Read(p []byte) (int, error) { return 0, errors.New("read fail") }
func (b *brokenFile) Close() error               { return nil }

func TestFileSaver_RestoreDump_ScannerError(t *testing.T) {
	orig := openFile
	openFile = func(string) (*os.File, error) {
		_ = &brokenFile{}
		return (*os.File)(nil), errors.New("scanner fail")
	}
	defer func() { openFile = orig }()

	fs := &FileSaver{FileStoragePath: "ignored", Storage: memory.NewMemStorage()}
	err := fs.RestoreDump(context.Background())
	require.Error(t, err)
}

func TestFileSaver_SaveDump_WriteError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "dump.txt")

	mockStorage := memory.NewMemStorage()
	mockStorage.Data[""] = &metric.Gauge{Name: "cpu", Value: float64(1.234)}

	orig := openFileWriter
	openFileWriter = func(string, int, os.FileMode) (*os.File, error) {
		return os.NewFile(0, ""), nil
	}
	defer func() { openFileWriter = orig }()

	fs := &FileSaver{FileStoragePath: path, Storage: mockStorage}
	err := fs.SaveDump(context.Background())
	require.Error(t, err)
}
