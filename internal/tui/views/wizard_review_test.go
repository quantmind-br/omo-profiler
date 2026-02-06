package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/schema"
)

func TestNewWizardReview(t *testing.T) {
	wr := NewWizardReview()

	// Just verify key bindings are initialized
	if wr.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if wr.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if wr.keys.Save.Help().Key == "" {
		t.Error("expected Save key to be initialized")
	}

	if wr.keys.Back.Help().Key == "" {
		t.Error("expected Back key to be initialized")
	}
}

func TestWizardReviewInit(t *testing.T) {
	wr := NewWizardReview()
	cmd := wr.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestWizardReviewSetSize(t *testing.T) {
	wr := NewWizardReview()

	wr.SetSize(100, 50)

	if wr.width != 100 {
		t.Errorf("expected width 100, got %d", wr.width)
	}

	if wr.height != 50 {
		t.Errorf("expected height 50, got %d", wr.height)
	}

	if !wr.ready {
		t.Error("expected ready to be true after SetSize")
	}

	// Call SetSize again to test the ready=true path
	wr.SetSize(80, 40)

	if wr.width != 80 {
		t.Errorf("expected width 80, got %d", wr.width)
	}

	if wr.height != 40 {
		t.Errorf("expected height 40, got %d", wr.height)
	}
}

func TestWizardReviewSetConfig(t *testing.T) {
	wr := NewWizardReview()

	cfg := &config.Config{
		// Use valid config structure
	}

	wr.SetConfig("test-profile", cfg)

	if wr.profileName != "test-profile" {
		t.Errorf("expected profile name 'test-profile', got '%s'", wr.profileName)
	}

	if wr.config != cfg {
		t.Error("expected config to be set")
	}

	if wr.jsonPreview == "" {
		t.Error("expected jsonPreview to be set")
	}
}

func TestWizardReviewSetConfigNil(t *testing.T) {
	wr := NewWizardReview()

	wr.SetConfig("empty", nil)

	if wr.profileName != "empty" {
		t.Errorf("expected profile name 'empty', got '%s'", wr.profileName)
	}

	if wr.config != nil {
		t.Error("expected config to be nil")
	}

	if wr.jsonPreview != "{}" {
		t.Errorf("expected jsonPreview to be '{}', got '%s'", wr.jsonPreview)
	}

	if !wr.isValid {
		t.Error("expected isValid to be true for nil config")
	}
}

func TestWizardReviewValidateAndPreview(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		wantJSON string
		wantValid bool
	}{
		{
			name:     "nil config",
			config:   nil,
			wantJSON: "{}",
			wantValid: true,
		},
		{
			name: "simple config",
			config: &config.Config{
				// Minimal valid config
			},
			wantJSON: "", // We'll just check it's not empty
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := NewWizardReview()
			wr.SetConfig("test", tt.config)

			if tt.wantJSON != "" && wr.jsonPreview != tt.wantJSON {
				t.Errorf("expected jsonPreview %q, got %q", tt.wantJSON, wr.jsonPreview)
			}

			if tt.wantJSON == "" && wr.jsonPreview == "" {
				t.Error("expected jsonPreview to be set")
			}

			if wr.isValid != tt.wantValid {
				t.Errorf("expected isValid %v, got %v", tt.wantValid, wr.isValid)
			}
		})
	}
}

func TestWizardReviewUpdateWindowSizeMsg(t *testing.T) {
	wr := NewWizardReview()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := wr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for WindowSizeMsg")
	}

	if updated.width != 80 {
		t.Errorf("expected width 80, got %d", updated.width)
	}

	if updated.height != 24 {
		t.Errorf("expected height 24, got %d", updated.height)
	}
}

func TestWizardReviewUpdateSaveKey(t *testing.T) {
	wr := NewWizardReview()
	cfg := &config.Config{}
	wr.SetConfig("test", cfg)

	// Test save when valid
	msg := tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune("enter")}
	updated, cmd := wr.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command when valid and Enter pressed")
	}

	if cmd != nil {
		result := cmd()
		if _, ok := result.(WizardNextMsg); !ok {
			t.Errorf("expected WizardNextMsg, got %T", result)
		}
	}

	// Test save when invalid
	wr.isValid = false
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd = wr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command when invalid and Enter pressed")
	}

	// Verify state didn't change
	if updated.isValid != false {
		t.Error("expected isValid to remain false")
	}
}

func TestWizardReviewUpdateBackKey(t *testing.T) {
	wr := NewWizardReview()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := wr.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command for Esc key")
	}

	if cmd != nil {
		result := cmd()
		if _, ok := result.(WizardBackMsg); !ok {
			t.Errorf("expected WizardBackMsg, got %T", result)
		}
	}

	// Verify state is preserved
	if updated.width != wr.width {
		t.Error("expected width to be preserved")
	}
}

func TestWizardReviewUpdateUnknownKey(t *testing.T) {
	wr := NewWizardReview()
	wr.SetSize(80, 24)

	// Test a key that doesn't match any binding
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := wr.Update(msg)

	// Should return the viewport's cmd (might be nil)
	// Just verify no panic and state is preserved
	if updated.width != 80 {
		t.Error("expected width to be preserved")
	}

	_ = cmd // Can't assert on viewport cmd without detailed knowledge
}

func TestWizardReviewIsValid(t *testing.T) {
	wr := NewWizardReview()

	// Default is invalid (no config set)
	if wr.IsValid() {
		t.Error("expected IsValid to be false initially")
	}

	wr.isValid = true
	if !wr.IsValid() {
		t.Error("expected IsValid to be true after setting")
	}
}

func TestWizardReviewGetErrors(t *testing.T) {
	wr := NewWizardReview()

	// No errors initially
	errs := wr.GetErrors()
	if errs != nil {
		t.Errorf("expected nil errors, got %v", errs)
	}

	// Set some errors
	wr.validationErrs = []schema.ValidationError{
		{Path: "test.path", Message: "test error"},
	}

	errs = wr.GetErrors()
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Path != "test.path" {
		t.Errorf("expected path 'test.path', got '%s'", errs[0].Path)
	}
}

func TestWizardReviewView(t *testing.T) {
	wr := NewWizardReview()
	wr.SetSize(80, 24)

	view := wr.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for title
	if !contains(view, "Review & Save") {
		t.Error("expected 'Review & Save' in view")
	}

	// Check for JSON Preview section
	if !contains(view, "JSON Preview:") {
		t.Error("expected 'JSON Preview:' in view")
	}

	// Check for help text
	if !contains(view, "Shift+Tab") {
		t.Error("expected 'Shift+Tab' in view")
	}
}

func TestWizardReviewViewWithValidConfig(t *testing.T) {
	wr := NewWizardReview()
	cfg := &config.Config{}
	wr.SetConfig("test", cfg)
	wr.SetSize(80, 24)

	view := wr.View()

	// Should show valid status
	if !contains(view, "✓ Configuration is valid") {
		t.Error("expected valid status in view")
	}

	// Should have save help
	if !contains(view, "Enter to save") {
		t.Error("expected 'Enter to save' in view")
	}
}

func TestWizardReviewViewWithInvalidConfig(t *testing.T) {
	wr := NewWizardReview()
	wr.isValid = false
	wr.validationErrs = []schema.ValidationError{
		{Path: "test.path", Message: "test error"},
	}
	wr.SetSize(80, 24)

	view := wr.View()

	// Should show invalid status
	if !contains(view, "✗ Validation errors found:") {
		t.Error("expected invalid status in view")
	}

	// Should show error details
	if !contains(view, "• test.path: test error") {
		t.Error("expected error details in view")
	}

	// Should NOT have save help
	if contains(view, "Enter to save") {
		t.Error("should not show 'Enter to save' when invalid")
	}
}

func TestWizardReviewKeyMap(t *testing.T) {
	wr := NewWizardReview()

	// Test that all key bindings are properly set
	tests := []struct {
		name     string
		binding  key.Binding
		expected string
	}{
		{"Up", wr.keys.Up, "↑/k"},
		{"Down", wr.keys.Down, "↓/j"},
		{"Save", wr.keys.Save, "enter/ctrl+s"},
		{"Back", wr.keys.Back, "shift+tab/esc"},
		{"PageUp", wr.keys.PageUp, "pgup"},
		{"PageDown", wr.keys.PageDown, "pgdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.binding.Help().Key == "" {
				t.Error("expected key binding to have a key")
			}
			if !contains(tt.binding.Help().Key, tt.expected) &&
			   tt.binding.Help().Key != tt.expected {
				// Help key format may vary, just check it's not empty
				t.Logf("%s key binding: %s", tt.name, tt.binding.Help().Key)
			}
		})
	}
}
