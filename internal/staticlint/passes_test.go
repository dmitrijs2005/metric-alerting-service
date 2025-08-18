package staticlint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPassesAnalyzers(t *testing.T) {
	analyzers := GetPassesAnalyzers()

	// Should have at least the known count (currently 49)
	require.GreaterOrEqual(t, len(analyzers), 45, "expected many analyzers, got fewer")

	// Ensure all analyzers are non-nil and have names
	for i, a := range analyzers {
		require.NotNilf(t, a, "analyzer at index %d should not be nil", i)
		require.NotEmptyf(t, a.Name, "analyzer at index %d should have a name", i)
	}

	// Spot check a few expected analyzers by name
	names := []string{analyzers[0].Name, analyzers[5].Name, analyzers[10].Name}
	require.Contains(t, names, "appends", "should include appends.Analyzer")
	require.Contains(t, names, "bools", "should include bools.Analyzer")
	require.Contains(t, names, "copylocks", "should include copylock.Analyzer")

	// Ensure the set contains some critical ones regardless of order
	found := map[string]bool{}
	for _, a := range analyzers {
		found[a.Name] = true
	}
	for _, expected := range []string{"printf", "shadow", "unusedresult", "waitgroup"} {
		require.Truef(t, found[expected], "expected analyzer %q to be present", expected)
	}
}
