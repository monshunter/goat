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

func (p *PatchExecutor) Run() error {
	log.Infof("Patching project")
	if err := p.initMainPackageInfos(); err != nil {
		log.Errorf("Failed to init main package infos: %v", err)
		return err
	}
	if err := p.prepare(); err != nil {
		log.Errorf("Failed to prepare: %v", err)
		return err
	}

	if !p.changed {
		log.Infof("No files with +goat:delete, +goat:insert found, no need to apply")
		return nil
	}
	if err := p.apply(); err != nil {
		log.Errorf("Failed to apply patch: %v", err)
		return err
	}
	log.Infof("Patch applied")
	return nil
}

// initMainPackageInfos initializes the main package infos
func (p *PatchExecutor) initMainPackageInfos() error {
	log.Infof("Getting main package infos")
	mainPkgInfos, err := getMainPackageInfosWithConfig(".", p.goModule, p.cfg.Ignores, p.cfg.SkipNestedModules)
	if err != nil {
		return err
	}
	p.mainPackageInfos = mainPkgInfos
	return nil
}

func (p *PatchExecutor) prepare() error {
	log.Infof("Preparing files")
	files, err := prepareFiles(p.cfg)
	if err != nil {
		log.Errorf("Failed to prepare files: %v", err)
		return err
	}
	goatFiles, err := p.prepareContents(files)
	if err != nil {
		log.Errorf("Failed to prepare contents: %v", err)
		return err
	}
	p.filesContents = make(map[string]string, len(goatFiles))
	for _, goatFile := range goatFiles {
		p.filesContents[goatFile.filename] = goatFile.content
	}
	return nil
}

// prepareContents prepares the contents of the files
func (p *PatchExecutor) prepareContents(files []string) ([]goatFile, error) {
	if p.cfg.Threads == 1 {
		return p.prepareContentsSequential(files)
	}
	return p.prepareContentsParallel(files)
}

// prepareContentsSequential prepares the contents of the files sequentially
func (p *PatchExecutor) prepareContentsSequential(files []string) ([]goatFile, error) {
	goatFiles := make([]goatFile, 0, len(files))
	for _, file := range files {
		goatFile, err := p.prepareContent(file)
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
func (p *PatchExecutor) prepareContentsParallel(files []string) ([]goatFile, error) {
	goatFiles := make([]goatFile, 0, len(files))
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, p.cfg.Threads)
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
			goatFile, err := p.prepareContent(file)
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
func (p *PatchExecutor) prepareContent(file string) (goatFile, error) {
	var content string
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file: %v", err)
		return goatFile{}, err
	}
	updated := false
	count := 0
	// handle // + goat:delete
	count, content, err = handleGoatDelete(p.cfg.PrinterConfig(), string(contentBytes), p.goatImportPath, p.goatPackageAlias)
	if err != nil {
		log.Errorf("Failed to handle goat delete: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	p.changed = p.changed || updated
	// handle // + goat:insert
	count, content, err = handleGoatInsert(p.cfg.PrinterConfig(), content, p.goatImportPath, p.goatPackageAlias)
	if err != nil {
		log.Errorf("Failed to handle goat insert: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	p.changed = p.changed || updated
	// handle // + goat:generate
	count, content, err = resetGoatGenerate(content)
	if err != nil {
		log.Errorf("Failed to reset goat generate: %v", err)
		return goatFile{}, err
	}
	updated = updated || count > 0
	// handle // + goat:main
	isMainEntry := false
	for _, mainPkgInfo := range p.mainPackageInfos {
		if mainPkgInfo.MainFile == file {
			isMainEntry = true
			break
		}
	}
	if isMainEntry {
		count, content, err = resetGoatMain(p.cfg.PrinterConfig(), content, p.goatImportPath, p.goatPackageAlias)
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
func (p *PatchExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := utils.GoatPackageImportPath(p.goModule, p.cfg.GoatPackagePath)
	files := make([]string, 0)
	for file := range p.filesContents {
		files = append(files, file)
	}
	slices.Sort(files)
	for _, file := range files {
		content := p.filesContents[file]
		count, newContent, err := utils.Replace(content, increment.TrackStmtPlaceHolder,
			increment.IncreamentReplaceStmt(p.cfg.GoatPackageAlias, start))
		if err != nil {
			log.Errorf("Failed to replace track stmt: %v", err)
			return 0, err
		}
		p.fileTrackIdStartMap[file] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increment.TrackImportPathPlaceHolder),
			increment.IncreamentReplaceImport(p.cfg.GoatPackageAlias, importPath))
		if err != nil {
			log.Errorf("Failed to replace track import: %v", err)
			return 0, err
		}
		p.filesContents[file] = newContent
	}
	return start - 1, nil
}

// apply applies the patch
func (p *PatchExecutor) apply() error {
	log.Infof("Applying patch")
	count, err := p.replaceTracks()
	if err != nil {
		log.Errorf("Failed to replace tracks: %v", err)
		return err
	}
	log.Infof("Total replaced tracks: %d", count)

	err = p.applyTracks()
	if err != nil {
		log.Errorf("Failed to apply tracks: %v", err)
		return err
	}

	// apply goat_generated.go
	componentTrackIdxs := getComponentTrackIdxs(p.fileTrackIdStartMap, p.mainPackageInfos)
	values := increment.NewValues(p.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}
	trackIdxs := getTotalTrackIdxs(p.fileTrackIdStartMap)
	// remove goat_generated.go if no track idxs
	if len(trackIdxs) == 0 {
		err = values.Remove(p.cfg.GoatGeneratedFile())
		if err != nil {
			log.Errorf("Failed to remove goat_generated.go: %v", err)
			return err
		}
		return nil
	}

	values.AddTrackIds(trackIdxs)

	if values.IsEmpty() {
		log.Infof("No tracking points found, skip saving generated file")
		return nil
	}

	log.Infof("Saving generated file %s", p.cfg.GoatGeneratedFile())
	err = values.Save(p.cfg.GoatGeneratedFile())
	if err != nil {
		log.Errorf("Failed to save goat_generated.go: %v", err)
		return err
	}

	// apply main entry
	if err := applyMainEntries(p.cfg, p.goModule, p.mainPackageInfos, componentTrackIdxs); err != nil {
		log.Errorf("Failed to apply main entries: %v", err)
		return err
	}
	return nil
}

// applyTracks applies the tracks
func (p *PatchExecutor) applyTracks() error {
	if p.cfg.Threads == 1 {
		return p.applyTracksSequential()
	}
	return p.applyTracksParallel()
}

// applyTracksSequential applies the tracks sequentially
func (p *PatchExecutor) applyTracksSequential() error {
	for file, content := range p.filesContents {
		err := utils.FormatAndSave(file, []byte(content), p.cfg.PrinterConfig())
		if err != nil {
			log.Errorf("Failed to format and save file: %v", err)
			return err
		}
	}
	return nil
}

// applyTracksParallel applies the tracks in parallel
func (p *PatchExecutor) applyTracksParallel() error {
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, p.cfg.Threads)
	errChan := make(chan error, len(p.filesContents))
	wg.Add(len(p.filesContents))
	for file, content := range p.filesContents {
		sem <- struct{}{}
		go func(file string, content string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			err := utils.FormatAndSave(file, []byte(content), p.cfg.PrinterConfig())
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
