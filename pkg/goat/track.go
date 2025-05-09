package goat

import (
	"fmt"
	"sort"
	"sync"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/maininfo"
	"github.com/monshunter/goat/pkg/tracking"
	"github.com/monshunter/goat/pkg/tracking/increment"
	"github.com/monshunter/goat/pkg/utils"
)

// TrackExecutor is the executor for the track
type TrackExecutor struct {
	cfg                 *config.Config
	changes             []*diff.FileChange
	mainPackageInfos    []maininfo.MainPackageInfo
	trackers            []tracking.Tracker
	fileTrackIdStartMap map[string]trackIdxInterval
	goModule            string
}

// NewTrackExecutor creates a new track executor
func NewTrackExecutor(cfg *config.Config) *TrackExecutor {
	return &TrackExecutor{
		cfg:                 cfg,
		fileTrackIdStartMap: make(map[string]trackIdxInterval),
		goModule:            config.GoModuleName(),
	}
}

// Run runs the track executor
func (p *TrackExecutor) Run() error {
	log.Infof("Tracking project")
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

	log.Infof("Replaced %d tracking points", count)

	componentTrackIdxs := getComponentTrackIdxs(p.fileTrackIdStartMap, p.mainPackageInfos)

	values := increment.NewValues(p.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}

	values.AddTrackIds(getTotalTrackIdxs(p.fileTrackIdStartMap))
	log.Infof("Saving generated file %s", p.cfg.GoatGeneratedFile())
	if err = values.Save(p.cfg.GoatGeneratedFile()); err != nil {
		return fmt.Errorf("failed to save generated file %s: %w", p.cfg.GoatGeneratedFile(), err)
	}

	log.Infof("Saving tracking points to %d files", len(p.trackers))
	if err := p.saveTracks(); err != nil {
		return fmt.Errorf("failed to save tracking points: %w", err)
	}

	log.Infof("Applying main entries")
	if err := applyMainEntries(p.cfg, p.goModule, p.mainPackageInfos, componentTrackIdxs); err != nil {
		return fmt.Errorf("failed to apply main entries: %w", err)
	}

	log.Infof("Track applied successfully with %d tracking points", count)
	return nil
}

// initChanges initializes the changes
func (p *TrackExecutor) initChanges() error {
	log.Infof("Getting code differences")
	changes, err := getDiff(p.cfg)
	if err != nil {
		return fmt.Errorf("failed to get code differences: %w", err)
	}
	// sort the changes by path
	// this sort is important for the incremental tracker,
	// because the incremental tracker will use the changes to generate the tracking points
	// and the tracking points are sorted by the path
	// so the tracking points will be in the same order for each time run
	// the same track between the two same commits
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	p.changes = changes
	log.Debugf("Found %d file changes", len(changes))
	return nil
}

// initMainPackageInfos initializes the main package infos
func (p *TrackExecutor) initMainPackageInfos() error {
	log.Infof("Getting main package infos")
	mainPkgInfos, err := getMainPackageInfos(".", p.goModule, p.cfg.Ignores)
	if err != nil {
		return fmt.Errorf("failed to get main package info: %w", err)
	}
	p.mainPackageInfos = mainPkgInfos
	log.Debugf("Found %d main packages", len(mainPkgInfos))
	return nil
}

// initTracks initializes the trackers
func (p *TrackExecutor) initTracks() error {
	log.Infof("Initializing trackers")
	if p.cfg.Threads == 1 {
		return p.initTracksSequential()
	}
	return p.initTracksParallel()
}

// initTracksSequential initializes the trackers sequentially
func (p *TrackExecutor) initTracksSequential() error {
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

// initTracksParallel initializes the trackers in parallel
func (p *TrackExecutor) initTracksParallel() error {
	trackers := make([]tracking.Tracker, len(p.changes))
	sem := make(chan struct{}, p.cfg.Threads)
	errChan := make(chan error, len(p.changes))
	wg := sync.WaitGroup{}
	wg.Add(len(p.changes))
	for i, change := range p.changes {
		sem <- struct{}{}
		go func(i int, change *diff.FileChange) {
			defer func() {
				<-sem
				wg.Done()
			}()
			tracker, err := p.handleDiffChange(change)
			if err != nil {
				errChan <- fmt.Errorf("failed to handle file change %s: %w", change.Path, err)
				return
			}
			trackers[i] = tracker
		}(i, change)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	p.trackers = trackers
	log.Debugf("Initialized %d trackers", len(trackers))
	return nil
}

// handleDiffChange handles the diff change
func (p *TrackExecutor) handleDiffChange(change *diff.FileChange) (tracking.Tracker, error) {
	granularity := p.cfg.GetGranularity()
	tracker, err := tracking.NewIncrementalTrack(".", change,
		increment.TrackImportPathPlaceHolder, increment.GetPackageInsertStmts(),
		granularity, p.cfg.PrinterConfig())
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
func (p *TrackExecutor) replaceTracks() (int, error) {
	log.Infof("Replacing tracks")
	start := 1
	importPath := utils.GoatPackageImportPath(p.goModule, p.cfg.GoatPackagePath)
	for i, tracker := range p.trackers {
		count, newContent, err := utils.Replace(string(tracker.Content()), increment.TrackStmtPlaceHolder,
			increment.IncreamentReplaceStmt(p.cfg.GoatPackageAlias, start))
		if err != nil || count != tracker.Count() {
			return 0, fmt.Errorf("failed to replace statements in %s: expected=%d, actual=%d: %w",
				tracker.Target(), tracker.Count(), count, err)
		}
		p.fileTrackIdStartMap[p.changes[i].Path] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increment.TrackImportPathPlaceHolder),
			increment.IncreamentReplaceImport(p.cfg.GoatPackageAlias, importPath))
		if err != nil {
			return 0, fmt.Errorf("failed to replace import in %s: %w", tracker.Target(), err)
		}
		tracker.SetContent([]byte(newContent))
		log.Debugf("Replaced %d tracking points in %s", count, tracker.Target())
	}
	return start - 1, nil
}

// saveTracks saves the trackers
func (p *TrackExecutor) saveTracks() error {
	if p.cfg.Threads == 1 {
		return p.saveTracksSequential()
	}
	return p.saveTracksParallel()
}

// saveTracksSequential saves the trackers sequentially
func (p *TrackExecutor) saveTracksSequential() error {
	for _, tracker := range p.trackers {
		if err := utils.FormatAndSave(tracker.Target(), tracker.Content(), p.cfg.PrinterConfig()); err != nil {
			return fmt.Errorf("failed to save tracker for %s: %w", tracker.Target(), err)
		}
	}
	log.Debugf("Saved %d tracking files", len(p.trackers))
	return nil
}

// saveTracksParallel saves the trackers in parallel
func (p *TrackExecutor) saveTracksParallel() error {
	wg := sync.WaitGroup{}
	wg.Add(len(p.trackers))
	sem := make(chan struct{}, p.cfg.Threads)
	errChan := make(chan error, len(p.trackers))
	for _, tracker := range p.trackers {
		sem <- struct{}{}
		go func(tracker tracking.Tracker) {
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := utils.FormatAndSave(tracker.Target(), tracker.Content(), p.cfg.PrinterConfig()); err != nil {
				errChan <- fmt.Errorf("failed to save tracker for %s: %w", tracker.Target(), err)
				return
			}
		}(tracker)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	log.Debugf("Saved %d tracker files", len(p.trackers))
	return nil
}
