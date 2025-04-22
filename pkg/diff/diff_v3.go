package diff

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

// This version cannot recognize the scenario of file migration and modification.
// All migrated files are regarded as deleted and newly created.
//
//	Blame the limited capabilities of go-git. I have tried various methods but still got no result.
//
// DifferV3 is the third version of the code difference analyzer
type DifferV3 struct {
	cfg      *config.Config
	repoInfo *repoInfo
}

// NewDifferV3 creates a new code DifferV3
func NewDifferV3(cfg *config.Config) (*DifferV3, error) {
	repoInfo, err := newRepoInfo(cfg.OldBranch, cfg.NewBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo info: %w", err)
	}
	d := &DifferV3{
		repoInfo: repoInfo,
		cfg:      cfg,
	}
	return d, nil
}

// AnalyzeChanges analyzes code changes between two branches
func (d *DifferV3) AnalyzeChanges() ([]*FileChange, error) {
	changes, err := d.repoInfo.getObjectChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to compare branch DifferV3ences: %w", err)
	}
	fileChanges := make([]*FileChange, len(changes))
	errChan := make(chan error, len(changes))
	sem := make(chan struct{}, d.cfg.Threads)
	var wg sync.WaitGroup
	wg.Add(len(changes))
	for i, change := range changes {
		sem <- struct{}{} // acquire worker slot
		go func(idx int, c *object.Change) {
			defer func() {
				<-sem // release worker slot
				wg.Done()
			}()
			fc, err := d.analyzeChange(c)
			if err != nil {
				errChan <- err
				return
			}
			if fc != nil && len(fc.LineChanges) > 0 {
				fileChanges[idx] = fc
			}

		}(i, change)
	}

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	return filterValidFileChanges(fileChanges), nil
}

// analyzeChange analyzes a single file change
func (d *DifferV3) analyzeChange(change *object.Change) (*FileChange, error) {

	action, err := change.Action()
	if err != nil {
		return nil, fmt.Errorf("failed to get change action: %w", err)
	}

	if change.From.Name != change.To.Name && change.From.Name != "" && change.To.Name != "" {
		return nil, nil
	}
	switch action {
	case merkletrie.Insert, merkletrie.Modify:
		return d.handleInsert(change)
	default:
		return nil, nil
	}
}

// handleInsert handles insert or modify operations
func (d *DifferV3) handleInsert(change *object.Change) (*FileChange, error) {
	fileName := change.To.Name
	if !utils.IsTargetFile(fileName, d.cfg.Ignores) {
		return nil, nil
	}
	patch, err := change.Patch()
	if err != nil || patch == nil {
		return nil, err
	}
	fc := FileChange{
		Path: fileName,
	}
	lineChanges := make([]LineChange, 0)
	for _, filePatch := range patch.FilePatches() {
		lineChanges = append(lineChanges, getLineChange(filePatch)...)
	}
	fc.LineChanges = lineChanges
	return &fc, nil
}
