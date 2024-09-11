package util

import "strings"

func GetPathFromParameters(scheme, base, namespace, env, app string) string {
	path := strings.Replace(scheme, "{base}", base, 1)
	path = strings.Replace(path, "{namespace}", namespace, 1)
	path = strings.Replace(path, "{app}", app, 1)
	path = strings.Replace(path, "{env}", env, 1)
	return path
}
