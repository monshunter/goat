package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func GetAstTree(fileName string, content []byte) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, content, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

func FormatAst(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8, Indent: 0}
	err := cfg.Fprint(&buf, fset, f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func AddCodes(fset *token.FileSet, f *ast.File, position int, codes []string) ([]byte, error) {
	var src bytes.Buffer
	if err := printer.Fprint(&src, fset, f); err != nil {
		return nil, err
	}
	srcStr := src.String()
	addLen := 0
	for _, code := range codes {
		addLen += len(code) + 1
	}
	// 对于每个插入位置，将打印语句插入到源代码字符串中
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
	content, err := FormatAst(newFset, newF)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func AddImport(pkgPath, alias string, filename string, content []byte) ([]byte, error) {
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
	content, err = FormatAst(fset, f)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func DeleteImport(pkgPath, alias string, filename string, content []byte) ([]byte, error) {
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
	content, err = FormatAst(fset, f)
	if err != nil {
		return nil, err
	}
	return content, nil
}

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

func IsTargetFile(fileName string, excludes []string) bool {
	return IsTargetDir(filepath.Dir(fileName), excludes) && IsGoFile(fileName)
}

func IsTargetDir(dir string, excludes []string) bool {
	// check if the dir is in the excludes
	if dir == "vendor" || dir == "testdata" || dir == "node_modules" {
		return false
	}
	// check if the dir is in the excludes
	segments := strings.Split(dir, "/")
	for _, segment := range segments {
		if segment == "testdata" {
			return false
		}
	}
	// check if the file is in the excludes
	for _, exclude := range excludes {
		if dir == exclude || strings.HasPrefix(dir, exclude) {
			return false
		}
	}
	return true
}

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
