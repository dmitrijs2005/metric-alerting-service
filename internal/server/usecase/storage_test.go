package usecase

import (
	"context"
	"reflect"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
)

var (
	metric1 = &metric.Counter{Name: "counter1", Value: 1}
	metric2 = &metric.Gauge{Name: "gauge1", Value: 1.234}
)

func prepareTestStorage() storage.Storage {
	s := memory.NewMemStorage()

	s.Data["counter|counter1"] = metric1
	s.Data["gauge|gauge1"] = metric2

	return s
}

func TestHTTPServer_retrieveMetric(t *testing.T) {

	ctx := context.Background()

	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		args    args
		want    metric.Metric
		name    string
		wantErr bool
	}{
		{name: "Counter OK", args: args{string(metric.MetricTypeCounter), "counter1"}, want: &metric.Counter{Name: "counter1", Value: 1}},
		{name: "Gauge OK", args: args{string(metric.MetricTypeGauge), "gauge1"}, want: &metric.Gauge{Name: "gauge1", Value: 1.234}},
		{name: "Unknown", args: args{"unknown", "u1"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestStorage()

			got, err := RetrieveMetric(ctx, s, tt.args.metricType, tt.args.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.retrieveMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.retrieveMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_updateMetric(t *testing.T) {

	s := prepareTestStorage()

	// a := "http://localhost:8080"
	// s := memory.NewMemStorage()

	// m1 := &metric.Counter{Name: "counter1", Value: 1}
	// m2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	// s.Data["counter|counter1"] = m1
	// s.Data["gauge|gauge1"] = m2

	ctx := context.Background()

	type args struct {
		metricValue any
		m           metric.Metric
	}
	tests := []struct {
		wantValue any
		args      args
		name      string
		wantErr   bool
	}{
		{name: "Counter OK", args: args{m: metric1, metricValue: int64(2)}, wantErr: false, wantValue: int64(3)},
		{name: "Gauge OK", args: args{m: metric2, metricValue: float64(2.345)}, wantErr: false, wantValue: float64(2.345)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// s := &HTTPServer{
			// 	Address: a,
			// 	Storage: s,
			// }
			if err := UpdateMetric(ctx, s, tt.args.m, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			m, err := RetrieveMetric(ctx, s, string(tt.args.m.GetType()), tt.args.m.GetName())
			if err != nil {
				t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if m.GetValue() != tt.wantValue {
				t.Errorf("HTTPServer.updateMetric() error = wrong value, %v, wanted: %v ", m.GetValue(), tt.wantValue)
			}

		})
	}
}
