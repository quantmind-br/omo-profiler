package views

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
)

func (w *WizardOther) SetConfig(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection
	// Disabled agents
	for _, a := range cfg.DisabledAgents {
		if _, ok := w.disabledAgents[a]; ok {
			w.disabledAgents[a] = true
		}
	}

	// Disabled skills
	for _, s := range cfg.DisabledSkills {
		if _, ok := w.disabledSkills[s]; ok {
			w.disabledSkills[s] = true
		}
	}

	// Disabled commands
	for _, c := range cfg.DisabledCommands {
		if _, ok := w.disabledCommands[c]; ok {
			w.disabledCommands[c] = true
		}
	}

	// Disabled MCPs
	if len(cfg.DisabledMCPs) > 0 {
		w.disabledMcps.SetValue(strings.Join(cfg.DisabledMCPs, ", "))
	}

	// Disabled Tools
	if len(cfg.DisabledTools) > 0 {
		w.disabledTools.SetValue(strings.Join(cfg.DisabledTools, ", "))
	}

	// MCP Env Allowlist
	if len(cfg.MCPEnvAllowlist) > 0 {
		w.mcpEnvAllowlist.SetValue(strings.Join(cfg.MCPEnvAllowlist, ", "))
	}

	// Auto update
	if cfg.AutoUpdate != nil {
		w.autoUpdate = *cfg.AutoUpdate
	}

	if cfg.HashlineEdit != nil {
		w.hashlineEdit = *cfg.HashlineEdit
	}

	if cfg.ModelFallback != nil {
		w.modelFallback = *cfg.ModelFallback
	}

	if cfg.StartWork != nil && cfg.StartWork.AutoCommit != nil {
		w.startWorkAutoCommit = *cfg.StartWork.AutoCommit
	}

	// Experimental
	if cfg.Experimental != nil {
		if cfg.Experimental.AggressiveTruncation != nil {
			w.expAggressiveTrunc = *cfg.Experimental.AggressiveTruncation
		}
		if cfg.Experimental.AutoResume != nil {
			w.expAutoResume = *cfg.Experimental.AutoResume
		}
		if cfg.Experimental.TruncateAllToolOutputs != nil {
			w.expTruncateAllOutputs = *cfg.Experimental.TruncateAllToolOutputs
		}
		if cfg.Experimental.PreemptiveCompaction != nil {
			w.expPreemptiveCompaction = *cfg.Experimental.PreemptiveCompaction
		}
		if cfg.Experimental.TaskSystem != nil {
			w.expTaskSystem = *cfg.Experimental.TaskSystem
		}
		if cfg.Experimental.PluginLoadTimeoutMs != nil {
			w.expPluginLoadTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.Experimental.PluginLoadTimeoutMs))
		}
		if cfg.Experimental.SafeHookCreation != nil {
			w.expSafeHookCreation = *cfg.Experimental.SafeHookCreation
		}
		if cfg.Experimental.HashlineEdit != nil {
			w.expHashlineEdit = *cfg.Experimental.HashlineEdit
		}
		if cfg.Experimental.DisableOmoEnv != nil {
			w.expDisableOmoEnv = *cfg.Experimental.DisableOmoEnv
		}
		if cfg.Experimental.ModelFallbackTitle != nil {
			w.expModelFallbackTitle = *cfg.Experimental.ModelFallbackTitle
		}
		if cfg.Experimental.MaxTools != nil {
			w.expMaxTools.SetValue(fmt.Sprintf("%d", *cfg.Experimental.MaxTools))
		}

		if cfg.Experimental.DynamicContextPruning != nil {
			dcp := cfg.Experimental.DynamicContextPruning
			if dcp.Enabled != nil {
				w.dcpEnabled = *dcp.Enabled
			}
			if dcp.Notification != "" {
				for i, v := range dcpNotificationValues {
					if v == dcp.Notification {
						w.dcpNotificationIdx = i
						break
					}
				}
			}
			if dcp.TurnProtection != nil {
				if dcp.TurnProtection.Enabled != nil {
					w.dcpTurnProtEnabled = *dcp.TurnProtection.Enabled
				}
				if dcp.TurnProtection.Turns != nil {
					w.dcpTurnProtTurns.SetValue(fmt.Sprintf("%d", *dcp.TurnProtection.Turns))
				}
			}
			if len(dcp.ProtectedTools) > 0 {
				w.dcpProtectedTools.SetValue(strings.Join(dcp.ProtectedTools, ", "))
			}
			if dcp.Strategies != nil {
				if dcp.Strategies.Deduplication != nil && dcp.Strategies.Deduplication.Enabled != nil {
					w.dcpDeduplicationEnabled = *dcp.Strategies.Deduplication.Enabled
				}
				if dcp.Strategies.SupersedeWrites != nil {
					if dcp.Strategies.SupersedeWrites.Enabled != nil {
						w.dcpSupersedeWritesEnabled = *dcp.Strategies.SupersedeWrites.Enabled
					}
					if dcp.Strategies.SupersedeWrites.Aggressive != nil {
						w.dcpSupersedeWritesAggressive = *dcp.Strategies.SupersedeWrites.Aggressive
					}
				}
				if dcp.Strategies.PurgeErrors != nil {
					if dcp.Strategies.PurgeErrors.Enabled != nil {
						w.dcpPurgeErrorsEnabled = *dcp.Strategies.PurgeErrors.Enabled
					}
					if dcp.Strategies.PurgeErrors.Turns != nil {
						w.dcpPurgeErrorsTurns.SetValue(fmt.Sprintf("%d", *dcp.Strategies.PurgeErrors.Turns))
					}
				}
			}
		}
	}

	// Claude Code
	if cfg.ClaudeCode != nil {
		if cfg.ClaudeCode.MCP != nil {
			w.ccMcp = *cfg.ClaudeCode.MCP
		}
		if cfg.ClaudeCode.Commands != nil {
			w.ccCommands = *cfg.ClaudeCode.Commands
		}
		if cfg.ClaudeCode.Skills != nil {
			w.ccSkills = *cfg.ClaudeCode.Skills
		}
		if cfg.ClaudeCode.Agents != nil {
			w.ccAgents = *cfg.ClaudeCode.Agents
		}
		if cfg.ClaudeCode.Hooks != nil {
			w.ccHooks = *cfg.ClaudeCode.Hooks
		}
		if cfg.ClaudeCode.Plugins != nil {
			w.ccPlugins = *cfg.ClaudeCode.Plugins
		}
		if len(cfg.ClaudeCode.PluginsOverride) > 0 {
			w.ccPluginsOverride.SetValue(serializeMapStringBool(cfg.ClaudeCode.PluginsOverride))
		}
	}

	// Sisyphus Agent
	if cfg.SisyphusAgent != nil {
		if cfg.SisyphusAgent.Disabled != nil {
			w.saDisabled = *cfg.SisyphusAgent.Disabled
		}
		if cfg.SisyphusAgent.DefaultBuilderEnabled != nil {
			w.saDefaultBuilderEnabled = *cfg.SisyphusAgent.DefaultBuilderEnabled
		}
		if cfg.SisyphusAgent.PlannerEnabled != nil {
			w.saPlannerEnabled = *cfg.SisyphusAgent.PlannerEnabled
		}
		if cfg.SisyphusAgent.ReplacePlan != nil {
			w.saReplacePlan = *cfg.SisyphusAgent.ReplacePlan
		}
		if cfg.SisyphusAgent.TDD != nil {
			w.saTDD = *cfg.SisyphusAgent.TDD
		}
	}

	// Ralph Loop
	if cfg.RalphLoop != nil {
		if cfg.RalphLoop.Enabled != nil {
			w.rlEnabled = *cfg.RalphLoop.Enabled
		}
		if cfg.RalphLoop.DefaultMaxIterations != nil {
			w.rlDefaultMaxIterations.SetValue(fmt.Sprintf("%d", *cfg.RalphLoop.DefaultMaxIterations))
		}
		if cfg.RalphLoop.StateDir != "" {
			w.rlStateDir.SetValue(cfg.RalphLoop.StateDir)
		}
		if cfg.RalphLoop.DefaultStrategy != "" {
			for i, s := range ralphLoopStrategies {
				if s == cfg.RalphLoop.DefaultStrategy {
					w.rlDefaultStrategyIdx = i
					break
				}
			}
		}
	}

	// Background Task
	if cfg.BackgroundTask != nil {
		if cfg.BackgroundTask.DefaultConcurrency != nil {
			w.btDefaultConcurrency.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.DefaultConcurrency))
		}
		if len(cfg.BackgroundTask.ProviderConcurrency) > 0 {
			w.btProviderConcurrency.SetValue(serializeMapStringInt(cfg.BackgroundTask.ProviderConcurrency))
		}
		if cfg.BackgroundTask.ModelConcurrency != nil {
			w.btModelConcurrency.SetValue(serializeMapStringInt(cfg.BackgroundTask.ModelConcurrency))
		}
		if cfg.BackgroundTask.MaxDepth != nil {
			w.btMaxDepth.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.MaxDepth))
		}
		if cfg.BackgroundTask.StaleTimeoutMs != nil {
			w.btStaleTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.StaleTimeoutMs))
		}
		if cfg.BackgroundTask.MessageStalenessTimeoutMs != nil {
			w.btMessageStalenessTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.MessageStalenessTimeoutMs))
		}
		if cfg.BackgroundTask.TaskTtlMs != nil {
			w.btTaskTtlMs.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.TaskTtlMs))
		}
		if cfg.BackgroundTask.SessionGoneTimeoutMs != nil {
			w.btSessionGoneTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.SessionGoneTimeoutMs))
		}
		if cfg.BackgroundTask.SyncPollTimeoutMs != nil {
			w.btSyncPollTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.SyncPollTimeoutMs))
		}
		if cfg.BackgroundTask.MaxToolCalls != nil {
			w.btMaxToolCalls.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.MaxToolCalls))
		}
		if cfg.BackgroundTask.CircuitBreaker != nil {
			if cfg.BackgroundTask.CircuitBreaker.Enabled != nil {
				w.btCircuitBreakerEnabled = *cfg.BackgroundTask.CircuitBreaker.Enabled
			}
			if cfg.BackgroundTask.CircuitBreaker.MaxToolCalls != nil {
				w.btCircuitBreakerMaxCalls.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.CircuitBreaker.MaxToolCalls))
			}
			if cfg.BackgroundTask.CircuitBreaker.ConsecutiveThreshold != nil {
				w.btCircuitBreakerConsecutive.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.CircuitBreaker.ConsecutiveThreshold))
			}
		}
	}

	// Notification
	if cfg.Notification != nil {
		if cfg.Notification.ForceEnable != nil {
			w.notifForceEnable = *cfg.Notification.ForceEnable
		}
	}

	// Git Master
	if cfg.GitMaster != nil {
		if cfg.GitMaster.CommitFooter != nil {
			switch v := cfg.GitMaster.CommitFooter.(type) {
			case bool:
				w.gmCommitFooter = v
			case string:
				w.gmCommitFooterText.SetValue(v)
			}
		}
		if cfg.GitMaster.IncludeCoAuthoredBy != nil {
			w.gmIncludeCoAuthoredBy = *cfg.GitMaster.IncludeCoAuthoredBy
		}
		if cfg.GitMaster.GitEnvPrefix != "" {
			w.gmGitEnvPrefix.SetValue(cfg.GitMaster.GitEnvPrefix)
		}
	}

	// Comment Checker
	if cfg.CommentChecker != nil {
		if cfg.CommentChecker.CustomPrompt != "" {
			w.ccCustomPrompt.SetValue(cfg.CommentChecker.CustomPrompt)
		}
	}

	// Babysitting
	if cfg.Babysitting != nil && cfg.Babysitting.TimeoutMs != nil {
		w.babysittingTimeoutMs.SetValue(fmt.Sprintf("%g", *cfg.Babysitting.TimeoutMs))
	}

	// Browser Automation Engine
	if cfg.BrowserAutomationEngine != nil {
		for i, v := range browserProviders {
			if v == cfg.BrowserAutomationEngine.Provider {
				w.browserProviderIdx = i
				break
			}
		}
	}

	// Tmux
	if cfg.Tmux != nil {
		if cfg.Tmux.Enabled != nil {
			w.tmuxEnabled = *cfg.Tmux.Enabled
		}
		if cfg.Tmux.Layout != "" {
			for i, v := range tmuxLayouts {
				if v == cfg.Tmux.Layout {
					w.tmuxLayoutIdx = i
					break
				}
			}
		}
		if cfg.Tmux.MainPaneSize != nil {
			w.tmuxMainPaneSize.SetValue(fmt.Sprintf("%g", *cfg.Tmux.MainPaneSize))
		}
		if cfg.Tmux.MainPaneMinWidth != nil {
			w.tmuxMainPaneMinWidth.SetValue(fmt.Sprintf("%g", *cfg.Tmux.MainPaneMinWidth))
		}
		if cfg.Tmux.AgentPaneMinWidth != nil {
			w.tmuxAgentPaneMinWidth.SetValue(fmt.Sprintf("%g", *cfg.Tmux.AgentPaneMinWidth))
		}
		if cfg.Tmux.Isolation != "" {
			for i, v := range tmuxIsolations {
				if v == cfg.Tmux.Isolation {
					w.tmuxIsolationIdx = i
					break
				}
			}
		}
	}

	// Websearch
	if cfg.Websearch != nil {
		for i, v := range websearchProviders {
			if v == cfg.Websearch.Provider {
				w.websearchProviderIdx = i
				break
			}
		}
	}

	// Sisyphus
	if cfg.Sisyphus != nil && cfg.Sisyphus.Tasks != nil {
		if cfg.Sisyphus.Tasks.StoragePath != "" {
			w.sisyphusTasksStoragePath.SetValue(cfg.Sisyphus.Tasks.StoragePath)
		}
		if cfg.Sisyphus.Tasks.TaskListID != "" {
			w.sisyphusTasksTaskListID.SetValue(cfg.Sisyphus.Tasks.TaskListID)
		}
		if cfg.Sisyphus.Tasks.ClaudeCodeCompat != nil {
			w.sisyphusTasksClaudeCodeCompat = *cfg.Sisyphus.Tasks.ClaudeCodeCompat
		}
	}

	// New Task System Enabled
	if cfg.NewTaskSystemEnabled != nil {
		w.newTaskSystemEnabled = *cfg.NewTaskSystemEnabled
	}

	// Default Run Agent
	if cfg.DefaultRunAgent != "" {
		w.defaultRunAgent.SetValue(cfg.DefaultRunAgent)
	}

	if cfg.ModelCapabilities != nil {
		if cfg.ModelCapabilities.Enabled != nil {
			w.mcEnabled = *cfg.ModelCapabilities.Enabled
		}
		if cfg.ModelCapabilities.AutoRefreshOnStart != nil {
			w.mcAutoRefreshOnStart = *cfg.ModelCapabilities.AutoRefreshOnStart
		}
		if cfg.ModelCapabilities.RefreshTimeoutMs != nil {
			w.mcRefreshTimeoutMs.SetValue(fmt.Sprintf("%d", *cfg.ModelCapabilities.RefreshTimeoutMs))
		}
		if cfg.ModelCapabilities.SourceURL != "" {
			w.mcSourceURL.SetValue(cfg.ModelCapabilities.SourceURL)
		}
	}

	if cfg.Openclaw != nil {
		if raw, err := json.MarshalIndent(cfg.Openclaw, "", "  "); err == nil {
			w.openclawEditor.SetValue(string(raw))
		}
	}

	if cfg.RuntimeFallback != nil {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, cfg.RuntimeFallback, "", "  "); err == nil {
			w.runtimeFallbackEditor.SetValue(prettyJSON.String())
		} else {
			w.runtimeFallbackEditor.SetValue(string(cfg.RuntimeFallback))
		}
	}

	// Skills JSON
	if cfg.Skills != nil {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, cfg.Skills, "", "  "); err == nil {
			w.skillsEditor.SetValue(prettyJSON.String())
		} else {
			w.skillsEditor.SetValue(string(cfg.Skills))
		}
	}

}

func (w *WizardOther) Apply(cfg *config.Config, selection *profile.FieldSelection) {
	w.selection = selection

	cfg.DisabledAgents = nil
	if w.fieldSelected(disabledAgentsFieldPath) {
		var agents []string
		for _, a := range disableableAgents {
			if w.disabledAgents[a] {
				agents = append(agents, a)
			}
		}
		cfg.DisabledAgents = emptySliceIfSelected(true, agents)
	}

	cfg.DisabledSkills = nil
	if w.fieldSelected(disabledSkillsFieldPath) {
		var skills []string
		for _, s := range disableableSkills {
			if w.disabledSkills[s] {
				skills = append(skills, s)
			}
		}
		cfg.DisabledSkills = emptySliceIfSelected(true, skills)
	}

	cfg.DisabledCommands = nil
	if w.fieldSelected(disabledCommandsFieldPath) {
		var commands []string
		for _, c := range disableableCommands {
			if w.disabledCommands[c] {
				commands = append(commands, c)
			}
		}
		cfg.DisabledCommands = emptySliceIfSelected(true, commands)
	}

	cfg.DisabledMCPs = nil
	if w.fieldSelected(disabledMcpsFieldPath) {
		var mcps []string
		for _, m := range strings.Split(w.disabledMcps.Value(), ",") {
			if s := strings.TrimSpace(m); s != "" {
				mcps = append(mcps, s)
			}
		}
		cfg.DisabledMCPs = emptySliceIfSelected(true, mcps)
	}

	cfg.DisabledTools = nil
	if w.fieldSelected(disabledToolsFieldPath) {
		var tools []string
		for _, t := range strings.Split(w.disabledTools.Value(), ",") {
			if s := strings.TrimSpace(t); s != "" {
				tools = append(tools, s)
			}
		}
		cfg.DisabledTools = emptySliceIfSelected(true, tools)
	}

	cfg.MCPEnvAllowlist = nil
	if w.fieldSelected(mcpEnvAllowlistFieldPath) {
		var envVars []string
		for _, e := range strings.Split(w.mcpEnvAllowlist.Value(), ",") {
			if s := strings.TrimSpace(e); s != "" {
				envVars = append(envVars, s)
			}
		}
		cfg.MCPEnvAllowlist = emptySliceIfSelected(true, envVars)
	}

	cfg.AutoUpdate = nil
	if w.fieldSelected(autoUpdateFieldPath) {
		cfg.AutoUpdate = wizardOtherBoolPtr(w.autoUpdate)
	}

	cfg.HashlineEdit = nil
	if w.fieldSelected(hashlineEditFieldPath) {
		cfg.HashlineEdit = wizardOtherBoolPtr(w.hashlineEdit)
	}

	cfg.ModelFallback = nil
	if w.fieldSelected(modelFallbackFieldPath) {
		cfg.ModelFallback = wizardOtherBoolPtr(w.modelFallback)
	}

	cfg.StartWork = nil
	if w.fieldSelected(startWorkAutoCommitFieldPath) {
		cfg.StartWork = &config.StartWorkConfig{AutoCommit: wizardOtherBoolPtr(w.startWorkAutoCommit)}
	}

	cfg.Experimental = nil
	if w.expHasData() {
		exp := &config.ExperimentalConfig{}
		if w.fieldSelected("experimental.aggressive_truncation") {
			exp.AggressiveTruncation = wizardOtherBoolPtr(w.expAggressiveTrunc)
		}
		if w.fieldSelected("experimental.auto_resume") {
			exp.AutoResume = wizardOtherBoolPtr(w.expAutoResume)
		}
		if w.fieldSelected("experimental.truncate_all_tool_outputs") {
			exp.TruncateAllToolOutputs = wizardOtherBoolPtr(w.expTruncateAllOutputs)
		}
		if w.fieldSelected("experimental.preemptive_compaction") {
			exp.PreemptiveCompaction = wizardOtherBoolPtr(w.expPreemptiveCompaction)
		}
		if w.fieldSelected("experimental.task_system") {
			exp.TaskSystem = wizardOtherBoolPtr(w.expTaskSystem)
		}
		if w.fieldSelected("experimental.plugin_load_timeout_ms") {
			if i, err := strconv.Atoi(strings.TrimSpace(w.expPluginLoadTimeoutMs.Value())); err == nil && i > 0 {
				exp.PluginLoadTimeoutMs = intPtr(i)
			}
		}
		if w.fieldSelected("experimental.safe_hook_creation") {
			exp.SafeHookCreation = wizardOtherBoolPtr(w.expSafeHookCreation)
		}
		if w.fieldSelected("experimental.hashline_edit") {
			exp.HashlineEdit = wizardOtherBoolPtr(w.expHashlineEdit)
		}
		if w.fieldSelected("experimental.disable_omo_env") {
			exp.DisableOmoEnv = wizardOtherBoolPtr(w.expDisableOmoEnv)
		}
		if w.fieldSelected("experimental.model_fallback_title") {
			exp.ModelFallbackTitle = wizardOtherBoolPtr(w.expModelFallbackTitle)
		}
		if w.fieldSelected("experimental.max_tools") {
			if v := parsePositiveInt64(w.expMaxTools.Value()); v != nil {
				exp.MaxTools = v
			}
		}
		if w.selectedWithPrefix("experimental.dynamic_context_pruning.") {
			dcp := &config.DynamicContextPruningConfig{}
			if w.fieldSelected("experimental.dynamic_context_pruning.enabled") {
				dcp.Enabled = wizardOtherBoolPtr(w.dcpEnabled)
			}
			if w.fieldSelected("experimental.dynamic_context_pruning.notification") {
				dcp.Notification = dcpNotificationValues[w.dcpNotificationIdx]
			}
			if w.fieldSelected("experimental.dynamic_context_pruning.turn_protection.enabled") || w.fieldSelected("experimental.dynamic_context_pruning.turn_protection.turns") {
				dcp.TurnProtection = &config.TurnProtectionConfig{}
				if w.fieldSelected("experimental.dynamic_context_pruning.turn_protection.enabled") {
					dcp.TurnProtection.Enabled = wizardOtherBoolPtr(w.dcpTurnProtEnabled)
				}
				if w.fieldSelected("experimental.dynamic_context_pruning.turn_protection.turns") {
					if i, err := strconv.Atoi(strings.TrimSpace(w.dcpTurnProtTurns.Value())); err == nil {
						dcp.TurnProtection.Turns = intPtr(i)
					}
				}
			}
			if w.fieldSelected("experimental.dynamic_context_pruning.protected_tools") {
				var tools []string
				for _, t := range strings.Split(w.dcpProtectedTools.Value(), ",") {
					if s := strings.TrimSpace(t); s != "" {
						tools = append(tools, s)
					}
				}
				dcp.ProtectedTools = emptySliceIfSelected(true, tools)
			}
			if w.selectedWithPrefix("experimental.dynamic_context_pruning.strategies.") {
				dcp.Strategies = &config.StrategiesConfig{}
				if w.fieldSelected("experimental.dynamic_context_pruning.strategies.deduplication.enabled") {
					dcp.Strategies.Deduplication = &config.DeduplicationConfig{Enabled: wizardOtherBoolPtr(w.dcpDeduplicationEnabled)}
				}
				if w.fieldSelected("experimental.dynamic_context_pruning.strategies.supersede_writes.enabled") || w.fieldSelected("experimental.dynamic_context_pruning.strategies.supersede_writes.aggressive") {
					dcp.Strategies.SupersedeWrites = &config.SupersedeWritesConfig{}
					if w.fieldSelected("experimental.dynamic_context_pruning.strategies.supersede_writes.enabled") {
						dcp.Strategies.SupersedeWrites.Enabled = wizardOtherBoolPtr(w.dcpSupersedeWritesEnabled)
					}
					if w.fieldSelected("experimental.dynamic_context_pruning.strategies.supersede_writes.aggressive") {
						dcp.Strategies.SupersedeWrites.Aggressive = wizardOtherBoolPtr(w.dcpSupersedeWritesAggressive)
					}
				}
				if w.fieldSelected("experimental.dynamic_context_pruning.strategies.purge_errors.enabled") || w.fieldSelected("experimental.dynamic_context_pruning.strategies.purge_errors.turns") {
					dcp.Strategies.PurgeErrors = &config.PurgeErrorsConfig{}
					if w.fieldSelected("experimental.dynamic_context_pruning.strategies.purge_errors.enabled") {
						dcp.Strategies.PurgeErrors.Enabled = wizardOtherBoolPtr(w.dcpPurgeErrorsEnabled)
					}
					if w.fieldSelected("experimental.dynamic_context_pruning.strategies.purge_errors.turns") {
						if i, err := strconv.Atoi(strings.TrimSpace(w.dcpPurgeErrorsTurns.Value())); err == nil {
							dcp.Strategies.PurgeErrors.Turns = intPtr(i)
						}
					}
				}
			}
			exp.DynamicContextPruning = dcp
		}
		cfg.Experimental = exp
	}

	cfg.ClaudeCode = nil
	if w.ccHasData() {
		cc := &config.ClaudeCodeConfig{}
		if w.fieldSelected("claude_code.mcp") {
			cc.MCP = wizardOtherBoolPtr(w.ccMcp)
		}
		if w.fieldSelected("claude_code.commands") {
			cc.Commands = wizardOtherBoolPtr(w.ccCommands)
		}
		if w.fieldSelected("claude_code.skills") {
			cc.Skills = wizardOtherBoolPtr(w.ccSkills)
		}
		if w.fieldSelected("claude_code.agents") {
			cc.Agents = wizardOtherBoolPtr(w.ccAgents)
		}
		if w.fieldSelected("claude_code.hooks") {
			cc.Hooks = wizardOtherBoolPtr(w.ccHooks)
		}
		if w.fieldSelected("claude_code.plugins") {
			cc.Plugins = wizardOtherBoolPtr(w.ccPlugins)
		}
		if w.fieldSelected("claude_code.plugins_override") {
			cc.PluginsOverride = emptyMapStringBoolIfSelected(true, parseMapStringBool(strings.TrimSpace(w.ccPluginsOverride.Value())))
		}
		cfg.ClaudeCode = cc
	}

	cfg.SisyphusAgent = nil
	if w.saHasData() {
		sa := &config.SisyphusAgentConfig{}
		if w.fieldSelected("sisyphus_agent.disabled") {
			sa.Disabled = wizardOtherBoolPtr(w.saDisabled)
		}
		if w.fieldSelected("sisyphus_agent.default_builder_enabled") {
			sa.DefaultBuilderEnabled = wizardOtherBoolPtr(w.saDefaultBuilderEnabled)
		}
		if w.fieldSelected("sisyphus_agent.planner_enabled") {
			sa.PlannerEnabled = wizardOtherBoolPtr(w.saPlannerEnabled)
		}
		if w.fieldSelected("sisyphus_agent.replace_plan") {
			sa.ReplacePlan = wizardOtherBoolPtr(w.saReplacePlan)
		}
		if w.fieldSelected("sisyphus_agent.tdd") {
			sa.TDD = wizardOtherBoolPtr(w.saTDD)
		}
		cfg.SisyphusAgent = sa
	}

	cfg.RalphLoop = nil
	if w.rlHasData() {
		rl := &config.RalphLoopConfig{}
		if w.fieldSelected("ralph_loop.enabled") {
			rl.Enabled = wizardOtherBoolPtr(w.rlEnabled)
		}
		if w.fieldSelected("ralph_loop.default_max_iterations") {
			if i, err := strconv.Atoi(strings.TrimSpace(w.rlDefaultMaxIterations.Value())); err == nil && i > 0 {
				rl.DefaultMaxIterations = intPtr(i)
			}
		}
		if w.fieldSelected("ralph_loop.state_dir") {
			rl.StateDir = strings.TrimSpace(w.rlStateDir.Value())
		}
		if w.fieldSelected("ralph_loop.default_strategy") {
			rl.DefaultStrategy = ralphLoopStrategies[w.rlDefaultStrategyIdx]
		}
		cfg.RalphLoop = rl
	}

	cfg.BackgroundTask = nil
	if w.btHasData() {
		bt := &config.BackgroundTaskConfig{}
		if w.fieldSelected("background_task.default_concurrency") {
			if i, err := strconv.Atoi(strings.TrimSpace(w.btDefaultConcurrency.Value())); err == nil && i > 0 {
				bt.DefaultConcurrency = intPtr(i)
			}
		}
		if w.fieldSelected("background_task.provider_concurrency") {
			bt.ProviderConcurrency = parseMapStringInt(strings.TrimSpace(w.btProviderConcurrency.Value()))
			if bt.ProviderConcurrency == nil {
				bt.ProviderConcurrency = map[string]int{}
			}
		}
		if w.fieldSelected("background_task.model_concurrency") {
			bt.ModelConcurrency = parseMapStringInt(strings.TrimSpace(w.btModelConcurrency.Value()))
			if bt.ModelConcurrency == nil {
				bt.ModelConcurrency = map[string]int{}
			}
		}
		if w.fieldSelected("background_task.max_depth") {
			bt.MaxDepth = parsePositiveInt64(w.btMaxDepth.Value())
		}
		if w.fieldSelected("background_task.stale_timeout_ms") {
			bt.StaleTimeoutMs = parsePositiveIntWithMinimum(w.btStaleTimeoutMs.Value(), 60000)
		}
		if w.fieldSelected("background_task.message_staleness_timeout_ms") {
			bt.MessageStalenessTimeoutMs = parsePositiveIntWithMinimum(w.btMessageStalenessTimeoutMs.Value(), 60000)
		}
		if w.fieldSelected("background_task.task_ttl_ms") {
			bt.TaskTtlMs = parsePositiveIntWithMinimum(w.btTaskTtlMs.Value(), 300000)
		}
		if w.fieldSelected("background_task.session_gone_timeout_ms") {
			bt.SessionGoneTimeoutMs = parsePositiveIntWithMinimum(w.btSessionGoneTimeoutMs.Value(), 10000)
		}
		if w.fieldSelected("background_task.sync_poll_timeout_ms") {
			bt.SyncPollTimeoutMs = parsePositiveIntWithMinimum(w.btSyncPollTimeoutMs.Value(), 60000)
		}
		if w.fieldSelected("background_task.max_tool_calls") {
			bt.MaxToolCalls = parsePositiveInt64(w.btMaxToolCalls.Value())
		}
		if w.selectedWithPrefix("background_task.circuit_breaker.") {
			cb := &config.BackgroundCircuitBreaker{}
			if w.fieldSelected("background_task.circuit_breaker.enabled") {
				cb.Enabled = wizardOtherBoolPtr(w.btCircuitBreakerEnabled)
			}
			if w.fieldSelected("background_task.circuit_breaker.max_tool_calls") {
				cb.MaxToolCalls = parsePositiveInt64(w.btCircuitBreakerMaxCalls.Value())
			}
			if w.fieldSelected("background_task.circuit_breaker.consecutive_threshold") {
				cb.ConsecutiveThreshold = parsePositiveInt64(w.btCircuitBreakerConsecutive.Value())
			}
			bt.CircuitBreaker = cb
		}
		cfg.BackgroundTask = bt
	}

	cfg.Notification = nil
	if w.fieldSelected("notification.force_enable") {
		cfg.Notification = &config.NotificationConfig{ForceEnable: wizardOtherBoolPtr(w.notifForceEnable)}
	}

	cfg.GitMaster = nil
	if w.selectedWithPrefix("git_master.") || (w.selection == nil && (w.gmCommitFooter || strings.TrimSpace(w.gmCommitFooterText.Value()) != "" || w.gmIncludeCoAuthoredBy || strings.TrimSpace(w.gmGitEnvPrefix.Value()) != "")) {
		gm := &config.GitMasterConfig{}
		if w.fieldSelected("git_master.commit_footer") {
			if text := strings.TrimSpace(w.gmCommitFooterText.Value()); text != "" {
				gm.CommitFooter = text
			} else {
				gm.CommitFooter = w.gmCommitFooter
			}
		}
		if w.fieldSelected("git_master.include_co_authored_by") {
			gm.IncludeCoAuthoredBy = wizardOtherBoolPtr(w.gmIncludeCoAuthoredBy)
		}
		if w.fieldSelected("git_master.git_env_prefix") {
			gm.GitEnvPrefix = strings.TrimSpace(w.gmGitEnvPrefix.Value())
		}
		cfg.GitMaster = gm
	}

	cfg.CommentChecker = nil
	if w.fieldSelected(commentCheckerCustomPromptFieldPath) {
		cfg.CommentChecker = &config.CommentCheckerConfig{CustomPrompt: w.ccCustomPrompt.Value()}
	}

	cfg.Babysitting = nil
	if w.fieldSelected(babysittingTimeoutFieldPath) {
		if f, err := strconv.ParseFloat(strings.TrimSpace(w.babysittingTimeoutMs.Value()), 64); err == nil && f > 0 {
			cfg.Babysitting = &config.BabysittingConfig{TimeoutMs: float64Ptr(f)}
		} else if w.selection != nil {
			cfg.Babysitting = &config.BabysittingConfig{}
		}
	}

	cfg.BrowserAutomationEngine = nil
	if w.fieldSelected(browserProviderFieldPath) {
		cfg.BrowserAutomationEngine = &config.BrowserAutomationEngineConfig{Provider: browserProviders[w.browserProviderIdx]}
	}

	cfg.Tmux = nil
	if w.tmuxHasData() {
		tmux := &config.TmuxConfig{}
		if w.fieldSelected("tmux.enabled") {
			tmux.Enabled = wizardOtherBoolPtr(w.tmuxEnabled)
		}
		if w.fieldSelected("tmux.layout") {
			tmux.Layout = tmuxLayouts[w.tmuxLayoutIdx]
		}
		if w.fieldSelected("tmux.main_pane_size") {
			if f, err := strconv.ParseFloat(strings.TrimSpace(w.tmuxMainPaneSize.Value()), 64); err == nil && f > 0 {
				tmux.MainPaneSize = float64Ptr(f)
			}
		}
		if w.fieldSelected("tmux.main_pane_min_width") {
			if f, err := strconv.ParseFloat(strings.TrimSpace(w.tmuxMainPaneMinWidth.Value()), 64); err == nil && f > 0 {
				tmux.MainPaneMinWidth = float64Ptr(f)
			}
		}
		if w.fieldSelected("tmux.agent_pane_min_width") {
			if f, err := strconv.ParseFloat(strings.TrimSpace(w.tmuxAgentPaneMinWidth.Value()), 64); err == nil && f > 0 {
				tmux.AgentPaneMinWidth = float64Ptr(f)
			}
		}
		if w.fieldSelected("tmux.isolation") {
			tmux.Isolation = tmuxIsolations[w.tmuxIsolationIdx]
		}
		cfg.Tmux = tmux
	}

	cfg.Websearch = nil
	if w.fieldSelected(websearchProviderFieldPath) {
		cfg.Websearch = &config.WebsearchConfig{Provider: websearchProviders[w.websearchProviderIdx]}
	}

	cfg.Sisyphus = nil
	if w.sisyphusHasData() {
		tasks := &config.SisyphusTasksConfig{}
		if w.fieldSelected("sisyphus.tasks.storage_path") {
			tasks.StoragePath = strings.TrimSpace(w.sisyphusTasksStoragePath.Value())
		}
		if w.fieldSelected("sisyphus.tasks.task_list_id") {
			tasks.TaskListID = strings.TrimSpace(w.sisyphusTasksTaskListID.Value())
		}
		if w.fieldSelected("sisyphus.tasks.claude_code_compat") {
			tasks.ClaudeCodeCompat = wizardOtherBoolPtr(w.sisyphusTasksClaudeCodeCompat)
		}
		cfg.Sisyphus = &config.SisyphusConfig{Tasks: tasks}
	}

	cfg.NewTaskSystemEnabled = nil
	if w.fieldSelected(newTaskSystemEnabledFieldPath) {
		cfg.NewTaskSystemEnabled = wizardOtherBoolPtr(w.newTaskSystemEnabled)
	}

	cfg.DefaultRunAgent = ""
	if w.fieldSelected(defaultRunAgentFieldPath) {
		cfg.DefaultRunAgent = strings.TrimSpace(w.defaultRunAgent.Value())
	}

	cfg.ModelCapabilities = nil
	if w.modelCapabilitiesHasData() {
		mc := &config.ModelCapabilitiesConfig{}
		if w.fieldSelected("model_capabilities.enabled") {
			mc.Enabled = wizardOtherBoolPtr(w.mcEnabled)
		}
		if w.fieldSelected("model_capabilities.auto_refresh_on_start") {
			mc.AutoRefreshOnStart = wizardOtherBoolPtr(w.mcAutoRefreshOnStart)
		}
		if w.fieldSelected("model_capabilities.refresh_timeout_ms") {
			mc.RefreshTimeoutMs = parsePositiveInt64(w.mcRefreshTimeoutMs.Value())
		}
		if w.fieldSelected("model_capabilities.source_url") {
			mc.SourceURL = strings.TrimSpace(w.mcSourceURL.Value())
		}
		cfg.ModelCapabilities = mc
	}

	cfg.Openclaw = nil
	if strings.TrimSpace(w.openclawEditor.Value()) != "" {
		var parsed config.OpenclawConfig
		if err := json.Unmarshal([]byte(w.openclawEditor.Value()), &parsed); err == nil && w.openclawHasData() {
			openclaw := &config.OpenclawConfig{}
			if w.fieldSelected(openclawEnabledFieldPath) {
				if parsed.Enabled != nil {
					openclaw.Enabled = wizardOtherBoolPtr(*parsed.Enabled)
				} else if w.selection != nil {
					openclaw.Enabled = wizardOtherBoolPtr(false)
				}
			}
			if w.fieldSelected(openclawGatewaysFieldPath) {
				if parsed.Gateways != nil {
					openclaw.Gateways = parsed.Gateways
				} else {
					openclaw.Gateways = map[string]*config.OpenclawGateway{}
				}
			}
			if w.fieldSelected(openclawHooksFieldPath) {
				if parsed.Hooks != nil {
					openclaw.Hooks = parsed.Hooks
				} else {
					openclaw.Hooks = map[string]*config.OpenclawHook{}
				}
			}
			if w.fieldSelected(openclawReplyListenerFieldPath) {
				if parsed.ReplyListener != nil {
					openclaw.ReplyListener = parsed.ReplyListener
				} else {
					openclaw.ReplyListener = &config.OpenclawReplyListenerConfig{}
				}
			}
			cfg.Openclaw = openclaw
		}
	}

	cfg.RuntimeFallback = nil
	if w.fieldSelected(runtimeFallbackFieldPath) {
		if v := strings.TrimSpace(w.runtimeFallbackEditor.Value()); v != "" {
			cfg.RuntimeFallback = json.RawMessage(v)
		}
	}

	cfg.Skills = nil
	if w.fieldSelected(skillsFieldPath) {
		if v := strings.TrimSpace(w.skillsEditor.Value()); v != "" {
			cfg.Skills = json.RawMessage(v)
		}
	}
}

