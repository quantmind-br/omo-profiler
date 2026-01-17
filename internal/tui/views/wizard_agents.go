package views

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/config"
)

// All 15 agents
var allAgents = []string{
	"build",
	"plan",
	"Sisyphus",
	"Sisyphus-Junior",
	"OpenCode-Builder",
	"Prometheus (Planner)",
	"Metis (Plan Consultant)",
	"Momus (Plan Reviewer)",
	"oracle",
	"librarian",
	"explore",
	"frontend-ui-ux-engineer",
	"document-writer",
	"multimodal-looker",
	"orchestrator-sisyphus",
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
	fieldPrompt
	fieldPromptAppend
	fieldDisable
	fieldDescription
	fieldMode
	fieldColor
	fieldPermEdit
	fieldPermBash
	fieldPermWebfetch
	fieldPermDoomLoop
	fieldPermExtDir
)

type agentConfig struct {
	enabled      bool
	expanded     bool
	model        textinput.Model
	variant      textinput.Model
	category     textinput.Model
	temperature  textinput.Model
	topP         textinput.Model
	skills       textinput.Model
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
}

func newAgentConfig() agentConfig {
	model := textinput.New()
	model.Placeholder = "model name"
	model.Width = 30

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

	return agentConfig{
		model:        model,
		variant:      variant,
		category:     category,
		temperature:  temperature,
		topP:         topP,
		skills:       skills,
		prompt:       prompt,
		promptAppend: promptAppend,
		description:  description,
		color:        color,
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
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
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
				ac.model.SetValue(agentCfg.Model)
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
			if agentCfg.Prompt != "" {
				ac.prompt.SetValue(agentCfg.Prompt)
			}
			if agentCfg.PromptAppend != "" {
				ac.promptAppend.SetValue(agentCfg.PromptAppend)
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

		if v := ac.model.Value(); v != "" {
			agentCfg.Model = v
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
		if v := ac.prompt.Value(); v != "" {
			agentCfg.Prompt = v
		}
		if v := ac.promptAppend.Value(); v != "" {
			agentCfg.PromptAppend = v
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

func (w WizardAgents) Update(msg tea.Msg) (WizardAgents, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	currentAgent := allAgents[w.cursor]
	ac := w.agents[currentAgent]

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		// When in form editing mode
		if w.inForm && ac.expanded {
			switch msg.String() {
			case "esc":
				w.inForm = false
				return w, nil
			case "tab":
				w.focusedField++
				if w.focusedField > fieldPermExtDir {
					w.focusedField = fieldModel
				}
				return w, nil
			case "shift+tab":
				if w.focusedField == fieldModel {
					w.focusedField = fieldPermExtDir
				} else {
					w.focusedField--
				}
				return w, nil
			case "left", "right":
				// Cycle through options for dropdown fields
				switch w.focusedField {
				case fieldDisable:
					ac.disable = !ac.disable
				case fieldMode:
					if msg.String() == "right" {
						ac.modeIdx = (ac.modeIdx + 1) % len(agentModes)
					} else {
						ac.modeIdx = (ac.modeIdx - 1 + len(agentModes)) % len(agentModes)
					}
				case fieldPermEdit:
					if msg.String() == "right" {
						ac.permEditIdx = (ac.permEditIdx + 1) % len(permissionValues)
					} else {
						ac.permEditIdx = (ac.permEditIdx - 1 + len(permissionValues)) % len(permissionValues)
					}
				case fieldPermBash:
					if msg.String() == "right" {
						ac.permBashIdx = (ac.permBashIdx + 1) % len(permissionValues)
					} else {
						ac.permBashIdx = (ac.permBashIdx - 1 + len(permissionValues)) % len(permissionValues)
					}
				case fieldPermWebfetch:
					if msg.String() == "right" {
						ac.permWebfetchIdx = (ac.permWebfetchIdx + 1) % len(permissionValues)
					} else {
						ac.permWebfetchIdx = (ac.permWebfetchIdx - 1 + len(permissionValues)) % len(permissionValues)
					}
				case fieldPermDoomLoop:
					if msg.String() == "right" {
						ac.permDoomLoopIdx = (ac.permDoomLoopIdx + 1) % len(permissionValues)
					} else {
						ac.permDoomLoopIdx = (ac.permDoomLoopIdx - 1 + len(permissionValues)) % len(permissionValues)
					}
				case fieldPermExtDir:
					if msg.String() == "right" {
						ac.permExtDirIdx = (ac.permExtDirIdx + 1) % len(permissionValues)
					} else {
						ac.permExtDirIdx = (ac.permExtDirIdx - 1 + len(permissionValues)) % len(permissionValues)
					}
				}
				return w, nil
			}

			// Update focused text input
			switch w.focusedField {
			case fieldModel:
				ac.model.Focus()
				ac.model, cmd = ac.model.Update(msg)
				cmds = append(cmds, cmd)
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
				}
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

	renderField := func(label string, field agentFormField, value string) string {
		style := fieldStyle
		if w.inForm && w.focusedField == field {
			style = focusStyle
		}
		return indent + style.Render(fmt.Sprintf("%-12s: ", label)) + value
	}

	renderDropdown := func(label string, field agentFormField, options []string, idx int) string {
		style := fieldStyle
		if w.inForm && w.focusedField == field {
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
		if w.inForm && w.focusedField == field {
			style = focusStyle
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent + style.Render(fmt.Sprintf("%-12s: ", label)) + checkbox + " [←/→]"
	}

	lines = append(lines, "")
	lines = append(lines, renderField("model", fieldModel, ac.model.View()))
	lines = append(lines, renderField("variant", fieldVariant, ac.variant.View()))
	lines = append(lines, renderField("category", fieldCategory, ac.category.View()))
	lines = append(lines, renderField("temperature", fieldTemperature, ac.temperature.View()))
	lines = append(lines, renderField("top_p", fieldTopP, ac.topP.View()))
	lines = append(lines, renderField("skills", fieldSkills, ac.skills.View()))
	lines = append(lines, renderField("prompt", fieldPrompt, ac.prompt.View()))
	lines = append(lines, renderField("prompt_append", fieldPromptAppend, ac.promptAppend.View()))
	lines = append(lines, renderBool("disable", fieldDisable, ac.disable))
	lines = append(lines, renderField("description", fieldDescription, ac.description.View()))
	lines = append(lines, renderDropdown("mode", fieldMode, agentModes, ac.modeIdx))
	lines = append(lines, renderField("color", fieldColor, ac.color.View()))
	lines = append(lines, "")
	lines = append(lines, indent+fieldStyle.Render("── Permissions ──"))
	lines = append(lines, renderDropdown("edit", fieldPermEdit, permissionValues, ac.permEditIdx))
	lines = append(lines, renderDropdown("bash", fieldPermBash, permissionValues, ac.permBashIdx))
	lines = append(lines, renderDropdown("webfetch", fieldPermWebfetch, permissionValues, ac.permWebfetchIdx))
	lines = append(lines, renderDropdown("doom_loop", fieldPermDoomLoop, permissionValues, ac.permDoomLoopIdx))
	lines = append(lines, renderDropdown("external_dir", fieldPermExtDir, permissionValues, ac.permExtDirIdx))
	lines = append(lines, "")

	return lines
}

func (w WizardAgents) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	title := titleStyle.Render("Configure Agents")
	desc := helpStyle.Render("Space to enable/disable • Enter to expand • Tab to next step")

	if w.inForm {
		desc = helpStyle.Render("Tab/Shift+Tab to navigate • ←/→ for options • Esc to close form")
	}

	content := w.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		"",
		content,
	)
}
