package tracking

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
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

type TrackCodeProvider interface {
	Position() CodeInsertPosition
	Stmts() []string
}

type TrackTemplateProvider interface {
	ImportSpec() (pkgPath, alias string)
	FrontTrackCodeProvider() TrackCodeProvider
	BackTrackCodeProvider() TrackCodeProvider
}

type InsertPosition struct {
	position CodeInsertPosition
	line     int
	column   int
}

type InsertPositions []InsertPosition

func (p *InsertPositions) Insert(position CodeInsertPosition, line int, column int) {
	*p = append(*p, InsertPosition{position: position, line: line, column: column})
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

// FunctionScopesOfAST returns the function scopes of the ast
func FunctionScopesOfAST(filename string, content []byte) (BlockScopes, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	nodes, err := functionNodesOfAST(astFile)
	if err != nil {
		return nil, err
	}

	blockScopes := BlockScopes{}
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.FuncDecl:
			blockScopes = append(blockScopes, BlockScope{
				StartLine: fset.Position(n.Body.Lbrace).Line,
				EndLine:   fset.Position(n.Body.Rbrace).Line,
			})
		case *ast.FuncLit:
			blockScopes = append(blockScopes, BlockScope{
				StartLine: fset.Position(n.Body.Lbrace).Line,
				EndLine:   fset.Position(n.Body.Rbrace).Line,
			})
		case *ast.File:
			blockScopes = append(blockScopes, BlockScope{
				StartLine: fset.Position(n.Pos()).Line,
				EndLine:   fset.Position(n.End()).Line,
			})
		}
	}
	blockScopes.Sort()
	return blockScopes, nil
}

// TrackScopes is a list of track scopes
type TrackScopes []TrackScope

// Sort sorts the track scopes by start line
func (t TrackScopes) Sort() {
	slices.SortFunc(t, func(a, b TrackScope) int {
		if a.StartLine == b.StartLine {
			return a.EndLine - b.EndLine
		}
		return a.StartLine - b.StartLine
	})
}

// Search returns the index of the track scope that contains the line
// Call after Sort, so the idx is the latest track scope that contains the line
func (t TrackScopes) Search(line int) int {
	idx := -1
	for i, trackScope := range t {
		if trackScope.Contains(line) {
			idx = i
		}
	}
	return idx
}

// Print prints the track scopes
func (t TrackScopes) Print() {
	fmt.Println("==============TrackScopes==============")
	for i, trackScope := range t {
		fmt.Println()
		fmt.Printf("==========track scope %d start==========\n", i)
		trackScope.Print()
		fmt.Printf("===========track scope %d end===========\n", i)
		fmt.Println()
	}
	fmt.Println("========================================")
}

// TrackScope is a scope of a function
// [StartLine, EndLine] is the range of the scope
type TrackScope struct {
	StartLine int
	EndLine   int
	node      *ast.BlockStmt
	Children  TrackScopes
}

// NewTrackScope creates a new track scope
func NewTrackScope(startLine int, endLine int, node *ast.BlockStmt) *TrackScope {
	return &TrackScope{
		StartLine: startLine,
		EndLine:   endLine,
		node:      node,
	}
}

func (f *TrackScope) AddChild(child TrackScope) {
	f.Children = append(f.Children, child)
}

// String returns the string representation of the track scope
func (f *TrackScope) String() string {
	return fmt.Sprintf("TrackScope{StartLine: %d, EndLine: %d}", f.StartLine, f.EndLine)
}

// IsEmpty checks if the track scope is empty
func (f *TrackScope) IsEmpty() bool {
	return f.StartLine == 0 && f.EndLine == 0
}

// IsValid checks if the track scope is valid
func (f *TrackScope) IsValid() bool {
	return f.StartLine < f.EndLine
}

// Contains checks if the track scope contains the line
func (f *TrackScope) Contains(line int) bool {
	return line > f.StartLine && line < f.EndLine
}

// ContainsRange checks if the track scope contains the range
func (f *TrackScope) ContainsRange(start, end int) bool {
	return start > f.StartLine && end < f.EndLine
}

// IsLeaf checks if the track scope is a leaf
func (f *TrackScope) IsLeaf() bool {
	return len(f.Children) == 0
}

// Search searches the track scope that contains the line
func (f *TrackScope) Search(line int) *TrackScope {
	for _, child := range f.Children {
		if child.Contains(line) {
			return child.Search(line)
		}
	}
	return f
}

// PrepareChildren prepares the children of the track scope
func (f *TrackScope) PrepareChildren(fset *token.FileSet) error {
	if f.node == nil {
		return nil
	}
	blockNodes, err := blockNodesOfFunction(f.node.List)
	if err != nil {
		return err
	}
	for _, blockNode := range blockNodes {
		if blockNode == nil {
			continue
		}
		switch block := blockNode.(type) {
		case *ast.IfStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}

		case *ast.ForStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}
		case *ast.RangeStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}
		case *ast.CaseClause:
			if len(block.Body) > 0 {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body[0].Pos()).Line - 1,
					EndLine:   fset.Position(block.Body[len(block.Body)-1].End()).Line + 1,
					node:      &ast.BlockStmt{List: block.Body},
				})
			}
		case *ast.CommClause:
			if len(block.Body) > 0 {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body[0].Pos()).Line - 1,
					EndLine:   fset.Position(block.Body[len(block.Body)-1].End()).Line + 1,
					node:      &ast.BlockStmt{List: block.Body},
				})
			}
		case *ast.SwitchStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}
		case *ast.TypeSwitchStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}
		case *ast.SelectStmt:
			if block.Body != nil {
				f.AddChild(TrackScope{
					StartLine: fset.Position(block.Body.Lbrace).Line,
					EndLine:   fset.Position(block.Body.Rbrace).Line,
					node:      block.Body,
				})
			}
		}
	}
	f.Children.Sort()

	for i := range f.Children {
		child := &f.Children[i]
		err := child.PrepareChildren(fset)
		if err != nil {
			return err
		}
	}
	children := []TrackScope{}
	if len(f.Children) > 0 {
		if f.StartLine < f.Children[0].StartLine {
			children = append(children, TrackScope{
				StartLine: f.StartLine,
				EndLine:   f.Children[0].StartLine,
				node:      nil,
			})
		}
		// fill the gap between children
		// for i < j, if children[i].EndLine < children[j].StartLine, then
		// fill a new child that child.startLine is children[i].EndLine
		// and child.EndLine is children[j].StartLine
		k := 0
		for i := range len(f.Children) {
			child := f.Children[i]
			if children[k].EndLine < child.StartLine {
				children = append(children, TrackScope{
					StartLine: children[k].EndLine,
					EndLine:   child.StartLine,
					node:      nil,
				})
				k++
			}
			children = append(children, child)
			k++
		}

		if f.EndLine > f.Children[len(f.Children)-1].EndLine {
			children = append(children, TrackScope{
				StartLine: f.Children[len(f.Children)-1].EndLine,
				EndLine:   f.EndLine,
				node:      nil,
			})
		}
	}
	f.Children = children
	return nil
}

func (f *TrackScope) Print() {
	fmt.Println(f.String())
	fmt.Printf("-----traverse children start: len(%d)----\n", len(f.Children))
	for _, child := range f.Children {
		child.Print()
	}
	fmt.Printf("-----traverse children end: len(%d)------\n", len(f.Children))
}

// TrackScopesOfAST returns the track scopes of the ast
func TrackScopesOfAST(filename string, content []byte) (TrackScopes, error) {

	// find all the function scopes
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	// find all function scopes
	trackScopes, err := functionTrackScopes(fset, astFile)
	if err != nil {
		return nil, err
	}
	// find all the block scopes
	for i := range trackScopes {
		trackScope := &trackScopes[i]
		err := trackScope.PrepareChildren(fset)
		if err != nil {
			return nil, err
		}
	}
	return trackScopes, nil
}

// functionTrackScopes returns the track scopes of the function
func functionTrackScopes(fset *token.FileSet, astFile *ast.File) (TrackScopes, error) {
	trackScopes := TrackScopes{}
	nodes, err := functionNodesOfAST(astFile)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.FuncDecl:
			trackScopes = append(trackScopes, TrackScope{
				StartLine: fset.Position(n.Body.Lbrace).Line,
				EndLine:   fset.Position(n.Body.Rbrace).Line,
				node:      n.Body,
			})
		case *ast.FuncLit:
			trackScopes = append(trackScopes, TrackScope{
				StartLine: fset.Position(n.Body.Lbrace).Line,
				EndLine:   fset.Position(n.Body.Rbrace).Line,
				node:      n.Body,
			})
		}
	}
	trackScopes.Sort()
	return trackScopes, nil
}

// functionNodesOfAST returns the nodes of the function
func functionNodesOfAST(astFile *ast.File) ([]ast.Node, error) {
	nodes := []ast.Node{}
	nodes = append(nodes, astFile)
	ast.Inspect(astFile, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		switch n := node.(type) {
		case *ast.FuncDecl:
			nodes = append(nodes, n)
		case *ast.FuncLit:
			nodes = append(nodes, n)
		}
		return true
	})
	return nodes, nil
}

// blockNodesOfFunction returns the nodes of the function
func blockNodesOfFunction(stmts []ast.Stmt) ([]ast.Stmt, error) {
	blockNodes := []ast.Stmt{}
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		switch stmt := stmt.(type) {
		case *ast.IfStmt:
			blockNodes = append(blockNodes, stmt)
			var elseStmt ast.Stmt
			elseStmt = stmt.Else
			for elseStmt != nil {
				switch s := elseStmt.(type) {
				case *ast.BlockStmt:
					blockNodes = append(blockNodes, &ast.IfStmt{
						Body: s,
					})
					elseStmt = nil
				case *ast.IfStmt:
					blockNodes = append(blockNodes, s)
					elseStmt = s.Else
				}
			}
		case *ast.ForStmt:
			blockNodes = append(blockNodes, stmt)
		case *ast.RangeStmt:
			blockNodes = append(blockNodes, stmt)
		case *ast.CaseClause:
			blockNodes = append(blockNodes, stmt)
		case *ast.CommClause:
			blockNodes = append(blockNodes, stmt)
		case *ast.LabeledStmt:
			blockNodes = append(blockNodes, stmt.Stmt)
		case *ast.SwitchStmt:
			blockNodes = append(blockNodes, stmt)
		case *ast.SelectStmt:
			blockNodes = append(blockNodes, stmt)
		case *ast.TypeSwitchStmt:
			blockNodes = append(blockNodes, stmt)

		}
	}
	return blockNodes, nil
}

type scopeKey struct {
	startLine, endLine int
}

type patchScope struct {
	scopeKey
	// indicesInfo[i]:
	// 0: the line is not new line
	// 1: the line is new line or comment but has not been marked inserted
	// 2: the line has been marked inserted
	marks []int // len(marks) == max(endLine - startLine - 1, 0)
}

func newTrackScopeIndex(startLine, endLine int) *patchScope {
	length := endLine - startLine - 1
	var marks []int
	if length > 0 {
		marks = make([]int, length)
	}
	return &patchScope{
		scopeKey: scopeKey{startLine: startLine, endLine: endLine},
		marks:    marks,
	}
}

// initMarks initializes the marks of the track scope index
// marks[i]: the i-th line has changed
func (t *patchScope) initMarks(marks []bool) {
	for i := range t.marks {
		if marks[t.startLine+i+1] {
			t.marks[i] = 1
		}
	}
}

func (t *patchScope) isInserted(line int) bool {
	return t.marks[line-t.startLine-1] == 2
}

func (t *patchScope) isNewLine(line int) bool {
	return t.marks[line-t.startLine-1] == 1
}

func (t *patchScope) canInsert(line int) bool {
	if t.isInserted(line) {
		return false
	}
	// Look forward. If a character that is not a new line is found, an insertion cannot be made.
	// It should immediately be determined whether this line has already been marked as inserted.
	// If so, return false; otherwise, return true.
	for j := line - 1; j > t.startLine; j-- {
		if t.isNewLine(j) {
			continue
		}
		// j first landed here is the first non-new line
		return !t.isInserted(j)
	}
	// The entire interval does not contain a non-new line, so it can be inserted.
	return true
}

// markInserted marks the line and all new lines before and after it as inserted
func (t *patchScope) markInserted(line int) {
	if t.marks[line-t.startLine-1] == 1 {
		t.marks[line-t.startLine-1] = 2
	}
	// Mark all new lines before and after this line as inserted
	for i := line - 1; i > t.startLine && t.isNewLine(i); i-- {
		t.marks[i-t.startLine-1] = 2
	}
	for i := line + 1; i < t.endLine && t.isNewLine(i); i++ {
		t.marks[i-t.startLine-1] = 2
	}
}
