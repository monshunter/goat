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
	"github.com/monshunter/goat/pkg/tracking/increament"
	"github.com/monshunter/goat/pkg/utils"
)

type IncreamentTrack struct {
	basePath                    string
	fileChange                  *diff.FileChange
	provider                    TrackTemplateProvider
	count                       int
	content                     []byte
	fileName                    string
	insertedPositions           InsertPositions
	singleLineInsertedPositions InsertPositions
	visitedInsertedPositions    map[InsertPosition]struct{}
	lastBlockInsertLine         int
	granularity                 config.Granularity
	importPathPlaceHolder       string
	trackStmtPlaceHolders       []string
	source                      []string
	sourceLength                int
	blockScopes                 BlockScopes
	functionScopes              BlockScopes
	printerConfig               *printer.Config
	trackScopes                 TrackScopes
	visitedTrackScopes          map[TrackScopeKey]struct{}
}

type TrackScopeKey struct {
	StartLine, EndLine int
}

func NewIncreamentTrack(basePath string, fileChange *diff.FileChange,
	importPathPlaceHolder string, trackStmtPlaceHolders []string,
	provider TrackTemplateProvider, granularity config.Granularity, printerConfig *printer.Config) (*IncreamentTrack, error) {
	fileName := fileChange.Path
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(basePath, fileName)
	}
	fileName = filepath.Clean(fileName)
	if provider == nil {
		provider = defaultIncrementTemplateProvider()
	}

	if importPathPlaceHolder == "" {
		importPathPlaceHolder = increament.TrackImportPathPlaceHolder
	}
	if len(trackStmtPlaceHolders) == 0 {
		trackStmtPlaceHolders = increament.GetPackageInsertData()
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	blockScopes, err := BlockScopesOfAST(fileName, content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze block scopes in %s: %w", fileName, err)
	}
	functionScopes, err := FunctionScopesOfAST(fileName, content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze function scopes in %s: %w", fileName, err)
	}

	trackScopes, err := TrackScopesOfAST(fileName, content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze tracking scopes in %s: %w", fileName, err)
	}

	log.Debugf("Tracking file: %s (blocks=%d, funcs=%d, trackScopes=%d)",
		fileName, len(blockScopes), len(functionScopes), len(trackScopes))

	return &IncreamentTrack{
		basePath:                    basePath,
		fileChange:                  fileChange,
		provider:                    provider,
		fileName:                    fileName,
		content:                     content,
		insertedPositions:           InsertPositions{},
		singleLineInsertedPositions: InsertPositions{},
		granularity:                 granularity,
		importPathPlaceHolder:       importPathPlaceHolder,
		trackStmtPlaceHolders:       trackStmtPlaceHolders,
		source:                      strings.Split(string(content), "\n"),
		sourceLength:                len(content),
		blockScopes:                 blockScopes,
		functionScopes:              functionScopes,
		visitedInsertedPositions:    make(map[InsertPosition]struct{}),
		printerConfig:               printerConfig,
		trackScopes:                 trackScopes,
		visitedTrackScopes:          make(map[TrackScopeKey]struct{}),
	}, nil
}

func (t *IncreamentTrack) doInsert() ([]byte, error) {

	if t.count == 0 {
		return t.content, nil
	}

	frontStmts, backStmts, frontComments, backComments := t.getContentsToInsert()
	adjustLength := calculateInsertLength(t.insertedPositions, t.trackStmtPlaceHolders, frontStmts, backStmts, frontComments, backComments)
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
	deltaArrayLength := len(deltaArray)
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
			pos := t.insertedPositions[posIdx]
			lines := doInsert(&buf, pos, t.trackStmtPlaceHolders, frontStmts, backStmts, frontComments, backComments)
			delta += lines
			posIdx++
		}
		if i == deltaLine {
			deltaArray[deltaIdx] = delta
			deltaIdx++
			if deltaIdx < deltaArrayLength {
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

	// handle inserted statements for single line function
	if deltaArrayLength > 0 {
		// adjust the line number of the single line inserted statements
		if deltaIdx < deltaArrayLength && deltaArray[deltaIdx] == -1 {
			deltaArray[deltaIdx] = delta
		}
		for k := 1; k < len(deltaArray); k++ {
			if deltaArray[k] == -1 {
				deltaArray[k] = deltaArray[k-1]
			} else {
				deltaArray[k] += deltaArray[k-1]
			}
		}

		for k := 0; k < len(deltaArray); k++ {
			t.singleLineInsertedPositions[k].line += deltaArray[k]
		}

		newSources := strings.Split(buf.String(), "\n")
		adjustLength = calculateInsertLength(t.singleLineInsertedPositions, t.trackStmtPlaceHolders, frontStmts, backStmts, frontComments, backComments)
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
				_ = doInsert(&newBuf, t.singleLineInsertedPositions[k], t.trackStmtPlaceHolders, frontStmts, backStmts, frontComments, backComments)
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
		t.content = newBuf.Bytes()
	} else {
		t.content = buf.Bytes()
	}
	newFset, newF, err := utils.GetAstTree("", t.content)
	if err != nil {
		log.Errorf("Failed to get ast tree, file: %s, error: %v", t.fileName, err)
		return nil, err
	}
	content, err := utils.FormatAst(t.printerConfig, newFset, newF)
	if err != nil {
		log.Errorf("Failed to format ast, file: %s, error: %v", t.fileName, err)
		return nil, err
	}
	t.content = content
	return t.content, nil
}

func (t *IncreamentTrack) getContentsToInsert() (
	frontStmts []string, backStmts []string, frontComments []string, backComments []string) {
	if t.provider != nil {
		if t.provider.FrontTrackCodeProvider() != nil {
			frontComments = t.provider.FrontTrackCodeProvider().Comments()
			frontStmts = t.provider.FrontTrackCodeProvider().Stmts()
		}
		if t.provider.BackTrackCodeProvider() != nil {
			backComments = t.provider.BackTrackCodeProvider().Comments()
			backStmts = t.provider.BackTrackCodeProvider().Stmts()
		}
	}
	return
}

func (t *IncreamentTrack) checkAndMarkInsert(position CodeInsertPosition, line int) {
	if t.granularity.IsLine() {
		t.checkAndMarkInsertByLine(position, line)
	} else if t.granularity.IsBlock() {
		t.checkAndMarkInsertStmtByBlock(position, line)
	} else if t.granularity.IsFunc() {
		t.checkAndMarkInsertStmtByFunc(position, line)
	} else if t.granularity.IsScope() {
		t.checkAndMarkInsertStmtByScope(position, line)
	}
}

func (t *IncreamentTrack) checkAndMarkInsertStmtByScope(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {

		id := t.trackScopes.Search(line)
		if id == -1 {
			return
		}
		trackScope := t.trackScopes[id].Search(line)
		if _, ok := t.visitedTrackScopes[TrackScopeKey{StartLine: trackScope.StartLine, EndLine: trackScope.EndLine}]; ok {
			return
		}
		t.visitedTrackScopes[TrackScopeKey{StartLine: trackScope.StartLine, EndLine: trackScope.EndLine}] = struct{}{}
		t.markInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndMarkInsertStmtByFunc(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {
		var valid bool
		line, valid = t.getInsertPositionInFunctionBody(line)
		if !valid {
			return
		}
		t.markInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndMarkInsertByLine(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {
		t.markInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndMarkInsertStmtByBlock(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {
		// check if the content between the last inserted line and the current line are all comments or empty lines
		next := t.lastBlockInsertLine + 1
		var content string
		for next < line {
			content = strings.TrimSpace(t.source[next-1])
			if content == "" {
				next++
				continue
			}

			if utils.IsGoComment(content) {
				next++
				continue
			}

			break
		}

		if next != line {
			t.markInsert(position, CodeInsertTypeStmt, line)
		} else {
			lastInsertBlockScope := t.blockScopes.Search(t.lastBlockInsertLine)
			currentBlockScope := t.blockScopes.Search(line)
			if lastInsertBlockScope != currentBlockScope {
				t.markInsert(position, CodeInsertTypeStmt, line)
			}
		}
		t.lastBlockInsertLine = line
	}
}

func (t *IncreamentTrack) adjustLastBlockInsertLine(from, to int) {
	if t.granularity.IsBlock() {
		if t.lastBlockInsertLine == from {
			t.lastBlockInsertLine = to
		}
	}
}
func (t *IncreamentTrack) isLineChanged(line int) bool {
	lineChange := t.fileChange.LineChanges.Search(line)
	return lineChange != -1
}

func (t *IncreamentTrack) isLineChangedRange(start, end int) bool {
	for i := start; i <= end; i++ {
		if t.isLineChanged(i) {
			return true
		}
	}
	return false
}

func (t *IncreamentTrack) markSpecialInsert(position CodeInsertPosition, codeType CodeInsertType, line int) {
	if t.granularity.IsFunc() {
		var valid bool
		line, valid = t.getInsertPositionInFunctionBody(line)
		if !valid {
			return
		}
	} else if t.granularity.IsScope() {
		id := t.trackScopes.Search(line)
		if id == -1 {
			return
		}
		trackScope := t.trackScopes[id].Search(line)
		if _, ok := t.visitedTrackScopes[TrackScopeKey{StartLine: trackScope.StartLine, EndLine: trackScope.EndLine}]; ok {
			return
		}
		t.visitedTrackScopes[TrackScopeKey{StartLine: trackScope.StartLine, EndLine: trackScope.EndLine}] = struct{}{}
	}
	t.markInsert(position, codeType, line)
}

func (t *IncreamentTrack) markInsert(position CodeInsertPosition, codeType CodeInsertType, line int) {
	for utils.IsGoComment(t.source[line-1]) {
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
	key := InsertPosition{position: position, codeType: codeType, line: line}
	if _, ok := t.visitedInsertedPositions[key]; ok {
		return
	}
	t.insertedPositions.Insert(position, codeType, line, 0)
	t.visitedInsertedPositions[key] = struct{}{}
	t.count++
}

func (t *IncreamentTrack) isInFunctionScopes(line int) bool {
	return t.functionScopes.Search(line) > 0
}

func (t *IncreamentTrack) getInsertPositionInFunctionBody(line int) (insertLine int, valid bool) {
	idx := t.functionScopes.Search(line)
	if idx == 0 {
		return -1, false
	}
	return t.functionScopes[idx].StartLine + 1, true
}

func (t *IncreamentTrack) TargetFile() string {
	return t.fileName
}

func (t *IncreamentTrack) Track() (int, error) {
	var err error
	t.insertedPositions.Reset()
	t.singleLineInsertedPositions.Reset()
	clear(t.visitedInsertedPositions)
	log.Debugf("Adding tracking to file: %s", t.fileName)
	_, err = t.addStmts()
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
			log.Errorf("failed to add import: %v", err)
			return 0, err
		}
		t.content = content

		// do provider insert
		log.Debugf("Adding provider import to file: %s", t.fileName)
		t.insertedPositions.Reset()
		t.singleLineInsertedPositions.Reset()
		clear(t.visitedInsertedPositions)
		pkgPath, alias := t.provider.ImportSpec()
		content, err = utils.AddImport(t.printerConfig, pkgPath, alias, t.fileName, t.content)
		if err != nil {
			log.Errorf("Failed to add import: %v", err)
			return 0, err
		}
		t.content = content
	}
	log.Debugf("Added %d tracking points to %s", t.count, t.fileName)
	return t.count, nil
}

func (t *IncreamentTrack) Count() int {
	return t.count
}

func (t *IncreamentTrack) Replace(target string, replace func(older string) (newer string)) (int, error) {
	if len(t.content) == 0 {
		return 0, fmt.Errorf("no content to calibrate")
	}

	content := string(t.content)
	count, newContent, err := utils.Replace(content, target, replace)
	if err != nil {
		return 0, err
	}
	t.content = []byte(newContent)
	return count, nil
}

func (t *IncreamentTrack) Bytes() []byte {
	return t.content
}

func (t *IncreamentTrack) Save(path string) error {
	perm := os.FileMode(0644)
	if path == "" {
		path = t.fileName
		log.Debugf("Applying tracking to file: %s", path)
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Errorf("Failed to get file info: %v, file: %s", err, path)
			return err
		}
		if fileInfo.IsDir() {
			log.Errorf("Path is a directory: %s", path)
			return fmt.Errorf("path is a directory: %s", path)
		}
		perm = fileInfo.Mode().Perm()
	}
	return os.WriteFile(path, t.content, perm)
}

func (t *IncreamentTrack) addStmts() ([]byte, error) {
	log.Debugf("Adding tracking to file: %s", t.fileName)
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
				t.insertSingleLineStmt(pos, CodeInsertTypeStmt, CodeInsertPositionFront)
			}
			t.processStatements(decl.Body.List, fset)
			t.processSpecialStatements(decl.Body, fset)
		case *ast.GenDecl:
			t.processGlobalValueSpecs(decl.Specs, fset)
			t.processGlobalFunctionLit(decl.Specs, fset)
		}
	}
	return t.doInsert()
}

func (t *IncreamentTrack) processGlobalValueSpecs(specs []ast.Spec, fset *token.FileSet) {
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
							t.insertSingleLineStmt(pos, CodeInsertTypeStmt, CodeInsertPositionFront)
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

func (t *IncreamentTrack) processGlobalFunctionLit(specs []ast.Spec, fset *token.FileSet) {
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
							t.processSpecialStatements(n.Body, fset)
						}
						return false
					}
					return true
				})
			}
		}
	}
}

func (t *IncreamentTrack) processSpecialStatements(node ast.Node, fset *token.FileSet) {
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
				t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
				if n.Else != nil {
					elseStmt, ok := n.Else.(*ast.BlockStmt)
					if ok && len(elseStmt.List) > 0 {
						t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
							fset.Position(elseStmt.Lbrace).Line+1)
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
						t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
							fset.Position(caseClause.Colon).Line+1)
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
						t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
							fset.Position(caseClause.Colon).Line+1)
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
				t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Colon).Line+1)
			}

		case *ast.CommClause:
			// return true
			changed = t.isLineChanged(fset.Position(n.Case).Line)
			if !changed && n.Comm != nil {
				changed = t.isLineChangedRange(fset.Position(n.Comm.Pos()).Line, fset.Position(n.Comm.End()).Line)
			}
			if changed {
				t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Colon).Line+1)
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
				t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
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
				t.markSpecialInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
			}
		}
		return true
	})
}

// processStatements analyzes and modifies statements by inserting additional code
// nodes before each statement in the function body.
func (t *IncreamentTrack) processStatements(statList []ast.Stmt, fset *token.FileSet) {
	for _, stmt := range statList {
		if stmt == nil {
			continue
		}
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.adjustLastBlockInsertLine(fset.Position(s.Pos()).Line, fset.Position(s.End()).Line)
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
			t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.List, fset)
		case *ast.ReturnStmt:
			t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.analyzeAndModifyExpr(s.Results, fset)
		case *ast.DeferStmt:
			if s.Call != nil && s.Call.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)
			}
		case *ast.SelectStmt:
			if s.Body != nil {
				t.processStatements(s.Body.List, fset)
			}
		case *ast.GoStmt:
			t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
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
				t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
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
			t.checkAndMarkInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
		}
	}
}

// analyzeAndModifyExpr analyzes and modifies expressions by processing any function literals found.
// It works in conjunction with processStatements to recursively handle nested expressions.
func (t *IncreamentTrack) analyzeAndModifyExpr(exprList []ast.Expr, fset *token.FileSet) {
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
				t.insertSingleLineStmt(pos, CodeInsertTypeStmt, CodeInsertPositionFront)
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

func (t *IncreamentTrack) insertSingleLineStmt(pos token.Position, codeType CodeInsertType, position CodeInsertPosition) {
	if t.isLineChanged(pos.Line) {
		t.count++
		t.singleLineInsertedPositions.Insert(position, codeType, pos.Line, pos.Column)
	}
}

func calculateInsertLength(insertedPositions InsertPositions, defaultInserts []string, frontStmts []string,
	backStmts []string, frontComments []string, backComments []string) int {
	insertLen := 0
	for _, pos := range insertedPositions {

		for _, stmt := range defaultInserts {
			insertLen += len(stmt) + 1
		}
		if pos.position.IsFront() {
			for _, comment := range frontComments {
				insertLen += len(comment) + 1
			}
			for _, stmt := range frontStmts {
				insertLen += len(stmt) + 1
			}
		} else {
			for _, comment := range backComments {
				insertLen += len(comment) + 1
			}
			for _, stmt := range backStmts {
				insertLen += len(stmt) + 1
			}
		}
	}
	return insertLen
}

func doInsert(buf *bytes.Buffer, pos InsertPosition, trackStmtPlaceHolders []string,
	frontStmts []string, backStmts []string, frontComments []string, backComments []string) int {
	// write default track stmt place holders
	lines := 0
	for _, trackStmtPlaceHolder := range trackStmtPlaceHolders {
		buf.WriteString(trackStmtPlaceHolder)
		buf.WriteByte('\n')
		lines++
	}
	// write user defined provider insert
	if pos.position.IsFront() {
		if len(frontStmts) > 0 {
			buf.WriteString(config.TrackUserComment)
			buf.WriteByte('\n')
			buf.WriteString(config.TrackTipsComment)
			buf.WriteByte('\n')
			lines += 2
			for _, content := range frontComments {
				buf.WriteString(content)
				buf.WriteByte('\n')
				lines++
			}
			for _, content := range frontStmts {
				buf.WriteString(content)
				buf.WriteByte('\n')
				lines++
			}
			buf.WriteString(config.TrackEndComment)
			buf.WriteByte('\n')
			lines++
		}

	} else {
		if len(backStmts) > 0 {
			buf.WriteString(config.TrackUserComment)
			buf.WriteByte('\n')
			buf.WriteString(config.TrackTipsComment)
			buf.WriteByte('\n')
			lines += 2
			for _, content := range backComments {
				buf.WriteString(content)
				buf.WriteByte('\n')
				lines++
			}
			for _, content := range backStmts {
				buf.WriteString(content)
				buf.WriteByte('\n')
				lines++
			}
			buf.WriteString(config.TrackEndComment)
			buf.WriteByte('\n')
			lines++
		}
	}
	return lines
}

// --- Interface Implementations ---

// incrementTemplateProvider implements TrackTemplateProvider
type incrementTemplateProvider struct {
	frontCodeProvider TrackCodeProvider
	backCodeProvider  TrackCodeProvider
}

func defaultIncrementTemplateProvider() *incrementTemplateProvider {
	return &incrementTemplateProvider{
		frontCodeProvider: &incrementCodeProvider{position: CodeInsertPositionFront},
		backCodeProvider:  &incrementCodeProvider{position: CodeInsertPositionBack},
	}
}

func (p *incrementTemplateProvider) ImportSpec() (pkgPath, alias string) {
	// Example: Use "fmt" package for Println
	return "", ""
}

func (p *incrementTemplateProvider) FrontTrackCodeProvider() TrackCodeProvider {
	// Return a slice containing instances of TrackCodeProvider
	return p.frontCodeProvider
}

func (p *incrementTemplateProvider) BackTrackCodeProvider() TrackCodeProvider {
	// Return a slice containing instances of TrackCodeProvider
	return p.backCodeProvider
}

// incrementCodeProvider implements TrackCodeProvider
type incrementCodeProvider struct {
	position CodeInsertPosition
	// Add other fields if needed, e.g., specific comments or statement details
}

func (p *incrementCodeProvider) Position() CodeInsertPosition {
	return p.position
}

func (p *incrementCodeProvider) Comments() []string {
	// Example: Return a specific comment or nil
	return []string{}
}

func (p *incrementCodeProvider) Stmts() []string {
	// Example: Return the string representation of the statement
	// Ideally, this should format the Stmt() result, but for simplicity:
	return []string{}
}

// Ensure the new types implement the interfaces (compile-time check)
var _ TrackTemplateProvider = (*incrementTemplateProvider)(nil)
var _ TrackCodeProvider = (*incrementCodeProvider)(nil)
