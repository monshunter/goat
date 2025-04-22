package diff

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

// DifferV2 is the second version of the code difference analyzer
type DifferV2 struct {
	cfg      *config.Config
	repoInfo *repoInfo
}

// NewDifferV2 creates a new code difference analyzer
func NewDifferV2(cfg *config.Config) (*DifferV2, error) {
	repoInfo, err := newRepoInfo(cfg.OldBranch, cfg.NewBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to create repo info: %w", err)
	}
	d := &DifferV2{
		repoInfo: repoInfo,
		cfg:      cfg,
	}
	return d, nil
}

// AnalyzeChanges analyzes code changes between two branches
func (d *DifferV2) AnalyzeChanges() ([]*FileChange, error) {
	filePatches, err := d.repoInfo.getFilePatches()
	if err != nil {
		return nil, fmt.Errorf("failed to get file patches: %w", err)
	}
	fileChanges := make([]*FileChange, len(filePatches))
	sem := make(chan struct{}, d.cfg.Threads) // concurrent workers,default 10
	var wg sync.WaitGroup
	wg.Add(len(filePatches))
	for i, filePatch := range filePatches {
		sem <- struct{}{} // acquire worker slot
		go func(idx int, filePath diff.FilePatch) {
			defer func() {
				<-sem // release worker slot
				wg.Done()
			}()
			fc := d.analyzeChange(filePatch)
			if fc != nil && len(fc.LineChanges) > 0 {
				fileChanges[idx] = fc
			}
		}(i, filePatch)
	}
	wg.Wait()
	return filterValidFileChanges(fileChanges), nil
}

// analyzeChange analyzes a file patch and returns a FileChange
func (d *DifferV2) analyzeChange(filePatch diff.FilePatch) *FileChange {
	from, to := filePatch.Files()
	if (from == nil && to == nil) || (from != nil && to == nil) {
		return nil
	}
	if !utils.IsTargetFile(to.Path(), d.cfg.Ignores) {
		return nil
	}
	lineChanges := getLineChange(filePatch)
	if len(lineChanges) == 0 {
		return nil
	}
	return &FileChange{Path: to.Path(), LineChanges: lineChanges}
}
