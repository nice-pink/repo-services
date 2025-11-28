package util

import (
	"github.com/nice-pink/repo-services/pkg/models"
)

func LogPrefix(app models.App) string {
	return app.Name + "(" + app.Env + "):"
}

func GetAppDescription(app models.App) string {
	output := ""
	if app.Namespace != "" {
		output += app.Namespace + "/"
	}
	output += app.Name + "(" + app.Env + "):" + app.Tag + " Path: " + app.File + ", Image: " + app.Image
	return output
}
