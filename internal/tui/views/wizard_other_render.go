package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var wizOtherCategoryStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9E2AF"))

func (w WizardOther) renderContent() string {
	var lines []string

	selectedStyle := wizOtherSelectedStyle
	enabledStyle := wizOtherEnabledStyle
	labelStyle := wizOtherLabelStyle
	catStyle := wizOtherCategoryStyle

	for ci, catName := range otherCategoryNames {
		cat := otherCategory(ci)

		// Category header cursor
		catCursor := "  "
		if cat == w.currentCategory && !w.inCategory && !w.inSubSection {
			catCursor = selectedStyle.Render("> ")
		}

		catIcon := "▶"
		if w.categoryExpanded[cat] {
			catIcon = "▼"
		}

		lines = append(lines, fmt.Sprintf("%s%s %s", catCursor, catIcon, catStyle.Render(catName)))

		if !w.categoryExpanded[cat] {
			continue
		}

		// Render sections within expanded category
		sectionIndent := "   "
		for _, section := range categorySections[ci] {
			name := otherSectionNames[int(section)]

			cursor := "  "
			if w.currentSection == section && w.inCategory && !w.inSubSection {
				cursor = selectedStyle.Render("> ")
			}

			// Simple boolean / top-level value sections render inline
			if path := w.topLevelFieldPath(section); path != "" {
				checkbox := "[ ]"
				if w.fieldSelected(path) {
					checkbox = enabledStyle.Render("[✓]")
				}
				value := ""
				var valid bool
				switch section {
				case sectionAutoUpdate:
					value = onOff(w.autoUpdate)
					valid = w.fieldSelected(path)
				case sectionNewTaskSystemEnabled:
					value = onOff(w.newTaskSystemEnabled)
					valid = w.fieldSelected(path)
				case sectionHashlineEdit:
					value = onOff(w.hashlineEdit)
					valid = w.fieldSelected(path)
				case sectionModelFallback:
					value = onOff(w.modelFallback)
					valid = w.fieldSelected(path)
				case sectionStartWork:
					value = onOff(w.startWorkAutoCommit)
					valid = w.fieldSelected(path)
				}
				if value != "" {
					if valid {
						value += enabledStyle.Render(" ✓")
					}
					if w.simpleValueFocused && section == w.currentSection && w.inCategory && !w.inSubSection {
						value = labelStyle.Render(value)
					}
					line := fmt.Sprintf("%s%s%s %s: %s", sectionIndent, cursor, checkbox, labelStyle.Render(name), value)
					lines = append(lines, line)
					continue
				}
			}

			// Expandable sections
			expandIcon := "▶"
			if w.sectionExpanded[section] {
				expandIcon = "▼"
			}

			lines = append(lines, fmt.Sprintf("%s%s%s %s", sectionIndent, cursor, expandIcon, labelStyle.Render(name)))
			if w.sectionExpanded[section] {
				for _, sl := range w.renderSubSection(section) {
					lines = append(lines, sectionIndent+sl)
				}
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (w WizardOther) renderSubSection(section otherSection) []string {
	var lines []string
	indent := "      "
	selectedStyle := wizOtherSelectedStyle
	enabledStyle := wizOtherEnabledStyle
	labelStyle := wizOtherLabelStyle
	valueStyle := wizOtherDimStyle

	rowPrefix := func(idx int) string {
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == idx {
			cursor = selectedStyle.Render("> ")
		}
		return indent + cursor
	}

	renderInclude := func(idx int, path, label string) string {
		checkbox := "[ ]"
		if w.fieldSelected(path) {
			checkbox = enabledStyle.Render("[✓]")
		}
		style := valueStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == idx && !w.subValueFocused {
			style = labelStyle
		}
		return rowPrefix(idx) + checkbox + " " + style.Render(label)
	}

	renderBoolField := func(idx int, path, label string, value bool) string {
		checkbox := "[ ]"
		if w.fieldSelected(path) {
			checkbox = enabledStyle.Render("[✓]")
		}
		style := valueStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == idx && !w.subValueFocused {
			style = labelStyle
		}
		valueRender := onOff(value)
		if w.fieldSelected(path) {
			valueRender += enabledStyle.Render(" ✓")
		}
		if w.inSubSection && w.currentSection == section && w.subCursor == idx && w.subValueFocused {
			valueRender = labelStyle.Render(valueRender)
		}
		return rowPrefix(idx) + checkbox + " " + style.Render(label+": ") + valueRender
	}

	renderValueField := func(idx int, path, label, value string) string {
		checkbox := "[ ]"
		if w.fieldSelected(path) {
			checkbox = enabledStyle.Render("[✓]")
		}
		style := valueStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == idx && !w.subValueFocused {
			style = labelStyle
		}
		valueRender := value
		if w.inSubSection && w.currentSection == section && w.subCursor == idx && w.subValueFocused {
			valueRender = labelStyle.Render(value)
		}
		return rowPrefix(idx) + checkbox + " " + style.Render(label+": ") + valueRender
	}

	switch section {
	case sectionDisabledMcps:
		lines = append(lines, renderInclude(0, disabledMcpsFieldPath, "disabled_mcps"))
		lines = append(lines, renderValueField(1, disabledMcpsFieldPath, "values", w.disabledMcps.View()))
	case sectionDisabledAgents:
		lines = append(lines, renderInclude(0, disabledAgentsFieldPath, "disabled_agents"))
		for i, agent := range disableableAgents {
			lines = append(lines, renderValueField(i+1, disabledAgentsFieldPath, agent, onOff(w.disabledAgents[agent])))
		}
	case sectionDisabledSkills:
		lines = append(lines, renderInclude(0, disabledSkillsFieldPath, "disabled_skills"))
		for i, skill := range disableableSkills {
			lines = append(lines, renderValueField(i+1, disabledSkillsFieldPath, skill, onOff(w.disabledSkills[skill])))
		}
	case sectionDisabledCommands:
		lines = append(lines, renderInclude(0, disabledCommandsFieldPath, "disabled_commands"))
		for i, cmd := range disableableCommands {
			lines = append(lines, renderValueField(i+1, disabledCommandsFieldPath, cmd, onOff(w.disabledCommands[cmd])))
		}
	case sectionDisabledTools:
		lines = append(lines, renderInclude(0, disabledToolsFieldPath, "disabled_tools"))
		lines = append(lines, renderValueField(1, disabledToolsFieldPath, "values", w.disabledTools.View()))
	case sectionMcpEnvAllowlist:
		lines = append(lines, renderInclude(0, mcpEnvAllowlistFieldPath, "mcp_env_allowlist"))
		lines = append(lines, renderValueField(1, mcpEnvAllowlistFieldPath, "values", w.mcpEnvAllowlist.View()))
	case sectionExperimental:
		lines = append(lines, renderBoolField(0, "experimental.aggressive_truncation", "aggressive_truncation", w.expAggressiveTrunc))
		lines = append(lines, renderBoolField(1, "experimental.disable_live_parent_wake_routing", "disable_live_parent_wake_routing", w.expDisableLiveParentWakeRouting))
		lines = append(lines, renderBoolField(2, "experimental.truncate_all_tool_outputs", "truncate_all_tool_outputs", w.expTruncateAllOutputs))
		lines = append(lines, renderBoolField(3, "experimental.dynamic_context_pruning.enabled", "dynamic_context_pruning.enabled", w.dcpEnabled))
		lines = append(lines, renderValueField(4, "experimental.dynamic_context_pruning.notification", "dynamic_context_pruning.notification", dcpNotificationValues[w.dcpNotificationIdx]))
		lines = append(lines, renderBoolField(5, "experimental.dynamic_context_pruning.turn_protection.enabled", "dynamic_context_pruning.turn_protection.enabled", w.dcpTurnProtEnabled))
		lines = append(lines, renderValueField(6, "experimental.dynamic_context_pruning.turn_protection.turns", "dynamic_context_pruning.turn_protection.turns", w.dcpTurnProtTurns.View()))
		lines = append(lines, renderValueField(7, "experimental.dynamic_context_pruning.protected_tools", "dynamic_context_pruning.protected_tools", w.dcpProtectedTools.View()))
		lines = append(lines, renderBoolField(8, "experimental.dynamic_context_pruning.strategies.deduplication.enabled", "dynamic_context_pruning.strategies.deduplication.enabled", w.dcpDeduplicationEnabled))
		lines = append(lines, renderBoolField(9, "experimental.dynamic_context_pruning.strategies.supersede_writes.enabled", "dynamic_context_pruning.strategies.supersede_writes.enabled", w.dcpSupersedeWritesEnabled))
		lines = append(lines, renderBoolField(10, "experimental.dynamic_context_pruning.strategies.supersede_writes.aggressive", "dynamic_context_pruning.strategies.supersede_writes.aggressive", w.dcpSupersedeWritesAggressive))
		lines = append(lines, renderBoolField(11, "experimental.dynamic_context_pruning.strategies.purge_errors.enabled", "dynamic_context_pruning.strategies.purge_errors.enabled", w.dcpPurgeErrorsEnabled))
		lines = append(lines, renderValueField(12, "experimental.dynamic_context_pruning.strategies.purge_errors.turns", "dynamic_context_pruning.strategies.purge_errors.turns", w.dcpPurgeErrorsTurns.View()))
		lines = append(lines, renderBoolField(13, "experimental.preemptive_compaction", "preemptive_compaction", w.expPreemptiveCompaction))
		lines = append(lines, renderBoolField(14, "experimental.task_system", "task_system", w.expTaskSystem))
		lines = append(lines, renderValueField(15, "experimental.plugin_load_timeout_ms", "plugin_load_timeout_ms", w.expPluginLoadTimeoutMs.View()))
		lines = append(lines, renderBoolField(16, "experimental.safe_hook_creation", "safe_hook_creation", w.expSafeHookCreation))
		lines = append(lines, renderBoolField(17, "experimental.hashline_edit", "hashline_edit", w.expHashlineEdit))
		lines = append(lines, renderBoolField(18, "experimental.disable_omo_env", "disable_omo_env", w.expDisableOmoEnv))
		lines = append(lines, renderBoolField(19, "experimental.model_fallback_title", "model_fallback_title", w.expModelFallbackTitle))
		lines = append(lines, renderValueField(20, "experimental.max_tools", "max_tools", w.expMaxTools.View()))
	case sectionClaudeCode:
		lines = append(lines, renderBoolField(0, "claude_code.mcp", "mcp", w.ccMcp))
		lines = append(lines, renderBoolField(1, "claude_code.commands", "commands", w.ccCommands))
		lines = append(lines, renderBoolField(2, "claude_code.skills", "skills", w.ccSkills))
		lines = append(lines, renderBoolField(3, "claude_code.agents", "agents", w.ccAgents))
		lines = append(lines, renderBoolField(4, "claude_code.hooks", "hooks", w.ccHooks))
		lines = append(lines, renderBoolField(5, "claude_code.plugins", "plugins", w.ccPlugins))
		lines = append(lines, renderValueField(6, "claude_code.plugins_override", "plugins_override", w.ccPluginsOverride.View()))
		lines = append(lines, renderValueField(7, "claude_code.anthropic_provider", "anthropic_provider", w.ccAnthropicProvider.View()))
	case sectionSisyphusAgent:
		lines = append(lines, renderBoolField(0, "sisyphus_agent.disabled", "disabled", w.saDisabled))
		lines = append(lines, renderBoolField(1, "sisyphus_agent.default_builder_enabled", "default_builder_enabled", w.saDefaultBuilderEnabled))
		lines = append(lines, renderBoolField(2, "sisyphus_agent.planner_enabled", "planner_enabled", w.saPlannerEnabled))
		lines = append(lines, renderBoolField(3, "sisyphus_agent.replace_plan", "replace_plan", w.saReplacePlan))
		lines = append(lines, renderBoolField(4, "sisyphus_agent.tdd", "tdd", w.saTDD))
	case sectionRalphLoop:
		lines = append(lines, renderBoolField(0, "ralph_loop.enabled", "enabled", w.rlEnabled))
		lines = append(lines, renderValueField(1, "ralph_loop.default_max_iterations", "default_max_iterations", w.rlDefaultMaxIterations.View()))
		lines = append(lines, renderValueField(2, "ralph_loop.state_dir", "state_dir", w.rlStateDir.View()))
		lines = append(lines, renderValueField(3, "ralph_loop.default_strategy", "default_strategy", ralphLoopStrategies[w.rlDefaultStrategyIdx]))
	case sectionBackgroundTask:
		lines = append(lines, renderValueField(0, "background_task.provider_concurrency", "provider_concurrency", w.btProviderConcurrency.View()))
		lines = append(lines, renderValueField(1, "background_task.model_concurrency", "model_concurrency", w.btModelConcurrency.View()))
		lines = append(lines, renderValueField(2, "background_task.default_concurrency", "default_concurrency", w.btDefaultConcurrency.View()))
		lines = append(lines, renderValueField(3, "background_task.stale_timeout_ms", "stale_timeout_ms", w.btStaleTimeoutMs.View()))
		lines = append(lines, renderValueField(4, "background_task.message_staleness_timeout_ms", "message_staleness_timeout_ms", w.btMessageStalenessTimeoutMs.View()))
		lines = append(lines, renderValueField(5, "background_task.sync_poll_timeout_ms", "sync_poll_timeout_ms", w.btSyncPollTimeoutMs.View()))
		lines = append(lines, renderValueField(6, "background_task.max_depth", "max_depth", w.btMaxDepth.View()))
		lines = append(lines, renderValueField(7, "background_task.task_ttl_ms", "task_ttl_ms", w.btTaskTtlMs.View()))
		lines = append(lines, renderValueField(8, "background_task.session_gone_timeout_ms", "session_gone_timeout_ms", w.btSessionGoneTimeoutMs.View()))
		lines = append(lines, renderValueField(9, "background_task.max_tool_calls", "max_tool_calls", w.btMaxToolCalls.View()))
		lines = append(lines, renderBoolField(10, "background_task.circuit_breaker.enabled", "circuit_breaker.enabled", w.btCircuitBreakerEnabled))
		lines = append(lines, renderValueField(11, "background_task.circuit_breaker.max_tool_calls", "circuit_breaker.max_tool_calls", w.btCircuitBreakerMaxCalls.View()))
		lines = append(lines, renderValueField(12, "background_task.circuit_breaker.consecutive_threshold", "circuit_breaker.consecutive_threshold", w.btCircuitBreakerConsecutive.View()))
	case sectionNotification:
		lines = append(lines, renderBoolField(0, "notification.force_enable", "force_enable", w.notifForceEnable))
	case sectionGitMaster:
		lines = append(lines, renderBoolField(0, "git_master.commit_footer", "commit_footer", w.gmCommitFooter))
		lines = append(lines, renderValueField(1, "git_master.commit_footer", "commit_footer_text", w.gmCommitFooterText.View()))
		lines = append(lines, renderBoolField(2, "git_master.include_co_authored_by", "include_co_authored_by", w.gmIncludeCoAuthoredBy))
		lines = append(lines, renderValueField(3, "git_master.git_env_prefix", "git_env_prefix", w.gmGitEnvPrefix.View()))
	case sectionCommentChecker:
		lines = append(lines, renderValueField(0, commentCheckerCustomPromptFieldPath, "custom_prompt", w.ccCustomPrompt.View()))
	case sectionBabysitting:
		lines = append(lines, renderValueField(0, babysittingTimeoutFieldPath, "timeout_ms", w.babysittingTimeoutMs.View()))
	case sectionBrowserAutomationEngine:
		lines = append(lines, renderValueField(0, browserProviderFieldPath, "provider", browserProviders[w.browserProviderIdx]))
	case sectionTmux:
		lines = append(lines, renderBoolField(0, "tmux.enabled", "enabled", w.tmuxEnabled))
		lines = append(lines, renderValueField(1, "tmux.layout", "layout", tmuxLayouts[w.tmuxLayoutIdx]))
		lines = append(lines, renderValueField(2, "tmux.main_pane_size", "main_pane_size", w.tmuxMainPaneSize.View()))
		lines = append(lines, renderValueField(3, "tmux.main_pane_min_width", "main_pane_min_width", w.tmuxMainPaneMinWidth.View()))
		lines = append(lines, renderValueField(4, "tmux.agent_pane_min_width", "agent_pane_min_width", w.tmuxAgentPaneMinWidth.View()))
		lines = append(lines, renderValueField(5, "tmux.isolation", "isolation", tmuxIsolations[w.tmuxIsolationIdx]))
	case sectionWebsearch:
		lines = append(lines, renderValueField(0, websearchProviderFieldPath, "provider", websearchProviders[w.websearchProviderIdx]))
	case sectionSisyphus:
		lines = append(lines, renderValueField(0, "sisyphus.tasks.storage_path", "tasks.storage_path", w.sisyphusTasksStoragePath.View()))
		lines = append(lines, renderValueField(1, "sisyphus.tasks.task_list_id", "tasks.task_list_id", w.sisyphusTasksTaskListID.View()))
		lines = append(lines, renderBoolField(2, "sisyphus.tasks.claude_code_compat", "tasks.claude_code_compat", w.sisyphusTasksClaudeCodeCompat))
	case sectionDefaultRunAgent:
		lines = append(lines, renderValueField(0, defaultRunAgentFieldPath, "value", w.defaultRunAgent.View()))
	case sectionModelCapabilities:
		lines = append(lines, renderBoolField(0, "model_capabilities.enabled", "enabled", w.mcEnabled))
		lines = append(lines, renderBoolField(1, "model_capabilities.auto_refresh_on_start", "auto_refresh_on_start", w.mcAutoRefreshOnStart))
		lines = append(lines, renderValueField(2, "model_capabilities.refresh_timeout_ms", "refresh_timeout_ms", w.mcRefreshTimeoutMs.View()))
		lines = append(lines, renderValueField(3, "model_capabilities.source_url", "source_url", w.mcSourceURL.View()))
	case sectionOpenclaw:
		lines = append(lines, renderBoolField(0, openclawEnabledFieldPath, "enabled", strings.Contains(w.openclawEditor.Value(), `"enabled": true`)))
		lines = append(lines, renderInclude(1, openclawGatewaysFieldPath, "gateways"))
		lines = append(lines, renderInclude(2, openclawHooksFieldPath, "hooks"))
		lines = append(lines, renderInclude(3, openclawReplyListenerFieldPath, "reply_listener"))
		lines = append(lines, rowPrefix(4)+w.openclawEditor.View())
	case sectionRuntimeFallback:
		lines = append(lines, renderInclude(0, runtimeFallbackFieldPath, "runtime_fallback"))
		lines = append(lines, rowPrefix(1)+w.runtimeFallbackEditor.View())
	case sectionSkillsJson:
		lines = append(lines, renderInclude(0, skillsFieldPath, "skills"))
		lines = append(lines, rowPrefix(1)+w.skillsEditor.View())
	case sectionAgentOrder:
		lines = append(lines, renderInclude(0, agentOrderFieldPath, "agent_order"))
		lines = append(lines, renderValueField(1, agentOrderFieldPath, "values", w.agentOrder.View()))
	case sectionKeywordDetector:
		lines = append(lines, renderInclude(0, keywordDetectorFieldPath, "disabled_keywords"))
		for i, kw := range disableableKeywords {
			lines = append(lines, renderValueField(i+1, keywordDetectorFieldPath, kw, onOff(w.disabledKeywords[kw])))
		}
	case sectionTeamMode:
		lines = append(lines, renderBoolField(0, "team_mode.enabled", "enabled", w.tmEnabled))
		lines = append(lines, renderBoolField(1, "team_mode.tmux_visualization", "tmux_visualization", w.tmTmuxVisualization))
		lines = append(lines, renderValueField(2, "team_mode.max_parallel_members", "max_parallel_members", w.tmMaxParallelMembers.View()))
		lines = append(lines, renderValueField(3, "team_mode.max_members", "max_members", w.tmMaxMembers.View()))
		lines = append(lines, renderValueField(4, "team_mode.max_messages_per_run", "max_messages_per_run", w.tmMaxMessagesPerRun.View()))
		lines = append(lines, renderValueField(5, "team_mode.max_wall_clock_minutes", "max_wall_clock_minutes", w.tmMaxWallClockMinutes.View()))
		lines = append(lines, renderValueField(6, "team_mode.max_member_turns", "max_member_turns", w.tmMaxMemberTurns.View()))
		lines = append(lines, renderValueField(7, "team_mode.base_dir", "base_dir", w.tmBaseDir.View()))
		lines = append(lines, renderValueField(8, "team_mode.message_payload_max_bytes", "message_payload_max_bytes", w.tmMessagePayloadMaxBytes.View()))
		lines = append(lines, renderValueField(9, "team_mode.recipient_unread_max_bytes", "recipient_unread_max_bytes", w.tmRecipientUnreadMaxBytes.View()))
		lines = append(lines, renderValueField(10, "team_mode.mailbox_poll_interval_ms", "mailbox_poll_interval_ms", w.tmMailboxPollIntervalMs.View()))
	case sectionMonitor:
		lines = append(lines, renderBoolField(0, "monitor.enabled", "enabled", w.monEnabled))
		lines = append(lines, renderBoolField(1, "monitor.live_mode_enabled", "live_mode_enabled", w.monLiveModeEnabled))
		lines = append(lines, renderValueField(2, "monitor.allowed_commands", "allowed_commands", w.monAllowedCommands.View()))
		lines = append(lines, renderValueField(3, "monitor.max_monitors_per_session", "max_monitors_per_session", w.monMaxMonitors.View()))
		lines = append(lines, renderValueField(4, "monitor.max_runtime_ms", "max_runtime_ms", w.monMaxRuntimeMs.View()))
		lines = append(lines, renderValueField(5, "monitor.batch_max_lines", "batch_max_lines", w.monBatchMaxLines.View()))
		lines = append(lines, renderValueField(6, "monitor.batch_max_bytes", "batch_max_bytes", w.monBatchMaxBytes.View()))
		lines = append(lines, renderValueField(7, "monitor.flush_interval_ms", "flush_interval_ms", w.monFlushIntervalMs.View()))
		lines = append(lines, renderValueField(8, "monitor.ring_max_lines", "ring_max_lines", w.monRingMaxLines.View()))
		lines = append(lines, renderValueField(9, "monitor.line_max_bytes", "line_max_bytes", w.monLineMaxBytes.View()))
		lines = append(lines, renderValueField(10, "monitor.pattern_max_length", "pattern_max_length", w.monPatternMaxLength.View()))
	case sectionCodegraph:
		lines = append(lines, renderBoolField(0, "codegraph.enabled", "enabled", w.cgEnabled))
		lines = append(lines, renderBoolField(1, "codegraph.auto_provision", "auto_provision", w.cgAutoProvision))
		lines = append(lines, renderValueField(2, "codegraph.install_dir", "install_dir", w.cgInstallDir.View()))
		lines = append(lines, renderBoolField(3, "codegraph.telemetry", "telemetry", w.cgTelemetry))
		lines = append(lines, renderValueField(4, "codegraph.watch_debounce_ms", "watch_debounce_ms", w.cgWatchDebounceMs.View()))
	case sectionTui:
		lines = append(lines, renderBoolField(0, "tui.sidebar.enabled", "sidebar.enabled", w.tuiSidebarEnabled))
	}

	lines = append(lines, "")
	return lines
}
