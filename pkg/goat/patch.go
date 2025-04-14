package goat

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking"
	"github.com/monshunter/goat/pkg/tracking/increament"
)

// PatchExecutor is the executor for the patch
type PatchExecutor struct {
	cfg                 *config.Config
	changes             []*diff.FileChange
	mainPkgInfo         *maininfo.MainInfo
	trackers            []tracking.Tracker
	fileTrackIdStartMap map[string]trackIdxInterval
	goModule            string
}

// NewPatchExecutor creates a new patch executor
func NewPatchExecutor(cfg *config.Config) *PatchExecutor {
	return &PatchExecutor{
		cfg:                 cfg,
		fileTrackIdStartMap: make(map[string]trackIdxInterval),
		goModule:            config.GoModuleName(cfg.ProjectRoot),
	}
}

// Run runs the patch executor
func (p *PatchExecutor) Run() error {
	if err := p.initChanges(); err != nil {
		return err
	}
	if err := p.initMainInfo(); err != nil {
		return err
	}
	// debugChanges(p.changes)
	// debugMainInfo(p.mainPkgInfo)

	if err := p.initTracks(); err != nil {
		return err
	}
	// fmt.Println("before replace", string(p.trackers[0].Bytes()))

	count, err := p.replaceTracks()
	if err != nil {
		return err
	}
	log.Printf("replaced %d tracks", count)
	// fmt.Println("after replace", string(p.trackers[0].Bytes()))
	componentTrackIdxs := p.getComponentTrackIdxs()
	// debugComponentTrackIdxs(componentTrackIdxs)
	// return nil
	values := increament.NewValues(p.cfg.GoatPackageName, p.cfg.AppVersion, p.cfg.AppName, p.cfg.Race)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}

	values.AddTrackIds(p.getTotalTrackIdxs())

	err = values.Save(filepath.Join(p.cfg.ProjectRoot, p.cfg.GoatPackagePath, "goat_generated.go"))
	if err != nil {
		return err
	}
	// return nil
	if err := p.saveTracks(); err != nil {
		return err
	}
	if err := p.applyMainEntry(); err != nil {
		return err
	}
	return nil
}

// initChanges initializes the changes
func (p *PatchExecutor) initChanges() error {
	changes, err := getDiff(p.cfg)
	if err != nil {
		log.Printf("failed to get differ: %v", err)
		return err
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	p.changes = changes
	return nil
}

// initMainInfo initializes the main package info
func (p *PatchExecutor) initMainInfo() error {
	mainPkgInfo, err := maininfo.NewMainInfo(p.cfg.ProjectRoot, p.goModule)
	if err != nil {
		log.Printf("failed to get main info: %v", err)
		return err
	}
	if len(mainPkgInfo.MainPackageInfos) == 0 {
		log.Printf("warning: no main package info found")
		return fmt.Errorf("warning: no main package info found")
	}

	p.mainPkgInfo = mainPkgInfo
	return nil
}

// initTracks initializes the trackers
func (p *PatchExecutor) initTracks() error {
	handle := func(change *diff.FileChange) (tracking.Tracker, error) {
		granularity := p.cfg.GetGranularity()
		tracker, err := tracking.NewIncreamentTrack(p.cfg.ProjectRoot, change, nil, granularity)
		if err != nil {
			log.Printf("failed to get tracker: %v", err)
			return nil, err
		}
		_, err = tracker.Track()
		if err != nil {
			log.Printf("failed to track: %v", err)
			return nil, err
		}
		return tracker, nil
	}
	trackers := make([]tracking.Tracker, len(p.changes))
	for i, change := range p.changes {
		tracker, err := handle(change)
		if err != nil {
			log.Printf("failed to get tracker: %v", err)
			return err
		}
		trackers[i] = tracker
	}

	p.trackers = trackers
	return nil
}

// replaceTracks replaces the tracks
func (p *PatchExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := filepath.Join(p.goModule, p.cfg.GoatPackagePath)
	for i, tracker := range p.trackers {
		count, err := tracker.Replace(tracking.DefaultTrackStmt, tracking.IncreamentReplaceStmt(p.cfg.GoatPackageAlias, start))
		if err != nil || count != tracker.Count() {
			log.Printf("failed to replace stmt: i: %d, err: %v, count: %d, expected: %d\n", i, err, count, tracker.Count())
			return 0, err
		}
		p.fileTrackIdStartMap[p.changes[i].Path] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, err = tracker.Replace(fmt.Sprintf("%q", tracking.DefaultImportPath),
			tracking.IncreamentReplaceImport(p.cfg.GoatPackageAlias, importPath))
		if err != nil {
			log.Printf("failed to replace import: i: %d, err: %v, count: %d, expected: %d\n", i, err, count, tracker.Count())
			return 0, err
		}
	}
	return start - 1, nil
}

// saveTracks saves the trackers
func (p *PatchExecutor) saveTracks() error {
	for _, tracker := range p.trackers {
		err := tracker.Save("")
		if err != nil {
			log.Printf("failed to save tracker: %v", err)
			return err
		}
	}
	return nil
}

// getComponentTrackIdxs returns the componentTrackIdxs
func (p *PatchExecutor) getComponentTrackIdxs() []componentTrackIdx {
	// packageTrackIdxMap: package -> trackIdxs
	packageTrackIdxMap := make(map[string][]int)
	for path, interval := range p.fileTrackIdStartMap {
		pkg := filepath.Dir(path)
		ids := make([]int, 0, interval.end-interval.start+1)
		for i := interval.start; i <= interval.end; i++ {
			ids = append(ids, i)
		}
		packageTrackIdxMap[pkg] = append(packageTrackIdxMap[pkg], ids...)
	}

	// componentTrackIdxs: componentId -> component -> trackIdxs
	componentTrackIdxs := make([]componentTrackIdx, 0)
	for i, mainInfo := range p.mainPkgInfo.MainPackageInfos {
		trackIdxs := make([]int, 0)
		for _, pkg := range mainInfo.Imports {
			ids, ok := packageTrackIdxMap[pkg]
			if !ok {
				continue
			}
			trackIdxs = append(trackIdxs, ids...)
		}
		sort.Ints(trackIdxs)
		component := componentTrackIdx{
			componentId: i + 1,
			component:   mainInfo.MainDir,
			trackIdx:    trackIdxs,
		}
		componentTrackIdxs = append(componentTrackIdxs, component)
	}
	return componentTrackIdxs
}

func (p *PatchExecutor) getTotalTrackIdxs() []int {
	idxs := make([]int, 0)
	for _, interval := range p.fileTrackIdStartMap {
		for i := interval.start; i <= interval.end; i++ {
			idxs = append(idxs, i)
		}
	}

	sort.Ints(idxs)
	slow, fast := 0, 0
	for fast < len(idxs) {
		if idxs[slow] == idxs[fast] {
			fast++
		} else {
			slow++
			idxs[slow] = idxs[fast]
		}
	}
	return idxs[:slow+1]
}

func (p *PatchExecutor) applyMainEntry() error {
	importPath := filepath.Join(p.goModule, p.cfg.GoatPackagePath)
	for i, mainInfo := range p.mainPkgInfo.MainPackageInfos {
		if !p.cfg.IsMainEntry(mainInfo.MainDir) {
			continue
		}
		codes := increament.GetMainEntryInitData(p.cfg.GoatPackageAlias, i+1)
		_, err := mainInfo.ApplyMainEntry(p.cfg.GoatPackageAlias, importPath, codes)
		if err != nil {
			log.Printf("failed to apply main entry: %v", err)
			return err
		}
	}
	return nil
}
