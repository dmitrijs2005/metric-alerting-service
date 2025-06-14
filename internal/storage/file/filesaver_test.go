package file

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/assert"
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
	if err != nil {
		t.Fatalf("SaveDump failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "counter1:counter:123") || !strings.Contains(content, "gauge1:gauge:1.234") {
		t.Errorf("unexpected file content: %s", content)
	}

	// restore dump
	stor2 := memory.NewMemStorage()
	fs2 := NewFileSaver(tmpFile, stor2)

	err = fs2.RestoreDump(ctx)
	if err != nil {
		t.Fatalf("RestoreDump failed: %v", err)
	}

	if len(stor2.Data) != len(stor.Data) {
		t.Errorf("expected 2 added metrics, got %d", len(stor2.Data))
	}

	m, err := stor2.Retrieve(ctx, metric.MetricTypeCounter, "counter1")
	assert.NoError(t, err)
	assert.Equal(t, m.GetValue(), int64(123))

	m2, err := stor2.Retrieve(ctx, metric.MetricTypeGauge, "gauge1")
	assert.NoError(t, err)
	assert.Equal(t, m2.GetValue(), float64(1.234))
}
