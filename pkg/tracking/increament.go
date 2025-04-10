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
	basePath   string
	fileChange *diff.FileChange
	provider   TrackTemplateProvider
	count      int
	content    []byte
	fileName   string
}

func NewIncreamentTrack(basePath string, fileChange *diff.FileChange, provider TrackTemplateProvider) (*IncreamentTrack, error) {
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
		basePath:   basePath,
		fileChange: fileChange,
		provider:   provider,
		fileName:   fileName,
		content:    content,
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

func (t *IncreamentTrack) formatAST(fset *token.FileSet, f *ast.File) ([]byte, error) {
	var buf bytes.Buffer
	// cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4, Indent: 0}
	err := cfg.Fprint(&buf, fset, f)
	if err != nil {
		return nil, err
	}
	t.content = buf.Bytes()
	return t.content, nil
}

func (t *IncreamentTrack) Track() (int, error) {
	var err error
	_, err = t.addImport()
	if err != nil {
		return 0, err
	}
	_, err = t.addStmts()
	if err != nil {
		return 0, err
	}
	_, err = t.addComments()
	if err != nil {
		return 0, err
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
	return t.formatAST(fset, f)
}

func (t *IncreamentTrack) addStmts() ([]byte, error) {
	fset, f, err := t.getAstTree()
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.FuncDecl); ok {
			decl.Body.List = t.processStatements(decl.Body.List, fset)
		}
	}
	return t.formatAST(fset, f)
}

// processStatements analyzes and modifies statements by inserting additional code
// nodes before each statement in the function body.
func (t *IncreamentTrack) processStatements(statList []ast.Stmt, fset *token.FileSet) []ast.Stmt {
	// 遍历函数体中的语句
	newStatList := make([]ast.Stmt, 0, len(statList))
	for _, stmt := range statList {
		//可以根据语句类型进一步处理

		switch s := stmt.(type) {
		case *ast.AssignStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Rhs = t.analyzeAndModifyExpr(s.Rhs, fset)
		case *ast.IfStmt:
			s.Body.List = t.processStatements(s.Body.List, fset)
			if s.Else != nil {
				switch s.Else.(type) {
				case *ast.IfStmt:
					s.Else = t.processStatements([]ast.Stmt{s.Else.(*ast.IfStmt)}, fset)[0]
				case *ast.BlockStmt:
					block := s.Else.(*ast.BlockStmt)
					block.List = t.processStatements(block.List, fset)
					s.Else = block
				}
			}
		case *ast.ForStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Body.List = t.processStatements(s.Body.List, fset)
		case *ast.RangeStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Body.List = t.processStatements(s.Body.List, fset)
		case *ast.SwitchStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Body.List = t.processStatements(s.Body.List, fset)
		case *ast.CommClause:
			s.Body = t.processStatements(s.Body, fset)
		case *ast.CaseClause:
			s.Body = t.processStatements(s.Body, fset)
		case *ast.BlockStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.List = t.processStatements(s.List, fset)
		case *ast.ReturnStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			for i, result := range s.Results {
				s.Results[i] = t.analyzeAndModifyExpr([]ast.Expr{result}, fset)[0]
			}
		case *ast.DeferStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.SelectStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Body.List = t.processStatements(s.Body.List, fset)
		case *ast.GoStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExpr([]ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.TypeSwitchStmt:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
			s.Body.List = t.processStatements(s.Body.List, fset)
		case *ast.ExprStmt:
			switch s.X.(type) {
			case *ast.CallExpr:
				ns := t.newStmt(CodeInjectTypeFront)
				if ns != nil {
					newStatList = append(newStatList, ns)
				}
				expr := s.X.(*ast.CallExpr)
				if expr.Fun != nil {
					expr.Fun = t.analyzeAndModifyExpr([]ast.Expr{expr.Fun}, fset)[0]
				}
			default:
			}
		default:
			ns := t.newStmt(CodeInjectTypeFront)
			if ns != nil {
				newStatList = append(newStatList, ns)
			}
		}
		newStatList = append(newStatList, stmt)
	}
	return newStatList
}

// analyzeAndModifyExpr analyzes and modifies expressions by processing any function literals found.
// It works in conjunction with processStatements to recursively handle nested expressions.
func (t *IncreamentTrack) analyzeAndModifyExpr(exprList []ast.Expr, fset *token.FileSet) []ast.Expr {
	newExprList := make([]ast.Expr, 0, len(exprList))
	for _, expr := range exprList {
		switch expr := expr.(type) {
		case *ast.FuncLit:
			expr.Body.List = t.processStatements(expr.Body.List, fset)
		}
		newExprList = append(newExprList, expr)
	}
	return newExprList
}

func (t *IncreamentTrack) addComments() ([]byte, error) {
	fset, f, err := t.getAstTree()
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.FuncDecl); ok {
			t.newComment(f, decl.Pos()-1, CodeInjectTypeFront)
			decl.Body.List = t.processComment(f, decl.Body.List, fset)
		}
	}
	return t.formatAST(fset, f)
}

func (t *IncreamentTrack) processComment(f *ast.File, statList []ast.Stmt, fset *token.FileSet) []ast.Stmt {
	// 遍历函数体中的语句
	newStatList := make([]ast.Stmt, 0, len(statList))
	for _, stmt := range statList {
		if _, ok := stmt.(*ast.IfStmt); !ok {
			t.newComment(f, stmt.Pos()-1, CodeInjectTypeFront)
		}
		// t.insertComment(f, stmt.Pos()-1)
		//可以根据语句类型进一步处理
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			s.Rhs = t.analyzeAndModifyExprAndInsertComment(f, s.Rhs, fset)
		case *ast.IfStmt:
			t.newComment(f, s.Body.Lbrace, CodeInjectTypeFront)
			s.Body.List = t.processComment(f, s.Body.List, fset)
			if s.Else != nil {
				switch s.Else.(type) {
				case *ast.IfStmt:
					s.Else = t.processComment(f, []ast.Stmt{s.Else.(*ast.IfStmt)}, fset)[0]
				case *ast.BlockStmt:
					t.newComment(f, s.Else.Pos(), CodeInjectTypeFront)
					block := s.Else.(*ast.BlockStmt)
					block.List = t.processComment(f, block.List, fset)

					s.Else = block
				}
			}
		case *ast.ForStmt:
			s.Body.List = t.processComment(f, s.Body.List, fset)
		case *ast.RangeStmt:
			s.Body.List = t.processComment(f, s.Body.List, fset)
		case *ast.SwitchStmt:
			s.Body.List = t.processComment(f, s.Body.List, fset)
		case *ast.CommClause:
			s.Body = t.processComment(f, s.Body, fset)
		case *ast.CaseClause:
			s.Body = t.processComment(f, s.Body, fset)
		case *ast.BlockStmt:
			s.List = t.processComment(f, s.List, fset)
		case *ast.ReturnStmt:
			for i, result := range s.Results {
				s.Results[i] = t.analyzeAndModifyExprAndInsertComment(f, []ast.Expr{result}, fset)[0]
			}
		case *ast.DeferStmt:
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExprAndInsertComment(f, []ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.SelectStmt:
			s.Body.List = t.processComment(f, s.Body.List, fset)
		case *ast.GoStmt:
			if s.Call != nil && s.Call.Fun != nil {
				s.Call.Fun = t.analyzeAndModifyExprAndInsertComment(f, []ast.Expr{s.Call.Fun}, fset)[0]
			}
		case *ast.TypeSwitchStmt:
			s.Body.List = t.processComment(f, s.Body.List, fset)
		case *ast.ExprStmt:
			switch s.X.(type) {
			case *ast.CallExpr:
				expr := s.X.(*ast.CallExpr)
				if expr.Fun != nil {
					expr.Fun = t.analyzeAndModifyExprAndInsertComment(f, []ast.Expr{expr.Fun}, fset)[0]
				}
			}
		}
		newStatList = append(newStatList, stmt)
	}
	return newStatList
}

func (t *IncreamentTrack) analyzeAndModifyExprAndInsertComment(f *ast.File, exprList []ast.Expr, fset *token.FileSet) []ast.Expr {
	newExprList := make([]ast.Expr, 0, len(exprList))
	for _, expr := range exprList {
		switch expr := expr.(type) {
		case *ast.FuncLit:
			expr.Body.List = t.processComment(f, expr.Body.List, fset)
		}
		newExprList = append(newExprList, expr)
	}
	return newExprList
}

func (t *IncreamentTrack) newComment(f *ast.File, pos token.Pos, codeType CodeInjectType) {
	var provider TrackCodeProvider
	if codeType == CodeInjectTypeFront {
		provider = t.provider.FrontTrackCodeProvider()
	} else {
		provider = t.provider.BackTrackCodeProvider()
	}
	if provider != nil {
		cm := provider.Comments()
		if cm != nil && len(cm.List) > 0 {
			cm.List[0].Slash = pos
			f.Comments = append(f.Comments, cm)
		}
	}
}

func (t *IncreamentTrack) newStmt(codeType CodeInjectType) ast.Stmt {
	var provider TrackCodeProvider
	if codeType == CodeInjectTypeFront {
		provider = t.provider.FrontTrackCodeProvider()
	} else {
		provider = t.provider.BackTrackCodeProvider()
	}
	if provider != nil {
		t.count++
		return provider.Stmt()
	}
	return nil
}

// --- Interface Implementations ---

// incrementTemplateProvider implements TrackTemplateProvider
type incrementTemplateProvider struct {
	frontCodeProvider TrackCodeProvider
	backCodeProvider  TrackCodeProvider
}

func defaultIncrementTemplateProvider() *incrementTemplateProvider {
	return &incrementTemplateProvider{
		frontCodeProvider: &incrementCodeProvider{codeType: CodeInjectTypeFront},
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
	codeType CodeInjectType
	// Add other fields if needed, e.g., specific comments or statement details
}

func (p *incrementCodeProvider) Type() CodeInjectType {
	return p.codeType
}

func (p *incrementCodeProvider) Comments() *ast.CommentGroup {
	// Example: Return a specific comment or nil
	return &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: "// +goat:track",
			},
		},
	}
}

func (p *incrementCodeProvider) Stmt() ast.Stmt {
	// Example: Return the fmt.Println statement node
	// This could be made more dynamic based on provider configuration
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("goat"), // Assumes "fmt" is imported
				Sel: ast.NewIdent("Track"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: `TRACK_ID`, // Example tracking message
				},
			},
		},
	}
}

func (p *incrementCodeProvider) StmtValue() string {
	// Example: Return the string representation of the statement
	// Ideally, this should format the Stmt() result, but for simplicity:
	return `goat.Track(TRACK_ID)`
}

// Ensure the new types implement the interfaces (compile-time check)
var _ TrackTemplateProvider = (*incrementTemplateProvider)(nil)
var _ TrackCodeProvider = (*incrementCodeProvider)(nil)
