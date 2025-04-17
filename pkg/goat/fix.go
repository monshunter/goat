package goat

import (
	"fmt"
	"log"
	"os"
	"slices"

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
	if err := f.initMainPackageInfos(); err != nil {
		return err
	}
	if err := f.prepare(); err != nil {
		return err
	}

	if !f.changed {
		log.Printf("no files +goat:delete, +goat:insert found, no need to apply")
		return nil
	}
	if err := f.apply(); err != nil {
		return err
	}
	return nil
}

func (f *FixExecutor) initMainPackageInfos() error {
	mainPkgInfos, err := getMainPackageInfos(".", f.goModule)
	if err != nil {
		return err
	}
	f.mainPackageInfos = mainPkgInfos
	return nil
}

func (f *FixExecutor) prepare() error {
	files, err := prepareFiles(f.cfg)

	if err != nil {
		return err
	}

	for _, file := range files {
		var content string
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		updated := false
		count := 0
		// handle // + goat:delete
		count, content, err = handleGoatDelete(f.cfg.PrinterConfig(), string(contentBytes), f.goatImportPath, f.goatPackageAlias)
		if err != nil {
			return err
		}
		updated = updated || count > 0
		f.changed = f.changed || updated
		// handle // + goat:insert
		count, content, err = handleGoatInsert(f.cfg.PrinterConfig(), content, f.goatImportPath, f.goatPackageAlias)
		if err != nil {
			return err
		}
		updated = updated || count > 0
		f.changed = f.changed || updated
		// handle // + goat:generate
		count, content, err = resetGoatGenerate(content)
		if err != nil {
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
			return 0, err
		}
		f.fileTrackIdStartMap[file] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increament.TrackImportPathPlaceHolder),
			increament.IncreamentReplaceImport(f.cfg.GoatPackageAlias, importPath))
		if err != nil {
			return 0, err
		}
		f.filesContents[file] = newContent
	}
	return start - 1, nil
}

func (f *FixExecutor) apply() error {
	count, err := f.replaceTracks()
	if err != nil {
		return err
	}
	log.Printf("replaced %d tracks", count)

	err = f.applyTracks()
	if err != nil {
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
			return err
		}
		return nil
	}

	values.AddTrackIds(trackIdxs)
	err = values.Save(f.cfg.GoatGeneratedFile())
	if err != nil {
		return err
	}

	// apply main entry
	if err := applyMainEntry(f.cfg, f.goModule, f.mainPackageInfos, componentTrackIdxs); err != nil {
		return err
	}
	return nil
}

func (f *FixExecutor) applyTracks() error {
	for file, content := range f.filesContents {
		fset, fileAst, err := utils.GetAstTree("", []byte(content))
		if err != nil {
			log.Printf("error: failed to get ast tree: %s, file: %s\n", err, file)
			return err
		}
		contentBytes, err := utils.FormatAst(f.cfg.PrinterConfig(), fset, fileAst)
		if err != nil {
			log.Printf("error: failed to format ast: %s, file: %s\n", err, file)
			return err
		}
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("error: failed to get file info: %s, file: %s\n", err, file)
			return err
		}
		err = os.WriteFile(file, contentBytes, info.Mode().Perm())
		if err != nil {
			log.Printf("error: failed to write file: %s, file: %s\n", err, file)
			return err
		}
	}
	return nil
}
