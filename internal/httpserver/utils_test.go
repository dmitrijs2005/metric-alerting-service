package httpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ContentTypeIsCompressable(tt.args.contentType))
		})
	}
}
