package diff

import (
	"fmt"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
// DifferV4 Code DifferV4ence Analyzer
type DifferV4 struct {
	cfg           *config.Config
	repo          *git.Repository
	stableHash    plumbing.Hash
	publishHash   plumbing.Hash
	publishCommit *object.Commit
	stableCommit  *object.Commit
}

// NewDifferV4 creates a new code DifferV4
func NewDifferV4(cfg *config.Config) (*DifferV4, error) {
	repo, err := git.PlainOpen(cfg.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	d := &DifferV4{
		repo: repo,
		cfg:  cfg,
	}
	stableHash, err := resolveRef(d.repo, cfg.StableBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve stable branch: %w", err)
	}

	publishHash, err := resolveRef(d.repo, cfg.PublishBranch)
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
func (d *DifferV4) AnalyzeChanges() ([]*FileChange, error) {
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
	changes, err := object.DiffTree(stableTree, publishTree)

	if err != nil {
		return nil, fmt.Errorf("failed to compare branch DifferV4ences: %w", err)
	}
	// Analyze changes
	// Process changes concurrently with worker pool
	fileChanges := make([]*FileChange, len(changes))
	errChan := make(chan error, len(changes))
	sem := make(chan struct{}, d.cfg.Threads) // concurrent workers,default 10
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
func (d *DifferV4) analyzeChange(change *object.Change) (*FileChange, error) {

	action, err := change.Action()
	if err != nil {
		return nil, fmt.Errorf("failed to get change action: %w", err)
	}

	if change.From.Name != change.To.Name && change.From.Name != "" && change.To.Name != "" {
		return nil, nil
	}
	switch action {
	case merkletrie.Insert, merkletrie.Modify:
		// fmt.Println("handleInsert", "from", change.From.Name, "to", change.To.Name, action)
		return d.handleInsert(change)
	default:
		return nil, nil
	}
}

// handleInsert handles insert or modify operations
func (d *DifferV4) handleInsert(change *object.Change) (*FileChange, error) {
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

// GetRepoPath returns the path of the repository
func (d *DifferV4) GetRepoPath() string {
	return d.cfg.ProjectRoot
}
