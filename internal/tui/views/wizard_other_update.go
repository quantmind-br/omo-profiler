package views

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
)

type fieldBinding struct {
	section   otherSection
	subCursor int
	update    func(*WizardOther, tea.KeyMsg) tea.Cmd
}

func (w *WizardOther) updateTextInputField(model *textinput.Model, msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc":
		model.Blur()
		w.inSubSection = false
		w.refreshView()
		return nil
	case "up", "k":
		model.Blur()
		if w.subCursor > 0 {
			w.subCursor--
		}
		w.refreshView()
		return nil
	case "down", "j":
		model.Blur()
		w.subCursor++
		w.refreshView()
		return nil
	case "tab":
		model.Blur()
		w.inSubSection = false
		w.refreshView()
		return nil
	case "enter", " ":
		model.Focus()
		*model, cmd = model.Update(msg)
		return cmd
	default:
		model.Focus()
		*model, cmd = model.Update(msg)
		return cmd
	}
}

func (w *WizardOther) updateTextareaField(model *textarea.Model, msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc", "tab":
		model.Blur()
		w.inSubSection = false
		w.refreshView()
		return nil
	default:
		model.Focus()
		*model, cmd = model.Update(msg)
		return cmd
	}
}

func (w *WizardOther) fieldBindings() []fieldBinding {
	return []fieldBinding{
		{section: sectionExperimental, subCursor: 6, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.dcpTurnProtTurns, msg) }},
		{section: sectionExperimental, subCursor: 7, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.dcpProtectedTools, msg) }},
		{section: sectionExperimental, subCursor: 12, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.dcpPurgeErrorsTurns, msg)
		}},
		{section: sectionExperimental, subCursor: 15, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.expPluginLoadTimeoutMs, msg)
		}},
		{section: sectionExperimental, subCursor: 20, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.expMaxTools, msg) }},
		{section: sectionClaudeCode, subCursor: 6, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.ccPluginsOverride, msg) }},
		{section: sectionBackgroundTask, subCursor: 0, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btProviderConcurrency, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btModelConcurrency, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 2, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btDefaultConcurrency, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 3, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.btStaleTimeoutMs, msg) }},
		{section: sectionBackgroundTask, subCursor: 4, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btMessageStalenessTimeoutMs, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 5, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btSyncPollTimeoutMs, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 6, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.btMaxDepth, msg) }},
		{section: sectionBackgroundTask, subCursor: 7, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.btTaskTtlMs, msg) }},
		{section: sectionBackgroundTask, subCursor: 8, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btSessionGoneTimeoutMs, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 9, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.btMaxToolCalls, msg) }},
		{section: sectionBackgroundTask, subCursor: 10, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btCircuitBreakerMaxCalls, msg)
		}},
		{section: sectionBackgroundTask, subCursor: 11, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.btCircuitBreakerConsecutive, msg)
		}},
		{section: sectionDisabledMcps, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.disabledMcps, msg) }},
		{section: sectionDisabledTools, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.disabledTools, msg) }},
		{section: sectionMcpEnvAllowlist, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.mcpEnvAllowlist, msg) }},
		{section: sectionRalphLoop, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.rlDefaultMaxIterations, msg)
		}},
		{section: sectionRalphLoop, subCursor: 2, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.rlStateDir, msg) }},
		{section: sectionGitMaster, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.gmCommitFooterText, msg)
		}},
		{section: sectionGitMaster, subCursor: 3, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.gmGitEnvPrefix, msg) }},
		{section: sectionCommentChecker, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.ccCustomPrompt, msg) }},
		{section: sectionBabysitting, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.babysittingTimeoutMs, msg)
		}},
		{section: sectionTmux, subCursor: 2, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.tmuxMainPaneSize, msg) }},
		{section: sectionTmux, subCursor: 3, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.tmuxMainPaneMinWidth, msg)
		}},
		{section: sectionTmux, subCursor: 4, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.tmuxAgentPaneMinWidth, msg)
		}},
		{section: sectionSisyphus, subCursor: 0, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.sisyphusTasksStoragePath, msg)
		}},
		{section: sectionSisyphus, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.sisyphusTasksTaskListID, msg)
		}},
		{section: sectionDefaultRunAgent, subCursor: -1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.defaultRunAgent, msg) }},
		{section: sectionModelCapabilities, subCursor: 2, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextInputField(&w.mcRefreshTimeoutMs, msg)
		}},
		{section: sectionModelCapabilities, subCursor: 3, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextInputField(&w.mcSourceURL, msg) }},
		{section: sectionOpenclaw, subCursor: 4, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextareaField(&w.openclawEditor, msg) }},
		{section: sectionRuntimeFallback, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd {
			return w.updateTextareaField(&w.runtimeFallbackEditor, msg)
		}},
		{section: sectionSkillsJson, subCursor: 1, update: func(w *WizardOther, msg tea.KeyMsg) tea.Cmd { return w.updateTextareaField(&w.skillsEditor, msg) }},
	}
}

func (w *WizardOther) dispatchFocusedField(msg tea.KeyMsg) (tea.Cmd, bool) {
	for _, binding := range w.fieldBindings() {
		if binding.section != w.currentSection {
			continue
		}
		if binding.subCursor >= 0 && binding.subCursor != w.subCursor {
			continue
		}
		return binding.update(w, msg), true
	}
	return nil, false
}

func (w WizardOther) Update(msg tea.Msg) (WizardOther, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		if w.inSubSection && !w.subValueFocused {
			switch msg.String() {
			case "esc", "tab":
				w.inSubSection = false
				w.subValueFocused = false
				w.refreshView()
				return w, nil
			case "up", "k":
				if w.subCursor > 0 {
					w.subCursor--
				}
				w.refreshView()
				return w, nil
			case "down", "j":
				w.subCursor++
				w.refreshView()
				return w, nil
			case " ":
				w.toggleSubItem()
				w.refreshView()
				return w, nil
			case "enter", "right", "l":
				path := w.subSectionFieldPath(w.currentSection, w.subCursor)
				if path == "" && (w.currentSection != sectionOpenclaw || w.subCursor != 4) && (w.currentSection != sectionRuntimeFallback || w.subCursor != 1) && (w.currentSection != sectionSkillsJson || w.subCursor != 1) {
					return w, nil
				}
				w.subValueFocused = true
				w.refreshView()
				return w, nil
			}
		}

		if w.inSubSection {
			if w.currentSection == sectionExperimental && w.subCursor == 4 {
				switch msg.String() {
				case "right", "l":
					w.dcpNotificationIdx = (w.dcpNotificationIdx + 1) % len(dcpNotificationValues)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.dcpNotificationIdx = (w.dcpNotificationIdx - 1 + len(dcpNotificationValues)) % len(dcpNotificationValues)
					w.refreshView()
					return w, nil
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 3 {
				switch msg.String() {
				case "right", "l":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx + 1) % len(ralphLoopStrategies)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx - 1 + len(ralphLoopStrategies)) % len(ralphLoopStrategies)
					w.refreshView()
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 5 {
				switch msg.String() {
				case "right", "l":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx + 1) % len(tmuxIsolations)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx - 1 + len(tmuxIsolations)) % len(tmuxIsolations)
					w.refreshView()
					return w, nil
				}
			}

			if w.currentSection == sectionBrowserAutomationEngine && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.browserProviderIdx = (w.browserProviderIdx + 1) % len(browserProviders)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.browserProviderIdx = (w.browserProviderIdx - 1 + len(browserProviders)) % len(browserProviders)
					w.refreshView()
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 1 {
				switch msg.String() {
				case "right", "l":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx + 1) % len(tmuxLayouts)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx - 1 + len(tmuxLayouts)) % len(tmuxLayouts)
					w.refreshView()
					return w, nil
				}
			}

			if w.currentSection == sectionWebsearch && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.websearchProviderIdx = (w.websearchProviderIdx + 1) % len(websearchProviders)
					w.refreshView()
					return w, nil
				case "left", "h":
					w.websearchProviderIdx = (w.websearchProviderIdx - 1 + len(websearchProviders)) % len(websearchProviders)
					w.refreshView()
					return w, nil
				}
			}

			if cmd, ok := w.dispatchFocusedField(msg); ok {
				return w, cmd
			}

			switch msg.String() {

			case "esc":
				w.inSubSection = false
			case "up", "k":
				if w.subCursor > 0 {
					w.subCursor--
				}
			case "down", "j":
				w.subCursor++
			case " ", "enter":
				w.toggleSubItem()
			case "tab":
				w.inSubSection = false
			}
			w.refreshView()
			return w, nil
		}

		switch {
		case key.Matches(msg, w.keys.Up):
			w.simpleValueFocused = false
			items := w.buildVisibleItems()
			idx := w.findCurrentVisibleIndex(items)
			if idx > 0 {
				w.applyCursorPosition(items[idx-1])
			}
		case key.Matches(msg, w.keys.Down):
			w.simpleValueFocused = false
			items := w.buildVisibleItems()
			idx := w.findCurrentVisibleIndex(items)
			if idx < len(items)-1 {
				w.applyCursorPosition(items[idx+1])
			}
		case key.Matches(msg, w.keys.Toggle):
			if w.inCategory {
				w.toggleSection()
			}
		case key.Matches(msg, w.keys.Expand):
			if !w.inCategory {
				// On category header: toggle expand
				w.categoryExpanded[w.currentCategory] = !w.categoryExpanded[w.currentCategory]
				if w.categoryExpanded[w.currentCategory] {
					sections := categorySections[int(w.currentCategory)]
					if len(sections) > 0 {
						w.currentSection = sections[0]
						w.inCategory = true
					}
				}
			} else {
				// On a section within an expanded category
				if w.isSimpleBooleanSection(w.currentSection) {
					w.toggleSection()
				} else {
					w.sectionExpanded[w.currentSection] = !w.sectionExpanded[w.currentSection]
					if w.sectionExpanded[w.currentSection] {
						w.inSubSection = true
						w.subCursor = 0
						w.subValueFocused = false
					}
				}
			}
		case key.Matches(msg, w.keys.Right):
			if !w.inCategory {
				// On category header: expand and enter first section
				if !w.categoryExpanded[w.currentCategory] {
					w.categoryExpanded[w.currentCategory] = true
					sections := categorySections[int(w.currentCategory)]
					if len(sections) > 0 {
						w.currentSection = sections[0]
						w.inCategory = true
					}
				}
			} else {
				// On a section: expand section
				if w.isSimpleBooleanSection(w.currentSection) {
					w.simpleValueFocused = true
				} else if !w.inSubSection && !w.sectionExpanded[w.currentSection] {
					w.sectionExpanded[w.currentSection] = true
					w.inSubSection = true
					w.subCursor = 0
					w.subValueFocused = false
				}
			}
		case key.Matches(msg, w.keys.Left):
			if !w.inCategory {
				// On category header: collapse category
				if w.categoryExpanded[w.currentCategory] {
					w.collapseCategory(w.currentCategory)
				}
			} else {
				// On a section within a category
				if w.isSimpleBooleanSection(w.currentSection) {
					w.simpleValueFocused = false
				} else if !w.inSubSection && w.sectionExpanded[w.currentSection] {
					w.sectionExpanded[w.currentSection] = false
				} else if !w.inSubSection {
					// Section already collapsed, go back to category header
					w.inCategory = false
				}
			}
		case key.Matches(msg, w.keys.Next):
			return w, func() tea.Msg { return WizardNextMsg{} }
		case key.Matches(msg, w.keys.Back):
			if w.inCategory {
				w.inCategory = false
			} else {
				return w, func() tea.Msg { return WizardBackMsg{} }
			}
		}
	}

	// Update viewport
	w.refreshView()
	w.viewport, cmd = w.viewport.Update(msg)

	return w, cmd
}

func (w *WizardOther) toggleSection() {
	switch w.currentSection {
	case sectionAutoUpdate:
		if w.simpleValueFocused {
			w.autoUpdate = !w.autoUpdate
		} else {
			w.toggleFieldSelection(autoUpdateFieldPath)
		}
	case sectionNewTaskSystemEnabled:
		if w.simpleValueFocused {
			w.newTaskSystemEnabled = !w.newTaskSystemEnabled
		} else {
			w.toggleFieldSelection(newTaskSystemEnabledFieldPath)
		}
	case sectionHashlineEdit:
		if w.simpleValueFocused {
			w.hashlineEdit = !w.hashlineEdit
		} else {
			w.toggleFieldSelection(hashlineEditFieldPath)
		}
	case sectionModelFallback:
		if w.simpleValueFocused {
			w.modelFallback = !w.modelFallback
		} else {
			w.toggleFieldSelection(modelFallbackFieldPath)
		}
	case sectionStartWork:
		if w.simpleValueFocused {
			w.startWorkAutoCommit = !w.startWorkAutoCommit
		} else {
			w.toggleFieldSelection(startWorkAutoCommitFieldPath)
		}
	case sectionDefaultRunAgent:
		w.toggleFieldSelection(defaultRunAgentFieldPath)
	case sectionDisabledMcps:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledMcps))
	case sectionDisabledAgents:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledAgents))
	case sectionDisabledSkills:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledSkills))
	case sectionDisabledCommands:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledCommands))
	case sectionDisabledTools:
		w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledTools))
	}
}

func (w *WizardOther) toggleSubItem() {
	// Handle Row 0 inclusion toggle for disabled-list sections
	if w.subCursor == 0 {
		switch w.currentSection {
		case sectionDisabledMcps, sectionDisabledAgents, sectionDisabledSkills,
			sectionDisabledCommands, sectionDisabledTools:
			w.toggleFieldSelection(w.topLevelFieldPath(w.currentSection))
			return
		}
	}

	path := w.subSectionFieldPath(w.currentSection, w.subCursor)
	if path != "" && !w.subValueFocused {
		w.toggleFieldSelection(path)
		return
	}

	switch w.currentSection {
	case sectionDisabledAgents:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableAgents) {
			agent := disableableAgents[idx]
			w.disabledAgents[agent] = !w.disabledAgents[agent]
		}
	case sectionDisabledSkills:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableSkills) {
			skill := disableableSkills[idx]
			w.disabledSkills[skill] = !w.disabledSkills[skill]
		}
	case sectionDisabledCommands:
		idx := w.subCursor - 1
		if idx >= 0 && idx < len(disableableCommands) {
			cmd := disableableCommands[idx]
			w.disabledCommands[cmd] = !w.disabledCommands[cmd]
		}
	case sectionExperimental:
		switch w.subCursor {
		case 0:
			w.expAggressiveTrunc = !w.expAggressiveTrunc
		case 1:
			w.expAutoResume = !w.expAutoResume
		case 2:
			w.expTruncateAllOutputs = !w.expTruncateAllOutputs
		case 3:
			w.dcpEnabled = !w.dcpEnabled
		case 5:
			w.dcpTurnProtEnabled = !w.dcpTurnProtEnabled
		case 8:
			w.dcpDeduplicationEnabled = !w.dcpDeduplicationEnabled
		case 9:
			w.dcpSupersedeWritesEnabled = !w.dcpSupersedeWritesEnabled
		case 10:
			w.dcpSupersedeWritesAggressive = !w.dcpSupersedeWritesAggressive
		case 11:
			w.dcpPurgeErrorsEnabled = !w.dcpPurgeErrorsEnabled
		case 13:
			w.expPreemptiveCompaction = !w.expPreemptiveCompaction
		case 14:
			w.expTaskSystem = !w.expTaskSystem
		case 16:
			w.expSafeHookCreation = !w.expSafeHookCreation
		case 17:
			w.expHashlineEdit = !w.expHashlineEdit
		case 18:
			w.expDisableOmoEnv = !w.expDisableOmoEnv
		case 19:
			w.expModelFallbackTitle = !w.expModelFallbackTitle
		}
	case sectionClaudeCode:
		switch w.subCursor {
		case 0:
			w.ccMcp = !w.ccMcp
		case 1:
			w.ccCommands = !w.ccCommands
		case 2:
			w.ccSkills = !w.ccSkills
		case 3:
			w.ccAgents = !w.ccAgents
		case 4:
			w.ccHooks = !w.ccHooks
		case 5:
			w.ccPlugins = !w.ccPlugins
		}
	case sectionSisyphusAgent:
		switch w.subCursor {
		case 0:
			w.saDisabled = !w.saDisabled
		case 1:
			w.saDefaultBuilderEnabled = !w.saDefaultBuilderEnabled
		case 2:
			w.saPlannerEnabled = !w.saPlannerEnabled
		case 3:
			w.saReplacePlan = !w.saReplacePlan
		case 4:
			w.saTDD = !w.saTDD
		}
	case sectionRalphLoop:
		if w.subCursor == 0 {
			w.rlEnabled = !w.rlEnabled
		}
	case sectionNotification:
		if w.subCursor == 0 {
			w.notifForceEnable = !w.notifForceEnable
		}
	case sectionGitMaster:
		switch w.subCursor {
		case 0:
			w.gmCommitFooter = !w.gmCommitFooter
		case 2:
			w.gmIncludeCoAuthoredBy = !w.gmIncludeCoAuthoredBy
		}
	case sectionTmux:
		if w.subCursor == 0 {
			w.tmuxEnabled = !w.tmuxEnabled
		}
	case sectionSisyphus:
		if w.subCursor == 2 {
			w.sisyphusTasksClaudeCodeCompat = !w.sisyphusTasksClaudeCodeCompat
		}
	case sectionModelCapabilities:
		switch w.subCursor {
		case 0:
			w.mcEnabled = !w.mcEnabled
		case 1:
			w.mcAutoRefreshOnStart = !w.mcAutoRefreshOnStart
		}
	case sectionOpenclaw:
		if w.subCursor == 0 {
			var parsed config.OpenclawConfig
			if err := json.Unmarshal([]byte(w.openclawEditor.Value()), &parsed); err == nil && parsed.Enabled != nil {
				parsed.Enabled = wizardOtherBoolPtr(!*parsed.Enabled)
				if raw, err := json.MarshalIndent(parsed, "", "  "); err == nil {
					w.openclawEditor.SetValue(string(raw))
				}
			}
		}
	}
}

// Category navigation helpers

type visibleItemKind int

const (
	itemCategory visibleItemKind = iota
	itemSection
)

type visibleItem struct {
	kind     visibleItemKind
	category otherCategory
	section  otherSection // only valid when kind == itemSection
}

func (w *WizardOther) buildVisibleItems() []visibleItem {
	var items []visibleItem
	for ci := range otherCategoryNames {
		cat := otherCategory(ci)
		items = append(items, visibleItem{kind: itemCategory, category: cat})
		if w.categoryExpanded[cat] {
			for _, sec := range categorySections[ci] {
				items = append(items, visibleItem{kind: itemSection, category: cat, section: sec})
			}
		}
	}
	return items
}

func (w *WizardOther) findCurrentVisibleIndex(items []visibleItem) int {
	if w.inCategory {
		for i, item := range items {
			if item.kind == itemSection && item.section == w.currentSection {
				return i
			}
		}
	}
	for i, item := range items {
		if item.kind == itemCategory && item.category == w.currentCategory {
			return i
		}
	}
	return 0
}

func (w *WizardOther) applyCursorPosition(item visibleItem) {
	if item.kind == itemCategory {
		w.currentCategory = item.category
		w.inCategory = false
	} else {
		w.currentCategory = item.category
		w.currentSection = item.section
		w.inCategory = true
	}
}

func (w *WizardOther) collapseCategory(cat otherCategory) {
	w.categoryExpanded[cat] = false
	for _, sec := range categorySections[int(cat)] {
		w.sectionExpanded[sec] = false
	}
	w.inCategory = false
	w.inSubSection = false
	w.subValueFocused = false
	w.simpleValueFocused = false
}
