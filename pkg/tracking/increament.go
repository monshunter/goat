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

	"github.com/monshunter/goat/pkg/config"
	"github.com/monshunter/goat/pkg/diff"
	"github.com/monshunter/goat/pkg/tracking/increament"
	"github.com/monshunter/goat/pkg/utils"
)

type IncreamentTrack struct {
	basePath               string
	fileChange             *diff.FileChange
	provider               TrackTemplateProvider
	count                  int
	content                []byte
	fileName               string
	positionInserts        InsertPositions
	visitedPositionInserts map[InsertPosition]struct{}
	lastBlockInsertLine    int
	granularity            config.Granularity
	importPathPlaceHolder  string
	trackStmtPlaceHolders  []string
	source                 []string
	blockScopes            BlockScopes
	printerConfig          *printer.Config
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
		return nil, err
	}

	blockScopes, err := BlockScopesOfGoAST(fileName, content)
	if err != nil {
		return nil, err
	}
	return &IncreamentTrack{
		basePath:               basePath,
		fileChange:             fileChange,
		provider:               provider,
		fileName:               fileName,
		content:                content,
		positionInserts:        InsertPositions{},
		granularity:            granularity,
		importPathPlaceHolder:  importPathPlaceHolder,
		trackStmtPlaceHolders:  trackStmtPlaceHolders,
		source:                 strings.Split(string(content), "\n"),
		blockScopes:            blockScopes,
		visitedPositionInserts: make(map[InsertPosition]struct{}),
		printerConfig:          printerConfig,
	}, nil
}

func (t *IncreamentTrack) doInsert(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var src bytes.Buffer
	if err := printer.Fprint(&src, fset, f); err != nil {
		return nil, err
	}
	frontStmts, backStmts, frontComments, backComments, insertLen := t.getContentsToInsert()

	srcStr := src.String()
	positionInsert := t.positionInserts
	positionInsert.Unique()
	positionInsert.Sort()
	// For each insertion position, insert the print statement into the source code string
	var buf bytes.Buffer
	buf.Grow(len(srcStr) + insertLen)
	posIdx := 0
	lines := 0
	i, j := 0, 0
	for ; i < len(srcStr) && posIdx < len(positionInsert); i++ {
		if lines == positionInsert[posIdx].positions-1 {
			pos := positionInsert[posIdx]
			buf.WriteString(srcStr[j:i])
			// write default track stmt place holders
			for _, trackStmtPlaceHolder := range t.trackStmtPlaceHolders {
				buf.WriteString(trackStmtPlaceHolder)
				buf.WriteByte('\n')
			}
			// write user defined provider insert
			if pos.position.IsFront() {
				if len(frontStmts) > 0 {
					buf.WriteString(config.TrackUserComment)
					buf.WriteByte('\n')
					buf.WriteString(config.TrackTipsComment)
					buf.WriteByte('\n')
					for _, content := range frontComments {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					for _, content := range frontStmts {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					buf.WriteString(config.TrackEndComment)
					buf.WriteByte('\n')
				}

			} else {
				if len(backStmts) > 0 {
					buf.WriteString(config.TrackUserComment)
					buf.WriteByte('\n')
					buf.WriteString(config.TrackTipsComment)
					buf.WriteByte('\n')
					for _, content := range backComments {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					for _, content := range backStmts {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					buf.WriteString(config.TrackEndComment)
					buf.WriteByte('\n')
				}
			}
			j = i
			posIdx++
		} else if srcStr[i] == '\n' {
			lines++
		}
	}
	buf.WriteString(srcStr[i:])
	t.content = buf.Bytes()
	newFset, newF, err := utils.GetAstTree("", t.content)
	if err != nil {
		return nil, err
	}
	content, err := utils.FormatAst(t.printerConfig, newFset, newF)
	if err != nil {
		return nil, err
	}
	t.content = content
	return t.content, nil
}

func (t *IncreamentTrack) getContentsToInsert() (
	frontStmts []string, backStmts []string, frontComments []string, backComments []string, insertLen int) {
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

	for _, pos := range t.positionInserts {
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
	return
}

func (t *IncreamentTrack) checkAndInsert(position CodeInsertPosition, line int) {
	if t.granularity.IsLine() {
		t.checkAndInsertByLine(position, line)
	} else if t.granularity.IsBlock() {
		t.checkAndInsertStmtByBlock(position, line)
	} else if t.granularity.IsFunc() {
		t.checkAndInsertStmtByFunc(position, line)
	}
}

func (t *IncreamentTrack) checkAndInsertStmtByFunc(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {
		t.addInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndInsertByLine(position CodeInsertPosition, line int) {
	if t.isLineChanged(line) {
		t.addInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndInsertStmtByBlock(position CodeInsertPosition, line int) {
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
			t.addInsert(position, CodeInsertTypeStmt, line)
		} else {
			lastInsertBlockScope := t.blockScopes.Search(t.lastBlockInsertLine)
			currentBlockScope := t.blockScopes.Search(line)
			if lastInsertBlockScope != currentBlockScope {
				t.addInsert(position, CodeInsertTypeStmt, line)
			}
		}
		t.lastBlockInsertLine = line
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

func (t *IncreamentTrack) addInsert(position CodeInsertPosition, codeType CodeInsertType, line int) {

	for utils.IsGoComment(t.source[line-1]) {
		line++
	}
	// Add a check to avoid duplicate inserts
	// Use visitedPositionInserts map to record the inserted positions,
	// This can prevent duplicate inserts in multiple AST scans.
	// It is important for ensuring the correctness and avoiding unnecessary duplicate tracking.
	key := InsertPosition{position: position, codeType: codeType, positions: line}
	if _, ok := t.visitedPositionInserts[key]; ok {
		return
	}
	t.positionInserts.Insert(position, codeType, line)
	t.visitedPositionInserts[key] = struct{}{}
	t.count++
}

func (t *IncreamentTrack) Track() (int, error) {
	var err error
	t.positionInserts.Reset()
	_, err = t.addStmts()
	if err != nil {
		return 0, err
	}
	if t.count > 0 {
		// do default insert
		t.positionInserts.Reset()
		content, err := utils.AddImport(t.printerConfig, t.importPathPlaceHolder, "", t.fileName, t.content)
		if err != nil {
			return 0, err
		}
		t.content = content

		// do provider insert
		t.positionInserts.Reset()
		pkgPath, alias := t.provider.ImportSpec()
		content, err = utils.AddImport(t.printerConfig, pkgPath, alias, t.fileName, t.content)
		if err != nil {
			return 0, err
		}
		t.content = content
	}
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
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("path is a directory: %s", path)
		}
		perm = fileInfo.Mode().Perm()
	}
	return os.WriteFile(path, t.content, perm)
}

func (t *IncreamentTrack) addStmts() ([]byte, error) {
	fset, f, err := utils.GetAstTree("", t.content)
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.FuncDecl); ok {
			t.processStatements(decl.Body.List, fset)
			t.processSpecialStatements(decl.Body, fset)
		}
	}
	return t.doInsert(fset, f)
}

func (t *IncreamentTrack) processSpecialStatements(node ast.Node, fset *token.FileSet) {
	ast.Inspect(node, func(n ast.Node) bool {
		var changed bool
		switch n := n.(type) {
		case *ast.IfStmt:
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
				t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
				if n.Else != nil {
					elseStmt, ok := n.Else.(*ast.BlockStmt)
					if ok {
						t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
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
			if changed {
				for _, subStmt := range n.Body.List {
					if caseClause, ok := subStmt.(*ast.CaseClause); ok && len(caseClause.Body) > 0 {
						t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
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
				changed = t.isLineChangedRange(fset.Position(n.Assign.Pos()).Line,

					fset.Position(n.Assign.End()).Line)
			}

			// Judge if the type switch statement is changed:
			// 1. If the type switch statement is changed, it means the whole type switch statement may be modified
			// 2. Even if the Init and Assign parts are not changed, the change of the type switch statement may affect the whole statement structure
			// 3. In this case, we need to add tracking statements to each case clause
			if changed {
				for _, subStmt := range n.Body.List {
					if caseClause, ok := subStmt.(*ast.CaseClause); ok && len(caseClause.Body) > 0 {
						t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt,
							fset.Position(caseClause.Colon).Line+1)
					}
				}
			}
		case *ast.CaseClause:
			// return true
			changed = t.isLineChanged(fset.Position(n.Case).Line)
			if !changed && len(n.List) > 0 {
				for _, expr := range n.List {
					changed = t.isLineChangedRange(fset.Position(expr.Pos()).Line,
						fset.Position(expr.End()).Line)
					if changed {
						break
					}
				}
			}
			if changed {
				t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Colon).Line+1)
			}

		case *ast.CommClause:
			// return true
			changed = t.isLineChanged(fset.Position(n.Case).Line)
			if !changed && n.Comm != nil {
				changed = t.isLineChangedRange(fset.Position(n.Comm.Pos()).Line,
					fset.Position(n.Comm.End()).Line)
			}
			if changed {
				t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Colon).Line+1)
			}
		case *ast.RangeStmt:
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
				changed = t.isLineChangedRange(fset.Position(n.X.Pos()).Line,
					fset.Position(n.X.End()).Line)
			}

			if changed {
				t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
			}
		case *ast.ForStmt:
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
				t.addInsert(CodeInsertPositionFront, CodeInsertTypeStmt, fset.Position(n.Body.Lbrace).Line+1)
			}
		}
		return true
	})
}

// processStatements analyzes and modifies statements by inserting additional code
// nodes before each statement in the function body.
func (t *IncreamentTrack) processStatements(statList []ast.Stmt, fset *token.FileSet) {
	// 遍历函数体中的语句
	for _, stmt := range statList {
		//可以根据语句类型进一步处理
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.analyzeAndModifyExpr(s.Rhs, fset)
		case *ast.IfStmt:
			t.processStatements(s.Body.List, fset)
			if s.Else != nil {
				switch s.Else.(type) {
				case *ast.IfStmt:
					t.processStatements([]ast.Stmt{s.Else.(*ast.IfStmt)}, fset)
				case *ast.BlockStmt:
					t.processStatements(s.Else.(*ast.BlockStmt).List, fset)
				}
			}
		case *ast.ForStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.RangeStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.SwitchStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.CommClause:
			t.processStatements(s.Body, fset)
		case *ast.CaseClause:
			t.processStatements(s.Body, fset)
		case *ast.BlockStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.List, fset)
		case *ast.ReturnStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.analyzeAndModifyExpr(s.Results, fset)
		case *ast.DeferStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)
			}
		case *ast.SelectStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.GoStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)
			}
		case *ast.TypeSwitchStmt:
			// t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.ExprStmt:
			switch s.X.(type) {
			case *ast.CallExpr:
				t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
				expr := s.X.(*ast.CallExpr)
				if expr.Fun != nil {
					t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)
				}
				if expr.Args != nil {
					t.analyzeAndModifyExpr(expr.Args, fset)
				}
			default:
			}
		default:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
		}
	}
}

// analyzeAndModifyExpr analyzes and modifies expressions by processing any function literals found.
// It works in conjunction with processStatements to recursively handle nested expressions.
func (t *IncreamentTrack) analyzeAndModifyExpr(exprList []ast.Expr, fset *token.FileSet) {
	for _, expr := range exprList {
		switch expr := expr.(type) {
		case *ast.FuncLit:
			t.processStatements(expr.Body.List, fset)
		case *ast.CallExpr:
			if expr.Fun != nil {
				t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)
			}
			if expr.Args != nil {
				t.analyzeAndModifyExpr(expr.Args, fset)
			}
		}
	}
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
