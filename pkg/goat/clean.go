package goat

import (
	"log"
	"os"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

func RunClean(cfg *config.Config) error {
	return nil
}

type CleanExecutor struct {
	cfg              *config.Config
	goatImportPath   string
	goatPackageAlias string
	files            []goatFile
}

func NewCleanExecutor(cfg *config.Config) *CleanExecutor {
	executor := &CleanExecutor{
		cfg:   cfg,
		files: make([]goatFile, 0),
	}
	executor.goatImportPath = utils.GoatPackageImportPath(config.GoModuleName(), cfg.GoatPackagePath)
	executor.goatPackageAlias = cfg.GoatPackageAlias
	return executor
}

func (e *CleanExecutor) Run() error {
	if err := e.prepare(); err != nil {
		return err
	}
	if err := e.clean(); err != nil {
		return err
	}
	return nil
}

func (e *CleanExecutor) prepare() error {
	var err error
	files, err := prepareFiles(e.cfg)
	if err != nil {
		return err
	}
	if err := e.prepareContents(files); err != nil {
		return err
	}
	return nil
}

func (e *CleanExecutor) prepareContents(files []string) error {
	for _, file := range files {
		content, changed, err := e.prepareContent(file)
		if err != nil {
			return err
		}
		if changed {
			e.files = append(e.files, goatFile{
				filename: file,
				content:  content,
			})
		}
	}
	return nil
}
func (e *CleanExecutor) prepareContent(filename string) (string, bool, error) {
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", false, err
	}
	content := string(contentBytes)
	changed := false
	// handle +goat:delete
	count, newContent, err := utils.ReplaceWithRegexp(config.TrackDeleteEndRegexp,
		content, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:insert
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackInsertRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:generate
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:main
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:user
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackUserEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		return "", false, err
	}
	changed = changed || count > 0

	if changed {
		// remove import
		bytes, err := utils.DeleteImport(e.cfg.PrinterConfig(), e.goatImportPath, e.goatPackageAlias, "", []byte(newContent))
		if err != nil {
			return "", false, err
		}
		return string(bytes), true, nil
	}

	return newContent, changed, nil
}

func (e *CleanExecutor) clean() error {
	for _, file := range e.files {
		if err := os.WriteFile(file.filename, []byte(file.content), 0644); err != nil {
			return err
		}
	}
	log.Printf("cleaned %d files", len(e.files))
	os.Remove(e.cfg.GoatGeneratedFile())
	// remove goat package if empty
	empty, err := utils.IsDirEmpty(e.cfg.GoatPackagePath)
	if err != nil {
		return err
	}
	if empty {
		os.RemoveAll(e.cfg.GoatPackagePath)
	}
	return nil
}
