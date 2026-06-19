package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing/object"
)

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
