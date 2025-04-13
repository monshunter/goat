package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Granularity int

const (
	GranularityLine  Granularity = 1
	GranularityBlock Granularity = 2
	GranularityFunc  Granularity = 3
)

const (
	GranularityLineStr  = "line"
	GranularityBlockStr = "block"
	GranularityFuncStr  = "func"
)

func (g Granularity) String() string {
	return []string{GranularityLineStr, GranularityBlockStr, GranularityFuncStr}[g-1]
}

func (g Granularity) Int() int {
	return int(g)
}

func (g Granularity) IsLine() bool {
	return g == GranularityLine
}

func (g Granularity) IsBlock() bool {
	return g == GranularityBlock
}

func (g Granularity) IsFunc() bool {
	return g == GranularityFunc
}

// Config configuration struct
type Config struct {
	// Root path
	ProjectRoot string `yaml:"projectRoot"` // absolute path
	// Stable branch name
	StableBranch string `yaml:"stableBranch"` // commit hash or branch name or tag name or .
	// Publish branch name
	PublishBranch string `yaml:"publishBranch"` // commit hash or branch name or tag name or .
	// Files or directories to ignore
	Ignores []string `yaml:"ignores"`
	// Goat package name
	GoatPackageName string `yaml:"goatPackageName"`
	// Goat package alias
	GoatPackageAlias string `yaml:"goatPackageAlias"`
	// Goat package path
	GoatPackagePath string `yaml:"goatPackagePath"`
	// Granularity
	Granularity *string `yaml:"granularity"` // line, block, func
	// Precision
	DiffPrecision *int `yaml:"diffPrecision"` // 1~4
	// Threads
	Threads int `yaml:"threads"` // 1~128
	// Race
	Race bool `yaml:"race"` // true, false
	// Clone branch
	CloneBranch bool `yaml:"cloneBranch"` // true, false
	// Main packages to track
	MainPackages []string `yaml:"mainPackages"`
	// Main package coverage strategy
	MainPackageTrackStrategy string `yaml:"mainPackageTrackStrategy"` // all, package // default: all
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	// Repo path check
	if config.ProjectRoot == "" {
		return nil, fmt.Errorf("project root is required")
	}
	projectRoot, err := filepath.Abs(config.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	config.ProjectRoot = projectRoot

	// Set default values
	if config.StableBranch == "" {
		return nil, fmt.Errorf("stable branch is required")
	}
	if config.PublishBranch == "" {
		return nil, fmt.Errorf("publish branch is required")
	}
	// Granularity check
	if config.Granularity != nil {
		switch *config.Granularity {
		case GranularityLineStr, GranularityBlockStr, GranularityFuncStr:
		default:
			return nil, fmt.Errorf("invalid granularity: %s", *config.Granularity)
		}
	}

	// Diff precision check
	if config.DiffPrecision != nil {
		if *config.DiffPrecision < 1 || *config.DiffPrecision > 4 {
			return nil, fmt.Errorf("invalid diff precision: %d", *config.DiffPrecision)
		}
	}

	// Threads check
	if config.Threads == 0 {
		config.Threads = runtime.NumCPU()
	}

	return &config, nil
}

func Init(projectPath string) error {
	cfg := fmt.Sprintf(CONFIG_TEMPLATE, projectPath)
	return os.WriteFile(filepath.Join(projectPath, ".goat", "config.yaml"), []byte(cfg), 0644)
}
