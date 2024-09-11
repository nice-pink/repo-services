package manifest

import (
	"path"

	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/repo-services/pkg/exceptional"
	"github.com/nice-pink/repo-services/pkg/models"
	"github.com/nice-pink/repo-services/pkg/util"
)

type ManifestHandler struct {
	exceptionalHandler *exceptional.ExceptionalHandler
}

func NewManifestHandler(exceptionalHandler *exceptional.ExceptionalHandler) *ManifestHandler {
	return &ManifestHandler{
		exceptionalHandler: exceptionalHandler,
	}
}

func (h *ManifestHandler) GetCurrentTags(app models.App) string {
	pattern := h.ImagePattern(app)
	extensions := []string{".yaml"}
	tags, err := filesystem.GetRegexInAllFiles(app.Path, false, pattern, `${2}`, extensions)
	if err != nil {
		log.Err(err, "get image tag")
		return ""
	}
	if len(tags) > 0 {
		return tags[0]
	}
	log.Info(util.LogPrefix(app), "Current tags:", tags)
	return ""
}

func (h *ManifestHandler) GetCurrentTag(app models.App) string {
	pattern := h.ImagePattern(app)
	tag, err := filesystem.GetRegexInFile(app.File, pattern, `${2}`, false)
	if err != nil {
		log.Err(err, "get image tag")
		return ""
	}
	log.Info(util.LogPrefix(app), "Current tag:", tag)
	return tag
}

func (h *ManifestHandler) SetTag(dest models.App) bool {
	currentTag := h.GetCurrentTag(dest)
	var err error
	pattern := h.ImagePattern(dest)
	tag := `${1}` + dest.Tag
	if dest.File != "" {
		err = filesystem.ReplaceRegexInFile(dest.File, pattern, tag, false)
		// update e.g. labels, annotations, etc. - ignore error
		filesystem.ReplaceInFile(dest.File, currentTag, dest.Tag, false)
	} else {
		err = filesystem.ReplaceRegexInAllFiles(dest.Path, true, pattern, tag)
		// update e.g. labels, annotations, etc. - ignore error
		filesystem.ReplaceInAllFiles(dest.Path, true, currentTag, dest.Tag)
	}

	if err != nil {
		log.Err(err, "Replacment error")
		return false
	}
	log.Info(util.LogPrefix(dest), "Updated tag:", dest.Tag)
	return true
}

func (h *ManifestHandler) SetTagWithSource(src, dest models.App) bool {
	tag := h.GetCurrentTag(src)
	if tag == "" {
		return false
	}
	dest.Tag = tag
	return h.SetTag(dest)
}

// image

func (h *ManifestHandler) getImageName(app models.App) string {
	if h.exceptionalHandler != nil {
		image := h.exceptionalHandler.GetImage(app)
		if image != "" {
			return image
		}
	}
	if app.Image != "" {
		return app.Image
	}
	return app.Name
}

func (h *ManifestHandler) updateImageName(app *models.App) {
	image := h.getImageName(*app)
	app.Image = image
}

func (h *ManifestHandler) getImageFile(app models.App, srcFolder string) string {
	if h.exceptionalHandler != nil {
		filepath := h.exceptionalHandler.GetFilePath(app)
		if filepath != "" {
			return path.Join(srcFolder, filepath)
		}
	}
	return path.Join(app.Path, app.File)
}

func (h *ManifestHandler) updateImageFile(app *models.App, srcFolder string) {
	imageFile := h.getImageFile(*app, srcFolder)
	app.File = imageFile
}

func (h *ManifestHandler) ImagePattern(app models.App) string {
	return `(.*[/| ]` + h.getImageName(app) + `:)([a-zA-Z0-9_-].*)`
}

// app

func (h *ManifestHandler) BuildApp(name, env, namespace, image, scheme, base, imageFileName, srcFolder, tag string) models.App {
	folder := util.GetPathFromParameters(scheme, base, namespace, env, name)
	folder = path.Join(srcFolder, folder)
	app := models.App{
		Path:      folder,
		Name:      name,
		Namespace: namespace,
		Env:       env,
		Tag:       tag,
		Image:     image,
		File:      imageFileName,
	}
	h.updateImageName(&app)
	h.updateImageFile(&app, srcFolder)

	return app
}

// Path      string
// 	Name      string
// 	Namespace string
// 	Tag       string
// 	Image     string
// 	File      string
// 	Env       string
// 	Envs      []ExceptionalEnvDef
