package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewImport(t *testing.T) {
	imp := NewImport()

	if imp.textInput.Placeholder != "Path to JSON file..." {
		t.Errorf("Expected placeholder 'Path to JSON file...', got '%s'", imp.textInput.Placeholder)
	}

	if !imp.textInput.Focused() {
		t.Error("Expected text input to be focused")
	}
}

func TestImportSetSize(t *testing.T) {
	imp := NewImport()
	imp.SetSize(100, 50)

	if imp.width != 100 {
		t.Errorf("Expected width 100, got %d", imp.width)
	}

	if imp.height != 50 {
		t.Errorf("Expected height 50, got %d", imp.height)
	}
}

func TestImportUpdate(t *testing.T) {
	imp := NewImport()

	// Test typing
	imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	if imp.GetPath() != "test" {
		t.Errorf("Expected path 'test', got '%s'", imp.GetPath())
	}
}

func TestImportCancel(t *testing.T) {
	imp := NewImport()

	_, cmd := imp.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command from Esc key")
	}

	msg := cmd()
	if _, ok := msg.(ImportCancelMsg); !ok {
		t.Errorf("Expected ImportCancelMsg, got %T", msg)
	}
}

func TestImportDoneWithEmptyPath(t *testing.T) {
	imp := NewImport()

	imp, cmd := imp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Expected nil command for empty path")
	}

	if imp.err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestImportDoneWithNonExistentFile(t *testing.T) {
	imp := NewImport()

	// Type a non-existent path
	for _, r := range "/nonexistent/file.json" {
		imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	imp, cmd := imp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Expected nil command for non-existent file")
	}

	if imp.err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestImportDoneWithValidFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(tmpFile, []byte(`{}`), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	imp := NewImport()

	// Type the path
	for _, r := range tmpFile {
		imp, _ = imp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	_, cmd := imp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command from Enter key")
	}

	msg := cmd()
	if _, ok := msg.(ImportDoneMsg); !ok {
		t.Errorf("Expected ImportDoneMsg, got %T", msg)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/test.json",
			expected: filepath.Join(home, "test.json"),
		},
		{
			name:     "absolute path",
			input:    "/home/user/test.json",
			expected: "/home/user/test.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.input)
			if err != nil {
				t.Fatalf("expandPath failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExpandPathRelative(t *testing.T) {
	result, err := expandPath("./test.json")
	if err != nil {
		t.Fatalf("expandPath failed: %v", err)
	}

	// Result should be absolute
	if !filepath.IsAbs(result) {
		t.Errorf("Expected absolute path, got '%s'", result)
	}
}

func TestImportView(t *testing.T) {
	imp := NewImport()
	view := imp.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Check that title is present
	if !contains(view, "Import Profile") {
		t.Error("Expected 'Import Profile' in view")
	}

	// Check that help is present
	if !contains(view, "import") {
		t.Error("Expected 'import' in help text")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
