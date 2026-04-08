package views

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
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

var thinkingTypes = []string{"", "enabled", "disabled"}
var effortLevels = []string{"", "none", "minimal", "low", "medium", "high", "xhigh"}
var verbosityLevels = []string{"", "low", "medium", "high"}

var (
	wizCatPurple = lipgloss.Color("#7D56F4")
	wizCatGray   = lipgloss.Color("#6C7086")
	wizCatText   = lipgloss.Color("#CDD6F4")
	wizCatRed    = lipgloss.Color("#F38BA8")
	wizCatGreen  = lipgloss.Color("#A6E3A1")
)

var (
	wizCatLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizCatText)
	wizCatDimStyle      = lipgloss.NewStyle().Foreground(wizCatGray)
	wizCatSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizCatPurple)
	wizCatTextStyle     = lipgloss.NewStyle().Foreground(wizCatText)
	wizCatErrorStyle    = lipgloss.NewStyle().Foreground(wizCatRed)
	wizCatValidStyle    = lipgloss.NewStyle().Foreground(wizCatGreen)
)

// validateCategoryField returns an inline validation indicator for range fields (temperature, top_p).
// Validation runs on field-exit (focused=false).
func validateCategoryField(label, value string, focused bool) string {
	switch label {
	case "temperature":
		if value == "" {
			return ""
		}
		if focused {
			return ""
		}
		if v, err := strconv.ParseFloat(value, 64); err != nil || math.IsNaN(v) || v < 0 || v > 2 {
			return wizCatErrorStyle.Render(" ✗ must be 0-2")
		}
		return wizCatValidStyle.Render(" ✓")
	case "top_p":
		if value == "" {
			return ""
		}
		if focused {
			return ""
		}
		if v, err := strconv.ParseFloat(value, 64); err != nil || math.IsNaN(v) || v < 0 || v > 1 {
			return wizCatErrorStyle.Render(" ✗ must be 0-1")
		}
		return wizCatValidStyle.Render(" ✓")
	}
	return ""
}

type categoryFormField int

const (
	catFieldName categoryFormField = iota
	catFieldModel
	catFieldVariant
	catFieldDescription
	catFieldIsUnstable
	catFieldDisable
	catFieldTemperature
	catFieldTopP
	catFieldMaxTokens
	catFieldThinkingType
	catFieldThinkingBudget
	catFieldReasoningEffort
	catFieldTextVerbosity
	catFieldTools
	catFieldPromptAppend
	catFieldMaxPromptTokens
	catFieldFallbackModels
)

var selectableCategoryFields = []categoryFormField{
	catFieldModel,
	catFieldVariant,
	catFieldDescription,
	catFieldIsUnstable,
	catFieldDisable,
	catFieldTemperature,
	catFieldTopP,
	catFieldMaxTokens,
	catFieldMaxPromptTokens,
	catFieldThinkingType,
	catFieldThinkingBudget,
	catFieldReasoningEffort,
	catFieldTextVerbosity,
	catFieldTools,
	catFieldPromptAppend,
	catFieldFallbackModels,
}

type categoryConfig struct {
	name               string
	nameInput          textinput.Model
	modelValue         string
	modelDisplay       string
	description        textinput.Model
	isUnstable         bool
	disable            bool
	variant            textinput.Model
	temperature        textinput.Model
	topP               textinput.Model
	maxTokens          textinput.Model
	maxPromptTokens    textinput.Model
	thinkingTypeIdx    int
	thinkingBudget     textinput.Model
	reasoningEffortIdx int
	textVerbosityIdx   int
	tools              textinput.Model
	promptAppend       textarea.Model
	fallbackModels     textinput.Model
	expanded           bool
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

func newCategoryConfig() categoryConfig {
	nameInput := textinput.New()
	nameInput.Placeholder = "category-name"
	nameInput.Width = 30

	description := textinput.New()
	description.Placeholder = "description"
	description.Width = 40

	variant := textinput.New()
	variant.Placeholder = "variant"
	variant.Width = 30

	temperature := textinput.New()
	temperature.Placeholder = "0.0-2.0"
	temperature.Width = 10

	topP := textinput.New()
	topP.Placeholder = "0.0-1.0"
	topP.Width = 10

	maxTokens := textinput.New()
	maxTokens.Placeholder = "e.g. 4096"
	maxTokens.Width = 10

	maxPromptTokens := textinput.New()
	maxPromptTokens.Placeholder = "e.g. 100000"
	maxPromptTokens.Width = 10

	thinkingBudget := textinput.New()
	thinkingBudget.Placeholder = "e.g. 10000"
	thinkingBudget.Width = 10

	tools := textinput.New()
	tools.Placeholder = "tool1:true, tool2:false"
	tools.Width = 40

	fallbackModels := textinput.New()
	fallbackModels.Placeholder = `"model-id" or ["model1", "model2"]`
	fallbackModels.Width = 40

	promptAppend := textarea.New()
	promptAppend.Placeholder = "Append to prompt..."
	promptAppend.SetWidth(50)
	promptAppend.SetHeight(3)

	saveDisplayNameInput := textinput.New()
	saveDisplayNameInput.Placeholder = "Display name"
	saveDisplayNameInput.Width = 30

	saveProviderInput := textinput.New()
	saveProviderInput.Placeholder = "Provider (optional)"
	saveProviderInput.Width = 30

	return categoryConfig{
		nameInput:            nameInput,
		description:          description,
		variant:              variant,
		temperature:          temperature,
		topP:                 topP,
		maxTokens:            maxTokens,
		maxPromptTokens:      maxPromptTokens,
		thinkingBudget:       thinkingBudget,
		tools:                tools,
		promptAppend:         promptAppend,
		fallbackModels:       fallbackModels,
		saveDisplayNameInput: saveDisplayNameInput,
		saveProviderInput:    saveProviderInput,
	}
}

type wizardCategoriesKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	New      key.Binding
	Delete   key.Binding
	Expand   key.Binding
	Next     key.Binding
	Back     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
}

func newWizardCategoriesKeyMap() wizardCategoriesKeyMap {
	return wizardCategoriesKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "collapse"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "expand"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new category"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Expand: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
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
	}
}

// WizardCategories is step 3: Category configuration
type WizardCategories struct {
	categories   []*categoryConfig
	selection    *profile.FieldSelection
	cursor       int
	inForm       bool
	focusedField categoryFormField
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
	keys         wizardCategoriesKeyMap
}

func NewWizardCategories() WizardCategories {
	return WizardCategories{
		categories: []*categoryConfig{},
		keys:       newWizardCategoriesKeyMap(),
	}
}

func (w *WizardCategories) populateDefaults(selection *profile.FieldSelection) {
	w.selection = selection
	w.categories = []*categoryConfig{}
	for _, def := range defaultCategories {
		cc := newCategoryConfig()
		cc.name = def.Name
		cc.nameInput.SetValue(def.Name)
		cc.description.SetValue(def.Description)
		cc.promptAppend.SetValue(def.PromptAppend)
		w.categories = append(w.categories, &cc)
	}
	if selection != nil {
		selection.SetSelected("categories.*.description", true)
		selection.SetSelected("categories.*.prompt_append", true)
	}
}

func (w WizardCategories) Init() tea.Cmd {
	return nil
}

func (w *WizardCategories) SetSize(width, height int) {
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
	if w.categories == nil {
		return
	}

	for _, cc := range w.categories {
		cc.nameInput.Width = layout.MediumFieldWidth(width)
		cc.description.Width = layout.MediumFieldWidth(width)
		cc.variant.Width = layout.MediumFieldWidth(width)
		cc.tools.Width = layout.WideFieldWidth(width, 10)
		cc.maxPromptTokens.Width = 10
		cc.promptAppend.SetWidth(layout.WideFieldWidth(width, 10))
		cc.fallbackModels.Width = layout.WideFieldWidth(width, 10)
		cc.saveDisplayNameInput.Width = layout.MediumFieldWidth(width)
		cc.saveProviderInput.Width = layout.MediumFieldWidth(width)
		cc.modelSelector.SetSize(width, height)
	}
	w.viewport.SetContent(w.renderContent())
}

func (w *WizardCategories) SetConfig(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection
	w.canonicalizeSelectionPaths()

	if cfg.Categories == nil {
		if len(w.categories) == 0 {
			w.populateDefaults(selection)
		}
		return
	}
	w.categories = []*categoryConfig{}

	for name, catCfg := range cfg.Categories {
		cc := newCategoryConfig()
		cc.name = name
		cc.nameInput.SetValue(name)

		if catCfg.Model != "" {
			cc.modelValue = catCfg.Model
			cc.modelDisplay = catCfg.Model
		}
		if catCfg.FallbackModels != nil {
			switch v := catCfg.FallbackModels.(type) {
			case string:
				cc.fallbackModels.SetValue(v)
			default:
				if b, err := json.Marshal(v); err == nil {
					cc.fallbackModels.SetValue(string(b))
				}
			}
		}
		if catCfg.Description != "" {
			cc.description.SetValue(catCfg.Description)
		}
		if catCfg.IsUnstableAgent != nil {
			cc.isUnstable = *catCfg.IsUnstableAgent
		}
		if catCfg.Disable != nil {
			cc.disable = *catCfg.Disable
		}
		if catCfg.Variant != "" {
			cc.variant.SetValue(catCfg.Variant)
		}
		if catCfg.Temperature != nil {
			cc.temperature.SetValue(fmt.Sprintf("%.1f", *catCfg.Temperature))
		}
		if catCfg.TopP != nil {
			cc.topP.SetValue(fmt.Sprintf("%.1f", *catCfg.TopP))
		}
		if catCfg.MaxTokens != nil {
			cc.maxTokens.SetValue(fmt.Sprintf("%.0f", *catCfg.MaxTokens))
		}
		if catCfg.MaxPromptTokens != nil {
			cc.maxPromptTokens.SetValue(fmt.Sprintf("%d", *catCfg.MaxPromptTokens))
		}

		if catCfg.Thinking != nil {
			for i, t := range thinkingTypes {
				if t == catCfg.Thinking.Type {
					cc.thinkingTypeIdx = i
					break
				}
			}
			if catCfg.Thinking.BudgetTokens != nil {
				cc.thinkingBudget.SetValue(fmt.Sprintf("%.0f", *catCfg.Thinking.BudgetTokens))
			}
		}

		if catCfg.ReasoningEffort != "" {
			for i, e := range effortLevels {
				if e == catCfg.ReasoningEffort {
					cc.reasoningEffortIdx = i
					break
				}
			}
		}

		if catCfg.TextVerbosity != "" {
			for i, v := range verbosityLevels {
				if v == catCfg.TextVerbosity {
					cc.textVerbosityIdx = i
					break
				}
			}
		}

		if len(catCfg.Tools) > 0 {
			cc.tools.SetValue(serializeMapStringBool(catCfg.Tools))
		}
		if catCfg.PromptAppend != "" {
			cc.promptAppend.SetValue(catCfg.PromptAppend)
		}

		w.categories = append(w.categories, &cc)
	}
}

func (w *WizardCategories) Apply(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection
	w.canonicalizeSelectionPaths()

	if !w.hasSelectedCategoryFields() {
		cfg.Categories = nil
		return
	}

	cfg.Categories = make(map[string]*config.CategoryConfig)

	for _, cc := range w.categories {
		name := strings.TrimSpace(cc.nameInput.Value())
		if name == "" {
			continue
		}

		catCfg := &config.CategoryConfig{}

		if w.isCategoryFieldSelected(catFieldModel) {
			catCfg.Model = cc.modelValue
		}
		if w.isCategoryFieldSelected(catFieldFallbackModels) {
			v := strings.TrimSpace(cc.fallbackModels.Value())
			v = strings.TrimSpace(v)
			var parsed any
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				catCfg.FallbackModels = parsed
			} else {
				catCfg.FallbackModels = v
			}
		}
		if w.isCategoryFieldSelected(catFieldDescription) {
			catCfg.Description = cc.description.Value()
		}
		if w.isCategoryFieldSelected(catFieldIsUnstable) {
			catCfg.IsUnstableAgent = &cc.isUnstable
		}
		if w.isCategoryFieldSelected(catFieldDisable) {
			catCfg.Disable = &cc.disable
		}
		if w.isCategoryFieldSelected(catFieldVariant) {
			catCfg.Variant = cc.variant.Value()
		}
		if w.isCategoryFieldSelected(catFieldTemperature) {
			v := strings.TrimSpace(cc.temperature.Value())
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.Temperature = &f
			}
		}
		if w.isCategoryFieldSelected(catFieldTopP) {
			v := strings.TrimSpace(cc.topP.Value())
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.TopP = &f
			}
		}
		if w.isCategoryFieldSelected(catFieldMaxTokens) {
			v := strings.TrimSpace(cc.maxTokens.Value())
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.MaxTokens = &f
			}
		}
		if w.isCategoryFieldSelected(catFieldMaxPromptTokens) {
			v := strings.TrimSpace(cc.maxPromptTokens.Value())
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				catCfg.MaxPromptTokens = &i
			}
		}

		if w.isCategoryFieldSelected(catFieldThinkingType) || w.isCategoryFieldSelected(catFieldThinkingBudget) {
			catCfg.Thinking = &config.ThinkingConfig{
				Type: thinkingTypes[cc.thinkingTypeIdx],
			}
			if w.isCategoryFieldSelected(catFieldThinkingBudget) {
				v := strings.TrimSpace(cc.thinkingBudget.Value())
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					catCfg.Thinking.BudgetTokens = &f
				}
			}
		}

		if w.isCategoryFieldSelected(catFieldReasoningEffort) {
			catCfg.ReasoningEffort = effortLevels[cc.reasoningEffortIdx]
		}
		if w.isCategoryFieldSelected(catFieldTextVerbosity) {
			catCfg.TextVerbosity = verbosityLevels[cc.textVerbosityIdx]
		}
		if w.isCategoryFieldSelected(catFieldTools) {
			v := strings.TrimSpace(cc.tools.Value())
			if v == "" {
				catCfg.Tools = map[string]bool{}
			} else {
				catCfg.Tools = parseMapStringBool(v)
			}
		}
		if w.isCategoryFieldSelected(catFieldPromptAppend) {
			catCfg.PromptAppend = cc.promptAppend.Value()
		}

		cfg.Categories[name] = catCfg
	}

	if len(cfg.Categories) == 0 {
		cfg.Categories = nil
	}
}

func categorySelectionPath(field categoryFormField) (string, bool) {
	switch field {
	case catFieldModel:
		return "categories.*.model", true
	case catFieldVariant:
		return "categories.*.variant", true
	case catFieldDescription:
		return "categories.*.description", true
	case catFieldIsUnstable:
		return "categories.*.is_unstable_agent", true
	case catFieldDisable:
		return "categories.*.disable", true
	case catFieldTemperature:
		return "categories.*.temperature", true
	case catFieldTopP:
		return "categories.*.top_p", true
	case catFieldMaxTokens:
		return "categories.*.max_tokens", true
	case catFieldMaxPromptTokens:
		return "categories.*.max_prompt_tokens", true
	case catFieldThinkingType:
		return "categories.*.thinking.type", true
	case catFieldThinkingBudget:
		return "categories.*.thinking.budget_tokens", true
	case catFieldReasoningEffort:
		return "categories.*.reasoning_effort", true
	case catFieldTextVerbosity:
		return "categories.*.text_verbosity", true
	case catFieldTools:
		return "categories.*.tools", true
	case catFieldPromptAppend:
		return "categories.*.prompt_append", true
	case catFieldFallbackModels:
		return "categories.*.fallback_models", true
	default:
		return "", false
	}
}

func categorySelectionAliases(field categoryFormField) []string {
	switch field {
	case catFieldMaxTokens:
		return []string{"categories.*.maxTokens"}
	case catFieldThinkingBudget:
		return []string{"categories.*.thinking.budgetTokens"}
	case catFieldReasoningEffort:
		return []string{"categories.*.reasoningEffort"}
	case catFieldTextVerbosity:
		return []string{"categories.*.textVerbosity"}
	default:
		return nil
	}
}

func (w *WizardCategories) canonicalizeSelectionPaths() {
	if w.selection == nil {
		return
	}

	for _, field := range selectableCategoryFields {
		path, ok := categorySelectionPath(field)
		if !ok {
			continue
		}

		aliases := categorySelectionAliases(field)
		if w.selection.IsSelected(path) {
			for _, alias := range aliases {
				w.selection.SetSelected(alias, false)
			}
			continue
		}

		selected := slices.ContainsFunc(aliases, func(alias string) bool {
			return w.selection.IsSelected(alias)
		})

		if selected {
			w.selection.SetSelected(path, true)
		}

		for _, alias := range aliases {
			w.selection.SetSelected(alias, false)
		}
	}
}

func (w WizardCategories) isCategoryFieldSelected(field categoryFormField) bool {
	path, ok := categorySelectionPath(field)
	if !ok {
		return false
	}
	if w.selection == nil {
		return true
	}
	if w.selection.IsSelected(path) {
		return true
	}
	return slices.ContainsFunc(categorySelectionAliases(field), func(alias string) bool {
		return w.selection.IsSelected(alias)
	})
}

func (w WizardCategories) hasSelectedCategoryFields() bool {
	return slices.ContainsFunc(selectableCategoryFields, w.isCategoryFieldSelected)
}

func (w *WizardCategories) toggleCategoryFieldSelection(field categoryFormField) {
	if w.selection == nil {
		return
	}

	path, ok := categorySelectionPath(field)
	if !ok {
		return
	}

	selected := w.isCategoryFieldSelected(field)
	w.selection.SetSelected(path, !selected)
	for _, alias := range categorySelectionAliases(field) {
		w.selection.SetSelected(alias, false)
	}
}

func (w *WizardCategories) updateFieldFocus(cc *categoryConfig) {
	cc.nameInput.Blur()
	cc.description.Blur()
	cc.variant.Blur()
	cc.temperature.Blur()
	cc.topP.Blur()
	cc.maxTokens.Blur()
	cc.maxPromptTokens.Blur()
	cc.thinkingBudget.Blur()
	cc.tools.Blur()
	cc.promptAppend.Blur()
	cc.fallbackModels.Blur()

	switch w.focusedField {
	case catFieldName:
		cc.nameInput.Focus()
	case catFieldModel:
	case catFieldDescription:
		cc.description.Focus()
	case catFieldVariant:
		cc.variant.Focus()
	case catFieldTemperature:
		cc.temperature.Focus()
	case catFieldTopP:
		cc.topP.Focus()
	case catFieldMaxTokens:
		cc.maxTokens.Focus()
	case catFieldMaxPromptTokens:
		cc.maxPromptTokens.Focus()
	case catFieldThinkingBudget:
		cc.thinkingBudget.Focus()
	case catFieldTools:
		cc.tools.Focus()
	case catFieldPromptAppend:
		cc.promptAppend.Focus()
	case catFieldFallbackModels:
		cc.fallbackModels.Focus()
	}
}

// getLineForField calculates the viewport line for the focused field
func (w WizardCategories) getLineForField(field categoryFormField) int {
	baseLine := 0
	for i := 0; i < w.cursor; i++ {
		baseLine++ // category header
		if w.categories[i].expanded {
			baseLine += 22 // expanded form ~22 lines
		}
	}
	baseLine++ // current category header
	baseLine++ // empty line

	fieldOffsets := map[categoryFormField]int{
		catFieldName:            0,
		catFieldModel:           1,
		catFieldVariant:         2,
		catFieldDescription:     3,
		catFieldIsUnstable:      4,
		catFieldDisable:         5,
		catFieldTemperature:     6,
		catFieldTopP:            7,
		catFieldMaxTokens:       8,
		catFieldMaxPromptTokens: 9,
		catFieldThinkingType:    12, // after empty line + thinking label
		catFieldThinkingBudget:  13,
		catFieldReasoningEffort: 15,
		catFieldTextVerbosity:   16,
		catFieldTools:           18,
		catFieldPromptAppend:    19,
		catFieldFallbackModels:  20,
	}

	return baseLine + fieldOffsets[field]
}

// ensureFieldVisible scrolls the viewport to keep the focused field visible
func (w *WizardCategories) ensureFieldVisible() {
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

func (w WizardCategories) Update(msg tea.Msg) (WizardCategories, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	var currentCategory *categoryConfig
	if len(w.categories) > 0 && w.cursor < len(w.categories) {
		currentCategory = w.categories[w.cursor]
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		if currentCategory != nil && currentCategory.selectingModel {
			currentCategory.modelSelector.SetSize(msg.Width, msg.Height)
		}
		return w, nil

	case ModelSelectedMsg:
		if currentCategory != nil {
			currentCategory.modelValue = msg.ModelID
			currentCategory.modelDisplay = msg.DisplayName
			currentCategory.selectingModel = false
		}
		return w, nil

	case ModelSelectorCancelMsg:
		if currentCategory != nil {
			currentCategory.selectingModel = false
		}
		return w, nil

	case PromptSaveCustomMsg:
		if currentCategory != nil {
			currentCategory.savingCustomModel = true
			currentCategory.customModelToSave = msg.ModelID
			currentCategory.savePromptAnswer = ""
			currentCategory.saveDisplayNameInput.SetValue("")
			currentCategory.saveProviderInput.SetValue("")
			currentCategory.saveError = ""
		}
		return w, nil

	case tea.KeyMsg:
		if currentCategory != nil && currentCategory.selectingModel {
			currentCategory.modelSelector, cmd = currentCategory.modelSelector.Update(msg)
			return w, cmd
		}

		if currentCategory != nil && currentCategory.savingCustomModel {
			return w.handleSaveCustomModel(currentCategory, msg)
		}

		if w.inForm && len(w.categories) > 0 && w.cursor < len(w.categories) {
			cc := w.categories[w.cursor]
			if cc.expanded {
				switch msg.String() {
				case "esc":
					w.inForm = false
					cc.expanded = false
					return w, nil
				case "down", "j":
					w.focusedField++
					if w.focusedField > catFieldFallbackModels {
						w.focusedField = catFieldName
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "up", "k":
					if w.focusedField == catFieldName {
						w.focusedField = catFieldFallbackModels
					} else {
						w.focusedField--
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "tab":
					w.focusedField++
					if w.focusedField > catFieldFallbackModels {
						w.focusedField = catFieldName
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "shift+tab":
					if w.focusedField == catFieldName {
						w.focusedField = catFieldFallbackModels
					} else {
						w.focusedField--
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case " ":
					if _, ok := categorySelectionPath(w.focusedField); ok {
						w.toggleCategoryFieldSelection(w.focusedField)
						w.viewport.SetContent(w.renderContent())
						w.ensureFieldVisible()
						return w, nil
					}
				case "enter":
					if w.focusedField == catFieldModel {
						cc.selectingModel = true
						cc.modelSelector = NewModelSelector()
						cc.modelSelector.SetSize(w.width, w.height)
						return w, nil
					}
					if w.focusedField == catFieldIsUnstable {
						cc.isUnstable = !cc.isUnstable
						return w, nil
					}
					if w.focusedField == catFieldDisable {
						cc.disable = !cc.disable
						return w, nil
					}
					// Cycle through options for dropdown fields
					switch w.focusedField {
					case catFieldThinkingType:
						cc.thinkingTypeIdx = (cc.thinkingTypeIdx + 1) % len(thinkingTypes)
					case catFieldReasoningEffort:
						cc.reasoningEffortIdx = (cc.reasoningEffortIdx + 1) % len(effortLevels)
					case catFieldTextVerbosity:
						cc.textVerbosityIdx = (cc.textVerbosityIdx + 1) % len(verbosityLevels)
					}
					return w, nil
				}

				switch w.focusedField {
				case catFieldName:
					cc.nameInput.Focus()
					cc.nameInput, cmd = cc.nameInput.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldDescription:
					cc.description.Focus()
					cc.description, cmd = cc.description.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldVariant:
					cc.variant.Focus()
					cc.variant, cmd = cc.variant.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldTemperature:
					cc.temperature.Focus()
					cc.temperature, cmd = cc.temperature.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldTopP:
					cc.topP.Focus()
					cc.topP, cmd = cc.topP.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldMaxTokens:
					cc.maxTokens.Focus()
					cc.maxTokens, cmd = cc.maxTokens.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldMaxPromptTokens:
					cc.maxPromptTokens.Focus()
					cc.maxPromptTokens, cmd = cc.maxPromptTokens.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldThinkingBudget:
					cc.thinkingBudget.Focus()
					cc.thinkingBudget, cmd = cc.thinkingBudget.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldTools:
					cc.tools.Focus()
					cc.tools, cmd = cc.tools.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldPromptAppend:
					cc.promptAppend.Focus()
					cc.promptAppend, cmd = cc.promptAppend.Update(msg)
					cmds = append(cmds, cmd)
				case catFieldFallbackModels:
					cc.fallbackModels.Focus()
					cc.fallbackModels, cmd = cc.fallbackModels.Update(msg)
					cmds = append(cmds, cmd)
				}

				return w, tea.Batch(cmds...)
			}
		}

		switch {
		case key.Matches(msg, w.keys.Up):
			if w.cursor > 0 {
				w.cursor--
			}
		case key.Matches(msg, w.keys.Down):
			if w.cursor < len(w.categories)-1 {
				w.cursor++
			}
		case key.Matches(msg, w.keys.Right):
			// Only when not in form mode
			if !w.inForm && len(w.categories) > 0 && w.cursor < len(w.categories) {
				cc := w.categories[w.cursor]
				if !cc.expanded {
					cc.expanded = true
					w.inForm = true
					w.focusedField = catFieldName
					w.updateFieldFocus(cc)
				}
			}
		case key.Matches(msg, w.keys.Left):
			// Only when not in form mode
			if !w.inForm && len(w.categories) > 0 && w.cursor < len(w.categories) {
				cc := w.categories[w.cursor]
				if cc.expanded {
					cc.expanded = false
					w.inForm = false
				}
			}
		case key.Matches(msg, w.keys.New):
			newCat := newCategoryConfig()
			w.categories = append(w.categories, &newCat)
			w.cursor = len(w.categories) - 1
			w.categories[w.cursor].expanded = true
			w.inForm = true
			w.focusedField = catFieldName
			w.categories[w.cursor].nameInput.Focus()
		case key.Matches(msg, w.keys.Delete):
			if len(w.categories) > 0 && w.cursor < len(w.categories) {
				w.categories = append(w.categories[:w.cursor], w.categories[w.cursor+1:]...)
				if w.cursor >= len(w.categories) && w.cursor > 0 {
					w.cursor--
				}
				if len(w.categories) == 0 {
					w.inForm = false
				}
			}
		case key.Matches(msg, w.keys.Expand):
			if len(w.categories) > 0 && w.cursor < len(w.categories) {
				cc := w.categories[w.cursor]
				cc.expanded = !cc.expanded
				if cc.expanded {
					w.inForm = true
					w.focusedField = catFieldName
					w.updateFieldFocus(cc)
				} else {
					w.inForm = false
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

	w.viewport.SetContent(w.renderContent())
	w.viewport, cmd = w.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return w, tea.Batch(cmds...)
}

func (w WizardCategories) renderContent() string {
	var lines []string

	if len(w.categories) == 0 {
		lines = append(lines, wizCatDimStyle.Render("No categories defined. Press 'n' to create one."))
	}

	for i, cc := range w.categories {
		cursor := "  "
		if i == w.cursor {
			cursor = wizCatSelectedStyle.Render("> ")
		}

		expandIcon := ""
		if cc.expanded {
			expandIcon = " ▼"
		} else {
			expandIcon = " ▶"
		}

		nameStyle := wizCatDimStyle
		if i == w.cursor {
			nameStyle = wizCatLabelStyle
		}

		displayName := cc.nameInput.Value()
		if displayName == "" {
			displayName = "(unnamed)"
		}

		line := fmt.Sprintf("%s%s%s", cursor, nameStyle.Render(displayName), expandIcon)
		lines = append(lines, line)

		if cc.expanded {
			lines = append(lines, w.renderCategoryForm(cc)...)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (w WizardCategories) renderCategoryForm(cc *categoryConfig) []string {
	var lines []string

	indent := "      "
	labelFmt := "%-16s: "
	if layout.IsCompact(w.width) {
		indent = "    "
		labelFmt = "%-8s: "
	}

	renderField := func(label string, field categoryFormField, value string) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		return indent + style.Render(fmt.Sprintf(labelFmt, label)) + value
	}

	renderSelectableField := func(label string, field categoryFormField, value string) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}

		checkbox := "[ ]"
		if w.isCategoryFieldSelected(field) {
			checkbox = "[✓]"
		}

		return indent + style.Render(checkbox+" "+fmt.Sprintf(labelFmt, label)) + value
	}

	renderDropdown := func(label string, field categoryFormField, options []string, idx int) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		checkbox := "[ ]"
		if w.isCategoryFieldSelected(field) {
			checkbox = "[✓]"
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent + style.Render(checkbox+" "+fmt.Sprintf(labelFmt, label)) + val + " [Enter]"
	}

	renderBool := func(label string, field categoryFormField, val bool) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		includeCheckbox := "[ ]"
		if w.isCategoryFieldSelected(field) {
			includeCheckbox = "[✓]"
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent + style.Render(includeCheckbox+" "+fmt.Sprintf(labelFmt, label)) + checkbox + " [Enter]"
	}

	lines = append(lines, "")
	lines = append(lines, renderField("name", catFieldName, cc.nameInput.View()))
	modelDisplayValue := cc.modelDisplay
	if modelDisplayValue == "" {
		modelDisplayValue = "[Select model...]"
	}
	lines = append(lines, renderSelectableField("model", catFieldModel, modelDisplayValue))
	lines = append(lines, renderSelectableField("variant", catFieldVariant, cc.variant.View()))
	lines = append(lines, renderSelectableField("description", catFieldDescription, cc.description.View()))
	lines = append(lines, renderBool("is_unstable", catFieldIsUnstable, cc.isUnstable))
	lines = append(lines, renderBool("disable", catFieldDisable, cc.disable))
	lines = append(lines, renderSelectableField("temperature", catFieldTemperature, cc.temperature.View())+validateCategoryField("temperature", cc.temperature.Value(), w.inForm && w.focusedField == catFieldTemperature))
	lines = append(lines, renderSelectableField("top_p", catFieldTopP, cc.topP.View())+validateCategoryField("top_p", cc.topP.Value(), w.inForm && w.focusedField == catFieldTopP))
	lines = append(lines, renderSelectableField("max_tokens", catFieldMaxTokens, cc.maxTokens.View()))
	lines = append(lines, renderSelectableField("max_prompt_tokens", catFieldMaxPromptTokens, cc.maxPromptTokens.View()))
	lines = append(lines, "")
	lines = append(lines, indent+wizCatDimStyle.Render("── Thinking ──"))
	lines = append(lines, renderDropdown("type", catFieldThinkingType, thinkingTypes, cc.thinkingTypeIdx))
	lines = append(lines, renderSelectableField("budget_tokens", catFieldThinkingBudget, cc.thinkingBudget.View()))
	lines = append(lines, "")
	lines = append(lines, renderDropdown("reasoning_effort", catFieldReasoningEffort, effortLevels, cc.reasoningEffortIdx))
	lines = append(lines, renderDropdown("text_verbosity", catFieldTextVerbosity, verbosityLevels, cc.textVerbosityIdx))
	lines = append(lines, "")
	lines = append(lines, renderSelectableField("tools", catFieldTools, cc.tools.View()))
	lines = append(lines, renderSelectableField("prompt_append", catFieldPromptAppend, cc.promptAppend.View()))
	lines = append(lines, renderSelectableField("fallback_models", catFieldFallbackModels, cc.fallbackModels.View()))
	lines = append(lines, "")

	return lines
}

func (w WizardCategories) View() string {
	if len(w.categories) > 0 && w.cursor < len(w.categories) {
		cc := w.categories[w.cursor]
		if cc.selectingModel {
			return cc.modelSelector.View()
		}
		if cc.savingCustomModel {
			return w.renderSaveCustomPrompt(cc)
		}
	}

	title := wizCatLabelStyle.Render("Configure Categories")
	desc := wizCatDimStyle.Render("n: new • d: delete • →: expand • ←: collapse • Enter: edit • Tab: next step")

	if w.inForm {
		desc = wizCatDimStyle.Render("↑/↓/Tab: navigate • Space: toggle include • Enter: select/toggle • Esc: close form")
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

func (w WizardCategories) handleSaveCustomModel(cc *categoryConfig, msg tea.KeyMsg) (WizardCategories, tea.Cmd) {
	if cc.savePromptAnswer == "" {
		switch msg.String() {
		case "y", "Y":
			cc.savePromptAnswer = "y"
			cc.saveFocusedField = 0
			cc.saveDisplayNameInput.Focus()
			return w, textinput.Blink
		case "n", "N":
			cc.savingCustomModel = false
			cc.savePromptAnswer = ""
			return w, nil
		case "esc":
			cc.savingCustomModel = false
			cc.savePromptAnswer = ""
			return w, nil
		}
		return w, nil
	}

	switch msg.String() {
	case "enter":
		displayName := strings.TrimSpace(cc.saveDisplayNameInput.Value())
		if displayName == "" {
			cc.saveError = "Display name is required"
			return w, nil
		}

		registry, err := models.Load()
		if err != nil {
			cc.saveError = err.Error()
			return w, nil
		}

		newModel := models.RegisteredModel{
			DisplayName: displayName,
			ModelID:     cc.customModelToSave,
			Provider:    strings.TrimSpace(cc.saveProviderInput.Value()),
		}

		if err := registry.Add(newModel); err != nil {
			cc.saveError = err.Error()
			return w, nil
		}

		cc.modelDisplay = displayName
		cc.savingCustomModel = false
		cc.savePromptAnswer = ""
		cc.saveError = ""
		return w, nil

	case "esc":
		cc.savingCustomModel = false
		cc.savePromptAnswer = ""
		return w, nil

	case "tab":
		cc.saveFocusedField = (cc.saveFocusedField + 1) % 2
		if cc.saveFocusedField == 0 {
			cc.saveDisplayNameInput.Focus()
			cc.saveProviderInput.Blur()
		} else {
			cc.saveProviderInput.Focus()
			cc.saveDisplayNameInput.Blur()
		}
		return w, nil

	case "shift+tab":
		cc.saveFocusedField = (cc.saveFocusedField + 1) % 2
		if cc.saveFocusedField == 0 {
			cc.saveDisplayNameInput.Focus()
			cc.saveProviderInput.Blur()
		} else {
			cc.saveProviderInput.Focus()
			cc.saveDisplayNameInput.Blur()
		}
		return w, nil
	}

	var cmd tea.Cmd
	if cc.saveFocusedField == 0 {
		cc.saveDisplayNameInput, cmd = cc.saveDisplayNameInput.Update(msg)
	} else {
		cc.saveProviderInput, cmd = cc.saveProviderInput.Update(msg)
	}
	cc.saveError = ""
	return w, cmd
}

func (w WizardCategories) renderSaveCustomPrompt(cc *categoryConfig) string {
	var lines []string
	lines = append(lines, wizCatSelectedStyle.Render("Custom Model"))
	lines = append(lines, "")
	lines = append(lines, wizCatTextStyle.Render(fmt.Sprintf("Model ID: %s", cc.customModelToSave)))
	lines = append(lines, "")

	if cc.savePromptAnswer == "" {
		lines = append(lines, wizCatTextStyle.Render("Save this model for future use? (y/n)"))
		lines = append(lines, "")
		lines = append(lines, wizCatDimStyle.Render("[y] yes  [n] no  [Esc] cancel"))
	} else {
		lines = append(lines, wizCatTextStyle.Render("Display name:"))
		lines = append(lines, cc.saveDisplayNameInput.View())
		lines = append(lines, "")
		lines = append(lines, wizCatTextStyle.Render("Provider (optional):"))
		lines = append(lines, cc.saveProviderInput.View())
		lines = append(lines, "")
		if cc.saveError != "" {
			lines = append(lines, wizCatErrorStyle.Render("Error: "+cc.saveError))
			lines = append(lines, "")
		}
		lines = append(lines, wizCatDimStyle.Render("[Enter] save  [Tab] next field  [Esc] cancel"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
