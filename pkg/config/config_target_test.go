package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigIsTargetDir(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a temporary directory for testing and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		os.Chdir(originalWd)
	}()

	// Create test directory structure
	testDirs := []string{
		"pkg/utils",
		"vendor",
		"testdata",
		"node_modules",
		"pkg/testdata/utils",
		"subproject/src",
		"normal/src",
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create a nested go.mod file
	subprojectGoMod := filepath.Join("subproject", "go.mod")
	err = os.WriteFile(subprojectGoMod, []byte("module subproject\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nested go.mod: %v", err)
	}

	// Create config with validation
	cfg := &Config{
		Ignores:           []string{"custom_exclude"},
		SkipNestedModules: true,
		DiffPrecision:     2,      // Required field
		AppVersion:        "test", // Avoid git validation
	}
	err = cfg.Validate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v", err)
	}

	testCases := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "valid directory",
			dir:  "pkg/utils",
			want: true,
		},
		{
			name: "vendor directory",
			dir:  "vendor",
			want: false,
		},
		{
			name: "testdata directory",
			dir:  "testdata",
			want: false,
		},
		{
			name: "node_modules directory",
			dir:  "node_modules",
			want: false,
		},
		{
			name: "path with testdata",
			dir:  "pkg/testdata/utils",
			want: false,
		},
		{
			name: "custom excluded directory",
			dir:  "custom_exclude",
			want: false,
		},
		{
			name: "nested module directory",
			dir:  "subproject",
			want: false,
		},
		{
			name: "subdirectory of nested module",
			dir:  "subproject/src",
			want: false,
		},
		{
			name: "normal directory",
			dir:  "normal/src",
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cfg.IsTargetDir(tc.dir)
			if got != tc.want {
				t.Errorf("Config.IsTargetDir(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestConfigIsBelongNestedModule(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a temporary directory for testing and change to it
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		// Restore original working directory
		os.Chdir(originalWd)
	}()

	// Create test directory structure
	testDirs := []string{
		"subproject/src/utils",
		"subproject/src/deep/nested",
		"normal/src",
		"normal/src/utils",
	}

	for _, dir := range testDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create a nested go.mod file
	subprojectGoMod := filepath.Join("subproject", "go.mod")
	err = os.WriteFile(subprojectGoMod, []byte("module subproject\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nested go.mod: %v", err)
	}

	// Create config with validation
	cfg := &Config{
		DiffPrecision: 2,      // Required field
		AppVersion:    "test", // Avoid git validation
	}
	err = cfg.Validate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v", err)
	}

	testCases := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "directory with go.mod directly",
			dir:  "subproject",
			want: true,
		},
		{
			name: "subdirectory of nested module",
			dir:  "subproject/src",
			want: true,
		},
		{
			name: "deep subdirectory of nested module",
			dir:  "subproject/src/deep/nested",
			want: true,
		},
		{
			name: "directory without nested go.mod",
			dir:  "normal",
			want: false,
		},
		{
			name: "subdirectory without nested go.mod",
			dir:  "normal/src",
			want: false,
		},
		{
			name: "project root (should not be considered nested)",
			dir:  ".",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cfg.IsBelongNestedModule(tc.dir)
			if got != tc.want {
				t.Errorf("Config.IsBelong(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}

	// Test caching - call the same directory twice to ensure cache works
	t.Run("caching test", func(t *testing.T) {
		// First call
		result1 := cfg.IsBelongNestedModule("subproject/src")
		// Second call should use cache
		result2 := cfg.IsBelongNestedModule("subproject/src")

		if result1 != result2 {
			t.Errorf("Caching failed: first call returned %v, second call returned %v", result1, result2)
		}

		if !result1 {
			t.Errorf("Expected subproject/src to belong to nested module, got %v", result1)
		}
	})
}

func TestConfigIsTargetFile(t *testing.T) {
	cfg := &Config{
		DiffPrecision: 2,      // Required field
		AppVersion:    "test", // Avoid git validation
	}
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v", err)
	}

	testCases := []struct {
		name     string
		fileName string
		want     bool
	}{
		{
			name:     "valid go file",
			fileName: "pkg/utils/file.go",
			want:     true,
		},
		{
			name:     "test file",
			fileName: "pkg/utils/file_test.go",
			want:     false,
		},
		{
			name:     "non-go file",
			fileName: "pkg/utils/file.txt",
			want:     false,
		},
		{
			name:     "vendor file",
			fileName: "vendor/pkg/file.go",
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cfg.IsTargetFile(tc.fileName)
			if got != tc.want {
				t.Errorf("Config.IsTargetFile(%q) = %v, want %v", tc.fileName, got, tc.want)
			}
		})
	}
}
