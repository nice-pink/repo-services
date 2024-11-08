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
	var tag = flag.String("tag", "", "Image tag to set.")
	gitFlags := util.GetGitFlags()
	flag.Parse()

	if *flags.Help {
		flag.Usage()
		PrintExamples()
		os.Exit(0)
	}

	log.Info("*** Start")
	log.Info(os.Args)

	if *tag == "" {
		log.Error("no -tag")
		PrintExamples()
		os.Exit(2)
	}

	if err := runner.Deploy(*tag, *flags.ExceptionalAppsFile, flags, gitFlags); err != nil {
		os.Exit(2)
	}
}

func PrintExamples() {
	log.Info()
	log.Info("--- Examples:")
	log.Info("bin/deploy -app test-app -namespace test -srcPath examples/repo -base base/resources -env dev -tag abc")
}
