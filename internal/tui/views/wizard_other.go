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

// Disableable agents (10)
var disableableAgents = []string{
	"Sisyphus",
	"oracle",
	"librarian",
	"explore",
	"frontend-ui-ux-engineer",
	"document-writer",
	"multimodal-looker",
	"Metis (Plan Consultant)",
	"Momus (Plan Reviewer)",
	"orchestrator-sisyphus",
}

// Disableable skills (3)
var disableableSkills = []string{
	"playwright",
	"frontend-ui-ux",
	"git-master",
}

// Disableable commands (2)
var disableableCommands = []string{
	"init-deep",
	"start-work",
}

var dcpNotificationValues = []string{"", "off", "minimal", "detailed"}

// Sections in the other settings
type otherSection int

const (
	sectionDisabledMcps otherSection = iota
	sectionDisabledAgents
	sectionDisabledSkills
	sectionDisabledCommands
	sectionAutoUpdate
	sectionExperimental
	sectionClaudeCode
	sectionSisyphusAgent
	sectionRalphLoop
	sectionBackgroundTask
	sectionNotification
	sectionGitMaster
	sectionCommentChecker
	sectionSkillsJson
)

var otherSectionNames = []string{
	"Disabled MCPs",
	"Disabled Agents",
	"Disabled Skills",
	"Disabled Commands",
	"Auto Update",
	"Experimental",
	"Claude Code",
	"Sisyphus Agent",
	"Ralph Loop",
	"Background Task",
	"Notification",
	"Git Master",
	"Comment Checker",
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

// WizardOther is step 4: Other settings
type WizardOther struct {
	// Disabled lists
	disabledMcps     textinput.Model
	disabledAgents   map[string]bool
	disabledSkills   map[string]bool
	disabledCommands map[string]bool

	// Auto update
	autoUpdate bool

	// Experimental flags
	expAggressiveTrunc    bool
	expAutoResume         bool
	expPreemptiveCompact  bool
	expTruncateAllOutputs bool
	expDcpForCompaction   bool
	expThreshold          textinput.Model

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

	// Ralph Loop
	rlEnabled              bool
	rlDefaultMaxIterations textinput.Model
	rlStateDir             textinput.Model

	// Background Task
	btDefaultConcurrency  textinput.Model
	btProviderConcurrency textinput.Model
	btModelConcurrency    textinput.Model

	// Notification
	notifForceEnable bool

	// Git Master
	gmCommitFooter        bool
	gmIncludeCoAuthoredBy bool

	// Comment Checker
	ccCustomPrompt textinput.Model

	// Skills JSON
	skillsEditor textarea.Model

	// UI State
	currentSection  otherSection
	sectionExpanded map[otherSection]bool
	subCursor       int
	inSubSection    bool
	viewport        viewport.Model
	ready           bool
	width           int
	height          int
	keys            wizardOtherKeyMap
}

func NewWizardOther() WizardOther {
	disabledMcps := textinput.New()
	disabledMcps.Placeholder = "mcp1, mcp2, ..."
	disabledMcps.Width = 50

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

	ccPrompt := textinput.New()
	ccPrompt.Placeholder = "custom prompt..."
	ccPrompt.Width = 50

	expThreshold := textinput.New()
	expThreshold.Placeholder = "0.8"
	expThreshold.Width = 10

	dcpTurnProtTurns := textinput.New()
	dcpTurnProtTurns.Placeholder = "3"
	dcpTurnProtTurns.Width = 10

	dcpProtectedTools := textinput.New()
	dcpProtectedTools.Placeholder = "tool1, tool2"
	dcpProtectedTools.Width = 40

	dcpPurgeErrorsTurns := textinput.New()
	dcpPurgeErrorsTurns.Placeholder = "5"
	dcpPurgeErrorsTurns.Width = 10

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

	return WizardOther{
		disabledMcps:           disabledMcps,
		disabledAgents:         disabledAgents,
		disabledSkills:         disabledSkills,
		disabledCommands:       disabledCommands,
		rlDefaultMaxIterations: rlMaxIter,
		rlStateDir:             rlStateDir,
		btDefaultConcurrency:   btConcurrency,
		btProviderConcurrency:  btProviderConcurrency,
		btModelConcurrency:     btModelConcurrency,
		ccCustomPrompt:         ccPrompt,
		expThreshold:           expThreshold,
		dcpTurnProtTurns:       dcpTurnProtTurns,
		dcpProtectedTools:      dcpProtectedTools,
		dcpPurgeErrorsTurns:    dcpPurgeErrorsTurns,
		ccPluginsOverride:      ccPluginsOverride,
		skillsEditor:           skillsEditor,
		sectionExpanded:        sectionExpanded,
		keys:                   newWizardOtherKeyMap(),
	}
}

func (w WizardOther) Init() tea.Cmd {
	return nil
}

func (w *WizardOther) SetSize(width, height int) {
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

func (w *WizardOther) SetConfig(cfg *config.Config) {
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

	// Auto update
	if cfg.AutoUpdate != nil {
		w.autoUpdate = *cfg.AutoUpdate
	}

	// Experimental
	if cfg.Experimental != nil {
		if cfg.Experimental.AggressiveTruncation != nil {
			w.expAggressiveTrunc = *cfg.Experimental.AggressiveTruncation
		}
		if cfg.Experimental.AutoResume != nil {
			w.expAutoResume = *cfg.Experimental.AutoResume
		}
		if cfg.Experimental.PreemptiveCompaction != nil {
			w.expPreemptiveCompact = *cfg.Experimental.PreemptiveCompaction
		}
		if cfg.Experimental.TruncateAllToolOutputs != nil {
			w.expTruncateAllOutputs = *cfg.Experimental.TruncateAllToolOutputs
		}
		if cfg.Experimental.DcpForCompaction != nil {
			w.expDcpForCompaction = *cfg.Experimental.DcpForCompaction
		}
		if cfg.Experimental.PreemptiveCompactionThreshold != nil {
			w.expThreshold.SetValue(fmt.Sprintf("%v", *cfg.Experimental.PreemptiveCompactionThreshold))
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
	}

	// Background Task
	if cfg.BackgroundTask != nil {
		if cfg.BackgroundTask.DefaultConcurrency != nil {
			w.btDefaultConcurrency.SetValue(fmt.Sprintf("%d", *cfg.BackgroundTask.DefaultConcurrency))
		}
		if len(cfg.BackgroundTask.ProviderConcurrency) > 0 {
			w.btProviderConcurrency.SetValue(serializeMapStringInt(cfg.BackgroundTask.ProviderConcurrency))
		}
		if len(cfg.BackgroundTask.ModelConcurrency) > 0 {
			w.btModelConcurrency.SetValue(serializeMapStringInt(cfg.BackgroundTask.ModelConcurrency))
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
			w.gmCommitFooter = *cfg.GitMaster.CommitFooter
		}
		if cfg.GitMaster.IncludeCoAuthoredBy != nil {
			w.gmIncludeCoAuthoredBy = *cfg.GitMaster.IncludeCoAuthoredBy
		}
	}

	// Comment Checker
	if cfg.CommentChecker != nil {
		if cfg.CommentChecker.CustomPrompt != "" {
			w.ccCustomPrompt.SetValue(cfg.CommentChecker.CustomPrompt)
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

func (w *WizardOther) Apply(cfg *config.Config) {
	// Disabled agents
	var agents []string
	for _, a := range disableableAgents {
		if w.disabledAgents[a] {
			agents = append(agents, a)
		}
	}
	cfg.DisabledAgents = agents
	if len(cfg.DisabledAgents) == 0 {
		cfg.DisabledAgents = nil
	}

	// Disabled skills
	var skills []string
	for _, s := range disableableSkills {
		if w.disabledSkills[s] {
			skills = append(skills, s)
		}
	}
	cfg.DisabledSkills = skills
	if len(cfg.DisabledSkills) == 0 {
		cfg.DisabledSkills = nil
	}

	// Disabled commands
	var commands []string
	for _, c := range disableableCommands {
		if w.disabledCommands[c] {
			commands = append(commands, c)
		}
	}
	cfg.DisabledCommands = commands
	if len(cfg.DisabledCommands) == 0 {
		cfg.DisabledCommands = nil
	}

	// Disabled MCPs
	if v := w.disabledMcps.Value(); v != "" {
		var mcps []string
		for _, m := range strings.Split(v, ",") {
			if s := strings.TrimSpace(m); s != "" {
				mcps = append(mcps, s)
			}
		}
		cfg.DisabledMCPs = mcps
	}
	if len(cfg.DisabledMCPs) == 0 {
		cfg.DisabledMCPs = nil
	}

	// Auto update
	if w.autoUpdate {
		cfg.AutoUpdate = &w.autoUpdate
	}

	// Experimental - only set if any flag is true or threshold has value
	expHasData := w.expAggressiveTrunc || w.expAutoResume || w.expPreemptiveCompact ||
		w.expTruncateAllOutputs || w.expDcpForCompaction ||
		w.expThreshold.Value() != "" ||
		w.dcpEnabled || w.dcpNotificationIdx > 0 || w.dcpTurnProtEnabled ||
		w.dcpTurnProtTurns.Value() != "" || w.dcpProtectedTools.Value() != "" ||
		w.dcpDeduplicationEnabled || w.dcpSupersedeWritesEnabled ||
		w.dcpSupersedeWritesAggressive || w.dcpPurgeErrorsEnabled ||
		w.dcpPurgeErrorsTurns.Value() != ""
	if expHasData {
		cfg.Experimental = &config.ExperimentalConfig{}
		if w.expAggressiveTrunc {
			cfg.Experimental.AggressiveTruncation = &w.expAggressiveTrunc
		}
		if w.expAutoResume {
			cfg.Experimental.AutoResume = &w.expAutoResume
		}
		if w.expPreemptiveCompact {
			cfg.Experimental.PreemptiveCompaction = &w.expPreemptiveCompact
		}
		if w.expTruncateAllOutputs {
			cfg.Experimental.TruncateAllToolOutputs = &w.expTruncateAllOutputs
		}
		if w.expDcpForCompaction {
			cfg.Experimental.DcpForCompaction = &w.expDcpForCompaction
		}
		if v := w.expThreshold.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				cfg.Experimental.PreemptiveCompactionThreshold = &f
			}
			// If parse fails, field is not set (skip, don't error)
		}

		dcpHasData := w.dcpEnabled || w.dcpNotificationIdx > 0 || w.dcpTurnProtEnabled ||
			w.dcpTurnProtTurns.Value() != "" || w.dcpProtectedTools.Value() != "" ||
			w.dcpDeduplicationEnabled || w.dcpSupersedeWritesEnabled ||
			w.dcpSupersedeWritesAggressive || w.dcpPurgeErrorsEnabled ||
			w.dcpPurgeErrorsTurns.Value() != ""
		if dcpHasData {
			cfg.Experimental.DynamicContextPruning = &config.DynamicContextPruningConfig{}
			dcp := cfg.Experimental.DynamicContextPruning

			if w.dcpEnabled {
				dcp.Enabled = &w.dcpEnabled
			}
			if w.dcpNotificationIdx > 0 {
				dcp.Notification = dcpNotificationValues[w.dcpNotificationIdx]
			}

			if w.dcpTurnProtEnabled || w.dcpTurnProtTurns.Value() != "" || w.dcpProtectedTools.Value() != "" {
				dcp.TurnProtection = &config.TurnProtectionConfig{}
				if w.dcpTurnProtEnabled {
					dcp.TurnProtection.Enabled = &w.dcpTurnProtEnabled
				}
				if v := w.dcpTurnProtTurns.Value(); v != "" {
					if i, err := strconv.Atoi(v); err == nil {
						dcp.TurnProtection.Turns = &i
					}
				}
				if v := w.dcpProtectedTools.Value(); v != "" {
					var tools []string
					for _, t := range strings.Split(v, ",") {
						if s := strings.TrimSpace(t); s != "" {
							tools = append(tools, s)
						}
					}
					if len(tools) > 0 {
						dcp.ProtectedTools = tools
					}
				}
			}

			if w.dcpDeduplicationEnabled || w.dcpSupersedeWritesEnabled || w.dcpSupersedeWritesAggressive || w.dcpPurgeErrorsEnabled || w.dcpPurgeErrorsTurns.Value() != "" {
				dcp.Strategies = &config.StrategiesConfig{}
				if w.dcpDeduplicationEnabled {
					dcp.Strategies.Deduplication = &config.DeduplicationConfig{Enabled: &w.dcpDeduplicationEnabled}
				}
				if w.dcpSupersedeWritesEnabled || w.dcpSupersedeWritesAggressive {
					dcp.Strategies.SupersedeWrites = &config.SupersedeWritesConfig{}
					if w.dcpSupersedeWritesEnabled {
						dcp.Strategies.SupersedeWrites.Enabled = &w.dcpSupersedeWritesEnabled
					}
					if w.dcpSupersedeWritesAggressive {
						dcp.Strategies.SupersedeWrites.Aggressive = &w.dcpSupersedeWritesAggressive
					}
				}
				if w.dcpPurgeErrorsEnabled || w.dcpPurgeErrorsTurns.Value() != "" {
					dcp.Strategies.PurgeErrors = &config.PurgeErrorsConfig{}
					if w.dcpPurgeErrorsEnabled {
						dcp.Strategies.PurgeErrors.Enabled = &w.dcpPurgeErrorsEnabled
					}
					if v := w.dcpPurgeErrorsTurns.Value(); v != "" {
						if i, err := strconv.Atoi(v); err == nil {
							dcp.Strategies.PurgeErrors.Turns = &i
						}
					}
				}
			}
		}
	}

	// Claude Code - only set if any flag is true or plugins_override has value
	ccHasData := w.ccMcp || w.ccCommands || w.ccSkills || w.ccAgents || w.ccHooks || w.ccPlugins ||
		w.ccPluginsOverride.Value() != ""
	if ccHasData {
		cfg.ClaudeCode = &config.ClaudeCodeConfig{}
		if w.ccMcp {
			cfg.ClaudeCode.MCP = &w.ccMcp
		}
		if w.ccCommands {
			cfg.ClaudeCode.Commands = &w.ccCommands
		}
		if w.ccSkills {
			cfg.ClaudeCode.Skills = &w.ccSkills
		}
		if w.ccAgents {
			cfg.ClaudeCode.Agents = &w.ccAgents
		}
		if w.ccHooks {
			cfg.ClaudeCode.Hooks = &w.ccHooks
		}
		if w.ccPlugins {
			cfg.ClaudeCode.Plugins = &w.ccPlugins
		}
		if v := w.ccPluginsOverride.Value(); v != "" {
			cfg.ClaudeCode.PluginsOverride = parseMapStringBool(v)
		}
	}

	// Sisyphus Agent - only set if any flag is true
	if w.saDisabled || w.saDefaultBuilderEnabled || w.saPlannerEnabled || w.saReplacePlan {
		cfg.SisyphusAgent = &config.SisyphusAgentConfig{}
		if w.saDisabled {
			cfg.SisyphusAgent.Disabled = &w.saDisabled
		}
		if w.saDefaultBuilderEnabled {
			cfg.SisyphusAgent.DefaultBuilderEnabled = &w.saDefaultBuilderEnabled
		}
		if w.saPlannerEnabled {
			cfg.SisyphusAgent.PlannerEnabled = &w.saPlannerEnabled
		}
		if w.saReplacePlan {
			cfg.SisyphusAgent.ReplacePlan = &w.saReplacePlan
		}
	}

	// Ralph Loop
	if w.rlEnabled || w.rlDefaultMaxIterations.Value() != "" || w.rlStateDir.Value() != "" {
		cfg.RalphLoop = &config.RalphLoopConfig{}
		if w.rlEnabled {
			cfg.RalphLoop.Enabled = &w.rlEnabled
		}
		if v := w.rlDefaultMaxIterations.Value(); v != "" {
			var i int
			fmt.Sscanf(v, "%d", &i)
			if i > 0 {
				cfg.RalphLoop.DefaultMaxIterations = &i
			}
		}
		if v := w.rlStateDir.Value(); v != "" {
			cfg.RalphLoop.StateDir = v
		}
	}

	// Background Task
	btHasData := w.btDefaultConcurrency.Value() != "" ||
		w.btProviderConcurrency.Value() != "" ||
		w.btModelConcurrency.Value() != ""
	if btHasData {
		cfg.BackgroundTask = &config.BackgroundTaskConfig{}
		if v := w.btDefaultConcurrency.Value(); v != "" {
			var i int
			fmt.Sscanf(v, "%d", &i)
			if i > 0 {
				cfg.BackgroundTask.DefaultConcurrency = &i
			}
		}
		if v := w.btProviderConcurrency.Value(); v != "" {
			cfg.BackgroundTask.ProviderConcurrency = parseMapStringInt(v)
		}
		if v := w.btModelConcurrency.Value(); v != "" {
			cfg.BackgroundTask.ModelConcurrency = parseMapStringInt(v)
		}
	}

	// Notification
	if w.notifForceEnable {
		cfg.Notification = &config.NotificationConfig{
			ForceEnable: &w.notifForceEnable,
		}
	}

	// Git Master
	if w.gmCommitFooter || w.gmIncludeCoAuthoredBy {
		cfg.GitMaster = &config.GitMasterConfig{}
		if w.gmCommitFooter {
			cfg.GitMaster.CommitFooter = &w.gmCommitFooter
		}
		if w.gmIncludeCoAuthoredBy {
			cfg.GitMaster.IncludeCoAuthoredBy = &w.gmIncludeCoAuthoredBy
		}
	}

	// Comment Checker
	if v := w.ccCustomPrompt.Value(); v != "" {
		cfg.CommentChecker = &config.CommentCheckerConfig{
			CustomPrompt: v,
		}
	}

	// Skills JSON
	if v := w.skillsEditor.Value(); strings.TrimSpace(v) != "" {
		cfg.Skills = json.RawMessage(v)
	} else {
		cfg.Skills = nil
	}
}

func (w WizardOther) Update(msg tea.Msg) (WizardOther, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.SetSize(msg.Width, msg.Height)
		return w, nil

	case tea.KeyMsg:
		if w.inSubSection {
			if w.currentSection == sectionExperimental && w.subCursor == 5 {
				switch msg.String() {
				case "esc":
					w.expThreshold.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "up", "k":
					w.expThreshold.Blur()
					if w.subCursor > 0 {
						w.subCursor--
					}
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "down", "j":
					w.expThreshold.Blur()
					w.subCursor++
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "tab":
					w.expThreshold.Blur()
					w.inSubSection = false
					w.viewport.SetContent(w.renderContent())
					return w, nil
				case "enter", " ":
					w.expThreshold.Focus()
					w.expThreshold, cmd = w.expThreshold.Update(msg)
					return w, cmd
				default:
					w.expThreshold.Focus()
					w.expThreshold, cmd = w.expThreshold.Update(msg)
					return w, cmd
				}
			}

			if w.currentSection == sectionExperimental && w.subCursor == 9 {
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

			if w.currentSection == sectionExperimental && w.subCursor == 10 {
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

			if w.currentSection == sectionExperimental && w.subCursor == 15 {
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

			if w.currentSection == sectionExperimental && w.subCursor == 7 {
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

			if w.currentSection == sectionSkillsJson {
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
			if w.currentSection > 0 {
				w.currentSection--
			}
		case key.Matches(msg, w.keys.Down):
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
			}
		case key.Matches(msg, w.keys.Right):
			if !w.inSubSection && !w.sectionExpanded[w.currentSection] {
				w.sectionExpanded[w.currentSection] = true
				w.inSubSection = true
				w.subCursor = 0
			}
		case key.Matches(msg, w.keys.Left):
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
		w.autoUpdate = !w.autoUpdate
	}
}

func (w *WizardOther) toggleSubItem() {
	switch w.currentSection {
	case sectionDisabledAgents:
		if w.subCursor < len(disableableAgents) {
			agent := disableableAgents[w.subCursor]
			w.disabledAgents[agent] = !w.disabledAgents[agent]
		}
	case sectionDisabledSkills:
		if w.subCursor < len(disableableSkills) {
			skill := disableableSkills[w.subCursor]
			w.disabledSkills[skill] = !w.disabledSkills[skill]
		}
	case sectionDisabledCommands:
		if w.subCursor < len(disableableCommands) {
			cmd := disableableCommands[w.subCursor]
			w.disabledCommands[cmd] = !w.disabledCommands[cmd]
		}
	case sectionExperimental:
		switch w.subCursor {
		case 0:
			w.expAggressiveTrunc = !w.expAggressiveTrunc
		case 1:
			w.expAutoResume = !w.expAutoResume
		case 2:
			w.expPreemptiveCompact = !w.expPreemptiveCompact
		case 3:
			w.expTruncateAllOutputs = !w.expTruncateAllOutputs
		case 4:
			w.expDcpForCompaction = !w.expDcpForCompaction
		case 5:
			// threshold textinput - handled in Update() via focus, not toggle
		case 6:
			w.dcpEnabled = !w.dcpEnabled
		case 7:
		case 8:
			w.dcpTurnProtEnabled = !w.dcpTurnProtEnabled
		case 9:
		case 10:
		case 11:
			w.dcpDeduplicationEnabled = !w.dcpDeduplicationEnabled
		case 12:
			w.dcpSupersedeWritesEnabled = !w.dcpSupersedeWritesEnabled
		case 13:
			w.dcpSupersedeWritesAggressive = !w.dcpSupersedeWritesAggressive
		case 14:
			w.dcpPurgeErrorsEnabled = !w.dcpPurgeErrorsEnabled
		case 15:
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
		case 6:
			// plugins_override textinput - handled in Update()
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
		case 1:
			w.gmIncludeCoAuthoredBy = !w.gmIncludeCoAuthoredBy
		}
	}
}

func (w WizardOther) renderContent() string {
	var lines []string

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))

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

		// Simple sections without expansion
		if section == sectionAutoUpdate {
			checkbox := "[ ]"
			if w.autoUpdate {
				checkbox = enabledStyle.Render("[✓]")
			}
			line := fmt.Sprintf("%s%s %s", cursor, checkbox, labelStyle.Render(name))
			lines = append(lines, line)
			continue
		}

		line := fmt.Sprintf("%s%s %s", cursor, expandIcon, labelStyle.Render(name))
		lines = append(lines, line)

		// Render expanded content
		if w.sectionExpanded[section] {
			subLines := w.renderSubSection(section)
			lines = append(lines, subLines...)
		}
	}

	_ = dimStyle
	_ = disabledStyle
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (w WizardOther) renderSubSection(section otherSection) []string {
	var lines []string
	indent := "      "

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	renderCheckbox := func(idx int, label string, checked bool) string {
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == idx {
			cursor = selectedStyle.Render("> ")
		}

		checkbox := "[ ]"
		if checked {
			checkbox = enabledStyle.Render("[✓]")
		}

		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == idx {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}

		return indent + cursor + checkbox + " " + style.Render(label)
	}

	_ = disabledStyle

	switch section {
	case sectionDisabledMcps:
		lines = append(lines, indent+"  "+w.disabledMcps.View())

	case sectionDisabledAgents:
		for i, agent := range disableableAgents {
			lines = append(lines, renderCheckbox(i, agent, w.disabledAgents[agent]))
		}

	case sectionDisabledSkills:
		for i, skill := range disableableSkills {
			lines = append(lines, renderCheckbox(i, skill, w.disabledSkills[skill]))
		}

	case sectionDisabledCommands:
		for i, cmd := range disableableCommands {
			lines = append(lines, renderCheckbox(i, cmd, w.disabledCommands[cmd]))
		}

	case sectionExperimental:
		lines = append(lines, renderCheckbox(0, "aggressive_truncation", w.expAggressiveTrunc))
		lines = append(lines, renderCheckbox(1, "auto_resume", w.expAutoResume))
		lines = append(lines, renderCheckbox(2, "preemptive_compaction", w.expPreemptiveCompact))
		lines = append(lines, renderCheckbox(3, "truncate_all_tool_outputs", w.expTruncateAllOutputs))
		lines = append(lines, renderCheckbox(4, "dcp_for_compaction", w.expDcpForCompaction))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("preemptive_compaction_threshold: ")+w.expThreshold.View())

		lines = append(lines, renderCheckbox(6, "dcp_enabled", w.dcpEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("dcp_notification: ")+dcpNotificationValues[w.dcpNotificationIdx])

		lines = append(lines, renderCheckbox(8, "dcp_turn_protection_enabled", w.dcpTurnProtEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 9 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 9 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("dcp_turn_protection_turns: ")+w.dcpTurnProtTurns.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 10 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 10 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("dcp_protected_tools: ")+w.dcpProtectedTools.View())

		lines = append(lines, renderCheckbox(11, "dcp_deduplication_enabled", w.dcpDeduplicationEnabled))
		lines = append(lines, renderCheckbox(12, "dcp_supersede_writes_enabled", w.dcpSupersedeWritesEnabled))
		lines = append(lines, renderCheckbox(13, "dcp_supersede_writes_aggressive", w.dcpSupersedeWritesAggressive))
		lines = append(lines, renderCheckbox(14, "dcp_purge_errors_enabled", w.dcpPurgeErrorsEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 15 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 15 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("dcp_purge_errors_turns: ")+w.dcpPurgeErrorsTurns.View())

	case sectionClaudeCode:
		lines = append(lines, renderCheckbox(0, "mcp", w.ccMcp))
		lines = append(lines, renderCheckbox(1, "commands", w.ccCommands))
		lines = append(lines, renderCheckbox(2, "skills", w.ccSkills))
		lines = append(lines, renderCheckbox(3, "agents", w.ccAgents))
		lines = append(lines, renderCheckbox(4, "hooks", w.ccHooks))
		lines = append(lines, renderCheckbox(5, "plugins", w.ccPlugins))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("plugins_override: ")+w.ccPluginsOverride.View())

	case sectionSisyphusAgent:
		lines = append(lines, renderCheckbox(0, "disabled", w.saDisabled))
		lines = append(lines, renderCheckbox(1, "default_builder_enabled", w.saDefaultBuilderEnabled))
		lines = append(lines, renderCheckbox(2, "planner_enabled", w.saPlannerEnabled))
		lines = append(lines, renderCheckbox(3, "replace_plan", w.saReplacePlan))

	case sectionRalphLoop:
		lines = append(lines, renderCheckbox(0, "enabled", w.rlEnabled))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("max_iterations: ")+w.rlDefaultMaxIterations.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("state_dir: ")+w.rlStateDir.View())

	case sectionBackgroundTask:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("provider_concurrency: ")+w.btProviderConcurrency.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("model_concurrency: ")+w.btModelConcurrency.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
		}
		lines = append(lines, indent+cursor+style.Render("default_concurrency: ")+w.btDefaultConcurrency.View())

	case sectionNotification:
		lines = append(lines, renderCheckbox(0, "force_enable", w.notifForceEnable))

	case sectionGitMaster:
		lines = append(lines, renderCheckbox(0, "commit_footer", w.gmCommitFooter))
		lines = append(lines, renderCheckbox(1, "include_co_authored_by", w.gmIncludeCoAuthoredBy))

	case sectionCommentChecker:
		lines = append(lines, indent+"  custom_prompt: "+w.ccCustomPrompt.View())

	case sectionSkillsJson:
		lines = append(lines, indent+w.skillsEditor.View())
	}

	lines = append(lines, "")
	return lines
}

func (w WizardOther) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CDD6F4"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	title := titleStyle.Render("Other Settings")
	desc := helpStyle.Render("Enter to expand • Space to toggle • Tab next • Shift+Tab back")

	if w.inSubSection {
		desc = helpStyle.Render("Space/Enter to toggle • Esc to close section")
	}

	content := w.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
		"",
		content,
	)
}
