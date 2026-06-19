package util

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

func GitPush(repoPath, msg string, gitFlags GitFlags) error {
	if *gitFlags.Push {
		slog.Default().Info("git_push", "repo", repoPath)
		repoHandle := NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.Token, *gitFlags.User, *gitFlags.Email)
		if err := repoHandle.CommitPushLocalRepo(repoPath, msg, true); err != nil {
			return err
		}
	}
	return nil
}

func GitClone(url, baseFolder string, flags GitFlags) error {
	if url != "" {
		slog.Default().Info("git_clone", "url", url)
		repoHandle := NewRepoHandle(*flags.SshKeyPath, *flags.Token, *flags.User, *flags.Email)
		if err := repoHandle.Clone(url, baseFolder, *flags.Branch, *flags.Shallow, false); err != nil {
			return err
		}
	}
	return nil
}

///////////////////

type RepoHandle struct {
	userName   string
	userEmail  string
	sshKeyPath string
	token      string
	repo       *git.Repository
}

func NewRepoHandle(sshKeyPath, token, userName, userEmail string) *RepoHandle {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &RepoHandle{
		userName:   userName,
		userEmail:  userEmail,
		sshKeyPath: sshKeyPath,
		token:      token,
	}
}

// Repo returns the underlying *git.Repository. May be nil before Open or Clone.
func (g *RepoHandle) Repo() *git.Repository { return g.repo }

// NOTE:
// For save usage do:
// - Clone() (shallow)
// - PullLocalRepo()
// - Do your changes to the repo
// - CommitPushLocalRepo()

// open

func (g *RepoHandle) Open(path string) error {
	var err error
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		slog.Default().Error("git_open", "err", err)
		return err
	}
	return nil
}

// Clone
func (g *RepoHandle) Clone(url string, dest string, branch string, shallow bool, repoSubfolder bool) error {
	// set clone options
	cloneOpt := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	// setup ssh auth
	if g.sshKeyPath != "" {
		auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
		if err != nil {
			slog.Default().Error("git_clone_ssh_key", "err", err)
			return err
		}
		cloneOpt.Auth = auth
	} else if g.token != "" {
		cloneOpt.Auth = &http.BasicAuth{
			Username: "github-actions",
			Password: g.token,
		}
	}

	if branch != "" {
		cloneOpt.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	if shallow {
		cloneOpt.Depth = 1
		cloneOpt.SingleBranch = true
		cloneOpt.ShallowSubmodules = true
	}

	// clone repo
	if repoSubfolder {
		path := strings.Split(url, "/")
		dest = filepath.Join(dest, path[len(path)-1])
	}

	var err error
	g.repo, err = git.PlainClone(dest, cloneOpt)
	if err != nil {
		slog.Default().Error("git_clone", "err", err)
		return err
	}

	// ... retrieving the branch being pointed by HEAD
	_, err = g.repo.Head()
	if err != nil {
		slog.Default().Error("git_head", "err", err)
	}
	return err
}

// Pull repo
func (g *RepoHandle) PullLocalRepo(path string) error {
	var err error
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		slog.Default().Error("git_open", "err", err)
		return err
	}

	workDir, err := g.repo.Worktree()
	if err != nil {
		slog.Default().Error("git_worktree", "err", err)
		return err
	}

	// fetch opt
	fetchOpt := &git.FetchOptions{
		Depth: 1,
		Force: true,
	}

	// pull opt
	pullOpt := &git.PullOptions{
		SingleBranch: true,
		Depth:        1,
		Force:        true,
	}

	// setup ssh auth
	if g.sshKeyPath != "" {
		auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
		if err != nil {
			slog.Default().Error("git_pull_ssh_key", "err", err)
			return err
		}
		pullOpt.Auth = auth
		fetchOpt.Auth = auth
	} else if g.token != "" {
		pullOpt.Auth = &http.BasicAuth{
			Username: "github-actions",
			Password: g.token,
		}
		fetchOpt.Auth = &http.BasicAuth{
			Username: "github-actions",
			Password: g.token,
		}
	}

	// Fetch remote — check return value; a network/auth error must not be silently swallowed.
	if err := g.repo.Fetch(fetchOpt); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	err = workDir.Pull(pullOpt)
	if err == git.NoErrAlreadyUpToDate {
		// do nothing
	} else if err != nil {
		slog.Default().Error("git_pull", "err", err)
	} else {
		slog.Default().Info("git_pulled")
	}

	return err
}

// Pull, commit and push repo.
func (g *RepoHandle) CommitPushLocalRepo(path string, message string, verbose bool) error {
	// Open folder as git repo
	var err error
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		slog.Default().Error("git_open", "err", err)
		return err
	}

	// Get worktree
	workDir, err := g.repo.Worktree()
	if err != nil {
		slog.Default().Error("git_worktree", "err", err)
		return err
	}

	// Get status
	status, err := workDir.Status()
	if err != nil {
		slog.Default().Error("git_status", "err", err)
		return err
	}
	if verbose {
		slog.Default().Info("git_status", "status", status.String())
	}

	// Add all files, with changes.
	for path := range status {
		slog.Default().Info("git_add", "path", path)
		workDir.Add(path)
	}

	// Commit changes
	commit, err := workDir.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.userName,
			Email: g.userEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		slog.Default().Error("git_commit", "err", err)
		return err
	}

	// Print commit object
	obj, err := g.repo.CommitObject(commit)
	if err != nil {
		slog.Default().Error("git_commit_object", "err", err)
		return err
	}
	slog.Default().Info("git_commit_object", "commit", obj.String())

	err = g.Push()
	if err != nil {
		return err
	}

	slog.Default().Info("git_push_ok")
	return nil
}

func (g *RepoHandle) Push() error {
	pushOpt := &git.PushOptions{}

	// setup ssh auth
	if g.sshKeyPath != "" {
		// Open key file for auth
		auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
		if err != nil {
			slog.Default().Error("git_push_ssh_key", "err", err)
			return err
		}
		pushOpt.Auth = auth
	} else if g.token != "" {
		pushOpt.Auth = &http.BasicAuth{
			Username: "github-actions",
			Password: g.token,
		}
	}

	// Push
	err := g.repo.Push(pushOpt)
	if err != nil {
		slog.Default().Error("git_push", "err", err)
	}
	return err
}

// Reset local repo to origin/main HEAD and clean unstaged files.
func (g *RepoHandle) ResetToRemoteHead(path string) error {
	// Open key file for auth
	auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
	if err != nil {
		slog.Default().Error("git_reset_ssh_key", "err", err)
		return err
	}

	// Open folder as git repo
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		slog.Default().Error("git_open", "err", err)
		return err
	}

	// Get worktree
	workDir, err := g.repo.Worktree()
	if err != nil {
		slog.Default().Error("git_worktree", "err", err)
		return err
	}

	// Fetch remote
	err = g.repo.Fetch(&git.FetchOptions{
		Auth: auth,
	})
	if err != nil {
		slog.Default().Error("git_fetch", "err", err)
	}

	// Get remote head
	remoteRef, err := g.repo.Reference(plumbing.Main, true)
	if err != nil {
		slog.Default().Error("git_reset_remote_head", "err", err)
		return err
	}

	// Reset
	workDir.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: remoteRef.Hash(),
	})

	// Clean git repo
	err = workDir.Clean(&git.CleanOptions{})
	if err != nil {
		slog.Default().Error("git_clean", "err", err)
	}

	return nil
}

