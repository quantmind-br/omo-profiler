package views

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
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
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

type fallbackModelEntry struct {
	model           string
	modelDisplay    string
	variant         string
	reasoningEffort string
	rawJSON         string
	isRawJSON       bool
}

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

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// validateAgentField returns an inline validation indicator for the given field.
// For simple fields (color), validation runs on every keystroke (focused=true).
// For range fields (temperature, top_p), validation runs on field-exit (focused=false).
func validateAgentField(label, value string, focused bool) string {
	switch label {
	case "color":
		if value == "" {
			return ""
		}
		if !validationHexRe.MatchString(value) {
			return wizAgentErrorStyle.Render(" ✗ invalid hex")
		}
		return wizAgentValidStyle.Render(" ✓")
	case "temperature":
		if value == "" {
			return ""
		}
		if focused {
			return ""
		}
		if v, err := strconv.ParseFloat(value, 64); err != nil || math.IsNaN(v) || v < 0 || v > 2 {
			return wizAgentErrorStyle.Render(" ✗ must be 0-2")
		}
		return wizAgentValidStyle.Render(" ✓")
	case "top_p":
		if value == "" {
			return ""
		}
		if focused {
			return ""
		}
		if v, err := strconv.ParseFloat(value, 64); err != nil || math.IsNaN(v) || v < 0 || v > 1 {
			return wizAgentErrorStyle.Render(" ✗ must be 0-1")
		}
		return wizAgentValidStyle.Render(" ✓")
	}
	return ""
}

func (w WizardAgents) lastFieldForCurrentAgent() agentFormField {
	if allAgents[w.cursor] == "hephaestus" {
		return fieldAllowNonGpt
	}
	return fieldCompactionVariant
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
	wizAgentValidStyle    = lipgloss.NewStyle().Foreground(wizAgentGreen)
)

var validationHexRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

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
	fieldUltraworkModel
	fieldUltraworkVariant
	fieldReasoningEffort
	fieldTextVerbosity
	fieldProviderOptions
	fieldPermEdit
	fieldPermBash
	fieldPermWebfetch
	fieldPermTask
	fieldPermDoomLoop
	fieldPermExtDir
	fieldFallbackModels
	fieldCompactionModel
	fieldCompactionVariant
	fieldAllowNonGpt
)

type agentConfig struct {
	enabled               bool
	expanded              bool
	modelValue            string
	modelDisplay          string
	variant               textinput.Model
	category              textinput.Model
	temperature           textinput.Model
	topP                  textinput.Model
	skills                textinput.Model
	tools                 textinput.Model
	fallbackModels        textinput.Model
	editingFallbackModels bool
	fallbackEntries       []fallbackModelEntry
	fallbackFocusedIdx    int
	fallbackEditField     int
	fallbackEditInput     textinput.Model
	fallbackEditingField  bool
	fallbackEditingRaw    bool
	fallbackRawInput      textinput.Model
	prompt                textarea.Model
	promptAppend          textarea.Model
	disable               bool
	description           textinput.Model
	modeIdx               int
	color                 textinput.Model
	maxTokens             textinput.Model
	thinkingTypeIdx       int
	thinkingBudget        textinput.Model
	ultraworkModel        textinput.Model
	ultraworkVariant      textinput.Model
	compactionModel       textinput.Model
	compactionVariant     textinput.Model
	allowNonGpt           bool
	reasoningEffortIdx    int
	textVerbosityIdx      int
	providerOptions       map[string]interface{}
	editingProviderOpts   bool
	provOptKeys           []string
	provOptValues         []textinput.Model
	provOptFocusedIdx     int
	provOptEditingVal     bool
	provOptNewKey         textinput.Model
	provOptAddingKey      bool
	// Permissions
	permEditIdx         int
	permBashIdx         int
	permWebfetchIdx     int
	permTaskIdx         int
	permDoomLoopIdx     int
	permExtDirIdx       int
	originalBash        interface{} // Preserve bash object through edit cycle
	editingBashPerms    bool
	bashConvertingToObj bool
	bashRuleKeys        []string
	bashRulePermIdx     []int
	bashRuleFocusedIdx  int
	bashRuleNewTool     textinput.Model
	bashAddingRule      bool
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

	fallbackModels := textinput.New()
	fallbackModels.Placeholder = `"model-id" or ["model1", "model2"]`
	fallbackModels.Width = 40

	fallbackEditInput := textinput.New()
	fallbackEditInput.Placeholder = "variant"
	fallbackEditInput.Width = 30

	fallbackRawInput := textinput.New()
	fallbackRawInput.Placeholder = `{"model":"id"}`
	fallbackRawInput.Width = 40

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

	ultraworkModel := textinput.New()
	ultraworkModel.Placeholder = "model ID"
	ultraworkModel.Width = 30

	ultraworkVariant := textinput.New()
	ultraworkVariant.Placeholder = "variant (optional)"
	ultraworkVariant.Width = 30

	compactionModel := textinput.New()
	compactionModel.Placeholder = "model ID"
	compactionModel.Width = 30

	compactionVariant := textinput.New()
	compactionVariant.Placeholder = "variant (optional)"
	compactionVariant.Width = 30

	saveDisplayNameInput := textinput.New()
	saveDisplayNameInput.Placeholder = "Display name"
	saveDisplayNameInput.Width = 30

	saveProviderInput := textinput.New()
	saveProviderInput.Placeholder = "Provider (optional)"
	saveProviderInput.Width = 30

	bashRuleNewTool := textinput.New()
	bashRuleNewTool.Placeholder = "tool name"
	bashRuleNewTool.Width = 20

	return agentConfig{
		variant:              variant,
		category:             category,
		temperature:          temperature,
		topP:                 topP,
		skills:               skills,
		tools:                tools,
		fallbackModels:       fallbackModels,
		fallbackEditInput:    fallbackEditInput,
		fallbackRawInput:     fallbackRawInput,
		prompt:               prompt,
		promptAppend:         promptAppend,
		description:          description,
		color:                color,
		maxTokens:            maxTokens,
		thinkingBudget:       thinkingBudget,
		ultraworkModel:       ultraworkModel,
		ultraworkVariant:     ultraworkVariant,
		compactionModel:      compactionModel,
		compactionVariant:    compactionVariant,
		saveDisplayNameInput: saveDisplayNameInput,
		saveProviderInput:    saveProviderInput,
		bashRuleNewTool:      bashRuleNewTool,
	}
}

func hasAdvancedFallbackFields(entry map[string]interface{}) bool {
	for key := range entry {
		if key != "model" && key != "variant" && key != "reasoningEffort" {
			return true
		}
	}
	return false
}

func parseFallbackEntries(value interface{}) []fallbackModelEntry {
	if value == nil {
		return nil
	}

	var entries []fallbackModelEntry
	appendString := func(model string) {
		entries = append(entries, fallbackModelEntry{model: model, modelDisplay: model})
	}
	appendObject := func(entry map[string]interface{}) {
		fe := fallbackModelEntry{}
		if m, ok := entry["model"].(string); ok {
			fe.model = m
			fe.modelDisplay = m
		}
		if v, ok := entry["variant"].(string); ok {
			fe.variant = v
		}
		if r, ok := entry["reasoningEffort"].(string); ok {
			fe.reasoningEffort = r
		}
		if hasAdvancedFallbackFields(entry) {
			fe.isRawJSON = true
			if raw, err := json.Marshal(entry); err == nil {
				fe.rawJSON = string(raw)
			}
		}
		entries = append(entries, fe)
	}

	switch v := value.(type) {
	case string:
		appendString(v)
	case []string:
		for _, item := range v {
			appendString(item)
		}
	case []interface{}:
		for _, item := range v {
			switch entry := item.(type) {
			case string:
				appendString(entry)
			case map[string]interface{}:
				appendObject(entry)
			}
		}
	}

	return entries
}

func refreshFallbackRawInput(ac *agentConfig) {
	if len(ac.fallbackEntries) == 0 {
		ac.fallbackModels.SetValue("")
		return
	}

	var fallback interface{}
	if len(ac.fallbackEntries) == 1 && ac.fallbackEntries[0].variant == "" && ac.fallbackEntries[0].reasoningEffort == "" && !ac.fallbackEntries[0].isRawJSON {
		fallback = ac.fallbackEntries[0].model
	} else {
		arr := make([]interface{}, len(ac.fallbackEntries))
		for i, fe := range ac.fallbackEntries {
			if fe.isRawJSON {
				var parsed interface{}
				if err := json.Unmarshal([]byte(fe.rawJSON), &parsed); err == nil {
					arr[i] = parsed
				} else {
					arr[i] = fe.model
				}
				continue
			}
			if fe.variant != "" || fe.reasoningEffort != "" {
				obj := map[string]interface{}{"model": fe.model}
				if fe.variant != "" {
					obj["variant"] = fe.variant
				}
				if fe.reasoningEffort != "" {
					obj["reasoningEffort"] = fe.reasoningEffort
				}
				arr[i] = obj
			} else {
				arr[i] = fe.model
			}
		}
		fallback = arr
	}

	if raw, err := json.Marshal(fallback); err == nil {
		ac.fallbackModels.SetValue(string(raw))
	}
}

func formatFallbackEntry(entry fallbackModelEntry) string {
	if entry.isRawJSON {
		text := entry.rawJSON
		if text == "" {
			text = entry.model
		}
		return fmt.Sprintf("raw %s", text)
	}

	parts := []string{entry.modelDisplay}
	if entry.variant != "" {
		parts = append(parts, "variant="+entry.variant)
	}
	if entry.reasoningEffort != "" {
		parts = append(parts, "reasoning="+entry.reasoningEffort)
	}
	return strings.Join(parts, " • ")
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
	selection    *profile.FieldSelection
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
	overhead := 4
	if layout.IsShort(height) {
		overhead = 3
	}
	if !w.ready {
		w.viewport = viewport.New(width, height-overhead)
		w.ready = true
	} else {
		w.viewport.Width = width
		w.viewport.Height = height - overhead
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
		ac.fallbackModels.Width = layout.WideFieldWidth(width, 10)
		ac.fallbackEditInput.Width = layout.MediumFieldWidth(width)
		ac.fallbackRawInput.Width = layout.WideFieldWidth(width, 10)
		ac.description.Width = layout.WideFieldWidth(width, 10)
		ac.prompt.SetWidth(layout.WideFieldWidth(width, 10))
		ac.promptAppend.SetWidth(layout.WideFieldWidth(width, 10))
		ac.ultraworkModel.Width = layout.MediumFieldWidth(width)
		ac.ultraworkVariant.Width = layout.MediumFieldWidth(width)
		ac.compactionModel.Width = layout.MediumFieldWidth(width)
		ac.compactionVariant.Width = layout.MediumFieldWidth(width)
		ac.saveDisplayNameInput.Width = layout.MediumFieldWidth(width)
		ac.saveProviderInput.Width = layout.MediumFieldWidth(width)
		ac.bashRuleNewTool.Width = layout.MediumFieldWidth(width)
		ac.modelSelector.SetSize(width, height)
	}
	w.viewport.SetContent(w.renderContent())
}

func (w *WizardAgents) SetConfig(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection
	w.canonicalizeSelectionPaths()
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
			ac.fallbackEntries = parseFallbackEntries(agentCfg.FallbackModels)
			ac.fallbackFocusedIdx = 0
			if agentCfg.FallbackModels != nil {
				if raw, err := json.Marshal(agentCfg.FallbackModels); err == nil {
					ac.fallbackModels.SetValue(string(raw))
				}
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
				if bashObj, ok := agentCfg.Permission.Bash.(map[string]interface{}); ok {
					ac.bashRuleKeys = sortedKeys(bashObj)
					ac.bashRulePermIdx = make([]int, len(ac.bashRuleKeys))
					for i, k := range ac.bashRuleKeys {
						if v, ok2 := bashObj[k].(string); ok2 {
							for j, pv := range permissionValues {
								if pv == v {
									ac.bashRulePermIdx[i] = j
									break
								}
							}
						}
					}
				}
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
			if agentCfg.Ultrawork != nil {
				ac.ultraworkModel.SetValue(agentCfg.Ultrawork.Model)
				ac.ultraworkVariant.SetValue(agentCfg.Ultrawork.Variant)
			} else {
				ac.ultraworkModel.SetValue("")
				ac.ultraworkVariant.SetValue("")
			}
			if agentCfg.Compaction != nil {
				ac.compactionModel.SetValue(agentCfg.Compaction.Model)
				ac.compactionVariant.SetValue(agentCfg.Compaction.Variant)
			} else {
				ac.compactionModel.SetValue("")
				ac.compactionVariant.SetValue("")
			}
			if name == "hephaestus" {
				ac.allowNonGpt = false
				if agentCfg.AllowNonGptModel != nil {
					ac.allowNonGpt = *agentCfg.AllowNonGptModel
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
			ac.providerOptions = agentCfg.ProviderOptions
			if ac.providerOptions != nil {
				ac.provOptKeys = sortedKeys(ac.providerOptions)
				ac.provOptValues = make([]textinput.Model, len(ac.provOptKeys))
				for i, k := range ac.provOptKeys {
					v := textinput.New()
					v.Width = 30
					v.SetValue(fmt.Sprintf("%v", ac.providerOptions[k]))
					ac.provOptValues[i] = v
				}
			} else {
				ac.provOptKeys = nil
				ac.provOptValues = nil
			}
		}
	}
}

func (w *WizardAgents) Apply(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection
	w.canonicalizeSelectionPaths()

	if selection == nil {
		w.applyAllAgentFields(cfg)
		return
	}

	if !w.hasSelectedAgentFields() {
		cfg.Agents = nil
		return
	}

	cfg.Agents = make(map[string]*config.AgentConfig)

	for name, ac := range w.agents {
		if !ac.enabled {
			continue
		}

		agentCfg := &config.AgentConfig{}
		hasSelectedFields := false

		if w.isAgentFieldSelected(fieldModel) {
			agentCfg.Model = ac.modelValue
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldFallbackModels) {
			agentCfg.FallbackModels = buildAgentFallbackModelsValue(ac)
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldVariant) {
			agentCfg.Variant = ac.variant.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldCategory) {
			agentCfg.Category = ac.category.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldTemperature) {
			if v := strings.TrimSpace(ac.temperature.Value()); v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					agentCfg.Temperature = &f
				}
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldTopP) {
			if v := strings.TrimSpace(ac.topP.Value()); v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					agentCfg.TopP = &f
				}
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldSkills) {
			if v := strings.TrimSpace(ac.skills.Value()); v != "" {
				parts := strings.Split(v, ",")
				var skills []string
				for _, p := range parts {
					if s := strings.TrimSpace(p); s != "" {
						skills = append(skills, s)
					}
				}
				agentCfg.Skills = skills
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldTools) {
			agentCfg.Tools = parseMapStringBool(ac.tools.Value())
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldPrompt) {
			agentCfg.Prompt = ac.prompt.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldPromptAppend) {
			agentCfg.PromptAppend = ac.promptAppend.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldDisable) {
			disable := ac.disable
			agentCfg.Disable = &disable
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldDescription) {
			agentCfg.Description = ac.description.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldMode) {
			agentCfg.Mode = agentModes[ac.modeIdx]
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldColor) {
			agentCfg.Color = ac.color.Value()
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldMaxTokens) {
			if val := strings.TrimSpace(ac.maxTokens.Value()); val != "" {
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					agentCfg.MaxTokens = &f
				}
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldThinkingType) || w.isAgentFieldSelected(fieldThinkingBudget) {
			agentCfg.Thinking = &config.ThinkingConfig{}
			if w.isAgentFieldSelected(fieldThinkingType) {
				agentCfg.Thinking.Type = thinkingTypes[ac.thinkingTypeIdx]
			}
			if w.isAgentFieldSelected(fieldThinkingBudget) {
				if val := strings.TrimSpace(ac.thinkingBudget.Value()); val != "" {
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						agentCfg.Thinking.BudgetTokens = &f
					}
				}
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldUltraworkModel) || w.isAgentFieldSelected(fieldUltraworkVariant) {
			agentCfg.Ultrawork = &config.UltraworkConfig{}
			if w.isAgentFieldSelected(fieldUltraworkModel) {
				agentCfg.Ultrawork.Model = ac.ultraworkModel.Value()
			}
			if w.isAgentFieldSelected(fieldUltraworkVariant) {
				agentCfg.Ultrawork.Variant = ac.ultraworkVariant.Value()
			}
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldReasoningEffort) {
			agentCfg.ReasoningEffort = effortLevels[ac.reasoningEffortIdx]
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldTextVerbosity) {
			agentCfg.TextVerbosity = verbosityLevels[ac.textVerbosityIdx]
			hasSelectedFields = true
		}
		if w.isAgentFieldSelected(fieldProviderOptions) {
			agentCfg.ProviderOptions = buildAgentProviderOptionsValue(ac)
			hasSelectedFields = true
		}

		permissionSelected := false
		if w.isAgentFieldSelected(fieldPermEdit) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.Edit = permissionValues[ac.permEditIdx]
			permissionSelected = true
		}
		if w.isAgentFieldSelected(fieldPermBash) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.Bash = buildAgentBashPermissionValue(ac)
			permissionSelected = true
		}
		if w.isAgentFieldSelected(fieldPermWebfetch) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.Webfetch = permissionValues[ac.permWebfetchIdx]
			permissionSelected = true
		}
		if w.isAgentFieldSelected(fieldPermTask) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.Task = permissionValues[ac.permTaskIdx]
			permissionSelected = true
		}
		if w.isAgentFieldSelected(fieldPermDoomLoop) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.DoomLoop = permissionValues[ac.permDoomLoopIdx]
			permissionSelected = true
		}
		if w.isAgentFieldSelected(fieldPermExtDir) {
			if agentCfg.Permission == nil {
				agentCfg.Permission = &config.PermissionConfig{}
			}
			agentCfg.Permission.ExternalDirectory = permissionValues[ac.permExtDirIdx]
			permissionSelected = true
		}
		if permissionSelected {
			hasSelectedFields = true
		}

		if w.isAgentFieldSelected(fieldCompactionModel) || w.isAgentFieldSelected(fieldCompactionVariant) {
			agentCfg.Compaction = &config.CompactionConfig{}
			if w.isAgentFieldSelected(fieldCompactionModel) {
				agentCfg.Compaction.Model = ac.compactionModel.Value()
			}
			if w.isAgentFieldSelected(fieldCompactionVariant) {
				agentCfg.Compaction.Variant = ac.compactionVariant.Value()
			}
			hasSelectedFields = true
		}

		if name == "hephaestus" && w.isAgentFieldSelected(fieldAllowNonGpt) {
			allowNonGpt := ac.allowNonGpt
			agentCfg.AllowNonGptModel = &allowNonGpt
			hasSelectedFields = true
		}

		if hasSelectedFields {
			cfg.Agents[name] = agentCfg
		}
	}

	if len(cfg.Agents) == 0 {
		cfg.Agents = nil
	}
}

func (w *WizardAgents) applyAllAgentFields(cfg *config.Config) {
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
		agentCfg.FallbackModels = buildAgentFallbackModelsValue(ac)
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
		if ac.permEditIdx > 0 || ac.permBashIdx > 0 || len(ac.bashRuleKeys) > 0 ||
			ac.permWebfetchIdx > 0 || ac.permTaskIdx > 0 || ac.permDoomLoopIdx > 0 || ac.permExtDirIdx > 0 {
			agentCfg.Permission = &config.PermissionConfig{}
			if ac.permEditIdx > 0 {
				agentCfg.Permission.Edit = permissionValues[ac.permEditIdx]
			}
			if bashValue := buildAgentBashPermissionValue(ac); bashValue != nil {
				agentCfg.Permission.Bash = bashValue
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

		if m := ac.ultraworkModel.Value(); m != "" {
			agentCfg.Ultrawork = &config.UltraworkConfig{
				Model:   m,
				Variant: ac.ultraworkVariant.Value(),
			}
		} else {
			agentCfg.Ultrawork = nil
		}

		if m := ac.compactionModel.Value(); m != "" {
			agentCfg.Compaction = &config.CompactionConfig{
				Model:   m,
				Variant: ac.compactionVariant.Value(),
			}
		} else {
			agentCfg.Compaction = nil
		}

		if name == "hephaestus" {
			if ac.allowNonGpt {
				b := true
				agentCfg.AllowNonGptModel = &b
			} else {
				b := false
				agentCfg.AllowNonGptModel = &b
			}
		} else {
			agentCfg.AllowNonGptModel = nil
		}

		agentCfg.ReasoningEffort = effortLevels[ac.reasoningEffortIdx]
		agentCfg.TextVerbosity = verbosityLevels[ac.textVerbosityIdx]

		agentCfg.ProviderOptions = buildAgentProviderOptionsValue(ac)

		cfg.Agents[name] = agentCfg
	}

	// Remove empty agents map
	if len(cfg.Agents) == 0 {
		cfg.Agents = nil
	}
}

var selectableAgentFields = []agentFormField{
	fieldModel,
	fieldFallbackModels,
	fieldVariant,
	fieldCategory,
	fieldTemperature,
	fieldTopP,
	fieldSkills,
	fieldTools,
	fieldPrompt,
	fieldPromptAppend,
	fieldDisable,
	fieldDescription,
	fieldMode,
	fieldColor,
	fieldMaxTokens,
	fieldThinkingType,
	fieldThinkingBudget,
	fieldUltraworkModel,
	fieldUltraworkVariant,
	fieldReasoningEffort,
	fieldTextVerbosity,
	fieldProviderOptions,
	fieldPermEdit,
	fieldPermBash,
	fieldPermWebfetch,
	fieldPermTask,
	fieldPermDoomLoop,
	fieldPermExtDir,
	fieldCompactionModel,
	fieldCompactionVariant,
	fieldAllowNonGpt,
}

func buildAgentFallbackModelsValue(ac *agentConfig) interface{} {
	if len(ac.fallbackEntries) > 0 {
		if len(ac.fallbackEntries) == 1 && ac.fallbackEntries[0].variant == "" && ac.fallbackEntries[0].reasoningEffort == "" && !ac.fallbackEntries[0].isRawJSON {
			return ac.fallbackEntries[0].model
		}

		arr := make([]interface{}, len(ac.fallbackEntries))
		for i, fe := range ac.fallbackEntries {
			if fe.isRawJSON {
				var parsed interface{}
				if json.Unmarshal([]byte(fe.rawJSON), &parsed) == nil {
					arr[i] = parsed
				} else {
					arr[i] = fe.model
				}
				continue
			}

			if fe.variant != "" || fe.reasoningEffort != "" {
				obj := map[string]interface{}{"model": fe.model}
				if fe.variant != "" {
					obj["variant"] = fe.variant
				}
				if fe.reasoningEffort != "" {
					obj["reasoningEffort"] = fe.reasoningEffort
				}
				arr[i] = obj
				continue
			}

			arr[i] = fe.model
		}
		return arr
	}

	v := strings.TrimSpace(ac.fallbackModels.Value())
	if v == "" {
		return nil
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(v), &parsed); err == nil {
		return parsed
	}

	return v
}

func buildAgentBashPermissionValue(ac *agentConfig) interface{} {
	if len(ac.bashRuleKeys) > 0 {
		bashMap := make(map[string]interface{})
		for i, k := range ac.bashRuleKeys {
			if i < len(ac.bashRulePermIdx) && ac.bashRulePermIdx[i] > 0 {
				bashMap[k] = permissionValues[ac.bashRulePermIdx[i]]
			}
		}
		if len(bashMap) > 0 {
			return bashMap
		}
		return nil
	}

	if ac.permBashIdx > 0 {
		return permissionValues[ac.permBashIdx]
	}

	return nil
}

func buildAgentProviderOptionsValue(ac *agentConfig) map[string]interface{} {
	if len(ac.provOptKeys) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for i, k := range ac.provOptKeys {
		if k == "" {
			continue
		}
		if i >= len(ac.provOptValues) {
			continue
		}

		raw := ac.provOptValues[i].Value()
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			result[k] = f
		} else if b, err := strconv.ParseBool(raw); err == nil {
			result[k] = b
		} else {
			result[k] = raw
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func agentSelectionPath(field agentFormField) (string, bool) {
	switch field {
	case fieldModel:
		return "agents.*.model", true
	case fieldFallbackModels:
		return "agents.*.fallback_models", true
	case fieldVariant:
		return "agents.*.variant", true
	case fieldCategory:
		return "agents.*.category", true
	case fieldTemperature:
		return "agents.*.temperature", true
	case fieldTopP:
		return "agents.*.top_p", true
	case fieldSkills:
		return "agents.*.skills", true
	case fieldTools:
		return "agents.*.tools", true
	case fieldPrompt:
		return "agents.*.prompt", true
	case fieldPromptAppend:
		return "agents.*.prompt_append", true
	case fieldDisable:
		return "agents.*.disable", true
	case fieldDescription:
		return "agents.*.description", true
	case fieldMode:
		return "agents.*.mode", true
	case fieldColor:
		return "agents.*.color", true
	case fieldMaxTokens:
		return "agents.*.max_tokens", true
	case fieldThinkingType:
		return "agents.*.thinking.type", true
	case fieldThinkingBudget:
		return "agents.*.thinking.budget_tokens", true
	case fieldUltraworkModel:
		return "agents.*.ultrawork.model", true
	case fieldUltraworkVariant:
		return "agents.*.ultrawork.variant", true
	case fieldReasoningEffort:
		return "agents.*.reasoning_effort", true
	case fieldTextVerbosity:
		return "agents.*.text_verbosity", true
	case fieldProviderOptions:
		return "agents.*.provider_options", true
	case fieldPermEdit:
		return "agents.*.permission.edit", true
	case fieldPermBash:
		return "agents.*.permission.bash", true
	case fieldPermWebfetch:
		return "agents.*.permission.webfetch", true
	case fieldPermTask:
		return "agents.*.permission.task", true
	case fieldPermDoomLoop:
		return "agents.*.permission.doom_loop", true
	case fieldPermExtDir:
		return "agents.*.permission.external_directory", true
	case fieldCompactionModel:
		return "agents.*.compaction.model", true
	case fieldCompactionVariant:
		return "agents.*.compaction.variant", true
	case fieldAllowNonGpt:
		return "agents.*.allow_non_gpt_model", true
	default:
		return "", false
	}
}

func agentSelectionAliases(field agentFormField) []string {
	switch field {
	case fieldMaxTokens:
		return []string{"agents.*.maxTokens"}
	case fieldThinkingBudget:
		return []string{"agents.*.thinking.budgetTokens"}
	case fieldReasoningEffort:
		return []string{"agents.*.reasoningEffort"}
	case fieldTextVerbosity:
		return []string{"agents.*.textVerbosity"}
	case fieldProviderOptions:
		return []string{"agents.*.providerOptions"}
	default:
		return nil
	}
}

func (w *WizardAgents) canonicalizeSelectionPaths() {
	if w.selection == nil {
		return
	}

	for _, field := range selectableAgentFields {
		path, ok := agentSelectionPath(field)
		if !ok {
			continue
		}

		aliases := agentSelectionAliases(field)
		if w.selection.IsSelected(path) {
			for _, alias := range aliases {
				w.selection.SetSelected(alias, false)
			}
			continue
		}

		selected := false
		for _, alias := range aliases {
			if w.selection.IsSelected(alias) {
				selected = true
				break
			}
		}
		if selected {
			w.selection.SetSelected(path, true)
		}

		for _, alias := range aliases {
			w.selection.SetSelected(alias, false)
		}
	}
}

func (w WizardAgents) isAgentFieldSelected(field agentFormField) bool {
	path, ok := agentSelectionPath(field)
	if !ok {
		return false
	}
	if w.selection == nil {
		return true
	}
	if w.selection.IsSelected(path) {
		return true
	}
	for _, alias := range agentSelectionAliases(field) {
		if w.selection.IsSelected(alias) {
			return true
		}
	}
	return false
}

func (w WizardAgents) hasSelectedAgentFields() bool {
	for _, field := range selectableAgentFields {
		if w.isAgentFieldSelected(field) {
			return true
		}
	}
	return false
}

func (w *WizardAgents) toggleAgentFieldSelection(field agentFormField) {
	if w.selection == nil {
		return
	}

	path, ok := agentSelectionPath(field)
	if !ok {
		return
	}

	selected := w.isAgentFieldSelected(field)
	w.selection.SetSelected(path, !selected)
	for _, alias := range agentSelectionAliases(field) {
		w.selection.SetSelected(alias, false)
	}
}

func (w *WizardAgents) updateFieldFocus(ac *agentConfig) {
	ac.variant.Blur()
	ac.category.Blur()
	ac.temperature.Blur()
	ac.topP.Blur()
	ac.skills.Blur()
	ac.tools.Blur()
	ac.fallbackModels.Blur()
	ac.prompt.Blur()
	ac.promptAppend.Blur()
	ac.description.Blur()
	ac.color.Blur()
	ac.maxTokens.Blur()
	ac.thinkingBudget.Blur()
	ac.ultraworkModel.Blur()
	ac.ultraworkVariant.Blur()
	ac.compactionModel.Blur()
	ac.compactionVariant.Blur()

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
	case fieldFallbackModels:
		if !ac.editingFallbackModels {
			ac.fallbackModels.Focus()
		}
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
	case fieldUltraworkModel:
		ac.ultraworkModel.Focus()
	case fieldUltraworkVariant:
		ac.ultraworkVariant.Focus()
	case fieldCompactionModel:
		ac.compactionModel.Focus()
	case fieldCompactionVariant:
		ac.compactionVariant.Focus()
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
			formHeight := 43
			if allAgents[i] == "hephaestus" {
				formHeight = 44
			}
			baseLine += formHeight
		}
	}
	baseLine++ // current agent header
	baseLine++ // empty line before form

	// Field offsets within the form
	fieldOffsets := map[agentFormField]int{
		fieldModel:             0,
		fieldVariant:           1,
		fieldCategory:          2,
		fieldTemperature:       3,
		fieldTopP:              4,
		fieldSkills:            5,
		fieldTools:             6,
		fieldPrompt:            7,
		fieldPromptAppend:      10,
		fieldDisable:           13,
		fieldDescription:       14,
		fieldMode:              15,
		fieldColor:             16,
		fieldMaxTokens:         17,
		fieldThinkingType:      18,
		fieldThinkingBudget:    19,
		fieldUltraworkModel:    20,
		fieldUltraworkVariant:  21,
		fieldReasoningEffort:   22,
		fieldTextVerbosity:     23,
		fieldProviderOptions:   24,
		fieldPermEdit:          27,
		fieldPermBash:          28,
		fieldPermWebfetch:      29,
		fieldPermTask:          30,
		fieldPermDoomLoop:      31,
		fieldPermExtDir:        32,
		fieldFallbackModels:    33,
		fieldCompactionModel:   34,
		fieldCompactionVariant: 35,
		fieldAllowNonGpt:       36,
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
		if ac.editingFallbackModels {
			ac.fallbackEntries = append(ac.fallbackEntries, fallbackModelEntry{
				model:        msg.ModelID,
				modelDisplay: msg.DisplayName,
			})
			ac.fallbackFocusedIdx = len(ac.fallbackEntries) - 1
			refreshFallbackRawInput(ac)
		} else {
			ac.modelValue = msg.ModelID
			ac.modelDisplay = msg.DisplayName
		}
		ac.selectingModel = false
		w.viewport.SetContent(w.renderContent())
		return w, nil

	case ModelSelectorCancelMsg:
		ac.selectingModel = false
		w.viewport.SetContent(w.renderContent())
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

		if ac.editingProviderOpts {
			return w.handleProviderOptsEditor(ac, msg)
		}

		if ac.editingFallbackModels {
			return w.handleFallbackModelsEditor(ac, msg)
		}

		if ac.editingBashPerms || ac.bashConvertingToObj {
			return w.handleBashPermsEditor(ac, msg)
		}

		// When in form editing mode
		if w.inForm && ac.expanded {
			lastField := w.lastFieldForCurrentAgent()
			nextField := func() {
				w.focusedField++
				if w.focusedField > lastField {
					w.focusedField = fieldModel
				}
				if w.focusedField == fieldAllowNonGpt && currentAgent != "hephaestus" {
					w.focusedField = fieldModel
				}
			}
			prevField := func() {
				if w.focusedField > fieldModel {
					w.focusedField--
				} else {
					w.focusedField = lastField
				}
				if w.focusedField == fieldAllowNonGpt && currentAgent != "hephaestus" {
					w.focusedField = lastField
				}
			}
			switch msg.String() {
			case "esc":
				w.inForm = false
				ac.expanded = false
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "down", "j":
				nextField()
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "up", "k":
				prevField()
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "tab":
				nextField()
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case "shift+tab":
				prevField()
				w.updateFieldFocus(ac)
				w.viewport.SetContent(w.renderContent())
				w.ensureFieldVisible()
				return w, nil
			case " ":
				if _, ok := agentSelectionPath(w.focusedField); ok {
					w.toggleAgentFieldSelection(w.focusedField)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				}
			case "enter":
				switch w.focusedField {
				case fieldModel:
					ac.selectingModel = true
					ac.modelSelector = NewModelSelector()
					ac.modelSelector.SetSize(w.width, w.height)
					return w, nil
				case fieldFallbackModels:
					ac.editingFallbackModels = true
					if ac.fallbackFocusedIdx >= len(ac.fallbackEntries) && len(ac.fallbackEntries) > 0 {
						ac.fallbackFocusedIdx = len(ac.fallbackEntries) - 1
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case fieldProviderOptions:
					ac.editingProviderOpts = true
					ac.provOptFocusedIdx = 0
					ac.provOptEditingVal = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case fieldDisable:
					ac.disable = !ac.disable
				case fieldAllowNonGpt:
					ac.allowNonGpt = !ac.allowNonGpt
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
					if len(ac.bashRuleKeys) > 0 {
						ac.editingBashPerms = true
						ac.bashRuleFocusedIdx = 0
						ac.bashAddingRule = false
						ac.bashConvertingToObj = false
					} else {
						ac.bashConvertingToObj = true
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
			case fieldFallbackModels:
				ac.fallbackModels, cmd = ac.fallbackModels.Update(msg)
				cmds = append(cmds, cmd)
				trimmed := strings.TrimSpace(ac.fallbackModels.Value())
				if trimmed == "" {
					ac.fallbackEntries = nil
					ac.fallbackFocusedIdx = 0
				} else {
					var parsed interface{}
					if json.Unmarshal([]byte(trimmed), &parsed) == nil {
						ac.fallbackEntries = parseFallbackEntries(parsed)
						if len(ac.fallbackEntries) == 0 {
							ac.fallbackFocusedIdx = 0
						} else if ac.fallbackFocusedIdx >= len(ac.fallbackEntries) {
							ac.fallbackFocusedIdx = len(ac.fallbackEntries) - 1
						}
					}
				}
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
			case fieldUltraworkModel:
				ac.ultraworkModel, cmd = ac.ultraworkModel.Update(msg)
				cmds = append(cmds, cmd)
			case fieldUltraworkVariant:
				ac.ultraworkVariant, cmd = ac.ultraworkVariant.Update(msg)
				cmds = append(cmds, cmd)
			case fieldCompactionModel:
				ac.compactionModel, cmd = ac.compactionModel.Update(msg)
				cmds = append(cmds, cmd)
			case fieldCompactionVariant:
				ac.compactionVariant, cmd = ac.compactionVariant.Update(msg)
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
	labelFmt := "%-12s: "
	if layout.IsCompact(w.width) {
		indent = "    "
		labelFmt = "%-8s: "
	}

	// Only show focus styling if this is the active agent being edited
	isActiveAgent := name == allAgents[w.cursor]

	renderField := func(label string, field agentFormField, value string) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		includeCheckbox := "[ ]"
		if w.isAgentFieldSelected(field) {
			includeCheckbox = "[✓]"
		}
		return indent[:len(indent)-2] + cursor + style.Render(includeCheckbox+" "+fmt.Sprintf(labelFmt, label)) + value
	}

	renderDropdown := func(label string, field agentFormField, options []string, idx int) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		includeCheckbox := "[ ]"
		if w.isAgentFieldSelected(field) {
			includeCheckbox = "[✓]"
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent[:len(indent)-2] + cursor + style.Render(includeCheckbox+" "+fmt.Sprintf(labelFmt, label)) + val + " [Enter]"
	}

	renderBool := func(label string, field agentFormField, val bool) string {
		style := wizAgentDimStyle
		cursor := "  "
		if w.inForm && isActiveAgent && w.focusedField == field {
			style = wizAgentSelectedStyle
			cursor = wizAgentCursorStyle.Render("> ")
		}
		includeCheckbox := "[ ]"
		if w.isAgentFieldSelected(field) {
			includeCheckbox = "[✓]"
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent[:len(indent)-2] + cursor + style.Render(includeCheckbox+" "+fmt.Sprintf(labelFmt, label)) + checkbox + " [Enter]"
	}

	lines = append(lines, "")
	modelDisplayValue := ac.modelDisplay
	if modelDisplayValue == "" {
		modelDisplayValue = "[Select model...]"
	}
	lines = append(lines, renderField("model", fieldModel, modelDisplayValue))
	lines = append(lines, renderField("variant", fieldVariant, ac.variant.View()))
	lines = append(lines, renderField("category", fieldCategory, ac.category.View()))
	lines = append(lines, renderField("temperature", fieldTemperature, ac.temperature.View())+validateAgentField("temperature", ac.temperature.Value(), isActiveAgent && w.focusedField == fieldTemperature))
	lines = append(lines, renderField("top_p", fieldTopP, ac.topP.View())+validateAgentField("top_p", ac.topP.Value(), isActiveAgent && w.focusedField == fieldTopP))
	lines = append(lines, renderField("skills", fieldSkills, ac.skills.View()))
	lines = append(lines, renderField("tools", fieldTools, ac.tools.View()))
	lines = append(lines, renderField("prompt", fieldPrompt, ac.prompt.View()))
	lines = append(lines, renderField("prompt_append", fieldPromptAppend, ac.promptAppend.View()))
	lines = append(lines, renderBool("disable", fieldDisable, ac.disable))
	lines = append(lines, renderField("description", fieldDescription, ac.description.View()))
	lines = append(lines, renderDropdown("mode", fieldMode, agentModes, ac.modeIdx))
	lines = append(lines, renderField("color", fieldColor, ac.color.View())+validateAgentField("color", ac.color.Value(), isActiveAgent && w.focusedField == fieldColor))
	lines = append(lines, renderField("maxTokens", fieldMaxTokens, ac.maxTokens.View()))
	lines = append(lines, renderDropdown("thinking", fieldThinkingType, thinkingTypes, ac.thinkingTypeIdx))
	lines = append(lines, renderField("thinkBudget", fieldThinkingBudget, ac.thinkingBudget.View()))
	lines = append(lines, renderField("ultraModel", fieldUltraworkModel, ac.ultraworkModel.View()))
	lines = append(lines, renderField("ultraVariant", fieldUltraworkVariant, ac.ultraworkVariant.View()))
	lines = append(lines, renderDropdown("reasoning", fieldReasoningEffort, effortLevels, ac.reasoningEffortIdx))
	lines = append(lines, renderDropdown("verbosity", fieldTextVerbosity, verbosityLevels, ac.textVerbosityIdx))
	if ac.editingProviderOpts {
		lines = append(lines, renderField("providerOpts", fieldProviderOptions, "[editing]"))
		if ac.provOptAddingKey {
			lines = append(lines, indent+"  "+wizAgentDimStyle.Render("new key: ")+ac.provOptNewKey.View())
		} else {
			if len(ac.provOptKeys) == 0 {
				lines = append(lines, indent+"  "+wizAgentDimStyle.Render("(empty) press 'a' to add"))
			}
			for i, k := range ac.provOptKeys {
				cursor := "  "
				if i == ac.provOptFocusedIdx {
					cursor = wizAgentCursorStyle.Render("> ")
				}
				val := ""
				if i < len(ac.provOptValues) {
					val = ac.provOptValues[i].View()
				}
				lines = append(lines, indent+cursor+wizAgentTextStyle.Render(k+": ")+val)
			}
			lines = append(lines, indent+"  "+wizAgentDimStyle.Render("a:add d:del ↑↓:nav enter:edit esc:done"))
		}
	} else {
		count := len(ac.provOptKeys)
		if count > 0 {
			lines = append(lines, renderField("providerOpts", fieldProviderOptions, fmt.Sprintf("%d options set [Enter to edit]", count)))
		} else {
			lines = append(lines, renderField("providerOpts", fieldProviderOptions, "(none) [Enter to edit]"))
		}
	}
	lines = append(lines, "")
	lines = append(lines, indent+wizAgentDimStyle.Render("── Permissions ──"))
	lines = append(lines, renderDropdown("edit", fieldPermEdit, permissionValues, ac.permEditIdx))

	// Bash permission - editable for both string and object modes
	if ac.editingBashPerms {
		lines = append(lines, indent+wizAgentSelectedStyle.Render("┌─ Editing Bash Permissions ─┐"))
		if len(ac.bashRuleKeys) == 0 && !ac.bashAddingRule {
			lines = append(lines, indent+"  "+wizAgentDimStyle.Render("(empty) press 'a' to add"))
		}
		for i, k := range ac.bashRuleKeys {
			cursor := "  "
			if i == ac.bashRuleFocusedIdx {
				cursor = wizAgentCursorStyle.Render("> ")
			}
			perm := "(none)"
			if i < len(ac.bashRulePermIdx) && ac.bashRulePermIdx[i] > 0 && ac.bashRulePermIdx[i] < len(permissionValues) {
				perm = permissionValues[ac.bashRulePermIdx[i]]
			}
			lines = append(lines, indent+cursor+wizAgentTextStyle.Render(k+": ")+perm+" [←/→]")
		}
		if ac.bashAddingRule {
			lines = append(lines, indent+"  "+wizAgentTextStyle.Render("New tool: ")+ac.bashRuleNewTool.View())
		}
		helpText := "a:add d:del esc:done [?] more"
		if ac.bashAddingRule {
			helpText = "enter:confirm esc:cancel"
		}
		lines = append(lines, indent+"  "+wizAgentDimStyle.Render(helpText))
	} else if ac.bashConvertingToObj {
		lines = append(lines, renderField("bash", fieldPermBash, "Convert to per-command rules? (y/n)"))
	} else if len(ac.bashRuleKeys) > 0 {
		lines = append(lines, renderField("bash", fieldPermBash, fmt.Sprintf("%d rules [Enter to edit]", len(ac.bashRuleKeys))))
	} else {
		lines = append(lines, renderDropdown("bash", fieldPermBash, permissionValues, ac.permBashIdx))
	}

	lines = append(lines, renderDropdown("webfetch", fieldPermWebfetch, permissionValues, ac.permWebfetchIdx))
	lines = append(lines, renderDropdown("task", fieldPermTask, permissionValues, ac.permTaskIdx))
	lines = append(lines, renderDropdown("doom_loop", fieldPermDoomLoop, permissionValues, ac.permDoomLoopIdx))
	lines = append(lines, renderDropdown("external_dir", fieldPermExtDir, permissionValues, ac.permExtDirIdx))
	lines = append(lines, "")
	lines = append(lines, indent+wizAgentDimStyle.Render("── Fallback ──"))
	if ac.editingFallbackModels {
		lines = append(lines, indent+wizAgentSelectedStyle.Render("┌─ Editing Fallback Models ─┐"))
		if len(ac.fallbackEntries) == 0 {
			lines = append(lines, indent+"  "+wizAgentDimStyle.Render("(empty) press 'a' to add"))
		}
		for i, entry := range ac.fallbackEntries {
			cursor := "  "
			if i == ac.fallbackFocusedIdx {
				cursor = wizAgentCursorStyle.Render("> ")
			}
			if i == ac.fallbackFocusedIdx && ac.fallbackEditingRaw {
				lines = append(lines, indent+cursor+wizAgentTextStyle.Render("raw ")+ac.fallbackRawInput.View())
				continue
			}
			if i == ac.fallbackFocusedIdx && ac.fallbackEditingField && ac.fallbackEditField == 1 {
				lines = append(lines, indent+cursor+wizAgentTextStyle.Render(entry.modelDisplay+" • variant=")+ac.fallbackEditInput.View())
				continue
			}
			lines = append(lines, indent+cursor+wizAgentTextStyle.Render(formatFallbackEntry(entry)))
		}
		helpText := "a:add d:del esc:done [?] more"
		if ac.fallbackEditingField || ac.fallbackEditingRaw {
			helpText = "enter:save esc:cancel"
		}
		lines = append(lines, indent+"  "+wizAgentDimStyle.Render(helpText))
		if raw := strings.TrimSpace(ac.fallbackModels.Value()); raw != "" {
			lines = append(lines, indent+"  "+wizAgentDimStyle.Render("raw: ")+raw)
		}
	} else if len(ac.fallbackEntries) > 0 {
		lines = append(lines, renderField("fallback", fieldFallbackModels, fmt.Sprintf("%d models [Enter to edit]", len(ac.fallbackEntries))))
	} else {
		lines = append(lines, renderField("fallback", fieldFallbackModels, "(none) [Enter to edit]"))
	}
	lines = append(lines, "")
	lines = append(lines, indent+wizAgentDimStyle.Render("── Compaction ──"))
	lines = append(lines, renderField("compModel", fieldCompactionModel, ac.compactionModel.View()))
	lines = append(lines, renderField("compVariant", fieldCompactionVariant, ac.compactionVariant.View()))
	if name == "hephaestus" {
		lines = append(lines, renderBool("allow_non_gpt", fieldAllowNonGpt, ac.allowNonGpt))
	}
	lines = append(lines, "")

	return lines
}

func (w WizardAgents) handleFallbackModelsEditor(ac *agentConfig, msg tea.KeyMsg) (WizardAgents, tea.Cmd) {
	if len(ac.fallbackEntries) == 0 {
		ac.fallbackFocusedIdx = 0
	} else if ac.fallbackFocusedIdx >= len(ac.fallbackEntries) {
		ac.fallbackFocusedIdx = len(ac.fallbackEntries) - 1
	}

	if ac.fallbackEditingRaw {
		if len(ac.fallbackEntries) > 0 && ac.fallbackFocusedIdx < len(ac.fallbackEntries) {
			entry := &ac.fallbackEntries[ac.fallbackFocusedIdx]
			switch msg.String() {
			case "enter", "esc":
				entry.rawJSON = ac.fallbackRawInput.Value()
				ac.fallbackEditingRaw = false
				ac.fallbackRawInput.Blur()
				refreshFallbackRawInput(ac)
				w.viewport.SetContent(w.renderContent())
				return w, nil
			default:
				var cmd tea.Cmd
				ac.fallbackRawInput, cmd = ac.fallbackRawInput.Update(msg)
				w.viewport.SetContent(w.renderContent())
				return w, cmd
			}
		}
		ac.fallbackEditingRaw = false
		ac.fallbackRawInput.Blur()
	}

	if ac.fallbackEditingField {
		if len(ac.fallbackEntries) > 0 && ac.fallbackFocusedIdx < len(ac.fallbackEntries) {
			entry := &ac.fallbackEntries[ac.fallbackFocusedIdx]
			switch msg.String() {
			case "enter":
				entry.variant = strings.TrimSpace(ac.fallbackEditInput.Value())
				ac.fallbackEditingField = false
				ac.fallbackEditInput.Blur()
				ac.fallbackEditField = 2
				refreshFallbackRawInput(ac)
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "esc":
				ac.fallbackEditingField = false
				ac.fallbackEditInput.Blur()
				w.viewport.SetContent(w.renderContent())
				return w, nil
			default:
				var cmd tea.Cmd
				ac.fallbackEditInput, cmd = ac.fallbackEditInput.Update(msg)
				w.viewport.SetContent(w.renderContent())
				return w, cmd
			}
		}
		ac.fallbackEditingField = false
		ac.fallbackEditInput.Blur()
	}

	switch msg.String() {
	case "esc":
		ac.fallbackEditingField = false
		ac.fallbackEditingRaw = false
		ac.fallbackEditInput.Blur()
		ac.fallbackRawInput.Blur()
		ac.editingFallbackModels = false
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "a":
		ac.selectingModel = true
		ac.modelSelector = NewModelSelector()
		ac.modelSelector.SetSize(w.width, w.height)
		return w, nil
	case "d":
		if len(ac.fallbackEntries) > 0 && ac.fallbackFocusedIdx < len(ac.fallbackEntries) {
			ac.fallbackEntries = append(ac.fallbackEntries[:ac.fallbackFocusedIdx], ac.fallbackEntries[ac.fallbackFocusedIdx+1:]...)
			if ac.fallbackFocusedIdx >= len(ac.fallbackEntries) && ac.fallbackFocusedIdx > 0 {
				ac.fallbackFocusedIdx--
			}
			ac.fallbackEditingField = false
			ac.fallbackEditingRaw = false
			ac.fallbackEditInput.Blur()
			ac.fallbackRawInput.Blur()
			refreshFallbackRawInput(ac)
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "up", "k":
		ac.fallbackEditingField = false
		ac.fallbackEditingRaw = false
		ac.fallbackEditInput.Blur()
		ac.fallbackRawInput.Blur()
		if ac.fallbackFocusedIdx > 0 {
			ac.fallbackFocusedIdx--
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "down", "j":
		ac.fallbackEditingField = false
		ac.fallbackEditingRaw = false
		ac.fallbackEditInput.Blur()
		ac.fallbackRawInput.Blur()
		if ac.fallbackFocusedIdx < len(ac.fallbackEntries)-1 {
			ac.fallbackFocusedIdx++
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "e":
		if len(ac.fallbackEntries) > 0 && ac.fallbackFocusedIdx < len(ac.fallbackEntries) {
			entry := &ac.fallbackEntries[ac.fallbackFocusedIdx]
			if entry.isRawJSON {
				ac.fallbackEditingRaw = true
				ac.fallbackRawInput.SetValue(entry.rawJSON)
				ac.fallbackRawInput.Focus()
				w.viewport.SetContent(w.renderContent())
				return w, textinput.Blink
			}
			switch ac.fallbackEditField {
			case 0:
				if entry.model == "" {
					ac.selectingModel = true
					ac.modelSelector = NewModelSelector()
					ac.modelSelector.SetSize(w.width, w.height)
					return w, nil
				}
				ac.fallbackEditField = 1
			case 1:
				ac.fallbackEditingField = true
				ac.fallbackEditInput.SetValue(entry.variant)
				ac.fallbackEditInput.Focus()
				w.viewport.SetContent(w.renderContent())
				return w, textinput.Blink
			case 2:
				current := 0
				for i, level := range effortLevels {
					if level == entry.reasoningEffort {
						current = i
						break
					}
				}
				entry.reasoningEffort = effortLevels[(current+1)%len(effortLevels)]
				ac.fallbackEditField = 0
			}
			refreshFallbackRawInput(ac)
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "r":
		if len(ac.fallbackEntries) > 0 && ac.fallbackFocusedIdx < len(ac.fallbackEntries) {
			entry := &ac.fallbackEntries[ac.fallbackFocusedIdx]
			if entry.isRawJSON {
				ac.fallbackEditingRaw = true
				ac.fallbackRawInput.SetValue(entry.rawJSON)
				ac.fallbackRawInput.Focus()
				w.viewport.SetContent(w.renderContent())
				return w, textinput.Blink
			} else {
				payload := map[string]interface{}{"model": entry.model}
				if entry.variant != "" {
					payload["variant"] = entry.variant
				}
				if entry.reasoningEffort != "" {
					payload["reasoningEffort"] = entry.reasoningEffort
				}
				entry.isRawJSON = true
				if raw, err := json.Marshal(payload); err == nil {
					entry.rawJSON = string(raw)
				}
			}
			refreshFallbackRawInput(ac)
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	}

	return w, nil
}

func (w WizardAgents) handleProviderOptsEditor(ac *agentConfig, msg tea.KeyMsg) (WizardAgents, tea.Cmd) {
	if ac.provOptAddingKey {
		switch msg.String() {
		case "esc":
			ac.provOptAddingKey = false
			ac.provOptNewKey.Blur()
			w.viewport.SetContent(w.renderContent())
			return w, nil
		case "enter":
			keyName := strings.TrimSpace(ac.provOptNewKey.Value())
			if keyName != "" {
				ac.provOptKeys = append(ac.provOptKeys, keyName)
				v := textinput.New()
				v.Width = 30
				ac.provOptValues = append(ac.provOptValues, v)
				ac.provOptFocusedIdx = len(ac.provOptKeys) - 1
			}
			ac.provOptAddingKey = false
			ac.provOptNewKey.Blur()
			ac.provOptNewKey.SetValue("")
			w.viewport.SetContent(w.renderContent())
			return w, nil
		default:
			var cmd tea.Cmd
			ac.provOptNewKey, cmd = ac.provOptNewKey.Update(msg)
			w.viewport.SetContent(w.renderContent())
			return w, cmd
		}
	}

	if ac.provOptEditingVal {
		switch msg.String() {
		case "esc", "enter":
			ac.provOptEditingVal = false
			if ac.provOptFocusedIdx < len(ac.provOptValues) {
				ac.provOptValues[ac.provOptFocusedIdx].Blur()
			}
			w.viewport.SetContent(w.renderContent())
			return w, nil
		}
		if ac.provOptFocusedIdx < len(ac.provOptValues) {
			var cmd tea.Cmd
			ac.provOptValues[ac.provOptFocusedIdx], cmd = ac.provOptValues[ac.provOptFocusedIdx].Update(msg)
			w.viewport.SetContent(w.renderContent())
			return w, cmd
		}
		return w, nil
	}

	switch msg.String() {
	case "esc":
		ac.editingProviderOpts = false
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "a":
		ac.provOptAddingKey = true
		ac.provOptNewKey = textinput.New()
		ac.provOptNewKey.Width = 20
		ac.provOptNewKey.Placeholder = "key name"
		ac.provOptNewKey.Focus()
		w.viewport.SetContent(w.renderContent())
		return w, textinput.Blink
	case "d":
		if len(ac.provOptKeys) > 0 && ac.provOptFocusedIdx < len(ac.provOptKeys) {
			ac.provOptKeys = append(ac.provOptKeys[:ac.provOptFocusedIdx], ac.provOptKeys[ac.provOptFocusedIdx+1:]...)
			ac.provOptValues = append(ac.provOptValues[:ac.provOptFocusedIdx], ac.provOptValues[ac.provOptFocusedIdx+1:]...)
			if ac.provOptFocusedIdx >= len(ac.provOptKeys) && ac.provOptFocusedIdx > 0 {
				ac.provOptFocusedIdx--
			}
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "up", "k":
		if ac.provOptFocusedIdx > 0 {
			ac.provOptFocusedIdx--
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "down", "j":
		if ac.provOptFocusedIdx < len(ac.provOptKeys)-1 {
			ac.provOptFocusedIdx++
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "enter":
		if ac.provOptFocusedIdx < len(ac.provOptValues) {
			ac.provOptEditingVal = true
			ac.provOptValues[ac.provOptFocusedIdx].Focus()
			w.viewport.SetContent(w.renderContent())
			return w, textinput.Blink
		}
		return w, nil
	}

	return w, nil
}

func (w WizardAgents) handleBashPermsEditor(ac *agentConfig, msg tea.KeyMsg) (WizardAgents, tea.Cmd) {
	if ac.bashConvertingToObj {
		switch msg.String() {
		case "y", "Y":
			ac.bashConvertingToObj = false
			permIdx := ac.permBashIdx
			if permIdx <= 0 {
				permIdx = 1
			}
			ac.bashRuleKeys = []string{"bash"}
			ac.bashRulePermIdx = []int{permIdx}
			ac.bashRuleFocusedIdx = 0
			ac.editingBashPerms = true
			w.viewport.SetContent(w.renderContent())
			return w, nil
		case "n", "N", "esc":
			ac.bashConvertingToObj = false
			w.viewport.SetContent(w.renderContent())
			return w, nil
		}
		return w, nil
	}

	if ac.bashAddingRule {
		switch msg.String() {
		case "esc":
			ac.bashAddingRule = false
			ac.bashRuleNewTool.Blur()
			w.viewport.SetContent(w.renderContent())
			return w, nil
		case "enter":
			name := strings.TrimSpace(ac.bashRuleNewTool.Value())
			if name != "" {
				ac.bashRuleKeys = append(ac.bashRuleKeys, name)
				ac.bashRulePermIdx = append(ac.bashRulePermIdx, 1)
				ac.bashRuleFocusedIdx = len(ac.bashRuleKeys) - 1
			}
			ac.bashAddingRule = false
			ac.bashRuleNewTool.SetValue("")
			ac.bashRuleNewTool.Blur()
			w.viewport.SetContent(w.renderContent())
			return w, nil
		}
		var cmd tea.Cmd
		ac.bashRuleNewTool, cmd = ac.bashRuleNewTool.Update(msg)
		w.viewport.SetContent(w.renderContent())
		return w, cmd
	}

	switch msg.String() {
	case "esc":
		ac.editingBashPerms = false
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "c":
		ac.bashRuleKeys = nil
		ac.bashRulePermIdx = nil
		ac.editingBashPerms = false
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "a":
		ac.bashAddingRule = true
		ac.bashRuleNewTool.SetValue("")
		ac.bashRuleNewTool.Focus()
		w.viewport.SetContent(w.renderContent())
		return w, textinput.Blink
	case "d":
		if len(ac.bashRuleKeys) > 0 && ac.bashRuleFocusedIdx < len(ac.bashRuleKeys) {
			ac.bashRuleKeys = append(ac.bashRuleKeys[:ac.bashRuleFocusedIdx], ac.bashRuleKeys[ac.bashRuleFocusedIdx+1:]...)
			ac.bashRulePermIdx = append(ac.bashRulePermIdx[:ac.bashRuleFocusedIdx], ac.bashRulePermIdx[ac.bashRuleFocusedIdx+1:]...)
			if ac.bashRuleFocusedIdx >= len(ac.bashRuleKeys) && ac.bashRuleFocusedIdx > 0 {
				ac.bashRuleFocusedIdx--
			}
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "up", "k":
		if ac.bashRuleFocusedIdx > 0 {
			ac.bashRuleFocusedIdx--
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "down", "j":
		if ac.bashRuleFocusedIdx < len(ac.bashRuleKeys)-1 {
			ac.bashRuleFocusedIdx++
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "left", "h":
		if ac.bashRuleFocusedIdx < len(ac.bashRulePermIdx) {
			idx := ac.bashRulePermIdx[ac.bashRuleFocusedIdx]
			idx--
			if idx < 1 {
				idx = len(permissionValues) - 1
			}
			ac.bashRulePermIdx[ac.bashRuleFocusedIdx] = idx
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	case "right", "l":
		if ac.bashRuleFocusedIdx < len(ac.bashRulePermIdx) {
			idx := ac.bashRulePermIdx[ac.bashRuleFocusedIdx]
			idx++
			if idx >= len(permissionValues) {
				idx = 1
			}
			ac.bashRulePermIdx[ac.bashRuleFocusedIdx] = idx
		}
		w.viewport.SetContent(w.renderContent())
		return w, nil
	}

	return w, nil
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
	desc := wizAgentDimStyle.Render("[Space] toggle  [Enter] expand  [Tab] next step")

	if w.inForm {
		desc = wizAgentDimStyle.Render("↑/↓/Tab: navigate • Space: toggle include • Enter: edit/value • Esc: close form")
	}

	content := w.viewport.View()

	if layout.IsShort(w.height) {
		return lipgloss.JoinVertical(lipgloss.Left,
			title,
			desc,
			content,
		)
	}

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
