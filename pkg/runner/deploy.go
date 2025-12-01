package runner

import (
	"errors"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/manifest"
	"github.com/nice-pink/repo-services/pkg/util"
)

func Deploy(tag, exceptionalAppsFile string, flags util.GeneralFlags, gitFlags util.GitFlags) error {
	if tag == "" {
		log.Error("no -tag")
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
	log.Info(util.GetAppDescription(app))
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
