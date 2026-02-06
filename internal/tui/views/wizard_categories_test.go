package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
)

func TestNewWizardCategories(t *testing.T) {
	wc := NewWizardCategories()

	if len(wc.categories) != 0 {
		t.Errorf("expected empty categories, got %d", len(wc.categories))
	}

	if wc.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", wc.cursor)
	}

	if wc.inForm {
		t.Error("expected inForm to be false initially")
	}

	if wc.ready {
		t.Error("expected ready to be false initially")
	}

	// Check key bindings
	if wc.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if wc.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if wc.keys.New.Help().Key == "" {
		t.Error("expected New key to be initialized")
	}

	if wc.keys.Delete.Help().Key == "" {
		t.Error("expected Delete key to be initialized")
	}

	if wc.keys.Expand.Help().Key == "" {
		t.Error("expected Expand key to be initialized")
	}

	if wc.keys.Next.Help().Key == "" {
		t.Error("expected Next key to be initialized")
	}

	if wc.keys.Back.Help().Key == "" {
		t.Error("expected Back key to be initialized")
	}
}

func TestWizardCategoriesInit(t *testing.T) {
	wc := NewWizardCategories()
	cmd := wc.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestWizardCategoriesSetSize(t *testing.T) {
	wc := NewWizardCategories()

	wc.SetSize(100, 50)

	if wc.width != 100 {
		t.Errorf("expected width 100, got %d", wc.width)
	}

	if wc.height != 50 {
		t.Errorf("expected height 50, got %d", wc.height)
	}

	if !wc.ready {
		t.Error("expected ready to be true after SetSize")
	}

	// Call SetSize again to test the ready=true path
	wc.SetSize(80, 40)

	if wc.width != 80 {
		t.Errorf("expected width 80, got %d", wc.width)
	}

	if wc.height != 40 {
		t.Errorf("expected height 40, got %d", wc.height)
	}
}

func TestWizardCategoriesSetConfig(t *testing.T) {
	wc := NewWizardCategories()

	cfg := &config.Config{
		Categories: map[string]*config.CategoryConfig{
			"coding": {
				Model:       "claude-sonnet-4",
				Description: "Coding tasks",
			},
			"writing": {
				Model:       "gpt-4",
				Description: "Writing tasks",
			},
		},
	}

	wc.SetConfig(cfg)

	if len(wc.categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(wc.categories))
	}

	// Check that both categories exist with correct data
	foundCoding := false
	foundWriting := false
	for _, cat := range wc.categories {
		if cat.name == "coding" {
			foundCoding = true
			if cat.modelValue != "claude-sonnet-4" {
				t.Errorf("expected coding model 'claude-sonnet-4', got %q", cat.modelValue)
			}
			if cat.description.Value() != "Coding tasks" {
				t.Errorf("expected coding description 'Coding tasks', got %q", cat.description.Value())
			}
		}
		if cat.name == "writing" {
			foundWriting = true
			if cat.modelValue != "gpt-4" {
				t.Errorf("expected writing model 'gpt-4', got %q", cat.modelValue)
			}
			if cat.description.Value() != "Writing tasks" {
				t.Errorf("expected writing description 'Writing tasks', got %q", cat.description.Value())
			}
		}
	}

	if !foundCoding {
		t.Error("expected to find 'coding' category")
	}

	if !foundWriting {
		t.Error("expected to find 'writing' category")
	}
}

func TestWizardCategoriesSetConfigNil(t *testing.T) {
	wc := NewWizardCategories()

	// Pass a config with nil Categories (not nil config itself)
	cfg := &config.Config{}
	wc.SetConfig(cfg)

	if len(wc.categories) != 0 {
		t.Errorf("expected 0 categories for nil Categories, got %d", len(wc.categories))
	}
}

func TestWizardCategoriesSetConfigNilCategories(t *testing.T) {
	wc := NewWizardCategories()

	cfg := &config.Config{}
	wc.SetConfig(cfg)

	if len(wc.categories) != 0 {
		t.Errorf("expected 0 categories for nil Categories, got %d", len(wc.categories))
	}
}

func TestWizardCategoriesSetConfigWithTemperature(t *testing.T) {
	wc := NewWizardCategories()

	temp := 0.7
	cfg := &config.Config{
		Categories: map[string]*config.CategoryConfig{
			"test": {
				Temperature: &temp,
			},
		},
	}

	wc.SetConfig(cfg)

	if wc.categories[0].temperature.Value() != "0.7" {
		t.Errorf("expected temperature '0.7', got %q", wc.categories[0].temperature.Value())
	}
}

func TestWizardCategoriesSetConfigWithThinking(t *testing.T) {
	wc := NewWizardCategories()

	cfg := &config.Config{
		Categories: map[string]*config.CategoryConfig{
			"test": {
				Thinking: &config.ThinkingConfig{
					Type: "enabled",
				},
			},
		},
	}

	wc.SetConfig(cfg)

	if wc.categories[0].thinkingTypeIdx != 1 { // "enabled" is at index 1
		t.Errorf("expected thinkingTypeIdx 1, got %d", wc.categories[0].thinkingTypeIdx)
	}
}

func TestWizardCategoriesApply(t *testing.T) {
	wc := NewWizardCategories()

	// Add a category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("test-category")
	newCat.modelValue = "claude-sonnet-4"
	newCat.description.SetValue("Test description")
	wc.categories = append(wc.categories, &newCat)

	cfg := &config.Config{}
	wc.Apply(cfg)

	if cfg.Categories == nil {
		t.Fatal("expected Categories to be set")
	}

	if len(cfg.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(cfg.Categories))
	}

	catCfg, ok := cfg.Categories["test-category"]
	if !ok {
		t.Fatal("expected category 'test-category' to exist")
	}

	if catCfg.Model != "claude-sonnet-4" {
		t.Errorf("expected model 'claude-sonnet-4', got %q", catCfg.Model)
	}

	if catCfg.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %q", catCfg.Description)
	}
}

func TestWizardCategoriesApplyEmptyName(t *testing.T) {
	wc := NewWizardCategories()

	// Add a category with empty name
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("")
	newCat.modelValue = "claude-sonnet-4"
	wc.categories = append(wc.categories, &newCat)

	cfg := &config.Config{}
	wc.Apply(cfg)

	// Should skip categories with empty names
	if len(cfg.Categories) != 0 {
		t.Errorf("expected empty Categories map, got %d entries", len(cfg.Categories))
	}
}

func TestWizardCategoriesApplyWithTemperature(t *testing.T) {
	wc := NewWizardCategories()

	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("test")
	newCat.temperature.SetValue("0.8")
	wc.categories = append(wc.categories, &newCat)

	cfg := &config.Config{}
	wc.Apply(cfg)

	catCfg := cfg.Categories["test"]
	if catCfg.Temperature == nil {
		t.Error("expected Temperature to be set")
	} else if *catCfg.Temperature != 0.8 {
		t.Errorf("expected temperature 0.8, got %f", *catCfg.Temperature)
	}
}

func TestWizardCategoriesApplyWithIsUnstable(t *testing.T) {
	wc := NewWizardCategories()

	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("test")
	newCat.isUnstable = true
	wc.categories = append(wc.categories, &newCat)

	cfg := &config.Config{}
	wc.Apply(cfg)

	catCfg := cfg.Categories["test"]
	if catCfg.IsUnstableAgent == nil {
		t.Error("expected IsUnstableAgent to be set")
	} else if !*catCfg.IsUnstableAgent {
		t.Error("expected IsUnstableAgent to be true")
	}
}

func TestWizardCategoriesUpdateWindowSizeMsg(t *testing.T) {
	wc := NewWizardCategories()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := wc.Update(msg)

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

func TestWizardCategoriesUpdateUpKey(t *testing.T) {
	wc := NewWizardCategories()

	// Add some categories
	newCat1 := newCategoryConfig()
	newCat1.nameInput.SetValue("cat1")
	newCat2 := newCategoryConfig()
	newCat2.nameInput.SetValue("cat2")
	wc.categories = append(wc.categories, &newCat1, &newCat2)
	wc.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Up key")
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", updated.cursor)
	}

	// Test at top
	wc.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = wc.Update(msg)

	if updated.cursor != 0 {
		t.Errorf("expected cursor to remain 0 at top, got %d", updated.cursor)
	}
}

func TestWizardCategoriesUpdateDownKey(t *testing.T) {
	wc := NewWizardCategories()

	// Add some categories
	newCat1 := newCategoryConfig()
	newCat1.nameInput.SetValue("cat1")
	newCat2 := newCategoryConfig()
	newCat2.nameInput.SetValue("cat2")
	wc.categories = append(wc.categories, &newCat1, &newCat2)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Down key")
	}

	if updated.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", updated.cursor)
	}
}

func TestWizardCategoriesUpdateNewKey(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'n' key")
	}

	if len(updated.categories) != 1 {
		t.Errorf("expected 1 category after 'n', got %d", len(updated.categories))
	}

	if !updated.inForm {
		t.Error("expected inForm to be true after creating new category")
	}

	if !updated.categories[0].expanded {
		t.Error("expected category to be expanded after creation")
	}

	if updated.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", updated.cursor)
	}
}

func TestWizardCategoriesUpdateDeleteKey(t *testing.T) {
	wc := NewWizardCategories()

	// Add some categories
	newCat1 := newCategoryConfig()
	newCat1.nameInput.SetValue("cat1")
	newCat2 := newCategoryConfig()
	newCat2.nameInput.SetValue("cat2")
	wc.categories = append(wc.categories, &newCat1, &newCat2)
	wc.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'd' key")
	}

	if len(updated.categories) != 1 {
		t.Errorf("expected 1 category after delete, got %d", len(updated.categories))
	}

	if updated.categories[0].nameInput.Value() != "cat2" {
		t.Errorf("expected remaining category 'cat2', got %q", updated.categories[0].nameInput.Value())
	}
}

func TestWizardCategoriesUpdateDeleteLastCategory(t *testing.T) {
	wc := NewWizardCategories()

	// Add one category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	wc.categories = append(wc.categories, &newCat)
	wc.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'd' key")
	}

	if len(updated.categories) != 0 {
		t.Errorf("expected 0 categories after deleting last one, got %d", len(updated.categories))
	}

	if updated.inForm {
		t.Error("expected inForm to be false when no categories")
	}
}

func TestWizardCategoriesUpdateExpandKey(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add a category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	wc.categories = append(wc.categories, &newCat)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter key")
	}

	if !updated.categories[0].expanded {
		t.Error("expected category to be expanded after Enter")
	}

	if !updated.inForm {
		t.Error("expected inForm to be true after expansion")
	}
}

func TestWizardCategoriesUpdateExpandKeyCollapse(t *testing.T) {
	wc := NewWizardCategories()

	// Add an expanded category - but NOT in form mode
	// The Expand key only works when not in form mode
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	// Don't set inForm - let the expand toggle work

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter key")
	}

	if updated.categories[0].expanded {
		t.Error("expected category to be collapsed after Enter on expanded")
	}

	if updated.inForm {
		t.Error("expected inForm to be false after collapse")
	}
}

func TestWizardCategoriesUpdateRightKey(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add a category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	wc.categories = append(wc.categories, &newCat)

	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Right key")
	}

	if !updated.categories[0].expanded {
		t.Error("expected category to be expanded after Right")
	}

	if !updated.inForm {
		t.Error("expected inForm to be true after Right")
	}
}

func TestWizardCategoriesUpdateLeftKey(t *testing.T) {
	wc := NewWizardCategories()

	// Add an expanded category - but NOT in form mode
	// The Left key only works when not in form mode
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	// Don't set inForm

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Left key")
	}

	if updated.categories[0].expanded {
		t.Error("expected category to be collapsed after Left")
	}

	if updated.inForm {
		t.Error("expected inForm to be false after collapse")
	}
}

func TestWizardCategoriesUpdateNextKey(t *testing.T) {
	wc := NewWizardCategories()

	msg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := wc.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Tab key")
	}

	result := cmd()
	if _, ok := result.(WizardNextMsg); !ok {
		t.Errorf("expected WizardNextMsg, got %T", result)
	}
}

func TestWizardCategoriesUpdateBackKey(t *testing.T) {
	wc := NewWizardCategories()

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	_, cmd := wc.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Shift+Tab key")
	}

	result := cmd()
	if _, ok := result.(WizardBackMsg); !ok {
		t.Errorf("expected WizardBackMsg, got %T", result)
	}
}

func TestWizardCategoriesUpdateFormEsc(t *testing.T) {
	wc := NewWizardCategories()

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in form mode")
	}

	if updated.inForm {
		t.Error("expected inForm to be false after Esc")
	}

	if updated.categories[0].expanded {
		t.Error("expected category to be collapsed after Esc")
	}
}

func TestWizardCategoriesUpdateFormNavigateDown(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.focusedField = catFieldName

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'j' key")
	}

	if updated.focusedField != catFieldModel {
		t.Errorf("expected focusedField to be catFieldModel, got %v", updated.focusedField)
	}
}

func TestWizardCategoriesUpdateFormNavigateUp(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.focusedField = catFieldModel

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'k' key")
	}

	if updated.focusedField != catFieldName {
		t.Errorf("expected focusedField to be catFieldName, got %v", updated.focusedField)
	}
}

func TestWizardCategoriesUpdateFormToggleIsUnstable(t *testing.T) {
	wc := NewWizardCategories()

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.focusedField = catFieldIsUnstable
	newCat.isUnstable = false

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter on is_unstable")
	}

	if !updated.categories[0].isUnstable {
		t.Error("expected isUnstable to be toggled to true")
	}
}

func TestWizardCategoriesUpdateFormCycleThinkingType(t *testing.T) {
	wc := NewWizardCategories()

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.focusedField = catFieldThinkingType
	newCat.thinkingTypeIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter on thinking_type")
	}

	if updated.categories[0].thinkingTypeIdx != 1 {
		t.Errorf("expected thinkingTypeIdx to be 1, got %d", updated.categories[0].thinkingTypeIdx)
	}
}

func TestWizardCategoriesUpdateModelSelectedMsg(t *testing.T) {
	wc := NewWizardCategories()

	// Add a category in model selection mode
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.selectingModel = true
	wc.categories = append(wc.categories, &newCat)

	msg := ModelSelectedMsg{
		ModelID:     "claude-sonnet-4",
		DisplayName: "Claude Sonnet 4",
	}

	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for ModelSelectedMsg")
	}

	if updated.categories[0].modelValue != "claude-sonnet-4" {
		t.Errorf("expected modelValue 'claude-sonnet-4', got %q", updated.categories[0].modelValue)
	}

	if updated.categories[0].modelDisplay != "Claude Sonnet 4" {
		t.Errorf("expected modelDisplay 'Claude Sonnet 4', got %q", updated.categories[0].modelDisplay)
	}

	if updated.categories[0].selectingModel {
		t.Error("expected selectingModel to be false after selection")
	}
}

func TestWizardCategoriesUpdateModelSelectorCancelMsg(t *testing.T) {
	wc := NewWizardCategories()

	// Add a category in model selection mode
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.selectingModel = true
	wc.categories = append(wc.categories, &newCat)

	msg := ModelSelectorCancelMsg{}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for ModelSelectorCancelMsg")
	}

	if updated.categories[0].selectingModel {
		t.Error("expected selectingModel to be false after cancel")
	}
}

func TestWizardCategoriesViewNoCategories(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	view := wc.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Configure Categories") {
		t.Error("expected 'Configure Categories' in view")
	}

	// The viewport content needs to be set
	wc.viewport.SetContent(wc.renderContent())
	view = wc.View()

	if !contains(view, "No categories defined") {
		t.Error("expected 'No categories defined' in view")
	}
}

func TestWizardCategoriesViewWithCategories(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add categories
	newCat1 := newCategoryConfig()
	newCat1.nameInput.SetValue("coding")
	newCat2 := newCategoryConfig()
	newCat2.nameInput.SetValue("writing")
	wc.categories = append(wc.categories, &newCat1, &newCat2)
	wc.viewport.SetContent(wc.renderContent())

	view := wc.View()

	if !contains(view, "coding") {
		t.Error("expected 'coding' in view")
	}

	if !contains(view, "writing") {
		t.Error("expected 'writing' in view")
	}

	if !contains(view, "Configure Categories") {
		t.Error("expected 'Configure Categories' in view")
	}
}

func TestWizardCategoriesViewExpanded(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("coding")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.viewport.SetContent(wc.renderContent())

	view := wc.View()

	if !contains(view, "name") {
		t.Error("expected 'name' field in expanded view")
	}

	if !contains(view, "model") {
		t.Error("expected 'model' field in expanded view")
	}

	if !contains(view, "Esc: close form") {
		t.Error("expected close form help in expanded view")
	}
}

func TestWizardCategoriesViewSelectingModel(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add a category in model selection mode
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("coding")
	newCat.selectingModel = true
	newCat.modelSelector = NewModelSelector()
	wc.categories = append(wc.categories, &newCat)

	view := wc.View()

	// Should show model selector view
	if !contains(view, "Select Model") {
		t.Error("expected 'Select Model' in view")
	}
}

func TestNewCategoryConfig(t *testing.T) {
	cc := newCategoryConfig()

	if cc.nameInput.Width != 30 {
		t.Errorf("expected nameInput width 30, got %d", cc.nameInput.Width)
	}

	if cc.temperature.Width != 10 {
		t.Errorf("expected temperature width 10, got %d", cc.temperature.Width)
	}

	if cc.tools.Width != 40 {
		t.Errorf("expected tools width 40, got %d", cc.tools.Width)
	}

	if cc.promptAppend.Height() != 3 {
		t.Errorf("expected promptAppend height 3, got %d", cc.promptAppend.Height())
	}
}

func TestWizardCategoriesRenderSaveCustomPrompt(t *testing.T) {
	wc := NewWizardCategories()

	// Create a category in save custom model mode
	cc := newCategoryConfig()
	cc.customModelToSave = "custom-model-123"
	cc.savePromptAnswer = ""

	view := wc.renderSaveCustomPrompt(&cc)

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Custom Model") {
		t.Error("expected 'Custom Model' in view")
	}

	if !contains(view, "custom-model-123") {
		t.Error("expected model ID in view")
	}

	if !contains(view, "Save this model") {
		t.Error("expected save prompt in view")
	}
}

func TestWizardCategoriesHandleSaveCustomModelYes(t *testing.T) {
	wc := NewWizardCategories()

	cc := newCategoryConfig()
	cc.nameInput.SetValue("cat1")
	cc.customModelToSave = "custom-model-123"
	cc.savePromptAnswer = ""
	cc.savingCustomModel = true
	wc.categories = append(wc.categories, &cc)
	wc.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	updated, cmd := wc.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil command for 'y' key (textinput.Blink)")
	}

	if !updated.categories[0].savingCustomModel {
		t.Error("expected savingCustomModel to remain true")
	}

	if updated.categories[0].savePromptAnswer != "y" {
		t.Errorf("expected savePromptAnswer 'y', got %q", updated.categories[0].savePromptAnswer)
	}
}

func TestWizardCategoriesHandleSaveCustomModelNo(t *testing.T) {
	wc := NewWizardCategories()

	cc := newCategoryConfig()
	cc.customModelToSave = "custom-model-123"
	cc.savingCustomModel = true
	wc.categories = append(wc.categories, &cc)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updated, cmd := wc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for 'n' key")
	}

	if updated.categories[0].savingCustomModel {
		t.Error("expected savingCustomModel to be false after 'n'")
	}

	if updated.categories[0].savePromptAnswer != "" {
		t.Error("expected savePromptAnswer to be cleared")
	}
}

func TestWizardCategoriesGetLineForField(t *testing.T) {
	wc := NewWizardCategories()

	// Add some categories
	newCat1 := newCategoryConfig()
	newCat1.nameInput.SetValue("cat1")
	newCat2 := newCategoryConfig()
	newCat2.nameInput.SetValue("cat2")
	newCat2.expanded = true
	wc.categories = append(wc.categories, &newCat1, &newCat2)
	wc.cursor = 1

	line := wc.getLineForField(catFieldName)

	// First category (collapsed) = 1 line
	// Second category header = 1 line
	// Empty line = 1 line
	// Field offset = 0
	// Expected: 1 + 1 + 1 + 0 = 3
	if line != 3 {
		t.Errorf("expected line 3, got %d", line)
	}
}

func TestWizardCategoriesEnsureFieldVisible(t *testing.T) {
	wc := NewWizardCategories()
	wc.SetSize(80, 24)

	// Add an expanded category
	newCat := newCategoryConfig()
	newCat.nameInput.SetValue("cat1")
	newCat.expanded = true
	wc.categories = append(wc.categories, &newCat)
	wc.inForm = true
	wc.focusedField = catFieldName

	// Should not panic
	wc.ensureFieldVisible()

	// Just verify the method runs without error
	if !wc.ready {
		t.Error("expected ready to be true after SetSize")
	}
}
