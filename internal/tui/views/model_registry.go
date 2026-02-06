package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/models"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
	"github.com/sahilm/fuzzy"
)

// Navigation messages
type ModelRegistryBackMsg struct{}
type ModelSavedMsg struct{ Model models.RegisteredModel }
type ModelDeletedMsg struct{ ModelID string }

type modelRegistryKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	New      key.Binding
	Import   key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Enter    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Esc      key.Binding
}

func newModelRegistryKeyMap() modelRegistryKeyMap {
	return modelRegistryKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Import: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "import"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
	}
}

type ModelRegistry struct {
	registry   *models.ModelsRegistry
	groups     []models.ProviderGroup
	flatModels []models.RegisteredModel
	cursor     int
	offset     int
	width      int
	height     int
	keys       modelRegistryKeyMap

	searchInput textinput.Model

	formMode  bool
	editMode  bool
	editingId string

	displayNameInput textinput.Model
	modelIdInput     textinput.Model
	providerInput    textinput.Model
	focusedField     int

	confirmDelete bool
	deleteTarget  string

	errorMsg  string
	loadError error
}

func NewModelRegistry() ModelRegistry {
	keys := newModelRegistryKeyMap()

	searchInput := textinput.New()
	searchInput.Placeholder = "type to filter..."
	searchInput.Width = 30

	displayNameInput := textinput.New()
	displayNameInput.Placeholder = "e.g., Claude Sonnet 4"
	displayNameInput.Width = 40

	modelIdInput := textinput.New()
	modelIdInput.Placeholder = "e.g., claude-sonnet-4-20250514"
	modelIdInput.Width = 40

	providerInput := textinput.New()
	providerInput.Placeholder = "e.g., Anthropic (optional)"
	providerInput.Width = 40

	m := ModelRegistry{
		keys:             keys,
		searchInput:      searchInput,
		displayNameInput: displayNameInput,
		modelIdInput:     modelIdInput,
		providerInput:    providerInput,
	}

	registry, err := models.Load()
	if err != nil {
		m.loadError = err
		return m
	}

	m.registry = registry
	m.groups = registry.ListByProvider()
	m.rebuildFlatModels()

	return m
}

func (m *ModelRegistry) rebuildFlatModels() {
	m.flatModels = nil
	for _, group := range m.groups {
		m.flatModels = append(m.flatModels, group.Models...)
	}
}

func (m ModelRegistry) getFilteredModels() []models.RegisteredModel {
	searchTerm := strings.TrimSpace(m.searchInput.Value())
	if searchTerm == "" {
		return m.flatModels
	}

	searchStrings := make([]string, len(m.flatModels))
	for i, model := range m.flatModels {
		provider := model.Provider
		if provider == "" {
			provider = "Other"
		}
		searchStrings[i] = fmt.Sprintf("%s/%s %s", provider, model.ModelID, model.DisplayName)
	}

	matches := fuzzy.Find(searchTerm, searchStrings)
	if len(matches) == 0 {
		return []models.RegisteredModel{}
	}

	filtered := make([]models.RegisteredModel, len(matches))
	for i, match := range matches {
		filtered[i] = m.flatModels[match.Index]
	}
	return filtered
}

func (m ModelRegistry) Init() tea.Cmd {
	return nil
}

func (m ModelRegistry) Update(msg tea.Msg) (ModelRegistry, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				m.confirmDelete = false
				target := m.deleteTarget
				m.deleteTarget = ""

				if err := m.registry.Delete(target); err != nil {
					m.errorMsg = fmt.Sprintf("Delete failed: %v", err)
					return m, nil
				}

				m.groups = m.registry.ListByProvider()
				m.rebuildFlatModels()
				if m.cursor >= len(m.flatModels) && len(m.flatModels) > 0 {
					m.cursor = len(m.flatModels) - 1
				}

				return m, func() tea.Msg {
					return ModelDeletedMsg{ModelID: target}
				}

			case "n", "N", "esc":
				m.confirmDelete = false
				m.deleteTarget = ""
				return m, nil
			}
			return m, nil
		}

		if m.formMode {
			switch msg.String() {
			case "enter":
				if err := m.validateAndSave(); err != nil {
					m.errorMsg = err.Error()
					return m, nil
				}

				m.formMode = false
				m.editMode = false
				m.editingId = ""
				m.errorMsg = ""
				m.resetForm()

				m.groups = m.registry.ListByProvider()
				m.rebuildFlatModels()

				return m, func() tea.Msg {
					return ModelSavedMsg{}
				}

			case "esc":
				m.formMode = false
				m.editMode = false
				m.editingId = ""
				m.errorMsg = ""
				m.resetForm()
				return m, nil

			case "tab":
				m.focusedField = (m.focusedField + 1) % 3
				m.updateFormFocus()
				m.errorMsg = ""
				return m, nil

			case "shift+tab":
				m.focusedField = (m.focusedField + 2) % 3
				m.updateFormFocus()
				m.errorMsg = ""
				return m, nil
			}

			oldValue := m.getFocusedInputValue()
			m.updateFocusedInput(msg)
			newValue := m.getFocusedInputValue()
			if oldValue != newValue {
				m.errorMsg = ""
			}

			return m, nil
		}

		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.cursor = 0
				m.offset = 0
				return m, nil
			case "enter", "down":
				m.searchInput.Blur()
				return m, nil
			}
			oldValue := m.searchInput.Value()
			m.searchInput, _ = m.searchInput.Update(msg)
			if oldValue != m.searchInput.Value() {
				m.cursor = 0
				m.offset = 0
			}
			return m, nil
		}

		filteredModels := m.getFilteredModels()
		visibleHeight := m.height - 10
		if visibleHeight < 5 {
			visibleHeight = 5
		}

		switch {
		case msg.String() == "/":
			m.searchInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(filteredModels)-1 {
				m.cursor++
				if m.cursor >= m.offset+visibleHeight {
					m.offset = m.cursor - visibleHeight + 1
				}
			}

		case key.Matches(msg, m.keys.New):
			m.enterAddMode()

		case key.Matches(msg, m.keys.Import):
			return m, func() tea.Msg {
				return NavToModelImportMsg{}
			}

		case key.Matches(msg, m.keys.Edit):
			if len(filteredModels) > 0 && m.cursor < len(filteredModels) {
				m.enterEditMode(filteredModels[m.cursor])
			}

		case key.Matches(msg, m.keys.Delete):
			if len(filteredModels) > 0 && m.cursor < len(filteredModels) {
				m.confirmDelete = true
				m.deleteTarget = filteredModels[m.cursor].ModelID
			}

		case key.Matches(msg, m.keys.Esc):
			return m, func() tea.Msg {
				return ModelRegistryBackMsg{}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *ModelRegistry) enterAddMode() {
	m.formMode = true
	m.editMode = false
	m.editingId = ""
	m.errorMsg = ""
	m.resetForm()
	m.focusedField = 0
	m.updateFormFocus()
}

func (m *ModelRegistry) enterEditMode(model models.RegisteredModel) {
	m.formMode = true
	m.editMode = true
	m.editingId = model.ModelID
	m.errorMsg = ""

	m.displayNameInput.SetValue(model.DisplayName)
	m.modelIdInput.SetValue(model.ModelID)
	m.providerInput.SetValue(model.Provider)

	m.focusedField = 0
	m.updateFormFocus()
}

func (m *ModelRegistry) resetForm() {
	m.displayNameInput.SetValue("")
	m.modelIdInput.SetValue("")
	m.providerInput.SetValue("")
}

func (m *ModelRegistry) updateFormFocus() {
	m.displayNameInput.Blur()
	m.modelIdInput.Blur()
	m.providerInput.Blur()

	switch m.focusedField {
	case 0:
		m.displayNameInput.Focus()
	case 1:
		m.modelIdInput.Focus()
	case 2:
		m.providerInput.Focus()
	}
}

func (m *ModelRegistry) updateFocusedInput(msg tea.Msg) {
	switch m.focusedField {
	case 0:
		m.displayNameInput, _ = m.displayNameInput.Update(msg)
	case 1:
		m.modelIdInput, _ = m.modelIdInput.Update(msg)
	case 2:
		m.providerInput, _ = m.providerInput.Update(msg)
	}
}

func (m *ModelRegistry) getFocusedInputValue() string {
	switch m.focusedField {
	case 0:
		return m.displayNameInput.Value()
	case 1:
		return m.modelIdInput.Value()
	case 2:
		return m.providerInput.Value()
	}
	return ""
}

func (m *ModelRegistry) validateAndSave() error {
	displayName := strings.TrimSpace(m.displayNameInput.Value())
	modelId := strings.TrimSpace(m.modelIdInput.Value())
	provider := strings.TrimSpace(m.providerInput.Value())

	if displayName == "" {
		return fmt.Errorf("Display name is required")
	}
	if modelId == "" {
		return fmt.Errorf("Model ID is required")
	}

	newModel := models.RegisteredModel{
		DisplayName: displayName,
		ModelID:     modelId,
		Provider:    provider,
	}

	if m.editMode {
		if err := m.registry.Update(m.editingId, newModel); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("Model with ID '%s' already exists", modelId)
			}
			return err
		}
	} else {
		if err := m.registry.Add(newModel); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("Model with ID '%s' already exists", modelId)
			}
			return err
		}
	}

	return nil
}

func (m ModelRegistry) View() string {
	if m.loadError != nil {
		errorView := lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Manage Models"),
			"",
			errorStyle.Render(fmt.Sprintf("Error loading models: %v", m.loadError)),
			"",
			grayStyle.Render("[esc] back"),
		)
		return errorView
	}

	if m.formMode {
		return m.renderForm()
	}

	if m.confirmDelete {
		return m.renderDeleteConfirm()
	}

	return m.renderList()
}

func (m ModelRegistry) renderList() string {
	title := titleStyle.Render("Manage Models")

	searchLine := "Search: " + m.searchInput.View()

	filteredModels := m.getFilteredModels()

	var content string
	if len(m.flatModels) == 0 {
		content = grayStyle.Render("No models registered yet. Press 'n' to add one.")
	} else if len(filteredModels) == 0 {
		content = grayStyle.Render("No models match the search.")
	} else {
		content = m.renderModelsList(filteredModels)
	}

	help := grayStyle.Render("[/] search  [n] new  [i] import  [e] edit  [d] delete  [esc] back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		searchLine,
		"",
		content,
		"",
		help,
	)
}

func (m ModelRegistry) renderModelsList(filteredModels []models.RegisteredModel) string {
	var lines []string

	visibleHeight := m.height - 10
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	scrollIndicatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	hasMoreAbove := m.offset > 0
	hasMoreBelow := m.offset+visibleHeight < len(filteredModels)

	if hasMoreAbove {
		lines = append(lines, scrollIndicatorStyle.Render("  ↑ more above"))
	}

	endIdx := m.offset + visibleHeight
	if endIdx > len(filteredModels) {
		endIdx = len(filteredModels)
	}

	for i := m.offset; i < endIdx; i++ {
		model := filteredModels[i]
		cursor := "  "
		itemStyle := normalStyle

		if i == m.cursor {
			cursor = accentStyle.Render("> ")
			itemStyle = selectedStyle
		}

		provider := model.Provider
		if provider == "" {
			provider = "Other"
		}

		displayName := fmt.Sprintf("%s/%s", provider, model.ModelID)
		line := fmt.Sprintf("%s%s %s",
			cursor,
			itemStyle.Render(displayName),
			grayStyle.Render("("+model.DisplayName+")"),
		)
		lines = append(lines, line)
	}

	if hasMoreBelow {
		lines = append(lines, scrollIndicatorStyle.Render("  ↓ more below"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m ModelRegistry) renderForm() string {
	var title string
	if m.editMode {
		title = titleStyle.Render("Edit Model")
	} else {
		title = titleStyle.Render("Add New Model")
	}

	formLines := []string{
		"",
		title,
		"",
		fmt.Sprintf("Display Name: %s", m.displayNameInput.View()),
		fmt.Sprintf("Model ID:     %s", m.modelIdInput.View()),
		fmt.Sprintf("Provider:     %s", m.providerInput.View()),
		"",
		grayStyle.Render("[Enter] save  [Esc] cancel"),
	}

	if m.errorMsg != "" {
		formLines = append(formLines, "")
		formLines = append(formLines, errorStyle.Render("Error: "+m.errorMsg))
	}

	return lipgloss.JoinVertical(lipgloss.Left, formLines...)
}

func (m ModelRegistry) renderDeleteConfirm() string {
	content := m.renderList()

	var targetName string
	for _, model := range m.flatModels {
		if model.ModelID == m.deleteTarget {
			targetName = model.DisplayName
			break
		}
	}

	confirmStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F9E2AF")).
		Background(lipgloss.Color("#45475A")).
		Padding(0, 1)

	confirmText := confirmStyle.Render(fmt.Sprintf("Delete '%s'? (y/n)", targetName))

	return lipgloss.JoinVertical(lipgloss.Left, content, "", confirmText)
}

func (m *ModelRegistry) SetSize(width, height int) {
	m.width = width
	m.height = height
	med := layout.MediumFieldWidth(width)
	m.searchInput.Width = med
	m.displayNameInput.Width = med
	m.modelIdInput.Width = med
	m.providerInput.Width = med
}

// IsEditing returns true when text input is active (form fields or search)
func (m ModelRegistry) IsEditing() bool {
	return m.formMode || m.searchInput.Focused()
}
