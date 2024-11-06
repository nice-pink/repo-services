package util

import (
	"os"
	"slices"
	"testing"
)

// env

func TestGetEnvString(t *testing.T) {
	os.Clearenv()
	name := "BLABLABLA"
	want := "click"
	got := GetEnvString(name, "click")
	if got != want {
		t.Error("GetEnvString(): got:", got, "!= want:", want)
	}
	// set
	os.Setenv(name, "click")
	got = GetEnvString(name, "clack")
	if got != want {
		t.Error("GetEnvString(): got:", got, "!= want:", want)
	}
}

func TestGetEnvBool(t *testing.T) {
	os.Clearenv()
	name := "BLABLABLA"
	want := true
	got := GetEnvBool(name, true)
	if got != want {
		t.Error("GetEnvBool(): got:", got, "!= want:", want)
	}
	// set
	os.Setenv(name, "true")
	got = GetEnvBool(name, false)
	if got != want {
		t.Error("GetEnvBool(): got:", got, "!= want:", want)
	}
}

// array

func TestRemoveFromStringArray(t *testing.T) {
	arr := []string{"hello", "world"}
	want := []string{"hello"}
	got := RemoveFromStringArray(arr, 1)
	if !slices.Equal(got, want) {
		t.Error("RemoveFromStringArray(): got:", got, "!= want:", want)
	}
}

func TestRemoveFromStringArrayInvalidIndex(t *testing.T) {
	arr := []string{"hello", "world"}
	want := []string{"hello", "world"}
	got := RemoveFromStringArray(arr, 3)
	if !slices.Equal(got, want) {
		t.Error("RemoveFromStringArray(): got:", got, "!= want:", want)
	}
}

func TestRemoveFromStringArrayEmpty(t *testing.T) {
	arr := []string{}
	want := []string{}
	got := RemoveFromStringArray(arr, 0)
	if !slices.Equal(got, want) {
		t.Error("RemoveFromStringArray(): got:", got, "!= want:", want)
	}
}
