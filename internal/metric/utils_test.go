package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMetricNameValid(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test1", args: args{n: "metric1"}, want: true},
		{name: "Test2", args: args{n: "metric1:counter"}, want: true},
		{name: "Test3", args: args{n: "metric1,counter"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMetricNameValid(tt.args.n)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMetric(t *testing.T) {

	c := &Counter{Name: "counter1", Value: 0}
	g := &Gauge{Name: "gauge1", Value: 0}

	type args struct {
		metricType MetricType
		metricName string
	}
	tests := []struct {
		args    args
		want    Metric
		err     error
		name    string
		wantErr bool
	}{
		{name: "Test Counter OK ", args: args{c.GetType(), c.GetName()}, want: c, wantErr: false, err: nil},
		{name: "Test Gauge OK ", args: args{g.GetType(), g.GetName()}, want: g, wantErr: false, err: nil},
		{name: "Test Error", args: args{"unknown", "unknown"}, want: nil, wantErr: true, err: ErrorInvalidMetricType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetric(tt.args.metricType, tt.args.metricName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

		})
	}
}
