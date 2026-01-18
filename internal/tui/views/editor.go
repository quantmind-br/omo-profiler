package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
)

// Message types
type EditorLoadMsg struct{ Profile *profile.Profile }
type EditorLoadErrorMsg struct{ Err error }
type EditorSaveMsg struct{}
type EditorSaveSuccessMsg struct{}
type EditorSaveErrorMsg struct{ Err error }
type EditorCancelMsg struct{}
type EditorSectionChangeMsg struct{ Section int }
type EditorValidationMsg struct{ Errors []schema.ValidationError }

// Editor sections
const (
	sectionName = iota
	sectionAgents
	sectionHooks
	sectionDisabled
	sectionOther
	sectionSkills
	sectionReview
)

var sectionNames = []string{
	"Name",
	"Agents",
	"Hooks",
	"Disabled",
	"Other",
	"Skills",
	"Review",
}

// Focus state
type editorFocus int

const (
	focusSidebar editorFocus = iota
	focusContent
)

// Styles
var (
	editorPurple  = lipgloss.Color("#7D56F4")
	editorMagenta = lipgloss.Color("#FF6AC1")
	editorGreen   = lipgloss.Color("#A6E3A1")
	editorRed     = lipgloss.Color("#F38BA8")
	editorYellow  = lipgloss.Color("#F9E2AF")
	editorGray    = lipgloss.Color("#6C7086")
	editorWhite   = lipgloss.Color("#CDD6F4")
)

var (
	editorTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(editorPurple)
	editorSubtitleStyle = lipgloss.NewStyle().Foreground(editorGray)
	editorErrorStyle    = lipgloss.NewStyle().Foreground(editorRed)
	editorSuccessStyle  = lipgloss.NewStyle().Foreground(editorGreen)
	editorHelpStyle     = lipgloss.NewStyle().Foreground(editorGray)
	editorActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(editorWhite).Background(editorPurple)
	editorInactiveStyle = lipgloss.NewStyle().Foreground(editorWhite)
	editorAccentStyle   = lipgloss.NewStyle().Foreground(editorMagenta)
	editorWarningStyle  = lipgloss.NewStyle().Foreground(editorYellow)
)

type editorKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Tab    key.Binding
	Enter  key.Binding
	Escape key.Binding
	Save   key.Binding
	Toggle key.Binding
}

func newEditorKeyMap() editorKeyMap {
	return editorKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
	}
}

type Editor struct {
	// Profile data
	originalProfile *profile.Profile
	workingProfile  *profile.Profile
	profileName     string

	// UI state
	section int
	focus   editorFocus
	width   int
	height  int
	ready   bool
	keys    editorKeyMap

	// Components
	nameInput textinput.Model
	viewport  viewport.Model

	// Content navigation
	contentCursor int
	contentItems  []string

	// Status
	loading       bool
	saving        bool
	err           error
	successMsg    string
	validationErr []schema.ValidationError

	// Hooks toggle state (which hooks are enabled)
	hooksEnabled map[string]bool

	// Available options
	availableAgents []string
	availableHooks  []string
}

func NewEditor(profileName string) Editor {
	ti := textinput.New()
	ti.Placeholder = "profile-name"
	ti.CharLimit = 64

	return Editor{
		profileName:     profileName,
		section:         sectionName,
		focus:           focusSidebar,
		keys:            newEditorKeyMap(),
		nameInput:       ti,
		loading:         true,
		hooksEnabled:    make(map[string]bool),
		availableAgents: []string{"build", "oracle", "explore", "librarian"},
		availableHooks:  []string{"pre-tool", "post-tool", "pre-response", "notification"},
	}
}

func (e Editor) Init() tea.Cmd {
	return e.loadProfile
}

func (e Editor) loadProfile() tea.Msg {
	p, err := profile.Load(e.profileName)
	if err != nil {
		return EditorLoadErrorMsg{Err: err}
	}
	return EditorLoadMsg{Profile: p}
}

func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height

	contentWidth := width - 20 // sidebar width
	contentHeight := height - 10

	if contentHeight < 1 {
		contentHeight = 1
	}

	if !e.ready {
		e.viewport = viewport.New(contentWidth, contentHeight)
		e.ready = true
	} else {
		e.viewport.Width = contentWidth
		e.viewport.Height = contentHeight
	}

	e.nameInput.Width = contentWidth - 4
}

func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)
		return e, nil

	case EditorLoadMsg:
		e.loading = false
		e.originalProfile = msg.Profile
		// Deep copy for working profile via JSON to avoid shared references
		e.workingProfile = &profile.Profile{
			Name: msg.Profile.Name,
			Path: msg.Profile.Path,
		}
		// Deep copy the config
		configBytes, err := json.Marshal(msg.Profile.Config)
		if err == nil {
			json.Unmarshal(configBytes, &e.workingProfile.Config)
		} else {
			// Fallback to shallow copy if marshal fails
			e.workingProfile.Config = msg.Profile.Config
		}
		e.nameInput.SetValue(msg.Profile.Name)
		e.initHooksState()
		e.updateContentItems()
		return e, nil

	case EditorLoadErrorMsg:
		e.loading = false
		e.err = msg.Err
		return e, nil

	case EditorSaveSuccessMsg:
		e.saving = false
		e.successMsg = "Profile saved successfully!"
		e.err = nil
		e.validationErr = nil
		return e, nil

	case EditorSaveErrorMsg:
		e.saving = false
		e.err = msg.Err
		return e, nil

	case EditorValidationMsg:
		e.saving = false
		e.validationErr = msg.Errors
		return e, nil

	case tea.KeyMsg:
		// Clear success message on any key
		e.successMsg = ""

		// Handle save shortcut globally
		if key.Matches(msg, e.keys.Save) {
			return e, e.saveProfile
		}

		if e.focus == focusSidebar {
			return e.handleSidebarKeys(msg)
		}
		return e.handleContentKeys(msg)
	}

	// Update sub-components
	if e.focus == focusContent && e.section == sectionName {
		var cmd tea.Cmd
		e.nameInput, cmd = e.nameInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if e.ready {
		var cmd tea.Cmd
		e.viewport, cmd = e.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return e, tea.Batch(cmds...)
}

func (e Editor) handleSidebarKeys(msg tea.KeyMsg) (Editor, tea.Cmd) {
	switch {
	case key.Matches(msg, e.keys.Up):
		if e.section > 0 {
			e.section--
			e.contentCursor = 0
			e.updateContentItems()
		}
	case key.Matches(msg, e.keys.Down):
		if e.section < len(sectionNames)-1 {
			e.section++
			e.contentCursor = 0
			e.updateContentItems()
		}
	case key.Matches(msg, e.keys.Tab), key.Matches(msg, e.keys.Enter):
		e.focus = focusContent
		e.contentCursor = 0
		if e.section == sectionName {
			e.nameInput.Focus()
			return e, textinput.Blink
		}
	case key.Matches(msg, e.keys.Escape):
		return e, func() tea.Msg { return EditorCancelMsg{} }
	}
	return e, nil
}

func (e Editor) handleContentKeys(msg tea.KeyMsg) (Editor, tea.Cmd) {
	switch {
	case key.Matches(msg, e.keys.Escape):
		e.focus = focusSidebar
		e.nameInput.Blur()
		return e, nil

	case key.Matches(msg, e.keys.Tab):
		e.focus = focusSidebar
		e.nameInput.Blur()
		return e, nil

	case key.Matches(msg, e.keys.Up):
		if e.contentCursor > 0 {
			e.contentCursor--
		}

	case key.Matches(msg, e.keys.Down):
		if e.contentCursor < len(e.contentItems)-1 {
			e.contentCursor++
		}

	case key.Matches(msg, e.keys.Toggle):
		// Space only toggles in hooks and disabled sections
		if e.section == sectionHooks || e.section == sectionDisabled {
			return e.handleContentAction()
		}

	case key.Matches(msg, e.keys.Enter):
		// Enter for save in review section
		if e.section == sectionReview {
			return e.handleContentAction()
		}
	}

	// Handle text input for name section
	if e.section == sectionName {
		var cmd tea.Cmd
		e.nameInput, cmd = e.nameInput.Update(msg)
		if e.workingProfile != nil {
			e.workingProfile.Name = e.nameInput.Value()
		}
		return e, cmd
	}

	return e, nil
}

func (e Editor) handleContentAction() (Editor, tea.Cmd) {
	if e.workingProfile == nil {
		return e, nil
	}

	switch e.section {
	case sectionHooks:
		if e.contentCursor < len(e.availableHooks) {
			hook := e.availableHooks[e.contentCursor]
			e.hooksEnabled[hook] = !e.hooksEnabled[hook]
			e.updateDisabledHooks()
		}

	case sectionDisabled:
		e.toggleDisabledItem()

	case sectionReview:
		// Save on enter in review section
		return e, e.saveProfile
	}

	return e, nil
}

func (e *Editor) initHooksState() {
	if e.workingProfile == nil {
		return
	}

	// All hooks enabled by default
	for _, h := range e.availableHooks {
		e.hooksEnabled[h] = true
	}

	// Mark disabled hooks
	for _, h := range e.workingProfile.Config.DisabledHooks {
		e.hooksEnabled[h] = false
	}
}

func (e *Editor) updateDisabledHooks() {
	if e.workingProfile == nil {
		return
	}

	var disabled []string
	for _, h := range e.availableHooks {
		if !e.hooksEnabled[h] {
			disabled = append(disabled, h)
		}
	}
	e.workingProfile.Config.DisabledHooks = disabled
}

func (e *Editor) toggleDisabledItem() {
	if e.workingProfile == nil || e.contentCursor >= len(e.contentItems) {
		return
	}

	item := e.contentItems[e.contentCursor]
	// Parse the item to determine type and value
	parts := strings.SplitN(item, ": ", 2)
	if len(parts) != 2 {
		return
	}

	category := parts[0]
	values := strings.Split(parts[1], ", ")
	if len(values) == 0 || values[0] == "(none)" {
		return
	}

	// Toggle first item in the list (simplified)
	switch category {
	case "MCPs":
		if len(e.workingProfile.Config.DisabledMCPs) > 0 {
			e.workingProfile.Config.DisabledMCPs = e.workingProfile.Config.DisabledMCPs[1:]
		}
	case "Agents":
		if len(e.workingProfile.Config.DisabledAgents) > 0 {
			e.workingProfile.Config.DisabledAgents = e.workingProfile.Config.DisabledAgents[1:]
		}
	case "Skills":
		if len(e.workingProfile.Config.DisabledSkills) > 0 {
			e.workingProfile.Config.DisabledSkills = e.workingProfile.Config.DisabledSkills[1:]
		}
	case "Commands":
		if len(e.workingProfile.Config.DisabledCommands) > 0 {
			e.workingProfile.Config.DisabledCommands = e.workingProfile.Config.DisabledCommands[1:]
		}
	}
	e.updateContentItems()
}

func (e *Editor) updateContentItems() {
	if e.workingProfile == nil {
		e.contentItems = []string{}
		return
	}

	switch e.section {
	case sectionName:
		e.contentItems = []string{"Profile Name"}

	case sectionAgents:
		e.contentItems = []string{}
		for name := range e.workingProfile.Config.Agents {
			e.contentItems = append(e.contentItems, name)
		}
		if len(e.contentItems) == 0 {
			e.contentItems = []string{"(no agents configured)"}
		}

	case sectionHooks:
		e.contentItems = e.availableHooks

	case sectionDisabled:
		e.contentItems = []string{
			fmt.Sprintf("MCPs: %s", formatList(e.workingProfile.Config.DisabledMCPs)),
			fmt.Sprintf("Agents: %s", formatList(e.workingProfile.Config.DisabledAgents)),
			fmt.Sprintf("Skills: %s", formatList(e.workingProfile.Config.DisabledSkills)),
			fmt.Sprintf("Commands: %s", formatList(e.workingProfile.Config.DisabledCommands)),
		}

	case sectionOther:
		e.contentItems = []string{
			fmt.Sprintf("Auto Update: %s", formatBoolPtr(e.workingProfile.Config.AutoUpdate)),
		}
		if e.workingProfile.Config.ClaudeCode != nil {
			e.contentItems = append(e.contentItems,
				fmt.Sprintf("MCP: %s", formatBoolPtr(e.workingProfile.Config.ClaudeCode.MCP)),
				fmt.Sprintf("Commands: %s", formatBoolPtr(e.workingProfile.Config.ClaudeCode.Commands)),
				fmt.Sprintf("Skills: %s", formatBoolPtr(e.workingProfile.Config.ClaudeCode.Skills)),
				fmt.Sprintf("Agents: %s", formatBoolPtr(e.workingProfile.Config.ClaudeCode.Agents)),
				fmt.Sprintf("Hooks: %s", formatBoolPtr(e.workingProfile.Config.ClaudeCode.Hooks)),
			)
		}

	case sectionSkills:
		if e.workingProfile.Config.Skills != nil {
			e.contentItems = []string{"(JSON data - view only)"}
		} else {
			e.contentItems = []string{"(no skills configured)"}
		}

	case sectionReview:
		e.contentItems = []string{"Save Profile"}
	}
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return strings.Join(items, ", ")
}

func formatBoolPtr(b *bool) string {
	if b == nil {
		return "(default)"
	}
	if *b {
		return "enabled"
	}
	return "disabled"
}

func (e Editor) saveProfile() tea.Msg {
	if e.workingProfile == nil {
		return EditorSaveErrorMsg{Err: fmt.Errorf("no profile loaded")}
	}

	// Validate name
	if err := profile.ValidateName(e.workingProfile.Name); err != nil {
		return EditorSaveErrorMsg{Err: fmt.Errorf("invalid name: %w", err)}
	}

	// Validate against schema
	validator, err := schema.GetValidator()
	if err != nil {
		return EditorSaveErrorMsg{Err: fmt.Errorf("validator error: %w", err)}
	}

	validationErrors, err := validator.Validate(&e.workingProfile.Config)
	if err != nil {
		return EditorSaveErrorMsg{Err: fmt.Errorf("validation error: %w", err)}
	}

	if len(validationErrors) > 0 {
		return EditorValidationMsg{Errors: validationErrors}
	}

	// Handle rename - save new profile FIRST, then delete old
	isRename := e.workingProfile.Name != e.originalProfile.Name

	// Save profile
	if err := e.workingProfile.Save(); err != nil {
		return EditorSaveErrorMsg{Err: err}
	}

	// Only delete old profile after successful save
	if isRename {
		if err := profile.Delete(e.originalProfile.Name); err != nil {
			// Log but don't fail - new profile is already saved
			// The old profile will remain as orphan but no data is lost
		}
	}

	return EditorSaveSuccessMsg{}
}

func (e Editor) View() string {
	if e.loading {
		return editorSubtitleStyle.Render("Loading profile...")
	}

	if e.err != nil && e.workingProfile == nil {
		return editorErrorStyle.Render(fmt.Sprintf("Error: %v", e.err))
	}

	title := editorTitleStyle.Render(fmt.Sprintf("Edit Profile: %s", e.profileName))

	// Build sidebar
	sidebar := e.renderSidebar()

	// Build content
	content := e.renderContent()

	// Layout
	sidebarWidth := 16
	sidebarBorder := lipgloss.RoundedBorder()
	sidebarStyle := lipgloss.NewStyle().
		Border(sidebarBorder).
		BorderForeground(e.borderColor(focusSidebar)).
		Width(sidebarWidth).
		Padding(0, 1)

	contentBorder := lipgloss.RoundedBorder()
	contentWidth := e.width - sidebarWidth - 6
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentStyle := lipgloss.NewStyle().
		Border(contentBorder).
		BorderForeground(e.borderColor(focusContent)).
		Width(contentWidth).
		Padding(0, 1)

	sidebarView := sidebarStyle.Render(sidebar)
	contentView := contentStyle.Render(content)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, " ", contentView)

	// Status bar
	var status string
	if e.successMsg != "" {
		status = editorSuccessStyle.Render(e.successMsg)
	} else if e.err != nil {
		status = editorErrorStyle.Render(fmt.Sprintf("Error: %v", e.err))
	} else if len(e.validationErr) > 0 {
		var errMsgs []string
		for _, ve := range e.validationErr {
			errMsgs = append(errMsgs, ve.Error())
		}
		status = editorErrorStyle.Render("Validation errors: " + strings.Join(errMsgs, "; "))
	} else if e.saving {
		status = editorSubtitleStyle.Render("Saving...")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		mainView,
		"",
		status,
	)
}

func (e Editor) borderColor(focus editorFocus) lipgloss.Color {
	if e.focus == focus {
		return editorPurple
	}
	return editorGray
}

func (e Editor) renderSidebar() string {
	var lines []string

	for i, name := range sectionNames {
		cursor := "  "
		style := editorInactiveStyle

		if i == e.section {
			cursor = "> "
			if e.focus == focusSidebar {
				style = editorActiveStyle
			} else {
				style = editorAccentStyle
			}
		}

		lines = append(lines, style.Render(cursor+name))
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderContent() string {
	if e.workingProfile == nil {
		return editorSubtitleStyle.Render("No profile loaded")
	}

	switch e.section {
	case sectionName:
		return e.renderNameSection()
	case sectionAgents:
		return e.renderAgentsSection()
	case sectionHooks:
		return e.renderHooksSection()
	case sectionDisabled:
		return e.renderDisabledSection()
	case sectionOther:
		return e.renderOtherSection()
	case sectionSkills:
		return e.renderSkillsSection()
	case sectionReview:
		return e.renderReviewSection()
	}

	return ""
}

func (e Editor) renderNameSection() string {
	label := editorSubtitleStyle.Render("Profile Name:")
	input := e.nameInput.View()

	return lipgloss.JoinVertical(lipgloss.Left, label, input)
}

func (e Editor) renderAgentsSection() string {
	if len(e.workingProfile.Config.Agents) == 0 {
		return editorSubtitleStyle.Render("No agents configured in this profile.")
	}

	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Configured Agents:"))
	lines = append(lines, "")

	i := 0
	for name, agent := range e.workingProfile.Config.Agents {
		cursor := "  "
		style := editorInactiveStyle
		if e.focus == focusContent && i == e.contentCursor {
			cursor = "> "
			style = editorActiveStyle
		}

		info := name
		if agent != nil && agent.Model != "" {
			info += fmt.Sprintf(" (%s)", agent.Model)
		}
		lines = append(lines, style.Render(cursor+info))
		i++
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderHooksSection() string {
	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Hooks (space to toggle):"))
	lines = append(lines, "")

	for i, hook := range e.availableHooks {
		cursor := "  "
		style := editorInactiveStyle
		if e.focus == focusContent && i == e.contentCursor {
			cursor = "> "
			style = editorActiveStyle
		}

		checkbox := "[ ]"
		if e.hooksEnabled[hook] {
			checkbox = "[x]"
		}

		lines = append(lines, style.Render(fmt.Sprintf("%s%s %s", cursor, checkbox, hook)))
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderDisabledSection() string {
	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Disabled Items:"))
	lines = append(lines, "")

	items := []struct {
		label string
		list  []string
	}{
		{"MCPs", e.workingProfile.Config.DisabledMCPs},
		{"Agents", e.workingProfile.Config.DisabledAgents},
		{"Skills", e.workingProfile.Config.DisabledSkills},
		{"Commands", e.workingProfile.Config.DisabledCommands},
	}

	for i, item := range items {
		cursor := "  "
		style := editorInactiveStyle
		if e.focus == focusContent && i == e.contentCursor {
			cursor = "> "
			style = editorActiveStyle
		}

		value := "(none)"
		if len(item.list) > 0 {
			value = strings.Join(item.list, ", ")
		}

		lines = append(lines, style.Render(fmt.Sprintf("%s%s: %s", cursor, item.label, value)))
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderOtherSection() string {
	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Other Settings:"))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Auto Update: %s", formatBoolPtr(e.workingProfile.Config.AutoUpdate)))

	if e.workingProfile.Config.ClaudeCode != nil {
		cc := e.workingProfile.Config.ClaudeCode
		lines = append(lines, "")
		lines = append(lines, editorSubtitleStyle.Render("Claude Code:"))
		lines = append(lines, fmt.Sprintf("  MCP: %s", formatBoolPtr(cc.MCP)))
		lines = append(lines, fmt.Sprintf("  Commands: %s", formatBoolPtr(cc.Commands)))
		lines = append(lines, fmt.Sprintf("  Skills: %s", formatBoolPtr(cc.Skills)))
		lines = append(lines, fmt.Sprintf("  Agents: %s", formatBoolPtr(cc.Agents)))
		lines = append(lines, fmt.Sprintf("  Hooks: %s", formatBoolPtr(cc.Hooks)))
	}

	if e.workingProfile.Config.SisyphusAgent != nil {
		sa := e.workingProfile.Config.SisyphusAgent
		lines = append(lines, "")
		lines = append(lines, editorSubtitleStyle.Render("Sisyphus Agent:"))
		lines = append(lines, fmt.Sprintf("  Disabled: %s", formatBoolPtr(sa.Disabled)))
		lines = append(lines, fmt.Sprintf("  Default Builder: %s", formatBoolPtr(sa.DefaultBuilderEnabled)))
		lines = append(lines, fmt.Sprintf("  Planner: %s", formatBoolPtr(sa.PlannerEnabled)))
	}

	if e.workingProfile.Config.Experimental != nil {
		exp := e.workingProfile.Config.Experimental
		lines = append(lines, "")
		lines = append(lines, editorSubtitleStyle.Render("Experimental:"))
		lines = append(lines, fmt.Sprintf("  Aggressive Truncation: %s", formatBoolPtr(exp.AggressiveTruncation)))
		lines = append(lines, fmt.Sprintf("  Auto Resume: %s", formatBoolPtr(exp.AutoResume)))
		lines = append(lines, fmt.Sprintf("  Preemptive Compaction: %s", formatBoolPtr(exp.PreemptiveCompaction)))
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderSkillsSection() string {
	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Skills Configuration:"))
	lines = append(lines, "")

	if e.workingProfile.Config.Skills == nil {
		lines = append(lines, editorSubtitleStyle.Render("(no skills configured)"))
	} else {
		// Pretty print skills JSON
		var prettyJSON []byte
		prettyJSON, err := json.MarshalIndent(e.workingProfile.Config.Skills, "", "  ")
		if err != nil {
			lines = append(lines, editorErrorStyle.Render("Error formatting skills"))
		} else {
			lines = append(lines, string(prettyJSON))
		}
	}

	return strings.Join(lines, "\n")
}

func (e Editor) renderReviewSection() string {
	var lines []string
	lines = append(lines, editorSubtitleStyle.Render("Review & Save:"))
	lines = append(lines, "")

	// Show JSON preview
	jsonData, err := json.MarshalIndent(e.workingProfile.Config, "", "  ")
	if err != nil {
		lines = append(lines, editorErrorStyle.Render("Error generating preview"))
	} else {
		// Truncate if too long
		preview := string(jsonData)
		previewLines := strings.Split(preview, "\n")
		maxLines := 15
		if len(previewLines) > maxLines {
			previewLines = previewLines[:maxLines]
			previewLines = append(previewLines, "...")
		}
		lines = append(lines, strings.Join(previewLines, "\n"))
	}

	lines = append(lines, "")

	// Save button
	cursor := "  "
	style := editorInactiveStyle
	if e.focus == focusContent && e.contentCursor == 0 {
		cursor = "> "
		style = editorActiveStyle
	}
	lines = append(lines, style.Render(cursor+"[ Save Profile ]"))

	return strings.Join(lines, "\n")
}

// ShouldReturn indicates if the editor should return to previous view
func (e Editor) ShouldReturn() bool {
	return false
}

// ProfileName returns the name of the profile being edited
func (e Editor) ProfileName() string {
	return e.profileName
}

// HasChanges returns true if the profile has been modified
func (e Editor) HasChanges() bool {
	if e.originalProfile == nil || e.workingProfile == nil {
		return false
	}

	origJSON, err1 := json.Marshal(e.originalProfile.Config)
	workJSON, err2 := json.Marshal(e.workingProfile.Config)

	if err1 != nil || err2 != nil {
		return true
	}

	return string(origJSON) != string(workJSON) || e.originalProfile.Name != e.workingProfile.Name
}
