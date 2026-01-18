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
)

var thinkingTypes = []string{"", "enabled", "disabled"}
var effortLevels = []string{"", "low", "medium", "high"}
var verbosityLevels = []string{"", "low", "medium", "high"}

type categoryFormField int

const (
	catFieldName categoryFormField = iota
	catFieldModel
	catFieldVariant
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
	model              textinput.Model
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
}

func newCategoryConfig() categoryConfig {
	nameInput := textinput.New()
	nameInput.Placeholder = "category-name"
	nameInput.Width = 30

	model := textinput.New()
	model.Placeholder = "model name"
	model.Width = 30

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

	return categoryConfig{
		nameInput:      nameInput,
		model:          model,
		variant:        variant,
		temperature:    temperature,
		topP:           topP,
		maxTokens:      maxTokens,
		thinkingBudget: thinkingBudget,
		tools:          tools,
		promptAppend:   promptAppend,
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
			cc.model.SetValue(catCfg.Model)
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

		if v := cc.model.Value(); v != "" {
			catCfg.Model = v
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
	cc.model.Blur()
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
		cc.model.Focus()
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

func (w WizardCategories) Update(msg tea.Msg) (WizardCategories, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
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
					return w, nil
				case "up", "k":
					if w.focusedField == catFieldName {
						w.focusedField = catFieldPromptAppend
					} else {
						w.focusedField--
					}
					w.updateFieldFocus(cc)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter":
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
				case catFieldModel:
					cc.model.Focus()
					cc.model, cmd = cc.model.Update(msg)
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

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	if len(w.categories) == 0 {
		lines = append(lines, dimStyle.Render("No categories defined. Press 'n' to create one."))
	}

	for i, cc := range w.categories {
		cursor := "  "
		if i == w.cursor {
			cursor = selectedStyle.Render("> ")
		}

		expandIcon := ""
		if cc.expanded {
			expandIcon = " ▼"
		} else {
			expandIcon = " ▶"
		}

		nameStyle := dimStyle
		if i == w.cursor {
			nameStyle = labelStyle
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
	fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	focusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	renderField := func(label string, field categoryFormField, value string) string {
		style := fieldStyle
		if w.inForm && w.focusedField == field {
			style = focusStyle
		}
		return indent + style.Render(fmt.Sprintf("%-16s: ", label)) + value
	}

	renderDropdown := func(label string, field categoryFormField, options []string, idx int) string {
		style := fieldStyle
		if w.inForm && w.focusedField == field {
			style = focusStyle
		}
		val := "(none)"
		if idx > 0 && idx < len(options) {
			val = options[idx]
		}
		return indent + style.Render(fmt.Sprintf("%-16s: ", label)) + val + " [Enter]"
	}

	lines = append(lines, "")
	lines = append(lines, renderField("name", catFieldName, cc.nameInput.View()))
	lines = append(lines, renderField("model", catFieldModel, cc.model.View()))
	lines = append(lines, renderField("variant", catFieldVariant, cc.variant.View()))
	lines = append(lines, renderField("temperature", catFieldTemperature, cc.temperature.View()))
	lines = append(lines, renderField("top_p", catFieldTopP, cc.topP.View()))
	lines = append(lines, renderField("max_tokens", catFieldMaxTokens, cc.maxTokens.View()))
	lines = append(lines, "")
	lines = append(lines, indent+fieldStyle.Render("── Thinking ──"))
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
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	title := titleStyle.Render("Configure Categories")
	desc := helpStyle.Render("n: new • d: delete • →: expand • ←: collapse • Enter: edit • Tab: next step")

	if w.inForm {
		desc = helpStyle.Render("↑/↓: navigate • Enter: cycle options • Esc: close form")
	}

	content := w.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		"",
		content,
	)
}
