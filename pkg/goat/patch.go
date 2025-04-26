package goat

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/monshunter/goat/pkg/log"
	"github.com/monshunter/goat/pkg/tracking/increment"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/utils"
)

// PatchExecutor is the executor for the patch
type PatchExecutor struct {
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

// NewPatchExecutor creates a new patch executor
func NewPatchExecutor(cfg *config.Config) *PatchExecutor {
	PatchExecutor := &PatchExecutor{
		cfg:                 cfg,
		fileTrackIdStartMap: make(map[string]trackIdxInterval),
		goModule:            config.GoModuleName(),
		filesContents:       make(map[string]string),
	}
	PatchExecutor.goatImportPath = utils.GoatPackageImportPath(PatchExecutor.goModule, PatchExecutor.cfg.GoatPackagePath)
	PatchExecutor.goatPackageAlias = cfg.GoatPackageAlias
	return PatchExecutor
}

func (f *PatchExecutor) Run() error {
	log.Infof("Patching project")
	if err := f.initMainPackageInfos(); err != nil {
		log.Errorf("Failed to init main package infos: %v", err)
		return err
	}
	if err := f.prepare(); err != nil {
		log.Errorf("Failed to prepare: %v", err)
		return err
	}

	if !f.changed {
		log.Infof("No files with +goat:delete, +goat:insert found, no need to apply")
		return nil
	}
	if err := f.apply(); err != nil {
		log.Errorf("Failed to apply patch: %v", err)
		return err
	}
	log.Infof("Patch applied")
	return nil
}

// initMainPackageInfos initializes the main package infos
func (f *PatchExecutor) initMainPackageInfos() error {
	log.Infof("Getting main package infos")
	mainPkgInfos, err := getMainPackageInfos(".", f.goModule, f.cfg.Ignores)
	if err != nil {
		return err
	}
	f.mainPackageInfos = mainPkgInfos
	return nil
}

func (f *PatchExecutor) prepare() error {
	log.Infof("Preparing files")
	files, err := prepareFiles(f.cfg)
	if err != nil {
		log.Errorf("Failed to prepare files: %v", err)
		return err
	}
	goatFiles, err := f.prepareContents(files)
	if err != nil {
		log.Errorf("Failed to prepare contents: %v", err)
		return err
	}
	f.filesContents = make(map[string]string, len(goatFiles))
	for _, goatFile := range goatFiles {
		f.filesContents[goatFile.filename] = goatFile.content
	}
	return nil
}

// prepareContents prepares the contents of the files
func (f *PatchExecutor) prepareContents(files []string) ([]goatFile, error) {
	if f.cfg.Threads == 1 {
		return f.prepareContentsSequential(files)
	}
	return f.prepareContentsParallel(files)
}

// prepareContentsSequential prepares the contents of the files sequentially
func (f *PatchExecutor) prepareContentsSequential(files []string) ([]goatFile, error) {
	goatFiles := make([]goatFile, 0, len(files))
	for _, file := range files {
		goatFile, err := f.prepareContent(file)
		if err != nil {
			log.Errorf("Failed to prepare content: %v", err)
			return nil, err
		}
		if goatFile.filename == "" || goatFile.content == "" {
			continue
		}
		goatFiles = append(goatFiles, goatFile)
	}
	return goatFiles, nil
}

// prepareContentsParallel prepares the contents of the files in parallel
func (f *PatchExecutor) prepareContentsParallel(files []string) ([]goatFile, error) {
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
				log.Errorf("Failed to prepare content: %v", err)
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
		if goatFile.filename == "" || goatFile.content == "" {
			continue
		}
		goatFiles = append(goatFiles, goatFile)
	}
	return goatFiles, nil
}

// prepareContent prepares the content of the file
func (f *PatchExecutor) prepareContent(file string) (goatFile, error) {
	var content string
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file: %v", err)
		return goatFile{}, err
	}
	updated := false
	count := 0
	// handle // + goat:delete
	count, content, err = handleGoatDelete(f.cfg.PrinterConfig(), string(contentBytes), f.goatImportPath, f.goatPackageAlias)
	if err != nil {
		log.Errorf("Failed to handle goat delete: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	f.changed = f.changed || updated
	// handle // + goat:insert
	count, content, err = handleGoatInsert(f.cfg.PrinterConfig(), content, f.goatImportPath, f.goatPackageAlias)
	if err != nil {
		log.Errorf("Failed to handle goat insert: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	f.changed = f.changed || updated
	// handle // + goat:generate
	count, content, err = resetGoatGenerate(content)
	if err != nil {
		log.Errorf("Failed to reset goat generate: %v", err)
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
			log.Errorf("Failed to reset goat main entry: %v", err)
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

// replaceTracks replaces the tracks in the files
func (f *PatchExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := utils.GoatPackageImportPath(f.goModule, f.cfg.GoatPackagePath)
	files := make([]string, 0)
	for file := range f.filesContents {
		files = append(files, file)
	}
	slices.Sort(files)
	for _, file := range files {
		content := f.filesContents[file]
		count, newContent, err := utils.Replace(content, increment.TrackStmtPlaceHolder,
			increment.IncreamentReplaceStmt(f.cfg.GoatPackageAlias, start))
		if err != nil {
			log.Errorf("Failed to replace track stmt: %v", err)
			return 0, err
		}
		f.fileTrackIdStartMap[file] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increment.TrackImportPathPlaceHolder),
			increment.IncreamentReplaceImport(f.cfg.GoatPackageAlias, importPath))
		if err != nil {
			log.Errorf("Failed to replace track import: %v", err)
			return 0, err
		}
		f.filesContents[file] = newContent
	}
	return start - 1, nil
}

// apply applies the patch
func (f *PatchExecutor) apply() error {
	log.Infof("Applying patch")
	count, err := f.replaceTracks()
	if err != nil {
		log.Errorf("Failed to replace tracks: %v", err)
		return err
	}
	log.Infof("Total replaced tracks: %d", count)

	err = f.applyTracks()
	if err != nil {
		log.Errorf("Failed to apply tracks: %v", err)
		return err
	}

	// apply goat_generated.go
	componentTrackIdxs := getComponentTrackIdxs(f.fileTrackIdStartMap, f.mainPackageInfos)
	values := increment.NewValues(f.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}
	trackIdxs := getTotalTrackIdxs(f.fileTrackIdStartMap)
	// remove goat_generated.go if no track idxs
	if len(trackIdxs) == 0 {
		err = values.Remove(f.cfg.GoatGeneratedFile())
		if err != nil {
			log.Errorf("Failed to remove goat_generated.go: %v", err)
			return err
		}
		return nil
	}

	values.AddTrackIds(trackIdxs)
	err = values.Save(f.cfg.GoatGeneratedFile())
	if err != nil {
		log.Errorf("Failed to save goat_generated.go: %v", err)
		return err
	}

	// apply main entry
	if err := applyMainEntries(f.cfg, f.goModule, f.mainPackageInfos, componentTrackIdxs); err != nil {
		log.Errorf("Failed to apply main entries: %v", err)
		return err
	}
	return nil
}

// applyTracks applies the tracks
func (f *PatchExecutor) applyTracks() error {
	if f.cfg.Threads == 1 {
		return f.applyTracksSequential()
	}
	return f.applyTracksParallel()
}

// applyTracksSequential applies the tracks sequentially
func (f *PatchExecutor) applyTracksSequential() error {
	for file, content := range f.filesContents {
		err := utils.FormatAndSave(file, []byte(content), f.cfg.PrinterConfig())
		if err != nil {
			log.Errorf("Failed to format and save file: %v", err)
			return err
		}
	}
	return nil
}

// applyTracksParallel applies the tracks in parallel
func (f *PatchExecutor) applyTracksParallel() error {
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
			err := utils.FormatAndSave(file, []byte(content), f.cfg.PrinterConfig())
			if err != nil {
				log.Errorf("Failed to format and save file: %v", err)
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
