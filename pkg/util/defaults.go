package util

// audio

var (
	DS_NAMESPACE       string = GetEnvString("DS_NAMESPACE", "")
	DS_ENV             string = GetEnvString("DS_ENV", "")
	DS_BASE            string = GetEnvString("DS_BASE", "")
	DS_SRC_FOLDER      string = GetEnvString("DS_SRC_FOLDER", "")
	DS_PATH_SCHEME     string = GetEnvString("DS_PATH_SCHEME", "{base}/{namespace}/{app}/{env}")
	DS_IMAGE_FILE_NAME string = GetEnvString("DS_IMAGE_FILE_NAME", "deployment.yaml")
)
