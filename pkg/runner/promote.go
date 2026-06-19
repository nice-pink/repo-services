package runner

import (
	"errors"
	"log/slog"

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
		slog.Default().Error("promote_no_current_tag")
		return errors.New("no current tag")
	}

	// build src app - overwrite env
	flags.Env = &srcEnv
	src := handler.BuildApp(flags, "")

	// log apps and set tag with source
	slog.Default().Info("promote_src", "app", util.GetAppDescription(src))
	slog.Default().Info("promote_dest", "app", util.GetAppDescription(app))
	if !handler.SetTagWithSource(src, app) {
		return errors.New("cannot set tag")
	}

	// git
	msg := "Promote " + app.Name + "(" + app.Env + ") version: " + src.Tag
	return util.GitPush(*flags.SrcPath, msg, gitFlags)
}
