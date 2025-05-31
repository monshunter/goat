package diff

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/monshunter/goat/pkg/config"
)

type DifferInterface interface {
	AnalyzeChanges() ([]*FileChange, error)
}

// FileChange represents file change information
type FileChange struct {
	Path        string      `json:"path"` // file path
	LineChanges LineChanges `json:"line_changes"`
}

// LineChange represents line-level change information
type LineChange struct {
	Start int `json:"start"` // starting line number of new code
	Lines int `json:"lines"` // number of lines of new code
}

// LineChanges is a list of line changes
type LineChanges []LineChange

// Search searches for a line change
func (l LineChanges) Search(line int) int {
	for i, change := range l {
		if line >= change.Start && line <= change.Start+change.Lines-1 {
			return i
		}
	}
	return -1
}

// repoInfo is the repository information
type repoInfo struct {
	repo      *git.Repository
	oldHash   plumbing.Hash
	newHash   plumbing.Hash
	oldCommit *object.Commit
	newCommit *object.Commit
	commits   map[plumbing.Hash]*object.Commit
}

// newRepoInfo creates a new repoInfo
func newRepoInfo(oldBranch, newBranch string) (*repoInfo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	if err := checkUncommittedChanges(repo); err != nil {
		return nil, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}

	oldHash, err := resolveRef(repo, oldBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve old branch: %w", err)
	}
	newHash, err := resolveRef(repo, newBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve new branch: %w", err)
	}

	// check if newBranch is the same as the current HEAD
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	if newHash != headRef.Hash() {
		return nil, fmt.Errorf("new branch(%s) is not the same as the current HEAD, please switch to the correct commit point before running the operation", newBranch)
	}

	oldCommit, err := repo.CommitObject(oldHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get old branch commit: %w", err)
	}

	newCommit, err := repo.CommitObject(newHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get new branch commit: %w", err)
	}

	// Because we have checked the newHash is the same as the current HEAD,
	// so we don't need to check if oldCommit is an ancestor of newCommit
	// sometimes the old commit is not an ancestor of new commit, but we still need to compare them
	// if isAncestor, err := oldCommit.IsAncestor(newCommit); err != nil {
	// 	return nil, fmt.Errorf("failed to check if old commit is an ancestor of new commit: %w", err)
	// } else if !isAncestor {
	// 	return nil, fmt.Errorf("old commit is not an ancestor of new commit")
	// }

	return &repoInfo{
		repo:      repo,
		oldHash:   oldHash,
		newHash:   newHash,
		oldCommit: oldCommit,
		newCommit: newCommit,
	}, nil
}

// getOldCommit returns the old commit
func (r *repoInfo) getOldCommit() *object.Commit {
	return r.oldCommit
}

// getNewCommit returns the new commit
func (r *repoInfo) getNewCommit() *object.Commit {
	return r.newCommit
}

// getRepo returns the repository
func (r *repoInfo) getRepo() *git.Repository {
	return r.repo
}

// getOldHash returns the old commit hash
func (r *repoInfo) getOldHash() plumbing.Hash {
	return r.oldHash
}

// getNewHash returns the new commit hash
func (r *repoInfo) getNewHash() plumbing.Hash {
	return r.newHash
}

// getObjectChanges returns the object changes
func (r *repoInfo) getObjectChanges() (object.Changes, error) {
	// Get trees for both commits
	oldTree, err := r.oldCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get old branch tree: %w", err)
	}

	newTree, err := r.newCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get new branch tree: %w", err)
	}

	// Compare trees
	changes, err := object.DiffTree(oldTree, newTree)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branch DifferV1ences: %w", err)
	}
	return changes, nil
}

// getFilePatches returns the file patches
func (r *repoInfo) getFilePatches() ([]diff.FilePatch, error) {
	patch, err := r.oldCommit.Patch(r.newCommit)
	if err != nil {
		return nil, err
	}
	// Analyze changes
	// Process changes concurrently with worker pool
	filePatches := patch.FilePatches()
	return filePatches, nil
}

// loadCommits loads the commits
func (r *repoInfo) loadCommits() error {
	if r.commits != nil {
		return nil
	}
	r.commits = make(map[plumbing.Hash]*object.Commit)
	commits, err := r.repo.CommitObjects()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}
	commits.ForEach(func(commit *object.Commit) error {
		r.commits[commit.Hash] = commit
		return nil
	})
	return nil
}

// resolveRef resolves a reference (branch name or commit hash) to a commit hash
func resolveRef(repo *git.Repository, ref string) (plumbing.Hash, error) {
	// Try to parse directly as hash
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err == nil {
		return *hash, nil
	}
	// Try to resolve as branch
	refName := plumbing.NewBranchReferenceName(ref)
	hash, err = repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		fmt.Printf("resolveRef: %s, hash: %s\n", ref, hash.String())
		return *hash, nil
	}

	// Try to resolve as tag
	refName = plumbing.NewTagReferenceName(ref)
	hash, err = repo.ResolveRevision(plumbing.Revision(refName))
	if err == nil {
		return *hash, nil
	}

	return plumbing.ZeroHash, fmt.Errorf("unable to resolve reference: %s", ref)
}

// filterValidFileChanges filters out nil entries from the file changes slice
// func filterValidFileChanges(fileChanges []*FileChange) []*FileChange {
// 	slow, fast := 0, 0
// 	for fast < len(fileChanges) {
// 		if fileChanges[fast] != nil {
// 			fileChanges[slow] = fileChanges[fast]
// 			slow++
// 		}
// 		fast++
// 	}
// 	return fileChanges[:slow]
// }

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
func filterValidFileChanges(fileChanges []*FileChange) []*FileChange {
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

// getLinesFromChunks gets the lines from the chunks
func getLinesFromChunks(chunks []diff.Chunk) int {
	count := 0
	for _, chunk := range chunks {
		count += len(strings.Split(chunk.Content(), "\n")) - 1
	}
	return count
}

// getLineChange gets the line changes from the file patch
func getLineChange(filePatch diff.FilePatch) []LineChange {
	from, to := filePatch.Files()
	if (from == nil && to == nil) || (from != nil && to == nil) {
		return nil
	}
	if from == nil && to != nil {
		lineChanges := make([]LineChange, 0)
		count := getLinesFromChunks(filePatch.Chunks())
		lineChanges = append(lineChanges, LineChange{Start: 1, Lines: count})
		if len(lineChanges) == 0 {
			return nil
		}
		return lineChanges
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
	return lineChanges
}

// checkUncommittedChanges checks if there are uncommitted changes in the working directory
func checkUncommittedChanges(repo *git.Repository) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if !status.IsClean() {
		// Check if there are all untracked files
		hasNonUntracked := false
		for _, fileStatus := range status {
			if fileStatus.Staging != git.Untracked || fileStatus.Worktree != git.Untracked {
				hasNonUntracked = true
				break
			}
		}
		if !hasNonUntracked {
			return nil
		}
		// Check if only config.ConfigYaml is modified
		configModifiedOnly := true
		for filePath, fileStatus := range status {
			if filePath != config.ConfigYaml &&
				(fileStatus.Staging != git.Unmodified || fileStatus.Worktree != git.Unmodified) {
				configModifiedOnly = false
				break
			}
		}
		if configModifiedOnly {
			return nil
		}
		return fmt.Errorf("there are uncommitted changes in the working directory")
	}
	return nil
}
