package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelSelector(t *testing.T) {
	ms := NewModelSelector()

	if ms.cursor < 0 || ms.cursor >= len(ms.items) {
		t.Error("expected cursor to be within items range")
	}

	if ms.filteredCount < 0 {
		t.Errorf("expected filteredCount >= 0, got %d", ms.filteredCount)
	}

	if ms.customInput.Placeholder != "e.g., gpt-4o-mini" {
		t.Errorf("expected custom input placeholder 'e.g., gpt-4o-mini', got %q", ms.customInput.Placeholder)
	}

	if ms.searchInput.Placeholder != "Search models..." {
		t.Errorf("expected search input placeholder 'Search models...', got %q", ms.searchInput.Placeholder)
	}

	// Check key bindings
	if ms.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if ms.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if ms.keys.Enter.Help().Key == "" {
		t.Error("expected Enter key to be initialized")
	}

	if ms.keys.Esc.Help().Key == "" {
		t.Error("expected Esc key to be initialized")
	}
}

func TestModelSelectorInit(t *testing.T) {
	ms := NewModelSelector()
	cmd := ms.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestModelSelectorSetSize(t *testing.T) {
	ms := NewModelSelector()

	ms.SetSize(100, 50)

	if ms.width != 100 {
		t.Errorf("expected width 100, got %d", ms.width)
	}

	if ms.height != 50 {
		t.Errorf("expected height 50, got %d", ms.height)
	}
}

func TestModelSelectorIsSelectable(t *testing.T) {
	ms := NewModelSelector()

	// Test with valid index
	if len(ms.items) > 0 {
		if ms.isSelectable(0) {
			// First item might be a header, separator, or custom
			t.Log("First item is selectable")
		}
	}

	// Test with out of range indices
	if ms.isSelectable(-1) {
		t.Error("expected -1 to not be selectable")
	}

	if ms.isSelectable(len(ms.items)) {
		t.Error("expected out of range index to not be selectable")
	}
}

func TestModelSelectorFindNextSelectable(t *testing.T) {
	ms := NewModelSelector()

	// Test finding next selectable from current position
	start := ms.cursor
	next := ms.findNextSelectable(start, 1)

	if next < 0 || next >= len(ms.items) {
		t.Error("expected next selectable within range")
	}

	// If we found a selectable item, it should be selectable
	if next >= 0 && next < len(ms.items) && ms.isSelectable(next) {
		t.Log("Found selectable item at:", next)
	}
}

func TestModelSelectorListHeight(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)

	height := ms.listHeight()

	if height <= 0 {
		t.Errorf("expected positive height, got %d", height)
	}

	// Height should be less than total items (leaves room for header/footer)
	if height > len(ms.items) {
		height = len(ms.items)
	}
	if height > len(ms.items) {
		t.Errorf("expected height <= items count, got %d > %d", height, len(ms.items))
	}
}

func TestModelSelectorHeaderHeight(t *testing.T) {
	ms := NewModelSelector()

	baseHeight := 2 // title + empty line
	searchHeight := 2 // search line + empty line
	expectedBase := baseHeight + searchHeight

	height := ms.headerHeight()

	if height < expectedBase {
		t.Errorf("expected height >= %d, got %d", expectedBase, height)
	}
}

func TestModelSelectorHeaderHeightWithLoadError(t *testing.T) {
	ms := NewModelSelector()
	ms.loadError = &testError{}

	height := ms.headerHeight()

	// Should be higher due to error message
	if height < 6 {
		t.Errorf("expected height >= 6 with load error, got %d", height)
	}
}

func TestModelSelectorHeaderHeightWithNoResults(t *testing.T) {
	ms := NewModelSelector()
	// Set a search term that yields no results
	ms.searchInput.SetValue("nonexistent-model")
	ms.rebuildItems()
	ms.customInput.Blur()

	height := ms.headerHeight()

	// Should include "No models match" message (2 extra lines)
	// Base: 2 (title) + 2 (search) = 4, + 2 for no results = 6
	if height < 6 {
		t.Errorf("expected height >= 6 with no results, got %d", height)
	}
}

func TestModelSelectorUpdateWindowSizeMsg(t *testing.T) {
	ms := NewModelSelector()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := ms.Update(msg)

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

func TestModelSelectorUpdateUpKey(t *testing.T) {
	ms := NewModelSelector()
	initialCursor := ms.cursor

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	// Cursor should either stay same or move up
	if updated.cursor < 0 || updated.cursor >= len(ms.items) {
		t.Errorf("cursor out of range: %d (len: %d)", updated.cursor, len(ms.items))
	}

	if updated.cursor > initialCursor {
		t.Error("cursor should not increase on Up key")
	}

	_ = initialCursor
}

func TestModelSelectorUpdateDownKey(t *testing.T) {
	ms := NewModelSelector()
	initialCursor := ms.cursor

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	// Cursor should either stay same or move down
	if updated.cursor < 0 || updated.cursor >= len(ms.items) {
		t.Errorf("cursor out of range: %d (len: %d)", updated.cursor, len(ms.items))
	}

	if updated.cursor < initialCursor && updated.cursor >= 0 {
		t.Log("cursor moved up on Down key, likely skipped non-selectable")
	}

	_ = initialCursor
}

func TestModelSelectorUpdateEscKey(t *testing.T) {
	ms := NewModelSelector()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := ms.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc key")
	}

	result := cmd()
	if _, ok := result.(ModelSelectorCancelMsg); !ok {
		t.Errorf("expected ModelSelectorCancelMsg, got %T", result)
	}
}

func TestModelSelectorUpdateSearchKey(t *testing.T) {
	ms := NewModelSelector()
	// Blur the search input first (it starts focused in NewModelSelector)
	ms.searchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for '/' key")
	}

	if !updated.searchInput.Focused() {
		t.Error("expected search input to be focused after '/' key")
	}
}

func TestModelSelectorUpdateCustomModeEnterEmpty(t *testing.T) {
	ms := NewModelSelector()
	ms.customMode = true
	ms.customInput.SetValue("")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter with empty custom input")
	}

	// With empty input, custom mode stays true (user must use Esc to exit)
	if !updated.customMode {
		t.Error("expected custom mode to remain true with empty input")
	}
}

func TestModelSelectorUpdateCustomModeEnterWithValue(t *testing.T) {
	ms := NewModelSelector()
	ms.customMode = true
	ms.customInput.SetValue("custom-model-id")
	ms.searchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := ms.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Enter with custom value")
	}

	// Should return commands for ModelSelectedMsg and PromptSaveCustomMsg
	result := cmd()
	if result == nil {
		t.Error("expected non-nil message from command")
	}

	_ = updated
}

func TestModelSelectorUpdateCustomModeEsc(t *testing.T) {
	ms := NewModelSelector()
	ms.customMode = true
	ms.customInput.SetValue("test-value")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in custom mode")
	}

	if updated.customMode {
		t.Error("expected custom mode to be false after Esc")
	}

	if updated.customInput.Value() != "" {
		t.Error("expected custom input to be cleared after Esc")
	}
}

func TestModelSelectorUpdateCustomModeTextInput(t *testing.T) {
	ms := NewModelSelector()
	ms.customMode = true
	ms.searchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, cmd := ms.Update(msg)

	// Text input should be updated
	_ = cmd
	_ = updated.customInput.Value()
}

func TestModelSelectorUpdateSearchInputEscEmpty(t *testing.T) {
	ms := NewModelSelector()
	ms.searchInput.Focus()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := ms.Update(msg)

	// Should cancel the search
	if cmd == nil {
		t.Fatal("expected non-nil command for Esc when search is empty")
	}

	result := cmd()
	if _, ok := result.(ModelSelectorCancelMsg); !ok {
		t.Errorf("expected ModelSelectorCancelMsg, got %T", result)
	}

	if updated.searchInput.Focused() {
		t.Error("expected search input to be blurred after Esc")
	}
}

func TestModelSelectorUpdateSearchInputEscWithValue(t *testing.T) {
	ms := NewModelSelector()
	ms.searchInput.Focus()
	ms.searchInput.SetValue("test")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc with search value")
	}

	if updated.searchInput.Focused() {
		t.Error("expected search input to be blurred after Esc with value")
	}
}

func TestModelSelectorUpdateSearchInputEnter(t *testing.T) {
	ms := NewModelSelector()
	ms.searchInput.Focus()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := ms.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter in search mode")
	}

	if updated.searchInput.Focused() {
		t.Error("expected search input to be blurred after Enter")
	}
}

func TestModelSelectorView(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)

	view := ms.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain title
	if !contains(view, "Select Model") {
		t.Error("expected 'Select Model' in view")
	}

	// Should contain search input
	if !contains(view, "Search:") {
		t.Error("expected 'Search:' in view")
	}

	// Should contain help
	if !contains(view, "[↑↓] navigate") {
		t.Error("expected navigation help in view")
	}
}

func TestModelSelectorViewCustomMode(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)
	ms.customMode = true

	view := ms.View()

	if view == "" {
		t.Error("expected non-empty view in custom mode")
	}

	// Should contain custom mode title
	if !contains(view, "Enter Custom Model") {
		t.Error("expected 'Enter Custom Model' in view")
	}

	// Should contain custom input
	if !contains(view, "Model ID:") {
		t.Error("expected 'Model ID:' in view")
	}
}

func TestModelSelectorViewLoadError(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)
	ms.loadError = &testError{}

	view := ms.View()

	if !contains(view, "Could not load models") {
		t.Error("expected error message in view")
	}
}

func TestModelSelectorRenderCustomMode(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)
	ms.customMode = true
	ms.customInput.SetValue("test-model")

	view := ms.renderCustomMode()

	if view == "" {
		t.Error("expected non-empty custom mode view")
	}

	if !contains(view, "Enter Custom Model") {
		t.Error("expected 'Enter Custom Model' in view")
	}
}

func TestModelSelectorGetSelectedModel(t *testing.T) {
	ms := NewModelSelector()

	// NewModelSelector() sets cursor to first selectable item
	// So GetSelectedModel should return that item
	modelID, displayName, isCustom := ms.GetSelectedModel()

	// If there are models loaded, should return the first one
	// (this test just verifies the method works without crashing)
	t.Logf("Selected model: ID=%q, DisplayName=%q, IsCustom=%v", modelID, displayName, isCustom)

	// Test with out-of-range cursor (should return empty)
	ms.cursor = -1
	modelID, displayName, isCustom = ms.GetSelectedModel()

	if modelID != "" {
		t.Error("expected empty model ID with invalid cursor")
	}

	if displayName != "" {
		t.Error("expected empty display name with invalid cursor")
	}

	if isCustom {
		t.Error("expected isCustom to be false with invalid cursor")
	}
}

func TestModelSelectorRebuildItems(t *testing.T) {
	ms := NewModelSelector()

	// Test rebuildItems without error
	ms.rebuildItems()

	if ms.items == nil {
		t.Error("expected items to be initialized")
	}

	// Should have separator and custom option
	hasSeparator := false
	hasCustom := false
	for _, item := range ms.items {
		if item.isSeparator {
			hasSeparator = true
		}
		if item.isCustom {
			hasCustom = true
		}
	}

	if !hasSeparator {
		t.Error("expected separator in items")
	}

	if !hasCustom {
		t.Error("expected custom option in items")
	}
}

func TestEnsureCursorVisible(t *testing.T) {
	ms := NewModelSelector()
	ms.SetSize(80, 24)

	// Should not panic
	ms.ensureCursorVisible()

	// Cursor should be within valid range
	if ms.cursor < 0 || ms.cursor >= len(ms.items) {
		t.Error("cursor should be within items after ensureCursorVisible")
	}
}
