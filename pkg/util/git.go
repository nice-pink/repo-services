package util

import (
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/goutil/pkg/repo"
)

func GitPush(repoPath, msg string, gitFlags GitFlags) error {
	if *gitFlags.GitPush {
		log.Info("Push to git.")
		repoHandle := repo.NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.GitUser, *gitFlags.GitEmail)
		if err := repoHandle.CommitPushLocalRepo(repoPath, msg, true); err != nil {
			return err
		}
	}
	return nil
}
