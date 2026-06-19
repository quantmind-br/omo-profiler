package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg creates a tea.KeyMsg for regular character keys (letters, numbers, symbols)
// Example: keyMsg("a"), keyMsg("1"), keyMsg(" ")
func keyMsg(k string) tea.KeyMsg {
	if len(k) == 0 {
		return tea.KeyMsg{}
	}
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(k),
	}
}

// keyMsgSpecial creates a tea.KeyMsg for special keys (arrows, enter, etc.)
// Example: keyMsgSpecial(tea.KeyRight), keyMsgSpecial(tea.KeyEnter)
func keyMsgSpecial(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{
		Type: k,
	}
}

// TestKeybindingsTestSetup validates that the test infrastructure compiles and runs
func TestKeybindingsTestSetup(t *testing.T) {
	// Verify keyMsg helper works
	msg := keyMsg("a")
	if msg.Type != tea.KeyRunes {
		t.Errorf("expected KeyRunes type, got %v", msg.Type)
	}
	if len(msg.Runes) != 1 || msg.Runes[0] != 'a' {
		t.Errorf("expected rune 'a', got %v", msg.Runes)
	}

	// Verify keyMsgSpecial helper works
	specialMsg := keyMsgSpecial(tea.KeyRight)
	if specialMsg.Type != tea.KeyRight {
		t.Errorf("expected KeyRight type, got %v", specialMsg.Type)
	}

	// Test infrastructure is ready
	t.Log("Test infrastructure setup complete")
}

func TestWizardCategoriesRightExpands(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = false
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))

	if !w.categories[0].expanded {
		t.Error("expected category to be expanded after pressing ctrl+right")
	}
	if !w.inForm {
		t.Error("expected inForm to be true after expanding with ctrl+right")
	}
}

func TestWizardCategoriesLeftCollapses(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = true
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))

	if w.categories[0].expanded {
		t.Error("expected category to be collapsed after pressing ctrl+left")
	}
	if w.inForm {
		t.Error("expected inForm to be false after collapsing with ctrl+left")
	}
}

func TestWizardCategoriesRightDoesNothingWhenExpanded(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = true
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))

	if !w.categories[0].expanded {
		t.Error("category should remain expanded")
	}
}

func TestWizardCategoriesLeftDoesNothingWhenCollapsed(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = false
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))

	if w.categories[0].expanded {
		t.Error("category should remain collapsed")
	}
}

func TestWizardCategoriesCtrlLeftBacksOutOfForm(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = true
	w.inForm = true
	w.focusedField = catFieldName

	// ctrl+← backs out of the form to the category list, mirroring ctrl+→.
	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))
	if w.categories[0].expanded || w.inForm {
		t.Error("ctrl+left should back out of the form to the category list")
	}
}

func TestWizardCategoriesPlainLeftStaysInForm(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	newCat.variant.SetValue("abc")
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = true
	w.inForm = true
	w.focusedField = catFieldVariant

	// Plain ← stays in the form: it is reserved for cursor movement inside the
	// focused text input, not for backing out.
	w, _ = w.Update(keyMsgSpecial(tea.KeyLeft))
	if !w.categories[0].expanded || !w.inForm {
		t.Error("plain left arrow should stay in the form for text-cursor navigation")
	}
}

func TestWizardCategoriesRightIgnoredInFormMode(t *testing.T) {
	w := NewWizardCategories()
	newCat := newCategoryConfig()
	w.categories = append(w.categories, &newCat)
	w.cursor = 0
	w.categories[0].expanded = false
	w.inForm = true

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))
	if w.categories[0].expanded {
		t.Error("ctrl+right should be ignored in form mode")
	}
}

func TestWizardAgentsRightExpands(t *testing.T) {
	w := NewWizardAgents()
	w.cursor = 0
	agentName := allAgents[0]
	w.agents[agentName].enabled = true
	w.agents[agentName].expanded = false
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))

	if !w.agents[agentName].expanded {
		t.Error("expected agent to be expanded after pressing ctrl+right")
	}
	if !w.inForm {
		t.Error("expected inForm to be true after expanding")
	}
}

func TestWizardAgentsLeftCollapses(t *testing.T) {
	w := NewWizardAgents()
	w.cursor = 0
	agentName := allAgents[0]
	w.agents[agentName].enabled = true
	w.agents[agentName].expanded = true
	w.inForm = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))

	if w.agents[agentName].expanded {
		t.Error("expected agent to be collapsed after pressing ctrl+left")
	}
	if w.inForm {
		t.Error("expected inForm to be false after collapsing")
	}
}

func TestWizardAgentsCtrlLeftBacksOutOfForm(t *testing.T) {
	w := NewWizardAgents()
	w.cursor = 0
	agentName := allAgents[0]
	w.agents[agentName].enabled = true
	w.agents[agentName].expanded = true
	w.inForm = true
	w.focusedField = fieldModel

	// ctrl+← backs out of the agent form to the agent list.
	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))
	if w.agents[agentName].expanded || w.inForm {
		t.Error("ctrl+left should back out of the agent form to the list")
	}
}

func TestWizardAgentsSpaceToggles(t *testing.T) {
	w := NewWizardAgents()
	w.cursor = 0
	agentName := allAgents[0]
	w.agents[agentName].enabled = false
	w.inForm = false

	w, _ = w.Update(keyMsg(" "))

	if !w.agents[agentName].enabled {
		t.Error("expected agent to be enabled after pressing space")
	}

	w, _ = w.Update(keyMsg(" "))

	if w.agents[agentName].enabled {
		t.Error("expected agent to be disabled after pressing space again")
	}
}

func TestWizardOtherRightExpandsSection(t *testing.T) {
	w := NewWizardOther()
	w.SetSize(80, 24)
	w.currentCategory = categoryDisabledFeatures
	w.categoryExpanded[categoryDisabledFeatures] = true
	w.inCategory = true
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = false
	w.inSubSection = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))

	if !w.sectionExpanded[sectionDisabledMcps] {
		t.Error("expected section to be expanded after pressing ctrl+right")
	}
	if !w.inSubSection {
		t.Error("expected inSubSection to be true after expanding with ctrl+right")
	}
	if w.subCursor != 0 {
		t.Errorf("expected subCursor to be 0, got %d", w.subCursor)
	}
}

func TestWizardOtherLeftCollapsesSection(t *testing.T) {
	w := NewWizardOther()
	w.SetSize(80, 24)
	w.currentCategory = categoryDisabledFeatures
	w.categoryExpanded[categoryDisabledFeatures] = true
	w.inCategory = true
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = true
	w.inSubSection = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))

	if w.sectionExpanded[sectionDisabledMcps] {
		t.Error("expected section to be collapsed after pressing ctrl+left")
	}
}

func TestWizardOtherRightDoesNothingWhenExpanded(t *testing.T) {
	w := NewWizardOther()
	w.SetSize(80, 24)
	w.currentCategory = categoryDisabledFeatures
	w.categoryExpanded[categoryDisabledFeatures] = true
	w.inCategory = true
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = true
	w.inSubSection = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlRight))

	if !w.sectionExpanded[sectionDisabledMcps] {
		t.Error("section should remain expanded")
	}
}

func TestWizardOtherLeftDoesNothingWhenCollapsed(t *testing.T) {
	w := NewWizardOther()
	w.SetSize(80, 24)
	w.currentCategory = categoryDisabledFeatures
	w.categoryExpanded[categoryDisabledFeatures] = true
	w.inCategory = true
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = false
	w.inSubSection = false

	w, _ = w.Update(keyMsgSpecial(tea.KeyCtrlLeft))

	// ctrl+left on a collapsed section goes back to category header
	if w.inCategory {
		t.Error("expected to go back to category header")
	}
}

func TestWizardOtherLeftRightIgnoredInSubSection(t *testing.T) {
	w := NewWizardOther()
	w.SetSize(80, 24)
	w.currentCategory = categoryDisabledFeatures
	w.categoryExpanded[categoryDisabledFeatures] = true
	w.inCategory = true
	w.currentSection = sectionDisabledAgents
	w.sectionExpanded[sectionDisabledAgents] = true
	w.inSubSection = true
	w.subCursor = 0

	w, _ = w.Update(keyMsgSpecial(tea.KeyLeft))
	if !w.sectionExpanded[sectionDisabledAgents] {
		t.Error("left arrow should be ignored when in subsection")
	}

	w.sectionExpanded[sectionDisabledAgents] = false
	w, _ = w.Update(keyMsgSpecial(tea.KeyRight))
	if w.sectionExpanded[sectionDisabledAgents] {
		t.Error("right arrow should be ignored when in subsection")
	}
}

func TestWizardHooksSpaceToggles(t *testing.T) {
	w := NewWizardHooks()
	w.SetSize(80, 24)
	w.cursor = 0
	hook := allHooks[0]
	initialState := w.disabled[hook]

	w, _ = w.Update(keyMsg(" "))

	if w.disabled[hook] == initialState {
		t.Error("expected hook to toggle after pressing space")
	}
}

func TestModelRegistryIsEditingInFormMode(t *testing.T) {
	m := NewModelRegistry()

	if m.IsEditing() {
		t.Error("expected IsEditing() to be false initially")
	}

	m.enterAddMode()
	if !m.IsEditing() {
		t.Error("expected IsEditing() to be true in form mode")
	}

	m.formMode = false
	m.resetForm()
	if m.IsEditing() {
		t.Error("expected IsEditing() to be false after exiting form mode")
	}
}

func TestModelRegistryIsEditingWithSearchFocused(t *testing.T) {
	m := NewModelRegistry()

	if m.IsEditing() {
		t.Error("expected IsEditing() to be false initially")
	}

	m.searchInput.Focus()
	if !m.IsEditing() {
		t.Error("expected IsEditing() to be true when search is focused")
	}

	m.searchInput.Blur()
	if m.IsEditing() {
		t.Error("expected IsEditing() to be false after search is blurred")
	}
}

func TestModelRegistryQKeyPassesThroughInFormMode(t *testing.T) {
	m := NewModelRegistry()
	m.enterAddMode()

	m, _ = m.Update(keyMsg("q"))

	if m.displayNameInput.Value() != "q" {
		t.Errorf("expected 'q' in displayNameInput, got %q", m.displayNameInput.Value())
	}

	if !m.formMode {
		t.Error("expected to remain in form mode after typing 'q'")
	}
}

func TestModelImportIsEditingWithSearchFocused(t *testing.T) {
	m := NewModelImport()

	if m.IsEditing() {
		t.Error("expected IsEditing() to be false initially")
	}

	m.searchInput.Focus()
	if !m.IsEditing() {
		t.Error("expected IsEditing() to be true when search is focused")
	}
}

func TestWizardHooksEnterDoesNotToggle(t *testing.T) {
	w := NewWizardHooks()
	w.SetSize(80, 24)
	w.cursor = 0
	hook := allHooks[0]
	initialState := w.disabled[hook]

	w, _ = w.Update(keyMsgSpecial(tea.KeyEnter))

	// Enter should NOT toggle - it may trigger next step or do nothing
	// But we check that disabled state is unchanged
	if w.disabled[hook] != initialState {
		t.Error("expected hook to NOT toggle after pressing enter (enter is not for toggle)")
	}
}
