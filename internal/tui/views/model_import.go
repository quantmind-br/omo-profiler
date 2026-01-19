package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/models"
	"github.com/sahilm/fuzzy"
)

type NavToModelImportMsg struct{}
type ModelImportBackMsg struct{}
type ModelImportDoneMsg struct {
	Imported int
	Skipped  int
}

type modelImportState int

const (
	stateImportLoading modelImportState = iota
	stateImportProviderList
	stateImportModelList
	stateImportError
)

type modelImportKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Space key.Binding
	Esc   key.Binding
	Retry key.Binding
}

func newModelImportKeyMap() modelImportKeyMap {
	return modelImportKeyMap{
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
			key.WithHelp("enter", "select/import"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Retry: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "retry"),
		),
	}
}

type ModelImport struct {
	state            modelImportState
	response         *models.ModelsDevResponse
	providers        []models.ProviderWithCount
	selectedProvider string
	providerModels   []models.ModelsDevModel
	selectedModels   map[string]bool
	cursor           int
	offset           int
	width            int
	height           int
	spinner          spinner.Model
	searchInput      textinput.Model
	errorMsg         string
	registry         *models.ModelsRegistry
	keys             modelImportKeyMap
}

func NewModelImport() ModelImport {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	searchInput := textinput.New()
	searchInput.Placeholder = "Search models..."
	searchInput.Width = 40

	registry, _ := models.Load()

	return ModelImport{
		state:          stateImportLoading,
		selectedModels: make(map[string]bool),
		spinner:        s,
		searchInput:    searchInput,
		registry:       registry,
		keys:           newModelImportKeyMap(),
	}
}

func (m ModelImport) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchModelsDevCmd,
	)
}

type fetchModelsDevMsg struct {
	response *models.ModelsDevResponse
	err      error
}

func fetchModelsDevCmd() tea.Msg {
	resp, err := models.FetchModelsDevRegistry()
	return fetchModelsDevMsg{response: resp, err: err}
}

func (m ModelImport) Update(msg tea.Msg) (ModelImport, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case fetchModelsDevMsg:
		if msg.err != nil {
			m.state = stateImportError
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.response = msg.response
		m.providers = msg.response.ListProviders()
		m.state = stateImportProviderList
		m.cursor = 0
		return m, nil

	case spinner.TickMsg:
		if m.state == stateImportLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch m.state {
		case stateImportProviderList:
			return m.handleProviderListKeys(msg)
		case stateImportModelList:
			return m.handleModelListKeys(msg)
		case stateImportError:
			return m.handleErrorKeys(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m ModelImport) handleProviderListKeys(msg tea.KeyMsg) (ModelImport, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.providers)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Enter):
		if len(m.providers) > 0 && m.cursor < len(m.providers) {
			m.selectedProvider = m.providers[m.cursor].ID
			m.providerModels = m.response.GetProviderModels(m.selectedProvider)
			m.selectedModels = make(map[string]bool)
			m.cursor = 0
			m.offset = 0
			m.searchInput.SetValue("")
			m.state = stateImportModelList
		}
	case key.Matches(msg, m.keys.Esc):
		return m, func() tea.Msg {
			return ModelImportBackMsg{}
		}
	}
	return m, nil
}

func (m ModelImport) handleModelListKeys(msg tea.KeyMsg) (ModelImport, tea.Cmd) {
	if m.searchInput.Focused() {
		switch msg.String() {
		case "esc":
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			return m, nil
		case "enter":
			m.searchInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	filteredModels := m.getFilteredModels()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(filteredModels)-1 {
			m.cursor++
		}
	case key.Matches(msg, m.keys.Space):
		if len(filteredModels) > 0 && m.cursor < len(filteredModels) {
			modelID := filteredModels[m.cursor].ID
			m.selectedModels[modelID] = !m.selectedModels[modelID]
		}
	case key.Matches(msg, m.keys.Enter):
		return m, m.importSelectedModels()
	case key.Matches(msg, m.keys.Esc):
		m.state = stateImportProviderList
		m.cursor = 0
		m.selectedModels = make(map[string]bool)
		m.searchInput.SetValue("")
		return m, nil
	case msg.String() == "/":
		m.searchInput.Focus()
		return m, nil
	}

	return m, nil
}

func (m ModelImport) handleErrorKeys(msg tea.KeyMsg) (ModelImport, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Retry):
		m.state = stateImportLoading
		m.errorMsg = ""
		return m, tea.Batch(m.spinner.Tick, fetchModelsDevCmd)
	case key.Matches(msg, m.keys.Esc):
		return m, func() tea.Msg {
			return ModelImportBackMsg{}
		}
	}
	return m, nil
}

func (m ModelImport) getFilteredModels() []models.ModelsDevModel {
	searchTerm := strings.TrimSpace(m.searchInput.Value())
	if searchTerm == "" {
		return m.providerModels
	}

	searchStrings := make([]string, len(m.providerModels))
	for i, model := range m.providerModels {
		searchStrings[i] = model.Name + " " + model.ID
	}

	matches := fuzzy.Find(searchTerm, searchStrings)
	if len(matches) == 0 {
		return []models.ModelsDevModel{}
	}

	filtered := make([]models.ModelsDevModel, len(matches))
	for i, match := range matches {
		filtered[i] = m.providerModels[match.Index]
	}
	return filtered
}

func (m ModelImport) importSelectedModels() tea.Cmd {
	return func() tea.Msg {
		imported := 0
		skipped := 0

		for modelID, selected := range m.selectedModels {
			if !selected {
				continue
			}

			var foundModel *models.ModelsDevModel
			for _, model := range m.providerModels {
				if model.ID == modelID {
					foundModel = &model
					break
				}
			}

			if foundModel == nil {
				continue
			}

			registeredModel := foundModel.ToRegisteredModel(m.selectedProvider)
			err := m.registry.Add(registeredModel)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					skipped++
				}
			} else {
				imported++
			}
		}

		return ModelImportDoneMsg{
			Imported: imported,
			Skipped:  skipped,
		}
	}
}

func (m ModelImport) View() string {
	switch m.state {
	case stateImportLoading:
		return m.renderLoading()
	case stateImportProviderList:
		return m.renderProviderList()
	case stateImportModelList:
		return m.renderModelList()
	case stateImportError:
		return m.renderError()
	}
	return ""
}

func (m ModelImport) renderLoading() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		titleStyle.Render("Import from models.dev"),
		"",
		m.spinner.View()+" Loading providers...",
	)
}

func (m ModelImport) renderProviderList() string {
	title := titleStyle.Render("Import from models.dev")

	var lines []string
	for i, provider := range m.providers {
		cursor := "  "
		itemStyle := normalStyle

		if i == m.cursor {
			cursor = accentStyle.Render("> ")
			itemStyle = selectedStyle
		}

		line := fmt.Sprintf("%s%s (%d models)",
			cursor,
			itemStyle.Render(provider.Name),
			provider.ModelCount,
		)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	help := grayStyle.Render("[↑↓] navigate  [enter] select  [esc] back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		content,
		"",
		help,
	)
}

func (m ModelImport) renderModelList() string {
	providerName := m.selectedProvider
	for _, p := range m.providers {
		if p.ID == m.selectedProvider {
			providerName = p.Name
			break
		}
	}

	title := titleStyle.Render(fmt.Sprintf("Import from %s", providerName))

	filteredModels := m.getFilteredModels()

	searchLine := "Search: " + m.searchInput.View()

	var lines []string
	visibleHeight := m.height - 10

	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visibleHeight {
		m.offset = m.cursor - visibleHeight + 1
	}

	for i := m.offset; i < len(filteredModels) && i < m.offset+visibleHeight; i++ {
		model := filteredModels[i]
		cursor := "  "
		checkbox := "[ ]"
		itemStyle := normalStyle

		if m.selectedModels[model.ID] {
			checkbox = "[x]"
		}

		if i == m.cursor {
			cursor = accentStyle.Render("> ")
			itemStyle = selectedStyle
		}

		capabilities := model.FormatCapabilities()
		line := fmt.Sprintf("%s%s %s %s",
			cursor,
			checkbox,
			itemStyle.Render(model.Name),
			grayStyle.Render(capabilities),
		)
		lines = append(lines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	selectedCount := 0
	for _, selected := range m.selectedModels {
		if selected {
			selectedCount++
		}
	}

	help := grayStyle.Render(fmt.Sprintf("%d selected  [space] toggle  [enter] import  [/] search  [esc] back", selectedCount))

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

func (m ModelImport) renderError() string {
	title := titleStyle.Render("Import from models.dev")
	errorText := errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
	help := grayStyle.Render("[r] retry  [esc] back")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		errorText,
		"",
		help,
	)
}

func (m *ModelImport) SetSize(width, height int) {
	m.width = width
	m.height = height
}
