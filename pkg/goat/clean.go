package goat

import (
	"os"

	"github.com/monshunter/goat/pkg/log"

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
	log.Infof("preparing files")
	if err := e.prepare(); err != nil {
		log.Errorf("failed to prepare: %v", err)
		return err
	}
	log.Infof("cleaning files")
	if err := e.clean(); err != nil {
		log.Errorf("failed to clean: %v", err)
		return err
	}
	log.Infof("cleaned files")
	return nil
}

func (e *CleanExecutor) prepare() error {
	var err error
	files, err := prepareFiles(e.cfg)
	if err != nil {
		log.Errorf("failed to prepare files: %v", err)
		return err
	}
	if err := e.prepareContents(files); err != nil {
		log.Errorf("failed to prepare contents: %v", err)
		return err
	}
	log.Infof("prepared %d files", len(e.files))
	return nil
}

func (e *CleanExecutor) prepareContents(files []string) error {
	for _, file := range files {
		content, changed, err := e.prepareContent(file)
		if err != nil {
			log.Errorf("failed to prepare content: %v", err)
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
	log.Debugf("preparing content for file: %s", filename)
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("failed to read file: %v", err)
		return "", false, err
	}
	content := string(contentBytes)
	changed := false
	// handle +goat:delete
	log.Debugf("replacing +goat:delete for file: %s", filename)
	count, newContent, err := utils.ReplaceWithRegexp(config.TrackDeleteEndRegexp,
		content, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:delete: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:insert
	log.Debugf("replacing +goat:insert for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackInsertRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:insert: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:generate
	log.Debugf("replacing +goat:generate for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackGenerateEndRegexp,
		newContent, func(older string) (newer string) {
			return ""
		})
	if err != nil {
		log.Errorf("failed to replace +goat:generate: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:main
	log.Debugf("replacing +goat:main for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackMainEntryEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("failed to replace +goat:main: %v", err)
		return "", false, err
	}
	changed = changed || count > 0
	// handle +goat:user
	log.Debugf("replacing +goat:user for file: %s", filename)
	count, newContent, err = utils.ReplaceWithRegexp(config.TrackUserEndRegexp, newContent, func(older string) (newer string) {
		return ""
	})
	if err != nil {
		log.Errorf("failed to replace +goat:user: %v", err)
		return "", false, err
	}
	changed = changed || count > 0

	if changed {
		// remove import
		log.Debugf("deleting import for file: %s", filename)
		bytes, err := utils.DeleteImport(e.cfg.PrinterConfig(), e.goatImportPath, e.goatPackageAlias, "", []byte(newContent))
		if err != nil {
			log.Errorf("failed to delete import: %v", err)
			return "", false, err
		}
		return string(bytes), true, nil
	}

	return newContent, changed, nil
}

func (e *CleanExecutor) clean() error {
	for _, file := range e.files {
		log.Debugf("cleaning file: %s", file.filename)
		if err := os.WriteFile(file.filename, []byte(file.content), 0644); err != nil {
			log.Errorf("failed to write file: %v", err)
			return err
		}
	}
	log.Infof("total cleaned files: %d", len(e.files))
	log.Debugf("removing goat generated file: %s", e.cfg.GoatGeneratedFile())
	os.Remove(e.cfg.GoatGeneratedFile())
	// remove goat package if empty
	log.Debugf("checking if goat package is empty: %s", e.cfg.GoatPackagePath)
	empty, err := utils.IsDirEmpty(e.cfg.GoatPackagePath)
	if err != nil {
		log.Errorf("failed to check if goat package is empty: %v", err)
		return err
	}
	if empty {
		log.Debugf("removing goat package: %s", e.cfg.GoatPackagePath)
		os.RemoveAll(e.cfg.GoatPackagePath)
	}
	return nil
}
