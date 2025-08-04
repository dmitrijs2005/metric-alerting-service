package buildinfo

import (
	"bytes"
	"testing"
)

func TestPrintBuildParam(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		value    string
		expected string
	}{
		{"WithValue", "Build version", "1.2.3", "Build version: 1.2.3\n"},
		{"WithoutValue", "Build version", "", "Build version: N/A\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			printBuildParam(&buf, tt.param, tt.value)

			if got := buf.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
