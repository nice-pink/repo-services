package exceptional

import (
	"os"

	"github.com/nice-pink/repo-services/pkg/models"
	yaml "gopkg.in/yaml.v3"
)

type ExceptionalHandler struct {
	Filepath string
}

func NewExceptionalHandler(filepath string) *ExceptionalHandler {
	return &ExceptionalHandler{Filepath: filepath}
}

func (h *ExceptionalHandler) GetFilePath(app models.App) string {
	if h.Filepath == "" {
		return ""
	}

	buf, err := os.ReadFile(h.Filepath)
	if err != nil {
		panic(err)
	}

	def := models.ExceptionalApps{}
	err = yaml.Unmarshal(buf, &def)
	if err != nil {
		return ""
	}
	for _, appNode := range def.Apps {
		if appNode.Name == app.Name {
			for _, envNode := range appNode.Envs {
				if envNode.Name == app.Env {
					return envNode.Path
				}
			}
		}
	}
	return ""
}

func (h *ExceptionalHandler) GetImage(app models.App) string {
	if h.Filepath == "" {
		return ""
	}

	buf, err := os.ReadFile(h.Filepath)
	if err != nil {
		panic(err)
	}

	def := models.ExceptionalApps{}
	err = yaml.Unmarshal(buf, &def)
	if err != nil {
		return ""
	}
	for _, appNode := range def.Apps {
		if appNode.Name == app.Name && appNode.Namespace == app.Namespace {
			return appNode.Image
		}
	}
	return ""
}
