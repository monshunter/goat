package tracking

import "slices"

// Tracker
type Tracker interface {
	Track() (int, error)
	Replace(target string, replace func(older string) (newer string)) (int, error)
	Bytes() []byte
	Save(path string) error
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
	position  CodeInsertPosition
	codeType  CodeInsertType
	positions int
}

type InsertPositions []InsertPosition

func (p *InsertPositions) Insert(position CodeInsertPosition, codeType CodeInsertType, positions int) {
	*p = append(*p, InsertPosition{position: position, codeType: codeType, positions: positions})
}

func (p *InsertPositions) Sort() {
	slices.SortFunc(*p, func(a, b InsertPosition) int {
		return a.positions - b.positions
	})
}
func (p *InsertPositions) Reset() {
	*p = InsertPositions{}
}
