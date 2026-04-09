package views

import (
	"fmt"
	"strconv"
	"strings"
)

// parseMapStringInt parses "key:val, key2:val2" into map[string]int
func parseMapStringInt(s string) map[string]int {
	if s == "" {
		return nil
	}
	result := make(map[string]int)
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
		if i, err := strconv.Atoi(val); err == nil {
			result[key] = i
		}
		// If Atoi fails, skip this entry
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// serializeMapStringInt converts map[string]int to "key:val, key2:val2"
func serializeMapStringInt(m map[string]int) string {
	if len(m) == 0 {
		return ""
	}
	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s:%d", k, v))
	}
	return strings.Join(pairs, ", ")
}

const (
	disabledMcpsFieldPath               = "disabled_mcps"
	disabledAgentsFieldPath             = "disabled_agents"
	disabledSkillsFieldPath             = "disabled_skills"
	disabledCommandsFieldPath           = "disabled_commands"
	disabledToolsFieldPath              = "disabled_tools"
	autoUpdateFieldPath                 = "auto_update"
	hashlineEditFieldPath               = "hashline_edit"
	modelFallbackFieldPath              = "model_fallback"
	newTaskSystemEnabledFieldPath       = "new_task_system_enabled"
	defaultRunAgentFieldPath            = "default_run_agent"
	runtimeFallbackFieldPath            = "runtime_fallback"
	skillsFieldPath                     = "skills"
	startWorkAutoCommitFieldPath        = "start_work.auto_commit"
	browserProviderFieldPath            = "browser_automation_engine.provider"
	websearchProviderFieldPath          = "websearch.provider"
	commentCheckerCustomPromptFieldPath = "comment_checker.custom_prompt"
	babysittingTimeoutFieldPath         = "babysitting.timeout_ms"
	mcpEnvAllowlistFieldPath            = "mcp_env_allowlist"
	openclawEnabledFieldPath            = "openclaw.enabled"
	openclawGatewaysFieldPath           = "openclaw.gateways.*.type"
	openclawHooksFieldPath              = "openclaw.hooks.*.enabled"
	openclawReplyListenerFieldPath      = "openclaw.reply_listener.discord_bot_token"
)

func parsePositiveInt64(input string) *int64 {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil || value <= 0 {
		return nil
	}
	return &value
}

func parseNonNegativeInt(input string) *int {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || value < 0 {
		return nil
	}
	return &value
}

func parsePositiveIntWithMinimum(input string, minimum int) *int {
	value := parseNonNegativeInt(input)
	if value == nil || *value < minimum {
		return nil
	}
	return value
}

func wizardOtherBoolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func emptySliceIfSelected(selected bool, values []string) []string {
	if !selected {
		return nil
	}
	if values == nil {
		return []string{}
	}
	return values
}

func emptyMapStringBoolIfSelected(selected bool, values map[string]bool) map[string]bool {
	if !selected {
		return nil
	}
	if values == nil {
		return map[string]bool{}
	}
	return values
}

func (w WizardOther) fieldSelected(path string) bool {
	if w.selection == nil {
		return true
	}
	return w.selection.IsSelected(path)
}

func (w *WizardOther) toggleFieldSelection(path string) {
	if w.selection == nil {
		return
	}
	w.selection.SetSelected(path, !w.selection.IsSelected(path))
}

func (w WizardOther) selectedWithPrefix(prefix string) bool {
	if w.selection == nil {
		return false
	}
	for _, path := range w.selection.SelectedPaths() {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (w WizardOther) expHasData() bool {
	if w.selection == nil {
		return w.expAggressiveTrunc || w.expAutoResume ||
			w.expTruncateAllOutputs || w.expPreemptiveCompaction || w.expTaskSystem ||
			w.expPluginLoadTimeoutMs.Value() != "" || w.expSafeHookCreation || w.expHashlineEdit ||
			w.expDisableOmoEnv || w.expModelFallbackTitle || w.expMaxTools.Value() != "" ||
			w.dcpEnabled || w.dcpNotificationIdx > 0 || w.dcpTurnProtEnabled ||
			w.dcpTurnProtTurns.Value() != "" || w.dcpProtectedTools.Value() != "" ||
			w.dcpDeduplicationEnabled || w.dcpSupersedeWritesEnabled ||
			w.dcpSupersedeWritesAggressive || w.dcpPurgeErrorsEnabled ||
			w.dcpPurgeErrorsTurns.Value() != ""
	}
	return w.selectedWithPrefix("experimental.")
}

func (w WizardOther) ccHasData() bool {
	if w.selection == nil {
		return w.ccMcp || w.ccCommands || w.ccSkills || w.ccAgents || w.ccHooks || w.ccPlugins || w.ccPluginsOverride.Value() != ""
	}
	return w.selectedWithPrefix("claude_code.")
}

func (w WizardOther) saHasData() bool {
	if w.selection == nil {
		return w.saDisabled || w.saDefaultBuilderEnabled || w.saPlannerEnabled || w.saReplacePlan || w.saTDD
	}
	return w.selectedWithPrefix("sisyphus_agent.")
}

func (w WizardOther) rlHasData() bool {
	if w.selection == nil {
		return w.rlEnabled || w.rlDefaultMaxIterations.Value() != "" || w.rlStateDir.Value() != "" || w.rlDefaultStrategyIdx > 0
	}
	return w.selectedWithPrefix("ralph_loop.")
}

func (w WizardOther) btHasData() bool {
	if w.selection == nil {
		return w.btDefaultConcurrency.Value() != "" || w.btProviderConcurrency.Value() != "" || w.btModelConcurrency.Value() != "" ||
			w.btMaxDepth.Value() != "" || w.btMaxDescendants.Value() != "" || w.btStaleTimeoutMs.Value() != "" ||
			w.btMessageStalenessTimeoutMs.Value() != "" || w.btTaskTtlMs.Value() != "" || w.btSessionGoneTimeoutMs.Value() != "" ||
			w.btSyncPollTimeoutMs.Value() != "" || w.btMaxToolCalls.Value() != "" || w.btCircuitBreakerEnabled ||
			w.btCircuitBreakerMaxCalls.Value() != "" || w.btCircuitBreakerConsecutive.Value() != ""
	}
	return w.selectedWithPrefix("background_task.")
}

func (w WizardOther) tmuxHasData() bool {
	if w.selection == nil {
		return w.tmuxEnabled || w.tmuxLayoutIdx > 0 || w.tmuxMainPaneSize.Value() != "" || w.tmuxMainPaneMinWidth.Value() != "" ||
			w.tmuxAgentPaneMinWidth.Value() != "" || w.tmuxIsolationIdx > 0
	}
	return w.selectedWithPrefix("tmux.")
}

func (w WizardOther) sisyphusHasData() bool {
	if w.selection == nil {
		return w.sisyphusTasksStoragePath.Value() != "" || w.sisyphusTasksTaskListID.Value() != "" || w.sisyphusTasksClaudeCodeCompat
	}
	return w.selectedWithPrefix("sisyphus.tasks.")
}

func (w WizardOther) modelCapabilitiesHasData() bool {
	if w.selection == nil {
		return w.mcEnabled || w.mcAutoRefreshOnStart || strings.TrimSpace(w.mcRefreshTimeoutMs.Value()) != "" || strings.TrimSpace(w.mcSourceURL.Value()) != ""
	}
	return w.selectedWithPrefix("model_capabilities.")
}

func (w WizardOther) openclawHasData() bool {
	if w.selection == nil {
		return strings.TrimSpace(w.openclawEditor.Value()) != ""
	}
	return w.selectedWithPrefix("openclaw.")
}

func (w WizardOther) topLevelFieldPath(section otherSection) string {
	switch section {
	case sectionDisabledMcps:
		return disabledMcpsFieldPath
	case sectionDisabledAgents:
		return disabledAgentsFieldPath
	case sectionDisabledSkills:
		return disabledSkillsFieldPath
	case sectionDisabledCommands:
		return disabledCommandsFieldPath
	case sectionDisabledTools:
		return disabledToolsFieldPath
	case sectionAutoUpdate:
		return autoUpdateFieldPath
	case sectionNewTaskSystemEnabled:
		return newTaskSystemEnabledFieldPath
	case sectionHashlineEdit:
		return hashlineEditFieldPath
	case sectionModelFallback:
		return modelFallbackFieldPath
	case sectionDefaultRunAgent:
		return defaultRunAgentFieldPath
	case sectionMcpEnvAllowlist:
		return mcpEnvAllowlistFieldPath
	case sectionRuntimeFallback:
		return runtimeFallbackFieldPath
	case sectionSkillsJson:
		return skillsFieldPath
	case sectionStartWork:
		return startWorkAutoCommitFieldPath
	default:
		return ""
	}
}

func (w WizardOther) isSimpleBooleanSection(section otherSection) bool {
	switch section {
	case sectionAutoUpdate, sectionNewTaskSystemEnabled, sectionHashlineEdit, sectionModelFallback, sectionStartWork:
		return true
	default:
		return false
	}
}

func (w WizardOther) subSectionFieldPath(section otherSection, idx int) string {
	switch section {
	case sectionExperimental:
		paths := []string{
			"experimental.aggressive_truncation",
			"experimental.auto_resume",
			"experimental.truncate_all_tool_outputs",
			"experimental.dynamic_context_pruning.enabled",
			"experimental.dynamic_context_pruning.notification",
			"experimental.dynamic_context_pruning.turn_protection.enabled",
			"experimental.dynamic_context_pruning.turn_protection.turns",
			"experimental.dynamic_context_pruning.protected_tools",
			"experimental.dynamic_context_pruning.strategies.deduplication.enabled",
			"experimental.dynamic_context_pruning.strategies.supersede_writes.enabled",
			"experimental.dynamic_context_pruning.strategies.supersede_writes.aggressive",
			"experimental.dynamic_context_pruning.strategies.purge_errors.enabled",
			"experimental.dynamic_context_pruning.strategies.purge_errors.turns",
			"experimental.preemptive_compaction",
			"experimental.task_system",
			"experimental.plugin_load_timeout_ms",
			"experimental.safe_hook_creation",
			"experimental.hashline_edit",
			"experimental.disable_omo_env",
			"experimental.model_fallback_title",
			"experimental.max_tools",
		}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionClaudeCode:
		paths := []string{"claude_code.mcp", "claude_code.commands", "claude_code.skills", "claude_code.agents", "claude_code.hooks", "claude_code.plugins", "claude_code.plugins_override"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionSisyphusAgent:
		paths := []string{"sisyphus_agent.disabled", "sisyphus_agent.default_builder_enabled", "sisyphus_agent.planner_enabled", "sisyphus_agent.replace_plan", "sisyphus_agent.tdd"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionRalphLoop:
		paths := []string{"ralph_loop.enabled", "ralph_loop.default_max_iterations", "ralph_loop.state_dir", "ralph_loop.default_strategy"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionBackgroundTask:
		paths := []string{"background_task.provider_concurrency", "background_task.model_concurrency", "background_task.default_concurrency", "background_task.stale_timeout_ms", "background_task.message_staleness_timeout_ms", "background_task.sync_poll_timeout_ms", "background_task.max_depth", "background_task.max_descendants", "background_task.task_ttl_ms", "background_task.session_gone_timeout_ms", "background_task.max_tool_calls", "background_task.circuit_breaker.enabled", "background_task.circuit_breaker.max_tool_calls", "background_task.circuit_breaker.consecutive_threshold"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionNotification:
		if idx == 0 {
			return "notification.force_enable"
		}
	case sectionGitMaster:
		paths := []string{"git_master.commit_footer", "git_master.commit_footer", "git_master.include_co_authored_by", "git_master.git_env_prefix"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionCommentChecker:
		if idx == 0 {
			return commentCheckerCustomPromptFieldPath
		}
	case sectionBabysitting:
		if idx == 0 {
			return babysittingTimeoutFieldPath
		}
	case sectionBrowserAutomationEngine:
		if idx == 0 {
			return browserProviderFieldPath
		}
	case sectionTmux:
		paths := []string{"tmux.enabled", "tmux.layout", "tmux.main_pane_size", "tmux.main_pane_min_width", "tmux.agent_pane_min_width", "tmux.isolation"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionWebsearch:
		if idx == 0 {
			return websearchProviderFieldPath
		}
	case sectionSisyphus:
		paths := []string{"sisyphus.tasks.storage_path", "sisyphus.tasks.task_list_id", "sisyphus.tasks.claude_code_compat"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionDefaultRunAgent:
		if idx == 0 {
			return defaultRunAgentFieldPath
		}
	case sectionModelCapabilities:
		paths := []string{"model_capabilities.enabled", "model_capabilities.auto_refresh_on_start", "model_capabilities.refresh_timeout_ms", "model_capabilities.source_url"}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	case sectionOpenclaw:
		paths := []string{openclawEnabledFieldPath, openclawGatewaysFieldPath, openclawHooksFieldPath, openclawReplyListenerFieldPath}
		if idx >= 0 && idx < len(paths) {
			return paths[idx]
		}
	}
	return ""
}

func onOff(v bool) string {
	if v {
		return "[on]"
	}
	return "[off]"
}
