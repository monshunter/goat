package tracking

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/monshunter/goat/pkg/log"

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	increament "github.com/monshunter/goat/pkg/tracking/increment"
	"github.com/monshunter/goat/pkg/utils"
)

type IncrementalTrack struct {
	basePath                    string
	fileChange                  *diff.FileChange
	count                       int
	content                     []byte
	fileName                    string
	insertedPositions           InsertPositions
	singleLineInsertedPositions InsertPositions
	visitedInsertedPositions    map[InsertPosition]struct{}
	granularity                 config.Granularity
	importPathPlaceHolder       string
	trackStmtPlaceHolders       []string
	source                      []string
	sourceLength                int
	functionScopes              BlockScopes
	printerConfig               *printer.Config
	trackScopes                 TrackScopes
	visitedTrackScopes          map[scopeKey]struct{}
	patchScopes                 map[scopeKey]*patchScope
	lineChanges                 []bool
	comments                    []bool
}

func NewIncrementalTrack(basePath string, fileChange *diff.FileChange,
	importPathPlaceHolder string, trackStmtPlaceHolders []string,
	granularity config.Granularity, printerConfig *printer.Config) (*IncrementalTrack, error) {
	fileName := fileChange.Path
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(basePath, fileName)
	}
	fileName = filepath.Clean(fileName)

	if importPathPlaceHolder == "" {
		importPathPlaceHolder = increament.TrackImportPathPlaceHolder
	}
	if len(trackStmtPlaceHolders) == 0 {
		trackStmtPlaceHolders = increament.GetPackageInsertStmts()
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	// analyze function scopes
	functionScopes, err := FunctionScopesOfAST(fileName, content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze function scopes in %s: %w", fileName, err)
	}

	// analyze tracking scopes
	var trackScopes TrackScopes
	if granularity.IsPatch() || granularity.IsScope() {
		trackScopes, err = TrackScopesOfAST(fileName, content)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze tracking scopes in %s: %w", fileName, err)
		}
	}
	source := strings.Split(string(content), "\n")
	comments := initComments(source)
	lineChanges := initLineChanges(len(source), fileChange.LineChanges)

	log.Debugf("Tracking file: %s (funcs=%d, trackScopes=%d)",
		fileName, len(functionScopes), len(trackScopes))

	return &IncrementalTrack{
		basePath:                    basePath,
		fileChange:                  fileChange,
		fileName:                    fileName,
		content:                     content,
		insertedPositions:           InsertPositions{},
		singleLineInsertedPositions: InsertPositions{},
		granularity:                 granularity,
		importPathPlaceHolder:       importPathPlaceHolder,
		trackStmtPlaceHolders:       trackStmtPlaceHolders,
		source:                      source,
		sourceLength:                len(source),
		functionScopes:              functionScopes,
		visitedInsertedPositions:    make(map[InsertPosition]struct{}),
		printerConfig:               printerConfig,
		trackScopes:                 trackScopes,
		visitedTrackScopes:          make(map[scopeKey]struct{}),
		patchScopes:                 make(map[scopeKey]*patchScope),
		lineChanges:                 lineChanges,
		comments:                    comments,
	}, nil
}

func (t *IncrementalTrack) doInsert() ([]byte, error) {

	if t.count == 0 {
		return t.content, nil
	}

	adjustLength := calculateInsertLength(t.insertedPositions, t.trackStmtPlaceHolders)
	t.insertedPositions.Unique()
	t.insertedPositions.Sort()
	t.singleLineInsertedPositions.Unique()
	t.singleLineInsertedPositions.Sort()

	posIdx := 0
	i := 0
	sources := t.source
	var buf bytes.Buffer
	deltaArray := make([]int, len(t.singleLineInsertedPositions))
	for i := range deltaArray {
		deltaArray[i] = -1
	}
	deltaIdx := -1
	deltaLine := -1
	if len(t.singleLineInsertedPositions) > 0 {
		deltaIdx = 0
		deltaLine = t.singleLineInsertedPositions[0].line - 1
	}
	delta := 0
	buf.Grow(t.sourceLength + adjustLength)
	for ; i < len(sources) && posIdx < len(t.insertedPositions); i++ {
		if i == t.insertedPositions[posIdx].line-1 {
			lines := doInsert(&buf, t.trackStmtPlaceHolders)
			delta += lines
			posIdx++
		}
		if i == deltaLine {
			deltaArray[deltaIdx] = delta
			deltaIdx++
			if deltaIdx < len(deltaArray) {
				deltaLine = t.singleLineInsertedPositions[deltaIdx].line - 1
			} else {
				deltaLine = -1
			}
			delta = 0
		}
		buf.WriteString(sources[i])
		buf.WriteByte('\n')
	}
	buf.WriteString(strings.Join(sources[i:], "\n"))
	var content []byte
	content = buf.Bytes()
	// handle inserted statements for single line function
	if len(deltaArray) > 0 {
		// adjust the line number of the single line inserted statements
		if deltaIdx < len(deltaArray) && deltaArray[deltaIdx] == -1 {
			deltaArray[deltaIdx] = delta
		}
		for k := 1; k < len(deltaArray); k++ {
			if deltaArray[k] == -1 {
				deltaArray[k] = deltaArray[k-1]
			} else {
				deltaArray[k] += deltaArray[k-1]
			}
		}

		for k := range deltaArray {
			t.singleLineInsertedPositions[k].line += deltaArray[k]
		}

		newSources := strings.Split(buf.String(), "\n")
		adjustLength = calculateInsertLength(t.singleLineInsertedPositions, t.trackStmtPlaceHolders)
		var newBuf bytes.Buffer
		newBuf.Grow(buf.Len() + adjustLength)
		k := 0
		i := 0
		for ; i < len(newSources) && k < len(t.singleLineInsertedPositions); i++ {
			src := newSources[i]
			if i == t.singleLineInsertedPositions[k].line-1 {
				// write column before
				column := t.singleLineInsertedPositions[k].column
				column -= 1
				newBuf.WriteString(src[:column])
				newBuf.WriteByte('\n')
				// write track stmt place holders
				_ = doInsert(&newBuf, t.trackStmtPlaceHolders)
				// write column after
				newBuf.WriteString(src[column:])
				newBuf.WriteByte('\n')
				k++
			} else {
				newBuf.WriteString(src)
				newBuf.WriteByte('\n')
			}
		}
		newBuf.WriteString(strings.Join(newSources[i:], "\n"))
		content = newBuf.Bytes()
	}

	newFset, newF, err := utils.GetAstTree("", content)
	if err != nil {
		log.Errorf("Failed to get ast tree, file: %s, error: %v", t.fileName, err)
		return nil, err
	}
	content, err = utils.FormatAst(t.printerConfig, newFset, newF)
	if err != nil {
		log.Errorf("Failed to format ast, file: %s, error: %v", t.fileName, err)
		return nil, err
	}
	return content, nil
}

func (t *IncrementalTrack) checkAndMarkInsert(line int) {
	if !t.isLineChanged(line) {
		return
	}
	t.forceMarkInsert(line)
}

func (t *IncrementalTrack) forceMarkInsert(line int) {
	if t.granularity.IsFunc() {
		t.markInsertByFunc(line)
	} else if t.granularity.IsScope() {
		t.markInsertByScope(line)
	} else if t.granularity.IsPatch() {
		t.markInsertByPatch(line)
	} else if t.granularity.IsLine() {
		t.markInsertByLine(line)
	}
}

func (t *IncrementalTrack) markInsertByScope(line int) {
	id := t.trackScopes.Search(line)
	if id == -1 {
		return
	}
	trackScope := t.trackScopes[id].Search(line)
	key := scopeKey{startLine: trackScope.StartLine, endLine: trackScope.EndLine}
	if _, ok := t.visitedTrackScopes[key]; ok {
		return
	}
	t.visitedTrackScopes[key] = struct{}{}
	t.markInsert(line)
}

func (t *IncrementalTrack) markInsertByFunc(line int) {
	var valid bool
	line, valid = t.getInsertPositionInFunctionBody(line)
	if !valid {
		return
	}
	t.markInsert(line)
}

func (t *IncrementalTrack) markInsertByLine(line int) {
	t.markInsert(line)
}

func (t *IncrementalTrack) markInsertByPatch(line int) {
	idx := t.trackScopes.Search(line)
	if idx == -1 {
		return
	}

	trackScope := t.trackScopes[idx].Search(line)
	key := scopeKey{startLine: trackScope.StartLine, endLine: trackScope.EndLine}
	patchScope := t.patchScopes[key]
	if patchScope == nil {
		patchScope = newTrackScopeIndex(t.trackScopes[idx].StartLine, t.trackScopes[idx].EndLine)
		patchScope.initMarks(t.lineChanges)
		patchScope.initMarks(t.comments)
		t.patchScopes[key] = patchScope
	}
	if patchScope.canInsert(line) {
		t.markInsert(line)
		patchScope.markInserted(line)
		t.patchScopes[key] = patchScope
	}
}

func (t *IncrementalTrack) isLineChanged(line int) bool {
	return t.lineChanges[line]
}

func (t *IncrementalTrack) isLineChangedRange(start, end int) bool {
	for i := start; i <= end; i++ {
		if t.isLineChanged(i) {
			return true
		}
	}
	return false
}

func (t *IncrementalTrack) markInsert(line int) {
	for t.comments[line] {
		line++
	}

	// Add a check to avoid inserts outside of function scopes which will cause error
	if !t.isInFunctionScopes(line) {
		return
	}

	// Add a check to avoid duplicate inserts
	// Use visitedPositionInserts map to record the inserted positions,
	// This can prevent duplicate inserts in multiple AST scans.
	// It is important for ensuring the correctness and avoiding unnecessary duplicate tracking.
	key := InsertPosition{line: line, column: 0}
	if _, ok := t.visitedInsertedPositions[key]; ok {
		return
	}
	t.insertedPositions.Insert(line, 0)
	t.visitedInsertedPositions[key] = struct{}{}
	t.count++
}

func (t *IncrementalTrack) isInFunctionScopes(line int) bool {
	return t.functionScopes.Search(line) > 0
}

func (t *IncrementalTrack) getInsertPositionInFunctionBody(line int) (insertLine int, valid bool) {
	idx := t.functionScopes.Search(line)
	if idx == 0 {
		return -1, false
	}
	return t.functionScopes[idx].StartLine + 1, true
}

// Target returns the target file name
func (t *IncrementalTrack) Target() string {
	return t.fileName
}

// Track adds tracking statements to the target file
func (t *IncrementalTrack) Track() (int, error) {
	var err error
	t.insertedPositions.Reset()
	t.singleLineInsertedPositions.Reset()
	clear(t.visitedInsertedPositions)
	log.Debugf("Adding tracking to file: %s", t.fileName)
	t.content, err = t.addStmts()
	if err != nil {
		return 0, fmt.Errorf("failed to add tracking statements to %s: %w", t.fileName, err)
	}
	if t.count > 0 {
		// do default insert
		log.Debugf("Adding default import to file: %s", t.fileName)
		t.insertedPositions.Reset()
		t.singleLineInsertedPositions.Reset()
		clear(t.visitedInsertedPositions)
		content, err := utils.AddImport(t.printerConfig, t.importPathPlaceHolder, "", t.fileName, t.content)
		if err != nil {
			log.Errorf("Failed to add import: %v", err)
			return 0, err
		}
		t.content = content
	}
	log.Debugf("Added %d tracking points to %s", t.count, t.fileName)
	return t.count, nil
}

// Count returns the number of tracking points
func (t *IncrementalTrack) Count() int {
	return t.count
}

// Content returns the content of the target file
func (t *IncrementalTrack) Content() []byte {
	return t.content
}

// SetContent sets the content of the target file
func (t *IncrementalTrack) SetContent(content []byte) {
	t.content = content
}

// addStmts adds tracking statements to the target file
func (t *IncrementalTrack) addStmts() ([]byte, error) {
	fset, f, err := utils.GetAstTree("", t.content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", t.fileName, err)
	}

	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Body == nil || len(decl.Body.List) == 0 {
				continue
			}
			if fset.Position(decl.Body.Lbrace).Line == fset.Position(decl.Body.Rbrace).Line {
				// skip single line function body
				if len(decl.Body.List) == 0 {
					continue
				}
				pos := fset.Position(decl.Body.List[0].Pos())
				t.insertSingleLineStmt(pos)
			}
			t.processStatements(decl.Body.List, fset)
			t.processControlStatements(decl.Body, fset)
		case *ast.GenDecl:
			t.processGlobalValueSpecs(decl.Specs, fset)
			t.processGlobalFunctionLit(decl.Specs, fset)
		}
	}
	return t.doInsert()
}

func (t *IncrementalTrack) processGlobalValueSpecs(specs []ast.Spec, fset *token.FileSet) {
	for _, spec := range specs {
		switch spec := spec.(type) {
		case *ast.ValueSpec:
			for _, value := range spec.Values {
				ast.Inspect(value, func(n ast.Node) bool {
					if n == nil {
						return false
					}
					switch n := n.(type) {
					case *ast.FuncLit:
						if n.Body == nil || len(n.Body.List) == 0 {
							return false
						}
						if fset.Position(n.Body.Lbrace).Line == fset.Position(n.Body.Rbrace).Line {
							pos := fset.Position(n.Body.List[0].Pos())
							t.insertSingleLineStmt(pos)
							return false
						}
						t.processStatements(n.Body.List, fset)
						return false
					}
					return true
				})
			}
		}
	}
}

func (t *IncrementalTrack) processGlobalFunctionLit(specs []ast.Spec, fset *token.FileSet) {
	for _, spec := range specs {
		switch spec := spec.(type) {
		case *ast.ValueSpec:
			for _, value := range spec.Values {
				ast.Inspect(value, func(n ast.Node) bool {
					if n == nil {
						return false
					}
					switch n := n.(type) {
					case *ast.FuncLit:
						if n.Body != nil {
							t.processControlStatements(n.Body, fset)
						}
						return false
					}
					return true
				})
			}
		}
	}
}

func (t *IncrementalTrack) processControlStatements(node ast.Node, fset *token.FileSet) {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		var changed bool
		switch n := n.(type) {
		case *ast.IfStmt:
			if n.Body == nil {
				break
			}
			// return true
			// Judge if the if statement is changed:
			// 1. If the if statement is changed, it means the whole if statement may be modified
			// 2. Even if the Init part is not changed, the change of the if statement may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to the if statement
			changed = t.isLineChanged(fset.Position(n.If).Line)
			if !changed && n.Init != nil {
				changed = t.isLineChangedRange(fset.Position(n.Init.Pos()).Line,
					fset.Position(n.Init.End()).Line)
			}
			if !changed && n.Cond != nil {
				changed = t.isLineChangedRange(fset.Position(n.Cond.Pos()).Line,
					fset.Position(n.Cond.End()).Line)
			}

			if changed {
				t.forceMarkInsert(fset.Position(n.Body.Lbrace).Line + 1)
				if n.Else != nil {
					elseStmt, ok := n.Else.(*ast.BlockStmt)
					if ok && len(elseStmt.List) > 0 {
						t.forceMarkInsert(fset.Position(elseStmt.Lbrace).Line + 1)
					}
				}
			}

		case *ast.SwitchStmt:
			// return true
			// Judge if the switch keyword line is changed:
			// 1. If the switch keyword line is changed, it means the whole switch statement may be modified
			// 2. Even if the Init and Tag parts are not changed, the change of the switch keyword line may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to each case clause
			changed = t.isLineChanged(fset.Position(n.Switch).Line)
			if !changed && n.Init != nil {
				changed = t.isLineChangedRange(fset.Position(n.Init.Pos()).Line,
					fset.Position(n.Init.End()).Line)
			}
			if !changed && n.Tag != nil {
				changed = t.isLineChangedRange(fset.Position(n.Tag.Pos()).Line,
					fset.Position(n.Tag.End()).Line)
			}
			if changed && n.Body != nil {
				for _, subStmt := range n.Body.List {
					if caseClause, ok := subStmt.(*ast.CaseClause); ok && len(caseClause.Body) > 0 {
						t.forceMarkInsert(fset.Position(caseClause.Colon).Line + 1)
					}
				}
			}
		case *ast.TypeSwitchStmt:
			// return true
			changed = t.isLineChanged(fset.Position(n.Switch).Line)
			if !changed && n.Init != nil {
				changed = t.isLineChangedRange(fset.Position(n.Init.Pos()).Line,
					fset.Position(n.Init.End()).Line)
			}
			if !changed && n.Assign != nil {
				changed = t.isLineChangedRange(fset.Position(n.Assign.Pos()).Line, fset.Position(n.Assign.End()).Line)
			}

			// Judge if the type switch statement is changed:
			// 1. If the type switch statement is changed, it means the whole type switch statement may be modified
			// 2. Even if the Init and Assign parts are not changed, the change of the type switch statement may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to each case clause
			if changed && n.Body != nil {
				for _, subStmt := range n.Body.List {
					if caseClause, ok := subStmt.(*ast.CaseClause); ok && len(caseClause.Body) > 0 {
						t.forceMarkInsert(fset.Position(caseClause.Colon).Line + 1)
					}
				}
			}
		case *ast.CaseClause:
			// return true
			changed = t.isLineChanged(fset.Position(n.Case).Line)
			if !changed && len(n.List) > 0 {
				for _, expr := range n.List {
					changed = t.isLineChangedRange(fset.Position(expr.Pos()).Line, fset.Position(expr.End()).Line)
					if changed {
						break
					}
				}
			}
			if changed {
				t.forceMarkInsert(fset.Position(n.Colon).Line + 1)
			}

		case *ast.CommClause:
			// return true
			changed = t.isLineChanged(fset.Position(n.Case).Line)
			if !changed && n.Comm != nil {
				changed = t.isLineChangedRange(fset.Position(n.Comm.Pos()).Line, fset.Position(n.Comm.End()).Line)
			}
			if changed {
				t.forceMarkInsert(fset.Position(n.Colon).Line + 1)
			}
		case *ast.RangeStmt:
			if n.Body == nil {
				break
			}
			// return true
			// Judge if the range statement is changed:
			// 1. If the range statement is changed, it means the whole range statement may be modified
			// 2. Even if the Key and Value parts are not changed, the change of the range statement may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to the range statement
			changed = t.isLineChanged(fset.Position(n.For).Line)
			if !changed && n.Key != nil {
				changed = t.isLineChangedRange(fset.Position(n.Key.Pos()).Line,
					fset.Position(n.Key.End()).Line)
			}
			if !changed && n.Value != nil {
				changed = t.isLineChangedRange(fset.Position(n.Value.Pos()).Line,
					fset.Position(n.Value.End()).Line)
			}
			if !changed && n.X != nil {
				changed = t.isLineChangedRange(fset.Position(n.X.Pos()).Line, fset.Position(n.X.End()).Line)
			}

			if changed {
				t.forceMarkInsert(fset.Position(n.Body.Lbrace).Line + 1)
			}
		case *ast.ForStmt:
			if n.Body == nil {
				break
			}
			// Judge if the for statement is changed:
			// 1. If the for statement is changed, it means the whole for statement may be modified
			// 2. Even if the Init and Assign parts are not changed, the change of the for statement may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to each case clause
			changed = t.isLineChanged(fset.Position(n.For).Line)
			if !changed && n.Init != nil {
				changed = t.isLineChangedRange(fset.Position(n.Init.Pos()).Line,
					fset.Position(n.Init.End()).Line)
			}
			if !changed && n.Cond != nil {
				changed = t.isLineChangedRange(fset.Position(n.Cond.Pos()).Line,
					fset.Position(n.Cond.End()).Line)
			}
			if !changed && n.Post != nil {
				changed = t.isLineChangedRange(fset.Position(n.Post.Pos()).Line,
					fset.Position(n.Post.End()).Line)
			}

			if changed {
				t.forceMarkInsert(fset.Position(n.Body.Lbrace).Line + 1)
			}
		}
		return true
	})
}

// processStatements analyzes and modifies statements by inserting additional code
// nodes before each statement in the function body.
func (t *IncrementalTrack) processStatements(statList []ast.Stmt, fset *token.FileSet) {
	for _, stmt := range statList {
		if stmt == nil {
			continue
		}
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
			t.analyzeAndModifyExpr(s.Rhs, fset)
		case *ast.IfStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
			if s.Else != nil {
				switch s.Else.(type) {
				case *ast.IfStmt:
					t.processStatements([]ast.Stmt{s.Else.(*ast.IfStmt)}, fset)
				case *ast.BlockStmt:
					t.processStatements(s.Else.(*ast.BlockStmt).List, fset)
				}
			}
		case *ast.ForStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.RangeStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.SwitchStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.CommClause:
			t.processStatements(s.Body, fset)
		case *ast.CaseClause:
			t.processStatements(s.Body, fset)
		case *ast.BlockStmt:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
			t.processStatements(s.List, fset)
		case *ast.ReturnStmt:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
			t.analyzeAndModifyExpr(s.Results, fset)
		case *ast.DeferStmt:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)
			}
		case *ast.SelectStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.GoStmt:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)
			}
		case *ast.TypeSwitchStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.ExprStmt:
			switch s.X.(type) {
			case *ast.CallExpr:
				t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
				expr := s.X.(*ast.CallExpr)
				if expr.Fun != nil {
					t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)
				}
				if expr.Args != nil {
					t.analyzeAndModifyExpr(expr.Args, fset)
				}
			default:
			}
		case *ast.LabeledStmt:
			t.processStatements([]ast.Stmt{s.Stmt}, fset)
		default:
			t.checkAndMarkInsert(fset.Position(s.Pos()).Line)
		}
	}
}

// analyzeAndModifyExpr analyzes and modifies expressions by processing any function literals found.
// It works in conjunction with processStatements to recursively handle nested expressions.
func (t *IncrementalTrack) analyzeAndModifyExpr(exprList []ast.Expr, fset *token.FileSet) {
	for _, expr := range exprList {
		if expr == nil {
			continue
		}
		switch expr := expr.(type) {
		case *ast.FuncLit:

			if expr.Body == nil || len(expr.Body.List) == 0 {
				continue
			}
			// Handle single line function
			if fset.Position(expr.Pos()).Line == fset.Position(expr.End()).Line {
				pos := fset.Position(expr.Body.List[0].Pos())
				t.insertSingleLineStmt(pos)
				continue
			}
			t.processStatements(expr.Body.List, fset)
		case *ast.CallExpr:
			if expr.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)
			}
			if expr.Args != nil {
				t.analyzeAndModifyExpr(expr.Args, fset)
			}
		case *ast.StructType:
			if expr.Fields != nil {
				for _, field := range expr.Fields.List {
					if field.Type != nil {
						t.analyzeAndModifyExpr([]ast.Expr{field.Type}, fset)
					}
				}
			}
		case *ast.CompositeLit:
			if expr.Elts != nil {
				for _, elt := range expr.Elts {
					t.analyzeAndModifyExpr([]ast.Expr{elt}, fset)
				}
			}
		case *ast.KeyValueExpr:
			if expr.Value != nil {
				t.analyzeAndModifyExpr([]ast.Expr{expr.Value}, fset)
			}
		case *ast.UnaryExpr:
			if expr.X != nil {
				t.analyzeAndModifyExpr([]ast.Expr{expr.X}, fset)
			}
		default:
		}
	}
}

func (t *IncrementalTrack) insertSingleLineStmt(pos token.Position) {
	if t.isLineChanged(pos.Line) {
		t.count++
		t.singleLineInsertedPositions.Insert(pos.Line, pos.Column)
	}
}

func calculateInsertLength(insertedPositions InsertPositions, defaultInserts []string) int {

	count := 0
	for _, stmt := range defaultInserts {
		count += len(stmt) + 1
	}
	return count * len(insertedPositions)
}

func doInsert(buf *bytes.Buffer, trackStmtPlaceHolders []string) int {
	// write default track stmt place holders
	for _, trackStmtPlaceHolder := range trackStmtPlaceHolders {
		buf.WriteString(trackStmtPlaceHolder)
		buf.WriteByte('\n')
	}
	return len(trackStmtPlaceHolders)
}

// initLineChanges initializes the line changes array
func initLineChanges(length int, lineChanges diff.LineChanges) []bool {
	// +1 for the line number, because the line number is 1-based
	res := make([]bool, length+1)
	for _, lineChange := range lineChanges {
		for i := lineChange.Start; i < lineChange.Start+lineChange.Lines; i++ {
			res[i] = true
		}
	}
	return res
}

// initComments initializes the comments array
func initComments(sources []string) []bool {
	// +1 for the line number, because the line number is 1-based
	comments := make([]bool, len(sources)+1)
	for i, source := range sources {
		comments[i+1] = utils.IsGoComment(source)
	}
	return comments
}
