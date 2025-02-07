package main

import (
	"flag"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/repo-services/pkg/util"
)

func main() {
	// parameters
	folder := flag.String("folder", "tmp", "folder")
	gitFlags := util.GetGitFlags()
	flag.Parse()

	log.Info("*** Start")
	log.Info(os.Args)

	util.GitClone(*gitFlags.Url, *folder, gitFlags)
}

func PrintExamples() {
	log.Info()
	log.Info("--- Examples:")
	log.Info("bin/deploy -app test-app -namespace test -srcPath examples/repo -base base/resources -env dev -tag abc")
}
