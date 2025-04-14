package utils

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
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
		astutil.AddNamedImport(fset, f, alias, pkgPath)
	}
	content, err = FormatAst(fset, f)
	if err != nil {
		return nil, err
	}
	return content, nil
}
