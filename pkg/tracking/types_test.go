package tracking

import (
	"testing"
)

func TestBlockScopeString(t *testing.T) {
	bs := BlockScope{StartLine: 10, EndLine: 20}
	expected := "BlockScope{StartLine: 10, EndLine: 20}"
	if bs.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bs.String())
	}
}

func TestBlockScopeIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		scope    BlockScope
		expected bool
	}{
		{"empty scope", BlockScope{StartLine: 0, EndLine: 0}, true},
		{"non-empty scope", BlockScope{StartLine: 10, EndLine: 20}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.scope.IsEmpty() != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", tt.scope.IsEmpty(), tt.expected)
			}
		})
	}
}

func TestBlockScopeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		scope    BlockScope
		expected bool
	}{
		{"valid scope", BlockScope{StartLine: 10, EndLine: 20}, true},
		{"invalid scope", BlockScope{StartLine: 20, EndLine: 10}, false},
		{"equal lines", BlockScope{StartLine: 10, EndLine: 10}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.scope.IsValid() != tt.expected {
				t.Errorf("IsValid() = %v, want %v", tt.scope.IsValid(), tt.expected)
			}
		})
	}
}

func TestBlockScopeContains(t *testing.T) {
	scope := BlockScope{StartLine: 10, EndLine: 20}
	tests := []struct {
		name     string
		line     int
		expected bool
	}{
		{"line before scope", 5, false},
		{"line at start", 10, false},
		{"line in scope", 15, true},
		{"line at end", 20, false},
		{"line after scope", 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if scope.Contains(tt.line) != tt.expected {
				t.Errorf("Contains(%d) = %v, want %v", tt.line, scope.Contains(tt.line), tt.expected)
			}
		})
	}
}

func TestBlockScopeContainsRange(t *testing.T) {
	scope := BlockScope{StartLine: 10, EndLine: 20}
	tests := []struct {
		name     string
		start    int
		end      int
		expected bool
	}{
		{"range inside scope", 12, 18, true},
		{"range outside scope", 5, 8, false},
		{"range overlapping start", 8, 15, false},
		{"range overlapping end", 15, 25, false},
		{"exact range", 10, 20, false},
		{"larger range", 5, 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if scope.ContainsRange(tt.start, tt.end) != tt.expected {
				t.Errorf("ContainsRange(%d, %d) = %v, want %v", tt.start, tt.end, scope.ContainsRange(tt.start, tt.end), tt.expected)
			}
		})
	}
}

func TestBlockScopesSort(t *testing.T) {
	scopes := BlockScopes{
		{StartLine: 30, EndLine: 40},
		{StartLine: 10, EndLine: 20},
		{StartLine: 20, EndLine: 30},
		{StartLine: 10, EndLine: 15},
	}

	expected := BlockScopes{
		{StartLine: 10, EndLine: 15},
		{StartLine: 10, EndLine: 20},
		{StartLine: 20, EndLine: 30},
		{StartLine: 30, EndLine: 40},
	}

	scopes.Sort()

	for i, scope := range scopes {
		if scope.StartLine != expected[i].StartLine || scope.EndLine != expected[i].EndLine {
			t.Errorf("Sort() index %d = %v, want %v", i, scope, expected[i])
		}
	}
}

func TestBlockScopesSearch(t *testing.T) {
	scopes := BlockScopes{
		{StartLine: 1, EndLine: 100}, // The whole file scope
		{StartLine: 10, EndLine: 20},
		{StartLine: 30, EndLine: 40},
		{StartLine: 50, EndLine: 60},
	}

	tests := []struct {
		name     string
		line     int
		expected int
	}{
		{"line in first scope", 15, 1},
		{"line in second scope", 35, 2},
		{"line in third scope", 55, 3},
		{"line outside scopes but in file", 25, 0},
		{"line outside scopes but in file", 45, 0},
		{"line in file scope", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if idx := scopes.Search(tt.line); idx != tt.expected {
				t.Errorf("Search(%d) = %v, want %v", tt.line, idx, tt.expected)
			}
		})
	}
}

func TestBlockScopesSearch2(t *testing.T) {
	// [{1 90} {12 15} {17 20} {22 25} {27 30} {32 89} {34 42} {45 53} {56 64} {77 83}]
	scopes := BlockScopes{
		{StartLine: 1, EndLine: 90},  // 0
		{StartLine: 12, EndLine: 15}, // 1
		{StartLine: 17, EndLine: 20}, // 2
		{StartLine: 22, EndLine: 25}, // 3
		{StartLine: 27, EndLine: 30}, // 4
		{StartLine: 32, EndLine: 89}, // 5
		{StartLine: 34, EndLine: 42}, // 6
		{StartLine: 45, EndLine: 53}, // 7
		{StartLine: 56, EndLine: 64}, // 8
		{StartLine: 77, EndLine: 83}, // 9
	}
	// line:  56
	// line:  66
	// line:  84
	// line:  86
	tests := []struct {
		name     string
		line     int
		expected int
	}{
		{"test line 56", 56, 5},
		{"test line 66", 66, 5},
		{"test line 84", 84, 5},
		{"test line 86", 86, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if idx := scopes.Search(tt.line); idx != tt.expected {
				t.Errorf("Search(%d) = %v, want %v", tt.line, idx, tt.expected)
			}
		})
	}
}

func TestBlockScopesUnique(t *testing.T) {
	// InsertPosition 结构是私有的，我们需要通过 InsertPositions 的方法来测试
	var positions InsertPositions
	positions.Insert(CodeInsertPositionFront, CodeInsertTypeComment, 10, 5)
	positions.Insert(CodeInsertPositionBack, CodeInsertTypeStmt, 20, 10)
	positions.Insert(CodeInsertPositionFront, CodeInsertTypeComment, 10, 5) // 重复项

	// 测试前长度
	if len(positions) != 3 {
		t.Errorf("Expected positions length to be 3, got %d", len(positions))
	}

	positions.Unique()

	// 测试后长度
	if len(positions) != 2 {
		t.Errorf("Expected unique positions length to be 2, got %d", len(positions))
	}
}

func TestBlockScopesReset(t *testing.T) {
	var positions InsertPositions
	positions.Insert(CodeInsertPositionFront, CodeInsertTypeComment, 10, 5)
	positions.Insert(CodeInsertPositionBack, CodeInsertTypeStmt, 20, 10)

	// 测试前长度
	if len(positions) != 2 {
		t.Errorf("Expected positions length to be 2, got %d", len(positions))
	}

	positions.Reset()

	// 测试后长度
	if len(positions) != 0 {
		t.Errorf("Expected reset positions length to be 0, got %d", len(positions))
	}
}

func TestBlockScopesOfGoAST(t *testing.T) {
	// 简单测试文件
	content := []byte(`
package example

func Example() {
	if true {
		// something
	}
	for i := 0; i < 10; i++ {
		// loop
	}
}
	`)

	scopes, err := BlockScopesOfGoAST("example.go", content)
	if err != nil {
		t.Fatalf("Failed to get block scopes: %v", err)
	}

	// 至少应该有这些区块：整个文件、函数、if语句、for语句
	if len(scopes) < 4 {
		t.Errorf("Expected at least 4 block scopes, got %d", len(scopes))
	}

	// 测试区块是否已排序
	for i := 1; i < len(scopes); i++ {
		if scopes[i-1].StartLine > scopes[i].StartLine {
			t.Errorf("Scopes not sorted at index %d", i)
		}
	}
}

func TestFunctionScopesOfGoAST(t *testing.T) {
	// 简单测试文件
	content := []byte(`
package example

func Example() {
	x := func() {
		// inner function
	}
	x()
}

var y = func() {
	// another function
}
	`)

	scopes, err := FunctionScopesOfGoAST("example.go", content)
	if err != nil {
		t.Fatalf("Failed to get function scopes: %v", err)
	}

	// 至少应该有这些区块：整个文件、Example函数、匿名函数x、匿名函数y
	if len(scopes) < 4 {
		t.Errorf("Expected at least 4 function scopes, got %d", len(scopes))
	}

	// 测试区块是否已排序
	for i := 1; i < len(scopes); i++ {
		if scopes[i-1].StartLine > scopes[i].StartLine {
			t.Errorf("Scopes not sorted at index %d", i)
		}
	}
}
