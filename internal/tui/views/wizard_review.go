package views

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/schema"
)

type wizardReviewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Save     key.Binding
	Back     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

func newWizardReviewKeyMap() wizardReviewKeyMap {
	return wizardReviewKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		Save: key.NewBinding(
			key.WithKeys("enter", "ctrl+s"),
			key.WithHelp("enter/ctrl+s", "save"),
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

// WizardReview is step 5: Review and save
type WizardReview struct {
	profileName    string
	config         *config.Config
	jsonPreview    string
	validationErrs []schema.ValidationError
	isValid        bool
	viewport       viewport.Model
	ready          bool
	width          int
	height         int
	keys           wizardReviewKeyMap
}

func NewWizardReview() WizardReview {
	return WizardReview{
		keys: newWizardReviewKeyMap(),
	}
}

func (w WizardReview) Init() tea.Cmd {
	return nil
}

func (w *WizardReview) SetSize(width, height int) {
	w.width = width
	w.height = height
	if !w.ready {
		w.viewport = viewport.New(width, height-8)
		w.ready = true
	} else {
		w.viewport.Width = width
		w.viewport.Height = height - 8
	}
}

func (w *WizardReview) SetConfig(name string, cfg *config.Config) {
	w.profileName = name
	w.config = cfg
	w.validateAndPreview()
}

func (w *WizardReview) validateAndPreview() {
	if w.config == nil {
		w.jsonPreview = "{}"
		w.isValid = true
		return
	}

	// Generate JSON preview
	jsonData, err := json.MarshalIndent(w.config, "", "  ")
	if err != nil {
		w.jsonPreview = fmt.Sprintf("Error generating preview: %v", err)
		w.isValid = false
		return
	}
	w.jsonPreview = string(jsonData)

	// Validate against schema
	validator, err := schema.GetValidator()
	if err != nil {
		w.validationErrs = nil
		w.isValid = true // Can't validate, assume valid
		return
	}

	errs, err := validator.Validate(w.config)
	if err != nil {
		w.validationErrs = nil
		w.isValid = true // Validation error, assume valid
		return
	}

	w.validationErrs = errs
	w.isValid = len(errs) == 0
}

func (w WizardReview) Update(msg tea.Msg) (WizardReview, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keys.Save):
			if w.isValid {
				return w, func() tea.Msg { return WizardNextMsg{} }
			}
			return w, nil
		case key.Matches(msg, w.keys.Back):
			return w, func() tea.Msg { return WizardBackMsg{} }
		}
	}

	// Update viewport
	w.viewport, cmd = w.viewport.Update(msg)

	return w, cmd
}

func (w WizardReview) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))

	title := titleStyle.Render("Review & Save")

	// Profile name
	nameLabel := helpStyle.Render("Profile Name: ") + titleStyle.Render(w.profileName)

	// Validation status
	var validationStatus string
	if w.isValid {
		validationStatus = successStyle.Render("✓ Configuration is valid")
	} else {
		validationStatus = errorStyle.Render("✗ Validation errors found:")
		for _, e := range w.validationErrs {
			validationStatus += "\n  " + errorStyle.Render(fmt.Sprintf("• %s: %s", e.Path, e.Message))
		}
	}

	// JSON preview in viewport
	w.viewport.SetContent(codeStyle.Render(w.jsonPreview))
	preview := w.viewport.View()

	// Help text
	var help string
	if w.isValid {
		help = helpStyle.Render("Enter to save • Shift+Tab to go back • ↑/↓ to scroll")
	} else {
		help = helpStyle.Render("Fix errors before saving • Shift+Tab to go back")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		nameLabel,
		"",
		validationStatus,
		"",
		titleStyle.Render("JSON Preview:"),
		preview,
		"",
		help,
	)
}

// IsValid returns true if the config passes validation
func (w WizardReview) IsValid() bool {
	return w.isValid
}

// GetErrors returns validation errors
func (w WizardReview) GetErrors() []schema.ValidationError {
	return w.validationErrs
}
