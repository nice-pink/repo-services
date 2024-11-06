package util

import "testing"

func TestGetPathFromParameters(t *testing.T) {
	scheme := "{base}/{namespace}/{env}/{app}"
	want := "b/n/e/a"
	got := GetPathFromParameters(scheme, "b", "n", "e", "a")
	if got != want {
		t.Error("GetPathFromParameters(): got:", got, "!= want:", want)
	}
}

func TestGetPathFromParametersEmpty(t *testing.T) {
	scheme := ""
	want := ""
	got := GetPathFromParameters(scheme, "b", "n", "e", "a")
	if got != want {
		t.Error("GetPathFromParameters(): got:", got, "!= want:", want)
	}
}
