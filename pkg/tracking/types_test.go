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
	// InsertPosition is private, we need to test it through the methods of InsertPositions
	var positions InsertPositions
	positions.Insert(CodeInsertPositionFront, 10, 5)
	positions.Insert(CodeInsertPositionBack, 20, 10)
	positions.Insert(CodeInsertPositionFront, 10, 5) // Duplicate item

	// Test the length before unique
	if len(positions) != 3 {
		t.Errorf("Expected positions length to be 3, got %d", len(positions))
	}

	positions.Unique()

	// Test the length after unique
	if len(positions) != 2 {
		t.Errorf("Expected unique positions length to be 2, got %d", len(positions))
	}
}

func TestBlockScopesReset(t *testing.T) {
	var positions InsertPositions
	positions.Insert(CodeInsertPositionFront, 10, 5)
	positions.Insert(CodeInsertPositionBack, 20, 10)

	// Test the length before reset
	if len(positions) != 2 {
		t.Errorf("Expected positions length to be 2, got %d", len(positions))
	}

	positions.Reset()

	// Test the length after reset
	if len(positions) != 0 {
		t.Errorf("Expected reset positions length to be 0, got %d", len(positions))
	}
}

func TestFunctionScopesOfGoAST(t *testing.T) {
	// Simple test file
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

	scopes, err := FunctionScopesOfAST("example.go", content)
	if err != nil {
		t.Fatalf("Failed to get function scopes: %v", err)
	}

	// At least these scopes should be present: the whole file, the Example function, the anonymous function x, the anonymous function y
	if len(scopes) < 4 {
		t.Errorf("Expected at least 4 function scopes, got %d", len(scopes))
	}

	// Test if the scopes are sorted
	for i := 1; i < len(scopes); i++ {
		if scopes[i-1].StartLine > scopes[i].StartLine {
			t.Errorf("Scopes not sorted at index %d", i)
		}
	}
}

func TestTrackScopeIndex_CanInsertAndMarkInserted(t *testing.T) {
	// Create a trackScopeIndex instance with interval (5,20)
	tsi := &patchScope{
		scopeKey: scopeKey{startLine: 5, endLine: 20},
		marks:    make([]int, 14), // 20-5-1 = 14
	}

	// Simulate lineChanges (line numbers start from 1)
	lineChanges := make([]bool, 21)
	// Mark continuous block 1
	lineChanges[6] = true
	lineChanges[7] = true
	lineChanges[8] = true
	// Mark continuous block 2
	lineChanges[10] = true
	lineChanges[11] = true
	lineChanges[12] = true
	// Mark continuous block 3
	lineChanges[14] = true
	lineChanges[15] = true
	lineChanges[16] = true
	lineChanges[17] = true
	lineChanges[18] = true

	// Initialize marks
	tsi.initMarks(lineChanges)

	// Test if initialization is correct
	for i := 6; i <= 8; i++ {
		if !tsi.isNewLine(i) {
			t.Fatalf("Line %d should be NewLine", i)
		}
	}
	if tsi.isNewLine(9) {
		t.Fatalf("Line 9 should not be NewLine")
	}

	// Test canInsert - Rows in consecutive block 1 should be insertable.
	if !tsi.canInsert(7) {
		t.Fatalf("Line 7 should be insertable")
	}

	// Mark continuous block 1 as inserted
	tsi.markInserted(7)

	// Now all lines in continuous block 1 should be marked as inserted
	for i := 6; i <= 8; i++ {
		if !tsi.isInserted(i) {
			t.Fatalf("After marking, line %d should be inserted", i)
		}
		// Try inserting again should fail
		if tsi.canInsert(i) {
			t.Fatalf("After marking, line %d should not be insertable again", i)
		}
	}

	// The continuous block 2 should still be insertable.
	if !tsi.canInsert(11) {
		t.Fatalf("Line 11 should be insertable")
	}

	// Mark continuous block 2 as inserted
	tsi.markInserted(11)

	// Continuous block 2 should be marked as inserted
	for i := 10; i <= 12; i++ {
		if !tsi.isInserted(i) {
			t.Fatalf("After marking, line %d should be inserted", i)
		}
		if tsi.canInsert(i) {
			t.Fatalf("After marking, line %d should not be insertable again", i)
		}
	}

	// Continuous block 3 should still be insertable
	if !tsi.canInsert(16) {
		t.Fatalf("Line 16 should be insertable")
	}

	// Mark continuous block 3 as inserted
	tsi.markInserted(16)

	// Continuous block 3 should be marked as inserted
	for i := 14; i <= 18; i++ {
		if !tsi.isInserted(i) {
			t.Fatalf("After marking, line %d should be inserted", i)
		}
		if tsi.canInsert(i) {
			t.Fatalf("After marking, line %d should not be insertable again", i)
		}
	}

	// Test if marking a previous continuous block does not affect the insertability of the next continuous block
	tsi = &patchScope{
		scopeKey: scopeKey{startLine: 5, endLine: 20},
		marks:    make([]int, 14), // 20-5-1 = 14
	}
	tsi.initMarks(lineChanges)

	// Mark only continuous block 1
	tsi.markInserted(7)

	// Continuous block 2 should still be insertable
	if !tsi.canInsert(11) {
		t.Fatalf("After marking block 1, block 2 (line 11) should still be insertable")
	}

	// Edge case: all lines are NewLine
	allNewLines := &patchScope{
		scopeKey: scopeKey{startLine: 30, endLine: 35},
		marks:    make([]int, 4), // 35-30-1 = 4
	}
	allChanges := make([]bool, 36)
	allChanges[31] = true
	allChanges[32] = true
	allChanges[33] = true
	allChanges[34] = true
	allNewLines.initMarks(allChanges)

	// When all lines are NewLine, the first line should be insertable
	if !allNewLines.canInsert(31) {
		t.Fatalf("When all lines are NewLine, the first line should be insertable")
	}

	// After marking, other lines should not be insertable
	allNewLines.markInserted(31)
	for i := 31; i <= 34; i++ {
		if !allNewLines.isInserted(i) {
			t.Fatalf("When all lines are NewLine, line %d should be inserted", i)
		}
		if i > 31 && allNewLines.canInsert(i) {
			t.Fatalf("When all lines are NewLine, line %d should not be insertable again", i)
		}
	}
}
