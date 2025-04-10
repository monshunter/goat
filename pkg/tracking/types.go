package tracking

import "go/ast"

// Tracker
type Tracker interface {
	Track() (int, error)
	Bytes() []byte
}

type CodeInjectType int

const (
	CodeInjectTypeFront CodeInjectType = 1
	CodeInjectTypeBack  CodeInjectType = 2
)

func (c CodeInjectType) String() string {
	return []string{"front", "back"}[c-1]
}

func (c CodeInjectType) Int() int {
	return int(c)
}

func (c CodeInjectType) IsFront() bool {
	return c == CodeInjectTypeFront
}

func (c CodeInjectType) IsBack() bool {
	return c == CodeInjectTypeBack
}

type TrackCodeProvider interface {
	Type() CodeInjectType
	Comments() *ast.CommentGroup
	Stmt() ast.Stmt
	StmtValue() string
}

type TrackTemplateProvider interface {
	ImportSpec() (pkgName, pkgPath, alias string)
	FrontTrackCodeProvider() TrackCodeProvider
	BackTrackCodeProvider() TrackCodeProvider
}
