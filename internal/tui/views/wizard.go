package views

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
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

var stepAbbreviations = map[int]string{
	StepName:       "N",
	StepCategories: "C",
	StepAgents:     "A",
	StepHooks:      "H",
	StepOther:      "O",
	StepReview:     "R",
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
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c/esc", "cancel"),
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
	selection           *profile.FieldSelection
	preservedUnknown    map[string]json.RawMessage
	editMode            bool // true when editing existing profile, false when creating new
	returnToList        bool // true when wizard was entered from list view

	// Sub-views for each step
	nameStep       WizardName
	categoriesStep WizardCategories
	agentsStep     WizardAgents
	hooksStep      WizardHooks
	otherStep      WizardOther
	reviewStep     WizardReview

	width         int
	height        int
	keys          wizardKeyMap
	err           error
	flashMsg      string
	confirmCancel bool // awaiting y/N confirmation to discard the wizard
}

// NewWizard creates a new wizard for creating a profile
func NewWizard() Wizard {
	otherStep := NewWizardOther()
	cfg := config.Config{}
	selection := profile.NewBlankSelection()
	otherStep.Apply(&cfg, selection)

	return Wizard{
		step:           StepName,
		config:         cfg,
		selection:      selection,
		editMode:       false,
		nameStep:       NewWizardName(),
		categoriesStep: NewWizardCategories(),
		agentsStep:     NewWizardAgents(),
		hooksStep:      NewWizardHooks(),
		otherStep:      otherStep,
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
	w.selection = profile.NewSelectionFromPresence(p.FieldPresence)
	w.preservedUnknown = p.PreservedUnknown
	w.nameStep.SetName(p.Name)
	w.categoriesStep.SetConfig(&w.config, w.selection)
	w.agentsStep.SetConfig(&w.config, w.selection)
	w.hooksStep.SetConfig(&w.config, w.selection)
	w.otherStep.SetConfig(&w.config, w.selection)
	w.reviewStep.SetConfig(p.Name, &w.config, w.selection, w.preservedUnknown)
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
	w.selection = profile.NewSelectionFromPresence(template.FieldPresence)
	w.preservedUnknown = template.PreservedUnknown
	w.categoriesStep.SetConfig(&w.config, w.selection)
	w.agentsStep.SetConfig(&w.config, w.selection)
	w.hooksStep.SetConfig(&w.config, w.selection)
	w.otherStep.SetConfig(&w.config, w.selection)
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
	headerLines := 4
	if layout.IsCompact(width) {
		headerLines = 3
	}
	if layout.IsShort(height) {
		headerLines--
		if headerLines < 1 {
			headerLines = 1
		}
	}
	contentHeight := height - headerLines
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
		if w.returnToList {
			return w, func() tea.Msg { return NavigateToListMsg{} }
		}
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
		w.flashMsg = ""
		// While confirming a discard, intercept y/N before anything else.
		if w.confirmCancel {
			switch msg.String() {
			case "y", "Y":
				w.confirmCancel = false
				return w, func() tea.Msg { return WizardCancelMsg{} }
			default:
				// n / N / esc / any other key resumes editing.
				w.confirmCancel = false
				return w, nil
			}
		}
		if key.Matches(msg, w.keys.Cancel) {
			// Guard against discarding unsaved work; cancel immediately if nothing entered.
			if w.hasUnsavedInput() {
				w.confirmCancel = true
				return w, nil
			}
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
		w.categoriesStep.SetConfig(&w.config, w.selection)
		return w, w.categoriesStep.Init()

	case StepCategories:
		w.categoriesStep.Apply(&w.config, w.selection)
		w.step = StepAgents
		w.agentsStep.SetConfig(&w.config, w.selection)
		return w, w.agentsStep.Init()

	case StepAgents:
		w.agentsStep.Apply(&w.config, w.selection)
		w.step = StepHooks
		w.hooksStep.SetConfig(&w.config, w.selection)
		return w, w.hooksStep.Init()

	case StepHooks:
		w.hooksStep.Apply(&w.config, w.selection)
		w.step = StepOther
		w.otherStep.SetConfig(&w.config, w.selection)
		return w, w.otherStep.Init()

	case StepOther:
		w.otherStep.Apply(&w.config, w.selection)
		w.step = StepReview
		w.reviewStep.SetConfig(w.profileName, &w.config, w.selection, w.preservedUnknown)
		return w, w.reviewStep.Init()

	case StepReview:
		// Validate config against schema FIRST
		validator, err := schema.GetValidator()
		if err != nil {
			w.err = fmt.Errorf("validator error: %w", err)
			return w, nil
		}

		validationErrors, err := validator.ValidateForSave(&w.config)
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
		profileName := w.profileName
		cfg := w.config
		selection := w.selection
		preservedUnknown := w.preservedUnknown
		editMode := w.editMode
		originalName := w.originalProfileName

		return w, func() tea.Msg {
			data, err := profile.MarshalSparse(&cfg, selection, preservedUnknown)
			if err != nil {
				return wizardSaveDoneMsg{err: err}
			}

			validator, err := schema.GetValidator()
			if err != nil {
				return wizardSaveDoneMsg{err: err}
			}

			validationErrors, err := validator.ValidateJSONForSave(data)
			if err != nil {
				return wizardSaveDoneMsg{err: err}
			}
			if len(validationErrors) > 0 {
				return wizardSaveDoneMsg{err: fmt.Errorf("validation failed: %s", validationErrors[0].Error())}
			}

			if err := config.EnsureDirs(); err != nil {
				return wizardSaveDoneMsg{err: err}
			}

			path := filepath.Join(config.ProfilesDir(), profileName+".json")
			if err := os.WriteFile(path, data, 0644); err != nil {
				return wizardSaveDoneMsg{err: err}
			}

			p := &profile.Profile{
				Name:             profileName,
				Config:           cfg,
				Path:             path,
				PreservedUnknown: preservedUnknown,
			}
			if editMode && p.Name != originalName {
				_ = profile.Delete(originalName)
			}
			return wizardSaveDoneMsg{profile: p}
		}
	}
	return w, nil
}

func (w Wizard) prevStep() (Wizard, tea.Cmd) {
	switch w.step {
	case StepCategories:
		w.categoriesStep.Apply(&w.config, w.selection)
		w.flashMsg = "Changes applied"
	case StepAgents:
		w.agentsStep.Apply(&w.config, w.selection)
		w.flashMsg = "Changes applied"
	case StepHooks:
		w.hooksStep.Apply(&w.config, w.selection)
		w.flashMsg = "Changes applied"
	case StepOther:
		w.otherStep.Apply(&w.config, w.selection)
		w.flashMsg = "Changes applied"
	}

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
	header := w.renderHeader()

	if w.confirmCancel {
		target := w.profileName
		if target == "" {
			target = w.nameStep.GetName()
		}
		if target == "" {
			target = "this profile"
		}
		dialog := layout.RenderConfirmDialog(target, "Discard")
		hint := wizNameDescStyle.Render("Your changes will be lost. Press y to discard, any other key to keep editing.")
		body := lipgloss.JoinVertical(lipgloss.Left, dialog, "", hint)
		if w.width > 0 && w.height > 0 {
			return lipgloss.Place(w.width, w.height, lipgloss.Center, lipgloss.Center, body)
		}
		return body
	}

	errorDisplay := ""
	if w.err != nil {
		errorDisplay = errorStyle.Render("⚠ " + w.err.Error())
	}

	flashDisplay := ""
	if w.flashMsg != "" {
		flashDisplay = successStyle.Render("✓ " + w.flashMsg)
	}

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

	if layout.IsShort(w.height) {
		if w.err != nil {
			return lipgloss.JoinVertical(lipgloss.Left, header, errorDisplay, content)
		}
		if w.flashMsg != "" {
			return lipgloss.JoinVertical(lipgloss.Left, header, flashDisplay, content)
		}
		return lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	if w.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, errorDisplay, "", content)
	}
	if w.flashMsg != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, flashDisplay, "", content)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, "", content)
}

func (w Wizard) renderHeader() string {
	title := titleStyle.Render("Create Profile")
	if w.editMode {
		title = titleStyle.Render("Edit Profile")
	}

	totalSteps := StepReview
	progressText := fmt.Sprintf("Step %d of %d", w.step, totalSteps)

	if layout.IsCompact(w.width) {
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
		var dots string
		for i := StepName; i <= StepReview; i++ {
			if i < w.step {
				dots += successStyle.Render("●")
			} else if i == w.step {
				dots += progressStyle.Render("●")
			} else {
				dots += dimStyle.Render("○")
			}
		}
		dots += " " + progressStyle.Render(stepNames[w.step]) + " " + dimStyle.Render(progressText)
		if layout.IsShort(w.height) {
			return dots
		}
		return lipgloss.JoinVertical(lipgloss.Left, title, dots)
	}

	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	// Use abbreviated names at narrow widths to avoid flickering between
	// full names and abbreviations as the step (and checkmark prefix) changes.
	useAbbrev := w.width <= 100

	var steps []string
	for i := StepName; i <= StepReview; i++ {
		stepNum := lipgloss.NewStyle().Bold(i == w.step)
		if useAbbrev {
			if i == w.step {
				steps = append(steps, progressStyle.Render(stepAbbreviations[i]))
			} else if i < w.step {
				steps = append(steps, successStyle.Render(stepAbbreviations[i]))
			} else {
				steps = append(steps, dimStyle.Render(stepAbbreviations[i]))
			}
		} else {
			if i == w.step {
				steps = append(steps, progressStyle.Render(stepNum.Render(stepNames[i])))
			} else if i < w.step {
				steps = append(steps, successStyle.Render("✓ "+stepNames[i]))
			} else {
				steps = append(steps, dimStyle.Render(stepNames[i]))
			}
		}
	}

	var sep string
	if useAbbrev {
		sep = " > "
	} else {
		sep = " → "
	}

	progress := lipgloss.JoinHorizontal(lipgloss.Top,
		steps[0], dimStyle.Render(sep),
		steps[1], dimStyle.Render(sep),
		steps[2], dimStyle.Render(sep),
		steps[3], dimStyle.Render(sep),
		steps[4], dimStyle.Render(sep),
		steps[5],
	)

	headerLine := lipgloss.JoinHorizontal(lipgloss.Top, progress, "  ", dimStyle.Render(progressText))

	// If still too wide, try abbreviated names (fallback for edge cases)
	if !useAbbrev && lipgloss.Width(headerLine) > w.width {
		var abbrSteps []string
		for i := StepName; i <= StepReview; i++ {
			if i == w.step {
				abbrSteps = append(abbrSteps, progressStyle.Render(stepAbbreviations[i]))
			} else if i < w.step {
				abbrSteps = append(abbrSteps, successStyle.Render(stepAbbreviations[i]))
			} else {
				abbrSteps = append(abbrSteps, dimStyle.Render(stepAbbreviations[i]))
			}
		}
		progress = lipgloss.JoinHorizontal(lipgloss.Top,
			abbrSteps[0], dimStyle.Render(" > "),
			abbrSteps[1], dimStyle.Render(" > "),
			abbrSteps[2], dimStyle.Render(" > "),
			abbrSteps[3], dimStyle.Render(" > "),
			abbrSteps[4], dimStyle.Render(" > "),
			abbrSteps[5],
		)
		headerLine = lipgloss.JoinHorizontal(lipgloss.Top, progress, "  ", dimStyle.Render(progressText))
	}

	// If still too wide, drop intermediates keeping first, current, last
	if lipgloss.Width(headerLine) > w.width {
		var minimal []string
		for i := StepName; i <= StepReview; i++ {
			if i == StepName || i == StepReview || i == w.step {
				if i == w.step {
					minimal = append(minimal, progressStyle.Render(stepAbbreviations[i]))
				} else if i < w.step {
					minimal = append(minimal, successStyle.Render(stepAbbreviations[i]))
				} else {
					minimal = append(minimal, dimStyle.Render(stepAbbreviations[i]))
				}
			}
		}
		progress = lipgloss.JoinHorizontal(lipgloss.Top,
			minimal[0], dimStyle.Render(" … "), minimal[1], dimStyle.Render(" … "), minimal[2],
		)
		headerLine = lipgloss.JoinHorizontal(lipgloss.Top, progress, "  ", dimStyle.Render(progressText))
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, headerLine)
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

// SetReturnToList sets the flag so cancel returns to list instead of dashboard
func (w *Wizard) SetReturnToList() {
	w.returnToList = true
}

// IsReviewStep returns true when on the final review step
func (w Wizard) IsReviewStep() bool {
	return w.step == StepReview
}

// IsConfirmingCancel reports whether the discard-confirmation prompt is showing.
func (w Wizard) IsConfirmingCancel() bool {
	return w.confirmCancel
}

// hasUnsavedInput reports whether cancelling now would discard user work.
// In edit mode there is always existing content to protect; otherwise it's
// "unsaved" once the user advances past step 1 or types a profile name.
func (w Wizard) hasUnsavedInput() bool {
	if w.editMode {
		return true
	}
	return w.step > StepName || w.nameStep.HasInput()
}
