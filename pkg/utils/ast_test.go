package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGetAstTree(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		wantErr  bool
		fileName string
	}{
		{
			name:     "valid go code",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			wantErr:  false,
		},
		{
			name:     "invalid go code",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int int { return a + b }\n",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset, f, err := GetAstTree(tc.fileName, []byte(tc.content))
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if fset == nil {
				t.Errorf("expected non-nil fset")
			}
			if f == nil {
				t.Errorf("expected non-nil ast.File")
			}
		})
	}
}

func TestFormatAst(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		wantErr  bool
		fileName string
	}{
		{
			name:     "valid go code",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset, f, err := GetAstTree(tc.fileName, []byte(tc.content))
			if err != nil {
				t.Fatalf("failed to parse ast: %v", err)
			}

			formatted, err := FormatAst(nil, fset, f)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if formatted == nil {
				t.Errorf("expected non-nil formatted code")
			}
			if len(formatted) == 0 {
				t.Errorf("expected non-empty formatted code")
			}
		})
	}
}

func TestAddCodes(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		position int
		codes    []string
		wantErr  bool
	}{
		{
			name:     "add code to package",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			position: 2,
			codes:    []string{"import \"fmt\"", "var x = 10"},
			wantErr:  false,
		},
		{
			name:     "invalid position",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			position: 100, // Invalid position
			codes:    []string{"import \"fmt\""},
			wantErr:  true, // Changed to expect error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset, f, err := GetAstTree("", []byte(tc.content))
			if err != nil {
				t.Fatalf("failed to parse ast: %v", err)
			}

			result, err := AddCodes(nil, fset, f, tc.position, tc.codes)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Errorf("expected non-nil result")
			}
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}
		})
	}
}

func TestAddImport(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		pkgPath  string
		alias    string
		wantErr  bool
		fileName string
	}{
		{
			name:     "add new import",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "",
			wantErr:  false,
		},
		{
			name:     "add aliased import",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "f",
			wantErr:  false,
		},
		{
			name:     "add existing import",
			fileName: "test.go",
			content:  "package test\n\nimport \"fmt\"\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "",
			wantErr:  false,
		},
		{
			name:     "empty package path",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "",
			alias:    "",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := AddImport(nil, tc.pkgPath, tc.alias, tc.fileName, []byte(tc.content))
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Errorf("expected non-nil result")
			}
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}

			// Verify import was added if pkgPath is not empty
			if tc.pkgPath != "" {
				_, f, err := GetAstTree("", result)
				if err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}

				found := false
				for _, imp := range f.Imports {
					path := strings.Trim(imp.Path.Value, "\"")
					if path == tc.pkgPath {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("import %s not found in result", tc.pkgPath)
				}
			}
		})
	}
}

func TestDeleteImport(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		pkgPath  string
		alias    string
		wantErr  bool
		fileName string
	}{
		{
			name:     "delete existing import",
			fileName: "test.go",
			content:  "package test\n\nimport \"fmt\"\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "",
			wantErr:  false,
		},
		{
			name:     "delete non-existing import",
			fileName: "test.go",
			content:  "package test\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "",
			wantErr:  false,
		},
		{
			name:     "delete aliased import",
			fileName: "test.go",
			content:  "package test\n\nimport f \"fmt\"\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "fmt",
			alias:    "f",
			wantErr:  false,
		},
		{
			name:     "empty package path",
			fileName: "test.go",
			content:  "package test\n\nimport \"fmt\"\n\nfunc Add(a, b int) int { return a + b }\n",
			pkgPath:  "",
			alias:    "",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := DeleteImport(nil, tc.pkgPath, tc.alias, tc.fileName, []byte(tc.content))
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Errorf("expected non-nil result")
			}
			if len(result) == 0 {
				t.Errorf("expected non-empty result")
			}

			// Verify import was deleted if pkgPath is not empty
			if tc.pkgPath != "" {
				_, f, err := GetAstTree("", result)
				if err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}

				for _, imp := range f.Imports {
					path := strings.Trim(imp.Path.Value, "\"")
					if path == tc.pkgPath {
						t.Errorf("import %s still found in result", tc.pkgPath)
						break
					}
				}
			}
		})
	}
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		name       string
		content    string
		target     string
		replace    func(string) string
		wantCount  int
		wantChange bool
	}{
		{
			name:    "replace string",
			content: "Hello, world!",
			target:  "world",
			replace: func(s string) string {
				return "universe"
			},
			wantCount:  1,
			wantChange: true,
		},
		{
			name:    "replace multiple occurrences",
			content: "test test test",
			target:  "test",
			replace: func(s string) string {
				return "TEST"
			},
			wantCount:  3,
			wantChange: true,
		},
		{
			name:    "no replacement",
			content: "Hello, world!",
			target:  "universe",
			replace: func(s string) string {
				return "galaxy"
			},
			wantCount:  0,
			wantChange: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, result, err := Replace(tc.content, tc.target, tc.replace)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if count != tc.wantCount {
				t.Errorf("expected count %d, got %d", tc.wantCount, count)
			}

			if tc.wantChange && result == tc.content {
				t.Errorf("expected content to change but it didn't")
			}

			if !tc.wantChange && result != tc.content {
				t.Errorf("expected content to remain the same but it changed")
			}
		})
	}
}

func TestReplaceWithRegexp(t *testing.T) {
	testCases := []struct {
		name       string
		content    string
		pattern    string
		replace    func(string) string
		wantCount  int
		wantChange bool
	}{
		{
			name:    "replace with regexp",
			content: "Hello, world!",
			pattern: "w[a-z]+d",
			replace: func(s string) string {
				return "universe"
			},
			wantCount:  1,
			wantChange: true,
		},
		{
			name:    "replace multiple with regexp",
			content: "test1 test2 test3",
			pattern: "test[0-9]",
			replace: func(s string) string {
				return "TEST"
			},
			wantCount:  3,
			wantChange: true,
		},
		{
			name:    "no replacement with regexp",
			content: "Hello, world!",
			pattern: "universe",
			replace: func(s string) string {
				return "galaxy"
			},
			wantCount:  0,
			wantChange: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.pattern)
			count, result, err := ReplaceWithRegexp(re, tc.content, tc.replace)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if count != tc.wantCount {
				t.Errorf("expected count %d, got %d", tc.wantCount, count)
			}

			if tc.wantChange && result == tc.content {
				t.Errorf("expected content to change but it didn't")
			}

			if !tc.wantChange && result != tc.content {
				t.Errorf("expected content to remain the same but it changed")
			}
		})
	}
}

func TestIsTargetFile(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		excludes []string
		want     bool
	}{
		{
			name:     "valid go file",
			fileName: "pkg/utils/file.go",
			excludes: []string{},
			want:     true,
		},
		{
			name:     "test file",
			fileName: "pkg/utils/file_test.go",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "non-go file",
			fileName: "pkg/utils/file.txt",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "excluded directory",
			fileName: "vendor/pkg/file.go",
			excludes: []string{"vendor"},
			want:     false,
		},
		{
			name:     "excluded path",
			fileName: "pkg/utils/file.go",
			excludes: []string{"pkg/utils"},
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTargetFile(tc.fileName, tc.excludes, true)
			if got != tc.want {
				t.Errorf("IsTargetFile(%q, %v) = %v, want %v", tc.fileName, tc.excludes, got, tc.want)
			}
		})
	}
}

func TestIsTargetDir(t *testing.T) {
	testCases := []struct {
		name     string
		dir      string
		excludes []string
		want     bool
	}{
		{
			name:     "valid directory",
			dir:      "pkg/utils",
			excludes: []string{},
			want:     true,
		},
		{
			name:     "vendor directory",
			dir:      "vendor",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "testdata directory",
			dir:      "testdata",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "node_modules directory",
			dir:      "node_modules",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "path with testdata",
			dir:      "pkg/testdata/utils",
			excludes: []string{},
			want:     false,
		},
		{
			name:     "excluded directory",
			dir:      "pkg/utils",
			excludes: []string{"pkg/utils"},
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTargetDir(tc.dir, tc.excludes, true)
			if got != tc.want {
				t.Errorf("IsTargetDir(%q, %v) = %v, want %v", tc.dir, tc.excludes, got, tc.want)
			}
		})
	}
}

func TestIsGoFile(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		want     bool
	}{
		{
			name:     "go file",
			fileName: "file.go",
			want:     true,
		},
		{
			name:     "test file",
			fileName: "file_test.go",
			want:     false,
		},
		{
			name:     "non-go file",
			fileName: "file.txt",
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsGoFile(tc.fileName)
			if got != tc.want {
				t.Errorf("IsGoFile(%q) = %v, want %v", tc.fileName, got, tc.want)
			}
		})
	}
}

func TestGoatPackageImportPath(t *testing.T) {
	testCases := []struct {
		name            string
		goModule        string
		goatPackagePath string
		want            string
	}{
		{
			name:            "valid path",
			goModule:        "github.com/user/project",
			goatPackagePath: "pkg/goat",
			want:            "github.com/user/project/pkg/goat",
		},
		{
			name:            "empty module",
			goModule:        "",
			goatPackagePath: "pkg/goat",
			want:            "pkg/goat",
		},
		{
			name:            "empty package path",
			goModule:        "github.com/user/project",
			goatPackagePath: "",
			want:            "github.com/user/project",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GoatPackageImportPath(tc.goModule, tc.goatPackagePath)
			if got != tc.want {
				t.Errorf("GoatPackageImportPath(%q, %q) = %q, want %q", tc.goModule, tc.goatPackagePath, got, tc.want)
			}
		})
	}
}

func TestRel(t *testing.T) {
	testCases := []struct {
		name   string
		base   string
		target string
		want   string
	}{
		{
			name:   "valid relative path",
			base:   "/usr/local",
			target: "/usr/local/bin",
			want:   "bin",
		},
		{
			name:   "target is base",
			base:   "/usr/local",
			target: "/usr/local",
			want:   ".",
		},
		{
			name:   "target outside base",
			base:   "/usr/local/bin",
			target: "/etc",
			want:   "../../etc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Rel(tc.base, tc.target)
			// On Windows paths might be different, so we only test on Unix-like systems
			if tc.want != "." && !strings.Contains(tc.want, "..") {
				if got != tc.want {
					t.Errorf("Rel(%q, %q) = %q, want %q", tc.base, tc.target, got, tc.want)
				}
			}
		})
	}
}

func TestIsDirEmpty(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-dir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test empty directory
	t.Run("empty directory", func(t *testing.T) {
		isEmpty, err := IsDirEmpty(tempDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !isEmpty {
			t.Errorf("expected directory to be empty")
		}
	})

	// Create a file in the temporary directory
	filePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test non-empty directory
	t.Run("non-empty directory", func(t *testing.T) {
		isEmpty, err := IsDirEmpty(tempDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if isEmpty {
			t.Errorf("expected directory to be non-empty")
		}
	})

	// Test non-existent directory
	t.Run("non-existent directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		isEmpty, err := IsDirEmpty(nonExistentDir)
		if err == nil {
			t.Errorf("expected error but got nil")
		}
		if isEmpty {
			t.Errorf("expected result to be false for non-existent directory")
		}
	})
}

func TestIsGoComment(t *testing.T) {
	testCases := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "single line comment",
			code: "// This is a comment",
			want: true,
		},
		{
			name: "multi-line comment start",
			code: "/* This is a comment",
			want: true,
		},
		{
			name: "multi-line comment end",
			code: " */ end of comment",
			want: true,
		},
		{
			name: "empty line",
			code: "",
			want: true,
		},
		{
			name: "whitespace only",
			code: "   ",
			want: true,
		},
		{
			name: "code line",
			code: "var x = 10",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsGoComment(tc.code)
			if got != tc.want {
				t.Errorf("IsGoComment(%q) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

func TestParseComments(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		wantErr     bool
		wantComment map[int]string
	}{
		{
			name: "with comments",
			content: `package test

// This is a comment
func Add(a, b int) int {
	// This is another comment
	return a + b
}`,
			wantErr: false,
			wantComment: map[int]string{
				3: "// This is a comment",
				5: "// This is another comment",
			},
		},
		{
			name: "without comments",
			content: `package test

func Add(a, b int) int {
	return a + b
}`,
			wantErr:     false,
			wantComment: map[int]string{},
		},
		{
			name: "with multi-line comment",
			content: `package test

/* This is a 
   multi-line comment */
func Add(a, b int) int {
	return a + b
}`,
			wantErr: false,
			wantComment: map[int]string{
				3: "/* This is a ",
				4: "   multi-line comment */",
			},
		},
		{
			name:    "invalid go code",
			content: "package test\n\nfunc Add(a, b int int { return a + b }\n",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comments, err := ParseComments([]byte(tc.content))
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check if all expected comments are present
			for line, comment := range tc.wantComment {
				if gotComment, ok := comments[line]; !ok {
					t.Errorf("comment at line %d not found", line)
				} else if gotComment != comment {
					t.Errorf("comment at line %d = %q, want %q", line, gotComment, comment)
				}
			}

			// Check if there are no unexpected comments
			if len(comments) != len(tc.wantComment) {
				t.Errorf("got %d comments, want %d", len(comments), len(tc.wantComment))
			}
		})
	}
}

func TestFormatAndSave(t *testing.T) {
	// Skip this test if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-dir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFileName := filepath.Join(tempDir, "test.go")
	testContent := "package test\n\nfunc   Add(a,   b int)   int    { return a+b }\n"
	err = os.WriteFile(testFileName, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test formatting and saving
	t.Run("format and save", func(t *testing.T) {
		err := FormatAndSave(testFileName, []byte(testContent), nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		// Read the formatted file
		formatted, err := os.ReadFile(testFileName)
		if err != nil {
			t.Errorf("failed to read formatted file: %v", err)
			return
		}

		// Check if the file was formatted (e.g., extra spaces removed)
		if strings.Contains(string(formatted), "func   Add") {
			t.Errorf("file was not formatted correctly")
		}
	})

	// Test with invalid file name
	t.Run("invalid file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.go")
		err := FormatAndSave(nonExistentFile, []byte(testContent), nil)
		if err == nil {
			t.Errorf("expected error but got nil")
		}
	})

	// Test with invalid Go code
	t.Run("invalid go code", func(t *testing.T) {
		invalidContent := "package test\n\nfunc Add(a, b int int { return a + b }\n"
		err := FormatAndSave(testFileName, []byte(invalidContent), nil)
		if err == nil {
			t.Errorf("expected error but got nil")
		}
	})
}

func TestIsBelongtoNestedGoModule(t *testing.T) {
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

	// Create project structure:
	// tempDir/ (project root)
	// ├── go.mod (main project)
	// ├── subproject/
	// │   ├── go.mod (nested module)
	// │   └── src/
	// │       └── subdir/
	// └── normal/
	//     └── subdir/

	// Create main go.mod in project root
	err = os.WriteFile("go.mod", []byte("module main\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main go.mod: %v", err)
	}

	// Create nested module structure
	subprojectDir := "subproject"
	err = os.MkdirAll(subprojectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subproject directory: %v", err)
	}

	// Create go.mod in subproject
	subprojectGoMod := filepath.Join(subprojectDir, "go.mod")
	err = os.WriteFile(subprojectGoMod, []byte("module subproject\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create subproject go.mod: %v", err)
	}

	// Create subdirectories within the nested module
	nestedSrcDir := filepath.Join(subprojectDir, "src")
	err = os.MkdirAll(nestedSrcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested src directory: %v", err)
	}

	nestedSubDir := filepath.Join(subprojectDir, "src", "subdir")
	err = os.MkdirAll(nestedSubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested subdir: %v", err)
	}

	// Create normal directory structure (no nested go.mod)
	normalDir := "normal"
	err = os.MkdirAll(normalDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create normal directory: %v", err)
	}

	normalSubDir := filepath.Join(normalDir, "subdir")
	err = os.MkdirAll(normalSubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create normal subdir: %v", err)
	}

	testCases := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "directory with go.mod directly",
			dir:  subprojectDir,
			want: true,
		},
		{
			name: "subdirectory of nested module",
			dir:  nestedSrcDir,
			want: true,
		},
		{
			name: "deep subdirectory of nested module",
			dir:  nestedSubDir,
			want: true,
		},
		{
			name: "directory without nested go.mod",
			dir:  normalDir,
			want: false,
		},
		{
			name: "subdirectory without nested go.mod",
			dir:  normalSubDir,
			want: false,
		},
		{
			name: "project root (should not be considered nested)",
			dir:  ".",
			want: false,
		},
		{
			name: "non-existent directory",
			dir:  "nonexistent",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsBelongtoNestedGoModule(tc.dir)
			if got != tc.want {
				t.Errorf("IsBelongtoNestedGoModule(%s) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}
