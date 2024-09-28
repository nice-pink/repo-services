package main

import (
	"flag"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/goutil/pkg/repo"
	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/manifest"
	"github.com/nice-pink/repo-services/pkg/util"
)

func main() {
	// parameters
	flags := util.GetGeneralFlags()
	var srcEnv = flag.String("srcEnv", util.DS_SRC_ENV, "Src environment for promotion. [default: staging]")
	gitFlags := util.GetGitFlags()
	flag.Parse()

	if *flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	log.Info("*** Start")
	log.Info(os.Args)

	eh := exceptional.NewExceptionalHandler(*flags.ExceptionalAppsFile)
	handler := manifest.NewManifestHandler(eh)

	// build dest app
	app := handler.BuildApp(flags, "")
	handler.GetCurrentTag(app)

	// build src app - overwrite env
	flags.Env = srcEnv
	src := handler.BuildApp(flags, "")

	// log apps and set tag with source
	log.Info("Src app:", util.GetAppDescription(src))
	log.Info("Dest app:", util.GetAppDescription(app))
	if !handler.SetTagWithSource(src, app) {
		os.Exit(2)
	}

	if *gitFlags.GitPush {
		log.Info("Push to git.")
		repoHandle := repo.NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.GitUser, *gitFlags.GitEmail)
		msg := "Promote " + app.Name + "(" + app.Env + ") version: " + src.Tag
		err := repoHandle.CommitPushLocalRepo(*flags.SrcPath, msg, true)
		if err != nil {
			os.Exit(2)
		}
	}
}
