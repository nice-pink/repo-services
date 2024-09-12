package util

import "flag"

type GeneralFlags struct {
	Help                *bool
	App                 *string
	Namespace           *string
	Env                 *string
	Base                *string
	Cluster             *string
	Image               *string
	ImageFile           *string
	ImageHistoryFile    *string
	PathScheme          *string
	ExceptionalAppsFile *string
	SrcPath             *string
}

func GetGeneralFlags() GeneralFlags {
	return GeneralFlags{
		Help:                flag.Bool("help", false, ""),
		App:                 flag.String("app", "", "App name."),
		Namespace:           flag.String("namespace", DS_NAMESPACE, "Namespace."),
		Env:                 flag.String("env", DS_ENV, "App environment."),
		Base:                flag.String("base", DS_BASE, "Base path for apps."),
		Cluster:             flag.String("cluster", DS_BASE, "Cluster name of app, if needed for pathScheme."),
		Image:               flag.String("image", "", "Image name. [default: *app]"),
		ImageFile:           flag.String("imageFileName", DS_IMAGE_FILE_NAME, "Name of file which contain the container image and version tag. [default: deployment.yaml]"),
		ImageHistoryFile:    flag.String("imageHistoryFileName", DS_IMAGE_HISTORY_FILE_NAME, "Name of file which contain the container image history."),
		PathScheme:          flag.String("pathScheme", DS_PATH_SCHEME, "Scheme for apps paths. [default: {base}/{namespace}/{app}/{env}]"),
		ExceptionalAppsFile: flag.String("exceptionalAppsFile", DS_EXCEPTIONAL_APPS_FILE, "Filepath to file specifying exceptional apps. E.g. imageName != appName; path exceptional; etc."),
		SrcPath:             flag.String("srcPath", DS_SRC_PATH, "Source folder (e.g. of repo)."),
	}
}