package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// defaultPrinterConfig is the default printer config
var defaultPrinterConfig = &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8, Indent: 0}

// GetAstTree parses the source code and returns the ast tree
func GetAstTree(fileName string, content []byte) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, content, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

// FormatAst formats the ast tree and returns the formatted code
func FormatAst(cfg *printer.Config, fset *token.FileSet, f *ast.File) ([]byte, error) {
	var buf bytes.Buffer
	if cfg == nil {
		cfg = defaultPrinterConfig
	}
	err := cfg.Fprint(&buf, fset, f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AddCodes adds the codes to the ast tree and returns the formatted code
func AddCodes(cfg *printer.Config, fset *token.FileSet, f *ast.File, position int, codes []string) ([]byte, error) {
	var src bytes.Buffer
	if cfg == nil {
		cfg = defaultPrinterConfig
	}
	if err := cfg.Fprint(&src, fset, f); err != nil {
		return nil, err
	}
	srcStr := src.String()
	addLen := 0
	for _, code := range codes {
		addLen += len(code) + 1
	}
	// For each insertion position, insert the print statement into the source code string
	var buf bytes.Buffer
	buf.Grow(len(srcStr) + addLen)
	lines := 0
	i := 0
	for ; i < len(srcStr); i++ {
		if lines == position-1 {
			break
		}
		if srcStr[i] == '\n' {
			lines++
		}
	}

	buf.WriteString(srcStr[:i])
	for _, code := range codes {
		buf.WriteString(code)
		buf.WriteByte('\n')
	}
	buf.WriteString(srcStr[i:])
	newFset, newF, err := GetAstTree("", buf.Bytes())
	if err != nil {
		return nil, err
	}
	content, err := FormatAst(cfg, newFset, newF)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// AddImport adds the import to the ast tree and returns the formatted code
func AddImport(cfg *printer.Config, pkgPath, alias string, filename string, content []byte) ([]byte, error) {
	if pkgPath == "" {
		return content, nil
	}
	fset, f, err := GetAstTree(filename, content)
	if err != nil {
		return nil, err
	}
	found := false
	for _, ipt := range f.Imports {
		if strings.Trim(ipt.Path.Value, "\"") == pkgPath {
			found = true
			break
		}
	}
	if !found {
		added := astutil.AddNamedImport(fset, f, alias, pkgPath)
		if !added {
			return nil, fmt.Errorf("failed to add import %s", pkgPath)
		}
	}
	content, err = FormatAst(cfg, fset, f)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// DeleteImport deletes the import from the ast tree and returns the formatted code
func DeleteImport(cfg *printer.Config, pkgPath, alias string, filename string, content []byte) ([]byte, error) {
	if pkgPath == "" {
		return content, nil
	}
	fset, f, err := GetAstTree(filename, content)
	if err != nil {
		return nil, err
	}
	found := false
	for _, ipt := range f.Imports {
		if strings.Trim(ipt.Path.Value, "\"") == pkgPath {
			found = true
			break
		}
	}
	if found {
		deleted := astutil.DeleteNamedImport(fset, f, alias, pkgPath)
		if !deleted {
			return nil, fmt.Errorf("failed to delete import %s", pkgPath)
		}
	}
	content, err = FormatAst(cfg, fset, f)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// Replace replaces the target string with the new string and returns the number of replacements and the new content
func Replace(content string, target string, replace func(older string) (newer string)) (int, string, error) {
	// Use regexp to replace the target string
	re := regexp.MustCompile(regexp.QuoteMeta(target))
	count := len(re.FindAllString(content, -1))
	if count == 0 {
		return 0, content, nil
	}
	newContent := re.ReplaceAllStringFunc(content, func(match string) string {
		return replace(match)
	})
	return count, newContent, nil
}

// ReplaceWithRegexp replaces the target string with the new string and returns the number of replacements and the new content
func ReplaceWithRegexp(re *regexp.Regexp, content string, replace func(older string) (newer string)) (int, string, error) {
	// Use regexp to replace the target string
	count := len(re.FindAllString(content, -1))
	if count == 0 {
		return 0, content, nil
	}
	newContent := re.ReplaceAllStringFunc(content, func(match string) string {
		return replace(match)
	})
	return count, newContent, nil
}

// IsGoFile checks if the file is a go file
func IsGoFile(fileName string) bool {
	if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
		return false
	}
	return true
}

// GoatPackageImportPath returns the import path of the goat package
func GoatPackageImportPath(goModule string, goatPackagePath string) string {
	return filepath.Join(goModule, goatPackagePath)
}

// Rel returns the relative path of target from base
func Rel(base string, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

// IsDirEmpty checks if a directory is empty
func IsDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // read one entry
	if err == nil {
		return false, nil // directory is not empty
	}

	return true, nil // directory is empty
}

// IsGoComment checks if the code is a golang comment
func IsGoComment(code string) bool {
	// Must regnize the following patterns:
	// 1. Single-line comment: // comment, note that "//" may have leading spaces
	// 2. Multi-line comment: /* comment */，note that "*" may have leading spaces
	// 3. Empty line is also considered as a comment
	code = strings.TrimSpace(code)
	if code == "" {
		return true
	}
	if strings.HasPrefix(code, "//") {
		return true
	}
	if strings.HasPrefix(code, "/*") || strings.HasPrefix(code, "*/") {
		return true
	}
	return false
}

// ParseComments parses the comments from the ast tree and returns the comments
func ParseComments(content []byte) (map[int]string, error) {
	fset, f, err := GetAstTree("", content)
	if err != nil {
		return nil, err
	}
	comments := make(map[int]string)
	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			start := fset.Position(comment.Pos()).Line
			text := strings.Split(comment.Text, "\n")
			for i, line := range text {
				comments[start+i] = line
			}
		}
	}
	return comments, nil
}

// FormatAndSave formats the ast tree and saves the formatted code to the file
func FormatAndSave(filename string, content []byte, cfg *printer.Config) error {
	fset, fileAst, err := GetAstTree("", content)
	if err != nil {
		return fmt.Errorf("failed to get ast tree: %v, file: %s", err, filename)
	}
	contentBytes, err := FormatAst(cfg, fset, fileAst)
	if err != nil {
		return fmt.Errorf("failed to format ast: %v, file: %s", err, filename)
	}
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v, file: %s", err, filename)
	}
	err = os.WriteFile(filename, contentBytes, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to write file: %v, file: %s", err, filename)
	}
	return nil
}
