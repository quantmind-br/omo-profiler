package views

import tea "github.com/charmbracelet/bubbletea"

// WizardStep defines the interface that all wizard steps must implement.
// Note: SetConfig/Apply are NOT part of this interface because WizardName
// uses SetName/GetName instead. The Wizard orchestrator calls these
// methods with concrete types.
type WizardStep interface {
	Init() tea.Cmd
	SetSize(width, height int)
	View() string
}

// Compile-time interface compliance checks
var (
	_ WizardStep = (*WizardName)(nil)
	_ WizardStep = (*WizardCategories)(nil)
	_ WizardStep = (*WizardAgents)(nil)
	_ WizardStep = (*WizardHooks)(nil)
	_ WizardStep = (*WizardOther)(nil)
	_ WizardStep = (*WizardReview)(nil)
)
