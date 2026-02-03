package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewExport(t *testing.T) {
	exp := NewExport("my-profile")

	if exp.profileName != "my-profile" {
		t.Errorf("Expected profile name 'my-profile', got '%s'", exp.profileName)
	}

	if exp.textInput.Placeholder != "Export path..." {
		t.Errorf("Expected placeholder 'Export path...', got '%s'", exp.textInput.Placeholder)
	}

	if !exp.textInput.Focused() {
		t.Error("Expected text input to be focused")
	}
}

func TestExportSetSize(t *testing.T) {
	exp := NewExport("test")
	exp.SetSize(100, 50)

	if exp.width != 100 {
		t.Errorf("Expected width 100, got %d", exp.width)
	}

	if exp.height != 50 {
		t.Errorf("Expected height 50, got %d", exp.height)
	}
}

func TestExportUpdate(t *testing.T) {
	exp := NewExport("test-profile")

	exp, _ = exp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	exp, _ = exp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	exp, _ = exp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	exp, _ = exp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})

	if exp.GetPath() != "/tmp" {
		t.Errorf("Expected path '/tmp', got '%s'", exp.GetPath())
	}
}

func TestExportCancel(t *testing.T) {
	exp := NewExport("test")

	_, cmd := exp.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command from Esc key")
	}

	msg := cmd()
	if _, ok := msg.(ExportCancelMsg); !ok {
		t.Errorf("Expected ExportCancelMsg, got %T", msg)
	}
}

func TestExportDoneWithEmptyPath(t *testing.T) {
	exp := NewExport("test")

	exp, cmd := exp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("Expected nil command for empty path")
	}

	if exp.err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestExportDoneWithValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	exp := NewExport("my-profile")

	exportPath := filepath.Join(tmpDir, "export.json")
	for _, r := range exportPath {
		exp, _ = exp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	_, cmd := exp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command from Enter key")
	}

	msg := cmd()
	doneMsg, ok := msg.(ExportDoneMsg)
	if !ok {
		t.Fatalf("Expected ExportDoneMsg, got %T", msg)
	}

	if doneMsg.Err != nil {
		t.Errorf("Expected no error, got %v", doneMsg.Err)
	}

	if doneMsg.Path == "" {
		t.Error("Expected non-empty path in ExportDoneMsg")
	}
}

func TestAutoRenameIfExists(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("file does not exist", func(t *testing.T) {
		path := filepath.Join(tmpDir, "new-file.json")
		result := autoRenameIfExists(path)

		if result != path {
			t.Errorf("Expected '%s', got '%s'", path, result)
		}
	})

	t.Run("file exists", func(t *testing.T) {
		existingFile := filepath.Join(tmpDir, "existing.json")
		if err := os.WriteFile(existingFile, []byte(`{}`), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result := autoRenameIfExists(existingFile)
		expected := filepath.Join(tmpDir, "existing-1.json")

		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("multiple files exist", func(t *testing.T) {
		baseFile := filepath.Join(tmpDir, "multi.json")
		if err := os.WriteFile(baseFile, []byte(`{}`), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		existing1 := filepath.Join(tmpDir, "multi-1.json")
		if err := os.WriteFile(existing1, []byte(`{}`), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result := autoRenameIfExists(baseFile)
		expected := filepath.Join(tmpDir, "multi-2.json")

		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestExpandExportPath(t *testing.T) {
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
			input:    "~/export.json",
			expected: filepath.Join(home, "export.json"),
		},
		{
			name:     "absolute path",
			input:    "/home/user/export.json",
			expected: "/home/user/export.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandExportPath(tt.input)
			if err != nil {
				t.Fatalf("expandExportPath failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExpandExportPathRelative(t *testing.T) {
	result, err := expandExportPath("./export.json")
	if err != nil {
		t.Fatalf("expandExportPath failed: %v", err)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("Expected absolute path, got '%s'", result)
	}
}

func TestExportView(t *testing.T) {
	exp := NewExport("test-profile")
	view := exp.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "Export Profile") {
		t.Error("Expected 'Export Profile' in view")
	}

	if !contains(view, "test-profile") {
		t.Error("Expected profile name in view")
	}
}

func TestExportGetProfileName(t *testing.T) {
	exp := NewExport("my-awesome-profile")

	if exp.GetProfileName() != "my-awesome-profile" {
		t.Errorf("Expected 'my-awesome-profile', got '%s'", exp.GetProfileName())
	}
}
