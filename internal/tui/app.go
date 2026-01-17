package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/tui/views"
)

type appState int

const (
	stateDashboard appState = iota
	stateList
	stateWizard
	stateEditor
	stateDiff
	stateImport
	stateExport
)

type App struct {
	state     appState
	width     int
	height    int
	help      help.Model
	showHelp  bool
	dashboard views.Dashboard
}

func NewApp() App {
	h := help.New()
	h.Styles.ShortKey = HelpStyle
	h.Styles.ShortDesc = HelpStyle
	h.Styles.FullKey = HelpStyle
	h.Styles.FullDesc = HelpStyle

	return App{
		state:     stateDashboard,
		help:      h,
		dashboard: views.NewDashboard(),
	}
}

func (a App) Init() tea.Cmd {
	return a.dashboard.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, Keys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(msg, Keys.Back):
			if a.state != stateDashboard {
				a.state = stateDashboard
				a.dashboard, cmd = a.dashboard.Update(nil)
				return a, a.dashboard.Refresh()
			}
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width
		a.dashboard.SetSize(msg.Width, msg.Height)

	case views.NavToListMsg:
		a.state = stateList
		return a, nil

	case views.NavToWizardMsg:
		a.state = stateWizard
		return a, nil

	case views.NavToEditorMsg:
		a.state = stateEditor
		return a, nil

	case views.NavToDiffMsg:
		a.state = stateDiff
		return a, nil

	case views.NavToImportMsg:
		a.state = stateImport
		return a, nil

	case views.NavToExportMsg:
		a.state = stateExport
		return a, nil
	}

	switch a.state {
	case stateDashboard:
		a.dashboard, cmd = a.dashboard.Update(msg)
	}

	return a, cmd
}

func (a App) View() string {
	var content string

	switch a.state {
	case stateDashboard:
		content = a.dashboard.View()
	case stateList:
		content = a.listView()
	case stateWizard:
		content = a.wizardView()
	case stateEditor:
		content = a.editorView()
	case stateDiff:
		content = a.diffView()
	case stateImport:
		content = a.importView()
	case stateExport:
		content = a.exportView()
	default:
		content = "Unknown state"
	}

	var helpView string
	if a.showHelp {
		helpView = a.help.View(Keys)
	} else {
		helpView = HelpStyle.Render("? help • q quit")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		"",
		helpView,
	)
}

func (a App) listView() string {
	return TitleStyle.Render("Profile List") + "\n\n" +
		SubtitleStyle.Render("(placeholder - profile list will appear here)")
}

func (a App) wizardView() string {
	return TitleStyle.Render("New Profile Wizard") + "\n\n" +
		SubtitleStyle.Render("(placeholder - wizard will appear here)")
}

func (a App) editorView() string {
	return TitleStyle.Render("Profile Editor") + "\n\n" +
		SubtitleStyle.Render("(placeholder - editor will appear here)")
}

func (a App) diffView() string {
	return TitleStyle.Render("Profile Diff") + "\n\n" +
		SubtitleStyle.Render("(placeholder - diff will appear here)")
}

func (a App) importView() string {
	return TitleStyle.Render("Import Profile") + "\n\n" +
		SubtitleStyle.Render("(placeholder - import will appear here)")
}

func (a App) exportView() string {
	return TitleStyle.Render("Export Profile") + "\n\n" +
		SubtitleStyle.Render("(placeholder - export will appear here)")
}
