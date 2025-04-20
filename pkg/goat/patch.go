package goat

import (
	"fmt"
	"sort"

	"github.com/monshunter/goat/pkg/log"

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
		goModule:            config.GoModuleName(),
	}
}

// Run runs the patch executor
func (p *PatchExecutor) Run() error {
	if err := p.initChanges(); err != nil {
		return fmt.Errorf("failed to initialize changes: %w", err)
	}

	if err := p.initMainPackageInfos(); err != nil {
		return fmt.Errorf("failed to initialize main packages: %w", err)
	}

	if err := p.initTracks(); err != nil {
		return fmt.Errorf("failed to initialize trackers: %w", err)
	}

	count, err := p.replaceTracks()
	if err != nil {
		return fmt.Errorf("failed to replace tracks: %w", err)
	}

	log.Debugf("Replaced %d tracking points", count)

	componentTrackIdxs := getComponentTrackIdxs(p.fileTrackIdStartMap, p.mainPackageInfos)

	values := increament.NewValues(p.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}

	values.AddTrackIds(getTotalTrackIdxs(p.fileTrackIdStartMap))

	if err = values.Save(p.cfg.GoatGeneratedFile()); err != nil {
		return fmt.Errorf("failed to save generated file %s: %w", p.cfg.GoatGeneratedFile(), err)
	}

	if err := p.saveTracks(); err != nil {
		return fmt.Errorf("failed to save tracking points: %w", err)
	}

	if err := applyMainEntry(p.cfg, p.goModule, p.mainPackageInfos, componentTrackIdxs); err != nil {
		return fmt.Errorf("failed to apply main entry: %w", err)
	}

	log.Infof("Patch applied successfully with %d tracking points", count)
	return nil
}

// initChanges initializes the changes
func (p *PatchExecutor) initChanges() error {
	changes, err := getDiff(p.cfg)
	if err != nil {
		return fmt.Errorf("failed to get code differences: %w", err)
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	p.changes = changes
	log.Debugf("Found %d file changes", len(changes))
	return nil
}

// initMainPackageInfos initializes the main package infos
func (p *PatchExecutor) initMainPackageInfos() error {
	mainPkgInfos, err := getMainPackageInfos(".", p.goModule, p.cfg.Ignores)
	if err != nil {
		return fmt.Errorf("failed to get main package info: %w", err)
	}
	p.mainPackageInfos = mainPkgInfos
	log.Debugf("Found %d main packages", len(mainPkgInfos))
	return nil
}

// initTracks initializes the trackers
func (p *PatchExecutor) initTracks() error {
	trackers := make([]tracking.Tracker, len(p.changes))
	for i, change := range p.changes {
		tracker, err := p.handleDiffChange(change)
		if err != nil {
			return fmt.Errorf("failed to handle file change %s: %w", change.Path, err)
		}
		trackers[i] = tracker
	}

	p.trackers = trackers
	log.Debugf("Initialized %d trackers", len(trackers))
	return nil
}

func (p *PatchExecutor) handleDiffChange(change *diff.FileChange) (tracking.Tracker, error) {
	granularity := p.cfg.GetGranularity()
	tracker, err := tracking.NewIncreamentTrack(".", change,
		increament.TrackImportPathPlaceHolder, increament.GetPackageInsertData(),
		nil, granularity, p.cfg.PrinterConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create incremental tracker: %w", err)
	}
	_, err = tracker.Track()
	if err != nil {
		return nil, fmt.Errorf("failed to track file: %w", err)
	}
	log.Debugf("Successfully tracked file: %s", change.Path)
	return tracker, nil
}

// replaceTracks replaces the tracks
func (p *PatchExecutor) replaceTracks() (int, error) {
	start := 1
	importPath := utils.GoatPackageImportPath(p.goModule, p.cfg.GoatPackagePath)
	for i, tracker := range p.trackers {
		count, err := tracker.Replace(increament.TrackStmtPlaceHolder, increament.IncreamentReplaceStmt(p.cfg.GoatPackageAlias, start))
		if err != nil || count != tracker.Count() {
			return 0, fmt.Errorf("failed to replace statements in %s: expected=%d, actual=%d: %w",
				tracker.TargetFile(), tracker.Count(), count, err)
		}
		p.fileTrackIdStartMap[p.changes[i].Path] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, err = tracker.Replace(fmt.Sprintf("%q", increament.TrackImportPathPlaceHolder),
			increament.IncreamentReplaceImport(p.cfg.GoatPackageAlias, importPath))
		if err != nil {
			return 0, fmt.Errorf("failed to replace import in %s: %w", tracker.TargetFile(), err)
		}
		log.Debugf("Replaced %d tracking points in %s", count, tracker.TargetFile())
	}
	return start - 1, nil
}

// saveTracks saves the trackers
func (p *PatchExecutor) saveTracks() error {
	totalSaved := 0
	for _, tracker := range p.trackers {
		if err := tracker.Save(""); err != nil {
			return fmt.Errorf("failed to save tracker for %s: %w", tracker.TargetFile(), err)
		}
		totalSaved++
	}
	log.Debugf("Saved %d tracker files", totalSaved)
	return nil
}
