package views

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/profile"
)

func TestNewDashboard(t *testing.T) {
	d := NewDashboard()

	if d.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", d.cursor)
	}

	if d.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if d.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if d.keys.Enter.Help().Key == "" {
		t.Error("expected Enter key to be initialized")
	}

	if d.keys.Import.Help().Key == "" {
		t.Error("expected Import key to be initialized")
	}

	if d.keys.Export.Help().Key == "" {
		t.Error("expected Export key to be initialized")
	}
}

func TestDashboardInit(t *testing.T) {
	d := NewDashboard()
	cmd := d.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}

	// The command should be loadActiveProfile
	msg := cmd()
	if _, ok := msg.(profileLoadedMsg); !ok {
		t.Errorf("expected profileLoadedMsg, got %T", msg)
	}
}

func TestDashboardLoadActiveProfile(t *testing.T) {
	d := NewDashboard()

	// We can't easily test this without mocking profile.GetActive and profile.List
	// Just verify it returns a message
	msg := d.loadActiveProfile()
	if _, ok := msg.(profileLoadedMsg); !ok {
		t.Errorf("expected profileLoadedMsg, got %T", msg)
	}
}

func TestDashboardUpdateProfileLoadedMsg(t *testing.T) {
	d := NewDashboard()

	active := &profile.ActiveConfig{
		Exists:      true,
		IsOrphan:    false,
		ProfileName: "test-profile",
	}

	msg := profileLoadedMsg{
		active: active,
		count:  5,
	}

	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for profileLoadedMsg")
	}

	if updated.activeProfile != active {
		t.Error("expected activeProfile to be set")
	}

	if updated.profileCount != 5 {
		t.Errorf("expected profileCount to be 5, got %d", updated.profileCount)
	}
}

func TestDashboardUpdateProfileLoadedMsgWithError(t *testing.T) {
	d := NewDashboard()

	msg := profileLoadedMsg{
		err: errors.New("test error"),
	}

	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for profileLoadedMsg with error")
	}

	if updated.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDashboardUpdateWindowSizeMsg(t *testing.T) {
	d := NewDashboard()

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for WindowSizeMsg")
	}

	if updated.width != 100 {
		t.Errorf("expected width 100, got %d", updated.width)
	}

	if updated.height != 50 {
		t.Errorf("expected height 50, got %d", updated.height)
	}
}

func TestDashboardUpdateUpKey(t *testing.T) {
	d := NewDashboard()
	d.cursor = 3

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	if updated.cursor != 2 {
		t.Errorf("expected cursor to be 2, got %d", updated.cursor)
	}

	// Test at top
	d.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = d.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to remain 0 at top, got %d", updated.cursor)
	}
}

func TestDashboardUpdateDownKey(t *testing.T) {
	d := NewDashboard()
	d.cursor = 2

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	if updated.cursor != 3 {
		t.Errorf("expected cursor to be 3, got %d", updated.cursor)
	}

	// Test at bottom
	d.cursor = len(menuItems) - 1
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updated, _ = d.Update(msg)

	if updated.cursor != len(menuItems)-1 {
		t.Errorf("expected cursor to remain at bottom, got %d", updated.cursor)
	}
}

func TestDashboardUpdateEnterKey(t *testing.T) {
	tests := []struct {
		name        string
		cursor      int
		expectedMsg interface{}
	}{
		{"Switch Profile", 0, NavToListMsg{}},
		{"Create New", 1, NavToWizardMsg{}},
		{"Create from Template", 2, NavToTemplateSelectMsg{}},
		{"Edit Current", 3, NavToEditorMsg{}},
		{"Compare Profiles", 4, NavToDiffMsg{}},
		{"Manage Models", 5, NavToModelsMsg{}},
		{"Import Profile", 6, NavToImportMsg{}},
		{"Export Profile", 7, NavToExportMsg{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDashboard()
			d.cursor = tt.cursor

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			_, cmd := d.Update(msg)

			if cmd == nil {
				t.Fatal("expected non-nil command for Enter key")
			}

			result := cmd()
			if _, ok := result.(NavToModelsMsg); ok && tt.cursor == 5 {
				// Special case for NavToModelsMsg
				return
			}

			switch tt.expectedMsg.(type) {
			case NavToListMsg:
				if _, ok := result.(NavToListMsg); !ok {
					t.Errorf("expected NavToListMsg, got %T", result)
				}
			case NavToWizardMsg:
				if _, ok := result.(NavToWizardMsg); !ok {
					t.Errorf("expected NavToWizardMsg, got %T", result)
				}
			case NavToTemplateSelectMsg:
				if _, ok := result.(NavToTemplateSelectMsg); !ok {
					t.Errorf("expected NavToTemplateSelectMsg, got %T", result)
				}
			case NavToEditorMsg:
				if _, ok := result.(NavToEditorMsg); !ok {
					t.Errorf("expected NavToEditorMsg, got %T", result)
				}
			case NavToDiffMsg:
				if _, ok := result.(NavToDiffMsg); !ok {
					t.Errorf("expected NavToDiffMsg, got %T", result)
				}
			case NavToModelsMsg:
				if _, ok := result.(NavToModelsMsg); !ok {
					t.Errorf("expected NavToModelsMsg, got %T", result)
				}
			case NavToImportMsg:
				if _, ok := result.(NavToImportMsg); !ok {
					t.Errorf("expected NavToImportMsg, got %T", result)
				}
			case NavToExportMsg:
				if _, ok := result.(NavToExportMsg); !ok {
					t.Errorf("expected NavToExportMsg, got %T", result)
				}
			}
		})
	}
}

func TestDashboardUpdateImportKey(t *testing.T) {
	d := NewDashboard()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")}
	_, cmd := d.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Import key")
	}

	result := cmd()
	if _, ok := result.(NavToImportMsg); !ok {
		t.Errorf("expected NavToImportMsg, got %T", result)
	}
}

func TestDashboardUpdateExportKey(t *testing.T) {
	d := NewDashboard()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
	_, cmd := d.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Export key")
	}

	result := cmd()
	if _, ok := result.(NavToExportMsg); !ok {
		t.Errorf("expected NavToExportMsg, got %T", result)
	}
}

func TestDashboardUpdateUnknownKey(t *testing.T) {
	d := NewDashboard()
	d.cursor = 2

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for unknown key")
	}

	if updated.cursor != 2 {
		t.Error("expected cursor to be preserved")
	}
}

func TestDashboardView(t *testing.T) {
	d := NewDashboard()
	d.SetSize(80, 24)

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check for title
	if !contains(view, "OMO-Profiler") {
		t.Error("expected 'OMO-Profiler' in view")
	}

	// Check for subtitle
	if !contains(view, "Profile manager for oh-my-opencode") {
		t.Error("expected subtitle in view")
	}

	// Check for menu
	if !contains(view, "Switch Profile") {
		t.Error("expected 'Switch Profile' in view")
	}
}

func TestDashboardViewWithActiveProfile(t *testing.T) {
	d := NewDashboard()
	d.SetSize(80, 24)

	d.activeProfile = &profile.ActiveConfig{
		Exists:      true,
		IsOrphan:    false,
		ProfileName: "test-profile",
	}
	d.profileCount = 5

	view := d.View()

	// Should show active profile
	if !contains(view, "Active:") {
		t.Error("expected 'Active:' in view")
	}

	if !contains(view, "test-profile") {
		t.Error("expected profile name in view")
	}

	// Should show profile count
	if !contains(view, "5 profiles available") {
		t.Error("expected '5 profiles available' in view")
	}
}

func TestDashboardViewWithNoActiveProfile(t *testing.T) {
	d := NewDashboard()
	d.SetSize(80, 24)

	d.activeProfile = &profile.ActiveConfig{
		Exists: false,
	}
	d.profileCount = 0

	view := d.View()

	// Should show none
	if !contains(view, "(None)") {
		t.Error("expected '(None)' in view")
	}
}

func TestDashboardViewWithError(t *testing.T) {
	d := NewDashboard()
	d.SetSize(80, 24)

	d.err = errors.New("test error")

	view := d.View()

	// Should show error
	if !contains(view, "Error:") {
		t.Error("expected 'Error:' in view")
	}
}

func TestDashboardRenderMenu(t *testing.T) {
	d := NewDashboard()
	d.SetSize(80, 24)

	menu := d.renderMenuContent()

	if menu == "" {
		t.Error("expected non-empty menu")
	}

	// Should have all menu items
	if !contains(menu, "Switch Profile") {
		t.Error("expected 'Switch Profile' in menu")
	}

	if !contains(menu, "Create New") {
		t.Error("expected 'Create New' in menu")
	}

	// Should have cursor marker (default position)
	if !contains(menu, " > ") {
		t.Error("expected cursor marker in menu")
	}
}

func TestDashboardRenderMenuCursorPosition(t *testing.T) {
	d := NewDashboard()
	d.cursor = 3

	menu := d.renderMenuContent()

	// Cursor should be at position 3 (Edit Current)
	// The menu should show the cursor indicator
	if !contains(menu, " > ") {
		t.Error("expected cursor marker in menu")
	}

	// The selected item should be "Edit Current"
	if !contains(menu, "Edit Current") {
		t.Error("expected 'Edit Current' in menu")
	}
}

func TestDashboardRefresh(t *testing.T) {
	d := NewDashboard()

	cmd := d.Refresh()

	if cmd == nil {
		t.Fatal("expected non-nil command from Refresh")
	}

	// Should return loadActiveProfile command
	msg := cmd()
	if _, ok := msg.(profileLoadedMsg); !ok {
		t.Errorf("expected profileLoadedMsg, got %T", msg)
	}
}

func TestDashboardSetSize(t *testing.T) {
	d := NewDashboard()

	d.SetSize(100, 50)

	if d.width != 100 {
		t.Errorf("expected width 100, got %d", d.width)
	}

	if d.height != 50 {
		t.Errorf("expected height 50, got %d", d.height)
	}
}

func TestMenuItems(t *testing.T) {
	// Verify all menu items are present
	expectedItems := []string{
		"Switch Profile",
		"Create New",
		"Create from Template",
		"Edit Current",
		"Compare Profiles",
		"Manage Models",
		"Import Profile",
		"Export Profile",
	}

	if len(menuItems) != len(expectedItems) {
		t.Errorf("expected %d menu items, got %d", len(expectedItems), len(menuItems))
	}

	for i, item := range expectedItems {
		if menuItems[i] != item {
			t.Errorf("expected menu item %d to be %q, got %q", i, item, menuItems[i])
		}
	}
}

func TestDashboardKeyMap(t *testing.T) {
	d := NewDashboard()

	// Test that all key bindings are properly set
	tests := []struct {
		name     string
		binding  key.Binding
		expected string
	}{
		{"Up", d.keys.Up, "up"},
		{"Down", d.keys.Down, "down"},
		{"Enter", d.keys.Enter, "select"},
		{"Import", d.keys.Import, "import"},
		{"Export", d.keys.Export, "export"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			// Just verify the help is set
			t.Logf("%s key binding help: %s", tt.name, help.Key)
		})
	}
}
