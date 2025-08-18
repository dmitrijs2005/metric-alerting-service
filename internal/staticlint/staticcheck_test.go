package staticlint

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func TestGetStaticCheckAnalyzers(t *testing.T) {
	analyzers := GetStaticCheckAnalyzers()

	require.NotEmpty(t, analyzers, "expected some analyzers to be returned")

	// Build a map to check for duplicates and lookup
	seen := make(map[string]*analysis.Analyzer)
	for _, a := range analyzers {
		if prev, ok := seen[a.Name]; ok {
			t.Fatalf("duplicate analyzer found: %s (%v vs %v)", a.Name, prev, a)
		}
		seen[a.Name] = a
	}

	// Spot-check required analyzers
	required := []string{"SA1000", "S1005", "ST1008", "QF1003"}
	for _, name := range required {
		if _, ok := seen[name]; !ok {
			t.Errorf("expected analyzer %s to be included", name)
		}
	}
}
