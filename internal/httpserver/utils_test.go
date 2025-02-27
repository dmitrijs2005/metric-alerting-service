package httpserver

import "testing"

func TestContentTypeIsCompressable(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"OK", args{"application/json"}, true},
		{"OK", args{"text/html whatever"}, true},
		{"Not OK", args{"some other"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentTypeIsCompressable(tt.args.contentType); got != tt.want {
				t.Errorf("ContentTypeIsCompressable() = %v, want %v", got, tt.want)
			}
		})
	}
}
