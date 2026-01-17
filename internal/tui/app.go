package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateDashboard appState = iota
	stateList
	stateWizard
	stateEditor
	stateDiff
)

type App struct {
	state    appState
	width    int
	height   int
	help     help.Model
	showHelp bool
}

func NewApp() App {
	h := help.New()
	h.Styles.ShortKey = HelpStyle
	h.Styles.ShortDesc = HelpStyle
	h.Styles.FullKey = HelpStyle
	h.Styles.FullDesc = HelpStyle

	return App{
		state: stateDashboard,
		help:  h,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, Keys.Help):
			a.showHelp = !a.showHelp
		case key.Matches(msg, Keys.Back):
			if a.state != stateDashboard {
				a.state = stateDashboard
			}
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width
	}
	return a, nil
}

func (a App) View() string {
	var content string

	switch a.state {
	case stateDashboard:
		content = a.dashboardView()
	case stateList:
		content = a.listView()
	case stateWizard:
		content = a.wizardView()
	case stateEditor:
		content = a.editorView()
	case stateDiff:
		content = a.diffView()
	default:
		content = "Unknown state"
	}

	helpView := a.help.View(Keys)
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

func (a App) dashboardView() string {
	title := TitleStyle.Render("omo-profiler")
	subtitle := SubtitleStyle.Render("Profile manager for oh-my-opencode")

	body := fmt.Sprintf(`
%s
%s

%s Dashboard View

  Press %s to toggle full help
  Press %s to quit
`,
		title,
		subtitle,
		AccentStyle.Render("→"),
		CyanAccentStyle.Render("?"),
		CyanAccentStyle.Render("q"),
	)

	if a.width > 0 && a.height > 0 {
		body += fmt.Sprintf("\n  Window: %dx%d", a.width, a.height)
	}

	return body
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
