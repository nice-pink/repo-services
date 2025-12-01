package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/nice-pink/goutil/pkg/log"
)

// import (
// 	"github.com/nice-pink/goutil/pkg/log"
// 	"github.com/nice-pink/goutil/pkg/repo"
// )

func GitPush(repoPath, msg string, gitFlags GitFlags) error {
	if *gitFlags.Push {
		log.Info("Push to git.")
		repoHandle := NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.Token, *gitFlags.User, *gitFlags.Email)
		if err := repoHandle.CommitPushLocalRepo(repoPath, msg, true); err != nil {
			return err
		}
	}
	return nil
}

func GitClone(url, baseFolder string, flags GitFlags) error {
	if url != "" {
		log.Info("Git clone.")
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
		log.Err(err, "open")
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
			log.Err(err, "key")
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
		log.Err(err, "clone")
		return err
	}

	// ... retrieving the branch being pointed by HEAD
	_, err = g.repo.Head()
	if err != nil {
		log.Err(err)
	}
	return err
}

// Pull repo
func (g *RepoHandle) PullLocalRepo(path string) error {
	var err error
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		log.Err(err, "open")
		return err
	}

	workDir, err := g.repo.Worktree()
	if err != nil {
		log.Err(err, "worktree")
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
			panic(err)
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

	// Fetch remote
	g.repo.Fetch(fetchOpt)

	err = workDir.Pull(pullOpt)
	if err == git.NoErrAlreadyUpToDate {
		// do nothing
	} else if err != nil {
		log.Err(err, "pull")
		// err = g.ResetToRemoteHead(path)
		// if err != nil {
		// 	log.Err(err)
		// } else {
		// 	err = workDir.Pull(pullOpt)
		// 	if err != nil {
		// 		log.Err(err)
		// 	}
		// }
	} else {
		log.Info("Pulled repo.")
	}

	return err
}

// Pull, commit and push repo.
func (g *RepoHandle) CommitPushLocalRepo(path string, message string, verbose bool) error {
	// Open folder as git repo
	var err error
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		log.Err(err, "open")
		return err
	}

	// Get worktree
	workDir, err := g.repo.Worktree()
	if err != nil {
		log.Err(err, "worktree")
		return err
	}
	// if verbose {
	// 	log.Info("workDir:")
	//  log.Info(workDir)
	// }

	// Pull remote
	// Note: The pull overwrites the current repo and undoes all changes!
	// if pull {
	// 	err = workDir.Pull(&git.PullOptions{
	// 		// RemoteName:   "origin",
	// 		SingleBranch: true,
	// 		Depth:        1,
	// 		Auth:         auth,
	// 		Force:        true,
	// 	})
	// 	if err == git.NoErrAlreadyUpToDate {
	// 		// do nothing
	// 	} else if err != nil {
	// 		log.Err(err, "pull")
	// 	}
	// 	if verbose {
	// 		log.Info("Pulled repo.")
	// 	}
	// }

	// Get status
	status, err := workDir.Status()
	if err != nil {
		log.Err(err, "status")
		return err
	}
	if verbose {
		log.Info("status:")
		log.Info(status.String())
	}

	// Add all files, with changes.
	for path := range status {
		fmt.Println("Added: " + path)
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
		log.Err(err, "commit")
		return err
	}

	// Print commit object
	obj, err := g.repo.CommitObject(commit)
	if err != nil {
		log.Err(err, "commit object")
		return err
	}
	fmt.Println(obj)

	err = g.Push()
	if err != nil {
		return err
	}

	// Success!
	fmt.Println("Success!")
	return nil
}

func (g *RepoHandle) Push() error {
	pushOpt := &git.PushOptions{}

	// setup ssh auth
	if g.sshKeyPath != "" {
		// Open key file for auth
		auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
		if err != nil {
			panic(err)
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
		log.Err(err, "push")
	}
	return err
}

// Reset local repo to origin/main HEAD and clean unstaged files.
func (g *RepoHandle) ResetToRemoteHead(path string) error {
	// Open key file for auth
	auth, err := ssh.NewPublicKeysFromFile("git", g.sshKeyPath, "")
	if err != nil {
		panic(err)
	}

	// Open folder as git repo
	g.repo, err = git.PlainOpen(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Get worktree
	workDir, err := g.repo.Worktree()
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Fetch remote
	err = g.repo.Fetch(&git.FetchOptions{
		Auth: auth,
	})
	if err != nil {
		log.Err(err, "fetch")
	}

	// Get remote head
	remoteRef, err := g.repo.Reference(plumbing.Main, true)
	if err != nil {
		fmt.Print("Could not get remote head.")
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
		fmt.Println(err)
	}

	return nil
}
