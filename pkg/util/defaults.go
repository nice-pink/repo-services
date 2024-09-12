package util

// audio

var (
	DS_NAMESPACE               string = GetEnvString("DS_NAMESPACE", "")
	DS_ENV                     string = GetEnvString("DS_ENV", "")
	DS_BASE                    string = GetEnvString("DS_BASE", "")
	DS_SRC_PATH                string = GetEnvString("DS_SRC_PATH", "")
	DS_SRC_ENV                 string = GetEnvString("DS_SRC_ENV", "staging")
	DS_PATH_SCHEME             string = GetEnvString("DS_PATH_SCHEME", "{base}/{namespace}/{app}/{env}")
	DS_EXCEPTIONAL_APPS_FILE   string = GetEnvString("DS_EXCEPTIONAL_APPS_FILE", "")
	DS_IMAGE_FILE_NAME         string = GetEnvString("DS_IMAGE_FILE_NAME", "deployment.yaml")
	DS_IMAGE_HISTORY_FILE_NAME string = GetEnvString("DS_IMAGE_HISTORY_FILE_NAME", "")
)
