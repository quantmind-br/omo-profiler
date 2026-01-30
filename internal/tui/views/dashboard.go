package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
)

type NavToListMsg struct{}
type NavToWizardMsg struct{}
type NavToEditorMsg struct{}
type NavToDiffMsg struct{}
type NavToImportMsg struct{}
type NavToExportMsg struct{}
type NavToModelsMsg struct{}
type NavToTemplateSelectMsg struct{}

const (
	menuSwitch = iota
	menuCreate
	menuCreateFromTemplate
	menuEdit
	menuCompare
	menuModels
	menuImport
	menuExport
)

var menuItems = []string{
	"Switch Profile",
	"Create New",
	"Create from Template",
	"Edit Current",
	"Compare Profiles",
	"Manage Models",
	"Import Profile",
	"Export Profile",
}

type Dashboard struct {
	activeProfile *profile.ActiveConfig
	profileCount  int
	cursor        int
	width         int
	height        int
	keys          dashboardKeyMap
	err           error
}

type dashboardKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Import key.Binding
	Export key.Binding
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9E2AF"))

	grayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#7D56F4"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4"))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6AC1"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))
)

func NewDashboard() Dashboard {
	return Dashboard{
		cursor: 0,
		keys: dashboardKeyMap{
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
			Import: key.NewBinding(
				key.WithKeys("i"),
				key.WithHelp("i", "import"),
			),
			Export: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "export"),
			),
		},
	}
}

func (d Dashboard) Init() tea.Cmd {
	return d.loadActiveProfile
}

func (d Dashboard) loadActiveProfile() tea.Msg {
	active, err := profile.GetActive()
	if err != nil {
		return profileLoadedMsg{err: err}
	}

	profiles, err := profile.List()
	if err != nil {
		return profileLoadedMsg{err: err}
	}

	return profileLoadedMsg{
		active: active,
		count:  len(profiles),
	}
}

type profileLoadedMsg struct {
	active *profile.ActiveConfig
	count  int
	err    error
}

func (d Dashboard) Update(msg tea.Msg) (Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case profileLoadedMsg:
		if msg.err != nil {
			d.err = msg.err
		} else {
			d.activeProfile = msg.active
			d.profileCount = msg.count
		}

	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Up):
			if d.cursor > 0 {
				d.cursor--
			}
		case key.Matches(msg, d.keys.Down):
			if d.cursor < len(menuItems)-1 {
				d.cursor++
			}
		case key.Matches(msg, d.keys.Enter):
			return d, d.handleSelect()
		case key.Matches(msg, d.keys.Import):
			return d, func() tea.Msg { return NavToImportMsg{} }
		case key.Matches(msg, d.keys.Export):
			return d, func() tea.Msg { return NavToExportMsg{} }
		}
	}

	return d, nil
}

func (d Dashboard) handleSelect() tea.Cmd {
	return func() tea.Msg {
		switch d.cursor {
		case menuSwitch:
			return NavToListMsg{}
		case menuCreate:
			return NavToWizardMsg{}
		case menuCreateFromTemplate:
			return NavToTemplateSelectMsg{}
		case menuEdit:
			return NavToEditorMsg{}
		case menuCompare:
			return NavToDiffMsg{}
		case menuModels:
			return NavToModelsMsg{}
		case menuImport:
			return NavToImportMsg{}
		case menuExport:
			return NavToExportMsg{}
		}
		return nil
	}
}

func (d Dashboard) View() string {
	title := titleStyle.Render("omo-profiler")
	subtitle := subtitleStyle.Render("Profile manager for oh-my-opencode")

	var profileStatus string
	if d.err != nil {
		profileStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Render(fmt.Sprintf("Error: %v", d.err))
	} else if d.activeProfile == nil {
		profileStatus = grayStyle.Render("Loading...")
	} else if !d.activeProfile.Exists {
		profileStatus = fmt.Sprintf("Active: %s", grayStyle.Render("(None)"))
	} else if d.activeProfile.IsOrphan {
		profileStatus = fmt.Sprintf("Active: %s", warningStyle.Render("(Custom)"))
	} else {
		profileStatus = fmt.Sprintf("Active: %s", successStyle.Render(d.activeProfile.ProfileName))
	}

	statsLine := subtitleStyle.Render(fmt.Sprintf("%d profiles available", d.profileCount))
	menuView := d.renderMenu()

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		subtitle,
		"",
		profileStatus,
		statsLine,
		"",
		menuView,
	)
}

func (d Dashboard) renderMenu() string {
	var lines []string

	for i, item := range menuItems {
		cursor := "   "
		itemStyle := normalStyle

		if i == d.cursor {
			cursor = accentStyle.Render(" > ")
			itemStyle = selectedStyle
		}

		line := fmt.Sprintf("%s%s", cursor, itemStyle.Render(" "+item+" "))
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (d Dashboard) Refresh() tea.Cmd {
	return d.loadActiveProfile
}

func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height
}
