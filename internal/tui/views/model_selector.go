package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/models"
	"github.com/sahilm/fuzzy"
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
	items         []selectorItem
	groups        []models.ProviderGroup
	cursor        int
	scrollOffset  int
	customMode    bool
	customInput   textinput.Model
	searchInput   textinput.Model
	filteredCount int
	width         int
	height        int
	keys          modelSelectorKeyMap
	loadError     error
}

func NewModelSelector() ModelSelector {
	customInput := textinput.New()
	customInput.Placeholder = "e.g., gpt-4o-mini"
	customInput.Width = 40

	searchInput := textinput.New()
	searchInput.Placeholder = "Search models..."
	searchInput.Width = 40
	searchInput.Focus()

	m := ModelSelector{
		customInput: customInput,
		searchInput: searchInput,
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
		m.filteredCount = 0
		return m
	}

	m.buildItems(registry)
	return m
}

func (m *ModelSelector) buildItems(registry *models.ModelsRegistry) {
	m.groups = registry.ListByProvider()
	m.rebuildItems()
}

func (m *ModelSelector) rebuildItems() {
	m.items = nil
	m.filteredCount = 0
	searchTerm := strings.TrimSpace(m.searchInput.Value())

	for _, group := range m.groups {
		providerName := group.Provider
		if providerName == "" {
			providerName = "Other"
		}

		if searchTerm == "" {
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
			m.filteredCount += len(group.Models)
			continue
		}

		searchStrings := make([]string, len(group.Models))
		for i, model := range group.Models {
			provider := model.Provider
			if provider == "" {
				provider = "Other"
			}
			searchStrings[i] = fmt.Sprintf("%s/%s %s", provider, model.ModelID, model.DisplayName)
		}

		matches := fuzzy.Find(searchTerm, searchStrings)
		if len(matches) == 0 {
			continue
		}

		m.items = append(m.items, selectorItem{
			isHeader: true,
			provider: providerName,
		})
		for _, match := range matches {
			idx := match.Index
			if idx < 0 || idx >= len(group.Models) {
				continue
			}
			m.items = append(m.items, selectorItem{
				model: &group.Models[idx],
			})
			m.filteredCount++
		}
	}

	// Add separator and custom option
	m.items = append(m.items, selectorItem{isSeparator: true})
	m.items = append(m.items, selectorItem{isCustom: true})

	// Set initial cursor to first selectable item
	m.cursor = m.findNextSelectable(0, 1)
	m.scrollOffset = 0
}

func (m ModelSelector) Init() tea.Cmd {
	return nil
}

func (m *ModelSelector) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ensureCursorVisible()
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

func (m ModelSelector) listHeight() int {
	if len(m.items) == 0 {
		return 0
	}

	headerHeight := m.headerHeight()

	footerHeight := 2
	available := m.height - headerHeight - footerHeight
	if available < 1 {
		available = 1
	}
	if available > len(m.items) {
		available = len(m.items)
	}

	return available
}

func (m ModelSelector) headerHeight() int {
	height := 2 // title + empty line
	if m.loadError != nil {
		height += 2
	}
	height += 2 // search line + empty line
	if m.filteredCount == 0 && strings.TrimSpace(m.searchInput.Value()) != "" {
		height += 2
	}
	return height
}

func (m *ModelSelector) ensureCursorVisible() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}

	visible := m.listHeight()
	if visible == 0 {
		m.scrollOffset = 0
		return
	}

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}

	maxOffset := len(m.items) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

func (m ModelSelector) Update(msg tea.Msg) (ModelSelector, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

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

		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc":
				if strings.TrimSpace(m.searchInput.Value()) == "" {
					m.searchInput.Blur()
					return m, func() tea.Msg {
						return ModelSelectorCancelMsg{}
					}
				}
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.rebuildItems()
				return m, nil
			case "enter", "down":
				m.searchInput.Blur()
				return m, nil
			}

			oldValue := m.searchInput.Value()
			m.searchInput, cmd = m.searchInput.Update(msg)
			if oldValue != m.searchInput.Value() {
				m.rebuildItems()
			}
			return m, cmd
		}

		// List mode
		switch {
		case msg.String() == "/":
			m.searchInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			newCursor := m.cursor - 1
			for newCursor >= 0 && !m.isSelectable(newCursor) {
				newCursor--
			}
			if newCursor >= 0 {
				m.cursor = newCursor
				m.ensureCursorVisible()
			}

		case key.Matches(msg, m.keys.Down):
			newCursor := m.cursor + 1
			for newCursor < len(m.items) && !m.isSelectable(newCursor) {
				newCursor++
			}
			if newCursor < len(m.items) {
				m.cursor = newCursor
				m.ensureCursorVisible()
			}

		case key.Matches(msg, m.keys.Enter):
			if m.cursor >= 0 && m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if item.isCustom {
					m.customMode = true
					m.customInput.SetValue("")
					m.customInput.Focus()
					m.searchInput.Blur()
					return m, textinput.Blink
				}
				if item.model != nil {
					modelValue := item.model.ModelID
					if item.model.Provider != "" {
						modelValue = item.model.Provider + "/" + item.model.ModelID
					}
					return m, func() tea.Msg {
						return ModelSelectedMsg{
							ModelID:     modelValue,
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

	var headerLines []string
	headerLines = append(headerLines, titleStyle.Render("Select Model"))
	headerLines = append(headerLines, "")

	if m.loadError != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
		headerLines = append(headerLines, errStyle.Render("Could not load models registry"))
		headerLines = append(headerLines, "")
	}

	searchLine := "Search: " + m.searchInput.View()
	headerLines = append(headerLines, searchLine)
	headerLines = append(headerLines, "")

	if m.filteredCount == 0 && strings.TrimSpace(m.searchInput.Value()) != "" {
		headerLines = append(headerLines, dimStyle.Render("No models match the search."))
		headerLines = append(headerLines, "")
	}

	visible := m.listHeight()
	start := 0
	end := 0
	if visible > 0 {
		start = m.scrollOffset
		if start < 0 {
			start = 0
		}
		end = start + visible
		if end > len(m.items) {
			end = len(m.items)
		}
	}

	var listLines []string
	for i := start; i < end; i++ {
		item := m.items[i]
		if item.isHeader {
			listLines = append(listLines, headerStyle.Render(item.provider))
			continue
		}

		if item.isSeparator {
			listLines = append(listLines, dimStyle.Render("───────────────────────"))
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
			listLines = append(listLines, cursor+text)
			continue
		}

		if item.model != nil {
			displayText := fmt.Sprintf("  %s", item.model.DisplayName)
			if i == m.cursor {
				displayText = style.Render(" " + item.model.DisplayName + " ")
			}
			listLines = append(listLines, cursor+displayText)
		}
	}

	var footerLines []string
	footerLines = append(footerLines, "")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	footerLines = append(footerLines, helpStyle.Render("[/] search  [↑↓] navigate  [Enter] select  [Esc] cancel"))

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, headerLines...),
		lipgloss.JoinVertical(lipgloss.Left, listLines...),
		lipgloss.JoinVertical(lipgloss.Left, footerLines...),
	)
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
