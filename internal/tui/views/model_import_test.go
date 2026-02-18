package views

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/models"
)

func TestNewModelImport(t *testing.T) {
	mi := NewModelImport()

	if mi.state != stateImportLoading {
		t.Errorf("expected stateImportLoading, got %v", mi.state)
	}

	if mi.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", mi.cursor)
	}

	if mi.offset != 0 {
		t.Errorf("expected offset to be 0, got %d", mi.offset)
	}

	if mi.providerOffset != 0 {
		t.Errorf("expected providerOffset to be 0, got %d", mi.providerOffset)
	}

	if mi.selectedModels == nil {
		t.Error("expected selectedModels to be initialized")
	}

	if len(mi.selectedModels) != 0 {
		t.Errorf("expected empty selectedModels, got %d items", len(mi.selectedModels))
	}

	if mi.registry == nil {
		t.Error("expected registry to be loaded")
	}

	if mi.providerSearchInput.Focused() {
		t.Error("expected provider search input to not be focused initially")
	}

	if mi.providerSearchInput.Value() != "" {
		t.Errorf("expected empty provider search input value, got %q", mi.providerSearchInput.Value())
	}

	if mi.providerSearchInput.Placeholder != "Search providers..." {
		t.Errorf("expected provider search placeholder 'Search providers...', got %q", mi.providerSearchInput.Placeholder)
	}

	// Check key bindings
	if mi.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if mi.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if mi.keys.Enter.Help().Key == "" {
		t.Error("expected Enter key to be initialized")
	}

	if mi.keys.Space.Help().Key == "" {
		t.Error("expected Space key to be initialized")
	}

	if mi.keys.Esc.Help().Key == "" {
		t.Error("expected Esc key to be initialized")
	}

	if mi.keys.Retry.Help().Key == "" {
		t.Error("expected Retry key to be initialized")
	}
}

func TestModelImportInit(t *testing.T) {
	mi := NewModelImport()
	cmd := mi.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}

	// The command should be a Batch with spinner.Tick and fetchModelsDevCmd
	batchCmd := cmd()
	// We can't easily test the batch contents, just verify it's not nil
	if batchCmd == nil {
		t.Error("expected batch command to return non-nil")
	}
}

func TestModelImportSetSize(t *testing.T) {
	mi := NewModelImport()

	mi.SetSize(100, 50)

	if mi.width != 100 {
		t.Errorf("expected width 100, got %d", mi.width)
	}

	if mi.height != 50 {
		t.Errorf("expected height 50, got %d", mi.height)
	}
}

func TestModelImportIsEditing(t *testing.T) {
	mi := NewModelImport()

	if mi.IsEditing() {
		t.Error("expected IsEditing to be false initially")
	}

	mi.searchInput.Focus()
	if !mi.IsEditing() {
		t.Error("expected IsEditing to be true when search is focused")
	}

	mi.searchInput.Blur()
	mi.providerSearchInput.Focus()
	if !mi.IsEditing() {
		t.Error("expected IsEditing to be true when provider search is focused")
	}
}

func TestModelImportUpdateFetchModelsDevMsgSuccess(t *testing.T) {
	mi := NewModelImport()

	// Create a mock response using the map structure
	mockResponse := &models.ModelsDevResponse{
		"anthropic": {
			ID:   "anthropic",
			Name: "Anthropic",
			Models: map[string]models.ModelsDevModel{
				"claude-3": {ID: "claude-3", Name: "Claude 3"},
			},
		},
		"openai": {
			ID:   "openai",
			Name: "OpenAI",
			Models: map[string]models.ModelsDevModel{
				"gpt-4": {ID: "gpt-4", Name: "GPT-4"},
			},
		},
	}

	msg := fetchModelsDevMsg{
		response: mockResponse,
		err:      nil,
	}

	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for successful fetch")
	}

	if updated.state != stateImportProviderList {
		t.Errorf("expected stateImportProviderList, got %v", updated.state)
	}

	if updated.response == nil {
		t.Error("expected response to be set")
	}

	if len(updated.providers) == 0 {
		t.Error("expected providers to be populated")
	}
}

func TestModelImportUpdateFetchModelsDevMsgError(t *testing.T) {
	mi := NewModelImport()

	msg := fetchModelsDevMsg{
		response: nil,
		err:      &testError{},
	}

	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for error fetch")
	}

	if updated.state != stateImportError {
		t.Errorf("expected stateImportError, got %v", updated.state)
	}

	if updated.errorMsg == "" {
		t.Error("expected errorMsg to be set")
	}
}

func TestModelImportUpdateSpinnerTickMsg(t *testing.T) {
	mi := NewModelImport()

	// Create a spinner tick message
	msg := spinner.TickMsg{}
	updated, cmd := mi.Update(msg)

	// Should return the spinner's command (might be nil)
	_ = cmd

	// State should still be loading
	if updated.state != stateImportLoading {
		t.Errorf("expected state to remain stateImportLoading, got %v", updated.state)
	}
}

func TestModelImportHandleProviderListKeysUp(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
		{ID: "openai", Name: "OpenAI", ModelCount: 3},
	}
	mi.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", updated.cursor)
	}

	// Test at top
	mi.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = mi.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to remain 0 at top, got %d", updated.cursor)
	}
}

func TestModelImportHandleProviderListKeysDown(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
		{ID: "openai", Name: "OpenAI", ModelCount: 3},
	}
	mi.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	if updated.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", updated.cursor)
	}
}

func TestModelImportHandleProviderListKeysEnter(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.response = &models.ModelsDevResponse{
		"anthropic": {
			ID:   "anthropic",
			Name: "Anthropic",
			Models: map[string]models.ModelsDevModel{
				"claude-3": {ID: "claude-3", Name: "Claude 3"},
			},
		},
	}
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 1},
	}
	mi.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter key")
	}

	if updated.state != stateImportModelList {
		t.Errorf("expected stateImportModelList, got %v", updated.state)
	}

	if updated.selectedProvider != "anthropic" {
		t.Errorf("expected selectedProvider 'anthropic', got %q", updated.selectedProvider)
	}

	if !updated.searchInput.Focused() {
		t.Error("expected search input to be focused after selecting provider")
	}

	if updated.providerSearchInput.Focused() {
		t.Error("expected provider search input to be blurred after selecting provider")
	}

	if updated.providerSearchInput.Value() != "" {
		t.Errorf("expected provider search input to be cleared after selecting provider, got %q", updated.providerSearchInput.Value())
	}
}

func TestModelImportHandleProviderListKeysSearchSlash(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providerSearchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for '/' key")
	}

	if !updated.providerSearchInput.Focused() {
		t.Error("expected provider search input to be focused after '/' key")
	}
}

func TestModelImportHandleProviderSearchEsc(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providerSearchInput.Focus()
	mi.providerSearchInput.SetValue("test")
	mi.cursor = 5
	mi.providerOffset = 2

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in provider search mode")
	}

	if updated.providerSearchInput.Focused() {
		t.Error("expected provider search input to be blurred after Esc")
	}

	if updated.providerSearchInput.Value() != "" {
		t.Errorf("expected provider search value to be cleared, got %q", updated.providerSearchInput.Value())
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", updated.cursor)
	}

	if updated.providerOffset != 0 {
		t.Errorf("expected providerOffset to be reset to 0, got %d", updated.providerOffset)
	}
}

func TestModelImportHandleProviderSearchEnter(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providerSearchInput.Focus()
	mi.providerSearchInput.SetValue("anth")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter in provider search mode")
	}

	if updated.providerSearchInput.Focused() {
		t.Error("expected provider search input to be blurred after Enter")
	}

	if updated.providerSearchInput.Value() != "anth" {
		t.Errorf("expected provider search value to be preserved, got %q", updated.providerSearchInput.Value())
	}
}

func TestModelImportHandleProviderSearchFiltering(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.providerSearchInput.Focus()
	mi.cursor = 3
	mi.providerOffset = 2

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updated, _ := mi.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", updated.cursor)
	}

	if updated.providerOffset != 0 {
		t.Errorf("expected providerOffset to be reset to 0, got %d", updated.providerOffset)
	}
}

func TestModelImportHandleProviderListEnterClearsProviderSearch(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.response = &models.ModelsDevResponse{
		"anthropic": {
			ID:   "anthropic",
			Name: "Anthropic",
			Models: map[string]models.ModelsDevModel{
				"claude-3": {ID: "claude-3", Name: "Claude 3"},
			},
		},
	}
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 1},
	}
	mi.cursor = 0
	mi.providerSearchInput.SetValue("anth")
	mi.providerSearchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter key")
	}

	if updated.state != stateImportModelList {
		t.Errorf("expected stateImportModelList, got %v", updated.state)
	}

	if updated.providerSearchInput.Value() != "" {
		t.Errorf("expected provider search input to be cleared after selecting provider, got %q", updated.providerSearchInput.Value())
	}

	if updated.providerSearchInput.Focused() {
		t.Error("expected provider search input to be blurred after selecting provider")
	}
}

func TestModelImportHandleProviderListKeysEsc(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := mi.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc key")
	}

	result := cmd()
	if _, ok := result.(ModelImportBackMsg); !ok {
		t.Errorf("expected ModelImportBackMsg, got %T", result)
	}
}

func TestModelImportHandleModelListKeysUp(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.providerModels = []models.ModelsDevModel{
		{ID: "model1", Name: "Model 1"},
		{ID: "model2", Name: "Model 2"},
	}
	mi.searchInput.Blur()
	mi.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", updated.cursor)
	}
}

func TestModelImportHandleModelListKeysDown(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.providerModels = []models.ModelsDevModel{
		{ID: "model1", Name: "Model 1"},
		{ID: "model2", Name: "Model 2"},
	}
	mi.searchInput.Blur()
	mi.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	if updated.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", updated.cursor)
	}
}

func TestModelImportHandleModelListKeysSpace(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.providerModels = []models.ModelsDevModel{
		{ID: "model1", Name: "Model 1"},
		{ID: "model2", Name: "Model 2"},
	}
	mi.searchInput.Blur()
	mi.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Space key")
	}

	if !updated.selectedModels["model1"] {
		t.Error("expected model1 to be selected after Space")
	}

	// Toggle again
	msg = tea.KeyMsg{Type: tea.KeySpace}
	updated, _ = mi.Update(msg)

	if updated.selectedModels["model1"] {
		t.Error("expected model1 to be deselected after second Space")
	}
}

func TestModelImportHandleModelListKeysEnter(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.providerModels = []models.ModelsDevModel{
		{ID: "model1", Name: "Model 1"},
	}
	mi.searchInput.Blur()
	mi.cursor = 0
	mi.selectedModels["model1"] = true

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := mi.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Enter key")
	}

	result := cmd()
	if _, ok := result.(ModelImportDoneMsg); !ok {
		t.Errorf("expected ModelImportDoneMsg, got %T", result)
	}
}

func TestModelImportHandleModelListKeysEsc(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.selectedProvider = "anthropic"
	mi.selectedModels["model1"] = true
	mi.searchInput.SetValue("test")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc key")
	}

	if updated.state != stateImportProviderList {
		t.Errorf("expected stateImportProviderList, got %v", updated.state)
	}

	// selectedProvider is NOT cleared - it remains for reference
	if updated.selectedProvider != "anthropic" {
		t.Errorf("expected selectedProvider to remain 'anthropic', got %q", updated.selectedProvider)
	}

	if len(updated.selectedModels) != 0 {
		t.Errorf("expected selectedModels to be cleared, got %d items", len(updated.selectedModels))
	}

	if updated.searchInput.Value() != "" {
		t.Errorf("expected search input to be cleared, got %q", updated.searchInput.Value())
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", updated.cursor)
	}
}

func TestModelImportHandleModelListKeysSearchSlash(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.searchInput.Blur()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for '/' key")
	}

	if !updated.searchInput.Focused() {
		t.Error("expected search input to be focused after '/' key")
	}
}

func TestModelImportHandleModelListSearchEsc(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.searchInput.Focus()
	mi.searchInput.SetValue("test")
	mi.cursor = 5

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := mi.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in search mode")
	}

	if updated.searchInput.Focused() {
		t.Error("expected search input to be blurred after Esc")
	}

	if updated.searchInput.Value() != "" {
		t.Errorf("expected search value to be cleared, got %q", updated.searchInput.Value())
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be reset to 0, got %d", updated.cursor)
	}
}

func TestModelImportHandleErrorKeysRetry(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportError
	mi.errorMsg = "test error"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updated, cmd := mi.Update(msg)

	// Should return a batch command
	if cmd == nil {
		t.Error("expected non-nil command for Retry key")
	}

	if updated.state != stateImportLoading {
		t.Errorf("expected stateImportLoading, got %v", updated.state)
	}

	if updated.errorMsg != "" {
		t.Errorf("expected errorMsg to be cleared, got %q", updated.errorMsg)
	}
}

func TestModelImportHandleErrorKeysEsc(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportError

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := mi.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc key")
	}

	result := cmd()
	if _, ok := result.(ModelImportBackMsg); !ok {
		t.Errorf("expected ModelImportBackMsg, got %T", result)
	}
}

func TestModelImportGetFilteredModelsNoSearch(t *testing.T) {
	mi := NewModelImport()
	mi.providerModels = []models.ModelsDevModel{
		{ID: "model1", Name: "Model 1"},
		{ID: "model2", Name: "Model 2"},
	}
	mi.searchInput.SetValue("")

	filtered := mi.getFilteredModels()

	if len(filtered) != len(mi.providerModels) {
		t.Errorf("expected %d models, got %d", len(mi.providerModels), len(filtered))
	}
}

func TestModelImportGetFilteredProvidersNoSearch(t *testing.T) {
	mi := NewModelImport()
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
		{ID: "openai", Name: "OpenAI", ModelCount: 3},
	}
	mi.providerSearchInput.SetValue("")

	filtered := mi.getFilteredProviders()

	if len(filtered) != 2 {
		t.Errorf("expected 2 providers, got %d", len(filtered))
	}
}

func TestModelImportGetFilteredProvidersWithSearch(t *testing.T) {
	mi := NewModelImport()
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
		{ID: "openai", Name: "OpenAI", ModelCount: 3},
	}
	mi.providerSearchInput.SetValue("anth")

	filtered := mi.getFilteredProviders()

	if len(filtered) != 1 {
		t.Errorf("expected 1 provider, got %d", len(filtered))
	}

	if filtered[0].ID != "anthropic" {
		t.Errorf("expected anthropic, got %q", filtered[0].ID)
	}
}

func TestModelImportGetFilteredProvidersNoMatch(t *testing.T) {
	mi := NewModelImport()
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
		{ID: "openai", Name: "OpenAI", ModelCount: 3},
	}
	mi.providerSearchInput.SetValue("nonexistent")

	filtered := mi.getFilteredProviders()

	if len(filtered) != 0 {
		t.Errorf("expected 0 providers, got %d", len(filtered))
	}
}

func TestModelImportGetFilteredModelsWithSearch(t *testing.T) {
	mi := NewModelImport()
	mi.providerModels = []models.ModelsDevModel{
		{ID: "claude-3", Name: "Claude 3"},
		{ID: "gpt-4", Name: "GPT-4"},
	}
	mi.selectedProvider = "anthropic"
	mi.searchInput.SetValue("claude")

	filtered := mi.getFilteredModels()

	// Should return models matching "claude"
	if len(filtered) != 1 {
		t.Errorf("expected 1 model, got %d", len(filtered))
	}

	if filtered[0].ID != "claude-3" {
		t.Errorf("expected claude-3, got %q", filtered[0].ID)
	}
}

func TestModelImportGetFilteredModelsNoMatch(t *testing.T) {
	mi := NewModelImport()
	mi.providerModels = []models.ModelsDevModel{
		{ID: "claude-3", Name: "Claude 3"},
	}
	mi.selectedProvider = "anthropic"
	mi.searchInput.SetValue("nonexistent")

	filtered := mi.getFilteredModels()

	if len(filtered) != 0 {
		t.Errorf("expected 0 models for non-matching search, got %d", len(filtered))
	}
}

func TestModelImportViewLoading(t *testing.T) {
	mi := NewModelImport()
	mi.width = 80
	mi.height = 24

	view := mi.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Import from models.dev") {
		t.Error("expected 'Import from models.dev' in view")
	}

	if !contains(view, "Loading providers") {
		t.Error("expected 'Loading providers' in view")
	}
}

func TestModelImportViewProviderList(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.width = 80
	mi.height = 24
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
	}

	view := mi.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Import from models.dev") {
		t.Error("expected 'Import from models.dev' in view")
	}

	if !contains(view, "Anthropic") {
		t.Error("expected 'Anthropic' in view")
	}

	if !contains(view, "5 models") {
		t.Error("expected '5 models' in view")
	}

	if !contains(view, "[↑↓] navigate") {
		t.Error("expected navigation help in view")
	}

	if !contains(view, "[/] search") {
		t.Error("expected search help in view")
	}

	if !contains(view, "Search") {
		t.Error("expected search line in view")
	}
}

func TestModelImportViewProviderListNoMatch(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportProviderList
	mi.width = 80
	mi.height = 24
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 5},
	}
	mi.providerSearchInput.SetValue("nonexistent")

	view := mi.View()

	if !contains(view, "No providers match the search.") {
		t.Error("expected no-match provider search message in view")
	}
}

func TestModelImportIsEditingProviderSearch(t *testing.T) {
	mi := NewModelImport()

	mi.providerSearchInput.Focus()
	if !mi.IsEditing() {
		t.Error("expected IsEditing to be true when provider search is focused")
	}

	mi.providerSearchInput.Blur()
	if mi.IsEditing() {
		t.Error("expected IsEditing to be false when provider search is blurred")
	}
}

func TestModelImportViewModelList(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportModelList
	mi.width = 80
	mi.height = 24
	mi.selectedProvider = "anthropic"
	mi.providers = []models.ProviderWithCount{
		{ID: "anthropic", Name: "Anthropic", ModelCount: 1},
	}
	mi.providerModels = []models.ModelsDevModel{
		{ID: "claude-3", Name: "Claude 3"},
	}
	mi.selectedModels = map[string]bool{"claude-3": true}

	view := mi.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Import from Anthropic") {
		t.Error("expected 'Import from Anthropic' in view")
	}

	if !contains(view, "claude-3") {
		t.Error("expected 'claude-3' in view")
	}

	if !contains(view, "[x]") {
		t.Error("expected checked checkbox in view")
	}

	if !contains(view, "1 selected") {
		t.Error("expected '1 selected' in view")
	}
}

func TestModelImportViewError(t *testing.T) {
	mi := NewModelImport()
	mi.state = stateImportError
	mi.width = 80
	mi.height = 24
	mi.errorMsg = "Failed to fetch"

	view := mi.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Error:") {
		t.Error("expected 'Error:' in view")
	}

	if !contains(view, "Failed to fetch") {
		t.Error("expected error message in view")
	}

	if !contains(view, "[r] retry") {
		t.Error("expected retry help in view")
	}
}
