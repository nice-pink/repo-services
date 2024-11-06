package util

import (
	"testing"

	"github.com/nice-pink/repo-services/pkg/models"
)

func TestLogPrefix(t *testing.T) {
	app := models.App{Name: "app", Env: "env"}
	want := "app(env):"
	got := LogPrefix(app)
	if want != got {
		t.Error("LogPrefix(): got:", got, "!= want:", want)
	}
}

func TestGetAppDescription(t *testing.T) {
	app := models.App{Name: "app", Env: "env", Namespace: "namespace", Tag: "tag", File: "file", Image: "image"}
	want := "namespace/app(env):tag Path: file, Image: image"
	got := GetAppDescription(app)
	if want != got {
		t.Error("GetAppDescription(): got:", got, "!= want:", want)
	}
}
