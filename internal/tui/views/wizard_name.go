package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
)

var (
	wizNameWhite = lipgloss.Color("#CDD6F4")
	wizNameGray  = lipgloss.Color("#6C7086")
	wizNameRed   = lipgloss.Color("#F38BA8")
	wizNameGreen = lipgloss.Color("#A6E3A1")
)

var (
	wizNameLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(wizNameWhite)
	wizNameDescStyle  = lipgloss.NewStyle().Foreground(wizNameGray)
	wizNameErrorStyle = lipgloss.NewStyle().Foreground(wizNameRed)
	wizNameValidStyle = lipgloss.NewStyle().Foreground(wizNameGreen)
)

type wizardNameKeyMap struct {
	Next   key.Binding
	Cancel key.Binding
}

func newWizardNameKeyMap() wizardNameKeyMap {
	return wizardNameKeyMap{
		Next: key.NewBinding(
			key.WithKeys("tab", "enter"),
			key.WithHelp("tab/enter", "next"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// WizardName is step 1: Profile name input
type WizardName struct {
	input  textinput.Model
	name   string
	err    error
	valid  bool
	width  int
	height int
	keys   wizardNameKeyMap
}

func NewWizardName() WizardName {
	ti := textinput.New()
	ti.Placeholder = "my-profile"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	return WizardName{
		input: ti,
		keys:  newWizardNameKeyMap(),
	}
}

func (w WizardName) Init() tea.Cmd {
	return textinput.Blink
}

func (w *WizardName) SetSize(width, height int) {
	w.width = width
	w.height = height
}

func (w *WizardName) SetName(name string) {
	w.name = name
	w.input.SetValue(name)
	w.validate()
}

func (w *WizardName) validate() {
	name := w.input.Value()
	if name == "" {
		w.err = profile.ErrEmptyName
		w.valid = false
		return
	}
	if err := profile.ValidateName(name); err != nil {
		w.err = err
		w.valid = false
		return
	}
	w.err = nil
	w.valid = true
	w.name = name
}

func (w WizardName) Update(msg tea.Msg) (WizardName, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keys.Next):
			w.validate()
			if w.valid {
				return w, func() tea.Msg { return WizardNextMsg{} }
			}
			return w, nil

		case key.Matches(msg, w.keys.Cancel):
			return w, func() tea.Msg { return WizardCancelMsg{} }
		}
	}

	// Update text input
	w.input, cmd = w.input.Update(msg)
	w.validate()

	return w, cmd
}

func (w WizardName) View() string {
	label := wizNameLabelStyle.Render("Profile Name")
	desc := wizNameDescStyle.Render("Enter a name for your profile (letters, numbers, hyphens, underscores only)")

	input := w.input.View()

	var status string
	if w.input.Value() == "" {
		status = wizNameDescStyle.Render("Required")
	} else if w.err != nil {
		status = wizNameErrorStyle.Render(fmt.Sprintf("✗ %s", w.err.Error()))
	} else {
		status = wizNameValidStyle.Render("✓ Valid name")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		desc,
		"",
		input,
		status,
	)
}

// IsComplete returns true if the name is valid
func (w WizardName) IsComplete() bool {
	return w.valid
}

// GetName returns the validated profile name
func (w WizardName) GetName() string {
	return w.name
}
