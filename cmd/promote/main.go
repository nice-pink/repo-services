package main

import (
	"flag"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/repo-services/pkg/runner"
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
		PrintExamples()
		os.Exit(0)
	}

	log.Info("*** Start")
	log.Info(os.Args)

	if err := runner.Promote(*srcEnv, *flags.ExceptionalAppsFile, flags, gitFlags); err != nil {
		os.Exit(2)
	}
}

func PrintExamples() {
	log.Info()
	log.Info("--- Examples:")
	log.Info("bin/promote -app test-app -namespace test -srcPath examples/repo -base base/resources -srcEnv dev")
}
