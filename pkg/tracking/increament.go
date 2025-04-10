package tracking

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/monshunter/goat/pkg/diff"
	"golang.org/x/tools/go/ast/astutil"
)

type IncreamentTrack struct {
	basePath            string
	fileChange          *diff.FileChange
	provider            TrackTemplateProvider
	count               int
	content             []byte
	fileName            string
	positionInserts     InsertPositions
	lastBlockInsertLine int
	granularity         Granularity
}

func NewIncreamentTrack(basePath string, fileChange *diff.FileChange,
	provider TrackTemplateProvider, granularity Granularity) (*IncreamentTrack, error) {
	fileName := fileChange.Path
	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(basePath, fileName)
	}
	fileName = filepath.Clean(fileName)
	if provider == nil {
		provider = defaultIncrementTemplateProvider()
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return &IncreamentTrack{
		basePath:        basePath,
		fileChange:      fileChange,
		provider:        provider,
		fileName:        fileName,
		content:         content,
		positionInserts: InsertPositions{},
		granularity:     granularity,
	}, nil
}

func (t *IncreamentTrack) getAstTree() (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", t.content, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

func (t *IncreamentTrack) formatOutput(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4, Indent: 0}
	err := cfg.Fprint(&buf, fset, f)
	if err != nil {
		return nil, err
	}
	t.content = buf.Bytes()
	return t.content, nil
}

func (t *IncreamentTrack) doInsert(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var src bytes.Buffer
	if err := printer.Fprint(&src, fset, f); err != nil {
		return nil, err
	}
	frontStmts, backStmts, frontComments, backComments, insertLen := t.getContentsToInsert()

	srcStr := src.String()
	positionInsert := t.positionInserts
	positionInsert.Sort()

	// 对于每个插入位置，将打印语句插入到源代码字符串中
	var buf bytes.Buffer
	buf.Grow(len(srcStr) + insertLen)
	posIdx := 0
	lines := 0
	i, j := 0, 0
	for ; i < len(srcStr) && posIdx < len(positionInsert); i++ {
		if lines == positionInsert[posIdx].positions-1 {
			pos := positionInsert[posIdx]
			buf.WriteString(srcStr[j:i])
			if pos.position.IsFront() {
				if len(frontStmts) > 0 {
					buf.WriteString("// +goat:start")
					buf.WriteByte('\n')
					for _, content := range frontComments {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					for _, content := range frontStmts {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					buf.WriteString("// +goat:end")
					buf.WriteByte('\n')
				}

			} else {
				if len(backStmts) > 0 {
					buf.WriteString("// +goat:start")
					buf.WriteByte('\n')
					for _, content := range backComments {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					for _, content := range backStmts {
						buf.WriteString(content)
						buf.WriteByte('\n')
					}
					buf.WriteString("// +goat:end")
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
	newFset, newF, err := t.getAstTree()
	if err != nil {
		return nil, err
	}
	return t.formatOutput(newFset, newF)
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
	lineChange := t.fileChange.LineChanges.Search(line)
	if lineChange != -1 {
		t.addInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndInsertByLine(position CodeInsertPosition, line int) {
	lineChange := t.fileChange.LineChanges.Search(line)
	if lineChange != -1 {
		t.addInsert(position, CodeInsertTypeStmt, line)
	}
}

func (t *IncreamentTrack) checkAndInsertStmtByBlock(position CodeInsertPosition, line int) {
	lineChange := t.fileChange.LineChanges.Search(line)
	if lineChange != -1 {
		if t.lastBlockInsertLine != line-1 {
			t.addInsert(position, CodeInsertTypeStmt, line)
		}
		t.lastBlockInsertLine = line
	}
}

func (t *IncreamentTrack) addInsert(position CodeInsertPosition, codeType CodeInsertType, line int) {
	t.positionInserts.Insert(position, codeType, line)
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
		t.positionInserts.Reset()
		_, err = t.addImport()
		if err != nil {
			return 0, err
		}
	}
	return t.count, nil
}

func (t *IncreamentTrack) Bytes() []byte {
	return t.content
}

func (t *IncreamentTrack) addImport() ([]byte, error) {
	fset, f, err := t.getAstTree()
	if err != nil {
		return nil, err
	}
	_, pkgPath, alias := t.provider.ImportSpec()
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
	return t.formatOutput(fset, f)
}

func (t *IncreamentTrack) addStmts() ([]byte, error) {
	fset, f, err := t.getAstTree()
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.FuncDecl); ok {
			t.processStatements(decl.Body.List, fset)
		}
	}
	return t.doInsert(fset, f)
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
			s.Rhs = t.analyzeAndModifyExpr(s.Rhs, fset)
		case *ast.IfStmt:
			t.processStatements(s.Body.List, fset)
			if s.Else != nil {
				switch s.Else.(type) {
				case *ast.IfStmt:
					t.processStatements([]ast.Stmt{s.Else.(*ast.IfStmt)}, fset)
				case *ast.BlockStmt:
					block := s.Else.(*ast.BlockStmt)
					t.processStatements(block.List, fset)
					s.Else = block
				}
			}
		case *ast.ForStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.RangeStmt:

			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.SwitchStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
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
			for i, result := range s.Results {
				s.Results[i] = t.analyzeAndModifyExpr([]ast.Expr{result}, fset)[0]
			}
		case *ast.DeferStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.SelectStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.GoStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.TypeSwitchStmt:
			t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
			t.processStatements(s.Body.List, fset)
		case *ast.ExprStmt:
			switch s.X.(type) {
			case *ast.CallExpr:
				t.checkAndInsert(CodeInsertPositionFront, fset.Position(s.Pos()).Line)
				expr := s.X.(*ast.CallExpr)
				if expr.Fun != nil {
					expr.Fun = t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)[0]
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
func (t *IncreamentTrack) analyzeAndModifyExpr(exprList []ast.Expr, fset *token.FileSet) []ast.Expr {
	newExprList := make([]ast.Expr, 0, len(exprList))
	for _, expr := range exprList {
		switch expr := expr.(type) {
		case *ast.FuncLit:
			t.processStatements(expr.Body.List, fset)
		}
		newExprList = append(newExprList, expr)
	}
	return newExprList
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
		// backCodeProvider:  &incrementCodeProvider{codeType: CodeInjectTypeBack},
	}
}

func (p *incrementTemplateProvider) ImportSpec() (pkgName, pkgPath, alias string) {
	// Example: Use "fmt" package for Println
	return "context", "context", "ctx"
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
	return []string{"// +goat:track"}
}

func (p *incrementCodeProvider) Stmts() []string {
	// Example: Return the string representation of the statement
	// Ideally, this should format the Stmt() result, but for simplicity:
	return []string{`goat.Track(TRACK_ID)`}
}

// Ensure the new types implement the interfaces (compile-time check)
var _ TrackTemplateProvider = (*incrementTemplateProvider)(nil)
var _ TrackCodeProvider = (*incrementCodeProvider)(nil)
