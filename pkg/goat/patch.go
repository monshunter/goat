package goat

import (
	"fmt"
	"log"
	"sort"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking"
	"github.com/monshunter/goat/pkg/tracking/increament"
	"github.com/monshunter/goat/pkg/utils"
)

// PatchExecutor is the executor for the patch
type PatchExecutor struct {
	cfg                 *config.Config
	changes             []*diff.FileChange
	mainPackageInfos    []maininfo.MainPackageInfo
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
	if err := p.initMainPackageInfos(); err != nil {
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

	componentTrackIdxs := getComponentTrackIdxs(p.fileTrackIdStartMap, p.mainPackageInfos)

	values := increament.NewValues(p.cfg.GoatPackageName, p.cfg.AppVersion, p.cfg.AppName, p.cfg.Race)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}

	values.AddTrackIds(getTotalTrackIdxs(p.fileTrackIdStartMap))

	err = values.Save(p.cfg.GoatGeneratedFile())
	if err != nil {
		return err
	}

	if err := p.saveTracks(); err != nil {
		return err
	}
	if err := applyMainEntry(p.cfg, p.goModule, p.mainPackageInfos, componentTrackIdxs); err != nil {
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

// initMainPackageInfos initializes the main package infos
func (p *PatchExecutor) initMainPackageInfos() error {
	mainPkgInfos, err := getMainPackageInfos(p.cfg.ProjectRoot, p.goModule)
	if err != nil {
		return err
	}
	p.mainPackageInfos = mainPkgInfos
	return nil
}

// initTracks initializes the trackers
func (p *PatchExecutor) initTracks() error {
	trackers := make([]tracking.Tracker, len(p.changes))
	for i, change := range p.changes {
		tracker, err := p.handleDiffChange(change)
		if err != nil {
			log.Printf("failed to get tracker: %v", err)
			return err
		}
		trackers[i] = tracker
	}

	p.trackers = trackers
	return nil
}

func (p *PatchExecutor) handleDiffChange(change *diff.FileChange) (tracking.Tracker, error) {
	granularity := p.cfg.GetGranularity()
	tracker, err := tracking.NewIncreamentTrack(p.cfg.ProjectRoot, change,
		increament.TrackImportPathPlaceHolder, increament.GetPackageInsertData(), nil, granularity)
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

// replaceTracks replaces the tracks
func (p *PatchExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := utils.GoatPackageImportPath(p.goModule, p.cfg.GoatPackagePath)
	for i, tracker := range p.trackers {
		count, err := tracker.Replace(increament.TrackStmtPlaceHolder, increament.IncreamentReplaceStmt(p.cfg.GoatPackageAlias, start))
		if err != nil || count != tracker.Count() {
			log.Printf("failed to replace stmt: i: %d, err: %v, count: %d, expected: %d\n", i, err, count, tracker.Count())
			return 0, err
		}
		p.fileTrackIdStartMap[p.changes[i].Path] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, err = tracker.Replace(fmt.Sprintf("%q", increament.TrackImportPathPlaceHolder),
			increament.IncreamentReplaceImport(p.cfg.GoatPackageAlias, importPath))
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
