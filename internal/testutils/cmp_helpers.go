// Package testutils provides helper functions for testing, such as struct comparison assertions.
package testutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertEqualStructs(t *testing.T, got, want interface{}) {
	t.Helper()
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Struct mismatch (-got +want):\n%s", diff)
	}
}
