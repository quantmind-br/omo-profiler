package diff

import (
	"strings"
	"testing"
)

// TestDiffTypeConstants verifies the DiffType constant values
func TestDiffTypeConstants(t *testing.T) {
	if DiffEqual != 0 {
		t.Errorf("Expected DiffEqual to be 0, got %d", DiffEqual)
	}
	if DiffAdded != 1 {
		t.Errorf("Expected DiffAdded to be 1, got %d", DiffAdded)
	}
	if DiffRemoved != 2 {
		t.Errorf("Expected DiffRemoved to be 2, got %d", DiffRemoved)
	}
}

// TestSplitLines tests the splitLines helper function
func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line without newline",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			input:    "hello\n",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multiple lines with trailing newline",
			input:    "line1\nline2\nline3\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{"", "", ""},
		},
		{
			name:     "single newline",
			input:    "\n",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

// TestComputeDiffIdentical tests ComputeDiff with identical content
func TestComputeDiffIdentical(t *testing.T) {
	json1 := []byte(`{
  "name": "test",
  "value": 123
}`)
	json2 := []byte(`{
  "name": "test",
  "value": 123
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if len(result.Left) != len(result.Right) {
		t.Errorf("Left and Right should have same length for identical content")
	}

	// All lines should be DiffEqual
	for i, line := range result.Left {
		if line.Type != DiffEqual {
			t.Errorf("Line %d on left should be DiffEqual, got %d", i, line.Type)
		}
		if line.LineNum != i+1 {
			t.Errorf("Line %d on left should have LineNum %d, got %d", i, i+1, line.LineNum)
		}
	}

	for i, line := range result.Right {
		if line.Type != DiffEqual {
			t.Errorf("Line %d on right should be DiffEqual, got %d", i, line.Type)
		}
		if line.LineNum != i+1 {
			t.Errorf("Line %d on right should have LineNum %d, got %d", i, i+1, line.LineNum)
		}
	}
}

// TestComputeDiffAddedLines tests ComputeDiff with added lines
func TestComputeDiffAddedLines(t *testing.T) {
	json1 := []byte(`{
  "name": "test"
}`)
	json2 := []byte(`{
  "name": "test",
  "value": 123
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Should have at least one DiffAdded line
	foundAdded := false
	for _, line := range result.Right {
		if line.Type == DiffAdded {
			foundAdded = true
			if line.LineNum == 0 {
				t.Error("Added line on right should not have LineNum 0")
			}
		}
	}

	if !foundAdded {
		t.Error("Expected to find at least one DiffAdded line on right side")
	}

	// Check that added lines have placeholder on left
	for i, line := range result.Left {
		if line.Type == DiffAdded {
			if line.Text != "" {
				t.Errorf("Added line %d on left should have empty text, got %q", i, line.Text)
			}
			if line.LineNum != 0 {
				t.Errorf("Added line %d on left should have LineNum 0, got %d", i, line.LineNum)
			}
		}
	}
}

// TestComputeDiffRemovedLines tests ComputeDiff with removed lines
func TestComputeDiffRemovedLines(t *testing.T) {
	json1 := []byte(`{
  "name": "test",
  "value": 123
}`)
	json2 := []byte(`{
  "name": "test"
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Should have at least one DiffRemoved line
	foundRemoved := false
	for _, line := range result.Left {
		if line.Type == DiffRemoved {
			foundRemoved = true
			if line.LineNum == 0 {
				t.Error("Removed line on left should not have LineNum 0")
			}
		}
	}

	if !foundRemoved {
		t.Error("Expected to find at least one DiffRemoved line on left side")
	}

	// Check that removed lines have placeholder on right
	for i, line := range result.Right {
		if line.Type == DiffRemoved {
			if line.Text != "" {
				t.Errorf("Removed line %d on right should have empty text, got %q", i, line.Text)
			}
			if line.LineNum != 0 {
				t.Errorf("Removed line %d on right should have LineNum 0, got %d", i, line.LineNum)
			}
		}
	}
}

// TestComputeDiffModifiedLines tests ComputeDiff with modified lines
func TestComputeDiffModifiedLines(t *testing.T) {
	json1 := []byte(`{
  "name": "old-value"
}`)
	json2 := []byte(`{
  "name": "new-value"
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Modified lines appear as removed + added
	foundRemoved := false
	foundAdded := false

	for _, line := range result.Left {
		if line.Type == DiffRemoved {
			foundRemoved = true
		}
	}

	for _, line := range result.Right {
		if line.Type == DiffAdded {
			foundAdded = true
		}
	}

	if !foundRemoved || !foundAdded {
		t.Error("Expected to find both removed and added lines for modified content")
	}
}

// TestComputeDiffBothEmpty tests ComputeDiff with both inputs empty
func TestComputeDiffBothEmpty(t *testing.T) {
	json1 := []byte("")
	json2 := []byte("")

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if len(result.Left) != 0 {
		t.Errorf("Expected empty left result, got %d lines", len(result.Left))
	}

	if len(result.Right) != 0 {
		t.Errorf("Expected empty right result, got %d lines", len(result.Right))
	}
}

// TestComputeDiffLeftEmpty tests ComputeDiff with left input empty
func TestComputeDiffLeftEmpty(t *testing.T) {
	json1 := []byte("")
	json2 := []byte(`{
  "name": "test"
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// All lines should be added
	for i, line := range result.Right {
		if line.Type != DiffAdded {
			t.Errorf("Line %d on right should be DiffAdded, got %d", i, line.Type)
		}
	}

	// Left should have placeholders
	for i, line := range result.Left {
		if line.Type != DiffAdded {
			t.Errorf("Line %d on left should be DiffAdded (placeholder), got %d", i, line.Type)
		}
		if line.Text != "" {
			t.Errorf("Line %d on left should have empty text, got %q", i, line.Text)
		}
		if line.LineNum != 0 {
			t.Errorf("Line %d on left should have LineNum 0, got %d", i, line.LineNum)
		}
	}
}

// TestComputeDiffRightEmpty tests ComputeDiff with right input empty
func TestComputeDiffRightEmpty(t *testing.T) {
	json1 := []byte(`{
  "name": "test"
}`)
	json2 := []byte("")

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// All lines should be removed
	for i, line := range result.Left {
		if line.Type != DiffRemoved {
			t.Errorf("Line %d on left should be DiffRemoved, got %d", i, line.Type)
		}
	}

	// Right should have placeholders
	for i, line := range result.Right {
		if line.Type != DiffRemoved {
			t.Errorf("Line %d on right should be DiffRemoved (placeholder), got %d", i, line.Type)
		}
		if line.Text != "" {
			t.Errorf("Line %d on right should have empty text, got %q", i, line.Text)
		}
		if line.LineNum != 0 {
			t.Errorf("Line %d on right should have LineNum 0, got %d", i, line.LineNum)
		}
	}
}

// TestComputeDiffLineNumberSequencing tests line number sequencing
func TestComputeDiffLineNumberSequencing(t *testing.T) {
	json1 := []byte(`line1
line2
line3`)
	json2 := []byte(`line1
line2
line3`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// For equal lines, line numbers should be sequential on both sides
	for i, line := range result.Left {
		if line.Type == DiffEqual {
			expectedLineNum := i + 1
			if line.LineNum != expectedLineNum {
				t.Errorf("Left line %d should have LineNum %d, got %d", i, expectedLineNum, line.LineNum)
			}
		}
	}

	for i, line := range result.Right {
		if line.Type == DiffEqual {
			expectedLineNum := i + 1
			if line.LineNum != expectedLineNum {
				t.Errorf("Right line %d should have LineNum %d, got %d", i, expectedLineNum, line.LineNum)
			}
		}
	}
}

// TestComputeDiffMultilineJSON tests ComputeDiff with multiline JSON structures
func TestComputeDiffMultilineJSON(t *testing.T) {
	json1 := []byte(`{
  "users": [
    {
      "name": "Alice",
      "age": 30
    },
    {
      "name": "Bob",
      "age": 25
    }
  ]
}`)
	json2 := []byte(`{
  "users": [
    {
      "name": "Alice",
      "age": 31
    },
    {
      "name": "Bob",
      "age": 25
    }
  ]
}`)

	result, err := ComputeDiff(json1, json2)
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	// Should have some equal lines and some different lines
	hasEqual := false
	hasDiff := false

	for _, line := range result.Left {
		if line.Type == DiffEqual {
			hasEqual = true
		}
		if line.Type == DiffRemoved || line.Type == DiffAdded {
			hasDiff = true
		}
	}

	if !hasEqual {
		t.Error("Expected to find some equal lines")
	}

	if !hasDiff {
		t.Error("Expected to find some different lines")
	}
}

// TestComputeUnifiedDiff_IdenticalContent tests ComputeUnifiedDiff with identical content
func TestComputeUnifiedDiff_IdenticalContent(t *testing.T) {
	oldContent := []byte("line1\nline2\nline3")
	newContent := []byte("line1\nline2\nline3")
	oldName := "old.json"
	newName := "new.json"

	diff := ComputeUnifiedDiff(oldName, newName, oldContent, newContent)

	expectedHeader := "--- old.json\n+++ new.json\n"
	if diff != expectedHeader {
		t.Errorf("Expected only headers for identical content, got:\n%q", diff)
	}
}

// TestComputeUnifiedDiff_WithDifferences tests ComputeUnifiedDiff with differences
func TestComputeUnifiedDiff_WithDifferences(t *testing.T) {
	oldContent := []byte("line1\nline2\nline3")
	newContent := []byte("line1\nline2.modified\nline3")
	oldName := "old.json"
	newName := "new.json"

	diff := ComputeUnifiedDiff(oldName, newName, oldContent, newContent)

	if !strings.Contains(diff, "-line2") {
		t.Errorf("Expected diff to contain -line2, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+line2.modified") {
		t.Errorf("Expected diff to contain +line2.modified, got:\n%s", diff)
	}
}

// TestComputeUnifiedDiff_HeadersPresent tests ComputeUnifiedDiff headers
func TestComputeUnifiedDiff_HeadersPresent(t *testing.T) {
	oldContent := []byte("a")
	newContent := []byte("b")
	oldName := "file1"
	newName := "file2"

	diff := ComputeUnifiedDiff(oldName, newName, oldContent, newContent)

	if !strings.HasPrefix(diff, "--- file1\n+++ file2\n") {
		t.Errorf("Expected unified diff headers, got:\n%s", diff)
	}
}
