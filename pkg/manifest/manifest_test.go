package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nice-pink/repo-services/pkg/models"
)

// TestManifestAtomicWrite verifies that SetTagInFileWithPattern writes
// atomically — either the old content or the new content is visible,
// never a partial state. We verify correctness (not the mid-write atomicity
// guarantee, which requires SIGKILL fault injection) by checking the file
// contains the new tag after the call and no temp files are left behind.
func TestManifestAtomicWrite(t *testing.T) {
	tmp := t.TempDir()
	manifestFile := filepath.Join(tmp, "deployment.yaml")

	original := `apiVersion: apps/v1
kind: Deployment
spec:
  containers:
  - name: myapp
    image: myapp:v0.0.1
`
	if err := os.WriteFile(manifestFile, []byte(original), 0644); err != nil {
		t.Fatalf("write original: %v", err)
	}

	pattern := `(.*[/| ]myapp:)([a-zA-Z0-9_-].*)`
	newTag := "v0.0.2"
	currentTag := "v0.0.1"

	replaced, err := SetTagInFileWithPattern(newTag, currentTag, manifestFile, pattern)
	if err != nil {
		t.Fatalf("SetTagInFileWithPattern: %v", err)
	}
	if !replaced {
		t.Error("expected replaced=true")
	}

	// Verify new content
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		t.Fatalf("read after write: %v", err)
	}
	if !strings.Contains(string(content), "myapp:v0.0.2") {
		t.Errorf("new tag not in file:\n%s", content)
	}
	if strings.Contains(string(content), "myapp:v0.0.1") {
		t.Errorf("old tag still in file:\n%s", content)
	}

	// Verify no temp files remain
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

// TestManifestAtomicWrite_NoPatternMatch verifies that when the regex pattern
// does not match, replaced=false is returned.
// Note: the plain-text currentTag replacement still runs (preserving original behaviour),
// so the file may change if currentTag appears in the content. This test uses
// a file where neither the pattern NOR the currentTag appears.
func TestManifestAtomicWrite_NoPatternMatch(t *testing.T) {
	tmp := t.TempDir()
	manifestFile := filepath.Join(tmp, "deployment.yaml")

	// File has no occurrence of myapp: or oldtag
	original := `image: otherapp:v9.9`
	if err := os.WriteFile(manifestFile, []byte(original), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	pattern := `(.*[/| ]myapp:)([a-zA-Z0-9_-].*)`
	// currentTag is "oldtag" which does not appear in the file
	replaced, err := SetTagInFileWithPattern("v2.0", "oldtag", manifestFile, pattern)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replaced {
		t.Error("expected replaced=false when pattern doesn't match")
	}

	// File should be unchanged (pattern didn't match, currentTag not in file)
	content, _ := os.ReadFile(manifestFile)
	if string(content) != original {
		t.Errorf("file changed when it should not have:\n%s", content)
	}
}

// TestAddTagToHistory_ReturnTrue verifies the bug fix: AddTagToHistory now
// returns true on success (was previously returning err != nil = true on error).
func TestAddTagToHistory_ReturnTrue(t *testing.T) {
	tmp := t.TempDir()
	histFile := filepath.Join(tmp, "history.yaml")
	// Create an empty history file so AppendToFile succeeds.
	if err := os.WriteFile(histFile, []byte(""), 0644); err != nil {
		t.Fatalf("create history file: %v", err)
	}

	h := NewManifestHandler(nil)
	app := models.App{
		Name:    "myapp",
		Env:     "dev",
		Tag:     "v1.2.3",
		History: histFile,
	}

	result := h.AddTagToHistory(app)
	if !result {
		t.Error("AddTagToHistory: expected true on success, got false (inverted return bug?)")
	}

	// Verify the tag was actually appended.
	content, err := os.ReadFile(histFile)
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	if !strings.Contains(string(content), "v1.2.3") {
		t.Errorf("tag not appended to history: %q", content)
	}
}

// TestAddTagToHistory_NoHistoryFile verifies that AddTagToHistory returns false
// when no history file is configured.
func TestAddTagToHistory_NoHistoryFile(t *testing.T) {
	h := NewManifestHandler(nil)
	app := models.App{
		Name: "myapp",
		Env:  "dev",
		Tag:  "v1.0.0",
		// History is empty
	}
	result := h.AddTagToHistory(app)
	if result {
		t.Error("AddTagToHistory: expected false when History is empty, got true")
	}
}
