package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
)

// Wizard message types
type WizardNextMsg struct{}
type WizardBackMsg struct{}
type WizardSaveMsg struct{ Profile *profile.Profile }
type WizardCancelMsg struct{}

// Internal message for async save completion
type wizardSaveDoneMsg struct {
	profile *profile.Profile
	err     error
}

const (
	StepName = iota + 1
	StepCategories
	StepAgents
	StepHooks
	StepOther
	StepReview
)

var stepNames = map[int]string{
	StepName:       "Name",
	StepCategories: "Categories",
	StepAgents:     "Agents",
	StepHooks:      "Hooks",
	StepOther:      "Other Settings",
	StepReview:     "Review & Save",
}

type wizardKeyMap struct {
	Next   key.Binding
	Back   key.Binding
	Cancel key.Binding
	Save   key.Binding
}

func newWizardKeyMap() wizardKeyMap {
	return wizardKeyMap{
		Next: key.NewBinding(
			key.WithKeys("tab", "enter"),
			key.WithHelp("tab/enter", "next"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab", "esc"),
			key.WithHelp("shift+tab/esc", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "cancel"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
	}
}

// Wizard is the main wizard struct for creating/editing profiles
type Wizard struct {
	step                int
	profileName         string
	originalProfileName string // for rename detection in edit mode
	config              config.Config
	editMode            bool // true when editing existing profile, false when creating new

	// Sub-views for each step
	nameStep       WizardName
	categoriesStep WizardCategories
	agentsStep     WizardAgents
	hooksStep      WizardHooks
	otherStep      WizardOther
	reviewStep     WizardReview

	width  int
	height int
	keys   wizardKeyMap
	err    error
}

// NewWizard creates a new wizard for creating a profile
func NewWizard() Wizard {
	return Wizard{
		step:           StepName,
		config:         config.Config{},
		editMode:       false,
		nameStep:       NewWizardName(),
		categoriesStep: NewWizardCategories(),
		agentsStep:     NewWizardAgents(),
		hooksStep:      NewWizardHooks(),
		otherStep:      NewWizardOther(),
		reviewStep:     NewWizardReview(),
		keys:           newWizardKeyMap(),
	}
}

// NewWizardForEdit creates a wizard for editing an existing profile
func NewWizardForEdit(p *profile.Profile) Wizard {
	w := NewWizard()
	w.editMode = true
	w.profileName = p.Name
	w.originalProfileName = p.Name
	w.config = p.Config
	w.nameStep.SetName(p.Name)
	w.categoriesStep.SetConfig(&w.config)
	w.agentsStep.SetConfig(&w.config)
	w.hooksStep.SetConfig(&w.config)
	w.otherStep.SetConfig(&w.config)
	w.reviewStep.SetConfig(p.Name, &w.config)
	return w
}

// NewWizardFromTemplate creates a wizard pre-populated with a template profile
func NewWizardFromTemplate(templateName string) (Wizard, error) {
	template, err := profile.Load(templateName)
	if err != nil {
		return Wizard{}, fmt.Errorf("template '%s' not found", templateName)
	}

	w := NewWizard()
	w.editMode = false
	w.config = template.Config
	w.categoriesStep.SetConfig(&w.config)
	w.agentsStep.SetConfig(&w.config)
	w.hooksStep.SetConfig(&w.config)
	w.otherStep.SetConfig(&w.config)
	return w, nil
}

func (w Wizard) Init() tea.Cmd {
	return tea.Batch(
		w.nameStep.Init(),
		w.categoriesStep.Init(),
		w.agentsStep.Init(),
		w.hooksStep.Init(),
		w.otherStep.Init(),
		w.reviewStep.Init(),
	)
}

func (w *Wizard) SetSize(width, height int) {
	w.width = width
	w.height = height
	// Reserve space for header/footer
	contentHeight := height - 6
	w.nameStep.SetSize(width, contentHeight)
	w.categoriesStep.SetSize(width, contentHeight)
	w.agentsStep.SetSize(width, contentHeight)
	w.hooksStep.SetSize(width, contentHeight)
	w.otherStep.SetSize(width, contentHeight)
	w.reviewStep.SetSize(width, contentHeight)
}

func (w Wizard) Update(msg tea.Msg) (Wizard, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case WizardNextMsg:
		return w.nextStep()

	case WizardBackMsg:
		return w.prevStep()

	case WizardCancelMsg:
		return w, func() tea.Msg { return NavigateToDashboardMsg{} }

	case WizardSaveMsg:
		return w, func() tea.Msg { return NavigateToDashboardMsg{} }

	case wizardSaveDoneMsg:
		if msg.err != nil {
			w.err = msg.err
			return w, nil
		}
		return w, func() tea.Msg { return WizardSaveMsg{Profile: msg.profile} }

	case tea.KeyMsg:
		// Global navigation keys
		if key.Matches(msg, w.keys.Cancel) {
			return w, func() tea.Msg { return WizardCancelMsg{} }
		}
	}

	// Delegate to current step
	switch w.step {
	case StepName:
		w.nameStep, cmd = w.nameStep.Update(msg)
		cmds = append(cmds, cmd)
		// Check for step completion
		if w.nameStep.IsComplete() {
			w.profileName = w.nameStep.GetName()
		}

	case StepCategories:
		w.categoriesStep, cmd = w.categoriesStep.Update(msg)
		cmds = append(cmds, cmd)

	case StepAgents:
		w.agentsStep, cmd = w.agentsStep.Update(msg)
		cmds = append(cmds, cmd)

	case StepHooks:
		w.hooksStep, cmd = w.hooksStep.Update(msg)
		cmds = append(cmds, cmd)

	case StepOther:
		w.otherStep, cmd = w.otherStep.Update(msg)
		cmds = append(cmds, cmd)

	case StepReview:
		w.reviewStep, cmd = w.reviewStep.Update(msg)
		cmds = append(cmds, cmd)
	}

	return w, tea.Batch(cmds...)
}

func (w Wizard) nextStep() (Wizard, tea.Cmd) {
	switch w.step {
	case StepName:
		if !w.nameStep.IsComplete() {
			return w, nil
		}
		w.profileName = w.nameStep.GetName()
		w.step = StepCategories
		w.categoriesStep.SetConfig(&w.config)
		return w, w.categoriesStep.Init()

	case StepCategories:
		w.categoriesStep.Apply(&w.config)
		w.step = StepAgents
		w.agentsStep.SetConfig(&w.config)
		return w, w.agentsStep.Init()

	case StepAgents:
		w.agentsStep.Apply(&w.config)
		w.step = StepHooks
		w.hooksStep.SetConfig(&w.config)
		return w, w.hooksStep.Init()

	case StepHooks:
		w.hooksStep.Apply(&w.config)
		w.step = StepOther
		w.otherStep.SetConfig(&w.config)
		return w, w.otherStep.Init()

	case StepOther:
		w.otherStep.Apply(&w.config)
		w.step = StepReview
		w.reviewStep.SetConfig(w.profileName, &w.config)
		return w, w.reviewStep.Init()

	case StepReview:
		// Validate config against schema FIRST
		validator, err := schema.GetValidator()
		if err != nil {
			w.err = fmt.Errorf("validator error: %w", err)
			return w, nil
		}

		validationErrors, err := validator.Validate(&w.config)
		if err != nil {
			w.err = fmt.Errorf("validation error: %w", err)
			return w, nil
		}

		if len(validationErrors) > 0 {
			// Format first validation error for display
			w.err = fmt.Errorf("validation failed: %s", validationErrors[0].Error())
			return w, nil
		}

		// Check for rename in edit mode
		if w.editMode && w.profileName != w.originalProfileName {
			// Check if new name already exists
			if profile.Exists(w.profileName) {
				w.err = fmt.Errorf("profile '%s' already exists", w.profileName)
				return w, nil
			}
		}

		// Capture values for async closure
		p := &profile.Profile{
			Name:   w.profileName,
			Config: w.config,
		}
		editMode := w.editMode
		originalName := w.originalProfileName

		return w, func() tea.Msg {
			if err := profile.Save(p); err != nil {
				return wizardSaveDoneMsg{err: err}
			}
			if editMode && p.Name != originalName {
				_ = profile.Delete(originalName)
			}
			return wizardSaveDoneMsg{profile: p}
		}
	}
	return w, nil
}

// prevStep navigates to the previous wizard step WITHOUT calling Apply()
// on the current step. This means any unsaved edits in the current step
// are intentionally discarded. When the user navigates forward again,
// SetConfig() will reload from the canonical w.config, restoring the
// last-applied state. This design gives users a natural "undo" mechanism:
// going back discards uncommitted changes in the current step.
func (w Wizard) prevStep() (Wizard, tea.Cmd) {
	switch w.step {
	case StepName:
		return w, func() tea.Msg { return WizardCancelMsg{} }
	case StepCategories:
		w.step = StepName
		return w, w.nameStep.Init()
	case StepAgents:
		w.step = StepCategories
		return w, w.categoriesStep.Init()
	case StepHooks:
		w.step = StepAgents
		return w, w.agentsStep.Init()
	case StepOther:
		w.step = StepHooks
		return w, w.hooksStep.Init()
	case StepReview:
		w.step = StepOther
		return w, w.otherStep.Init()
	}
	return w, nil
}

func (w Wizard) View() string {
	// Header with progress indicator
	header := w.renderHeader()

	// Current step content
	var content string
	switch w.step {
	case StepName:
		content = w.nameStep.View()
	case StepCategories:
		content = w.categoriesStep.View()
	case StepAgents:
		content = w.agentsStep.View()
	case StepHooks:
		content = w.hooksStep.View()
	case StepOther:
		content = w.otherStep.View()
	case StepReview:
		content = w.reviewStep.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
	)
}

func (w Wizard) renderHeader() string {
	// Progress bar
	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	var steps []string
	for i := StepName; i <= StepReview; i++ {
		stepNum := lipgloss.NewStyle().Bold(i == w.step)
		if i == w.step {
			steps = append(steps, progressStyle.Render(stepNum.Render(stepNames[i])))
		} else if i < w.step {
			steps = append(steps, successStyle.Render("✓ "+stepNames[i]))
		} else {
			steps = append(steps, dimStyle.Render(stepNames[i]))
		}
	}

	progress := lipgloss.JoinHorizontal(lipgloss.Top,
		steps[0], dimStyle.Render(" → "),
		steps[1], dimStyle.Render(" → "),
		steps[2], dimStyle.Render(" → "),
		steps[3], dimStyle.Render(" → "),
		steps[4], dimStyle.Render(" → "),
		steps[5],
	)

	title := titleStyle.Render("Create Profile")
	if w.editMode {
		title = titleStyle.Render("Edit Profile")
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, progress)
}

// GetProfile returns the built profile
func (w Wizard) GetProfile() *profile.Profile {
	return &profile.Profile{
		Name:   w.profileName,
		Config: w.config,
	}
}

// IsEditMode returns true if editing an existing profile
func (w Wizard) IsEditMode() bool {
	return w.editMode
}
