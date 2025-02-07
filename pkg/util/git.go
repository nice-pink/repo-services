package util

import (
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/goutil/pkg/repo"
)

func GitPush(repoPath, msg string, gitFlags GitFlags) error {
	if *gitFlags.Push {
		log.Info("Push to git.")
		repoHandle := repo.NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.User, *gitFlags.Email)
		if err := repoHandle.CommitPushLocalRepo(repoPath, msg, true); err != nil {
			return err
		}
	}
	return nil
}

func GitClone(url, baseFolder string, flags GitFlags) error {
	if url != "" {
		log.Info("Git clone.")
		repoHandle := repo.NewRepoHandle(*flags.SshKeyPath, *flags.User, *flags.Email)
		if err := repoHandle.Clone(url, baseFolder, *flags.Branch, *flags.Shallow, false); err != nil {
			return err
		}
	}
	return nil
}
