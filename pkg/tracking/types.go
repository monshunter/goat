package tracking

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

// Tracker
type Tracker interface {
	Track() (int, error)
	Replace(target string, replace func(older string) (newer string)) (int, error)
	Bytes() []byte
	Count() int
	Save(path string) error
	TargetFile() string
}

type CodeInsertPosition int

const (
	CodeInsertPositionFront CodeInsertPosition = 1
	CodeInsertPositionBack  CodeInsertPosition = 2
)

const (
	CodeInsertPositionFrontStr = "front"
	CodeInsertPositionBackStr  = "back"
)

func (c CodeInsertPosition) String() string {
	return []string{CodeInsertPositionFrontStr, CodeInsertPositionBackStr}[c-1]
}

func (c CodeInsertPosition) Int() int {
	return int(c)
}

func (c CodeInsertPosition) IsFront() bool {
	return c == CodeInsertPositionFront
}

func (c CodeInsertPosition) IsBack() bool {
	return c == CodeInsertPositionBack
}

type CodeInsertType int

const (
	CodeInsertTypeComment CodeInsertType = 1
	CodeInsertTypeStmt    CodeInsertType = 2
)

const (
	CodeInsertTypeCommentStr = "comment"
	CodeInsertTypeStmtStr    = "stmt"
)

func (c CodeInsertType) String() string {
	return []string{CodeInsertTypeCommentStr, CodeInsertTypeStmtStr}[c-1]
}

func (c CodeInsertType) Int() int {
	return int(c)
}

func (c CodeInsertType) IsComment() bool {
	return c == CodeInsertTypeComment
}

func (c CodeInsertType) IsStmt() bool {
	return c == CodeInsertTypeStmt
}

type TrackCodeProvider interface {
	Position() CodeInsertPosition
	Comments() []string
	Stmts() []string
}

type TrackTemplateProvider interface {
	ImportSpec() (pkgPath, alias string)
	FrontTrackCodeProvider() TrackCodeProvider
	BackTrackCodeProvider() TrackCodeProvider
}

type InsertPosition struct {
	position CodeInsertPosition
	codeType CodeInsertType
	line     int
	column   int
}

type InsertPositions []InsertPosition

func (p *InsertPositions) Insert(position CodeInsertPosition, codeType CodeInsertType, line int, column int) {
	*p = append(*p, InsertPosition{position: position, codeType: codeType, line: line, column: column})
}

func (p *InsertPositions) Sort() {
	slices.SortFunc(*p, func(a, b InsertPosition) int {
		if a.line == b.line {
			return a.position.Int() - b.position.Int()
		}
		return a.line - b.line
	})
}

func (p *InsertPositions) Unique() {
	// return
	unique := make(map[InsertPosition]struct{})
	for _, position := range *p {
		unique[position] = struct{}{}
	}
	tmp := InsertPositions{}
	for position := range unique {
		tmp = append(tmp, position)
	}
	*p = tmp
}

func (p *InsertPositions) Reset() {
	*p = InsertPositions{}
}

type BlockScope struct {
	StartLine int
	EndLine   int
}

func (b *BlockScope) String() string {
	return fmt.Sprintf("BlockScope{StartLine: %d, EndLine: %d}", b.StartLine, b.EndLine)
}

func (b *BlockScope) IsEmpty() bool {
	return b.StartLine == 0 && b.EndLine == 0
}

func (b *BlockScope) IsValid() bool {
	return b.StartLine < b.EndLine
}

func (b *BlockScope) Contains(line int) bool {
	return line > b.StartLine && line < b.EndLine
}

func (b *BlockScope) ContainsRange(start, end int) bool {
	return start > b.StartLine && end < b.EndLine
}

type BlockScopes []BlockScope

func (b BlockScopes) Sort() {
	slices.SortFunc(b, func(a, b BlockScope) int {
		if a.StartLine == b.StartLine {
			return a.EndLine - b.EndLine
		}
		return a.StartLine - b.StartLine
	})
}

// SearchErr uses binary search which may miss nested scopes since it only returns
// the first matching scope it finds. The linear search in Search checks all scopes
// and finds the innermost (most nested) scope by continuing to search after finding
// a match. Binary search stops after finding any match.
//
// For example, with scopes:
// 1. {1, 100}
// 2. {10, 20}
// 3. {15, 18}
//
// Searching line 16:
// - SearchErr may return scope 1 or 2 depending on search order
// - Search will always return scope 3 (most nested)
//
// Binary search is faster but doesn't handle nested scopes well.
// Linear search is slower but more accurate for nested cases.
// Search call after Sort
func (b BlockScopes) SearchWrongImplement(line int) int {
	// find lastest scope of the line
	l, r := 0, len(b)-1
	idx := 0
	for l <= r {
		mid := l + (r-l)/2
		if b[mid].StartLine < line {
			if b[mid].EndLine > line {
				idx = mid
			}
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return idx
}

// Search call after Sort
func (b BlockScopes) Search(line int) int {
	// find lastest scope of the line
	idx := 0
	for i, scope := range b {
		if scope.StartLine < line && scope.EndLine > line {
			idx = i
		}
	}
	return idx
}

// BlockScopesOfGoAST returns the block scopes of the go ast
func BlockScopesOfGoAST(filename string, content []byte) (BlockScopes, error) {

	fset := token.NewFileSet()
	blockScopes := BlockScopes{}
	astFile, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	blockScopes = append(blockScopes, BlockScope{
		StartLine: 1,
		EndLine:   len(lines),
	})
	for _, decl := range astFile.Decls {
		if declFunc, ok := decl.(*ast.FuncDecl); ok {
			if declFunc.Body == nil {
				continue
			}
			blockScopes = append(blockScopes, BlockScope{
				StartLine: fset.Position(declFunc.Body.Lbrace).Line,
				EndLine:   fset.Position(declFunc.Body.Rbrace).Line,
			})

			// Traverse the statements in the function body
			for _, stmt := range declFunc.Body.List {
				ast.Inspect(stmt, func(node ast.Node) bool {
					if node == nil {
						return false
					}
					switch stmt := node.(type) {
					case *ast.IfStmt:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body.Lbrace).Line,
								EndLine:   fset.Position(stmt.Body.Rbrace).Line,
							})
							if stmt.Else != nil {
								switch stmt.Else.(type) {
								case *ast.BlockStmt:
									block := stmt.Else.(*ast.BlockStmt)
									blockScopes = append(blockScopes, BlockScope{
										StartLine: fset.Position(block.Lbrace).Line,
										EndLine:   fset.Position(block.Rbrace).Line,
									})
								}
							}
						}

					case *ast.ForStmt:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body.Lbrace).Line,
								EndLine:   fset.Position(stmt.Body.Rbrace).Line,
							})
						}
					case *ast.RangeStmt:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body.Lbrace).Line,
								EndLine:   fset.Position(stmt.Body.Rbrace).Line,
							})
						}

					case *ast.CaseClause:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body[0].Pos()).Line - 1,
								EndLine:   fset.Position(stmt.Body[len(stmt.Body)-1].End()).Line + 1,
							})
						}

					case *ast.CommClause:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body[0].Pos()).Line - 1,
								EndLine:   fset.Position(stmt.Body[len(stmt.Body)-1].End()).Line + 1,
							})
						}
					case *ast.FuncLit:
						if stmt.Body != nil {
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(stmt.Body.Lbrace).Line,
								EndLine:   fset.Position(stmt.Body.Rbrace).Line,
							})
						}
					}
					return true
				})
			}
		}
	}
	blockScopes.Sort()
	return blockScopes, nil
}

// FunctionScopesOfGoAST returns the function scopes of the go ast
func FunctionScopesOfGoAST(filename string, content []byte) (BlockScopes, error) {

	fset := token.NewFileSet()
	blockScopes := BlockScopes{}
	astFile, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	// Add this scope to make the whole file as a function scope
	// This can make the search function works
	// But index 0 is not a function scope
	blockScopes = append(blockScopes, BlockScope{
		StartLine: 1,
		EndLine:   len(lines),
	})

	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Body == nil {
				continue
			}

			blockScopes = append(blockScopes, BlockScope{
				StartLine: fset.Position(decl.Body.Lbrace).Line,
				EndLine:   fset.Position(decl.Body.Rbrace).Line,
			})
			for _, stmt := range decl.Body.List {
				ast.Inspect(stmt, func(node ast.Node) bool {
					if node == nil {
						return false
					}
					switch n := node.(type) {
					case *ast.FuncLit:
						if n.Body == nil {
							return false
						}
						blockScopes = append(blockScopes, BlockScope{
							StartLine: fset.Position(n.Body.Lbrace).Line,
							EndLine:   fset.Position(n.Body.Rbrace).Line,
						})
					}
					return true
				})
			}
		case *ast.GenDecl:
			if len(decl.Specs) == 0 {
				continue
			}
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ValueSpec:
					ast.Inspect(spec, func(node ast.Node) bool {
						if node == nil {
							return false
						}
						switch n := node.(type) {
						case *ast.FuncLit:
							if n.Body != nil {
								return false
							}
							blockScopes = append(blockScopes, BlockScope{
								StartLine: fset.Position(n.Body.Lbrace).Line,
								EndLine:   fset.Position(n.Body.Rbrace).Line,
							})
						}
						return true
					})
				}
			}
		}
	}
	blockScopes.Sort()
	return blockScopes, nil
}
