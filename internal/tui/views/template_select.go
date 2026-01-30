package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
)

// NavToWizardFromTemplateMsg is emitted when user selects a template
type NavToWizardFromTemplateMsg struct {
	TemplateName string
}

// TemplateSelectCancelMsg is emitted when user cancels template selection
type TemplateSelectCancelMsg struct{}

type templateSelectKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Cancel key.Binding
}

func newTemplateSelectKeyMap() templateSelectKeyMap {
	return templateSelectKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// TemplateSelect is a view for selecting a profile as template
type TemplateSelect struct {
	profiles []string
	cursor   int
	width    int
	height   int
	keys     templateSelectKeyMap
}

// NewTemplateSelect creates a new template selection view
func NewTemplateSelect() TemplateSelect {
	profiles, _ := profile.List()
	return TemplateSelect{
		profiles: profiles,
		cursor:   0,
		keys:     newTemplateSelectKeyMap(),
	}
}

// Init initializes the view
func (t TemplateSelect) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (t TemplateSelect) Update(msg tea.Msg) (TemplateSelect, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keys.Up):
			if t.cursor > 0 {
				t.cursor--
			}
		case key.Matches(msg, t.keys.Down):
			if t.cursor < len(t.profiles)-1 {
				t.cursor++
			}
		case key.Matches(msg, t.keys.Select):
			if len(t.profiles) > 0 && t.cursor < len(t.profiles) {
				return t, func() tea.Msg {
					return NavToWizardFromTemplateMsg{
						TemplateName: t.profiles[t.cursor],
					}
				}
			}
		case key.Matches(msg, t.keys.Cancel):
			return t, func() tea.Msg {
				return TemplateSelectCancelMsg{}
			}
		}
	}
	return t, nil
}

// View renders the template selection view
func (t TemplateSelect) View() string {
	if len(t.profiles) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Create from Template"),
			"",
			normalStyle.Render("No profiles available to use as template."),
			"",
			normalStyle.Render("Press esc to go back."),
		)
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Create from Template"))
	lines = append(lines, "")
	lines = append(lines, normalStyle.Render("Select a profile to use as template:"))
	lines = append(lines, "")

	for i, profile := range t.profiles {
		cursor := "   "
		itemStyle := normalStyle

		if i == t.cursor {
			cursor = accentStyle.Render(" > ")
			itemStyle = selectedStyle
		}

		line := cursor + itemStyle.Render(" "+profile+" ")
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, normalStyle.Render("↑/↓ navigate • enter select • esc cancel"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// SetSize sets the view dimensions
func (t *TemplateSelect) SetSize(width, height int) {
	t.width = width
	t.height = height
}
