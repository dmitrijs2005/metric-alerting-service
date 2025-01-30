package storage

import (
	"reflect"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
	"github.com/stretchr/testify/assert"
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

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
