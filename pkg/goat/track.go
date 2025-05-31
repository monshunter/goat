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
	replacedFiles       int
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
func (t *TrackExecutor) Run() error {
	log.Infof("Tracking project")
	if err := t.initChanges(); err != nil {
		return fmt.Errorf("failed to initialize changes: %w", err)
	}

	if err := t.initMainPackageInfos(); err != nil {
		return fmt.Errorf("failed to initialize main packages: %w", err)
	}

	if err := t.initTracks(); err != nil {
		return fmt.Errorf("failed to initialize trackers: %w", err)
	}

	count, err := t.replaceTracks()
	if err != nil {
		return fmt.Errorf("failed to replace tracks: %w", err)
	}

	log.Infof("Replaced %d tracking points", count)

	componentTrackIdxs := getComponentTrackIdxs(t.fileTrackIdStartMap, t.mainPackageInfos)

	values := increment.NewValues(t.cfg)
	for _, component := range componentTrackIdxs {
		values.AddComponent(component.componentId, component.component, component.trackIdx)
	}

	values.AddTrackIds(getTotalTrackIdxs(t.fileTrackIdStartMap))

	if values.IsEmpty() {
		log.Infof("No tracking points found, skip saving generated file")
		return nil
	}

	log.Infof("Saving generated file %s", t.cfg.GoatGeneratedFile())
	if err = values.Save(t.cfg.GoatGeneratedFile()); err != nil {
		return fmt.Errorf("failed to save generated file %s: %w", t.cfg.GoatGeneratedFile(), err)
	}

	log.Infof("Saving tracking points to %d files", t.replacedFiles)
	if err := t.saveTracks(); err != nil {
		return fmt.Errorf("failed to save tracking points: %w", err)
	}

	log.Infof("Applying main entries")
	if err := applyMainEntries(t.cfg, t.goModule, t.mainPackageInfos, componentTrackIdxs); err != nil {
		return fmt.Errorf("failed to apply main entries: %w", err)
	}

	log.Infof("Track applied successfully with %d tracking points", count)
	return nil
}

// initChanges initializes the changes
func (t *TrackExecutor) initChanges() error {
	log.Infof("Getting code differences")
	changes, err := getDiff(t.cfg)
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

	t.changes = changes
	log.Debugf("Found %d file changes", len(changes))
	return nil
}

// initMainPackageInfos initializes the main package infos
func (t *TrackExecutor) initMainPackageInfos() error {
	log.Infof("Getting main package infos")
	mainPkgInfos, err := getMainPackageInfosWithConfig(".", t.goModule, t.cfg.Ignores, t.cfg.SkipNestedModules)
	if err != nil {
		return fmt.Errorf("failed to get main package info: %w", err)
	}
	t.mainPackageInfos = mainPkgInfos
	log.Debugf("Found %d main packages", len(mainPkgInfos))
	return nil
}

// initTracks initializes the trackers
func (t *TrackExecutor) initTracks() error {
	log.Infof("Initializing trackers")
	if t.cfg.Threads == 1 {
		return t.initTracksSequential()
	}
	return t.initTracksParallel()
}

// initTracksSequential initializes the trackers sequentially
func (t *TrackExecutor) initTracksSequential() error {
	trackers := make([]tracking.Tracker, len(t.changes))
	for i, change := range t.changes {
		tracker, err := t.handleDiffChange(change)
		if err != nil {
			return fmt.Errorf("failed to handle file change %s: %w", change.Path, err)
		}
		trackers[i] = tracker
	}

	t.trackers = trackers
	log.Debugf("Initialized %d trackers", len(trackers))
	return nil
}

// initTracksParallel initializes the trackers in parallel
func (t *TrackExecutor) initTracksParallel() error {
	trackers := make([]tracking.Tracker, len(t.changes))
	sem := make(chan struct{}, t.cfg.Threads)
	errChan := make(chan error, len(t.changes))
	wg := sync.WaitGroup{}
	wg.Add(len(t.changes))
	for i, change := range t.changes {
		sem <- struct{}{}
		go func(i int, change *diff.FileChange) {
			defer func() {
				<-sem
				wg.Done()
			}()
			tracker, err := t.handleDiffChange(change)
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
	t.trackers = trackers
	log.Debugf("Initialized %d trackers", len(trackers))
	return nil
}

// handleDiffChange handles the diff change
func (t *TrackExecutor) handleDiffChange(change *diff.FileChange) (tracking.Tracker, error) {
	granularity := t.cfg.GetGranularity()
	tracker, err := tracking.NewIncrementalTrack(".", change,
		increment.TrackImportPathPlaceHolder, increment.GetPackageInsertStmts(),
		granularity, t.cfg.PrinterConfig())
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
func (t *TrackExecutor) replaceTracks() (int, error) {
	log.Infof("Replacing tracks")
	start := 1
	importPath := utils.GoatPackageImportPath(t.goModule, t.cfg.GoatPackagePath)
	for i, tracker := range t.trackers {
		count, newContent, err := utils.Replace(string(tracker.Content()), increment.TrackStmtPlaceHolder,
			increment.IncreamentReplaceStmt(t.cfg.GoatPackageAlias, start))
		if err != nil || count != tracker.Count() {
			return 0, fmt.Errorf("failed to replace statements in %s: expected=%d, actual=%d: %w",
				tracker.Target(), tracker.Count(), count, err)
		}
		t.fileTrackIdStartMap[t.changes[i].Path] = trackIdxInterval{start: start, end: start + count - 1}
		start += count
		_, newContent, err = utils.Replace(newContent, fmt.Sprintf("%q", increment.TrackImportPathPlaceHolder),
			increment.IncreamentReplaceImport(t.cfg.GoatPackageAlias, importPath))
		if err != nil {
			return 0, fmt.Errorf("failed to replace import in %s: %w", tracker.Target(), err)
		}
		tracker.SetContent([]byte(newContent))
		if count > 0 {
			t.replacedFiles++
		}
		log.Debugf("Replaced %d tracking points in %s", count, tracker.Target())
	}
	return start - 1, nil
}

// saveTracks saves the trackers
func (t *TrackExecutor) saveTracks() error {
	if t.cfg.Threads == 1 {
		return t.saveTracksSequential()
	}
	return t.saveTracksParallel()
}

// saveTracksSequential saves the trackers sequentially
func (t *TrackExecutor) saveTracksSequential() error {
	for _, tracker := range t.trackers {
		if err := utils.FormatAndSave(tracker.Target(), tracker.Content(), t.cfg.PrinterConfig()); err != nil {
			return fmt.Errorf("failed to save tracker for %s: %w", tracker.Target(), err)
		}
	}
	log.Debugf("Saved %d tracking files", t.replacedFiles)
	return nil
}

// saveTracksParallel saves the trackers in parallel
func (t *TrackExecutor) saveTracksParallel() error {
	wg := sync.WaitGroup{}
	wg.Add(len(t.trackers))
	sem := make(chan struct{}, t.cfg.Threads)
	errChan := make(chan error, len(t.trackers))
	for _, tracker := range t.trackers {
		sem <- struct{}{}
		go func(tracker tracking.Tracker) {
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := utils.FormatAndSave(tracker.Target(), tracker.Content(), t.cfg.PrinterConfig()); err != nil {
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
	log.Debugf("Saved %d tracker files", t.replacedFiles)
	return nil
}

// applyMainEntries applies the main entries
