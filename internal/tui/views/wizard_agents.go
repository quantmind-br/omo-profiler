package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/models"
)

func parseMapStringBool(s string) map[string]bool {
	if s == "" {
		return nil
	}
	result := make(map[string]bool)
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		result[key] = val == "true"
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func serializeMapStringBool(m map[string]bool) string {
	if len(m) == 0 {
		return ""
	}
	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s:%t", k, v))
	}
	return strings.Join(pairs, ", ")
}

var allAgents = []string{
	"build",
	"plan",
	"sisyphus",
	"hephaestus",
	"sisyphus-junior",
	"OpenCode-Builder",
	"prometheus",
	"metis",
	"momus",
	"atlas",
	"oracle",
	"librarian",
	"explore",
	"multimodal-looker",
}

// Mode options for agents
var agentModes = []string{"", "subagent", "primary", "all"}

// Permission values
var permissionValues = []string{"", "ask", "allow", "deny"}

type agentFormField int

const (
	fieldModel agentFormField = iota
	fieldVariant
	fieldCategory
	fieldTemperature
	fieldTopP
	fieldSkills
	fieldTools
	fieldPrompt
	fieldPromptAppend
	fieldDisable
	fieldDescription
	fieldMode
	fieldColor
	fieldPermEdit
)

type agentConfig struct {
	enabled      bool
	expanded     bool
	modelValue   string
	modelDisplay string
	variant      textinput.Model
	category     textinput.Model
	temperature  textinput.Model
	topP         textinput.Model
	skills       textinput.Model
	tools        textinput.Model
	prompt       textarea.Model
	promptAppend textarea.Model
	disable      bool
	description  textinput.Model
	modeIdx      int
	color        textinput.Model
	// Permissions
	permEditIdx     int
	permBashIdx     int
	permWebfetchIdx int
	permDoomLoopIdx int
	permExtDirIdx   int
	// Model selector state
	selectingModel       bool
	modelSelector        ModelSelector
	savingCustomModel    bool
	customModelToSave    string
	savePromptAnswer     string
	saveDisplayNameInput textinput.Model
	saveProviderInput    textinput.Model
	saveFocusedField     int
	saveError            string
}

func newAgentConfig() agentConfig {
	variant := textinput.New()
	variant.Placeholder = "variant"
	variant.Width = 30

	category := textinput.New()
	category.Placeholder = "category"
	category.Width = 30

	temperature := textinput.New()
	temperature.Placeholder = "0.0-2.0"
	temperature.Width = 10

	topP := textinput.New()
	topP.Placeholder = "0.0-1.0"
	topP.Width = 10

	skills := textinput.New()
	skills.Placeholder = "skill1, skill2"
	skills.Width = 40

	tools := textinput.New()
	tools.Placeholder = "tool1:true, tool2:false"
	tools.Width = 40

	prompt := textarea.New()
	prompt.Placeholder = "Custom prompt..."
	prompt.SetWidth(50)
	prompt.SetHeight(3)

	promptAppend := textarea.New()
	promptAppend.Placeholder = "Append to prompt..."
	promptAppend.SetWidth(50)
	promptAppend.SetHeight(3)

	description := textinput.New()
	description.Placeholder = "description"
	description.Width = 40

	color := textinput.New()
	color.Placeholder = "#RRGGBB"
	color.Width = 10

	saveDisplayNameInput := textinput.New()
	saveDisplayNameInput.Placeholder = "Display name"
	saveDisplayNameInput.Width = 30

	saveProviderInput := textinput.New()
	saveProviderInput.Placeholder = "Provider (optional)"
	saveProviderInput.Width = 30

	return agentConfig{
		variant:              variant,
		category:             category,
		temperature:          temperature,
		topP:                 topP,
		skills:               skills,
		tools:                tools,
		prompt:               prompt,
		promptAppend:         promptAppend,
		description:          description,
		color:                color,
		saveDisplayNameInput: saveDisplayNameInput,
		saveProviderInput:    saveProviderInput,
	}
}

type wizardAgentsKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	Expand   key.Binding
	Next     key.Binding
	Back     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Left     key.Binding
	Right    key.Binding
}

func newWizardAgentsKeyMap() wizardAgentsKeyMap {
	return wizardAgentsKeyMap{
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
		Expand: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "expand/collapse"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next step"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "prev field"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand"),
		),
	}
}

// WizardAgents is step 2: Agent configuration
type WizardAgents struct {
	agents       map[string]*agentConfig
	cursor       int
	focusedField agentFormField
	inForm       bool // true when editing expanded agent form
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
	keys         wizardAgentsKeyMap
}

func NewWizardAgents() WizardAgents {
	agents := make(map[string]*agentConfig)
	for _, name := range allAgents {
		cfg := newAgentConfig()
		agents[name] = &cfg
	}

	return WizardAgents{
		agents: agents,
		keys:   newWizardAgentsKeyMap(),
	}
}

func (w WizardAgents) Init() tea.Cmd {
	return nil
}

func (w *WizardAgents) SetSize(width, height int) {
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

func (w *WizardAgents) SetConfig(cfg *config.Config) {
	if cfg.Agents == nil {
		return
	}
	for name, agentCfg := range cfg.Agents {
		if ac, ok := w.agents[name]; ok {
			ac.enabled = true
			if agentCfg.Model != "" {
				ac.modelValue = agentCfg.Model
				ac.modelDisplay = agentCfg.Model
			}
			if agentCfg.Variant != "" {
				ac.variant.SetValue(agentCfg.Variant)
			}
			if agentCfg.Category != "" {
				ac.category.SetValue(agentCfg.Category)
			}
			if agentCfg.Temperature != nil {
				ac.temperature.SetValue(fmt.Sprintf("%.1f", *agentCfg.Temperature))
			}
			if agentCfg.TopP != nil {
				ac.topP.SetValue(fmt.Sprintf("%.1f", *agentCfg.TopP))
			}
			if len(agentCfg.Skills) > 0 {
				ac.skills.SetValue(strings.Join(agentCfg.Skills, ", "))
			}
			if len(agentCfg.Tools) > 0 {
				ac.tools.SetValue(serializeMapStringBool(agentCfg.Tools))
			}
			if agentCfg.Disable != nil && *agentCfg.Disable {
				ac.disable = true
			}
			if agentCfg.Description != "" {
				ac.description.SetValue(agentCfg.Description)
			}
			if agentCfg.Mode != "" {
				for i, m := range agentModes {
					if m == agentCfg.Mode {
						ac.modeIdx = i
						break
					}
				}
			}
			if agentCfg.Color != "" {
				ac.color.SetValue(agentCfg.Color)
			}
			// Permissions
			if agentCfg.Permission != nil {
				for i, v := range permissionValues {
					if v == agentCfg.Permission.Edit {
						ac.permEditIdx = i
					}
					if bashStr, ok := agentCfg.Permission.Bash.(string); ok && v == bashStr {
						ac.permBashIdx = i
					}
					if v == agentCfg.Permission.Webfetch {
						ac.permWebfetchIdx = i
					}
					if v == agentCfg.Permission.DoomLoop {
						ac.permDoomLoopIdx = i
					}
					if v == agentCfg.Permission.ExternalDirectory {
						ac.permExtDirIdx = i
					}
				}
			}
		}
	}
}

func (w *WizardAgents) Apply(cfg *config.Config) {
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]*config.AgentConfig)
	}

	for name, ac := range w.agents {
		if !ac.enabled {
			delete(cfg.Agents, name)
			continue
		}

		agentCfg := &config.AgentConfig{}

		if ac.modelValue != "" {
			agentCfg.Model = ac.modelValue
		}
		if v := ac.variant.Value(); v != "" {
			agentCfg.Variant = v
		}
		if v := ac.category.Value(); v != "" {
			agentCfg.Category = v
		}
		if v := ac.temperature.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				agentCfg.Temperature = &f
			}
		}
		if v := ac.topP.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				agentCfg.TopP = &f
			}
		}
		if v := ac.skills.Value(); v != "" {
			parts := strings.Split(v, ",")
			var skills []string
			for _, p := range parts {
				if s := strings.TrimSpace(p); s != "" {
					skills = append(skills, s)
				}
			}
			if len(skills) > 0 {
				agentCfg.Skills = skills
			}
		}
		if v := ac.tools.Value(); v != "" {
			agentCfg.Tools = parseMapStringBool(v)
		}
		if ac.disable {
			b := true
			agentCfg.Disable = &b
		}
		if v := ac.description.Value(); v != "" {
			agentCfg.Description = v
		}
		if ac.modeIdx > 0 {
			agentCfg.Mode = agentModes[ac.modeIdx]
		}
		if v := ac.color.Value(); v != "" {
			agentCfg.Color = v
		}

		// Permissions
		if ac.permEditIdx > 0 || ac.permBashIdx > 0 || ac.permWebfetchIdx > 0 ||
			ac.permDoomLoopIdx > 0 || ac.permExtDirIdx > 0 {
			agentCfg.Permission = &config.PermissionConfig{}
			if ac.permEditIdx > 0 {
				agentCfg.Permission.Edit = permissionValues[ac.permEditIdx]
			}
			if ac.permBashIdx > 0 {
				agentCfg.Permission.Bash = permissionValues[ac.permBashIdx]
			}
			if ac.permWebfetchIdx > 0 {
				agentCfg.Permission.Webfetch = permissionValues[ac.permWebfetchIdx]
			}
			if ac.permDoomLoopIdx > 0 {
				agentCfg.Permission.DoomLoop = permissionValues[ac.permDoomLoopIdx]
			}
			if ac.permExtDirIdx > 0 {
				agentCfg.Permission.ExternalDirectory = permissionValues[ac.permExtDirIdx]
			}
		}

		cfg.Agents[name] = agentCfg
	}

	// Remove empty agents map
	if len(cfg.Agents) == 0 {
		cfg.Agents = nil
	}
}

func (w *WizardAgents) updateFieldFocus(ac *agentConfig) {
	ac.variant.Blur()
	ac.category.Blur()
	ac.temperature.Blur()
	ac.topP.Blur()
	ac.skills.Blur()
	ac.tools.Blur()
	ac.prompt.Blur()
	ac.promptAppend.Blur()
	ac.description.Blur()
	ac.color.Blur()

	switch w.focusedField {
	case fieldModel:
		// Model field uses selector, no focus needed
	case fieldVariant:
		ac.variant.Focus()
	case fieldCategory:
		ac.category.Focus()
	case fieldTemperature:
		ac.temperature.Focus()
	case fieldTopP:
		ac.topP.Focus()
	case fieldSkills:
		ac.skills.Focus()
	case fieldTools:
		ac.tools.Focus()
	case fieldPrompt:
		ac.prompt.Focus()
	case fieldPromptAppend:
		ac.promptAppend.Focus()
	case fieldDescription:
		ac.description.Focus()
	case fieldColor:
		ac.color.Focus()
	}
}

// getLineForField calculates the viewport line for the focused field
func (w WizardAgents) getLineForField(field agentFormField) int {
	// Base lines: count lines for agents before the current one
	baseLine := 0
	for i := 0; i < w.cursor; i++ {
		baseLine++ // agent header line
		ac := w.agents[allAgents[i]]
		if ac.expanded && ac.enabled {
			baseLine += 30 // expanded form ~30 lines
		}
	}
	baseLine++ // current agent header
	baseLine++ // empty line before form

	// Field offsets within the form
	fieldOffsets := map[agentFormField]int{
		fieldModel:        0,
		fieldVariant:      1,
		fieldCategory:     2,
		fieldTemperature:  3,
		fieldTopP:         4,
		fieldSkills:       5,
		fieldTools:        6,
		fieldPrompt:       7,
		fieldPromptAppend: 10,
		fieldDisable:      13,
		fieldDescription:  14,
		fieldMode:         15,
		fieldColor:        16,
		fieldPermEdit:     18,
	}

	return baseLine + fieldOffsets[field]
}

// ensureFieldVisible scrolls the viewport to keep the focused field visible
func (w *WizardAgents) ensureFieldVisible() {
	if !w.ready {
		return
	}
	line := w.getLineForField(w.focusedField)
	if line < w.viewport.YOffset {
		w.viewport.SetYOffset(line)
	} else if line >= w.viewport.YOffset+w.viewport.Height {
		w.viewport.SetYOffset(line - w.viewport.Height + 1)
	}
}

func (w WizardAgents) Update(msg tea.Msg) (WizardAgents, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	currentAgent := allAgents[w.cursor]
	ac := w.agents[currentAgent]

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)

		currentAgent := allAgents[w.cursor]
		if ac, ok := w.agents[currentAgent]; ok {
			if ac.selectingModel {
				ac.modelSelector.SetSize(msg.Width, msg.Height)
			}
		}
		return w, nil

	case ModelSelectedMsg:
		ac.modelValue = msg.ModelID
		ac.modelDisplay = msg.DisplayName
		ac.selectingModel = false
		return w, nil

	case ModelSelectorCancelMsg:
		ac.selectingModel = false
		return w, nil

	case PromptSaveCustomMsg:
		ac.savingCustomModel = true
		ac.customModelToSave = msg.ModelID
		ac.savePromptAnswer = ""
		ac.saveDisplayNameInput.SetValue("")
		ac.saveProviderInput.SetValue("")
		ac.saveError = ""
		return w, nil

	case tea.KeyMsg:
		if ac.selectingModel {
			ac.modelSelector, cmd = ac.modelSelector.Update(msg)
			return w, cmd
		}

		if ac.savingCustomModel {
			return w.handleSaveCustomModel(ac, msg)
		}

		// When in form editing mode
		if w.inForm && ac.expanded {
			switch msg.String() {
			case "esc":
				w.inForm = false
				ac.expanded = false
				return w, nil
			case "down", "j":
				w.focusedField++
				if w.focusedField > fieldPermEdit {
					w.focusedField = fieldModel
				}
				w.updateFieldFocus(ac)
				w.ensureFieldVisible()
				return w, nil
			case "up", "k":
				if w.focusedField > fieldModel {
					w.focusedField--
				} else {
					w.focusedField = fieldPermEdit
				}
				w.updateFieldFocus(ac)
				w.ensureFieldVisible()
				return w, nil
			case "tab":
				w.focusedField++
				if w.focusedField > fieldPermEdit {
					w.focusedField = fieldModel
				}
				w.updateFieldFocus(ac)
				w.ensureFieldVisible()
				return w, nil
			case "shift+tab":
				if w.focusedField > fieldModel {
					w.focusedField--
				} else {
					w.focusedField = fieldPermEdit
				}
				w.updateFieldFocus(ac)
				w.ensureFieldVisible()
				return w, nil
			case "enter":
				// Cycle through options for dropdown fields
				switch w.focusedField {
				case fieldModel:
					ac.selectingModel = true
					ac.modelSelector = NewModelSelector()
					ac.modelSelector.SetSize(w.width, w.height)
					return w, nil
				case fieldDisable:
					ac.disable = !ac.disable
				case fieldMode:
					ac.modeIdx = (ac.modeIdx + 1) % len(agentModes)
				case fieldPermEdit:
					ac.permEditIdx = (ac.permEditIdx + 1) % len(permissionValues)
				}
				return w, nil
			}

			// Update focused text input
			switch w.focusedField {
			case fieldModel:
				// Model uses selector, handled separately via enter key
			case fieldVariant:
				ac.variant.Focus()
				ac.variant, cmd = ac.variant.Update(msg)
				cmds = append(cmds, cmd)
			case fieldCategory:
				ac.category.Focus()
				ac.category, cmd = ac.category.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTemperature:
				ac.temperature.Focus()
				ac.temperature, cmd = ac.temperature.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTopP:
				ac.topP.Focus()
				ac.topP, cmd = ac.topP.Update(msg)
				cmds = append(cmds, cmd)
			case fieldSkills:
				ac.skills.Focus()
				ac.skills, cmd = ac.skills.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTools:
				ac.tools.Focus()
				ac.tools, cmd = ac.tools.Update(msg)
				cmds = append(cmds, cmd)
			case fieldPrompt:
				ac.prompt.Focus()
				ac.prompt, cmd = ac.prompt.Update(msg)
				cmds = append(cmds, cmd)
			case fieldPromptAppend:
				ac.promptAppend.Focus()
				ac.promptAppend, cmd = ac.promptAppend.Update(msg)
				cmds = append(cmds, cmd)
			case fieldDescription:
				ac.description.Focus()
				ac.description, cmd = ac.description.Update(msg)
				cmds = append(cmds, cmd)
			case fieldColor:
				ac.color.Focus()
				ac.color, cmd = ac.color.Update(msg)
				cmds = append(cmds, cmd)
			}

			return w, tea.Batch(cmds...)
		}

		// Navigation mode
		switch {
		case key.Matches(msg, w.keys.Up):
			if w.cursor > 0 {
				w.cursor--
			}
		case key.Matches(msg, w.keys.Down):
			if w.cursor < len(allAgents)-1 {
				w.cursor++
			}
		case key.Matches(msg, w.keys.Toggle):
			ac.enabled = !ac.enabled
		case key.Matches(msg, w.keys.Expand):
			if ac.enabled {
				ac.expanded = !ac.expanded
				if ac.expanded {
					w.inForm = true
					w.focusedField = fieldModel
					w.updateFieldFocus(ac)
				} else {
					w.inForm = false
				}
			}
		case key.Matches(msg, w.keys.Right):
			// Expand only when not in form mode and agent is enabled
			if !w.inForm && ac.enabled && !ac.expanded {
				ac.expanded = true
				w.inForm = true
				w.focusedField = fieldModel
				w.updateFieldFocus(ac)
			}
		case key.Matches(msg, w.keys.Left):
			// Collapse only when not in form mode
			if !w.inForm && ac.expanded {
				ac.expanded = false
				w.inForm = false
			}
		case key.Matches(msg, w.keys.Next):
			if !w.inForm {
				return w, func() tea.Msg { return WizardNextMsg{} }
			}
		case key.Matches(msg, w.keys.Back):
			if !w.inForm {
				return w, func() tea.Msg { return WizardBackMsg{} }
			}
		}
	}

	// Update viewport
	w.viewport.SetContent(w.renderContent())
	w.viewport, cmd = w.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return w, tea.Batch(cmds...)
}

func (w WizardAgents) renderContent() string {
	var lines []string

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))

	for i, name := range allAgents {
		ac := w.agents[name]

		cursor := "  "
		if i == w.cursor {
			cursor = selectedStyle.Render("> ")
		}

		checkbox := "[ ]"
		if ac.enabled {
			checkbox = enabledStyle.Render("[✓]")
		}

		expandIcon := ""
		if ac.enabled {
			if ac.expanded {
				expandIcon = " ▼"
			} else {
				expandIcon = " ▶"
			}
		}

		nameStyle := dimStyle
		if i == w.cursor {
			nameStyle = labelStyle
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, nameStyle.Render(name), expandIcon)
		lines = append(lines, line)

		// Show expanded form
		if ac.expanded && ac.enabled {
			lines = append(lines, w.renderAgentForm(name, ac)...)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (w WizardAgents) renderAgentForm(name string, ac *agentConfig) []string {
	var lines []string

	indent := "      "
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	focusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	// Only show focus styling if this is the active agent being edited
	isActiveAgent := name == allAgents[w.cursor]

	renderField := func(label string, field agentFormField, value string) string {
		style := fieldStyle
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = focusStyle
		}
		return indent + style.Render(fmt.Sprintf("%-12s: ", label)) + value
	}

	renderDropdown := func(label string, field agentFormField, options []string, idx int) string {
		style := fieldStyle
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = focusStyle
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent + style.Render(fmt.Sprintf("%-12s: ", label)) + val + " [←/→]"
	}

	renderBool := func(label string, field agentFormField, val bool) string {
		style := fieldStyle
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = focusStyle
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent + style.Render(fmt.Sprintf("%-12s: ", label)) + checkbox + " [←/→]"
	}

	lines = append(lines, "")
	modelDisplayValue := ac.modelDisplay
	if modelDisplayValue == "" {
		modelDisplayValue = "[Select model...]"
	}
	lines = append(lines, renderField("model", fieldModel, modelDisplayValue))
	lines = append(lines, renderField("variant", fieldVariant, ac.variant.View()))
	lines = append(lines, renderField("category", fieldCategory, ac.category.View()))
	lines = append(lines, renderField("temperature", fieldTemperature, ac.temperature.View()))
	lines = append(lines, renderField("top_p", fieldTopP, ac.topP.View()))
	lines = append(lines, renderField("skills", fieldSkills, ac.skills.View()))
	lines = append(lines, renderField("tools", fieldTools, ac.tools.View()))
	lines = append(lines, renderField("prompt", fieldPrompt, ac.prompt.View()))
	lines = append(lines, renderField("prompt_append", fieldPromptAppend, ac.promptAppend.View()))
	lines = append(lines, renderBool("disable", fieldDisable, ac.disable))
	lines = append(lines, renderField("description", fieldDescription, ac.description.View()))
	lines = append(lines, renderDropdown("mode", fieldMode, agentModes, ac.modeIdx))
	lines = append(lines, renderField("color", fieldColor, ac.color.View()))
	lines = append(lines, "")
	lines = append(lines, indent+fieldStyle.Render("── Permissions ──"))
	lines = append(lines, renderDropdown("edit", fieldPermEdit, permissionValues, ac.permEditIdx))
	lines = append(lines, "")

	return lines
}

func (w WizardAgents) handleSaveCustomModel(ac *agentConfig, msg tea.KeyMsg) (WizardAgents, tea.Cmd) {
	if ac.savePromptAnswer == "" {
		switch msg.String() {
		case "y", "Y":
			ac.savePromptAnswer = "y"
			ac.saveFocusedField = 0
			ac.saveDisplayNameInput.Focus()
			return w, textinput.Blink
		case "n", "N":
			ac.savingCustomModel = false
			ac.savePromptAnswer = ""
			return w, nil
		case "esc":
			ac.savingCustomModel = false
			ac.savePromptAnswer = ""
			return w, nil
		}
		return w, nil
	}

	switch msg.String() {
	case "enter":
		displayName := strings.TrimSpace(ac.saveDisplayNameInput.Value())
		if displayName == "" {
			ac.saveError = "Display name is required"
			return w, nil
		}

		registry, err := models.Load()
		if err != nil {
			ac.saveError = err.Error()
			return w, nil
		}

		newModel := models.RegisteredModel{
			DisplayName: displayName,
			ModelID:     ac.customModelToSave,
			Provider:    strings.TrimSpace(ac.saveProviderInput.Value()),
		}

		if err := registry.Add(newModel); err != nil {
			ac.saveError = err.Error()
			return w, nil
		}

		ac.modelDisplay = displayName
		ac.savingCustomModel = false
		ac.savePromptAnswer = ""
		ac.saveError = ""
		return w, nil

	case "esc":
		ac.savingCustomModel = false
		ac.savePromptAnswer = ""
		return w, nil

	case "tab":
		ac.saveFocusedField = (ac.saveFocusedField + 1) % 2
		if ac.saveFocusedField == 0 {
			ac.saveDisplayNameInput.Focus()
			ac.saveProviderInput.Blur()
		} else {
			ac.saveProviderInput.Focus()
			ac.saveDisplayNameInput.Blur()
		}
		return w, nil

	case "shift+tab":
		ac.saveFocusedField = (ac.saveFocusedField + 1) % 2
		if ac.saveFocusedField == 0 {
			ac.saveDisplayNameInput.Focus()
			ac.saveProviderInput.Blur()
		} else {
			ac.saveProviderInput.Focus()
			ac.saveDisplayNameInput.Blur()
		}
		return w, nil
	}

	var cmd tea.Cmd
	if ac.saveFocusedField == 0 {
		ac.saveDisplayNameInput, cmd = ac.saveDisplayNameInput.Update(msg)
	} else {
		ac.saveProviderInput, cmd = ac.saveProviderInput.Update(msg)
	}
	ac.saveError = ""
	return w, cmd
}

func (w WizardAgents) View() string {
	currentAgent := allAgents[w.cursor]
	ac := w.agents[currentAgent]

	if ac.selectingModel {
		return ac.modelSelector.View()
	}

	if ac.savingCustomModel {
		return w.renderSaveCustomPrompt(ac)
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	title := titleStyle.Render("Configure Agents")
	desc := helpStyle.Render("Space to enable/disable • Enter to expand • Tab to next step")

	if w.inForm {
		desc = helpStyle.Render("↑/↓/Tab: navigate • Enter: cycle options • Esc: close form")
	}

	content := w.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		"",
		content,
	)
}

func (w WizardAgents) renderSaveCustomPrompt(ac *agentConfig) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))

	var lines []string
	lines = append(lines, titleStyle.Render("Custom Model"))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render(fmt.Sprintf("Model ID: %s", ac.customModelToSave)))
	lines = append(lines, "")

	if ac.savePromptAnswer == "" {
		lines = append(lines, labelStyle.Render("Save this model for future use? (y/n)"))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("[y] yes  [n] no  [Esc] cancel"))
	} else {
		lines = append(lines, labelStyle.Render("Display name:"))
		lines = append(lines, ac.saveDisplayNameInput.View())
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Provider (optional):"))
		lines = append(lines, ac.saveProviderInput.View())
		lines = append(lines, "")
		if ac.saveError != "" {
			lines = append(lines, errStyle.Render("Error: "+ac.saveError))
			lines = append(lines, "")
		}
		lines = append(lines, helpStyle.Render("[Enter] save  [Tab] next field  [Esc] cancel"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
