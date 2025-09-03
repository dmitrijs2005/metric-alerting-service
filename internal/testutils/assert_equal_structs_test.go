package testutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type sample struct {
	ID   int
	Name string
}

func TestAssertEqualStructs_Success(t *testing.T) {
	got := sample{ID: 1, Name: "foo"}
	want := sample{ID: 1, Name: "foo"}

	AssertEqualStructs(t, got, want)
	if t.Failed() {
		t.Errorf("expected test to pass, but it failed")
	}
}

func TestCmpDiff_Fail(t *testing.T) {
	got := sample{ID: 1, Name: "foo"}
	want := sample{ID: 2, Name: "bar"}

	if diff := cmp.Diff(got, want); diff == "" {
		t.Errorf("expected diff but got none")
	} else {
		t.Logf("diff:\n%s", diff)
	}
}
