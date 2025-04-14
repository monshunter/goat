package config

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

type Granularity int

const (
	_ Granularity = iota
	GranularityLine
	GranularityBlock
	GranularityFunc
)

const (
	GranularityLineStr  = "line"
	GranularityBlockStr = "block"
	GranularityFuncStr  = "func"
)

func ToGranularity(s string) (Granularity, error) {
	switch s {
	case GranularityLineStr:
		return GranularityLine, nil
	case GranularityBlockStr:
		return GranularityBlock, nil
	case GranularityFuncStr:
		return GranularityFunc, nil
	default:
		return 0, fmt.Errorf("invalid granularity: %s", s)
	}
}

func (g Granularity) IsValid() bool {
	return g == GranularityLine || g == GranularityBlock || g == GranularityFunc
}

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
	// App name
	AppName string `yaml:"appName"` // goat
	// App version
	AppVersion string `yaml:"appVersion"` // 1.0.0
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
	Granularity string `yaml:"granularity"` // line, block, func
	// Diff precision
	DiffPrecision int `yaml:"diffPrecision"` // 1~2, 3&4 is not supported
	// Threads
	Threads int `yaml:"threads"` // 1~128
	// Race
	Race bool `yaml:"race"` // true, false
	// Clone branch
	CloneBranch bool `yaml:"cloneBranch"` // true, false
	// Main packages to track
	MainEntries []string `yaml:"mainEntries"`
	// Main package coverage strategy
	TrackStrategy string `yaml:"trackStrategy"` // project, package // default: project

}

func (c *Config) Validate() error {
	if c.Granularity == "" {
		c.Granularity = GranularityBlockStr
	}
	_, err := ToGranularity(c.Granularity)
	if err != nil {
		return fmt.Errorf("invalid granularity: %w", err)
	}

	if c.DiffPrecision < 1 || c.DiffPrecision > 2 {
		return fmt.Errorf("invalid diff precision: %d", c.DiffPrecision)
	}

	if c.Threads <= 0 {
		c.Threads = runtime.NumCPU()
	}

	if c.ProjectRoot == "" {
		return fmt.Errorf("project root is required")
	}

	projectRoot, err := filepath.Abs(c.ProjectRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	c.ProjectRoot = projectRoot

	if c.StableBranch == "" {
		c.StableBranch = "main"
	}

	if c.PublishBranch == "" {
		c.PublishBranch = "HEAD"
	}

	if c.Ignores == nil {
		c.Ignores = []string{".git", ".gitignore", ".DS_Store", ".idea", ".vscode", ".venv"}
	}

	if c.MainEntries == nil {
		c.MainEntries = []string{"*"}
	}

	if c.TrackStrategy == "" {
		c.TrackStrategy = "project"
	}

	if c.TrackStrategy != "project" && c.TrackStrategy != "package" {
		return fmt.Errorf("invalid track strategy: %s", c.TrackStrategy)
	}

	if c.GoatPackageName == "" {
		c.GoatPackageName = "goat"
	}

	if c.GoatPackageAlias == "" {
		c.GoatPackageAlias = "goat"
	}

	if c.GoatPackagePath == "" {
		c.GoatPackagePath = "goat"
	}
	return nil
}

func (c *Config) GetGranularity() Granularity {
	granularity, err := ToGranularity(c.Granularity)
	if err != nil {
		return GranularityBlock
	}
	return granularity
}

func (c *Config) IsMainEntry(entry string) bool {
	for _, mainEntry := range c.MainEntries {
		if mainEntry == "*" || mainEntry == entry {
			return true
		}
	}
	return false
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
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	return &config, nil
}

// InitWithConfig initializes configuration with a Config struct
func InitWithConfig(filename string, cfg *Config) error {
	// parse config template
	tmpl, err := template.New("config").Parse(CONFIG_TEMPLATE)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// create config file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// execute template
	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	return nil
}

// GetGoModuleName gets the module name from the go.mod file
func GoModuleName(root string) string {
	modFilePath := filepath.Join(root, "go.mod")
	content, err := os.ReadFile(modFilePath)
	if err != nil {
		return ""
	}
	modFile, err := modfile.Parse(modFilePath, content, nil)
	if err != nil {
		return ""
	}
	return modFile.Module.Mod.Path
}
