package goat

import (
	"fmt"
	"os"
	"slices"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking/increament"
	"github.com/monshunter/goat/pkg/utils"
)

type FixExecutor struct {
	cfg                 *config.Config
	mainPackageInfos    []maininfo.MainPackageInfo
	fileTrackIdStartMap map[string]trackIdxInterval
	// filesContents is the contents of the files
	filesContents    map[string]string
	goModule         string
	goatImportPath   string
	goatPackageAlias string
	// changed is true if any `// + goat:delete`, `// + goat:insert` is found
	changed bool
}

func NewFixExecutor(cfg *config.Config) *FixExecutor {
	fixExecutor := &FixExecutor{
		cfg:                 cfg,
		fileTrackIdStartMap: make(map[string]trackIdxInterval),
		goModule:            config.GoModuleName(),
		filesContents:       make(map[string]string),
	}
	fixExecutor.goatImportPath = utils.GoatPackageImportPath(fixExecutor.goModule, fixExecutor.cfg.GoatPackagePath)
	fixExecutor.goatPackageAlias = cfg.GoatPackageAlias
	return fixExecutor
}

func (f *FixExecutor) Run() error {
	log.Infof("init main package infos")
	if err := f.initMainPackageInfos(); err != nil {
		log.Errorf("failed to init main package infos: %v", err)
		return err
	}
	log.Infof("preparing files")
	if err := f.prepare(); err != nil {
		log.Errorf("failed to prepare: %v", err)
		return err
	}

	if !f.changed {
		log.Infof("no files +goat:delete, +goat:insert found, no need to apply")
		return nil
	}
	log.Infof("applying fix")
	if err := f.apply(); err != nil {
		log.Errorf("failed to apply fix: %v", err)
		return err
	}
	log.Infof("fix applied")
	return nil
}

func (f *FixExecutor) initMainPackageInfos() error {
	mainPkgInfos, err := getMainPackageInfos(".", f.goModule, f.cfg.Ignores)
	if err != nil {
		return err
	}
	f.mainPackageInfos = mainPkgInfos
	return nil
}

func (f *FixExecutor) prepare() error {
	files, err := prepareFiles(f.cfg)
	if err != nil {
		log.Errorf("failed to prepare files: %v", err)
		return err
	}

	for _, file := range files {
		var content string
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			log.Errorf("failed to read file: %v", err)
			return err
		}
		updated := false
		count := 0
		// handle // + goat:delete
		count, content, err = handleGoatDelete(f.cfg.PrinterConfig(), string(contentBytes), f.goatImportPath, f.goatPackageAlias)
		if err != nil {
			log.Errorf("failed to handle goat delete: %v", err)
			return err
		}
		updated = updated || count > 0
		f.changed = f.changed || updated
		// handle // + goat:insert
		count, content, err = handleGoatInsert(f.cfg.PrinterConfig(), content, f.goatImportPath, f.goatPackageAlias)
		if err != nil {
			log.Errorf("failed to handle goat insert: %v", err)
			return err
		}
		updated = updated || count > 0
		f.changed = f.changed || updated
		// handle // + goat:generate
		count, content, err = resetGoatGenerate(content)
		if err != nil {
			log.Errorf("failed to reset goat generate: %v", err)
			return err
		}
		updated = updated || count > 0
		// handle // + goat:main
		isMainEntry := false
		for _, mainPkgInfo := range f.mainPackageInfos {
			if mainPkgInfo.MainFile == file {
				isMainEntry = true
				break
			}
		}
		if isMainEntry {
			count, content, err = resetGoatMain(f.cfg.PrinterConfig(), content, f.goatImportPath, f.goatPackageAlias)
			if err != nil {
				log.Errorf("failed to reset goat main entry: %v", err)
				return err
			}
			updated = updated || count > 0
		}

		if updated {
			f.filesContents[file] = content
		}
	}
	return nil
}

func (f *FixExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := utils.GoatPackageImportPath(f.goModule, f.cfg.GoatPackagePath)
	files := make([]string, 0)
	for file := range f.filesContents {
		files = append(files, file)
	}
	slices.Sort(files)
	for _, file := range files {
		content := f.filesContents[file]
		count, newContent, err := utils.Replace(content, increament.TrackStmtPlaceHolder,
			increament.IncreamentReplaceStmt(f.cfg.GoatPackageAlias, start))
		if err != nil {
			log.Errorf("failed to replace track stmt: %v", err)
			return 0, err
		}
		f.fileTrackIdStartMap[file] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increament.TrackImportPathPlaceHolder),
			increament.IncreamentReplaceImport(f.cfg.GoatPackageAlias, importPath))
		if err != nil {
			log.Errorf("failed to replace track import: %v", err)
			return 0, err
		}
		f.filesContents[file] = newContent
	}
	return start - 1, nil
}

func (f *FixExecutor) apply() error {
	count, err := f.replaceTracks()
	if err != nil {
		log.Errorf("failed to replace tracks: %v", err)
		return err
	}
	log.Infof("total replaced tracks: %d", count)

	err = f.applyTracks()
	if err != nil {
		log.Errorf("failed to apply tracks: %v", err)
		return err
	}

	// apply goat_generated.go
	componentTrackIdxs := getComponentTrackIdxs(f.fileTrackIdStartMap, f.mainPackageInfos)
	values := increament.NewValues(f.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}
	trackIdxs := getTotalTrackIdxs(f.fileTrackIdStartMap)
	// remove goat_generated.go if no track idxs
	if len(trackIdxs) == 0 {
		err = values.Remove(f.cfg.GoatGeneratedFile())
		if err != nil {
			log.Errorf("failed to remove goat_generated.go: %v", err)
			return err
		}
		return nil
	}

	values.AddTrackIds(trackIdxs)
	err = values.Save(f.cfg.GoatGeneratedFile())
	if err != nil {
		log.Errorf("failed to save goat_generated.go: %v", err)
		return err
	}

	// apply main entry
	if err := applyMainEntry(f.cfg, f.goModule, f.mainPackageInfos, componentTrackIdxs); err != nil {
		log.Errorf("failed to apply main entry: %v", err)
		return err
	}
	return nil
}

func (f *FixExecutor) applyTracks() error {
	for file, content := range f.filesContents {
		fset, fileAst, err := utils.GetAstTree("", []byte(content))
		if err != nil {
			log.Errorf("failed to get ast tree: %v, file: %s\n", err, file)
			return err
		}
		contentBytes, err := utils.FormatAst(f.cfg.PrinterConfig(), fset, fileAst)
		if err != nil {
			log.Errorf("failed to format ast: %v, file: %s\n", err, file)
			return err
		}
		info, err := os.Stat(file)
		if err != nil {
			log.Errorf("failed to get file info: %v, file: %s\n", err, file)
			return err
		}
		err = os.WriteFile(file, contentBytes, info.Mode().Perm())
		if err != nil {
			log.Errorf("failed to write file: %v, file: %s\n", err, file)
			return err
		}
	}
	return nil
}
