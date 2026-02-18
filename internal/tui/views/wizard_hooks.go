package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

var (
	wizHooksPurple = lipgloss.Color("#7D56F4")
	wizHooksGreen  = lipgloss.Color("#A6E3A1")
	wizHooksRed    = lipgloss.Color("#F38BA8")
	wizHooksGray   = lipgloss.Color("#6C7086")
	wizHooksWhite  = lipgloss.Color("#CDD6F4")
)

var (
	wizHooksSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizHooksPurple)
	wizHooksEnabledStyle  = lipgloss.NewStyle().Foreground(wizHooksGreen)
	wizHooksDisabledStyle = lipgloss.NewStyle().Foreground(wizHooksRed)
	wizHooksDimStyle      = lipgloss.NewStyle().Foreground(wizHooksGray)
	wizHooksNameStyle     = lipgloss.NewStyle().Bold(true).Foreground(wizHooksWhite)
	wizHooksTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizHooksWhite)
	wizHooksHelpStyle     = lipgloss.NewStyle().Foreground(wizHooksGray)
)

var allHooks = []string{
	"todo-continuation-enforcer",
	"context-window-monitor",
	"session-recovery",
	"session-notification",
	"comment-checker",
	"grep-output-truncator",
	"tool-output-truncator",
	"question-label-truncator",
	"directory-agents-injector",
	"directory-readme-injector",
	"empty-task-response-detector",
	"think-mode",
	"anthropic-context-window-limit-recovery",
	"preemptive-compaction",
	"rules-injector",
	"background-notification",
	"auto-update-checker",
	"startup-toast",
	"keyword-detector",
	"agent-usage-reminder",
	"non-interactive-env",
	"interactive-bash-session",
	"thinking-block-validator",
	"ultrawork-model-override",
	"ralph-loop",
	"category-skill-reminder",
	"compaction-context-injector",
	"compaction-todo-preserver",
	"claude-code-hooks",
	"auto-slash-command",
	"edit-error-recovery",
	"json-error-recovery",
	"delegate-task-retry",
	"prometheus-md-only",
	"sisyphus-junior-notepad",
	"no-sisyphus-gpt",
	"start-work",
	"atlas",
	"unstable-agent-babysitter",
	"task-reminder",
	"task-resume-info",
	"stop-continuation-guard",
	"tasks-todowrite-disabler",
	"write-existing-file-guard",
	"anthropic-effort",
	"hashline-read-enhancer",
}

type wizardHooksKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Next     key.Binding
	Back     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

func newWizardHooksKeyMap() wizardHooksKeyMap {
	return wizardHooksKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next step"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab", "esc"),
			key.WithHelp("shift+tab/esc", "back"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdown", "page down"),
		),
	}
}

// WizardHooks is step 3: Hook configuration
type WizardHooks struct {
	disabled map[string]bool
	cursor   int
	viewport viewport.Model
	ready    bool
	width    int
	height   int
	keys     wizardHooksKeyMap
}

func NewWizardHooks() WizardHooks {
	disabled := make(map[string]bool)
	for _, hook := range allHooks {
		disabled[hook] = false // All enabled by default
	}

	return WizardHooks{
		disabled: disabled,
		keys:     newWizardHooksKeyMap(),
	}
}

func (w WizardHooks) Init() tea.Cmd {
	return nil
}

func (w *WizardHooks) SetSize(width, height int) {
	w.width = width
	w.height = height
	overhead := 4
	if layout.IsShort(height) {
		overhead = 2
	}
	if !w.ready {
		w.viewport = viewport.New(width, height-overhead)
		w.ready = true
	} else {
		w.viewport.Width = width
		w.viewport.Height = height - overhead
	}
	w.viewport.SetContent(w.renderContent())
}

func (w *WizardHooks) SetConfig(cfg *config.Config) {
	// Reset all to enabled
	for hook := range w.disabled {
		w.disabled[hook] = false
	}
	// Mark disabled ones
	for _, hook := range cfg.DisabledHooks {
		if _, ok := w.disabled[hook]; ok {
			w.disabled[hook] = true
		}
	}
}

func (w *WizardHooks) Apply(cfg *config.Config) {
	var disabled []string
	for _, hook := range allHooks {
		if w.disabled[hook] {
			disabled = append(disabled, hook)
		}
	}
	cfg.DisabledHooks = disabled
	if len(cfg.DisabledHooks) == 0 {
		cfg.DisabledHooks = nil
	}
}

func (w WizardHooks) Update(msg tea.Msg) (WizardHooks, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keys.Up):
			if w.cursor > 0 {
				w.cursor--
			}
		case key.Matches(msg, w.keys.Down):
			if w.cursor < len(allHooks)-1 {
				w.cursor++
			}
		case key.Matches(msg, w.keys.Toggle):
			hook := allHooks[w.cursor]
			w.disabled[hook] = !w.disabled[hook]
		case key.Matches(msg, w.keys.Next):
			return w, func() tea.Msg { return WizardNextMsg{} }
		case key.Matches(msg, w.keys.Back):
			return w, func() tea.Msg { return WizardBackMsg{} }
		case key.Matches(msg, w.keys.PageUp):
			w.cursor -= 10
			if w.cursor < 0 {
				w.cursor = 0
			}
		case key.Matches(msg, w.keys.PageDown):
			w.cursor += 10
			if w.cursor >= len(allHooks) {
				w.cursor = len(allHooks) - 1
			}
		}
	}

	// Update viewport content
	w.viewport.SetContent(w.renderContent())
	w.viewport, cmd = w.viewport.Update(msg)

	return w, cmd
}

func (w WizardHooks) renderContent() string {
	var lines []string

	selectedStyle := wizHooksSelectedStyle
	enabledStyle := wizHooksEnabledStyle
	disabledStyle := wizHooksDisabledStyle
	dimStyle := wizHooksDimStyle

	for i, hook := range allHooks {
		cursor := "  "
		if i == w.cursor {
			cursor = selectedStyle.Render("> ")
		}

		var checkbox string
		if w.disabled[hook] {
			checkbox = disabledStyle.Render("[✗]")
		} else {
			checkbox = enabledStyle.Render("[✓]")
		}

		nameStyle := dimStyle
		if i == w.cursor {
			nameStyle = wizHooksNameStyle
		}

		status := ""
		if w.disabled[hook] {
			status = disabledStyle.Render(" (disabled)")
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, nameStyle.Render(hook), status)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (w WizardHooks) View() string {
	titleStyle := wizHooksTitleStyle
	helpStyle := wizHooksHelpStyle

	disabledCount := 0
	for _, d := range w.disabled {
		if d {
			disabledCount++
		}
	}

	content := w.viewport.View()

	if layout.IsShort(w.height) {
		title := titleStyle.Render("Hooks")
		stats := helpStyle.Render(fmt.Sprintf(" (%d/%d disabled)", disabledCount, len(allHooks)))
		return lipgloss.JoinVertical(lipgloss.Left,
			title+stats,
			content,
		)
	}

	title := titleStyle.Render("Configure Hooks")
	desc := helpStyle.Render("Space to toggle • ✓ enabled • ✗ disabled • Tab next • Shift+Tab back")
	stats := helpStyle.Render(fmt.Sprintf("%d/%d hooks disabled", disabledCount, len(allHooks)))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		stats,
		"",
		content,
	)
}
