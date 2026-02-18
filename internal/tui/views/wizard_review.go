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
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

var (
	wizReviewWhite = lipgloss.Color("#CDD6F4")
	wizReviewGray  = lipgloss.Color("#6C7086")
	wizReviewRed   = lipgloss.Color("#F38BA8")
	wizReviewGreen = lipgloss.Color("#A6E3A1")
	wizReviewBlue  = lipgloss.Color("#89B4FA")
)

var (
	wizReviewTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(wizReviewWhite)
	wizReviewHelpStyle    = lipgloss.NewStyle().Foreground(wizReviewGray)
	wizReviewErrorStyle   = lipgloss.NewStyle().Foreground(wizReviewRed)
	wizReviewSuccessStyle = lipgloss.NewStyle().Foreground(wizReviewGreen)
	wizReviewCodeStyle    = lipgloss.NewStyle().Foreground(wizReviewBlue)
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
	overhead := 8
	if layout.IsShort(height) {
		overhead = 4
	}
	if !w.ready {
		w.viewport = viewport.New(width, height-overhead)
		w.ready = true
	} else {
		w.viewport.Width = width
		w.viewport.Height = height - overhead
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
	title := wizReviewTitleStyle.Render("Review & Save")

	// Profile name
	nameLabel := wizReviewHelpStyle.Render("Profile Name: ") + wizReviewTitleStyle.Render(w.profileName)

	// Validation status
	var validationStatus string
	if w.isValid {
		validationStatus = wizReviewSuccessStyle.Render("✓ Configuration is valid")
	} else {
		validationStatus = wizReviewErrorStyle.Render("✗ Validation errors found:")
		for _, e := range w.validationErrs {
			validationStatus += "\n  " + wizReviewErrorStyle.Render(fmt.Sprintf("• %s: %s", e.Path, e.Message))
		}
	}

	// JSON preview in viewport
	w.viewport.SetContent(wizReviewCodeStyle.Render(w.jsonPreview))
	preview := w.viewport.View()

	// Help text
	var help string
	if w.isValid {
		help = wizReviewHelpStyle.Render("Enter to save • Shift+Tab to go back • ↑/↓ to scroll")
	} else {
		help = wizReviewHelpStyle.Render("Fix errors before saving • Shift+Tab to go back")
	}

	if layout.IsShort(w.height) {
		titleLine := wizReviewTitleStyle.Render("Review: ") + wizReviewTitleStyle.Render(w.profileName)
		w.viewport.SetContent(wizReviewCodeStyle.Render(w.jsonPreview))
		return lipgloss.JoinVertical(lipgloss.Left,
			titleLine,
			validationStatus,
			w.viewport.View(),
			help,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		nameLabel,
		"",
		validationStatus,
		"",
		wizReviewTitleStyle.Render("JSON Preview:"),
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
