package config

import "encoding/json"

// Config is the root configuration struct
type Config struct {
	Schema           string                     `json:"$schema,omitempty"`
	DisabledMCPs     []string                   `json:"disabled_mcps,omitempty"`
	DisabledAgents   []string                   `json:"disabled_agents,omitempty"`
	DisabledSkills   []string                   `json:"disabled_skills,omitempty"`
	DisabledHooks    []string                   `json:"disabled_hooks,omitempty"`
	DisabledCommands []string                   `json:"disabled_commands,omitempty"`
	Agents           map[string]*AgentConfig    `json:"agents,omitempty"`
	Categories       map[string]*CategoryConfig `json:"categories,omitempty"`
	ClaudeCode       *ClaudeCodeConfig          `json:"claude_code,omitempty"`
	SisyphusAgent    *SisyphusAgentConfig       `json:"sisyphus_agent,omitempty"`
	CommentChecker   *CommentCheckerConfig      `json:"comment_checker,omitempty"`
	Experimental     *ExperimentalConfig        `json:"experimental,omitempty"`
	AutoUpdate       *bool                      `json:"auto_update,omitempty"`
	Skills           json.RawMessage            `json:"skills,omitempty"`
	RalphLoop        *RalphLoopConfig           `json:"ralph_loop,omitempty"`
	BackgroundTask   *BackgroundTaskConfig      `json:"background_task,omitempty"`
	Notification     *NotificationConfig        `json:"notification,omitempty"`
	GitMaster        *GitMasterConfig           `json:"git_master,omitempty"`
}

// AgentConfig has 14 optional fields
type AgentConfig struct {
	Model        string            `json:"model,omitempty"`
	Variant      string            `json:"variant,omitempty"`
	Category     string            `json:"category,omitempty"`
	Skills       []string          `json:"skills,omitempty"`
	Temperature  *float64          `json:"temperature,omitempty"`
	TopP         *float64          `json:"top_p,omitempty"`
	Prompt       string            `json:"prompt,omitempty"`
	PromptAppend string            `json:"prompt_append,omitempty"`
	Tools        map[string]bool   `json:"tools,omitempty"`
	Disable      *bool             `json:"disable,omitempty"`
	Description  string            `json:"description,omitempty"`
	Mode         string            `json:"mode,omitempty"`
	Color        string            `json:"color,omitempty"`
	Permission   *PermissionConfig `json:"permission,omitempty"`
}

// PermissionConfig - bash is interface{} to preserve string OR object
type PermissionConfig struct {
	Edit              string      `json:"edit,omitempty"`
	Bash              interface{} `json:"bash,omitempty"`
	Webfetch          string      `json:"webfetch,omitempty"`
	DoomLoop          string      `json:"doom_loop,omitempty"`
	ExternalDirectory string      `json:"external_directory,omitempty"`
}

// CategoryConfig
type CategoryConfig struct {
	Model           string          `json:"model"`
	Variant         string          `json:"variant,omitempty"`
	Temperature     *float64        `json:"temperature,omitempty"`
	TopP            *float64        `json:"top_p,omitempty"`
	MaxTokens       *float64        `json:"maxTokens,omitempty"`
	Thinking        *ThinkingConfig `json:"thinking,omitempty"`
	ReasoningEffort string          `json:"reasoningEffort,omitempty"`
	TextVerbosity   string          `json:"textVerbosity,omitempty"`
	Tools           map[string]bool `json:"tools,omitempty"`
	PromptAppend    string          `json:"prompt_append,omitempty"`
	IsUnstableAgent *bool           `json:"is_unstable_agent,omitempty"`
}

// ThinkingConfig
type ThinkingConfig struct {
	Type         string   `json:"type"`
	BudgetTokens *float64 `json:"budgetTokens,omitempty"`
}

// ClaudeCodeConfig
type ClaudeCodeConfig struct {
	MCP             *bool           `json:"mcp,omitempty"`
	Commands        *bool           `json:"commands,omitempty"`
	Skills          *bool           `json:"skills,omitempty"`
	Agents          *bool           `json:"agents,omitempty"`
	Hooks           *bool           `json:"hooks,omitempty"`
	Plugins         *bool           `json:"plugins,omitempty"`
	PluginsOverride map[string]bool `json:"plugins_override,omitempty"`
}

// SisyphusAgentConfig
type SisyphusAgentConfig struct {
	Disabled              *bool `json:"disabled,omitempty"`
	DefaultBuilderEnabled *bool `json:"default_builder_enabled,omitempty"`
	PlannerEnabled        *bool `json:"planner_enabled,omitempty"`
	ReplacePlan           *bool `json:"replace_plan,omitempty"`
}

// ExperimentalConfig
type ExperimentalConfig struct {
	AggressiveTruncation   *bool                        `json:"aggressive_truncation,omitempty"`
	AutoResume             *bool                        `json:"auto_resume,omitempty"`
	TruncateAllToolOutputs *bool                        `json:"truncate_all_tool_outputs,omitempty"`
	DynamicContextPruning  *DynamicContextPruningConfig `json:"dynamic_context_pruning,omitempty"`
}

// DynamicContextPruningConfig
type DynamicContextPruningConfig struct {
	Enabled        *bool                 `json:"enabled,omitempty"`
	Notification   string                `json:"notification,omitempty"`
	TurnProtection *TurnProtectionConfig `json:"turn_protection,omitempty"`
	ProtectedTools []string              `json:"protected_tools,omitempty"`
	Strategies     *StrategiesConfig     `json:"strategies,omitempty"`
}

// TurnProtectionConfig
type TurnProtectionConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
	Turns   *int  `json:"turns,omitempty"`
}

// StrategiesConfig
type StrategiesConfig struct {
	Deduplication   *DeduplicationConfig   `json:"deduplication,omitempty"`
	SupersedeWrites *SupersedeWritesConfig `json:"supersede_writes,omitempty"`
	PurgeErrors     *PurgeErrorsConfig     `json:"purge_errors,omitempty"`
}

// DeduplicationConfig
type DeduplicationConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// SupersedeWritesConfig
type SupersedeWritesConfig struct {
	Enabled    *bool `json:"enabled,omitempty"`
	Aggressive *bool `json:"aggressive,omitempty"`
}

// PurgeErrorsConfig
type PurgeErrorsConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
	Turns   *int  `json:"turns,omitempty"`
}

// RalphLoopConfig
type RalphLoopConfig struct {
	Enabled              *bool  `json:"enabled,omitempty"`
	DefaultMaxIterations *int   `json:"default_max_iterations,omitempty"`
	StateDir             string `json:"state_dir,omitempty"`
}

// BackgroundTaskConfig
type BackgroundTaskConfig struct {
	DefaultConcurrency  *int           `json:"defaultConcurrency,omitempty"`
	ProviderConcurrency map[string]int `json:"providerConcurrency,omitempty"`
	ModelConcurrency    map[string]int `json:"modelConcurrency,omitempty"`
	StaleTimeoutMs      *int           `json:"staleTimeoutMs,omitempty"`
}

// NotificationConfig
type NotificationConfig struct {
	ForceEnable *bool `json:"force_enable,omitempty"`
}

// GitMasterConfig
type GitMasterConfig struct {
	CommitFooter        *bool `json:"commit_footer,omitempty"`
	IncludeCoAuthoredBy *bool `json:"include_co_authored_by,omitempty"`
}

// CommentCheckerConfig
type CommentCheckerConfig struct {
	CustomPrompt string `json:"custom_prompt,omitempty"`
}
