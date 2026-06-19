package runner

import (
	"errors"
	"log/slog"

	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/manifest"
	"github.com/nice-pink/repo-services/pkg/util"
)

func Deploy(tag, exceptionalAppsFile string, flags util.GeneralFlags, gitFlags util.GitFlags) error {
	if tag == "" {
		slog.Default().Error("deploy_no_tag")
		return errors.New("no 'tag' defined")
	}

	if *gitFlags.Url != "" {
		err := util.GitClone(*gitFlags.Url, *flags.SrcPath, gitFlags)
		if err != nil {
			return err
		}
	}

	// exceptional apps handler
	eh := exceptional.NewExceptionalHandler(exceptionalAppsFile)
	handler := manifest.NewManifestHandler(eh)
	app := handler.BuildApp(flags, tag)

	// run
	slog.Default().Info("deploy_app", "app", util.GetAppDescription(app))
	if !handler.SetTag(app) {
		return errors.New("could not set tag")
	}

	if !*gitFlags.Push {
		return nil
	}

	// git
	msg := "Deploy " + *flags.App + "(" + *flags.Env + ") version: " + tag
	return util.GitPush(*flags.SrcPath, msg, gitFlags)
}
