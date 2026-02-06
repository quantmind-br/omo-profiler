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

var thinkingTypes = []string{"", "enabled", "disabled"}
var effortLevels = []string{"", "low", "medium", "high", "xhigh"}
var verbosityLevels = []string{"", "low", "medium", "high"}

var (
	wizCatPurple = lipgloss.Color("#7D56F4")
	wizCatGray   = lipgloss.Color("#6C7086")
	wizCatText   = lipgloss.Color("#CDD6F4")
	wizCatRed    = lipgloss.Color("#F38BA8")
)

var (
	wizCatLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizCatText)
	wizCatDimStyle      = lipgloss.NewStyle().Foreground(wizCatGray)
	wizCatSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizCatPurple)
	wizCatTextStyle     = lipgloss.NewStyle().Foreground(wizCatText)
	wizCatErrorStyle    = lipgloss.NewStyle().Foreground(wizCatRed)
)

type categoryFormField int

const (
	catFieldName categoryFormField = iota
	catFieldModel
	catFieldVariant
	catFieldDescription
	catFieldIsUnstable
	catFieldTemperature
	catFieldTopP
	catFieldMaxTokens
	catFieldThinkingType
	catFieldThinkingBudget
	catFieldReasoningEffort
	catFieldTextVerbosity
	catFieldTools
	catFieldPromptAppend
)

type categoryConfig struct {
	name               string
	nameInput          textinput.Model
	modelValue         string
	modelDisplay       string
	description        textinput.Model
	isUnstable         bool
	variant            textinput.Model
	temperature        textinput.Model
	topP               textinput.Model
	maxTokens          textinput.Model
	thinkingTypeIdx    int
	thinkingBudget     textinput.Model
	reasoningEffortIdx int
	textVerbosityIdx   int
	tools              textinput.Model
	promptAppend       textarea.Model
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

	thinkingBudget := textinput.New()
	thinkingBudget.Placeholder = "e.g. 10000"
	thinkingBudget.Width = 10

	tools := textinput.New()
	tools.Placeholder = "tool1:true, tool2:false"
	tools.Width = 40

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
		thinkingBudget:       thinkingBudget,
		tools:                tools,
		promptAppend:         promptAppend,
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

func (w WizardCategories) Init() tea.Cmd {
	return nil
}

func (w *WizardCategories) SetSize(width, height int) {
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

func (w *WizardCategories) SetConfig(cfg *config.Config) {
	w.categories = []*categoryConfig{}

	if cfg.Categories == nil {
		return
	}

	for name, catCfg := range cfg.Categories {
		cc := newCategoryConfig()
		cc.name = name
		cc.nameInput.SetValue(name)

		if catCfg.Model != "" {
			cc.modelValue = catCfg.Model
			cc.modelDisplay = catCfg.Model
		}
		if catCfg.Description != "" {
			cc.description.SetValue(catCfg.Description)
		}
		if catCfg.IsUnstableAgent != nil {
			cc.isUnstable = *catCfg.IsUnstableAgent
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

func (w *WizardCategories) Apply(cfg *config.Config) {
	cfg.Categories = make(map[string]*config.CategoryConfig)

	for _, cc := range w.categories {
		name := strings.TrimSpace(cc.nameInput.Value())
		if name == "" {
			continue
		}

		catCfg := &config.CategoryConfig{}

		if cc.modelValue != "" {
			catCfg.Model = cc.modelValue
		}
		if v := cc.description.Value(); v != "" {
			catCfg.Description = v
		}
		if cc.isUnstable {
			catCfg.IsUnstableAgent = &cc.isUnstable
		}
		if v := cc.variant.Value(); v != "" {
			catCfg.Variant = v
		}
		if v := cc.temperature.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.Temperature = &f
			}
		}
		if v := cc.topP.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.TopP = &f
			}
		}
		if v := cc.maxTokens.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				catCfg.MaxTokens = &f
			}
		}

		if cc.thinkingTypeIdx > 0 {
			catCfg.Thinking = &config.ThinkingConfig{
				Type: thinkingTypes[cc.thinkingTypeIdx],
			}
			if v := cc.thinkingBudget.Value(); v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					catCfg.Thinking.BudgetTokens = &f
				}
			}
		}

		if cc.reasoningEffortIdx > 0 {
			catCfg.ReasoningEffort = effortLevels[cc.reasoningEffortIdx]
		}
		if cc.textVerbosityIdx > 0 {
			catCfg.TextVerbosity = verbosityLevels[cc.textVerbosityIdx]
		}
		if v := cc.tools.Value(); v != "" {
			catCfg.Tools = parseMapStringBool(v)
		}
		if v := cc.promptAppend.Value(); v != "" {
			catCfg.PromptAppend = v
		}

		cfg.Categories[name] = catCfg
	}

	if len(cfg.Categories) == 0 {
		cfg.Categories = nil
	}
}

func (w *WizardCategories) updateFieldFocus(cc *categoryConfig) {
	cc.nameInput.Blur()
	cc.description.Blur()
	cc.variant.Blur()
	cc.temperature.Blur()
	cc.topP.Blur()
	cc.maxTokens.Blur()
	cc.thinkingBudget.Blur()
	cc.tools.Blur()
	cc.promptAppend.Blur()

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
	case catFieldThinkingBudget:
		cc.thinkingBudget.Focus()
	case catFieldTools:
		cc.tools.Focus()
	case catFieldPromptAppend:
		cc.promptAppend.Focus()
	}
}

// getLineForField calculates the viewport line for the focused field
func (w WizardCategories) getLineForField(field categoryFormField) int {
	baseLine := 0
	for i := 0; i < w.cursor; i++ {
		baseLine++ // category header
		if w.categories[i].expanded {
			baseLine += 20 // expanded form ~20 lines
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
		catFieldTemperature:     5,
		catFieldTopP:            6,
		catFieldMaxTokens:       7,
		catFieldThinkingType:    10, // after empty line + thinking label
		catFieldThinkingBudget:  11,
		catFieldReasoningEffort: 13,
		catFieldTextVerbosity:   14,
		catFieldTools:           16,
		catFieldPromptAppend:    17,
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
					if w.focusedField > catFieldPromptAppend {
						w.focusedField = catFieldName
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "up", "k":
					if w.focusedField == catFieldName {
						w.focusedField = catFieldPromptAppend
					} else {
						w.focusedField--
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "tab":
					w.focusedField++
					if w.focusedField > catFieldPromptAppend {
						w.focusedField = catFieldName
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
				case "shift+tab":
					if w.focusedField == catFieldName {
						w.focusedField = catFieldPromptAppend
					} else {
						w.focusedField--
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					w.ensureFieldVisible()
					return w, nil
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

	renderField := func(label string, field categoryFormField, value string) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		return indent + style.Render(fmt.Sprintf("%-16s: ", label)) + value
	}

	renderDropdown := func(label string, field categoryFormField, options []string, idx int) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent + style.Render(fmt.Sprintf("%-16s: ", label)) + val + " [Enter]"
	}

	renderBool := func(label string, field categoryFormField, val bool) string {
		style := wizCatDimStyle
		if w.inForm && w.focusedField == field {
			style = wizCatSelectedStyle
		}
		checkbox := "[ ]"
		if val {
			checkbox = "[✓]"
		}
		return indent + style.Render(fmt.Sprintf("%-16s: ", label)) + checkbox + " [Enter]"
	}

	lines = append(lines, "")
	lines = append(lines, renderField("name", catFieldName, cc.nameInput.View()))
	modelDisplayValue := cc.modelDisplay
	if modelDisplayValue == "" {
		modelDisplayValue = "[Select model...]"
	}
	lines = append(lines, renderField("model", catFieldModel, modelDisplayValue))
	lines = append(lines, renderField("variant", catFieldVariant, cc.variant.View()))
	lines = append(lines, renderField("description", catFieldDescription, cc.description.View()))
	lines = append(lines, renderBool("is_unstable", catFieldIsUnstable, cc.isUnstable))
	lines = append(lines, renderField("temperature", catFieldTemperature, cc.temperature.View()))
	lines = append(lines, renderField("top_p", catFieldTopP, cc.topP.View()))
	lines = append(lines, renderField("max_tokens", catFieldMaxTokens, cc.maxTokens.View()))
	lines = append(lines, "")
	lines = append(lines, indent+wizCatDimStyle.Render("── Thinking ──"))
	lines = append(lines, renderDropdown("type", catFieldThinkingType, thinkingTypes, cc.thinkingTypeIdx))
	lines = append(lines, renderField("budget_tokens", catFieldThinkingBudget, cc.thinkingBudget.View()))
	lines = append(lines, "")
	lines = append(lines, renderDropdown("reasoning_effort", catFieldReasoningEffort, effortLevels, cc.reasoningEffortIdx))
	lines = append(lines, renderDropdown("text_verbosity", catFieldTextVerbosity, verbosityLevels, cc.textVerbosityIdx))
	lines = append(lines, "")
	lines = append(lines, renderField("tools", catFieldTools, cc.tools.View()))
	lines = append(lines, renderField("prompt_append", catFieldPromptAppend, cc.promptAppend.View()))
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
		desc = wizCatDimStyle.Render("↑/↓/Tab: navigate • Enter: select/toggle • Esc: close form")
	}

	content := w.viewport.View()

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
