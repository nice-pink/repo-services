package util

import "testing"

// TestNewFlagsFromValues verifies that NewFlagsFromValues builds correctly
// pointered structs from plain values and that Push/Url are settable.
func TestNewFlagsFromValues(t *testing.T) {
	v := FlagValues{
		App:                 "my-app",
		Namespace:           "default",
		Env:                 "dev",
		Base:                "base/resources",
		Cluster:             "cluster1",
		Image:               "my-image",
		ImageFile:           "deployment.yaml",
		ImageHistoryFile:    "history.yaml",
		PathScheme:          "{base}/{namespace}/{app}/{env}",
		ExceptionalAppsFile: "/path/exceptional.yaml",
		SrcPath:             "/ops-repo",
		VersionInfo:         "v1.2.3",
		Push:                false,
		Shallow:             false,
		SshKeyPath:          "/keys/id_rsa",
		User:                "mcp-server",
		Email:               "mcp@example.com",
		Url:                 "",
		Branch:              "main",
		Token:               "ghp_test",
	}

	gf, gitF := NewFlagsFromValues(v)

	// GeneralFlags checks
	if *gf.App != "my-app" {
		t.Errorf("App: got %q, want %q", *gf.App, "my-app")
	}
	if *gf.Namespace != "default" {
		t.Errorf("Namespace: got %q, want %q", *gf.Namespace, "default")
	}
	if *gf.Env != "dev" {
		t.Errorf("Env: got %q, want %q", *gf.Env, "dev")
	}
	if *gf.Base != "base/resources" {
		t.Errorf("Base: got %q, want %q", *gf.Base, "base/resources")
	}
	if *gf.SrcPath != "/ops-repo" {
		t.Errorf("SrcPath: got %q, want %q", *gf.SrcPath, "/ops-repo")
	}
	if *gf.PathScheme != "{base}/{namespace}/{app}/{env}" {
		t.Errorf("PathScheme: got %q", *gf.PathScheme)
	}
	if *gf.ImageFile != "deployment.yaml" {
		t.Errorf("ImageFile: got %q", *gf.ImageFile)
	}
	if gf.Help == nil {
		t.Error("Help should be a non-nil pointer")
	}
	if *gf.Help != false {
		t.Error("Help should be false")
	}

	// GitFlags checks — especially Push=false and Url="" (the MCP safety contract)
	if *gitF.Push != false {
		t.Error("Push must default to false")
	}
	if *gitF.Url != "" {
		t.Error("Url must default to empty string")
	}
	if *gitF.Token != "ghp_test" {
		t.Errorf("Token: got %q", *gitF.Token)
	}
	if *gitF.SshKeyPath != "/keys/id_rsa" {
		t.Errorf("SshKeyPath: got %q", *gitF.SshKeyPath)
	}
}

// TestNewFlagsFromValues_PushFalseHardcoded verifies the invariant that the
// MCP server can set Push=false and the pointer indirection is correct.
func TestNewFlagsFromValues_PushFalseHardcoded(t *testing.T) {
	v := FlagValues{
		App:        "app",
		Env:        "prod",
		Base:       "base",
		PathScheme: "{base}/{app}/{env}",
		ImageFile:  "deployment.yaml",
		Push:       false,
		Url:        "",
	}

	_, gitF := NewFlagsFromValues(v)

	// Dereference the pointer — must not panic and must be false.
	if gitF.Push == nil {
		t.Fatal("Push pointer is nil")
	}
	if *gitF.Push {
		t.Error("Push must be false")
	}
	if gitF.Url == nil {
		t.Fatal("Url pointer is nil")
	}
	if *gitF.Url != "" {
		t.Error("Url must be empty")
	}
}
