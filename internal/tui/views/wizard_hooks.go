package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
)

// All 36 hooks - must stay in sync with upstream schema order
var allHooks = []string{
	"todo-continuation-enforcer",
	"context-window-monitor",
	"session-recovery",
	"session-notification",
	"comment-checker",
	"grep-output-truncator",
	"tool-output-truncator",
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
	"ralph-loop",
	"category-skill-reminder",
	"compaction-context-injector",
	"claude-code-hooks",
	"auto-slash-command",
	"edit-error-recovery",
	"delegate-task-retry",
	"prometheus-md-only",
	"sisyphus-junior-notepad",
	"start-work",
	"atlas",
	"unstable-agent-babysitter",
	"stop-continuation-guard",
	"tasks-todowrite-disabler",
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
	if !w.ready {
		w.viewport = viewport.New(width, height-4)
		w.ready = true
	} else {
		w.viewport.Width = width
		w.viewport.Height = height - 4
	}
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

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

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
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	title := titleStyle.Render("Configure Hooks")
	desc := helpStyle.Render("Space to toggle • ✓ enabled • ✗ disabled • Tab next • Shift+Tab back")

	disabledCount := 0
	for _, d := range w.disabled {
		if d {
			disabledCount++
		}
	}
	stats := helpStyle.Render(fmt.Sprintf("%d/%d hooks disabled", disabledCount, len(allHooks)))

	content := w.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		stats,
		"",
		content,
	)
}
