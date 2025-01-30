package storage

import (
	"fmt"
	"sync"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getKey(t *testing.T) {
	type args struct {
		metricType metrics.MetricType
		metricName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Test1", args: args{metricType: metrics.MetricTypeCounter, metricName: "counter1"}, want: "counter|counter1"},
		{name: "Test2", args: args{metricType: metrics.MetricTypeGauge, metricName: "gauge1"}, want: "gauge|gauge1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equal(t, tt.want, getKey(tt.args.metricType, tt.args.metricName))

			// if got := getKey(tt.args.metricType, tt.args.metricName); got != tt.want {
			// 	t.Errorf("getKey() = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestMemStorage_Retrieve(t *testing.T) {

	metric1 := &metrics.Counter{Name: "counter1", Value: 1}
	metric2 := &metrics.Gauge{Name: "gauge1", Value: 3.14}

	s := &MemStorage{
		Data: map[string]metrics.Metric{
			"counter|counter1": metric1,
			"gauge|gauge1":     metric2,
		},
		mu: sync.Mutex{},
	}

	type args struct {
		metricType metrics.MetricType
		metricName string
	}
	tests := []struct {
		name    string
		args    args
		want    metrics.Metric
		wantErr bool
	}{
		{name: "Retrieve Existing Metric (counter)", args: args{metricType: metrics.MetricTypeCounter, metricName: metric1.Name}, want: metric1, wantErr: false},
		{name: "Retrieve Existing Metric (gauge)", args: args{metricType: metrics.MetricTypeGauge, metricName: metric2.Name}, want: metric2, wantErr: false},
		{name: "Retrieve Non-Existing Metric", args: args{metricType: metrics.MetricTypeGauge, metricName: "unknown"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.Retrieve(tt.args.metricType, tt.args.metricName)

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

	metric1 := &metrics.Counter{Name: "counter1", Value: 1}
	metric2 := &metrics.Gauge{Name: "gauge1", Value: 3.14}

	s := &MemStorage{
		Data: map[string]metrics.Metric{
			"counter|counter1": metric1,
			"gauge|gauge1":     metric2,
		},
		mu: sync.Mutex{},
	}

	tests := []struct {
		name    string
		want    []metrics.Metric
		wantErr bool
	}{
		{name: "Get all metrics", want: []metrics.Metric{metric1, metric2}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.RetrieveAll()

			assert.NoError(t, err)
			assert.Equal(t, len(got), len(s.Data))
			assert.Equal(t, got, tt.want)

		})
	}
}

func TestMemStorage_Add(t *testing.T) {

	metric1 := &metrics.Counter{Name: "counter1", Value: 1}
	metric2 := &metrics.Gauge{Name: "gauge1", Value: 3.14}

	type args struct {
		metric metrics.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     string
	}{
		{name: "Add second metric", args: args{metric: metric2}, wantErr: false},
		{name: "Add same metric, should be an error", args: args{metric: metric1}, wantErr: true, err: MetricAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &MemStorage{
				Data: map[string]metrics.Metric{
					"counter|counter1": metric1,
				},
				mu: sync.Mutex{},
			}

			err := s.Add(tt.args.metric)
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

	mcb := &metrics.Counter{Name: "counter1", Value: 1}
	mgb := &metrics.Gauge{Name: "gauge1", Value: 1}

	s := &MemStorage{
		Data: map[string]metrics.Metric{
			fmt.Sprintf("%s|%s", mcb.GetType(), mcb.GetName()): mcb,
			fmt.Sprintf("%s|%s", mgb.GetType(), mgb.GetName()): mgb,
		},
		mu: sync.Mutex{},
	}

	type args struct {
		metric metrics.Metric
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
			err := s.Update(tt.args.metric, tt.args.value)
			require.NoError(t, err)
			key := getKey(tt.args.metric.GetType(), tt.args.metric.GetName())
			assert.Equal(t, tt.wantValue, s.Data[key].GetValue())
		})
	}
}
