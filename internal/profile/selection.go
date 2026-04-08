package profile

import (
	"sort"
	"strings"
)

var allFieldPaths = []string{
	"$schema",
	"disabled_mcps",
	"disabled_agents",
	"disabled_skills",
	"disabled_hooks",
	"disabled_commands",
	"hashline_edit",
	"model_fallback",
	"agents.*.model",
	"agents.*.fallback_models",
	"agents.*.variant",
	"agents.*.category",
	"agents.*.skills",
	"agents.*.temperature",
	"agents.*.top_p",
	"agents.*.prompt",
	"agents.*.prompt_append",
	"agents.*.tools",
	"agents.*.disable",
	"agents.*.description",
	"agents.*.mode",
	"agents.*.color",
	"agents.*.permission.edit",
	"agents.*.permission.bash",
	"agents.*.permission.webfetch",
	"agents.*.permission.task",
	"agents.*.permission.doom_loop",
	"agents.*.permission.external_directory",
	"agents.*.max_tokens",
	"agents.*.thinking.type",
	"agents.*.thinking.budget_tokens",
	"agents.*.reasoning_effort",
	"agents.*.text_verbosity",
	"agents.*.provider_options",
	"agents.*.ultrawork.model",
	"agents.*.ultrawork.variant",
	"agents.*.compaction.model",
	"agents.*.compaction.variant",
	"agents.*.allow_non_gpt_model",
	"categories.*.description",
	"categories.*.model",
	"categories.*.fallback_models",
	"categories.*.variant",
	"categories.*.temperature",
	"categories.*.top_p",
	"categories.*.max_tokens",
	"categories.*.thinking.type",
	"categories.*.thinking.budget_tokens",
	"categories.*.reasoning_effort",
	"categories.*.text_verbosity",
	"categories.*.tools",
	"categories.*.prompt_append",
	"categories.*.max_prompt_tokens",
	"categories.*.is_unstable_agent",
	"categories.*.disable",
	"claude_code.mcp",
	"claude_code.commands",
	"claude_code.skills",
	"claude_code.agents",
	"claude_code.hooks",
	"claude_code.plugins",
	"claude_code.plugins_override",
	"sisyphus_agent.disabled",
	"sisyphus_agent.default_builder_enabled",
	"sisyphus_agent.planner_enabled",
	"sisyphus_agent.replace_plan",
	"sisyphus_agent.tdd",
	"comment_checker.custom_prompt",
	"experimental.aggressive_truncation",
	"experimental.auto_resume",
	"experimental.preemptive_compaction",
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
	"experimental.task_system",
	"experimental.plugin_load_timeout_ms",
	"experimental.safe_hook_creation",
	"experimental.disable_omo_env",
	"experimental.hashline_edit",
	"experimental.model_fallback_title",
	"experimental.max_tools",
	"auto_update",
	"skills",
	"ralph_loop.enabled",
	"ralph_loop.default_max_iterations",
	"ralph_loop.state_dir",
	"ralph_loop.default_strategy",
	"runtime_fallback",
	"background_task.default_concurrency",
	"background_task.provider_concurrency",
	"background_task.model_concurrency",
	"background_task.max_depth",
	"background_task.max_descendants",
	"background_task.stale_timeout_ms",
	"background_task.message_staleness_timeout_ms",
	"background_task.task_ttl_ms",
	"background_task.session_gone_timeout_ms",
	"background_task.sync_poll_timeout_ms",
	"background_task.max_tool_calls",
	"background_task.circuit_breaker.enabled",
	"background_task.circuit_breaker.max_tool_calls",
	"background_task.circuit_breaker.consecutive_threshold",
	"notification.force_enable",
	"git_master.commit_footer",
	"git_master.include_co_authored_by",
	"git_master.git_env_prefix",
	"new_task_system_enabled",
	"disabled_tools",
	"babysitting.timeout_ms",
	"browser_automation_engine.provider",
	"tmux.enabled",
	"tmux.layout",
	"tmux.main_pane_size",
	"tmux.main_pane_min_width",
	"tmux.agent_pane_min_width",
	"tmux.isolation",
	"websearch.provider",
	"sisyphus.tasks.storage_path",
	"sisyphus.tasks.claude_code_compat",
	"sisyphus.tasks.task_list_id",
	"default_run_agent",
	"start_work.auto_commit",
	"openclaw.enabled",
	"openclaw.gateways.*.type",
	"openclaw.gateways.*.url",
	"openclaw.gateways.*.method",
	"openclaw.gateways.*.headers",
	"openclaw.gateways.*.command",
	"openclaw.gateways.*.timeout",
	"openclaw.hooks.*.enabled",
	"openclaw.hooks.*.gateway",
	"openclaw.hooks.*.instruction",
	"openclaw.reply_listener.discord_bot_token",
	"openclaw.reply_listener.discord_channel_id",
	"openclaw.reply_listener.discord_mention",
	"openclaw.reply_listener.authorized_discord_user_ids",
	"openclaw.reply_listener.telegram_bot_token",
	"openclaw.reply_listener.telegram_chat_id",
	"openclaw.reply_listener.poll_interval_ms",
	"openclaw.reply_listener.rate_limit_per_minute",
	"openclaw.reply_listener.max_message_length",
	"openclaw.reply_listener.include_prefix",
	"model_capabilities.enabled",
	"model_capabilities.auto_refresh_on_start",
	"model_capabilities.refresh_timeout_ms",
	"model_capabilities.source_url",
	"_migrations",
}

type FieldSelection struct {
	selected map[string]bool
}

func NewBlankSelection() *FieldSelection {
	return &FieldSelection{selected: make(map[string]bool)}
}

func NewSelectionFromPresence(presence map[string]bool) *FieldSelection {
	selection := NewBlankSelection()
	for _, path := range allFieldPaths {
		if presence[topLevelPath(path)] {
			selection.selected[path] = true
		}
	}
	return selection
}

func NewSelectionFromTemplate(templatePresence map[string]bool) *FieldSelection {
	return NewSelectionFromPresence(templatePresence)
}

func (s *FieldSelection) IsSelected(path string) bool {
	if s == nil {
		return false
	}

	if s.selected[path] {
		return true
	}

	for _, candidate := range wildcardCandidates(path) {
		if s.selected[candidate] {
			return true
		}
	}

	return false
}

func (s *FieldSelection) SetSelected(path string, selected bool) {
	if s == nil {
		return
	}
	if s.selected == nil {
		s.selected = make(map[string]bool)
	}
	if selected {
		s.selected[path] = true
		return
	}
	delete(s.selected, path)
}

func (s *FieldSelection) Toggle(path string) {
	if s == nil {
		return
	}
	s.SetSelected(path, !s.selected[path])
}

func (s *FieldSelection) SelectedPaths() []string {
	if s == nil || len(s.selected) == 0 {
		return nil
	}

	paths := make([]string, 0, len(s.selected))
	for path, selected := range s.selected {
		if selected {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func (s *FieldSelection) Clone() *FieldSelection {
	clone := NewBlankSelection()
	if s == nil {
		return clone
	}
	for path, selected := range s.selected {
		if selected {
			clone.selected[path] = true
		}
	}
	return clone
}

func topLevelPath(path string) string {
	parts := strings.SplitN(path, ".", 2)
	return parts[0]
}

func wildcardCandidates(path string) []string {
	parts := strings.Split(path, ".")
	if len(parts) < 3 {
		return nil
	}

	seen := make(map[string]struct{}, len(parts)-2)
	var candidates []string
	for i := 1; i < len(parts)-1; i++ {
		candidateParts := append([]string(nil), parts...)
		candidateParts[i] = "*"
		candidate := strings.Join(candidateParts, ".")
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}
	return candidates
}
