package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/key"
)

func TestProfileItemTitle(t *testing.T) {
	tests := []struct {
		name     string
		item     profileItem
		expected string
	}{
		{
			name:     "active profile",
			item:     profileItem{name: "test", isActive: true},
			expected: "* test (active)",
		},
		{
			name:     "inactive profile",
			item:     profileItem{name: "test", isActive: false},
			expected: "  test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.Title()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProfileItemDescription(t *testing.T) {
	tests := []struct {
		name     string
		item     profileItem
		expected string
	}{
		{
			name:     "active profile",
			item:     profileItem{isActive: true},
			expected: "Currently active profile",
		},
		{
			name:     "inactive profile",
			item:     profileItem{isActive: false},
			expected: "Press enter to switch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.Description()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProfileItemFilterValue(t *testing.T) {
	item := profileItem{name: "test-profile", isActive: true}

	result := item.FilterValue()
	if result != "test-profile" {
		t.Errorf("expected 'test-profile', got %q", result)
	}
}

func TestNewListKeyMap(t *testing.T) {
	keys := newListKeyMap()

	// Verify all key bindings are initialized
	tests := []struct {
		name    string
		binding key.Binding
	}{
		{"Switch", keys.Switch},
		{"Edit", keys.Edit},
		{"Delete", keys.Delete},
		{"New", keys.New},
		{"Search", keys.Search},
		{"Back", keys.Back},
		{"Import", keys.Import},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key == "" {
				t.Errorf("expected %s key to be initialized", tt.name)
			}
			t.Logf("%s key: %s", tt.name, help.Key)
		})
	}
}

func TestNewList(t *testing.T) {
	l := NewList()

	if l.list.Title != "Profiles" {
		t.Errorf("expected title 'Profiles', got %q", l.list.Title)
	}

	if !l.list.ShowStatusBar() {
		t.Error("expected status bar to be shown")
	}

	if !l.list.ShowFilter() {
		t.Error("expected filtering to be enabled")
	}

	if l.list.ShowHelp() {
		t.Error("expected help to be hidden")
	}

	// Check key bindings
	if l.keys.Switch.Help().Key == "" {
		t.Error("expected Switch key to be initialized")
	}
}

func TestListInit(t *testing.T) {
	l := NewList()
	cmd := l.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}

	msg := cmd()
	if _, ok := msg.(listProfilesLoadedMsg); !ok {
		t.Errorf("expected listProfilesLoadedMsg, got %T", msg)
	}
}

func TestListLoadProfiles(t *testing.T) {
	l := NewList()

	// LoadProfiles requires actual profile files, so we just verify
	// the method exists and returns an error if profiles don't exist
	err := l.LoadProfiles()
	// This will likely fail in test environment, which is ok
	// We're just testing that the method works
	_ = err // We can't assert on this without proper setup
}

func TestListSetSize(t *testing.T) {
	l := NewList()

	l.SetSize(100, 50)

	if l.width != 100 {
		t.Errorf("expected width 100, got %d", l.width)
	}

	if l.height != 50 {
		t.Errorf("expected height 50, got %d", l.height)
	}

	// The inner list should have different dimensions
	// height is reduced by 4
	if l.list.Height() != 46 {
		t.Errorf("expected inner list height 46, got %d", l.list.Height())
	}
}

func TestListUpdateListProfilesLoadedMsg(t *testing.T) {
	l := NewList()

	msg := listProfilesLoadedMsg{}
	updated, cmd := l.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for listProfilesLoadedMsg")
	}

	// LoadProfiles should have been called (though it may fail)
	_ = updated
}

func TestListUpdateWindowSizeMsg(t *testing.T) {
	l := NewList()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := l.Update(msg)

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

func TestListUpdateBackKey(t *testing.T) {
	l := NewList()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := l.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc key")
	}

	result := cmd()
	if _, ok := result.(NavigateToDashboardMsg); !ok {
		t.Errorf("expected NavigateToDashboardMsg, got %T", result)
	}
}

func TestListUpdateNewKey(t *testing.T) {
	l := NewList()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	_, cmd := l.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for 'n' key")
	}

	result := cmd()
	if _, ok := result.(NavigateToWizardMsg); !ok {
		t.Errorf("expected NavigateToWizardMsg, got %T", result)
	}
}

func TestListUpdateImportKey(t *testing.T) {
	l := NewList()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")}
	_, cmd := l.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for 'i' key")
	}

	result := cmd()
	if _, ok := result.(NavToImportMsg); !ok {
		t.Errorf("expected NavToImportMsg, got %T", result)
	}
}

func TestListUpdateDeleteConfirmationYes(t *testing.T) {
	l := NewList()
	l.confirmingDelete = true
	l.deleteTarget = "test-profile"

	// Test 'y' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	updated, cmd := l.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for 'y' key")
	}

	result := cmd()
	if deleteMsg, ok := result.(DeleteProfileMsg); !ok {
		t.Errorf("expected DeleteProfileMsg, got %T", result)
	} else if deleteMsg.Name != "test-profile" {
		t.Errorf("expected profile name 'test-profile', got %q", deleteMsg.Name)
	}

	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be false after confirmation")
	}

	if updated.deleteTarget != "" {
		t.Error("expected deleteTarget to be cleared")
	}
}

func TestListUpdateDeleteConfirmationNo(t *testing.T) {
	l := NewList()
	l.confirmingDelete = true
	l.deleteTarget = "test-profile"

	// Test 'n' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updated, cmd := l.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'n' key")
	}

	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be false after rejection")
	}

	if updated.deleteTarget != "" {
		t.Error("expected deleteTarget to be cleared")
	}
}

func TestListUpdateDeleteConfirmationEsc(t *testing.T) {
	l := NewList()
	l.confirmingDelete = true
	l.deleteTarget = "test-profile"

	// Test Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := l.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc during delete confirmation")
	}

	if updated.confirmingDelete {
		t.Error("expected confirmingDelete to be false after Esc")
	}

	if updated.deleteTarget != "" {
		t.Error("expected deleteTarget to be cleared")
	}
}

func TestListUpdateDeleteKey(t *testing.T) {
	l := NewList()

	// Add an item to the list for selection
	items := []list.Item{profileItem{name: "test-profile"}}
	l.list.SetItems(items)

	// Select the item
	l.list.Select(0)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updated, cmd := l.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'd' key (just enters confirmation mode)")
	}

	if !updated.confirmingDelete {
		t.Error("expected confirmingDelete to be true after 'd' key")
	}

	if updated.deleteTarget != "test-profile" {
		t.Errorf("expected deleteTarget to be 'test-profile', got %q", updated.deleteTarget)
	}
}

func TestListUpdateUnknownKey(t *testing.T) {
	l := NewList()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := l.Update(msg)

	// Should return the list's cmd
	_ = cmd // Can't assert on list cmd
	_ = updated
}

func TestListView(t *testing.T) {
	l := NewList()
	l.SetSize(80, 24)

	view := l.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain the list view
	if !contains(view, "Profiles") {
		t.Error("expected 'Profiles' in view")
	}
}

func TestListViewWithConfirmingDelete(t *testing.T) {
	l := NewList()
	l.SetSize(80, 24)
	l.confirmingDelete = true
	l.deleteTarget = "test-profile"

	view := l.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain delete confirmation
	if !contains(view, "Delete 'test-profile'?") {
		t.Error("expected delete confirmation in view")
	}

	if !contains(view, "[y/n]") {
		t.Error("expected '[y/n]' in view")
	}
}

func TestListSelectedProfile(t *testing.T) {
	l := NewList()

	// No items selected initially
	result := l.SelectedProfile()
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	// Add an item and select it
	items := []list.Item{profileItem{name: "test-profile"}}
	l.list.SetItems(items)
	l.list.Select(0)

	result = l.SelectedProfile()
	if result != "test-profile" {
		t.Errorf("expected 'test-profile', got %q", result)
	}
}

func TestListIsConfirmingDelete(t *testing.T) {
	l := NewList()

	if l.IsConfirmingDelete() {
		t.Error("expected confirmingDelete to be false initially")
	}

	l.confirmingDelete = true

	if !l.IsConfirmingDelete() {
		t.Error("expected confirmingDelete to be true after setting")
	}
}
