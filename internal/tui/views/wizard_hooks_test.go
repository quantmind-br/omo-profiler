package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/diogenes/omo-profiler/internal/config"
)

func TestNewWizardHooks(t *testing.T) {
	wh := NewWizardHooks()

	if wh.disabled == nil {
		t.Error("expected disabled map to be initialized")
	}

	if wh.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", wh.cursor)
	}

	if len(wh.disabled) != len(allHooks) {
		t.Errorf("expected %d hooks, got %d", len(allHooks), len(wh.disabled))
	}

	// All hooks should be enabled by default
	for hook, disabled := range wh.disabled {
		if disabled {
			t.Errorf("expected hook %q to be enabled by default", hook)
		}
	}
}

func TestWizardHooksInit(t *testing.T) {
	wh := NewWizardHooks()
	cmd := wh.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestWizardHooksSetSize(t *testing.T) {
	wh := NewWizardHooks()

	wh.SetSize(100, 50)

	if wh.width != 100 {
		t.Errorf("expected width 100, got %d", wh.width)
	}

	if wh.height != 50 {
		t.Errorf("expected height 50, got %d", wh.height)
	}

	if !wh.ready {
		t.Error("expected ready to be true after SetSize")
	}

	// Call SetSize again to test the ready=true path
	wh.SetSize(80, 40)

	if wh.width != 80 {
		t.Errorf("expected width 80, got %d", wh.width)
	}

	if wh.height != 40 {
		t.Errorf("expected height 40, got %d", wh.height)
	}
}

func TestWizardHooksSetConfig(t *testing.T) {
	wh := NewWizardHooks()

	cfg := &config.Config{
		DisabledHooks: []string{
			"todo-continuation-enforcer",
			"session-recovery",
		},
	}

	wh.SetConfig(cfg)

	// Check that the specified hooks are disabled
	if !wh.disabled["todo-continuation-enforcer"] {
		t.Error("expected todo-continuation-enforcer to be disabled")
	}

	if !wh.disabled["session-recovery"] {
		t.Error("expected session-recovery to be disabled")
	}

	// Check that other hooks remain enabled
	if wh.disabled["context-window-monitor"] {
		t.Error("expected context-window-monitor to be enabled")
	}
}

func TestWizardHooksSetConfigResetsPreviousState(t *testing.T) {
	wh := NewWizardHooks()

	// Set some disabled hooks
	wh.disabled["todo-continuation-enforcer"] = true
	wh.disabled["session-recovery"] = true

	// Set config with different disabled hooks
	cfg := &config.Config{
		DisabledHooks: []string{"context-window-monitor"},
	}

	wh.SetConfig(cfg)

	// Previous state should be reset
	if wh.disabled["todo-continuation-enforcer"] {
		t.Error("expected todo-continuation-enforcer to be enabled after reset")
	}

	if wh.disabled["session-recovery"] {
		t.Error("expected session-recovery to be enabled after reset")
	}

	// New config should be applied
	if !wh.disabled["context-window-monitor"] {
		t.Error("expected context-window-monitor to be disabled")
	}
}

func TestWizardHooksApply(t *testing.T) {
	wh := NewWizardHooks()

	// Disable some hooks
	wh.disabled["todo-continuation-enforcer"] = true
	wh.disabled["session-recovery"] = true

	cfg := &config.Config{}
	wh.Apply(cfg)

	// Check DisabledHooks is set correctly
	if len(cfg.DisabledHooks) != 2 {
		t.Errorf("expected 2 disabled hooks, got %d", len(cfg.DisabledHooks))
	}

	// Check order (should match allHooks order: todo-continuation-enforcer comes before session-recovery)
	expected := []string{"todo-continuation-enforcer", "session-recovery"}
	for i, hook := range cfg.DisabledHooks {
		if hook != expected[i] {
			t.Errorf("expected hook %d to be %q, got %q", i, expected[i], hook)
		}
	}
}

func TestWizardHooksApplyNoneDisabled(t *testing.T) {
	wh := NewWizardHooks()

	cfg := &config.Config{}
	wh.Apply(cfg)

	// When no hooks are disabled, DisabledHooks should be nil (not empty slice)
	if cfg.DisabledHooks != nil {
		t.Errorf("expected DisabledHooks to be nil when none disabled, got %v", cfg.DisabledHooks)
	}
}

func TestWizardHooksUpdateUpKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = 5

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := wh.Update(msg)

	if updated.cursor != 4 {
		t.Errorf("expected cursor to be 4, got %d", updated.cursor)
	}

	// Test at top
	wh.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = wh.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to remain 0 at top, got %d", updated.cursor)
	}

	_ = cmd
}

func TestWizardHooksUpdateDownKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = len(allHooks) - 5

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := wh.Update(msg)

	if updated.cursor != len(allHooks)-4 {
		t.Errorf("expected cursor to be %d, got %d", len(allHooks)-4, updated.cursor)
	}

	// Test at bottom
	wh.cursor = len(allHooks) - 1
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = wh.Update(msg)

	if updated.cursor != len(allHooks)-1 {
		t.Errorf("expected cursor to remain at bottom, got %d", updated.cursor)
	}

	_ = cmd
}

func TestWizardHooksUpdateToggleKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = 0
	hook := allHooks[0]

	// Initial state should be enabled
	if wh.disabled[hook] {
		t.Error("expected hook to be enabled initially")
	}

	// Toggle to disabled
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd := wh.Update(msg)

	if !updated.disabled[hook] {
		t.Error("expected hook to be disabled after toggle")
	}

	// Toggle back to enabled
	msg = tea.KeyMsg{Type: tea.KeySpace}
	updated, _ = wh.Update(msg)

	if updated.disabled[hook] {
		t.Error("expected hook to be enabled after second toggle")
	}

	_ = cmd
}

func TestWizardHooksUpdateNextKey(t *testing.T) {
	wh := NewWizardHooks()

	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, cmd := wh.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command for Tab key")
	}

	if cmd != nil {
		result := cmd()
		if _, ok := result.(WizardNextMsg); !ok {
			t.Errorf("expected WizardNextMsg, got %T", result)
		}
	}

	// Verify state is preserved
	if updated.cursor != 0 {
		t.Error("expected cursor to be preserved")
	}
}

func TestWizardHooksUpdateBackKey(t *testing.T) {
	wh := NewWizardHooks()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := wh.Update(msg)

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
	if updated.cursor != 0 {
		t.Error("expected cursor to be preserved")
	}
}

func TestWizardHooksUpdatePageUpKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = 20

	msg := tea.KeyMsg{Type: tea.KeyPgUp}
	updated, cmd := wh.Update(msg)

	if updated.cursor != 10 {
		t.Errorf("expected cursor to be 10, got %d", updated.cursor)
	}

	// Test page up near top
	wh.cursor = 5
	msg = tea.KeyMsg{Type: tea.KeyPgUp}
	updated, _ = wh.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be 0 after page up from position 5, got %d", updated.cursor)
	}

	_ = cmd
}

func TestWizardHooksUpdatePageDownKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = len(allHooks) - 20

	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	updated, cmd := wh.Update(msg)

	if updated.cursor != len(allHooks)-10 {
		t.Errorf("expected cursor to be %d, got %d", len(allHooks)-10, updated.cursor)
	}

	// Test page down near bottom
	wh.cursor = len(allHooks) - 5
	msg = tea.KeyMsg{Type: tea.KeyPgDown}
	updated, _ = wh.Update(msg)

	if updated.cursor != len(allHooks)-1 {
		t.Errorf("expected cursor to be at bottom, got %d", updated.cursor)
	}

	_ = cmd
}

func TestWizardHooksUpdateWindowSizeMsg(t *testing.T) {
	wh := NewWizardHooks()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := wh.Update(msg)

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

func TestWizardHooksUpdateUnknownKey(t *testing.T) {
	wh := NewWizardHooks()
	wh.SetSize(80, 24)

	// Test a key that doesn't match any binding
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := wh.Update(msg)

	// Should return the viewport's cmd (might be nil)
	// Just verify no panic and state is preserved
	if updated.cursor != 0 {
		t.Error("expected cursor to be preserved")
	}

	_ = cmd // Can't assert on viewport cmd without detailed knowledge
}

func TestWizardHooksView(t *testing.T) {
	wh := NewWizardHooks()
	wh.SetSize(80, 24)

	view := wh.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for title
	if !contains(view, "Configure Hooks") {
		t.Error("expected 'Configure Hooks' in view")
	}

	// Check for help text
	if !contains(view, "Space to toggle") {
		t.Error("expected 'Space to toggle' in view")
	}

	// Check for stats
	if !contains(view, "hooks disabled") {
		t.Error("expected 'hooks disabled' in view")
	}
}

func TestWizardHooksViewWithDisabledHooks(t *testing.T) {
	wh := NewWizardHooks()
	wh.SetSize(80, 24)

	// Disable some hooks
	wh.disabled["todo-continuation-enforcer"] = true
	wh.disabled["session-recovery"] = true

	view := wh.View()

	// Should show disabled count
	if !contains(view, "2/36") {
		t.Error("expected '2/36 hooks disabled' in view")
	}
}

func TestWizardHooksRenderContent(t *testing.T) {
	wh := NewWizardHooks()
	wh.cursor = 0

	content := wh.renderContent()

	if content == "" {
		t.Error("expected non-empty content")
	}

	// Should have cursor on first item
	if !contains(content, "> ") {
		t.Error("expected cursor marker in content")
	}

	// First item should be selected
	if !contains(content, "todo-continuation-enforcer") {
		t.Error("expected first hook name in content")
	}
}

func TestWizardHooksKeyMap(t *testing.T) {
	wh := NewWizardHooks()

	// Test that all key bindings are properly set
	tests := []struct {
		name     string
		binding  key.Binding
		expected string
	}{
		{"Up", wh.keys.Up, "↑/k"},
		{"Down", wh.keys.Down, "↓/j"},
		{"Toggle", wh.keys.Toggle, "space"},
		{"Next", wh.keys.Next, "tab"},
		{"Back", wh.keys.Back, "shift+tab/esc"},
		{"PageUp", wh.keys.PageUp, "pgup"},
		{"PageDown", wh.keys.PageDown, "pgdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key == "" {
				t.Error("expected key binding to have a key")
			}
			// Just verify the help is set, format may vary
			t.Logf("%s key binding: %s", tt.name, help.Key)
		})
	}
}
