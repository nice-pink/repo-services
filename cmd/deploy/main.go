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
	var tag = flag.String("tag", "", "Image tag to set.")
	gitFlags := util.GetGitFlags()
	flag.Parse()

	if *flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	log.Info("*** Start")
	log.Info(os.Args)

	if *tag == "" {
		log.Error("no -tag")
		os.Exit(2)
	}

	// exceptional apps handler
	eh := exceptional.NewExceptionalHandler(*flags.ExceptionalAppsFile)
	handler := manifest.NewManifestHandler(eh)

	// run
	app := handler.BuildApp(flags, *tag)
	log.Info(util.GetAppDescription(app))
	if !handler.SetTag(app) {
		os.Exit(2)
	}

	if *gitFlags.GitPush {
		log.Info("Push to git.")
		repoHandle := repo.NewRepoHandle(*gitFlags.SshKeyPath, *gitFlags.GitUser, *gitFlags.GitEmail)
		msg := "Deploy " + *flags.App + "(" + *flags.Env + ") version: " + *tag
		err := repoHandle.CommitPushLocalRepo(*flags.SrcPath, msg, true)
		if err != nil {
			os.Exit(2)
		}
	}
}

func PrintExamples() {
	log.Info()
	log.Info("--- Examples")
	log.Info("deploy app version: ./deploy -app processing-engine-streaming -namespace streaming -env prod -tag abcd")
}
