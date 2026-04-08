package views

import (
	"bytes"
	"encoding/json"
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
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

var (
	wizOtherPurple = lipgloss.Color("#7D56F4")
	wizOtherGreen  = lipgloss.Color("#A6E3A1")
	wizOtherRed    = lipgloss.Color("#F38BA8")
	wizOtherGray   = lipgloss.Color("#6C7086")
	wizOtherWhite  = lipgloss.Color("#CDD6F4")
)

var (
	wizOtherSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizOtherPurple)
	wizOtherEnabledStyle  = lipgloss.NewStyle().Foreground(wizOtherGreen)
	wizOtherDisabledStyle = lipgloss.NewStyle().Foreground(wizOtherRed)
	wizOtherDimStyle      = lipgloss.NewStyle().Foreground(wizOtherGray)
	wizOtherLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizOtherWhite)
	wizOtherHelpStyle     = lipgloss.NewStyle().Foreground(wizOtherGray)
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

// Disableable agents (9) - matches schema disabled_agents enum
var disableableAgents = []string{
	"sisyphus",
	"hephaestus",
	"prometheus",
	"oracle",
	"librarian",
	"explore",
	"multimodal-looker",
	"metis",
	"momus",
	"atlas",
}

// Disableable skills (3) - matches schema disabled_skills enum
var disableableSkills = []string{
	"playwright",
	"agent-browser",
	"dev-browser",
	"frontend-ui-ux",
	"git-master",
}

// Disableable commands
var disableableCommands = []string{
	"init-deep",
	"ralph-loop",
	"ulw-loop",
	"cancel-ralph",
	"refactor",
	"start-work",
	"stop-continuation",
	"remove-ai-slops",
}

var dcpNotificationValues = []string{"", "off", "minimal", "detailed"}
var browserProviders = []string{"", "playwright", "playwright-cli", "agent-browser", "dev-browser"}
var tmuxLayouts = []string{"", "main-horizontal", "main-vertical", "tiled", "even-horizontal", "even-vertical"}
var tmuxIsolations = []string{"", "inline", "window", "session"}
var websearchProviders = []string{"", "exa", "tavily"}
var ralphLoopStrategies = []string{"", "reset", "continue"}

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
	openclawEnabledFieldPath            = "openclaw.enabled"
	openclawGatewaysFieldPath           = "openclaw.gateways.*.type"
	openclawHooksFieldPath              = "openclaw.hooks.*.enabled"
	openclawReplyListenerFieldPath      = "openclaw.reply_listener.discord_bot_token"
)

// Sections in the other settings
type otherSection int

const (
	sectionDisabledMcps otherSection = iota
	sectionDisabledAgents
	sectionDisabledSkills
	sectionDisabledCommands
	sectionDisabledTools
	sectionAutoUpdate
	sectionExperimental
	sectionClaudeCode
	sectionSisyphusAgent
	sectionRalphLoop
	sectionBackgroundTask
	sectionNotification
	sectionGitMaster
	sectionCommentChecker
	sectionBabysitting
	sectionBrowserAutomationEngine
	sectionTmux
	sectionWebsearch
	sectionSisyphus
	sectionNewTaskSystemEnabled
	sectionDefaultRunAgent
	sectionHashlineEdit
	sectionModelFallback
	sectionStartWork
	sectionModelCapabilities
	sectionOpenclaw
	sectionRuntimeFallback
	sectionSkillsJson
)

var otherSectionNames = []string{
	"Disabled MCPs",
	"Disabled Agents",
	"Disabled Skills",
	"Disabled Commands",
	"Disabled Tools",
	"Auto Update",
	"Experimental",
	"Claude Code",
	"Sisyphus Agent",
	"Ralph Loop",
	"Background Task",
	"Notification",
	"Git Master",
	"Comment Checker",
	"Babysitting",
	"Browser Automation Engine",
	"Tmux",
	"Websearch",
	"Sisyphus",
	"New Task System Enabled",
	"Default Run Agent",
	"Hashline Edit",
	"Model Fallback",
	"Start Work",
	"Model Capabilities",
	"Openclaw (JSON)",
	"Runtime Fallback (JSON)",
	"Skills (JSON)",
}

type wizardOtherKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
	Expand key.Binding
	Next   key.Binding
	Back   key.Binding
	Left   key.Binding
	Right  key.Binding
}

func newWizardOtherKeyMap() wizardOtherKeyMap {
	return wizardOtherKeyMap{
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
			key.WithHelp("enter", "expand/edit"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next step"),
		),
		Back: key.NewBinding(
			key.WithKeys("shift+tab", "esc"),
			key.WithHelp("shift+tab/esc", "back"),
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

func int64Ptr(v int64) *int64 {
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
	case sectionRuntimeFallback:
		return runtimeFallbackFieldPath
	case sectionSkillsJson:
		return skillsFieldPath
	default:
		return ""
	}
}

func (w WizardOther) isSimpleBooleanSection(section otherSection) bool {
	switch section {
	case sectionAutoUpdate, sectionNewTaskSystemEnabled, sectionHashlineEdit, sectionModelFallback:
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
	case sectionStartWork:
		if idx == 0 {
			return startWorkAutoCommitFieldPath
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

func (w WizardOther) isBooleanField(section otherSection, idx int) bool {
	path := w.subSectionFieldPath(section, idx)
	switch path {
	case "experimental.aggressive_truncation", "experimental.auto_resume", "experimental.truncate_all_tool_outputs",
		"experimental.dynamic_context_pruning.enabled", "experimental.dynamic_context_pruning.turn_protection.enabled",
		"experimental.dynamic_context_pruning.strategies.deduplication.enabled",
		"experimental.dynamic_context_pruning.strategies.supersede_writes.enabled",
		"experimental.dynamic_context_pruning.strategies.supersede_writes.aggressive",
		"experimental.dynamic_context_pruning.strategies.purge_errors.enabled", "experimental.preemptive_compaction",
		"experimental.task_system", "experimental.safe_hook_creation", "experimental.hashline_edit",
		"experimental.disable_omo_env", "experimental.model_fallback_title", "claude_code.mcp", "claude_code.commands",
		"claude_code.skills", "claude_code.agents", "claude_code.hooks", "claude_code.plugins",
		"sisyphus_agent.disabled", "sisyphus_agent.default_builder_enabled", "sisyphus_agent.planner_enabled",
		"sisyphus_agent.replace_plan", "sisyphus_agent.tdd", "ralph_loop.enabled", "notification.force_enable",
		"git_master.commit_footer", "git_master.include_co_authored_by", "tmux.enabled",
		"sisyphus.tasks.claude_code_compat", startWorkAutoCommitFieldPath,
		"model_capabilities.enabled", "model_capabilities.auto_refresh_on_start", openclawEnabledFieldPath:
		return true
	default:
		return false
	}
}

func onOff(v bool) string {
	if v {
		return "[on]"
	}
	return "[off]"
}

// WizardOther is step 4: Other settings
type WizardOther struct {
	selection *profile.FieldSelection

	// Disabled lists
	disabledMcps     textinput.Model
	disabledAgents   map[string]bool
	disabledSkills   map[string]bool
	disabledCommands map[string]bool
	disabledTools    textinput.Model

	// Auto update
	autoUpdate          bool
	hashlineEdit        bool
	modelFallback       bool
	startWorkAutoCommit bool

	// Experimental flags
	expAggressiveTrunc      bool
	expAutoResume           bool
	expTruncateAllOutputs   bool
	expPreemptiveCompaction bool
	expTaskSystem           bool
	expPluginLoadTimeoutMs  textinput.Model
	expSafeHookCreation     bool
	expHashlineEdit         bool
	expDisableOmoEnv        bool
	expModelFallbackTitle   bool
	expMaxTools             textinput.Model

	dcpEnabled                   bool
	dcpNotificationIdx           int
	dcpTurnProtEnabled           bool
	dcpTurnProtTurns             textinput.Model
	dcpProtectedTools            textinput.Model
	dcpDeduplicationEnabled      bool
	dcpSupersedeWritesEnabled    bool
	dcpSupersedeWritesAggressive bool
	dcpPurgeErrorsEnabled        bool
	dcpPurgeErrorsTurns          textinput.Model

	// Claude Code
	ccMcp             bool
	ccCommands        bool
	ccSkills          bool
	ccAgents          bool
	ccHooks           bool
	ccPlugins         bool
	ccPluginsOverride textinput.Model

	// Sisyphus Agent
	saDisabled              bool
	saDefaultBuilderEnabled bool
	saPlannerEnabled        bool
	saReplacePlan           bool
	saTDD                   bool

	// Ralph Loop
	rlEnabled              bool
	rlDefaultMaxIterations textinput.Model
	rlStateDir             textinput.Model
	rlDefaultStrategyIdx   int

	// Background Task
	btDefaultConcurrency        textinput.Model
	btProviderConcurrency       textinput.Model
	btModelConcurrency          textinput.Model
	btMaxDepth                  textinput.Model
	btMaxDescendants            textinput.Model
	btStaleTimeoutMs            textinput.Model
	btMessageStalenessTimeoutMs textinput.Model
	btTaskTtlMs                 textinput.Model
	btSessionGoneTimeoutMs      textinput.Model
	btSyncPollTimeoutMs         textinput.Model
	btMaxToolCalls              textinput.Model
	btCircuitBreakerEnabled     bool
	btCircuitBreakerMaxCalls    textinput.Model
	btCircuitBreakerConsecutive textinput.Model

	// Notification
	notifForceEnable bool

	// Git Master
	gmCommitFooter        bool
	gmCommitFooterText    textinput.Model
	gmIncludeCoAuthoredBy bool
	gmGitEnvPrefix        textinput.Model

	// Comment Checker
	ccCustomPrompt textinput.Model

	// Babysitting
	babysittingTimeoutMs textinput.Model

	// Browser Automation Engine
	browserProviderIdx int

	// Tmux
	tmuxEnabled           bool
	tmuxLayoutIdx         int
	tmuxMainPaneSize      textinput.Model
	tmuxMainPaneMinWidth  textinput.Model
	tmuxAgentPaneMinWidth textinput.Model
	tmuxIsolationIdx      int

	// Websearch
	websearchProviderIdx int

	// Sisyphus
	sisyphusTasksStoragePath      textinput.Model
	sisyphusTasksTaskListID       textinput.Model
	sisyphusTasksClaudeCodeCompat bool

	// New Task System Enabled
	newTaskSystemEnabled bool

	// Default Run Agent
	defaultRunAgent textinput.Model

	mcEnabled            bool
	mcAutoRefreshOnStart bool
	mcRefreshTimeoutMs   textinput.Model
	mcSourceURL          textinput.Model

	openclawEditor textarea.Model

	// Skills JSON
	runtimeFallbackEditor textarea.Model
	skillsEditor          textarea.Model

	// UI State
	currentSection     otherSection
	sectionExpanded    map[otherSection]bool
	subCursor          int
	inSubSection       bool
	subValueFocused    bool
	simpleValueFocused bool
	viewport           viewport.Model
	ready              bool
	width              int
	height             int
	keys               wizardOtherKeyMap
}

func NewWizardOther() WizardOther {
	disabledMcps := textinput.New()
	disabledMcps.Placeholder = "mcp1, mcp2, ..."
	disabledMcps.Width = 50

	disabledTools := textinput.New()
	disabledTools.Placeholder = "tool1, tool2, ..."
	disabledTools.Width = 50

	rlMaxIter := textinput.New()
	rlMaxIter.Placeholder = "10"
	rlMaxIter.Width = 10

	rlStateDir := textinput.New()
	rlStateDir.Placeholder = "/path/to/state"
	rlStateDir.Width = 40

	btConcurrency := textinput.New()
	btConcurrency.Placeholder = "4"
	btConcurrency.Width = 10

	btProviderConcurrency := textinput.New()
	btProviderConcurrency.Placeholder = "anthropic:5, openai:3"
	btProviderConcurrency.Width = 40

	btModelConcurrency := textinput.New()
	btModelConcurrency.Placeholder = "claude-3:2, gpt-4:3"
	btModelConcurrency.Width = 40

	btStaleTimeoutMs := textinput.New()
	btStaleTimeoutMs.Placeholder = "60000"
	btStaleTimeoutMs.Width = 15

	btMaxDepth := textinput.New()
	btMaxDepth.Placeholder = "8"
	btMaxDepth.Width = 10

	btMaxDescendants := textinput.New()
	btMaxDescendants.Placeholder = "100"
	btMaxDescendants.Width = 10

	btMsgStaleTimeout := textinput.New()
	btMsgStaleTimeout.Placeholder = "60000"
	btMsgStaleTimeout.Width = 15

	btTaskTtlMs := textinput.New()
	btTaskTtlMs.Placeholder = "300000"
	btTaskTtlMs.Width = 15

	btSessionGoneTimeoutMs := textinput.New()
	btSessionGoneTimeoutMs.Placeholder = "10000"
	btSessionGoneTimeoutMs.Width = 15

	btSyncPollTimeoutMs := textinput.New()
	btSyncPollTimeoutMs.Placeholder = "60000"
	btSyncPollTimeoutMs.Width = 15

	btMaxToolCalls := textinput.New()
	btMaxToolCalls.Placeholder = "50"
	btMaxToolCalls.Width = 10

	btCircuitBreakerMaxCalls := textinput.New()
	btCircuitBreakerMaxCalls.Placeholder = "25"
	btCircuitBreakerMaxCalls.Width = 10

	btCircuitBreakerConsecutive := textinput.New()
	btCircuitBreakerConsecutive.Placeholder = "5"
	btCircuitBreakerConsecutive.Width = 10

	ccPrompt := textinput.New()
	ccPrompt.Placeholder = "custom prompt..."
	ccPrompt.Width = 50

	dcpTurnProtTurns := textinput.New()
	dcpTurnProtTurns.Placeholder = "3"
	dcpTurnProtTurns.Width = 10

	dcpProtectedTools := textinput.New()
	dcpProtectedTools.Placeholder = "tool1, tool2"
	dcpProtectedTools.Width = 40

	dcpPurgeErrorsTurns := textinput.New()
	dcpPurgeErrorsTurns.Placeholder = "5"
	dcpPurgeErrorsTurns.Width = 10

	expPluginLoadTimeoutMs := textinput.New()
	expPluginLoadTimeoutMs.Placeholder = "30000"
	expPluginLoadTimeoutMs.Width = 10

	expMaxTools := textinput.New()
	expMaxTools.Placeholder = "64"
	expMaxTools.Width = 10

	ccPluginsOverride := textinput.New()
	ccPluginsOverride.Placeholder = "serena:false, context7:true"
	ccPluginsOverride.Width = 40

	// Initialize disabled maps
	disabledAgents := make(map[string]bool)
	for _, a := range disableableAgents {
		disabledAgents[a] = false
	}

	disabledSkills := make(map[string]bool)
	for _, s := range disableableSkills {
		disabledSkills[s] = false
	}

	disabledCommands := make(map[string]bool)
	for _, c := range disableableCommands {
		disabledCommands[c] = false
	}

	sectionExpanded := make(map[otherSection]bool)

	skillsEditor := textarea.New()
	skillsEditor.Placeholder = `["skill1", "skill2"] or {"sources": [...]}`
	skillsEditor.SetWidth(60)
	skillsEditor.SetHeight(8)

	runtimeFallbackEditor := textarea.New()
	runtimeFallbackEditor.Placeholder = `true or {"enabled": true, "max_fallback_attempts": 3}`
	runtimeFallbackEditor.SetWidth(60)
	runtimeFallbackEditor.SetHeight(6)

	tmuxSize := textinput.New()
	tmuxSize.Placeholder = "0.75"
	tmuxSize.Width = 10

	tmuxMinWidth := textinput.New()
	tmuxMinWidth.Placeholder = "0.5"
	tmuxMinWidth.Width = 10

	tmuxAgentWidth := textinput.New()
	tmuxAgentWidth.Placeholder = "0.3"
	tmuxAgentWidth.Width = 10

	sisTasksStoragePath := textinput.New()
	sisTasksStoragePath.Placeholder = ".sisyphus/tasks"
	sisTasksStoragePath.Width = 40

	sisTasksTaskListID := textinput.New()
	sisTasksTaskListID.Placeholder = "default"
	sisTasksTaskListID.Width = 30

	babysittingTimeoutMs := textinput.New()
	babysittingTimeoutMs.Placeholder = "60000"
	babysittingTimeoutMs.Width = 10

	gmCommitFooterText := textinput.New()
	gmCommitFooterText.Placeholder = "Signed-off-by: ..."
	gmCommitFooterText.Width = 40

	gmGitEnvPrefix := textinput.New()
	gmGitEnvPrefix.Placeholder = "GIT_MASTER=1"
	gmGitEnvPrefix.Width = 25
	gmGitEnvPrefix.SetValue("GIT_MASTER=1")

	defaultRunAgent := textinput.New()
	defaultRunAgent.Placeholder = "build"
	defaultRunAgent.Width = 30

	mcRefreshTimeoutMs := textinput.New()
	mcRefreshTimeoutMs.Placeholder = "60000"
	mcRefreshTimeoutMs.Width = 10

	mcSourceURL := textinput.New()
	mcSourceURL.Placeholder = "https://models.dev/api"
	mcSourceURL.Width = 50

	openclawEditor := textarea.New()
	openclawEditor.Placeholder = `{"enabled": true, "gateways": {}, "hooks": {}, "replyListener": {}}`
	openclawEditor.SetWidth(60)
	openclawEditor.SetHeight(6)

	return WizardOther{
		disabledMcps:                disabledMcps,
		disabledAgents:              disabledAgents,
		disabledSkills:              disabledSkills,
		disabledCommands:            disabledCommands,
		disabledTools:               disabledTools,
		expPluginLoadTimeoutMs:      expPluginLoadTimeoutMs,
		expMaxTools:                 expMaxTools,
		rlDefaultMaxIterations:      rlMaxIter,
		rlStateDir:                  rlStateDir,
		btDefaultConcurrency:        btConcurrency,
		btProviderConcurrency:       btProviderConcurrency,
		btModelConcurrency:          btModelConcurrency,
		btMaxDepth:                  btMaxDepth,
		btMaxDescendants:            btMaxDescendants,
		btStaleTimeoutMs:            btStaleTimeoutMs,
		btMessageStalenessTimeoutMs: btMsgStaleTimeout,
		btTaskTtlMs:                 btTaskTtlMs,
		btSessionGoneTimeoutMs:      btSessionGoneTimeoutMs,
		btSyncPollTimeoutMs:         btSyncPollTimeoutMs,
		btMaxToolCalls:              btMaxToolCalls,
		btCircuitBreakerMaxCalls:    btCircuitBreakerMaxCalls,
		btCircuitBreakerConsecutive: btCircuitBreakerConsecutive,
		ccCustomPrompt:              ccPrompt,
		babysittingTimeoutMs:        babysittingTimeoutMs,
		gmCommitFooterText:          gmCommitFooterText,
		gmGitEnvPrefix:              gmGitEnvPrefix,
		tmuxMainPaneSize:            tmuxSize,
		tmuxMainPaneMinWidth:        tmuxMinWidth,
		tmuxAgentPaneMinWidth:       tmuxAgentWidth,
		sisyphusTasksStoragePath:    sisTasksStoragePath,
		sisyphusTasksTaskListID:     sisTasksTaskListID,
		defaultRunAgent:             defaultRunAgent,
		mcRefreshTimeoutMs:          mcRefreshTimeoutMs,
		mcSourceURL:                 mcSourceURL,
		dcpTurnProtTurns:            dcpTurnProtTurns,
		dcpProtectedTools:           dcpProtectedTools,
		dcpPurgeErrorsTurns:         dcpPurgeErrorsTurns,
		ccPluginsOverride:           ccPluginsOverride,
		openclawEditor:              openclawEditor,
		runtimeFallbackEditor:       runtimeFallbackEditor,
		skillsEditor:                skillsEditor,
		tmuxLayoutIdx:               2,
		tmuxIsolationIdx:            3,
		sectionExpanded:             sectionExpanded,
		keys:                        newWizardOtherKeyMap(),
	}
}

func (w WizardOther) Init() tea.Cmd {
	return nil
}

func (w *WizardOther) SetSize(width, height int) {
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
	if w.disabledAgents == nil {
		return
	}

	wide := layout.WideFieldWidth(width, 10)
	w.disabledMcps.Width = wide
	w.disabledTools.Width = wide
	w.rlDefaultMaxIterations.Width = layout.FixedSmallWidth()
	w.rlStateDir.Width = wide
	w.btDefaultConcurrency.Width = layout.FixedSmallWidth()
	w.btProviderConcurrency.Width = wide
	w.btModelConcurrency.Width = wide
	w.btStaleTimeoutMs.Width = layout.FixedSmallWidth()
	w.btMaxDepth.Width = layout.FixedSmallWidth()
	w.btMaxDescendants.Width = layout.FixedSmallWidth()
	w.btMessageStalenessTimeoutMs.Width = layout.FixedSmallWidth()
	w.btTaskTtlMs.Width = layout.FixedSmallWidth()
	w.btSessionGoneTimeoutMs.Width = layout.FixedSmallWidth()
	w.btSyncPollTimeoutMs.Width = layout.FixedSmallWidth()
	w.btMaxToolCalls.Width = layout.FixedSmallWidth()
	w.btCircuitBreakerMaxCalls.Width = layout.FixedSmallWidth()
	w.btCircuitBreakerConsecutive.Width = layout.FixedSmallWidth()
	w.ccCustomPrompt.Width = wide
	w.dcpTurnProtTurns.Width = layout.FixedSmallWidth()
	w.dcpProtectedTools.Width = wide
	w.dcpPurgeErrorsTurns.Width = layout.FixedSmallWidth()
	w.expPluginLoadTimeoutMs.Width = layout.FixedSmallWidth()
	w.expMaxTools.Width = layout.FixedSmallWidth()
	w.ccPluginsOverride.Width = wide
	w.babysittingTimeoutMs.Width = layout.FixedSmallWidth()
	w.gmCommitFooterText.Width = wide
	w.gmGitEnvPrefix.Width = wide
	w.tmuxMainPaneSize.Width = layout.FixedSmallWidth()
	w.tmuxMainPaneMinWidth.Width = layout.FixedSmallWidth()
	w.tmuxAgentPaneMinWidth.Width = layout.FixedSmallWidth()
	w.sisyphusTasksStoragePath.Width = wide
	w.sisyphusTasksTaskListID.Width = wide
	w.defaultRunAgent.Width = wide
	w.mcRefreshTimeoutMs.Width = layout.FixedSmallWidth()
	w.mcSourceURL.Width = wide
	w.openclawEditor.SetWidth(wide)
	w.runtimeFallbackEditor.SetWidth(wide)
	w.skillsEditor.SetWidth(wide)
	w.viewport.SetContent(w.renderContent())
}

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
		if cfg.BackgroundTask.MaxDescendants != nil {
			w.btMaxDescendants.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.MaxDescendants))
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
		if w.fieldSelected("background_task.max_descendants") {
			bt.MaxDescendants = parsePositiveInt64(w.btMaxDescendants.Value())
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
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "up", "k":
				if w.subCursor > 0 {
					w.subCursor--
				}
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "down", "j":
				w.subCursor++
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case " ":
				w.toggleSubItem()
				w.viewport.SetContent(w.renderContent())
				return w, nil
			case "enter", "right", "l":
				path := w.subSectionFieldPath(w.currentSection, w.subCursor)
				if path == "" && !(w.currentSection == sectionOpenclaw && w.subCursor == 4) && !(w.currentSection == sectionRuntimeFallback && w.subCursor == 1) && !(w.currentSection == sectionSkillsJson && w.subCursor == 1) {
					return w, nil
				}
				w.subValueFocused = true
				w.viewport.SetContent(w.renderContent())
				return w, nil
			}
		}

		if w.inSubSection {
			if w.currentSection == sectionExperimental && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.dcpTurnProtTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpTurnProtTurns.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpTurnProtTurns.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpTurnProtTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpTurnProtTurns.Focus()
					w.dcpTurnProtTurns, cmd = w.dcpTurnProtTurns.Update(msg)
					return w, cmd
				default:
					w.dcpTurnProtTurns.Focus()
					w.dcpTurnProtTurns, cmd = w.dcpTurnProtTurns.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 7 {
				switch msg.String() {
				case "esc":
					w.dcpProtectedTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpProtectedTools.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpProtectedTools.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpProtectedTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpProtectedTools.Focus()
					w.dcpProtectedTools, cmd = w.dcpProtectedTools.Update(msg)
					return w, cmd
				default:
					w.dcpProtectedTools.Focus()
					w.dcpProtectedTools, cmd = w.dcpProtectedTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 12 {
				switch msg.String() {
				case "esc":
					w.dcpPurgeErrorsTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.dcpPurgeErrorsTurns.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.dcpPurgeErrorsTurns.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.dcpPurgeErrorsTurns.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.dcpPurgeErrorsTurns.Focus()
					w.dcpPurgeErrorsTurns, cmd = w.dcpPurgeErrorsTurns.Update(msg)
					return w, cmd
				default:
					w.dcpPurgeErrorsTurns.Focus()
					w.dcpPurgeErrorsTurns, cmd = w.dcpPurgeErrorsTurns.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 4 {
				switch msg.String() {
				case "right", "l":
					w.dcpNotificationIdx = (w.dcpNotificationIdx + 1) % len(dcpNotificationValues)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.dcpNotificationIdx = (w.dcpNotificationIdx - 1 + len(dcpNotificationValues)) % len(dcpNotificationValues)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 15 {
				switch msg.String() {
				case "esc":
					w.expPluginLoadTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.expPluginLoadTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.expPluginLoadTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.expPluginLoadTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.expPluginLoadTimeoutMs.Focus()
					w.expPluginLoadTimeoutMs, cmd = w.expPluginLoadTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.expPluginLoadTimeoutMs.Focus()
					w.expPluginLoadTimeoutMs, cmd = w.expPluginLoadTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 20 {
				switch msg.String() {
				case "esc":
					w.expMaxTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.expMaxTools.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.expMaxTools.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.expMaxTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.expMaxTools.Focus()
					w.expMaxTools, cmd = w.expMaxTools.Update(msg)
					return w, cmd
				default:
					w.expMaxTools.Focus()
					w.expMaxTools, cmd = w.expMaxTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionClaudeCode && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.ccPluginsOverride.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.ccPluginsOverride.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.ccPluginsOverride.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.ccPluginsOverride.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.ccPluginsOverride.Focus()
					w.ccPluginsOverride, cmd = w.ccPluginsOverride.Update(msg)
					return w, cmd
				default:
					w.ccPluginsOverride.Focus()
					w.ccPluginsOverride, cmd = w.ccPluginsOverride.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 0 {
				switch msg.String() {
				case "esc":
					w.btProviderConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btProviderConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btProviderConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btProviderConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btProviderConcurrency.Focus()
					w.btProviderConcurrency, cmd = w.btProviderConcurrency.Update(msg)
					return w, cmd
				default:
					w.btProviderConcurrency.Focus()
					w.btProviderConcurrency, cmd = w.btProviderConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.btModelConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btModelConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btModelConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btModelConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btModelConcurrency.Focus()
					w.btModelConcurrency, cmd = w.btModelConcurrency.Update(msg)
					return w, cmd
				default:
					w.btModelConcurrency.Focus()
					w.btModelConcurrency, cmd = w.btModelConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDisabledMcps {
				switch msg.String() {
				case "esc":
					w.disabledMcps.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.disabledMcps.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.disabledMcps.Focus()
					w.disabledMcps, cmd = w.disabledMcps.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDisabledTools {
				switch msg.String() {
				case "esc":
					w.disabledTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.disabledTools.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.disabledTools.Focus()
					w.disabledTools, cmd = w.disabledTools.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.rlDefaultMaxIterations.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.rlDefaultMaxIterations.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.rlDefaultMaxIterations.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.rlDefaultMaxIterations.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.rlDefaultMaxIterations.Focus()
					w.rlDefaultMaxIterations, cmd = w.rlDefaultMaxIterations.Update(msg)
					return w, cmd
				default:
					w.rlDefaultMaxIterations.Focus()
					w.rlDefaultMaxIterations, cmd = w.rlDefaultMaxIterations.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.rlStateDir.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.rlStateDir.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.rlStateDir.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.rlStateDir.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.rlStateDir.Focus()
					w.rlStateDir, cmd = w.rlStateDir.Update(msg)
					return w, cmd
				default:
					w.rlStateDir.Focus()
					w.rlStateDir, cmd = w.rlStateDir.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRalphLoop && w.subCursor == 3 {
				switch msg.String() {
				case "right", "l":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx + 1) % len(ralphLoopStrategies)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.rlDefaultStrategyIdx = (w.rlDefaultStrategyIdx - 1 + len(ralphLoopStrategies)) % len(ralphLoopStrategies)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.btDefaultConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btDefaultConcurrency.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btDefaultConcurrency.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btDefaultConcurrency.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btDefaultConcurrency.Focus()
					w.btDefaultConcurrency, cmd = w.btDefaultConcurrency.Update(msg)
					return w, cmd
				default:
					w.btDefaultConcurrency.Focus()
					w.btDefaultConcurrency, cmd = w.btDefaultConcurrency.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.btStaleTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btStaleTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btStaleTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btStaleTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btStaleTimeoutMs.Focus()
					w.btStaleTimeoutMs, cmd = w.btStaleTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btStaleTimeoutMs.Focus()
					w.btStaleTimeoutMs, cmd = w.btStaleTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.btMessageStalenessTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMessageStalenessTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMessageStalenessTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMessageStalenessTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMessageStalenessTimeoutMs.Focus()
					w.btMessageStalenessTimeoutMs, cmd = w.btMessageStalenessTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btMessageStalenessTimeoutMs.Focus()
					w.btMessageStalenessTimeoutMs, cmd = w.btMessageStalenessTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 5 {
				switch msg.String() {
				case "esc":
					w.btSyncPollTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btSyncPollTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btSyncPollTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btSyncPollTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btSyncPollTimeoutMs.Focus()
					w.btSyncPollTimeoutMs, cmd = w.btSyncPollTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btSyncPollTimeoutMs.Focus()
					w.btSyncPollTimeoutMs, cmd = w.btSyncPollTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 6 {
				switch msg.String() {
				case "esc":
					w.btMaxDepth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxDepth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxDepth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxDepth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxDepth.Focus()
					w.btMaxDepth, cmd = w.btMaxDepth.Update(msg)
					return w, cmd
				default:
					w.btMaxDepth.Focus()
					w.btMaxDepth, cmd = w.btMaxDepth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 7 {
				switch msg.String() {
				case "esc":
					w.btMaxDescendants.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxDescendants.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxDescendants.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxDescendants.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxDescendants.Focus()
					w.btMaxDescendants, cmd = w.btMaxDescendants.Update(msg)
					return w, cmd
				default:
					w.btMaxDescendants.Focus()
					w.btMaxDescendants, cmd = w.btMaxDescendants.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 8 {
				switch msg.String() {
				case "esc":
					w.btTaskTtlMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btTaskTtlMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btTaskTtlMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btTaskTtlMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btTaskTtlMs.Focus()
					w.btTaskTtlMs, cmd = w.btTaskTtlMs.Update(msg)
					return w, cmd
				default:
					w.btTaskTtlMs.Focus()
					w.btTaskTtlMs, cmd = w.btTaskTtlMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 9 {
				switch msg.String() {
				case "esc":
					w.btSessionGoneTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btSessionGoneTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btSessionGoneTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btSessionGoneTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btSessionGoneTimeoutMs.Focus()
					w.btSessionGoneTimeoutMs, cmd = w.btSessionGoneTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.btSessionGoneTimeoutMs.Focus()
					w.btSessionGoneTimeoutMs, cmd = w.btSessionGoneTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 10 {
				switch msg.String() {
				case "esc":
					w.btMaxToolCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btMaxToolCalls.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btMaxToolCalls.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btMaxToolCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btMaxToolCalls.Focus()
					w.btMaxToolCalls, cmd = w.btMaxToolCalls.Update(msg)
					return w, cmd
				default:
					w.btMaxToolCalls.Focus()
					w.btMaxToolCalls, cmd = w.btMaxToolCalls.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 12 {
				switch msg.String() {
				case "esc":
					w.btCircuitBreakerMaxCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btCircuitBreakerMaxCalls.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btCircuitBreakerMaxCalls.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btCircuitBreakerMaxCalls.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btCircuitBreakerMaxCalls.Focus()
					w.btCircuitBreakerMaxCalls, cmd = w.btCircuitBreakerMaxCalls.Update(msg)
					return w, cmd
				default:
					w.btCircuitBreakerMaxCalls.Focus()
					w.btCircuitBreakerMaxCalls, cmd = w.btCircuitBreakerMaxCalls.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBackgroundTask && w.subCursor == 13 {
				switch msg.String() {
				case "esc":
					w.btCircuitBreakerConsecutive.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.btCircuitBreakerConsecutive.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.btCircuitBreakerConsecutive.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.btCircuitBreakerConsecutive.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.btCircuitBreakerConsecutive.Focus()
					w.btCircuitBreakerConsecutive, cmd = w.btCircuitBreakerConsecutive.Update(msg)
					return w, cmd
				default:
					w.btCircuitBreakerConsecutive.Focus()
					w.btCircuitBreakerConsecutive, cmd = w.btCircuitBreakerConsecutive.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionGitMaster && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.gmCommitFooterText.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.gmCommitFooterText.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.gmCommitFooterText.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.gmCommitFooterText.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.gmCommitFooterText.Focus()
					w.gmCommitFooterText, cmd = w.gmCommitFooterText.Update(msg)
					return w, cmd
				default:
					w.gmCommitFooterText.Focus()
					w.gmCommitFooterText, cmd = w.gmCommitFooterText.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionGitMaster && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.gmGitEnvPrefix.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.gmGitEnvPrefix.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.gmGitEnvPrefix.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.gmGitEnvPrefix.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.gmGitEnvPrefix.Focus()
					w.gmGitEnvPrefix, cmd = w.gmGitEnvPrefix.Update(msg)
					return w, cmd
				default:
					w.gmGitEnvPrefix.Focus()
					w.gmGitEnvPrefix, cmd = w.gmGitEnvPrefix.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 5 {
				switch msg.String() {
				case "right", "l":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx + 1) % len(tmuxIsolations)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.tmuxIsolationIdx = (w.tmuxIsolationIdx - 1 + len(tmuxIsolations)) % len(tmuxIsolations)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionStartWork {
				switch msg.String() {
				case "esc", "tab":
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.startWorkAutoCommit = !w.startWorkAutoCommit
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionCommentChecker {
				switch msg.String() {
				case "esc":
					w.ccCustomPrompt.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.ccCustomPrompt.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.ccCustomPrompt.Focus()
					w.ccCustomPrompt, cmd = w.ccCustomPrompt.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBabysitting {
				switch msg.String() {
				case "esc":
					w.babysittingTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.babysittingTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.babysittingTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.babysittingTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.babysittingTimeoutMs.Focus()
					w.babysittingTimeoutMs, cmd = w.babysittingTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.babysittingTimeoutMs.Focus()
					w.babysittingTimeoutMs, cmd = w.babysittingTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionBrowserAutomationEngine && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.browserProviderIdx = (w.browserProviderIdx + 1) % len(browserProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.browserProviderIdx = (w.browserProviderIdx - 1 + len(browserProviders)) % len(browserProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 1 {
				switch msg.String() {
				case "right", "l":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx + 1) % len(tmuxLayouts)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.tmuxLayoutIdx = (w.tmuxLayoutIdx - 1 + len(tmuxLayouts)) % len(tmuxLayouts)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.tmuxMainPaneSize.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxMainPaneSize.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxMainPaneSize.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxMainPaneSize.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxMainPaneSize.Focus()
					w.tmuxMainPaneSize, cmd = w.tmuxMainPaneSize.Update(msg)
					return w, cmd
				default:
					w.tmuxMainPaneSize.Focus()
					w.tmuxMainPaneSize, cmd = w.tmuxMainPaneSize.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.tmuxMainPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxMainPaneMinWidth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxMainPaneMinWidth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxMainPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxMainPaneMinWidth.Focus()
					w.tmuxMainPaneMinWidth, cmd = w.tmuxMainPaneMinWidth.Update(msg)
					return w, cmd
				default:
					w.tmuxMainPaneMinWidth.Focus()
					w.tmuxMainPaneMinWidth, cmd = w.tmuxMainPaneMinWidth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionTmux && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.tmuxAgentPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.tmuxAgentPaneMinWidth.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.tmuxAgentPaneMinWidth.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.tmuxAgentPaneMinWidth.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.tmuxAgentPaneMinWidth.Focus()
					w.tmuxAgentPaneMinWidth, cmd = w.tmuxAgentPaneMinWidth.Update(msg)
					return w, cmd
				default:
					w.tmuxAgentPaneMinWidth.Focus()
					w.tmuxAgentPaneMinWidth, cmd = w.tmuxAgentPaneMinWidth.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionWebsearch && w.subCursor == 0 {
				switch msg.String() {
				case "right", "l":
					w.websearchProviderIdx = (w.websearchProviderIdx + 1) % len(websearchProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "left", "h":
					w.websearchProviderIdx = (w.websearchProviderIdx - 1 + len(websearchProviders)) % len(websearchProviders)
					w.viewport.SetContent(w.renderContent())
					return w, nil
				}
			}

			if w.currentSection == sectionSisyphus && w.subCursor == 0 {
				switch msg.String() {
				case "esc":
					w.sisyphusTasksStoragePath.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.sisyphusTasksStoragePath.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.sisyphusTasksStoragePath.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.sisyphusTasksStoragePath.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.sisyphusTasksStoragePath.Focus()
					w.sisyphusTasksStoragePath, cmd = w.sisyphusTasksStoragePath.Update(msg)
					return w, cmd
				default:
					w.sisyphusTasksStoragePath.Focus()
					w.sisyphusTasksStoragePath, cmd = w.sisyphusTasksStoragePath.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionSisyphus && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.sisyphusTasksTaskListID.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.sisyphusTasksTaskListID.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.sisyphusTasksTaskListID.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.sisyphusTasksTaskListID.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.sisyphusTasksTaskListID.Focus()
					w.sisyphusTasksTaskListID, cmd = w.sisyphusTasksTaskListID.Update(msg)
					return w, cmd
				default:
					w.sisyphusTasksTaskListID.Focus()
					w.sisyphusTasksTaskListID, cmd = w.sisyphusTasksTaskListID.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionDefaultRunAgent {
				switch msg.String() {
				case "esc":
					w.defaultRunAgent.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.defaultRunAgent.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.defaultRunAgent.Focus()
					w.defaultRunAgent, cmd = w.defaultRunAgent.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionModelCapabilities && w.subCursor == 2 {
				switch msg.String() {
				case "esc":
					w.mcRefreshTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.mcRefreshTimeoutMs.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.mcRefreshTimeoutMs.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.mcRefreshTimeoutMs.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.mcRefreshTimeoutMs.Focus()
					w.mcRefreshTimeoutMs, cmd = w.mcRefreshTimeoutMs.Update(msg)
					return w, cmd
				default:
					w.mcRefreshTimeoutMs.Focus()
					w.mcRefreshTimeoutMs, cmd = w.mcRefreshTimeoutMs.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionModelCapabilities && w.subCursor == 3 {
				switch msg.String() {
				case "esc":
					w.mcSourceURL.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.mcSourceURL.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.mcSourceURL.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.mcSourceURL.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.mcSourceURL.Focus()
					w.mcSourceURL, cmd = w.mcSourceURL.Update(msg)
					return w, cmd
				default:
					w.mcSourceURL.Focus()
					w.mcSourceURL, cmd = w.mcSourceURL.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionOpenclaw && w.subCursor == 4 {
				switch msg.String() {
				case "esc":
					w.openclawEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.openclawEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.openclawEditor.Focus()
					w.openclawEditor, cmd = w.openclawEditor.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionRuntimeFallback && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.runtimeFallbackEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.runtimeFallbackEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.runtimeFallbackEditor.Focus()
					w.runtimeFallbackEditor, cmd = w.runtimeFallbackEditor.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionSkillsJson && w.subCursor == 1 {
				switch msg.String() {
				case "esc":
					w.skillsEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.skillsEditor.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				default:
					w.skillsEditor.Focus()
					w.skillsEditor, cmd = w.skillsEditor.Update(msg)
					return w, cmd
				}
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
			w.viewport.SetContent(w.renderContent())
			return w, nil
		}

		switch {
		case key.Matches(msg, w.keys.Up):
			w.simpleValueFocused = false
			if w.currentSection > 0 {
				w.currentSection--
			}
		case key.Matches(msg, w.keys.Down):
			w.simpleValueFocused = false
			if w.currentSection < sectionSkillsJson {
				w.currentSection++
			}
		case key.Matches(msg, w.keys.Toggle):
			w.toggleSection()
		case key.Matches(msg, w.keys.Expand):
			w.sectionExpanded[w.currentSection] = !w.sectionExpanded[w.currentSection]
			if w.sectionExpanded[w.currentSection] {
				w.inSubSection = true
				w.subCursor = 0
				w.subValueFocused = false
			}
		case key.Matches(msg, w.keys.Right):
			if w.isSimpleBooleanSection(w.currentSection) {
				w.simpleValueFocused = true
				break
			}
			if !w.inSubSection && !w.sectionExpanded[w.currentSection] {
				w.sectionExpanded[w.currentSection] = true
				w.inSubSection = true
				w.subCursor = 0
				w.subValueFocused = false
			}
		case key.Matches(msg, w.keys.Left):
			if w.isSimpleBooleanSection(w.currentSection) {
				w.simpleValueFocused = false
				break
			}
			if !w.inSubSection && w.sectionExpanded[w.currentSection] {
				w.sectionExpanded[w.currentSection] = false
			}
		case key.Matches(msg, w.keys.Next):
			return w, func() tea.Msg { return WizardNextMsg{} }
		case key.Matches(msg, w.keys.Back):
			return w, func() tea.Msg { return WizardBackMsg{} }
		}
	}

	// Update viewport
	w.viewport.SetContent(w.renderContent())
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
	}
}

func (w *WizardOther) toggleSubItem() {
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
	case sectionStartWork:
		if w.subCursor == 0 {
			w.startWorkAutoCommit = !w.startWorkAutoCommit
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

func (w WizardOther) renderContent() string {
	var lines []string

	selectedStyle := wizOtherSelectedStyle
	enabledStyle := wizOtherEnabledStyle
	labelStyle := wizOtherLabelStyle

	for i, name := range otherSectionNames {
		section := otherSection(i)
		cursor := "  "
		if section == w.currentSection && !w.inSubSection {
			cursor = selectedStyle.Render("> ")
		}

		expandIcon := "▶"
		if w.sectionExpanded[section] {
			expandIcon = "▼"
		}

		if path := w.topLevelFieldPath(section); path != "" {
			checkbox := "[ ]"
			if w.fieldSelected(path) {
				checkbox = enabledStyle.Render("[✓]")
			}
			value := ""
			switch section {
			case sectionAutoUpdate:
				value = onOff(w.autoUpdate)
			case sectionNewTaskSystemEnabled:
				value = onOff(w.newTaskSystemEnabled)
			case sectionHashlineEdit:
				value = onOff(w.hashlineEdit)
			case sectionModelFallback:
				value = onOff(w.modelFallback)
			}
			if value != "" {
				line := fmt.Sprintf("%s%s %s: %s", cursor, checkbox, labelStyle.Render(name), value)
				lines = append(lines, line)
				continue
			}
		}

		line := fmt.Sprintf("%s%s %s", cursor, expandIcon, labelStyle.Render(name))
		lines = append(lines, line)
		if w.sectionExpanded[section] {
			lines = append(lines, w.renderSubSection(section)...)
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
	case sectionExperimental:
		lines = append(lines, renderBoolField(0, "experimental.aggressive_truncation", "aggressive_truncation", w.expAggressiveTrunc))
		lines = append(lines, renderBoolField(1, "experimental.auto_resume", "auto_resume", w.expAutoResume))
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
		lines = append(lines, renderValueField(7, "background_task.max_descendants", "max_descendants", w.btMaxDescendants.View()))
		lines = append(lines, renderValueField(8, "background_task.task_ttl_ms", "task_ttl_ms", w.btTaskTtlMs.View()))
		lines = append(lines, renderValueField(9, "background_task.session_gone_timeout_ms", "session_gone_timeout_ms", w.btSessionGoneTimeoutMs.View()))
		lines = append(lines, renderValueField(10, "background_task.max_tool_calls", "max_tool_calls", w.btMaxToolCalls.View()))
		lines = append(lines, renderBoolField(11, "background_task.circuit_breaker.enabled", "circuit_breaker.enabled", w.btCircuitBreakerEnabled))
		lines = append(lines, renderValueField(12, "background_task.circuit_breaker.max_tool_calls", "circuit_breaker.max_tool_calls", w.btCircuitBreakerMaxCalls.View()))
		lines = append(lines, renderValueField(13, "background_task.circuit_breaker.consecutive_threshold", "circuit_breaker.consecutive_threshold", w.btCircuitBreakerConsecutive.View()))
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
	case sectionStartWork:
		lines = append(lines, renderBoolField(0, startWorkAutoCommitFieldPath, "auto_commit", w.startWorkAutoCommit))
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
	}

	lines = append(lines, "")
	return lines
}

func (w WizardOther) View() string {
	titleStyle := wizOtherLabelStyle
	helpStyle := wizOtherHelpStyle

	title := titleStyle.Render("Other Settings")
	desc := helpStyle.Render("Enter to expand • Space to toggle • Tab next • Shift+Tab back")

	if w.inSubSection {
		desc = helpStyle.Render("Space/Enter to toggle • Esc to close section")
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
