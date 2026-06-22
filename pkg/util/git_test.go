package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// TestIsMissingObjectErr covers the retry trigger: PullLocalRepo escalates from
// shallow to full fetch only when this matcher returns true. It must catch both
// the typed plumbing.ErrObjectNotFound and string-wrapped variants (the SSH
// transport in production wraps the error as a plain message).
func TestIsMissingObjectErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"typed", plumbing.ErrObjectNotFound, true},
		{"wrapped-typed", fmt.Errorf("pull failed: %w", plumbing.ErrObjectNotFound), true},
		{"string-only", errors.New("object not found"), true},
		{"unrelated", errors.New("authentication required"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isMissingObjectErr(tc.err); got != tc.want {
				t.Errorf("isMissingObjectErr(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// makeInitialCommit creates a minimal git repo with one commit so it can be cloned.
func makeInitialCommit(t *testing.T, dir string) {
	t.Helper()
	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	stub := filepath.Join(dir, "README")
	if err := os.WriteFile(stub, []byte("init"), 0644); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	if _, err := wt.Add("README"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := wt.Commit("initial", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()},
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

// TestPullLocalRepo_FetchErrorPropagates verifies that a fetch error is NOT silently
// swallowed. We clone a local bare repo, then change the remote URL to something
// unreachable, and verify PullLocalRepo returns an error.
//
// This is the regression guard for the discarded g.repo.Fetch(fetchOpt) return value.
func TestPullLocalRepo_FetchErrorPropagates(t *testing.T) {
	// Create a bare repo with a commit (bare repo must be non-empty to allow clone)
	bareDir := t.TempDir()
	bareRepo, err := gogit.PlainInit(bareDir, true)
	if err != nil {
		t.Fatalf("init bare repo: %v", err)
	}
	_ = bareRepo

	// Create a non-bare source repo, commit to it, then push to bare
	srcDir := t.TempDir()
	srcRepo, err := gogit.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("init src repo: %v", err)
	}
	stub := filepath.Join(srcDir, "README")
	if err := os.WriteFile(stub, []byte("init"), 0644); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	wt, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add("README"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := wt.Commit("initial", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()},
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}
	// Add the bare repo as remote and push
	if _, err := srcRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file://" + bareDir},
	}); err != nil {
		t.Fatalf("create remote: %v", err)
	}
	if err := srcRepo.Push(&gogit.PushOptions{RemoteName: "origin"}); err != nil {
		t.Fatalf("push to bare: %v", err)
	}

	// Clone the bare repo to a working directory
	workDir := t.TempDir()
	_, err = gogit.PlainClone(workDir, &gogit.CloneOptions{
		URL: "file://" + bareDir,
	})
	if err != nil {
		t.Fatalf("clone: %v", err)
	}

	// Now point the remote to an unreachable URL so the fetch must fail
	workRepo, err := gogit.PlainOpen(workDir)
	if err != nil {
		t.Fatalf("open workdir: %v", err)
	}
	if err := workRepo.DeleteRemote("origin"); err != nil {
		t.Fatalf("delete remote: %v", err)
	}
	if _, err := workRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file:///nonexistent/path/that/does/not/exist"},
	}); err != nil {
		t.Fatalf("create bad remote: %v", err)
	}

	// PullLocalRepo should return an error from Fetch, not succeed silently.
	rh := NewRepoHandle("", "", "test", "test@example.com")
	pullErr := rh.PullLocalRepo(workDir)
	if pullErr == nil {
		t.Error("expected PullLocalRepo to return an error when fetch fails with unreachable remote; got nil")
	}
}

// TestPullLocalRepo_BehindByMultipleCommits is the end-to-end smoke test for
// the multi-commit-behind path. Production saw a "object not found" failure
// here against the GitHub SSH transport; against go-git's local file:// the
// shallow fetch may quietly send the full pack, so this test mainly guards
// against a regression that breaks pulls when N>1 commits behind. The error
// matcher that triggers the retry is covered separately in TestIsMissingObjectErr.
func TestPullLocalRepo_BehindByMultipleCommits(t *testing.T) {
	bareDir := t.TempDir()
	if _, err := gogit.PlainInit(bareDir, true); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	srcDir := t.TempDir()
	srcRepo, err := gogit.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("init src: %v", err)
	}
	srcWT, err := srcRepo.Worktree()
	if err != nil {
		t.Fatalf("src worktree: %v", err)
	}
	commitSrc := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		if _, err := srcWT.Add(name); err != nil {
			t.Fatalf("add %s: %v", name, err)
		}
		if _, err := srcWT.Commit("c-"+name, &gogit.CommitOptions{
			Author: &object.Signature{Name: "t", Email: "t@t.com", When: time.Now()},
		}); err != nil {
			t.Fatalf("commit %s: %v", name, err)
		}
	}

	commitSrc("A", "a")
	if _, err := srcRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file://" + bareDir},
	}); err != nil {
		t.Fatalf("create remote: %v", err)
	}
	if err := srcRepo.Push(&gogit.PushOptions{RemoteName: "origin"}); err != nil {
		t.Fatalf("push A: %v", err)
	}

	workDir := t.TempDir()
	if _, err := gogit.PlainClone(workDir, &gogit.CloneOptions{
		URL: "file://" + bareDir,
	}); err != nil {
		t.Fatalf("clone: %v", err)
	}

	// Advance origin by 2 commits — workDir is now 2 behind.
	commitSrc("B", "b")
	commitSrc("C", "c")
	if err := srcRepo.Push(&gogit.PushOptions{RemoteName: "origin"}); err != nil {
		t.Fatalf("push B,C: %v", err)
	}

	rh := NewRepoHandle("", "", "test", "test@example.com")
	if err := rh.PullLocalRepo(workDir); err != nil {
		t.Fatalf("PullLocalRepo: %v", err)
	}

	// Confirm the chain is intact: every file from origin landed AND fsck-like
	// walk from HEAD reaches the root without missing objects.
	for _, f := range []string{"A", "B", "C"} {
		if _, err := os.Stat(filepath.Join(workDir, f)); err != nil {
			t.Errorf("expected %s in workDir after pull: %v", f, err)
		}
	}
	workRepo, err := gogit.PlainOpen(workDir)
	if err != nil {
		t.Fatalf("open workdir: %v", err)
	}
	head, err := workRepo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	commits, err := workRepo.Log(&gogit.LogOptions{From: head.Hash()})
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	n := 0
	if err := commits.ForEach(func(*object.Commit) error { n++; return nil }); err != nil {
		t.Fatalf("walk: %v", err)
	}
	if n < 3 {
		t.Errorf("expected ancestry walk to reach >=3 commits (A,B,C); got %d", n)
	}
}

// TestRepoHandleRepoAccessor verifies the new Repo() accessor returns the
// underlying *git.Repository after Open.
func TestRepoHandleRepoAccessor(t *testing.T) {
	dir := t.TempDir()
	if _, err := gogit.PlainInit(dir, false); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Write a stub file
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("test"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	rh := NewRepoHandle("", "", "test", "test@example.com")
	if err := rh.Open(dir); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if rh.Repo() == nil {
		t.Error("Repo() should return non-nil after Open")
	}
}
