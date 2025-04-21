package goat

import (
	"fmt"
	"os"
	"slices"
	"sync"

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
	goatFiles, err := f.prepareContents(files)
	if err != nil {
		log.Errorf("failed to prepare contents: %v", err)
		return err
	}
	f.filesContents = make(map[string]string, len(goatFiles))
	for _, goatFile := range goatFiles {
		f.filesContents[goatFile.filename] = goatFile.content
	}
	return nil
}

func (f *FixExecutor) prepareContents(files []string) ([]goatFile, error) {
	if f.cfg.Threads == 1 {
		return f.prepareContentsSequential(files)
	}
	return f.prepareContentsParallel(files)
}

func (f *FixExecutor) prepareContentsSequential(files []string) ([]goatFile, error) {
	goatFiles := make([]goatFile, 0, len(files))
	for _, file := range files {
		goatFile, err := f.prepareContent(file)
		if err != nil {
			log.Errorf("failed to prepare content: %v", err)
			return nil, err
		}
		goatFiles = append(goatFiles, goatFile)
	}
	return goatFiles, nil
}

func (f *FixExecutor) prepareContentsParallel(files []string) ([]goatFile, error) {
	goatFiles := make([]goatFile, 0, len(files))
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, f.cfg.Threads)
	errChan := make(chan error, len(files))
	fileChan := make(chan goatFile, len(files))
	wg.Add(len(files))
	for _, file := range files {
		sem <- struct{}{}
		go func(file string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			goatFile, err := f.prepareContent(file)
			if err != nil {
				log.Errorf("failed to prepare content: %v", err)
				errChan <- err
				return
			}
			fileChan <- goatFile
		}(file)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	close(fileChan)
	for goatFile := range fileChan {
		goatFiles = append(goatFiles, goatFile)
	}
	return goatFiles, nil
}

func (f *FixExecutor) prepareContent(file string) (goatFile, error) {
	var content string
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("failed to read file: %v", err)
		return goatFile{}, err
	}
	updated := false
	count := 0
	// handle // + goat:delete
	count, content, err = handleGoatDelete(f.cfg.PrinterConfig(), string(contentBytes), f.goatImportPath, f.goatPackageAlias)
	if err != nil {
		log.Errorf("failed to handle goat delete: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	f.changed = f.changed || updated
	// handle // + goat:insert
	count, content, err = handleGoatInsert(f.cfg.PrinterConfig(), content, f.goatImportPath, f.goatPackageAlias)
	if err != nil {
		log.Errorf("failed to handle goat insert: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	f.changed = f.changed || updated
	// handle // + goat:generate
	count, content, err = resetGoatGenerate(content)
	if err != nil {
		log.Errorf("failed to reset goat generate: %v", err)
		return goatFile{}, err
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
			return goatFile{}, err
		}
		updated = updated || count > 0
	}

	if !updated {
		return goatFile{}, nil
	}
	return goatFile{
		filename: file,
		content:  content,
	}, nil
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
	if f.cfg.Threads == 1 {
		return f.applyTracksSequential()
	}
	return f.applyTracksParallel()
}

func (f *FixExecutor) applyTracksSequential() error {
	for file, content := range f.filesContents {
		err := utils.FormatAndWrite(file, []byte(content), f.cfg.PrinterConfig())
		if err != nil {
			log.Errorf("failed to format and write file: %v", err)
			return err
		}
	}
	return nil
}

func (f *FixExecutor) applyTracksParallel() error {
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, f.cfg.Threads)
	errChan := make(chan error, len(f.filesContents))
	wg.Add(len(f.filesContents))
	for file, content := range f.filesContents {
		sem <- struct{}{}
		go func(file string, content string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			err := utils.FormatAndWrite(file, []byte(content), f.cfg.PrinterConfig())
			if err != nil {
				log.Errorf("failed to format and write file: %v", err)
				errChan <- err
				return
			}
		}(file, content)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
