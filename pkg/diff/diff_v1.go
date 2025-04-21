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
	cfg       *config.Config
	repo      *git.Repository
	oldHash   plumbing.Hash
	newHash   plumbing.Hash
	newCommit *object.Commit
	oldCommit *object.Commit
	commits   map[plumbing.Hash]*object.Commit
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
	// process input
	// resolve old branch in repo
	oldHash, err := resolveRef(d.repo, cfg.OldBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve old branch: %w", err)
	}
	// resolve new branch in repo
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
	oldTree, err := d.oldCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get old branch tree: %w", err)
	}

	newTree, err := d.newCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get new branch tree: %w", err)
	}

	// Compare trees
	changes, err := object.DiffTree(oldTree, newTree)
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
	file, err := d.newCommit.File(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	// Get file content
	lines, err := file.Lines()
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	// Get blame information for the file
	blame, err := git.Blame(d.newCommit, filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame information: %w", err)
	}
	// Use blame information to identify incremental code
	var changes []LineChange
	var currentChange *LineChange

	for i := range lines {
		line := blame.Lines[i]
		// Check if current line's commit is after old branch
		isNewLine := d.isCommitAfterStable(line.Hash, d.oldHash)
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

// isCommitAfterStable checks if the given commit is after the old branch commit
func (d *DifferV1) isCommitAfterStable(commitHash plumbing.Hash, oldHash plumbing.Hash) bool {
	// Return false if the commit is the same as old branch commit
	if commitHash == oldHash {
		return false
	}
	// Get the commit object
	commit, ok := d.commits[commitHash]
	if !ok {
		return false
	}
	return d.oldCommit.Committer.When.Before(commit.Committer.When)
}
