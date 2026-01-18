package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
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
	stateModels
)

// Toast message types
type toastType int

const (
	toastSuccess toastType = iota
	toastError
	toastInfo
)

type toastMsg struct {
	text     string
	typ      toastType
	duration time.Duration
}

type clearToastMsg struct{}

// Operation messages
type switchProfileDoneMsg struct {
	name string
	err  error
}

type deleteProfileDoneMsg struct {
	name string
	err  error
}

type App struct {
	state     appState
	prevState appState
	width     int
	height    int
	ready     bool

	// Help
	help     help.Model
	showHelp bool

	// Loading state
	spinner spinner.Model
	loading bool

	// Toast notification
	toast       string
	toastType   toastType
	toastActive bool

	// Views
	dashboard     views.Dashboard
	list          views.List
	wizard        views.Wizard
	editor        views.Editor
	diff          views.Diff
	modelRegistry views.ModelRegistry

	// Context for editor
	editProfileName string
}

func NewApp() App {
	h := help.New()
	h.Styles.ShortKey = HelpStyle
	h.Styles.ShortDesc = HelpStyle
	h.Styles.FullKey = HelpStyle
	h.Styles.FullDesc = HelpStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Purple)

	return App{
		state:     stateDashboard,
		help:      h,
		spinner:   s,
		dashboard: views.NewDashboard(),
		list:      views.NewList(),
		diff:      views.NewDiff(),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.dashboard.Init(),
		a.spinner.Tick,
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys - but don't intercept 'q' in Wizard or Editor (text input)
		switch {
		case key.Matches(msg, Keys.Quit):
			// Only quit with 'q' in Dashboard or List states
			// Wizard and Editor handle their own quit via ctrl+c
			if msg.String() == "q" && (a.state == stateWizard || a.state == stateEditor) {
				break
			}
			return a, tea.Quit
		case key.Matches(msg, Keys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(msg, Keys.Back):
			// Don't intercept Esc if a view handles it internally
			if a.state == stateWizard || a.state == stateEditor || a.state == stateDiff || a.state == stateModels {
				// Let the view handle it
				break
			}
			if a.state != stateDashboard {
				return a.navigateTo(stateDashboard)
			}
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width
		a.ready = true

		// Propagate to all views
		a.dashboard.SetSize(msg.Width, msg.Height-3)
		a.list.SetSize(msg.Width, msg.Height-3)
		a.wizard.SetSize(msg.Width, msg.Height-3)
		a.editor.SetSize(msg.Width, msg.Height-3)

	case spinner.TickMsg:
		if a.loading {
			a.spinner, cmd = a.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case clearToastMsg:
		a.toast = ""
		a.toastActive = false
		return a, nil

	case toastMsg:
		a.toast = msg.text
		a.toastType = msg.typ
		a.toastActive = true
		return a, tea.Tick(msg.duration, func(time.Time) tea.Msg {
			return clearToastMsg{}
		})

	// Navigation messages from Dashboard
	case views.NavToListMsg:
		a.list = views.NewList()
		a.list.SetSize(a.width, a.height-3)
		return a.navigateTo(stateList)

	case views.NavToWizardMsg:
		a.wizard = views.NewWizard()
		a.wizard.SetSize(a.width, a.height-3)
		return a.navigateTo(stateWizard)

	case views.NavToEditorMsg:
		// Edit current active profile using wizard
		active, err := profile.GetActive()
		if err != nil || active == nil || !active.Exists || active.IsOrphan {
			return a, a.showToast("No active profile to edit", toastError, 3*time.Second)
		}
		p, err := profile.Load(active.ProfileName)
		if err != nil {
			return a, a.showToast("Failed to load profile: "+err.Error(), toastError, 3*time.Second)
		}
		a.wizard = views.NewWizardForEdit(p)
		a.wizard.SetSize(a.width, a.height-3)
		return a.navigateTo(stateWizard)

	case views.NavToDiffMsg:
		a.diff = views.NewDiff()
		return a.navigateTo(stateDiff)

	case views.NavToImportMsg:
		return a, a.showToast("Import not yet implemented", toastInfo, 2*time.Second)

	case views.NavToExportMsg:
		return a, a.showToast("Export not yet implemented", toastInfo, 2*time.Second)

	case views.NavToModelsMsg:
		a.modelRegistry = views.NewModelRegistry()
		a.modelRegistry.SetSize(a.width, a.height-3)
		return a.navigateTo(stateModels)

	case views.ModelRegistryBackMsg:
		return a.navigateTo(stateDashboard)

	// Navigation from List
	case views.NavigateToDashboardMsg:
		return a.navigateTo(stateDashboard)

	case views.NavigateToWizardMsg:
		a.wizard = views.NewWizard()
		a.wizard.SetSize(a.width, a.height-3)
		return a.navigateTo(stateWizard)

	case views.SwitchProfileMsg:
		a.loading = true
		return a, tea.Batch(
			a.spinner.Tick,
			a.doSwitchProfile(msg.Name),
		)

	case switchProfileDoneMsg:
		a.loading = false
		if msg.err != nil {
			return a, a.showToast("Switch failed: "+msg.err.Error(), toastError, 3*time.Second)
		}
		cmds = append(cmds, a.showToast("Switched to: "+msg.name, toastSuccess, 3*time.Second))
		a.dashboard = views.NewDashboard()
		a.dashboard.SetSize(a.width, a.height-3)
		cmds = append(cmds, a.dashboard.Init())
		a.state = stateDashboard
		return a, tea.Batch(cmds...)

	case views.EditProfileMsg:
		p, err := profile.Load(msg.Name)
		if err != nil {
			return a, a.showToast("Failed to load profile: "+err.Error(), toastError, 3*time.Second)
		}
		a.wizard = views.NewWizardForEdit(p)
		a.wizard.SetSize(a.width, a.height-3)
		return a.navigateTo(stateWizard)

	case views.DeleteProfileMsg:
		a.loading = true
		return a, tea.Batch(
			a.spinner.Tick,
			a.doDeleteProfile(msg.Name),
		)

	case deleteProfileDoneMsg:
		a.loading = false
		if msg.err != nil {
			return a, a.showToast("Delete failed: "+msg.err.Error(), toastError, 3*time.Second)
		}
		cmds = append(cmds, a.showToast("Deleted: "+msg.name, toastSuccess, 3*time.Second))
		// Refresh list
		a.list = views.NewList()
		a.list.SetSize(a.width, a.height-3)
		cmds = append(cmds, a.list.Init())
		return a, tea.Batch(cmds...)

	// Wizard messages
	case views.WizardSaveMsg:
		cmds = append(cmds, a.showToast("Profile saved!", toastSuccess, 3*time.Second))
		a.dashboard = views.NewDashboard()
		a.dashboard.SetSize(a.width, a.height-3)
		cmds = append(cmds, a.dashboard.Init())
		a.state = stateDashboard
		return a, tea.Batch(cmds...)

	case views.WizardCancelMsg:
		return a.navigateTo(stateDashboard)

	// Editor messages
	case views.EditorSaveSuccessMsg:
		cmds = append(cmds, a.showToast("Profile saved!", toastSuccess, 3*time.Second))
		a.dashboard = views.NewDashboard()
		a.dashboard.SetSize(a.width, a.height-3)
		cmds = append(cmds, a.dashboard.Init())
		a.state = stateDashboard
		return a, tea.Batch(cmds...)

	case views.EditorCancelMsg:
		return a.navigateTo(stateDashboard)
	}

	// Delegate update to current view
	switch a.state {
	case stateDashboard:
		a.dashboard, cmd = a.dashboard.Update(msg)
		cmds = append(cmds, cmd)

	case stateList:
		a.list, cmd = a.list.Update(msg)
		cmds = append(cmds, cmd)

	case stateWizard:
		a.wizard, cmd = a.wizard.Update(msg)
		cmds = append(cmds, cmd)

	case stateEditor:
		a.editor, cmd = a.editor.Update(msg)
		cmds = append(cmds, cmd)

	case stateDiff:
		a.diff, cmd = a.diff.Update(msg)
		cmds = append(cmds, cmd)

	case stateModels:
		a.modelRegistry, cmd = a.modelRegistry.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a App) navigateTo(state appState) (App, tea.Cmd) {
	a.prevState = a.state
	a.state = state

	var cmd tea.Cmd
	switch state {
	case stateDashboard:
		a.dashboard = views.NewDashboard()
		a.dashboard.SetSize(a.width, a.height-3)
		cmd = a.dashboard.Init()
	case stateList:
		cmd = a.list.Init()
	case stateWizard:
		cmd = a.wizard.Init()
	case stateEditor:
		cmd = a.editor.Init()
	case stateDiff:
		cmd = a.diff.Init()
	case stateModels:
		cmd = a.modelRegistry.Init()
	}

	return a, cmd
}

func (a App) doSwitchProfile(name string) tea.Cmd {
	return func() tea.Msg {
		err := profile.SetActive(name)
		return switchProfileDoneMsg{name: name, err: err}
	}
}

func (a App) doDeleteProfile(name string) tea.Cmd {
	return func() tea.Msg {
		err := profile.Delete(name)
		return deleteProfileDoneMsg{name: name, err: err}
	}
}

func (a App) showToast(text string, typ toastType, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return toastMsg{text: text, typ: typ, duration: duration}
	}
}

func (a App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	var content string

	// Loading overlay
	if a.loading {
		content = lipgloss.Place(
			a.width,
			a.height-3,
			lipgloss.Center,
			lipgloss.Center,
			a.spinner.View()+" Loading...",
		)
	} else {
		switch a.state {
		case stateDashboard:
			content = a.dashboard.View()
		case stateList:
			content = a.list.View()
		case stateWizard:
			content = a.wizard.View()
		case stateEditor:
			content = a.editor.View()
		case stateDiff:
			content = a.diff.View()
		case stateImport:
			content = a.placeholderView("Import Profile")
		case stateExport:
			content = a.placeholderView("Export Profile")
		case stateModels:
			content = a.modelRegistry.View()
		default:
			content = "Unknown state"
		}
	}

	// Toast notification
	var toastView string
	if a.toastActive && a.toast != "" {
		var style lipgloss.Style
		switch a.toastType {
		case toastSuccess:
			style = SuccessStyle.Padding(0, 1)
		case toastError:
			style = ErrorStyle.Padding(0, 1)
		default:
			style = CyanAccentStyle.Padding(0, 1)
		}
		toastView = style.Render(a.toast)
	}

	// Help view
	var helpView string
	if a.showHelp {
		helpView = a.renderFullHelp()
	} else {
		helpView = a.renderShortHelp()
	}

	// Build final layout
	var parts []string
	parts = append(parts, content)

	if toastView != "" {
		parts = append(parts, toastView)
	}

	parts = append(parts, helpView)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (a App) renderShortHelp() string {
	var hints []string

	switch a.state {
	case stateDashboard:
		hints = []string{"↑↓ navigate", "enter select", "? help", "q quit"}
	case stateList:
		hints = []string{"enter switch", "e edit", "d delete", "n new", "/ search", "esc back"}
	case stateWizard:
		hints = []string{"tab/enter next", "shift+tab back", "ctrl+s save", "ctrl+c cancel"}
	case stateEditor:
		hints = []string{"tab switch focus", "ctrl+s save", "esc back"}
	case stateDiff:
		hints = []string{"tab switch pane", "enter select", "↑↓ scroll", "esc back"}
	case stateModels:
		hints = []string{"n new", "e edit", "d delete", "↑↓ navigate", "esc back"}
	default:
		hints = []string{"? help", "q quit"}
	}

	return HelpStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, joinWithSeparator(hints, " • ")...))
}

func (a App) renderFullHelp() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, "")

	// Global keys
	lines = append(lines, AccentStyle.Render("Global:"))
	lines = append(lines, HelpStyle.Render("  q/ctrl+c   Quit application"))
	lines = append(lines, HelpStyle.Render("  ?          Toggle help"))
	lines = append(lines, HelpStyle.Render("  esc        Back/Cancel"))
	lines = append(lines, "")

	// Context-specific keys
	switch a.state {
	case stateDashboard:
		lines = append(lines, AccentStyle.Render("Dashboard:"))
		lines = append(lines, HelpStyle.Render("  ↑/k        Move up"))
		lines = append(lines, HelpStyle.Render("  ↓/j        Move down"))
		lines = append(lines, HelpStyle.Render("  enter      Select menu item"))

	case stateList:
		lines = append(lines, AccentStyle.Render("Profile List:"))
		lines = append(lines, HelpStyle.Render("  ↑/k        Move up"))
		lines = append(lines, HelpStyle.Render("  ↓/j        Move down"))
		lines = append(lines, HelpStyle.Render("  enter      Switch to profile"))
		lines = append(lines, HelpStyle.Render("  e          Edit profile"))
		lines = append(lines, HelpStyle.Render("  d          Delete profile"))
		lines = append(lines, HelpStyle.Render("  n          New profile"))
		lines = append(lines, HelpStyle.Render("  /          Search profiles"))

	case stateWizard:
		lines = append(lines, AccentStyle.Render("Profile Wizard:"))
		lines = append(lines, HelpStyle.Render("  tab/enter  Next step"))
		lines = append(lines, HelpStyle.Render("  shift+tab  Previous step"))
		lines = append(lines, HelpStyle.Render("  ctrl+s     Save profile"))
		lines = append(lines, HelpStyle.Render("  ctrl+c     Cancel"))

	case stateEditor:
		lines = append(lines, AccentStyle.Render("Profile Editor:"))
		lines = append(lines, HelpStyle.Render("  ↑/k        Move up"))
		lines = append(lines, HelpStyle.Render("  ↓/j        Move down"))
		lines = append(lines, HelpStyle.Render("  tab        Switch focus"))
		lines = append(lines, HelpStyle.Render("  space      Toggle option"))
		lines = append(lines, HelpStyle.Render("  ctrl+s     Save changes"))

	case stateDiff:
		lines = append(lines, AccentStyle.Render("Profile Diff:"))
		lines = append(lines, HelpStyle.Render("  ↑/k        Scroll up"))
		lines = append(lines, HelpStyle.Render("  ↓/j        Scroll down"))
		lines = append(lines, HelpStyle.Render("  tab        Switch pane"))
		lines = append(lines, HelpStyle.Render("  enter      Select profile"))
		lines = append(lines, HelpStyle.Render("  pgup/pgdn  Page scroll"))

	case stateModels:
		lines = append(lines, AccentStyle.Render("Model Registry:"))
		lines = append(lines, HelpStyle.Render("  ↑/k        Move up"))
		lines = append(lines, HelpStyle.Render("  ↓/j        Move down"))
		lines = append(lines, HelpStyle.Render("  n          New model"))
		lines = append(lines, HelpStyle.Render("  e          Edit model"))
		lines = append(lines, HelpStyle.Render("  d          Delete model"))
		lines = append(lines, HelpStyle.Render("  enter      Confirm"))
		lines = append(lines, HelpStyle.Render("  esc        Back/Cancel"))
	}

	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("Press ? to close help"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (a App) placeholderView(title string) string {
	return TitleStyle.Render(title) + "\n\n" +
		SubtitleStyle.Render("(Coming soon)")
}

func joinWithSeparator(items []string, sep string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			result = append(result, sep)
		}
		result = append(result, item)
	}
	return result
}
