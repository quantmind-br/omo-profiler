package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/models"
)

// Messages
type ModelSelectedMsg struct {
	ModelID     string
	DisplayName string
	IsCustom    bool
}

type ModelSelectorCancelMsg struct{}

type PromptSaveCustomMsg struct {
	ModelID string
}

type selectorItem struct {
	isHeader    bool // provider header
	isCustom    bool // "Enter custom model..." option
	isSeparator bool // visual separator line
	provider    string
	model       *models.RegisteredModel
}

type modelSelectorKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Esc   key.Binding
}

func newModelSelectorKeyMap() modelSelectorKeyMap {
	return modelSelectorKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

type ModelSelector struct {
	items       []selectorItem
	cursor      int
	customMode  bool
	customInput textinput.Model
	width       int
	height      int
	keys        modelSelectorKeyMap
	loadError   error
}

func NewModelSelector() ModelSelector {
	customInput := textinput.New()
	customInput.Placeholder = "e.g., gpt-4o-mini"
	customInput.Width = 40

	m := ModelSelector{
		customInput: customInput,
		keys:        newModelSelectorKeyMap(),
	}

	registry, err := models.Load()
	if err != nil {
		m.loadError = err
		// Still show custom option even on error
		m.items = []selectorItem{
			{isSeparator: true},
			{isCustom: true},
		}
		m.cursor = 1 // Point to custom option
		return m
	}

	m.buildItems(registry)
	return m
}

func (m *ModelSelector) buildItems(registry *models.ModelsRegistry) {
	m.items = nil
	groups := registry.ListByProvider()

	for _, group := range groups {
		providerName := group.Provider
		if providerName == "" {
			providerName = "Other"
		}
		// Add provider header
		m.items = append(m.items, selectorItem{
			isHeader: true,
			provider: providerName,
		})
		// Add models under this provider
		for i := range group.Models {
			m.items = append(m.items, selectorItem{
				model: &group.Models[i],
			})
		}
	}

	// Add separator and custom option
	m.items = append(m.items, selectorItem{isSeparator: true})
	m.items = append(m.items, selectorItem{isCustom: true})

	// Set initial cursor to first selectable item
	m.cursor = m.findNextSelectable(0, 1)
}

func (m ModelSelector) Init() tea.Cmd {
	return nil
}

func (m *ModelSelector) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m ModelSelector) isSelectable(idx int) bool {
	if idx < 0 || idx >= len(m.items) {
		return false
	}
	item := m.items[idx]
	return !item.isHeader && !item.isSeparator
}

func (m ModelSelector) findNextSelectable(start, direction int) int {
	idx := start
	for {
		if m.isSelectable(idx) {
			return idx
		}
		idx += direction
		if idx < 0 || idx >= len(m.items) {
			return start // No selectable found, stay put
		}
	}
}

func (m ModelSelector) Update(msg tea.Msg) (ModelSelector, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.customMode {
			switch msg.String() {
			case "enter":
				value := m.customInput.Value()
				if value != "" {
					m.customMode = false
					return m, tea.Batch(
						func() tea.Msg {
							return ModelSelectedMsg{
								ModelID:     value,
								DisplayName: value,
								IsCustom:    true,
							}
						},
						func() tea.Msg {
							return PromptSaveCustomMsg{ModelID: value}
						},
					)
				}
				return m, nil
			case "esc":
				m.customMode = false
				m.customInput.SetValue("")
				return m, nil
			}

			// Update text input
			m.customInput, cmd = m.customInput.Update(msg)
			return m, cmd
		}

		// List mode
		switch {
		case key.Matches(msg, m.keys.Up):
			newCursor := m.cursor - 1
			for newCursor >= 0 && !m.isSelectable(newCursor) {
				newCursor--
			}
			if newCursor >= 0 {
				m.cursor = newCursor
			}

		case key.Matches(msg, m.keys.Down):
			newCursor := m.cursor + 1
			for newCursor < len(m.items) && !m.isSelectable(newCursor) {
				newCursor++
			}
			if newCursor < len(m.items) {
				m.cursor = newCursor
			}

		case key.Matches(msg, m.keys.Enter):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if item.isCustom {
					m.customMode = true
					m.customInput.SetValue("")
					m.customInput.Focus()
					return m, textinput.Blink
				}
				if item.model != nil {
					return m, func() tea.Msg {
						return ModelSelectedMsg{
							ModelID:     item.model.ModelID,
							DisplayName: item.model.DisplayName,
							IsCustom:    false,
						}
					}
				}
			}

		case key.Matches(msg, m.keys.Esc):
			return m, func() tea.Msg {
				return ModelSelectorCancelMsg{}
			}
		}
	}

	return m, nil
}

func (m ModelSelector) View() string {
	if m.customMode {
		return m.renderCustomMode()
	}
	return m.renderList()
}

func (m ModelSelector) renderList() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6C7086"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4")).Background(lipgloss.Color("#7D56F4"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6AC1"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))
	customStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#89B4FA"))

	var lines []string
	lines = append(lines, titleStyle.Render("Select Model"))
	lines = append(lines, "")

	if m.loadError != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
		lines = append(lines, errStyle.Render("Could not load models registry"))
		lines = append(lines, "")
	}

	for i, item := range m.items {
		if item.isHeader {
			lines = append(lines, headerStyle.Render(item.provider))
			continue
		}

		if item.isSeparator {
			lines = append(lines, dimStyle.Render("───────────────────────"))
			continue
		}

		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
			style = selectedStyle
		}

		if item.isCustom {
			text := "Enter custom model..."
			if i == m.cursor {
				text = style.Render(" " + text + " ")
			} else {
				text = customStyle.Render(text)
			}
			lines = append(lines, cursor+text)
			continue
		}

		if item.model != nil {
			displayText := fmt.Sprintf("  %s", item.model.DisplayName)
			if i == m.cursor {
				displayText = style.Render(" " + item.model.DisplayName + " ")
			}
			lines = append(lines, cursor+displayText)
		}
	}

	lines = append(lines, "")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	lines = append(lines, helpStyle.Render("[↑↓] navigate  [Enter] select  [Esc] cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m ModelSelector) renderCustomMode() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	var lines []string
	lines = append(lines, titleStyle.Render("Enter Custom Model"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Model ID:"))
	lines = append(lines, m.customInput.View())
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("[Enter] confirm  [Esc] cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// GetSelectedModel returns the selected model info for display
func (m ModelSelector) GetSelectedModel() (modelID, displayName string, isCustom bool) {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		if item.model != nil {
			return item.model.ModelID, item.model.DisplayName, false
		}
		if item.isCustom && m.customInput.Value() != "" {
			return m.customInput.Value(), m.customInput.Value(), true
		}
	}
	return "", "", false
}
