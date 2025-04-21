package diff

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/utils"
)

// DifferV2 Code DifferV2ence Analyzer
type DifferV2 struct {
	cfg       *config.Config
	repo      *git.Repository
	oldHash   plumbing.Hash
	newHash   plumbing.Hash
	newCommit *object.Commit
	oldCommit *object.Commit
}

// NewDifferV2 creates a new code DifferV2
func NewDifferV2(cfg *config.Config) (*DifferV2, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	if err := checkUncommittedChanges(repo); err != nil {
		return nil, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}
	d := &DifferV2{
		repo: repo,
		cfg:  cfg,
	}
	oldHash, err := resolveRef(d.repo, cfg.OldBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve old branch: %w", err)
	}

	newHash, err := resolveRef(d.repo, cfg.NewBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve new branch: %w", err)
	}

	oldCommit, err := d.repo.CommitObject(oldHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get old branch commit: %w", err)
	}

	newCommit, err := d.repo.CommitObject(newHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get new branch commit: %w", err)
	}
	d.oldHash = oldHash
	d.newHash = newHash
	d.oldCommit = oldCommit
	d.newCommit = newCommit
	return d, nil
}

// AnalyzeChangesV2 Correctness evaluates the correctness of AnalyzeChangesV2 implementation
// 1. Correctness:
//    - Correctly identifies added chunks in diff
//    - Properly handles file filtering for Go files
//    - Correctly calculates line numbers for changes
//    - Missing handling for renamed files (same as original AnalyzeChanges)
//    - Doesn't properly handle empty files (count could be 0)
//
// 2. Algorithm Analysis:
//    - Uses patch-based diffing which is more efficient than tree diff for some cases
//    - Complexity: O(n) where n is total lines in all changed files
//    - Memory efficient - processes files sequentially without storing all changes
//    - More accurate line counting than blame-based approach
//
// 3. Optimization Suggestions:
//    - Add parallel processing similar to original AnalyzeChanges
//    - Cache file content to avoid repeated splitting
//    - Pre-allocate lineChanges slice with estimated capacity
//    - Add special handling for large files (>1000 lines)
//    - Consider using line hashes for more accurate change detection
//    - Add metrics collection for performance monitoring
//
// 4. Comparison with Original:
//    - More accurate for line-level changes
//    - Simpler implementation
//    - Less memory overhead
//    - Missing commit-based filtering of changes
//    - Could be combined with blame approach for best results

func (d *DifferV2) AnalyzeChanges() ([]*FileChange, error) {
	patch, err := d.oldCommit.Patch(d.newCommit)
	if err != nil {
		// TODO: return err
		return nil, err
	}
	// Analyze changes
	// Process changes concurrently with worker pool
	filePatches := patch.FilePatches()
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
