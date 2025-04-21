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

// DifferV1 Code DifferV1ence Analyzer
type DifferV1 struct {
	cfg           *config.Config
	repo          *git.Repository
	stableHash    plumbing.Hash
	publishHash   plumbing.Hash
	publishCommit *object.Commit
	stableCommit  *object.Commit
	commits       map[plumbing.Hash]*object.Commit
}

// NewDifferV1 creates a new code DifferV1
func NewDifferV1(cfg *config.Config) (*DifferV1, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Check if there are uncommitted changes
	if err := checkUncommittedChanges(repo); err != nil {
		return nil, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}

	d := &DifferV1{
		repo:    repo,
		cfg:     cfg,
		commits: make(map[plumbing.Hash]*object.Commit),
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
	d.loadCommits()
	return d, nil
}

func (d *DifferV1) loadCommits() error {
	commits, err := d.repo.CommitObjects()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}
	commits.ForEach(func(commit *object.Commit) error {
		d.commits[commit.Hash] = commit
		return nil
	})
	return nil
}

// AnalyzeChanges analyzes code changes between two branches
func (d *DifferV1) AnalyzeChanges() ([]*FileChange, error) {
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
		return nil, fmt.Errorf("failed to compare branch DifferV1ences: %w", err)
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
	close(sem)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	return filterValidFileChanges(fileChanges), nil
}

// analyzeChange analyzes a single file change
func (d *DifferV1) analyzeChange(change *object.Change) (*FileChange, error) {
	action, err := change.Action()
	if err != nil {
		return nil, fmt.Errorf("failed to get change action: %w", err)
	}

	// Check if this is a rename operation
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
func (d *DifferV1) handleInsert(change *object.Change) (*FileChange, error) {
	fileName := change.To.Name
	if !utils.IsTargetFile(fileName, d.cfg.Ignores) {
		return nil, nil
	}
	fc := FileChange{
		Path: fileName,
	}
	lineChanges, err := d.getLineChanges(fileName)
	if err != nil {
		return nil, fmt.Errorf("get lines change failed: %w", err)
	}
	fc.LineChanges = lineChanges
	return &fc, nil
}

// GetLineChanges gets line-level change information for a file, focusing only on incremental code
func (d *DifferV1) getLineChanges(filepath string) ([]LineChange, error) {
	// Get file content
	file, err := d.publishCommit.File(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	// Get file content
	lines, err := file.Lines()
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	// Get blame information for the file
	blame, err := git.Blame(d.publishCommit, filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame information: %w", err)
	}
	// Use blame information to identify incremental code
	var changes []LineChange
	var currentChange *LineChange

	for i := range lines {
		line := blame.Lines[i]
		// Check if current line's commit is after stable branch
		isNewLine := d.isCommitAfterStable(line.Hash, d.stableHash)
		if isNewLine {
			if currentChange == nil {
				currentChange = &LineChange{
					Start: i + 1,
					Lines: 1,
				}
			} else if currentChange.Start+currentChange.Lines == i+1 {
				// Continuous incremental code lines, merge into current change
				currentChange.Lines++
			} else {
				// Discontinuous incremental code lines, create new change
				changes = append(changes, *currentChange)
				currentChange = &LineChange{
					Start: i + 1,
					Lines: 1,
				}
			}
		} else if currentChange != nil {
			changes = append(changes, *currentChange)
			currentChange = nil
		}
	}

	if currentChange != nil {
		changes = append(changes, *currentChange)
	}
	return changes, nil
}

// isCommitAfterStable checks if the given commit is after the stable branch commit
func (d *DifferV1) isCommitAfterStable(commitHash plumbing.Hash, stableHash plumbing.Hash) bool {
	// Return false if the commit is the same as stable branch commit
	if commitHash == stableHash {
		return false
	}
	// Get the commit object
	commit, ok := d.commits[commitHash]
	if !ok {
		return false
	}
	return d.stableCommit.Committer.When.Before(commit.Committer.When)
}
