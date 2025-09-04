package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
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

func TestHTTPServer_addNewMetric(t *testing.T) {

	s := prepareTestStorage()
	ctx := context.Background()

	type args struct {
		metricValue any
		metricType  string
		metricName  string
	}
	tests := []struct {
		args    args
		want    metric.Metric
		name    string
		wantErr bool
	}{
		{name: "Counter OK", args: args{metricType: "counter", metricName: "c2", metricValue: int64(1)}, wantErr: false, want: &metric.Counter{Name: "c2", Value: int64(1)}},
		{name: "Error", args: args{metricType: "unknown", metricName: "c2", metricValue: int64(1)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddNewMetric(ctx, s, tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.addNewMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.addNewMetric() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr {

				m, err := RetrieveMetric(ctx, s, tt.args.metricType, tt.args.metricName)
				if err != nil {
					t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(m, tt.want) {
					t.Errorf("HTTPServer.addNewMetric() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestHTTPServer_newMetricWithValue(t *testing.T) {
	type args struct {
		metricValue any
		metricType  string
		metricName  string
	}
	tests := []struct {
		args    args
		want    metric.Metric
		name    string
		wantErr bool
	}{
		{name: "Counter", args: args{metricType: "counter", metricName: "c1", metricValue: int64(1)}, wantErr: false, want: &metric.Counter{Name: "c1", Value: int64(1)}},
		{name: "Gauge", args: args{metricType: "gauge", metricName: "g1", metricValue: float64(1.234)}, wantErr: false, want: &metric.Gauge{Name: "g1", Value: float64(1.234)}},
		{name: "Gauge", args: args{metricType: "unknown", metricName: "g1", metricValue: float64(1.234)}, wantErr: true, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetricWithValue(tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.newMetricWithValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.newMetricWithValue() = %v, want %v", got, tt.want)
			}
		})
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

func TestHTTPServer_updateMetricByValue(t *testing.T) {

	s := prepareTestStorage()
	ctx := context.Background()
	type args struct {
		metricValue interface{}
		metricType  string
		metricName  string
	}
	tests := []struct {
		storage storage.Storage
		args    args
		want    metric.Metric
		name    string
		wantErr bool
	}{
		{name: "Counter1", storage: s, args: args{metricType: "counter", metricName: "c1", metricValue: int64(1)}, wantErr: false, want: &metric.Counter{Name: "c1", Value: int64(1)}},
		{name: "Counter2", storage: s, args: args{metricType: "counter", metricName: "c1", metricValue: int64(1)}, wantErr: false, want: &metric.Counter{Name: "c1", Value: int64(2)}},
		{name: "Error1", storage: s, args: args{metricType: "counter", metricName: "c1", metricValue: "wrongvalue"}, wantErr: true, want: &metric.Counter{}},
		{name: "Error2", storage: s, args: args{metricType: "counter", metricName: "x1", metricValue: "wrongvalue"}, wantErr: true, want: &metric.Counter{}},
		{name: "Error3", storage: faultyStorage{}, args: args{metricType: "counter", metricName: "x1", metricValue: "wrongvalue"}, wantErr: true, want: &metric.Counter{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateMetricByValue(ctx, tt.storage, tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.updateMetricByValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HTTPServer.updateMetricByValue() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

type UnknownMetric struct {
}

func (m *UnknownMetric) GetType() metric.MetricType {
	return metric.MetricType("unknown")
}
func (m *UnknownMetric) GetName() string {
	return "unknown"
}
func (m *UnknownMetric) GetValue() interface{} {
	return "unknown"
}
func (m *UnknownMetric) Update(interface{}) error {
	return nil
}

func TestHTTPServer_fillValue(t *testing.T) {
	type args struct {
		m metric.Metric
		r *dto.Metrics
	}
	tests := []struct {
		args    args
		name    string
		wantErr bool
	}{
		{name: "OK", args: args{m: &metric.Counter{Name: "c1", Value: int64(1)}, r: &dto.Metrics{}}, wantErr: false},
		{name: "Error", args: args{m: &metric.Gauge{Name: "g1", Value: float64(1.234)}, r: &dto.Metrics{}}, wantErr: false},
		{name: "Error", args: args{m: &UnknownMetric{}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FillValue(tt.args.m, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.fillValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
