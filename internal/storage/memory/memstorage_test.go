package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getKey(t *testing.T) {
	type args struct {
		metricType metric.MetricType
		metricName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Test1", args: args{metricType: metric.MetricTypeCounter, metricName: "counter1"}, want: "counter|counter1"},
		{name: "Test2", args: args{metricType: metric.MetricTypeGauge, metricName: "gauge1"}, want: "gauge|gauge1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getKey(tt.args.metricType, tt.args.metricName))
		})
	}
}

func TestMemStorage_Retrieve(t *testing.T) {

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 3.14}
	ctx := context.Background()

	s := &MemStorage{
		Data: map[string]metric.Metric{
			"counter|counter1": metric1,
			"gauge|gauge1":     metric2,
		},
		mu: sync.Mutex{},
	}

	type args struct {
		metricType metric.MetricType
		metricName string
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{name: "Retrieve Existing Metric (counter)", args: args{metricType: metric.MetricTypeCounter, metricName: metric1.Name}, want: metric1, wantErr: false},
		{name: "Retrieve Existing Metric (gauge)", args: args{metricType: metric.MetricTypeGauge, metricName: metric2.Name}, want: metric2, wantErr: false},
		{name: "Retrieve Non-Existing Metric", args: args{metricType: metric.MetricTypeGauge, metricName: "unknown"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.Retrieve(ctx, tt.args.metricType, tt.args.metricName)

			if !tt.wantErr {
				assert.NoError(t, err, "Expected no error for existing metric")
				assert.Equal(t, tt.want, got, "Retrieved metric should match the stored value")
			} else {
				assert.Error(t, err, "Expected an error for non-existing metric")
			}
		})
	}
}

func TestMemStorage_RetrieveAll(t *testing.T) {

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 3.14}
	ctx := context.Background()

	s := &MemStorage{
		Data: map[string]metric.Metric{
			"counter|counter1": metric1,
			"gauge|gauge1":     metric2,
		},
		mu: sync.Mutex{},
	}

	tests := []struct {
		name    string
		want    []metric.Metric
		wantErr bool
	}{
		{name: "Get all metrics", want: []metric.Metric{metric1, metric2}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.RetrieveAll(ctx)
			assert.NoError(t, err)

			for _, m := range got {
				found := false
				for _, mSource := range s.Data {
					if m.GetName() == mSource.GetName() && m.GetType() == mSource.GetType() {
						assert.Equal(t, m.GetValue(), mSource.GetValue())
						found = true
					}
				}
				assert.True(t, found)
			}

		})
	}
}

func TestMemStorage_Add(t *testing.T) {

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 3.14}
	ctx := context.Background()

	type args struct {
		metric metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     string
	}{
		{name: "Add second metric", args: args{metric: metric2}, wantErr: false},
		{name: "Add same metric, should be an error", args: args{metric: metric1}, wantErr: true, err: common.ErrorMetricAlreadyExists.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &MemStorage{
				Data: map[string]metric.Metric{
					"counter|counter1": metric1,
				},
				mu: sync.Mutex{},
			}

			err := s.Add(ctx, tt.args.metric)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				key := getKey(tt.args.metric.GetType(), tt.args.metric.GetName())
				assert.Equal(t, tt.args.metric, s.Data[key])
			}

		})
	}
}

func TestMemStorage_Update(t *testing.T) {

	ctx := context.Background()

	mcb := &metric.Counter{Name: "counter1", Value: 1}
	mgb := &metric.Gauge{Name: "gauge1", Value: 1}

	s := &MemStorage{
		Data: map[string]metric.Metric{
			fmt.Sprintf("%s|%s", mcb.GetType(), mcb.GetName()): mcb,
			fmt.Sprintf("%s|%s", mgb.GetType(), mgb.GetName()): mgb,
		},
		mu: sync.Mutex{},
	}

	type args struct {
		metric metric.Metric
		value  interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantValue interface{}
	}{
		{"Test Counter update", args{mcb, int64(1)}, int64(2)},
		{"Test Gauge update", args{mgb, float64(1)}, float64(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Update(ctx, tt.args.metric, tt.args.value)
			require.NoError(t, err)
			key := getKey(tt.args.metric.GetType(), tt.args.metric.GetName())
			assert.Equal(t, tt.wantValue, s.Data[key].GetValue())
		})
	}
}

func BenchmarkMemStorage_Add(b *testing.B) {
	stor := NewMemStorage()
	ctx := context.Background()

	// creating metrics before Reset timer
	metrics := make([]metric.Metric, b.N)
	for i := 0; i < b.N; i++ {
		m := metric.NewCounter(fmt.Sprintf("counter%d", i))
		m.Value = int64(i)
		metrics[i] = m
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stor.Add(ctx, metrics[i])
	}
}

func BenchmarkMemStorage_Update(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	m := metric.MustNewCounter("counter1", 1)
	_ = store.Add(ctx, m)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Update(ctx, m, int64(i))
	}
}

func BenchmarkStore_GetMetric(b *testing.B) {
	stor := NewMemStorage()
	ctx := context.Background()

	m := metric.NewCounter("counter1")
	m.Value = 123

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stor.Retrieve(ctx, metric.MetricTypeCounter, "counter1")
	}
}

func BenchmarkMemStorage_RetrieveAll(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		_ = store.Add(ctx, metric.MustNewCounter(fmt.Sprintf("counter%d", i), int64(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.RetrieveAll(ctx)
	}
}

func BenchmarkMemStorage_UpdateBatch(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	metrics := make([]metric.Metric, 100)
	for i := 0; i < 100; i++ {
		m := metric.MustNewGauge(fmt.Sprintf("gauge%d", i), float64(i))
		_ = store.Add(ctx, m)
		metrics[i] = m
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.UpdateBatch(ctx, &metrics)
	}
}

func BenchmarkMemStorage_ConcurrentRetrieve(b *testing.B) {
	store := NewMemStorage()
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		_ = store.Add(ctx, metric.MustNewCounter(fmt.Sprintf("counter%d", i), int64(i)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Retrieve(ctx, metric.MetricTypeCounter, "counter500")
		}
	})
}

func BenchmarkMemStorage_ConcurrentAdd(b *testing.B) {
	stor := NewMemStorage()
	ctx := context.Background()

	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddInt64(&counter, 1)
			m := metric.MustNewCounter(fmt.Sprintf("counter%d", i), i)
			_ = stor.Add(ctx, m)
		}
	})
}
