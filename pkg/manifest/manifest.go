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
		log.Err(err, "get image tag from: "+app.File)
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

	// only if successfully updated tag
	// h.AddTagToHistory(dest)

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

func (h *ManifestHandler) AddTagToHistory(app models.App) bool {
	if app.History == "" {
		return false
	}
	err := filesystem.AppendToFile(app.History, "- "+app.Tag, true)
	return err != nil
}

// image

func (h *ManifestHandler) updateWithExceptionalHandler(app *models.App, srcFolder, historyFileName string) {
	h.updateImagePaths(app, srcFolder, historyFileName)
	h.updateImageName(app)
}

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

func (h *ManifestHandler) getImageFile(app models.App, srcFolder string) (folder string, filename string) {
	if h.exceptionalHandler == nil {
		return app.Path, app.File
	}

	folder, filename = h.exceptionalHandler.GetFileInfo(app)
	if folder != "" {
		// log.Info("Found exceptional path:", folder)
		folder = path.Join(srcFolder, folder)
	} else {
		folder = app.Path
	}
	if filename == "" {
		filename = app.File
	}
	return folder, filename
}

func (h *ManifestHandler) updateImagePaths(app *models.App, srcFolder, historyFileName string) {
	folder, filename := h.getImageFile(*app, srcFolder)
	app.Path = folder
	app.File = path.Join(folder, filename)
	if historyFileName != "" {
		app.History = path.Join(folder, historyFileName)
	}
}

func (h *ManifestHandler) ImagePattern(app models.App) string {
	return `(.*[/| ]` + h.getImageName(app) + `:)([a-zA-Z0-9_-].*)`
}

// app

func (h *ManifestHandler) BuildApp(name, env, namespace, image, scheme, base, imageFileName, historyFileName, srcFolder, tag string) models.App {
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
		History:   historyFileName,
	}
	h.updateWithExceptionalHandler(&app, srcFolder, historyFileName)

	return app
}
