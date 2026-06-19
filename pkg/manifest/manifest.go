package manifest

import (
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/nice-pink/goutil/pkg/filesystem"
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
		slog.Default().Error("get_image_tag", "err", err)
		return ""
	}
	if len(tags) > 0 {
		return tags[0]
	}
	slog.Default().Info("get_current_tags", "prefix", util.LogPrefix(app), "tags", tags)
	return ""
}

func (h *ManifestHandler) GetCurrentTag(app models.App) string {
	pattern := h.ImagePattern(app)
	tag, err := filesystem.GetRegexInFile(app.File, pattern, `${2}`, false)
	if err != nil {
		slog.Default().Error("get_image_tag_from_file", "err", err, "file", app.File)
		return ""
	}
	slog.Default().Info("get_current_tag", "prefix", util.LogPrefix(app), "tag", tag)
	return tag
}

func (h *ManifestHandler) SetTag(dest models.App) bool {
	currentTag := h.GetCurrentTag(dest)
	if currentTag == "" {
		slog.Default().Error("set_tag_no_current")
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
		slog.Default().Error("set_tag_replacement", "err", err)
		return false
	}

	if replaced {
		// only if successfully updated tag
		h.AddTagToHistory(dest)
	}

	slog.Default().Info("set_tag_ok", "prefix", util.LogPrefix(dest), "tag", dest.Tag)
	return true
}

// atomicReplaceRegexInFile reads the file at filepath, applies the regex replacement,
// and writes the result atomically via a temp file + fsync + rename.
// This ensures no partial write is observable if the process is killed mid-write.
func atomicReplaceRegexInFile(filePath, pattern, replacement string) (bool, error) {
	// Read existing content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	newContent := re.ReplaceAll(content, []byte(replacement))
	if string(newContent) == string(content) {
		// No change — let caller know (replaced=false)
		return false, nil
	}

	// Write to a temp file in the same directory (same filesystem → atomic rename on POSIX)
	dir := path.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".manifest-*.tmp")
	if err != nil {
		return false, err
	}
	tmpName := tmp.Name()

	_, writeErr := tmp.Write(newContent)
	syncErr := tmp.Sync()
	closeErr := tmp.Close()

	if writeErr != nil || syncErr != nil {
		os.Remove(tmpName)
		if writeErr != nil {
			return false, writeErr
		}
		return false, syncErr
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return false, closeErr
	}

	if err := os.Rename(tmpName, filePath); err != nil {
		os.Remove(tmpName)
		return false, err
	}

	return true, nil
}

// atomicReplaceInFile replaces a plain string oldStr with newStr atomically.
func atomicReplaceInFile(filePath, oldStr, newStr string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newContent := []byte(strings.ReplaceAll(string(content), oldStr, newStr))
	if string(newContent) == string(content) {
		return nil
	}

	dir := path.Dir(filePath)
	tmp, err := os.CreateTemp(dir, ".manifest-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	_, writeErr := tmp.Write(newContent)
	syncErr := tmp.Sync()
	closeErr := tmp.Close()

	if writeErr != nil || syncErr != nil {
		os.Remove(tmpName)
		if writeErr != nil {
			return writeErr
		}
		return syncErr
	}
	if closeErr != nil {
		os.Remove(tmpName)
		return closeErr
	}

	return os.Rename(tmpName, filePath)
}


func SetTagInFileWithPattern(tag, currentTag, filePath, pattern string) (replaced bool, err error) {
	// Atomic write: apply primary regex replacement.
	replaced, err = atomicReplaceRegexInFile(filePath, pattern, "${1}"+tag)
	if err != nil {
		return false, err
	}
	if currentTag != "" {
		// update e.g. labels, annotations, etc. - ignore error
		atomicReplaceInFile(filePath, currentTag, tag)
	}
	return replaced, nil
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

// AddTagToHistory appends the tag to the app's history file.
// Returns true on success, false on failure or if no history file is configured.
// (Bug fix: was previously returning err != nil which was inverted.)
func (h *ManifestHandler) AddTagToHistory(app models.App) bool {
	if app.History == "" {
		return false
	}
	err := filesystem.AppendToFile(app.History, "- "+app.Tag, true)
	return err == nil
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
