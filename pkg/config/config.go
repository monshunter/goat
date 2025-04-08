package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config configuration struct
type Config struct {
	// Repository path
	RepoPath string `yaml:"repo_path"`
	// Stable branch name
	StableBranch string `yaml:"stable_branch"`
	// Publish branch name
	PublishBranch string `yaml:"publish_branch"`
	// Tracing function name
	TraceFuncName string `yaml:"trace_func_name"`
	// Files or directories to ignore
	Ignores []string `yaml:"ignores"`
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

	// Set default values
	if config.StableBranch == "" {
		config.StableBranch = "stable"
	}
	if config.PublishBranch == "" {
		config.PublishBranch = "publish"
	}
	if config.TraceFuncName == "" {
		config.TraceFuncName = "trace"
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(filename string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
