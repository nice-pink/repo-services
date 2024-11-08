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
	if currentTag == "" {
		log.Error("Can't find a current tag to replace.")
		return false
	}
	var err error
	replaced := false
	pattern := h.ImagePattern(dest)
	if dest.File != "" {
		replaced, err = SetTagInFileWithPattern(dest.Tag, currentTag, dest.File, pattern)
	} else {
		replaced, err = SetTagInFolderWithPattern(dest.Tag, currentTag, dest.Path, pattern)
	}

	if err != nil {
		log.Err(err, "Replacment error")
		return false
	}

	if replaced {
		// only if successfully updated tag
		h.AddTagToHistory(dest)
	}

	log.Info(util.LogPrefix(dest), "Updated tag:", dest.Tag)
	return true
}

func SetTagInFileWithPattern(tag, currentTag, filepath, pattern string) (replaced bool, err error) {
	replaced, err = filesystem.ReplaceRegexInFile(filepath, pattern, `${1}`+tag, false)
	if currentTag != "" {
		// update e.g. labels, annotations, etc. - ignore error
		filesystem.ReplaceInFile(filepath, currentTag, tag, false)
	}
	return replaced, err
}

func SetTagInFolderWithPattern(tag, currentTag, folder, pattern string) (replaced bool, err error) {
	replaced, err = filesystem.ReplaceRegexInAllFiles(folder, true, pattern, `${1}`+tag)
	if currentTag != "" {
		// update e.g. labels, annotations, etc. - ignore error
		filesystem.ReplaceInAllFiles(folder, true, currentTag, tag)
	}
	return replaced, err
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

func (h *ManifestHandler) BuildApp(flags util.GeneralFlags, tag string) models.App {
	folder := util.GetPathFromParameters(*flags.PathScheme, *flags.Base, *flags.Namespace, *flags.Env, *flags.App)
	folder = path.Join(*flags.SrcPath, folder)
	app := models.App{
		Path:      folder,
		Name:      *flags.App,
		Namespace: *flags.Namespace,
		Env:       *flags.Env,
		Tag:       tag,
		Image:     *flags.Image,
		File:      *flags.ImageFile,
		History:   *flags.ImageHistoryFile,
	}
	h.updateWithExceptionalHandler(&app, *flags.SrcPath, *flags.ImageHistoryFile)

	return app
}
