package staticlint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPublicAnalyzers(t *testing.T) {
	analyzers := GetPublicAnalyzers()

	require.Len(t, analyzers, 3, "expected exactly 3 analyzers")

	names := []string{
		analyzers[0].Name, // bodyclose
		analyzers[1].Name, // nakedret
		analyzers[2].Name, // ineffassign
	}

	require.Equal(t, "bodyclose", names[0], "first analyzer should be bodyclose")
	require.Equal(t, "nakedret", names[1], "second analyzer should be nakedret")
	require.Equal(t, "ineffassign", names[2], "third analyzer should be ineffassign")
}
