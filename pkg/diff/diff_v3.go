package diff

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// Due to go-git, this version will panic in a multi-threaded scenario.
// https://github.com/go-git/go-git/pull/186
// DifferV3 Code DifferV3ence Analyzer
type DifferV3 struct {
	stableBranch  string
	publishBranch string
	repoPath      string
	workers       int
	repo          *git.Repository
	stableHash    plumbing.Hash
	publishHash   plumbing.Hash
	publishCommit *object.Commit
	stableCommit  *object.Commit
}

// NewDifferV3 creates a new code DifferV3
func NewDifferV3(projectPath, stableBranch, publishBranch string, workers int) (*DifferV3, error) {
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	d := &DifferV3{
		repo:          repo,
		stableBranch:  stableBranch,
		publishBranch: publishBranch,
		repoPath:      projectPath,
		workers:       workers,
	}
	stableHash, err := resolveRef(d.repo, stableBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve stable branch: %w", err)
	}

	publishHash, err := resolveRef(d.repo, publishBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve publish branch: %w", err)
	}

	stableCommit, err := d.repo.CommitObject(stableHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get stable branch commit: %w", err)
	}

	publishCommit, err := d.repo.CommitObject(publishHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get publish branch commit: %w", err)
	}
	d.stableHash = stableHash
	d.publishHash = publishHash
	d.stableCommit = stableCommit
	d.publishCommit = publishCommit
	return d, nil
}

// AnalyzeChanges analyzes code changes between two branches
func (d *DifferV3) AnalyzeChanges() ([]*FileChange, error) {
	// Get trees for both commits
	stableTree, err := d.stableCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get stable branch tree: %w", err)
	}

	publishTree, err := d.publishCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get publish branch tree: %w", err)
	}

	// Compare trees
	// changes, err := object.DiffTree(stableTree, publishTree)
	changes, err := object.DiffTreeWithOptions(context.Background(), stableTree, publishTree, object.DefaultDiffTreeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branch DifferV3ences: %w", err)
	}
	// Analyze changes
	// Process changes concurrently with worker pool
	fileChanges := make([]*FileChange, len(changes))
	errChan := make(chan error, len(changes))
	sem := make(chan struct{}, d.workers) // concurrent workers,default 10
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
	if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
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

// GetRepoPath returns the path of the repository
func (d *DifferV3) GetRepoPath() string {
	return d.repoPath
}
