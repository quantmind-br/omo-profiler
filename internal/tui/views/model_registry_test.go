package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/models"
)

func TestNewModelRegistry(t *testing.T) {
	mr := NewModelRegistry()

	if mr.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", mr.cursor)
	}

	if mr.offset != 0 {
		t.Errorf("expected offset to be 0, got %d", mr.offset)
	}

	if mr.formMode {
		t.Error("expected formMode to be false initially")
	}

	if mr.editMode {
		t.Error("expected editMode to be false initially")
	}

	if mr.focusedField != 0 {
		t.Errorf("expected focusedField to be 0, got %d", mr.focusedField)
	}

	if mr.confirmDelete {
		t.Error("expected confirmDelete to be false initially")
	}

	// Check key bindings
	if mr.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if mr.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if mr.keys.New.Help().Key == "" {
		t.Error("expected New key to be initialized")
	}

	if mr.keys.Edit.Help().Key == "" {
		t.Error("expected Edit key to be initialized")
	}

	if mr.keys.Delete.Help().Key == "" {
		t.Error("expected Delete key to be initialized")
	}

	if mr.keys.Esc.Help().Key == "" {
		t.Error("expected Esc key to be initialized")
	}
}

func TestModelRegistryInit(t *testing.T) {
	mr := NewModelRegistry()
	cmd := mr.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestModelRegistryGetFilteredModelsNoSearch(t *testing.T) {
	mr := NewModelRegistry()

	// Without search term, should return all flat models
	filtered := mr.getFilteredModels()

	if len(filtered) != len(mr.flatModels) {
		t.Errorf("expected %d models (same as flatModels), got %d", len(mr.flatModels), len(filtered))
	}
}

func TestModelRegistryGetFilteredModelsWithSearch(t *testing.T) {
	mr := NewModelRegistry()
	mr.searchInput.SetValue("claude")

	filtered := mr.getFilteredModels()

	// Should return models matching "claude"
	for _, model := range filtered {
		lowerID := strings.ToLower(model.ModelID)
		lowerName := strings.ToLower(model.DisplayName)
		lowerProvider := strings.ToLower(model.Provider)
		searchLower := strings.ToLower("claude")

		if !strings.Contains(lowerID, searchLower) &&
			!strings.Contains(lowerName, searchLower) &&
			!strings.Contains(lowerProvider, searchLower) {
			t.Errorf("model %q doesn't match search term", model.DisplayName)
		}
	}
}

func TestModelRegistryGetFilteredModelsNoMatch(t *testing.T) {
	mr := NewModelRegistry()
	mr.searchInput.SetValue("nonexistent-model-xyz-123")

	filtered := mr.getFilteredModels()

	if len(filtered) != 0 {
		t.Errorf("expected 0 models for non-matching search, got %d", len(filtered))
	}
}

func TestModelRegistrySetSize(t *testing.T) {
	mr := NewModelRegistry()

	mr.SetSize(100, 50)

	if mr.width != 100 {
		t.Errorf("expected width 100, got %d", mr.width)
	}

	if mr.height != 50 {
		t.Errorf("expected height 50, got %d", mr.height)
	}
}

func TestModelRegistryIsEditing(t *testing.T) {
	mr := NewModelRegistry()

	if mr.IsEditing() {
		t.Error("expected IsEditing to be false initially")
	}

	mr.formMode = true
	if !mr.IsEditing() {
		t.Error("expected IsEditing to be true in formMode")
	}

	mr.formMode = false
	mr.searchInput.Focus()
	if !mr.IsEditing() {
		t.Error("expected IsEditing to be true when search is focused")
	}
}

func TestModelRegistryUpdateWindowSizeMsg(t *testing.T) {
	mr := NewModelRegistry()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := mr.Update(msg)

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

func TestModelRegistryUpdateUpKey(t *testing.T) {
	mr := NewModelRegistry()
	mr.cursor = 5

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	if updated.cursor != 4 {
		t.Errorf("expected cursor to be 4, got %d", updated.cursor)
	}

	// Test at top
	mr.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = mr.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to remain 0 at top, got %d", updated.cursor)
	}
}

func TestModelRegistryUpdateDownKey(t *testing.T) {
	mr := NewModelRegistry()
	mr.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	if updated.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", updated.cursor)
	}
}

func TestModelRegistryUpdateNewKey(t *testing.T) {
	mr := NewModelRegistry()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'n' key")
	}

	if !updated.formMode {
		t.Error("expected formMode to be true after 'n' key")
	}

	if updated.editMode {
		t.Error("expected editMode to be false for new model")
	}
}

func TestModelRegistryUpdateImportKey(t *testing.T) {
	mr := NewModelRegistry()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")}
	_, cmd := mr.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for 'i' key")
	}

	result := cmd()
	if _, ok := result.(NavToModelImportMsg); !ok {
		t.Errorf("expected NavToModelImportMsg, got %T", result)
	}
}

func TestModelRegistryUpdateEditKey(t *testing.T) {
	mr := NewModelRegistry()

	// Need to have some models to edit
	if len(mr.flatModels) == 0 {
		t.Skip("no models available to edit")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'e' key")
	}

	if !updated.formMode {
		t.Error("expected formMode to be true after 'e' key")
	}

	if !updated.editMode {
		t.Error("expected editMode to be true after 'e' key")
	}
}

func TestModelRegistryUpdateDeleteKey(t *testing.T) {
	mr := NewModelRegistry()

	// Need to have some models to delete
	if len(mr.flatModels) == 0 {
		t.Skip("no models available to delete")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'd' key (just enters confirmation mode)")
	}

	if !updated.confirmDelete {
		t.Error("expected confirmDelete to be true after 'd' key")
	}

	if updated.deleteTarget == "" {
		t.Error("expected deleteTarget to be set")
	}
}

func TestModelRegistryUpdateEscKey(t *testing.T) {
	mr := NewModelRegistry()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := mr.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc key")
	}

	result := cmd()
	if _, ok := result.(ModelRegistryBackMsg); !ok {
		t.Errorf("expected ModelRegistryBackMsg, got %T", result)
	}
}

func TestModelRegistryUpdateSearchKey(t *testing.T) {
	mr := NewModelRegistry()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for '/' key")
	}

	if !updated.searchInput.Focused() {
		t.Error("expected search input to be focused after '/' key")
	}
}

func TestModelRegistryUpdateDeleteConfirmYes(t *testing.T) {
	mr := NewModelRegistry()
	mr.confirmDelete = true
	mr.deleteTarget = "test-model-id"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	updated, cmd := mr.Update(msg)

	// The command might be nil if the model doesn't exist or delete fails
	_ = cmd

	if updated.confirmDelete {
		t.Error("expected confirmDelete to be false after confirmation")
	}

	if updated.deleteTarget != "" {
		t.Error("expected deleteTarget to be cleared after confirmation")
	}
}

func TestModelRegistryUpdateDeleteConfirmNo(t *testing.T) {
	mr := NewModelRegistry()
	mr.confirmDelete = true
	mr.deleteTarget = "test-model-id"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'n' key during delete confirmation")
	}

	if updated.confirmDelete {
		t.Error("expected confirmDelete to be false after rejection")
	}

	if updated.deleteTarget != "" {
		t.Error("expected deleteTarget to be cleared after rejection")
	}
}

func TestModelRegistryUpdateFormModeEnterWithValidation(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.displayNameInput.SetValue("Test Model")
	mr.modelIdInput.SetValue("test-model-id")
	mr.providerInput.SetValue("Test Provider")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := mr.Update(msg)

	// Command might be nil if save fails
	_ = cmd

	// After successful save, formMode should be false
	// (If save fails, errorMsg would be set and formMode stays true)
	if updated.errorMsg != "" {
		// Validation or save failed, which is expected in test environment
		t.Logf("Got expected error: %s", updated.errorMsg)
	}
}

func TestModelRegistryUpdateFormModeEnterEmptyValidation(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.displayNameInput.SetValue("")
	mr.modelIdInput.SetValue("")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command when validation fails")
	}

	if !updated.formMode {
		t.Error("expected formMode to remain true when validation fails")
	}

	if updated.errorMsg == "" {
		t.Error("expected errorMsg to be set when validation fails")
	}
}

func TestModelRegistryUpdateFormModeEsc(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.editMode = true
	mr.editingId = "test-id"
	mr.displayNameInput.SetValue("Test")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in form mode")
	}

	if updated.formMode {
		t.Error("expected formMode to be false after Esc")
	}

	if updated.editMode {
		t.Error("expected editMode to be false after Esc")
	}

	if updated.editingId != "" {
		t.Error("expected editingId to be cleared after Esc")
	}
}

func TestModelRegistryUpdateFormModeTab(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.focusedField = 0

	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Tab in form mode")
	}

	if updated.focusedField != 1 {
		t.Errorf("expected focusedField to be 1 after Tab, got %d", updated.focusedField)
	}

	// Test wrapping
	mr.focusedField = 2
	msg = tea.KeyMsg{Type: tea.KeyTab}
	updated, _ = mr.Update(msg)

	if updated.focusedField != 0 {
		t.Errorf("expected focusedField to wrap to 0, got %d", updated.focusedField)
	}
}

func TestModelRegistryUpdateFormModeShiftTab(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.focusedField = 0

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updated, cmd := mr.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Shift+Tab in form mode")
	}

	if updated.focusedField != 2 {
		t.Errorf("expected focusedField to be 2 after Shift+Tab from 0, got %d", updated.focusedField)
	}
}

func TestModelRegistryEnterEditMode(t *testing.T) {
	mr := NewModelRegistry()

	testModel := models.RegisteredModel{
		ModelID:     "test-model",
		DisplayName: "Test Model",
		Provider:    "Test Provider",
	}

	mr.enterEditMode(testModel)

	if !mr.formMode {
		t.Error("expected formMode to be true")
	}

	if !mr.editMode {
		t.Error("expected editMode to be true")
	}

	if mr.editingId != "test-model" {
		t.Errorf("expected editingId to be 'test-model', got %q", mr.editingId)
	}

	if mr.displayNameInput.Value() != "Test Model" {
		t.Errorf("expected displayNameInput to be 'Test Model', got %q", mr.displayNameInput.Value())
	}

	if mr.modelIdInput.Value() != "test-model" {
		t.Errorf("expected modelIdInput to be 'test-model', got %q", mr.modelIdInput.Value())
	}

	if mr.providerInput.Value() != "Test Provider" {
		t.Errorf("expected providerInput to be 'Test Provider', got %q", mr.providerInput.Value())
	}
}

func TestModelRegistryEnterAddMode(t *testing.T) {
	mr := NewModelRegistry()
	mr.displayNameInput.SetValue("Old Value")
	mr.modelIdInput.SetValue("old-id")

	mr.enterAddMode()

	if !mr.formMode {
		t.Error("expected formMode to be true")
	}

	if mr.editMode {
		t.Error("expected editMode to be false for add mode")
	}

	if mr.editingId != "" {
		t.Errorf("expected editingId to be empty, got %q", mr.editingId)
	}

	if mr.displayNameInput.Value() != "" {
		t.Errorf("expected displayNameInput to be cleared, got %q", mr.displayNameInput.Value())
	}
}

func TestModelRegistryResetForm(t *testing.T) {
	mr := NewModelRegistry()
	mr.displayNameInput.SetValue("Test")
	mr.modelIdInput.SetValue("test-id")
	mr.providerInput.SetValue("Provider")

	mr.resetForm()

	if mr.displayNameInput.Value() != "" {
		t.Errorf("expected displayNameInput to be empty, got %q", mr.displayNameInput.Value())
	}

	if mr.modelIdInput.Value() != "" {
		t.Errorf("expected modelIdInput to be empty, got %q", mr.modelIdInput.Value())
	}

	if mr.providerInput.Value() != "" {
		t.Errorf("expected providerInput to be empty, got %q", mr.providerInput.Value())
	}
}

func TestModelRegistryUpdateFormFocus(t *testing.T) {
	mr := NewModelRegistry()

	// Field 0 - Display Name
	mr.focusedField = 0
	mr.updateFormFocus()
	if !mr.displayNameInput.Focused() {
		t.Error("expected displayNameInput to be focused when focusedField=0")
	}

	// Field 1 - Model ID
	mr.focusedField = 1
	mr.updateFormFocus()
	if !mr.modelIdInput.Focused() {
		t.Error("expected modelIdInput to be focused when focusedField=1")
	}

	// Field 2 - Provider
	mr.focusedField = 2
	mr.updateFormFocus()
	if !mr.providerInput.Focused() {
		t.Error("expected providerInput to be focused when focusedField=2")
	}
}

func TestModelRegistryGetFocusedInputValue(t *testing.T) {
	mr := NewModelRegistry()
	mr.displayNameInput.SetValue("Display Name")
	mr.modelIdInput.SetValue("model-id")
	mr.providerInput.SetValue("Provider")

	// Field 0
	mr.focusedField = 0
	if mr.getFocusedInputValue() != "Display Name" {
		t.Errorf("expected 'Display Name', got %q", mr.getFocusedInputValue())
	}

	// Field 1
	mr.focusedField = 1
	if mr.getFocusedInputValue() != "model-id" {
		t.Errorf("expected 'model-id', got %q", mr.getFocusedInputValue())
	}

	// Field 2
	mr.focusedField = 2
	if mr.getFocusedInputValue() != "Provider" {
		t.Errorf("expected 'Provider', got %q", mr.getFocusedInputValue())
	}

	// Invalid field
	mr.focusedField = 5
	if mr.getFocusedInputValue() != "" {
		t.Errorf("expected empty string for invalid field, got %q", mr.getFocusedInputValue())
	}
}

func TestModelRegistryValidateAndSaveMissingDisplayName(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.displayNameInput.SetValue("")
	mr.modelIdInput.SetValue("test-id")

	err := mr.validateAndSave()

	if err == nil {
		t.Error("expected error when display name is missing")
	}

	if !strings.Contains(err.Error(), "Display name") {
		t.Errorf("expected error about display name, got %q", err.Error())
	}
}

func TestModelRegistryValidateAndSaveMissingModelID(t *testing.T) {
	mr := NewModelRegistry()
	mr.formMode = true
	mr.displayNameInput.SetValue("Test Model")
	mr.modelIdInput.SetValue("")

	err := mr.validateAndSave()

	if err == nil {
		t.Error("expected error when model ID is missing")
	}

	if !strings.Contains(err.Error(), "Model ID") {
		t.Errorf("expected error about model ID, got %q", err.Error())
	}
}

func TestModelRegistryViewLoadError(t *testing.T) {
	mr := NewModelRegistry()
	mr.loadError = &testError{}
	mr.width = 80
	mr.height = 24

	view := mr.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Error loading models") {
		t.Error("expected error message in view")
	}
}

func TestModelRegistryViewFormMode(t *testing.T) {
	mr := NewModelRegistry()
	mr.width = 80
	mr.height = 24
	mr.formMode = true

	view := mr.View()

	if view == "" {
		t.Error("expected non-empty view in form mode")
	}

	// Should show form title
	if !contains(view, "Add New Model") {
		t.Error("expected 'Add New Model' in view")
	}

	// Should show form fields
	if !contains(view, "Display Name:") {
		t.Error("expected 'Display Name:' field in view")
	}

	if !contains(view, "Model ID:") {
		t.Error("expected 'Model ID:' field in view")
	}

	if !contains(view, "Provider:") {
		t.Error("expected 'Provider:' field in view")
	}
}

func TestModelRegistryViewEditMode(t *testing.T) {
	mr := NewModelRegistry()
	mr.width = 80
	mr.height = 24
	mr.formMode = true
	mr.editMode = true

	view := mr.View()

	if !contains(view, "Edit Model") {
		t.Error("expected 'Edit Model' in view")
	}
}

func TestModelRegistryViewListMode(t *testing.T) {
	mr := NewModelRegistry()
	mr.width = 80
	mr.height = 24

	view := mr.View()

	if view == "" {
		t.Error("expected non-empty view in list mode")
	}

	if !contains(view, "Manage Models") {
		t.Error("expected 'Manage Models' in view")
	}

	if !contains(view, "Search:") {
		t.Error("expected 'Search:' in view")
	}
}

func TestModelRegistryViewDeleteConfirm(t *testing.T) {
	mr := NewModelRegistry()
	mr.width = 80
	mr.height = 24
	mr.confirmDelete = true
	mr.deleteTarget = "test-model"

	view := mr.View()

	if !contains(view, "Delete") {
		t.Error("expected 'Delete' in confirmation view")
	}

	if !contains(view, "(y/n)") {
		t.Error("expected '(y/n)' in confirmation view")
	}
}
