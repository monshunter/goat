package diff

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
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

type LineChanges []LineChange

func (l LineChanges) Search(line int) int {
	for i, change := range l {
		if line >= change.Start && line <= change.Start+change.Lines-1 {
			return i
		}
	}
	return -1
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

func getLinesFromChunks(chunks []diff.Chunk) int {
	count := 0
	for _, chunk := range chunks {
		count += len(strings.Split(chunk.Content(), "\n")) - 1
	}
	return count
}

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
