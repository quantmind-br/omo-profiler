package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestJoinWithSeparator(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		sep      string
		expected []string
	}{
		{
			name:     "empty slice",
			items:    []string{},
			sep:      ",",
			expected: nil,
		},
		{
			name:     "single item",
			items:    []string{"one"},
			sep:      ",",
			expected: []string{"one"},
		},
		{
			name:     "two items",
			items:    []string{"one", "two"},
			sep:      ",",
			expected: []string{"one", ",", "two"},
		},
		{
			name:     "three items",
			items:    []string{"one", "two", "three"},
			sep:      "•",
			expected: []string{"one", "•", "two", "•", "three"},
		},
		{
			name:     "multiple items with space separator",
			items:    []string{"a", "b", "c", "d"},
			sep:      " ",
			expected: []string{"a", " ", "b", " ", "c", " ", "d"},
		},
		{
			name:     "empty string separator",
			items:    []string{"a", "b"},
			sep:      "",
			expected: []string{"a", "", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinWithSeparator(tt.items, tt.sep)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("item %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestPlaceholderView(t *testing.T) {
	app := NewApp()

	result := app.placeholderView("Test Title")

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, "Test Title") {
		t.Error("expected title to be in result")
	}

	if !strings.Contains(result, "Coming soon") {
		t.Error("expected 'Coming soon' to be in result")
	}
}

func TestNewApp(t *testing.T) {
	app := NewApp()

	if app.state != stateDashboard {
		t.Errorf("expected initial state to be stateDashboard, got %v", app.state)
	}

	if app.help.Width != 0 {
		t.Error("expected help width to be 0 initially")
	}

	if app.width != 0 || app.height != 0 {
		t.Error("expected dimensions to be 0 initially")
	}

	if app.ready {
		t.Error("expected app to not be ready initially")
	}
}

func TestAppInit(t *testing.T) {
	app := NewApp()

	cmd := app.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}

	// Init should return a batch command with dashboard.Init and spinner.Tick
	// We can't easily test the batch contents, but we can verify it's not nil
}

func TestAppUpdateWindowSizeMsg(t *testing.T) {
	app := NewApp()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, cmd := app.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for WindowSizeMsg")
	}

	updatedApp, ok := updated.(App)
	if !ok {
		t.Fatal("expected App type from Update")
	}

	if updatedApp.width != 100 {
		t.Errorf("expected width 100, got %d", updatedApp.width)
	}

	if updatedApp.height != 50 {
		t.Errorf("expected height 50, got %d", updatedApp.height)
	}

	if !updatedApp.ready {
		t.Error("expected app to be ready after WindowSizeMsg")
	}

	if updatedApp.help.Width != 100 {
		t.Errorf("expected help width 100, got %d", updatedApp.help.Width)
	}
}

func TestAppShowToast(t *testing.T) {
	app := NewApp()

	cmd := app.showToast("Test message", toastSuccess, 3*1000000000)

	if cmd == nil {
		t.Fatal("expected non-nil command from showToast")
	}

	msg := cmd()
	toast, ok := msg.(toastMsg)
	if !ok {
		t.Fatalf("expected toastMsg, got %T", msg)
	}

	if toast.text != "Test message" {
		t.Errorf("expected message 'Test message', got '%s'", toast.text)
	}

	if toast.typ != toastSuccess {
		t.Errorf("expected type toastSuccess, got %v", toast.typ)
	}

	if toast.duration != 3*1000000000 {
		t.Errorf("expected duration 3s, got %v", toast.duration)
	}
}

func TestToastMessage(t *testing.T) {
	app := NewApp()

	msg := toastMsg{text: "Test", typ: toastInfo, duration: 1000000000}
	updated, cmd := app.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command from toastMsg (tick)")
	}

	updatedApp, ok := updated.(App)
	if !ok {
		t.Fatal("expected App type from Update")
	}

	if !updatedApp.toastActive {
		t.Error("expected toast to be active")
	}

	if updatedApp.toast != "Test" {
		t.Errorf("expected toast text 'Test', got '%s'", updatedApp.toast)
	}

	if updatedApp.toastType != toastInfo {
		t.Errorf("expected toast type toastInfo, got %v", updatedApp.toastType)
	}
}

func TestClearToastMessage(t *testing.T) {
	app := NewApp()
	app.toast = "Test"
	app.toastActive = true

	msg := clearToastMsg{}
	updated, cmd := app.Update(msg)

	if cmd != nil {
		t.Error("expected nil command from clearToastMsg")
	}

	updatedApp, ok := updated.(App)
	if !ok {
		t.Fatal("expected App type from Update")
	}

	if updatedApp.toastActive {
		t.Error("expected toast to be inactive")
	}

	if updatedApp.toast != "" {
		t.Errorf("expected empty toast, got '%s'", updatedApp.toast)
	}
}

func TestRenderShortHelp(t *testing.T) {
	app := NewApp()
	app.width = 100
	app.height = 50

	// Test each state
	states := []struct {
		state        appState
		expectedHint []string
	}{
		{stateDashboard, []string{"↑↓ navigate", "enter select", "? help", "q quit"}},
		{stateList, []string{"enter switch", "e edit", "d delete", "n new", "/ search", "esc back"}},
		{stateWizard, []string{"tab/enter next", "shift+tab back", "ctrl+s save", "ctrl+c cancel"}},
		{stateDiff, []string{"tab switch pane", "enter select", "↑↓ scroll", "esc back"}},
		// stateImport and stateExport use default help
		{stateImport, []string{"? help", "q quit"}},
		{stateExport, []string{"? help", "q quit"}},
		{stateModels, []string{"n new", "i import", "e edit", "d delete", "↑↓ navigate", "esc back"}},
		{stateModelImport, []string{"space toggle", "enter import", "/ search", "↑↓ navigate", "esc back"}},
		{stateTemplateSelect, []string{"↑↓ navigate", "enter select", "esc cancel"}},
	}

	for _, tt := range states {
		t.Run("help test", func(t *testing.T) {
			app.state = tt.state
			help := app.renderShortHelp()

			if help == "" {
				t.Error("expected non-empty help")
			}

			for _, hint := range tt.expectedHint {
				if !strings.Contains(help, hint) {
					t.Errorf("expected help to contain %q", hint)
				}
			}
		})
	}
}

func TestRenderFullHelp(t *testing.T) {
	app := NewApp()
	app.width = 100
	app.height = 50

	// Test a few states
	tests := []struct {
		name    string
		state   appState
		contains []string
	}{
		{
			name:  "dashboard",
			state: stateDashboard,
			contains: []string{
				"Keyboard Shortcuts",
				"Global:",
				"q/ctrl+c",
				"Dashboard:",
				"Move up",
				"Move down",
			},
		},
		{
			name:  "list",
			state: stateList,
			contains: []string{
				"Keyboard Shortcuts",
				"Profile List:",
				"Switch to profile",
				"Edit profile",
				"Delete profile",
			},
		},
		{
			name:  "wizard",
			state: stateWizard,
			contains: []string{
				"Profile Wizard:",
				"Next step",
				"Save profile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.state = tt.state
			help := app.renderFullHelp()

			if help == "" {
				t.Error("expected non-empty help")
			}

			for _, expected := range tt.contains {
				if !strings.Contains(help, expected) {
					t.Errorf("expected help to contain %q", expected)
				}
			}
		})
	}
}

func TestNavigateTo(t *testing.T) {
	app := NewApp()
	app.width = 100
	app.height = 50

	// Test navigating to each state
	states := []appState{
		stateDashboard,
		stateList,
		stateWizard,
		stateDiff,
		stateModels,
		stateModelImport,
		stateImport,
		stateExport,
		stateTemplateSelect,
	}

	for _, targetState := range states {
		t.Run("navigation test", func(t *testing.T) {
			app.state = stateDashboard // Start from dashboard

			updated, cmd := app.navigateTo(targetState)

			// Most states return a command from Init(), but some may return nil
			// The important thing is the state changes correctly
			if updated.state != targetState {
				t.Errorf("expected state %v, got %v", targetState, updated.state)
			}

			if updated.prevState != stateDashboard {
				t.Errorf("expected prevState to be stateDashboard, got %v", updated.prevState)
			}

			_ = cmd // We can't assert on cmd without properly initializing all views
		})
	}
}

func TestDoSwitchProfile(t *testing.T) {
	app := NewApp()

	cmd := app.doSwitchProfile("test-profile")

	if cmd == nil {
		t.Fatal("expected non-nil command from doSwitchProfile")
	}

	msg := cmd()
	result, ok := msg.(switchProfileDoneMsg)
	if !ok {
		t.Fatalf("expected switchProfileDoneMsg, got %T", msg)
	}

	if result.name != "test-profile" {
		t.Errorf("expected name 'test-profile', got '%s'", result.name)
	}
}

func TestDoDeleteProfile(t *testing.T) {
	app := NewApp()

	cmd := app.doDeleteProfile("test-profile")

	if cmd == nil {
		t.Fatal("expected non-nil command from doDeleteProfile")
	}

	msg := cmd()
	result, ok := msg.(deleteProfileDoneMsg)
	if !ok {
		t.Fatalf("expected deleteProfileDoneMsg, got %T", msg)
	}

	if result.name != "test-profile" {
		t.Errorf("expected name 'test-profile', got '%s'", result.name)
	}
}
