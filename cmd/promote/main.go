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
	var appName = flag.String("app", "", "App name.")
	var namespace = flag.String("namespace", util.DS_NAMESPACE, "Namespace.")
	var env = flag.String("env", util.DS_ENV, "App environment.")
	var srcEnv = flag.String("srcEnv", util.DS_SRC_ENV, "Src environment for promotion.")
	var base = flag.String("base", util.DS_BASE, "Base path for apps.")
	var pathScheme = flag.String("pathScheme", util.DS_PATH_SCHEME, "Scheme for apps paths.")
	var image = flag.String("image", "", "Image name.")
	var imageFileName = flag.String("imageFileName", util.DS_IMAGE_FILE_NAME, "Name of file which contain the container image and version tag.")
	var imageHistoryFileName = flag.String("imageHistoryFileName", util.DS_IMAGE_HISTORY_FILE_NAME, "Name of file which contain the container image history.")
	var exceptionalAppsFile = flag.String("exceptionalAppsFile", util.DS_EXCEPTIONAL_APPS_FILE, "Filepath to file specifying exceptional apps. E.g. imageName != appName; path exceptional; etc.")
	var srcFolder = flag.String("srcFolder", util.DS_SRC_FOLDER, "Source folder (e.g. of repo).")
	flag.Parse()

	log.Info("*** Start")
	log.Info(os.Args)

	eh := exceptional.NewExceptionalHandler(*exceptionalAppsFile)
	handler := manifest.NewManifestHandler(eh)

	app := handler.BuildApp(*appName, *env, *namespace, *image, *pathScheme, *base, *imageFileName, *imageHistoryFileName, *srcFolder, "")
	handler.GetCurrentTag(app)

	src := handler.BuildApp(*appName, *srcEnv, *namespace, *image, *pathScheme, *base, *imageFileName, *imageHistoryFileName, *srcFolder, "")
	log.Info("Src app:", util.GetAppDescription(src))
	log.Info("Dest app:", util.GetAppDescription(app))
	handler.SetTagWithSource(src, app)
}
