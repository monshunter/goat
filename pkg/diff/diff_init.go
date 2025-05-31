package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/log"
	"github.com/monshunter/goat/pkg/utils"
)

// DifferInit Code DifferInitence Analyzer
type DifferInit struct {
	cfg *config.Config
}

// NewDifferInit creates a new code DifferInit
func NewDifferInit(cfg *config.Config) (*DifferInit, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	if err := checkUncommittedChanges(repo); err != nil {
		return nil, fmt.Errorf("failed to check uncommitted changes: %w", err)
	}
	return &DifferInit{
		cfg: cfg,
	}, nil
}

// AnalyzeChanges analyzes code changes between new and initial commit
func (d *DifferInit) AnalyzeChanges() ([]*FileChange, error) {
	fileChanges := make([]*FileChange, 0)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Failed to walk: %v", err)
			return err
		}
		if info.IsDir() {
			if !utils.IsTargetDir(path, d.cfg.Ignores, d.cfg.SkipNestedModules) {
				return filepath.SkipDir
			}
			return nil
		}
		if !utils.IsGoFile(path) {
			return nil
		}
		// skip goat_generated.go
		if path == d.cfg.GoatGeneratedFile() {
			return nil
		}
		// get file content
		var content []byte
		content, err = os.ReadFile(path)
		if err != nil {
			log.Errorf("Failed to read file: %v", err)
			return err
		}
		// count lines
		lines := strings.Split(string(content), "\n")
		fileChanges = append(fileChanges, &FileChange{
			Path: utils.Rel(".", path),
			LineChanges: LineChanges{
				{
					Start: 1,
					Lines: len(lines),
				},
			},
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk: %w", err)
	}
	return fileChanges, nil
}
