package runner

import (
	"errors"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/goutil/pkg/repo"
	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/manifest"
	"github.com/nice-pink/repo-services/pkg/util"
)

func Promote(srcEnv, exceptionalAppsFile string, flags util.GeneralFlags, gitFlags util.GitFlags) error {
	eh := exceptional.NewExceptionalHandler(exceptionalAppsFile)
	handler := manifest.NewManifestHandler(eh)

	// build dest app
	app := handler.BuildApp(flags, "")
	currentTag := handler.GetCurrentTag(app)
	if currentTag == "" {
		log.Error("Can't get current tag.")
		return errors.New("no current tag")
	}

	// build src app - overwrite env
	flags.Env = &srcEnv
	src := handler.BuildApp(flags, "")

	// log apps and set tag with source
	log.Info("Src app:", util.GetAppDescription(src))
	log.Info("Dest app:", util.GetAppDescription(app))
	if !handler.SetTagWithSource(src, app) {
		return errors.New("cannot set tag")
	}

	if *gitFlags.GitPush {
		log.Info("Push to git.")
		repoHandle := repo.NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.GitUser, *gitFlags.GitEmail)
		msg := "Promote " + app.Name + "(" + app.Env + ") version: " + src.Tag
		if err := repoHandle.CommitPushLocalRepo(*flags.SrcPath, msg, true); err != nil {
			return err
		}
	}
	return nil
}
