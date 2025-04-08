package diff

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// Differ Code Difference Analyzer
type Differ struct {
	stableBranch  string
	publishBranch string
	repoPath      string
	workers       int
	repo          *git.Repository
	stableHash    plumbing.Hash
	publishHash   plumbing.Hash
	publishCommit *object.Commit
	stableCommit  *object.Commit
	commits       map[plumbing.Hash]*object.Commit
}

// NewDiffer creates a new code differ
func NewDiffer(projectPath, stableBranch, publishBranch string, workers int) (*Differ, error) {
	repo, err := git.PlainOpen(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	d := &Differ{
		repo:          repo,
		stableBranch:  stableBranch,
		publishBranch: publishBranch,
		repoPath:      projectPath,
		workers:       workers,
		commits:       make(map[plumbing.Hash]*object.Commit),
	}
	stableHash, err := d.resolveRef(stableBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve stable branch: %w", err)
	}

	publishHash, err := d.resolveRef(publishBranch)
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

func (d *Differ) loadCommits() error {
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
func (d *Differ) AnalyzeChanges() ([]*FileChange, error) {
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
		return nil, fmt.Errorf("failed to compare branch differences: %w", err)
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
func (d *Differ) analyzeChange(change *object.Change) (*FileChange, error) {
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
func (d *Differ) handleInsert(change *object.Change) (*FileChange, error) {
	fileName := change.To.Name
	if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
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

// resolveRef resolves a reference (branch name or commit hash) to a commit hash
func (d *Differ) resolveRef(ref string) (plumbing.Hash, error) {
	// Try to parse directly as hash
	hash, err := d.repo.ResolveRevision(plumbing.Revision(ref))
	if err == nil {
		return *hash, nil
	}
	// Try to resolve as branch
	refName := plumbing.NewBranchReferenceName(ref)
	hash, err = d.repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		fmt.Printf("resolveRef: %s, hash: %s\n", ref, hash.String())
		return *hash, nil
	}

	// Try to resolve as tag
	refName = plumbing.NewTagReferenceName(ref)
	hash, err = d.repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		return *hash, nil
	}

	return plumbing.ZeroHash, fmt.Errorf("unable to resolve reference: %s", ref)
}

// GetLineChanges gets line-level change information for a file, focusing only on incremental code
func (d *Differ) getLineChanges(filepath string) ([]LineChange, error) {
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
func (d *Differ) isCommitAfterStable(commitHash plumbing.Hash, stableHash plumbing.Hash) bool {
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

// GetRepoPath returns the path of the repository
func (d *Differ) GetRepoPath() string {
	return d.repoPath
}

// filterValidFileChanges filters out nil entries from the file changes slice
func filterValidFileChanges(fileChanges []*FileChange) []*FileChange {
	slow, fast := 0, 0
	for fast < len(fileChanges) {
		if fileChanges[fast] != nil {
			fileChanges[slow] = fileChanges[fast]
			slow++
		}
		fast++
	}
	return fileChanges[:slow]
}

// filterValidFileChangesOptimized evaluates the correctness of the optimized file changes filter
// 1. Correctness:
//   - Correctly filters out nil entries from the slice
//   - Maintains all non-nil entries in the result
//   - Handles edge cases (all nil, none nil, single element)
//   - Doesn't preserve original order (which is acceptable per function name)
//
// 2. Algorithm Analysis:
//   - Uses two-pointer technique (i from start, j from end)
//   - Swaps nil elements to the end of the slice
//   - Complexity: O(n) where n is length of slice
//   - More efficient than original filterValidFileChanges (single pass vs two-pass)
//   - Uses constant space (in-place modification)
//
// 3. Optimization Suggestions:
//   - Could use unsafe.Pointer for faster swaps
//   - Add early exit if no nil elements found
//   - Consider using SIMD instructions for bulk nil checks
//   - Add benchmark tests against original implementation
//   - Could parallelize nil checks for very large slices
//
// 4. Comparison with Original:
//   - Faster (O(n) vs O(2n))
//   - Doesn't preserve order (tradeoff for performance)
//   - Same memory usage (both in-place)
//   - Simpler implementation
//
// filterValidFileChangesOptimized ignore order
func filterValidFileChangesOptimized(fileChanges []*FileChange) []*FileChange {
	i, j := 0, len(fileChanges)-1
	for i <= j {
		if fileChanges[i] == nil {
			fileChanges[i], fileChanges[j] = fileChanges[j], fileChanges[i]
			j--
		} else {
			i++
		}
	}
	return fileChanges[:i]
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

func (d *Differ) AnalyzeChangesV2() ([]*FileChange, error) {
	patch, err := d.stableCommit.Patch(d.publishCommit)
	if err != nil {
		// TODO: return err
		return nil, err
	}
	// Analyze changes
	// Process changes concurrently with worker pool
	filePatches := patch.FilePatches()
	fileChanges := make([]*FileChange, len(filePatches))
	sem := make(chan struct{}, d.workers) // concurrent workers,default 10
	var wg sync.WaitGroup
	wg.Add(len(filePatches))
	for i, filePatch := range filePatches {
		sem <- struct{}{} // acquire worker slot
		go func(idx int, filePath diff.FilePatch) {
			defer func() {
				<-sem // release worker slot
				wg.Done()
			}()
			fc := d.analyzeChangeV2(filePatch)
			if fc != nil && len(fc.LineChanges) > 0 {
				fileChanges[idx] = fc
			}
		}(i, filePatch)
	}
	wg.Wait()
	return filterValidFileChangesOptimized(fileChanges), nil
}

func (d *Differ) analyzeChangeV2(filePatch diff.FilePatch) *FileChange {
	from, to := filePatch.Files()
	if (from == nil && to == nil) || (from != nil && to == nil) {
		return nil
	}
	if !strings.HasSuffix(to.Path(), ".go") || strings.HasSuffix(to.Path(), "_test.go") {
		return nil
	}
	if from == nil && to != nil {

		lineChanges := make([]LineChange, 0)
		count := getLinesFromChunks(filePatch.Chunks())
		lineChanges = append(lineChanges, LineChange{Start: 1, Lines: count})
		if len(lineChanges) == 0 {
			return nil
		}
		return &FileChange{Path: to.Path(), LineChanges: lineChanges}
	}

	lineChanges := make([]LineChange, 0)
	originLineNo := 0
	newLineNo := 0
	for _, chunk := range filePatch.Chunks() {
		lines := strings.Split(chunk.Content(), "\n")
		if chunk.Type() == diff.Add {
			start := newLineNo + 1
			count := len(lines) - 1
			lineChanges = append(lineChanges, LineChange{Start: start, Lines: count})
		}

		if chunk.Type() == diff.Equal {
			originLineNo += len(lines) - 1
			newLineNo += len(lines) - 1
		} else if chunk.Type() == diff.Delete {
			originLineNo += len(lines) - 1
		} else if chunk.Type() == diff.Add {
			newLineNo += len(lines) - 1
		}
	}
	if len(lineChanges) > 0 {
		return &FileChange{Path: to.Path(), LineChanges: lineChanges}
	}
	return nil
}

func getLinesFromChunks(chunks []diff.Chunk) int {
	count := 0
	for _, chunk := range chunks {
		count += len(strings.Split(chunk.Content(), "\n")) - 1
	}
	return count
}
