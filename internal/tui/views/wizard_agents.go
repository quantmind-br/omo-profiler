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
	"github.com/diogenes/omo-profiler/internal/tui/layout"
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
	"oracle",
	"librarian",
	"explore",
	"multimodal-looker",
	"atlas",
}

// Mode options for agents
var agentModes = []string{"", "subagent", "primary", "all"}

// Permission values
var permissionValues = []string{"", "ask", "allow", "deny"}

var (
	wizAgentPurple = lipgloss.Color("#7D56F4")
	wizAgentGray   = lipgloss.Color("#6C7086")
	wizAgentText   = lipgloss.Color("#CDD6F4")
	wizAgentRed    = lipgloss.Color("#F38BA8")
	wizAgentGreen  = lipgloss.Color("#A6E3A1")
	wizAgentPink   = lipgloss.Color("#FF6AC1")
)

var (
	wizAgentLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizAgentText)
	wizAgentDimStyle      = lipgloss.NewStyle().Foreground(wizAgentGray)
	wizAgentSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizAgentPurple)
	wizAgentEnabledStyle  = lipgloss.NewStyle().Foreground(wizAgentGreen)
	wizAgentCursorStyle   = lipgloss.NewStyle().Bold(true).Foreground(wizAgentPink)
	wizAgentTextStyle     = lipgloss.NewStyle().Foreground(wizAgentText)
	wizAgentErrorStyle    = lipgloss.NewStyle().Foreground(wizAgentRed)
)

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
	fieldMaxTokens
	fieldThinkingType
	fieldThinkingBudget
	fieldReasoningEffort
	fieldTextVerbosity
	fieldProviderOptions
	fieldPermEdit
	fieldPermBash
	fieldPermWebfetch
	fieldPermTask
	fieldPermDoomLoop
	fieldPermExtDir
)

type agentConfig struct {
	enabled                bool
	expanded               bool
	modelValue             string
	modelDisplay           string
	variant                textinput.Model
	category               textinput.Model
	temperature            textinput.Model
	topP                   textinput.Model
	skills                 textinput.Model
	tools                  textinput.Model
	prompt                 textarea.Model
	promptAppend           textarea.Model
	disable                bool
	description            textinput.Model
	modeIdx                int
	color                  textinput.Model
	maxTokens              textinput.Model
	thinkingTypeIdx        int
	thinkingBudget         textinput.Model
	reasoningEffortIdx     int
	textVerbosityIdx       int
	providerOptionsDisplay string
	// Permissions
	permEditIdx     int
	permBashIdx     int
	permWebfetchIdx int
	permTaskIdx     int
	permDoomLoopIdx int
	permExtDirIdx   int
	originalBash    interface{} // Preserve bash object through edit cycle
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

	maxTokens := textinput.New()
	maxTokens.Placeholder = "e.g. 8192"
	maxTokens.CharLimit = 10

	thinkingBudget := textinput.New()
	thinkingBudget.Placeholder = "e.g. 10000"
	thinkingBudget.CharLimit = 10

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
		maxTokens:            maxTokens,
		thinkingBudget:       thinkingBudget,
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

	// Guard against uninitialized struct (e.g. before navigation)
	if w.agents == nil {
		return
	}

	for _, ac := range w.agents {
		ac.variant.Width = layout.MediumFieldWidth(width)
		ac.category.Width = layout.MediumFieldWidth(width)
		ac.skills.Width = layout.WideFieldWidth(width, 10)
		ac.tools.Width = layout.WideFieldWidth(width, 10)
		ac.description.Width = layout.WideFieldWidth(width, 10)
		ac.prompt.SetWidth(layout.WideFieldWidth(width, 10))
		ac.promptAppend.SetWidth(layout.WideFieldWidth(width, 10))
		ac.saveDisplayNameInput.Width = layout.MediumFieldWidth(width)
		ac.saveProviderInput.Width = layout.MediumFieldWidth(width)
		ac.modelSelector.SetSize(width, height)
	}
	w.viewport.SetContent(w.renderContent())
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
			if agentCfg.Prompt != "" {
				ac.prompt.SetValue(agentCfg.Prompt)
			}
			if agentCfg.PromptAppend != "" {
				ac.promptAppend.SetValue(agentCfg.PromptAppend)
			}
			// Permissions
			if agentCfg.Permission != nil {
				ac.originalBash = agentCfg.Permission.Bash
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
					if v == agentCfg.Permission.Task {
						ac.permTaskIdx = i
					}
					if v == agentCfg.Permission.DoomLoop {
						ac.permDoomLoopIdx = i
					}
					if v == agentCfg.Permission.ExternalDirectory {
						ac.permExtDirIdx = i
					}
				}
			}
			if agentCfg.MaxTokens != nil {
				ac.maxTokens.SetValue(fmt.Sprintf("%.0f", *agentCfg.MaxTokens))
			}
			if agentCfg.Thinking != nil {
				for i, t := range thinkingTypes {
					if t == agentCfg.Thinking.Type {
						ac.thinkingTypeIdx = i
						break
					}
				}
				if agentCfg.Thinking.BudgetTokens != nil {
					ac.thinkingBudget.SetValue(fmt.Sprintf("%.0f", *agentCfg.Thinking.BudgetTokens))
				}
			}
			for i, e := range effortLevels {
				if e == agentCfg.ReasoningEffort {
					ac.reasoningEffortIdx = i
					break
				}
			}
			for i, v := range verbosityLevels {
				if v == agentCfg.TextVerbosity {
					ac.textVerbosityIdx = i
					break
				}
			}
			if len(agentCfg.ProviderOptions) > 0 {
				ac.providerOptionsDisplay = fmt.Sprintf("[%d options set]", len(agentCfg.ProviderOptions))
			} else {
				ac.providerOptionsDisplay = "(none)"
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

		var agentCfg *config.AgentConfig
		if existing, ok := cfg.Agents[name]; ok {
			agentCfg = existing
		} else {
			agentCfg = &config.AgentConfig{}
		}

		agentCfg.Model = ac.modelValue
		agentCfg.Variant = ac.variant.Value()
		agentCfg.Category = ac.category.Value()
		if v := ac.temperature.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				agentCfg.Temperature = &f
			}
		} else {
			agentCfg.Temperature = nil
		}
		if v := ac.topP.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				agentCfg.TopP = &f
			}
		} else {
			agentCfg.TopP = nil
		}
		if v := ac.skills.Value(); v != "" {
			parts := strings.Split(v, ",")
			var skills []string
			for _, p := range parts {
				if s := strings.TrimSpace(p); s != "" {
					skills = append(skills, s)
				}
			}
			agentCfg.Skills = skills
		} else {
			agentCfg.Skills = nil
		}
		agentCfg.Tools = parseMapStringBool(ac.tools.Value())
		if ac.disable {
			b := true
			agentCfg.Disable = &b
		} else {
			agentCfg.Disable = nil
		}
		agentCfg.Description = ac.description.Value()
		if ac.modeIdx > 0 {
			agentCfg.Mode = agentModes[ac.modeIdx]
		} else {
			agentCfg.Mode = ""
		}
		agentCfg.Color = ac.color.Value()
		agentCfg.Prompt = ac.prompt.Value()
		agentCfg.PromptAppend = ac.promptAppend.Value()

		// Permissions
		if ac.permEditIdx > 0 || ac.permBashIdx > 0 || ac.originalBash != nil ||
			ac.permWebfetchIdx > 0 || ac.permTaskIdx > 0 || ac.permDoomLoopIdx > 0 || ac.permExtDirIdx > 0 {
			agentCfg.Permission = &config.PermissionConfig{}
			if ac.permEditIdx > 0 {
				agentCfg.Permission.Edit = permissionValues[ac.permEditIdx]
			}
			if ac.permBashIdx > 0 {
				agentCfg.Permission.Bash = permissionValues[ac.permBashIdx]
			} else if ac.originalBash != nil {
				agentCfg.Permission.Bash = ac.originalBash
			}
			if ac.permWebfetchIdx > 0 {
				agentCfg.Permission.Webfetch = permissionValues[ac.permWebfetchIdx]
			}
			if ac.permTaskIdx > 0 {
				agentCfg.Permission.Task = permissionValues[ac.permTaskIdx]
			}
			if ac.permDoomLoopIdx > 0 {
				agentCfg.Permission.DoomLoop = permissionValues[ac.permDoomLoopIdx]
			}
			if ac.permExtDirIdx > 0 {
				agentCfg.Permission.ExternalDirectory = permissionValues[ac.permExtDirIdx]
			}
		} else {
			agentCfg.Permission = nil
		}

		if val := ac.maxTokens.Value(); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				agentCfg.MaxTokens = &f
			}
		} else {
			agentCfg.MaxTokens = nil
		}

		if ac.thinkingTypeIdx > 0 || ac.thinkingBudget.Value() != "" {
			if agentCfg.Thinking == nil {
				agentCfg.Thinking = &config.ThinkingConfig{}
			}
			if ac.thinkingTypeIdx > 0 {
				agentCfg.Thinking.Type = thinkingTypes[ac.thinkingTypeIdx]
			}
			if val := ac.thinkingBudget.Value(); val != "" {
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					agentCfg.Thinking.BudgetTokens = &f
				}
			}
		} else if agentCfg.Thinking != nil && ac.thinkingTypeIdx == 0 && ac.thinkingBudget.Value() == "" {
			agentCfg.Thinking = nil
		}

		agentCfg.ReasoningEffort = effortLevels[ac.reasoningEffortIdx]
		agentCfg.TextVerbosity = verbosityLevels[ac.textVerbosityIdx]

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
	ac.maxTokens.Blur()
	ac.thinkingBudget.Blur()

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
	case fieldMaxTokens:
		ac.maxTokens.Focus()
	case fieldThinkingBudget:
		ac.thinkingBudget.Focus()
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
			baseLine += 37 // expanded form ~37 lines
		}
	}
	baseLine++ // current agent header
	baseLine++ // empty line before form

	// Field offsets within the form
	fieldOffsets := map[agentFormField]int{
		fieldModel:           0,
		fieldVariant:         1,
		fieldCategory:        2,
		fieldTemperature:     3,
		fieldTopP:            4,
		fieldSkills:          5,
		fieldTools:           6,
		fieldPrompt:          7,
		fieldPromptAppend:    10,
		fieldDisable:         13,
		fieldDescription:     14,
		fieldMode:            15,
		fieldColor:           16,
		fieldMaxTokens:       17,
		fieldThinkingType:    18,
		fieldThinkingBudget:  19,
		fieldReasoningEffort: 20,
		fieldTextVerbosity:   21,
		fieldProviderOptions: 22,
		fieldPermEdit:        25,
		fieldPermBash:        26,
		fieldPermWebfetch:    27,
		fieldPermTask:        28,
		fieldPermDoomLoop:    29,
		fieldPermExtDir:      30,
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
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "down", "j":
				w.focusedField++
				if w.focusedField > fieldPermExtDir {
					w.focusedField = fieldModel
				}
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "up", "k":
				if w.focusedField > fieldModel {
					w.focusedField--
				} else {
					w.focusedField = fieldPermExtDir
				}
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "tab":
				w.focusedField++
				if w.focusedField > fieldPermExtDir {
					w.focusedField = fieldModel
				}
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "shift+tab":
				if w.focusedField > fieldModel {
					w.focusedField--
				} else {
					w.focusedField = fieldPermExtDir
				}
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "enter":
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
				case fieldThinkingType:
					ac.thinkingTypeIdx = (ac.thinkingTypeIdx + 1) % len(thinkingTypes)
				case fieldReasoningEffort:
					ac.reasoningEffortIdx = (ac.reasoningEffortIdx + 1) % len(effortLevels)
				case fieldTextVerbosity:
					ac.textVerbosityIdx = (ac.textVerbosityIdx + 1) % len(verbosityLevels)
				case fieldPermEdit:
					ac.permEditIdx = (ac.permEditIdx + 1) % len(permissionValues)
				case fieldPermBash:
					// Only cycle if bash is a string (not object)
					if _, isObject := ac.originalBash.(map[string]interface{}); !isObject {
						ac.permBashIdx = (ac.permBashIdx + 1) % len(permissionValues)
					}
				case fieldPermWebfetch:
					ac.permWebfetchIdx = (ac.permWebfetchIdx + 1) % len(permissionValues)
				case fieldPermTask:
					ac.permTaskIdx = (ac.permTaskIdx + 1) % len(permissionValues)
				case fieldPermDoomLoop:
					ac.permDoomLoopIdx = (ac.permDoomLoopIdx + 1) % len(permissionValues)
				case fieldPermExtDir:
					ac.permExtDirIdx = (ac.permExtDirIdx + 1) % len(permissionValues)
				}
				w.viewport.SetContent(w.renderContent())
				return w, nil
			}

			switch w.focusedField {
			case fieldVariant:
				ac.variant, cmd = ac.variant.Update(msg)
				cmds = append(cmds, cmd)
			case fieldCategory:
				ac.category, cmd = ac.category.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTemperature:
				ac.temperature, cmd = ac.temperature.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTopP:
				ac.topP, cmd = ac.topP.Update(msg)
				cmds = append(cmds, cmd)
			case fieldSkills:
				ac.skills, cmd = ac.skills.Update(msg)
				cmds = append(cmds, cmd)
			case fieldTools:
				ac.tools, cmd = ac.tools.Update(msg)
				cmds = append(cmds, cmd)
			case fieldPrompt:
				ac.prompt, cmd = ac.prompt.Update(msg)
				cmds = append(cmds, cmd)
			case fieldPromptAppend:
				ac.promptAppend, cmd = ac.promptAppend.Update(msg)
				cmds = append(cmds, cmd)
			case fieldDescription:
				ac.description, cmd = ac.description.Update(msg)
				cmds = append(cmds, cmd)
			case fieldColor:
				ac.color, cmd = ac.color.Update(msg)
				cmds = append(cmds, cmd)
			case fieldMaxTokens:
				ac.maxTokens, cmd = ac.maxTokens.Update(msg)
				cmds = append(cmds, cmd)
			case fieldThinkingBudget:
				ac.thinkingBudget, cmd = ac.thinkingBudget.Update(msg)
				cmds = append(cmds, cmd)
			}

			w.viewport.SetContent(w.renderContent())
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

	for i, name := range allAgents {
		ac := w.agents[name]

		cursor := "  "
		if i == w.cursor {
			cursor = wizAgentSelectedStyle.Render("> ")
		}

		checkbox := "[ ]"
		if ac.enabled {
			checkbox = wizAgentEnabledStyle.Render("[✓]")
		}

		expandIcon := ""
		if ac.enabled {
			if ac.expanded {
				expandIcon = " ▼"
			} else {
				expandIcon = " ▶"
			}
		}

		nameStyle := wizAgentDimStyle
		if i == w.cursor {
			nameStyle = wizAgentLabelStyle
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

	// Only show focus styling if this is the active agent being edited
	isActiveAgent := name == allAgents[w.cursor]

	renderField := func(label string, field agentFormField, value string) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		return indent[:4] + cursor + style.Render(fmt.Sprintf("%-12s: ", label)) + value
	}

	renderDropdown := func(label string, field agentFormField, options []string, idx int) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent[:4] + cursor + style.Render(fmt.Sprintf("%-12s: ", label)) + val + " [←/→]"
	}

	renderBool := func(label string, field agentFormField, val bool) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent[:4] + cursor + style.Render(fmt.Sprintf("%-12s: ", label)) + checkbox + " [←/→]"
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
	lines = append(lines, renderField("maxTokens", fieldMaxTokens, ac.maxTokens.View()))
	lines = append(lines, renderDropdown("thinking", fieldThinkingType, thinkingTypes, ac.thinkingTypeIdx))
	lines = append(lines, renderField("thinkBudget", fieldThinkingBudget, ac.thinkingBudget.View()))
	lines = append(lines, renderDropdown("reasoning", fieldReasoningEffort, effortLevels, ac.reasoningEffortIdx))
	lines = append(lines, renderDropdown("verbosity", fieldTextVerbosity, verbosityLevels, ac.textVerbosityIdx))
	lines = append(lines, renderField("providerOpts", fieldProviderOptions, ac.providerOptionsDisplay+" (read-only)"))
	lines = append(lines, "")
	lines = append(lines, indent+wizAgentDimStyle.Render("── Permissions ──"))
	lines = append(lines, renderDropdown("edit", fieldPermEdit, permissionValues, ac.permEditIdx))

	// Bash permission - special handling for object type
	bashValue := ""
	switch v := ac.originalBash.(type) {
	case map[string]interface{}:
		bashValue = fmt.Sprintf("[object: %d rules]", len(v))
	default:
		if ac.permBashIdx > 0 && ac.permBashIdx < len(permissionValues) {
			bashValue = permissionValues[ac.permBashIdx]
		} else {
			bashValue = "(none)"
		}
	}
	bashStyle := wizAgentDimStyle
	bashCursor := "  "
	if w.inForm && isActiveAgent && w.focusedField == fieldPermBash {
		bashStyle = wizAgentSelectedStyle
		bashCursor = wizAgentCursorStyle.Render("> ")
	}
	bashHint := " [←/→]"
	if _, isObject := ac.originalBash.(map[string]interface{}); isObject {
		bashHint = " (read-only)"
	}
	lines = append(lines, indent[:4]+bashCursor+bashStyle.Render(fmt.Sprintf("%-12s: ", "bash"))+bashValue+bashHint)

	lines = append(lines, renderDropdown("webfetch", fieldPermWebfetch, permissionValues, ac.permWebfetchIdx))
	lines = append(lines, renderDropdown("task", fieldPermTask, permissionValues, ac.permTaskIdx))
	lines = append(lines, renderDropdown("doom_loop", fieldPermDoomLoop, permissionValues, ac.permDoomLoopIdx))
	lines = append(lines, renderDropdown("external_dir", fieldPermExtDir, permissionValues, ac.permExtDirIdx))
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

	title := wizAgentLabelStyle.Render("Configure Agents")
	desc := wizAgentDimStyle.Render("Space to enable/disable • Enter to expand • Tab to next step")

	if w.inForm {
		desc = wizAgentDimStyle.Render("↑/↓/Tab: navigate • Enter: cycle options • Esc: close form")
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
	var lines []string
	lines = append(lines, wizAgentSelectedStyle.Render("Custom Model"))
	lines = append(lines, "")
	lines = append(lines, wizAgentTextStyle.Render(fmt.Sprintf("Model ID: %s", ac.customModelToSave)))
	lines = append(lines, "")

	if ac.savePromptAnswer == "" {
		lines = append(lines, wizAgentTextStyle.Render("Save this model for future use? (y/n)"))
		lines = append(lines, "")
		lines = append(lines, wizAgentDimStyle.Render("[y] yes  [n] no  [Esc] cancel"))
	} else {
		lines = append(lines, wizAgentTextStyle.Render("Display name:"))
		lines = append(lines, ac.saveDisplayNameInput.View())
		lines = append(lines, "")
		lines = append(lines, wizAgentTextStyle.Render("Provider (optional):"))
		lines = append(lines, ac.saveProviderInput.View())
		lines = append(lines, "")
		if ac.saveError != "" {
			lines = append(lines, wizAgentErrorStyle.Render("Error: "+ac.saveError))
			lines = append(lines, "")
		}
		lines = append(lines, wizAgentDimStyle.Render("[Enter] save  [Tab] next field  [Esc] cancel"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
