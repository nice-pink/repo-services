package main

import (
	"flag"
	"os"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/manifest"
	"github.com/nice-pink/repo-services/pkg/util"
)

func main() {
	// parameters
	flags := util.GetGeneralFlags()
	var tag = flag.String("tag", "", "Image tag to set.")
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
	handler.SetTag(app)
}
