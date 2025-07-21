package file

import (
	"context"
	"fmt"
	"os"
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
