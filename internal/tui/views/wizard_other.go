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

	// Disabled Tools
	if v := w.disabledTools.Value(); v != "" {
		var tools []string
		for _, t := range strings.Split(v, ",") {
			if s := strings.TrimSpace(t); s != "" {
				tools = append(tools, s)
			}
		}
		cfg.DisabledTools = tools
	}
	if len(cfg.DisabledTools) == 0 {
		cfg.DisabledTools = nil
	}

	// Auto update
	if w.autoUpdate {
		cfg.AutoUpdate = &w.autoUpdate
	}

	if w.hashlineEdit {
		cfg.HashlineEdit = &w.hashlineEdit
	}

	if w.modelFallback {
		cfg.ModelFallback = &w.modelFallback
	}

	if w.startWorkAutoCommit {
		cfg.StartWork = &config.StartWorkConfig{AutoCommit: &w.startWorkAutoCommit}
	}

	// Experimental - only set if any flag is true or DCP has value
	expHasData := w.expAggressiveTrunc || w.expAutoResume ||
		w.expTruncateAllOutputs || w.expPreemptiveCompaction || w.expTaskSystem ||
		w.expPluginLoadTimeoutMs.Value() != "" || w.expSafeHookCreation || w.expHashlineEdit ||
		w.expDisableOmoEnv || w.expModelFallbackTitle || w.expMaxTools.Value() != "" ||
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
		if w.expTruncateAllOutputs {
			cfg.Experimental.TruncateAllToolOutputs = &w.expTruncateAllOutputs
		}
		if w.expPreemptiveCompaction {
			cfg.Experimental.PreemptiveCompaction = &w.expPreemptiveCompaction
		}
		if w.expTaskSystem {
			cfg.Experimental.TaskSystem = &w.expTaskSystem
		}
		if v := w.expPluginLoadTimeoutMs.Value(); v != "" {
			if i, err := strconv.Atoi(v); err == nil && i > 0 {
				cfg.Experimental.PluginLoadTimeoutMs = &i
			}
		}
		if w.expSafeHookCreation {
			cfg.Experimental.SafeHookCreation = &w.expSafeHookCreation
		}
		if w.expHashlineEdit {
			cfg.Experimental.HashlineEdit = &w.expHashlineEdit
		}
		if w.expDisableOmoEnv {
			cfg.Experimental.DisableOmoEnv = &w.expDisableOmoEnv
		}
		if w.expModelFallbackTitle {
			cfg.Experimental.ModelFallbackTitle = &w.expModelFallbackTitle
		}
		if v := parsePositiveInt64(w.expMaxTools.Value()); v != nil {
			cfg.Experimental.MaxTools = v
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
	if w.saDisabled || w.saDefaultBuilderEnabled || w.saPlannerEnabled || w.saReplacePlan || w.saTDD {
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
		if w.saTDD {
			cfg.SisyphusAgent.TDD = &w.saTDD
		}
	}

	// Ralph Loop
	if w.rlEnabled || w.rlDefaultMaxIterations.Value() != "" || w.rlStateDir.Value() != "" || w.rlDefaultStrategyIdx > 0 {
		cfg.RalphLoop = &config.RalphLoopConfig{}
		if w.rlEnabled {
			cfg.RalphLoop.Enabled = &w.rlEnabled
		}
		if v := w.rlDefaultMaxIterations.Value(); v != "" {
			var i int
			_, _ = fmt.Sscanf(v, "%d", &i)
			if i > 0 {
				cfg.RalphLoop.DefaultMaxIterations = &i
			}
		}
		if v := w.rlStateDir.Value(); v != "" {
			cfg.RalphLoop.StateDir = v
		}
		if w.rlDefaultStrategyIdx > 0 {
			cfg.RalphLoop.DefaultStrategy = ralphLoopStrategies[w.rlDefaultStrategyIdx]
		}
	}

	// Background Task
	btHasData := w.btDefaultConcurrency.Value() != "" ||
		w.btProviderConcurrency.Value() != "" ||
		w.btModelConcurrency.Value() != "" ||
		w.btMaxDepth.Value() != "" ||
		w.btMaxDescendants.Value() != "" ||
		w.btStaleTimeoutMs.Value() != "" ||
		w.btMessageStalenessTimeoutMs.Value() != "" ||
		w.btTaskTtlMs.Value() != "" ||
		w.btSessionGoneTimeoutMs.Value() != "" ||
		w.btSyncPollTimeoutMs.Value() != "" ||
		w.btMaxToolCalls.Value() != "" ||
		w.btCircuitBreakerEnabled ||
		w.btCircuitBreakerMaxCalls.Value() != "" ||
		w.btCircuitBreakerConsecutive.Value() != ""
	if btHasData {
		cfg.BackgroundTask = &config.BackgroundTaskConfig{}
		if v := w.btDefaultConcurrency.Value(); v != "" {
			var i int
			_, _ = fmt.Sscanf(v, "%d", &i)
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
		if v := parsePositiveInt64(w.btMaxDepth.Value()); v != nil {
			cfg.BackgroundTask.MaxDepth = v
		}
		if v := parsePositiveInt64(w.btMaxDescendants.Value()); v != nil {
			cfg.BackgroundTask.MaxDescendants = v
		}
		if v := parsePositiveIntWithMinimum(w.btStaleTimeoutMs.Value(), 60000); v != nil {
			cfg.BackgroundTask.StaleTimeoutMs = v
		}
		if v := parsePositiveIntWithMinimum(w.btMessageStalenessTimeoutMs.Value(), 60000); v != nil {
			cfg.BackgroundTask.MessageStalenessTimeoutMs = v
		}
		if v := parsePositiveIntWithMinimum(w.btTaskTtlMs.Value(), 300000); v != nil {
			cfg.BackgroundTask.TaskTtlMs = v
		}
		if v := parsePositiveIntWithMinimum(w.btSessionGoneTimeoutMs.Value(), 10000); v != nil {
			cfg.BackgroundTask.SessionGoneTimeoutMs = v
		}
		if v := parsePositiveIntWithMinimum(w.btSyncPollTimeoutMs.Value(), 60000); v != nil {
			cfg.BackgroundTask.SyncPollTimeoutMs = v
		}
		if v := parsePositiveInt64(w.btMaxToolCalls.Value()); v != nil {
			cfg.BackgroundTask.MaxToolCalls = v
		}
		cbHasData := w.btCircuitBreakerEnabled ||
			w.btCircuitBreakerMaxCalls.Value() != "" ||
			w.btCircuitBreakerConsecutive.Value() != ""
		if cbHasData {
			cfg.BackgroundTask.CircuitBreaker = &config.BackgroundCircuitBreaker{}
			if w.btCircuitBreakerEnabled {
				cfg.BackgroundTask.CircuitBreaker.Enabled = &w.btCircuitBreakerEnabled
			}
			if v := parsePositiveInt64(w.btCircuitBreakerMaxCalls.Value()); v != nil {
				cfg.BackgroundTask.CircuitBreaker.MaxToolCalls = v
			}
			if v := parsePositiveInt64(w.btCircuitBreakerConsecutive.Value()); v != nil {
				cfg.BackgroundTask.CircuitBreaker.ConsecutiveThreshold = v
			}
		}
	}

	// Notification
	if w.notifForceEnable {
		cfg.Notification = &config.NotificationConfig{
			ForceEnable: &w.notifForceEnable,
		}
	}

	// Git Master
	gmCommitFooterText := strings.TrimSpace(w.gmCommitFooterText.Value())
	gmGitEnvPrefix := strings.TrimSpace(w.gmGitEnvPrefix.Value())
	if w.gmCommitFooter || gmCommitFooterText != "" || w.gmIncludeCoAuthoredBy || gmGitEnvPrefix != "" {
		cfg.GitMaster = &config.GitMasterConfig{}
		if gmCommitFooterText != "" {
			cfg.GitMaster.CommitFooter = gmCommitFooterText
		} else if w.gmCommitFooter {
			cfg.GitMaster.CommitFooter = w.gmCommitFooter
		}
		if w.gmIncludeCoAuthoredBy {
			cfg.GitMaster.IncludeCoAuthoredBy = &w.gmIncludeCoAuthoredBy
		}
		if gmGitEnvPrefix != "" {
			cfg.GitMaster.GitEnvPrefix = gmGitEnvPrefix
		}
	}

	// Comment Checker
	if v := w.ccCustomPrompt.Value(); v != "" {
		cfg.CommentChecker = &config.CommentCheckerConfig{
			CustomPrompt: v,
		}
	}

	// Babysitting
	if v := w.babysittingTimeoutMs.Value(); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.Babysitting = &config.BabysittingConfig{TimeoutMs: &f}
		}
	}

	// Browser Automation Engine
	if w.browserProviderIdx > 0 {
		cfg.BrowserAutomationEngine = &config.BrowserAutomationEngineConfig{
			Provider: browserProviders[w.browserProviderIdx],
		}
	}

	// Tmux
	tmuxHasData := w.tmuxEnabled || w.tmuxLayoutIdx > 0 ||
		w.tmuxMainPaneSize.Value() != "" || w.tmuxMainPaneMinWidth.Value() != "" || w.tmuxAgentPaneMinWidth.Value() != "" ||
		w.tmuxIsolationIdx > 0
	if tmuxHasData {
		cfg.Tmux = &config.TmuxConfig{}
		if w.tmuxEnabled {
			cfg.Tmux.Enabled = &w.tmuxEnabled
		}
		if w.tmuxLayoutIdx > 0 {
			cfg.Tmux.Layout = tmuxLayouts[w.tmuxLayoutIdx]
		}
		if v := w.tmuxMainPaneSize.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
				cfg.Tmux.MainPaneSize = &f
			}
		}
		if v := w.tmuxMainPaneMinWidth.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
				cfg.Tmux.MainPaneMinWidth = &f
			}
		}
		if v := w.tmuxAgentPaneMinWidth.Value(); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
				cfg.Tmux.AgentPaneMinWidth = &f
			}
		}
		if w.tmuxIsolationIdx > 0 {
			cfg.Tmux.Isolation = tmuxIsolations[w.tmuxIsolationIdx]
		}
	}

	// Websearch
	if w.websearchProviderIdx > 0 {
		cfg.Websearch = &config.WebsearchConfig{
			Provider: websearchProviders[w.websearchProviderIdx],
		}
	}

	// Sisyphus
	sisyphusHasData := w.sisyphusTasksStoragePath.Value() != "" ||
		w.sisyphusTasksTaskListID.Value() != "" || w.sisyphusTasksClaudeCodeCompat
	if sisyphusHasData {
		cfg.Sisyphus = &config.SisyphusConfig{Tasks: &config.SisyphusTasksConfig{}}
		if v := w.sisyphusTasksStoragePath.Value(); v != "" {
			cfg.Sisyphus.Tasks.StoragePath = v
		}
		if v := w.sisyphusTasksTaskListID.Value(); v != "" {
			cfg.Sisyphus.Tasks.TaskListID = v
		}
		if w.sisyphusTasksClaudeCodeCompat {
			cfg.Sisyphus.Tasks.ClaudeCodeCompat = &w.sisyphusTasksClaudeCodeCompat
		}
	}

	// New Task System Enabled
	if w.newTaskSystemEnabled {
		cfg.NewTaskSystemEnabled = &w.newTaskSystemEnabled
	}

	// Default Run Agent
	cfg.DefaultRunAgent = strings.TrimSpace(w.defaultRunAgent.Value())

	modelCapabilitiesHasData := w.mcEnabled || w.mcAutoRefreshOnStart ||
		strings.TrimSpace(w.mcRefreshTimeoutMs.Value()) != "" || strings.TrimSpace(w.mcSourceURL.Value()) != ""
	if modelCapabilitiesHasData {
		cfg.ModelCapabilities = &config.ModelCapabilitiesConfig{}
		if w.mcEnabled {
			cfg.ModelCapabilities.Enabled = &w.mcEnabled
		}
		if w.mcAutoRefreshOnStart {
			cfg.ModelCapabilities.AutoRefreshOnStart = &w.mcAutoRefreshOnStart
		}
		if v := parsePositiveInt64(w.mcRefreshTimeoutMs.Value()); v != nil {
			cfg.ModelCapabilities.RefreshTimeoutMs = v
		}
		if v := strings.TrimSpace(w.mcSourceURL.Value()); v != "" {
			cfg.ModelCapabilities.SourceURL = v
		}
	}

	if v := strings.TrimSpace(w.openclawEditor.Value()); v != "" {
		var parsed config.OpenclawConfig
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			cfg.Openclaw = &parsed
		}
	} else {
		cfg.Openclaw = nil
	}

	if v := w.runtimeFallbackEditor.Value(); strings.TrimSpace(v) != "" {
		cfg.RuntimeFallback = json.RawMessage(v)
	} else {
		cfg.RuntimeFallback = nil
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

			if w.currentSection == sectionOpenclaw {
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

			if w.currentSection == sectionRuntimeFallback {
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
	case sectionNewTaskSystemEnabled:
		w.newTaskSystemEnabled = !w.newTaskSystemEnabled
	case sectionHashlineEdit:
		w.hashlineEdit = !w.hashlineEdit
	case sectionModelFallback:
		w.modelFallback = !w.modelFallback
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
			w.expTruncateAllOutputs = !w.expTruncateAllOutputs
		case 3:
			w.dcpEnabled = !w.dcpEnabled
		case 4:
		case 5:
			w.dcpTurnProtEnabled = !w.dcpTurnProtEnabled
		case 6:
		case 7:
		case 8:
			w.dcpDeduplicationEnabled = !w.dcpDeduplicationEnabled
		case 9:
			w.dcpSupersedeWritesEnabled = !w.dcpSupersedeWritesEnabled
		case 10:
			w.dcpSupersedeWritesAggressive = !w.dcpSupersedeWritesAggressive
		case 11:
			w.dcpPurgeErrorsEnabled = !w.dcpPurgeErrorsEnabled
		case 12:
		case 13:
			w.expPreemptiveCompaction = !w.expPreemptiveCompaction
		case 14:
			w.expTaskSystem = !w.expTaskSystem
		case 15:
			// plugin_load_timeout_ms textinput - handled in Update()
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
		case 1:
			// commit_footer text input handled in Update()
		case 2:
			w.gmIncludeCoAuthoredBy = !w.gmIncludeCoAuthoredBy
		case 3:
			// git_env_prefix text input handled in Update()
		}
	case sectionBrowserAutomationEngine:
		if w.subCursor == 0 {
			w.browserProviderIdx = (w.browserProviderIdx + 1) % len(browserProviders)
		}
	case sectionTmux:
		switch w.subCursor {
		case 0:
			w.tmuxEnabled = !w.tmuxEnabled
		case 1:
			w.tmuxLayoutIdx = (w.tmuxLayoutIdx + 1) % len(tmuxLayouts)
		case 5:
			w.tmuxIsolationIdx = (w.tmuxIsolationIdx + 1) % len(tmuxIsolations)
		}
	case sectionWebsearch:
		if w.subCursor == 0 {
			w.websearchProviderIdx = (w.websearchProviderIdx + 1) % len(websearchProviders)
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
	}
}

func (w WizardOther) renderContent() string {
	var lines []string

	selectedStyle := wizOtherSelectedStyle
	enabledStyle := wizOtherEnabledStyle
	disabledStyle := wizOtherDisabledStyle
	dimStyle := wizOtherDimStyle
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

		// Simple sections without expansion
		if section == sectionAutoUpdate || section == sectionNewTaskSystemEnabled || section == sectionHashlineEdit || section == sectionModelFallback {
			checkbox := "[ ]"
			checked := false
			switch section {
			case sectionAutoUpdate:
				checked = w.autoUpdate
			case sectionNewTaskSystemEnabled:
				checked = w.newTaskSystemEnabled
			case sectionHashlineEdit:
				checked = w.hashlineEdit
			case sectionModelFallback:
				checked = w.modelFallback
			}
			if checked {
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

	selectedStyle := wizOtherSelectedStyle
	enabledStyle := wizOtherEnabledStyle
	disabledStyle := wizOtherDisabledStyle
	dimStyle := wizOtherDimStyle

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
			style = wizOtherLabelStyle
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

	case sectionDisabledTools:
		lines = append(lines, indent+"  "+w.disabledTools.View())

	case sectionExperimental:
		lines = append(lines, renderCheckbox(0, "aggressive_truncation", w.expAggressiveTrunc))
		lines = append(lines, renderCheckbox(1, "auto_resume", w.expAutoResume))
		lines = append(lines, renderCheckbox(2, "truncate_all_tool_outputs", w.expTruncateAllOutputs))

		lines = append(lines, renderCheckbox(3, "dcp_enabled", w.dcpEnabled))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("dcp_notification: ")+dcpNotificationValues[w.dcpNotificationIdx])

		lines = append(lines, renderCheckbox(5, "dcp_turn_protection_enabled", w.dcpTurnProtEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("dcp_turn_protection_turns: ")+w.dcpTurnProtTurns.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("dcp_protected_tools: ")+w.dcpProtectedTools.View())

		lines = append(lines, renderCheckbox(8, "dcp_deduplication_enabled", w.dcpDeduplicationEnabled))
		lines = append(lines, renderCheckbox(9, "dcp_supersede_writes_enabled", w.dcpSupersedeWritesEnabled))
		lines = append(lines, renderCheckbox(10, "dcp_supersede_writes_aggressive", w.dcpSupersedeWritesAggressive))
		lines = append(lines, renderCheckbox(11, "dcp_purge_errors_enabled", w.dcpPurgeErrorsEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 12 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 12 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("dcp_purge_errors_turns: ")+w.dcpPurgeErrorsTurns.View())

		lines = append(lines, renderCheckbox(13, "preemptive_compaction", w.expPreemptiveCompaction))
		lines = append(lines, renderCheckbox(14, "task_system", w.expTaskSystem))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 15 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 15 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("plugin_load_timeout_ms: ")+w.expPluginLoadTimeoutMs.View())

		lines = append(lines, renderCheckbox(16, "safe_hook_creation", w.expSafeHookCreation))
		lines = append(lines, renderCheckbox(17, "hashline_edit", w.expHashlineEdit))
		lines = append(lines, renderCheckbox(18, "disable_omo_env", w.expDisableOmoEnv))
		lines = append(lines, renderCheckbox(19, "model_fallback_title", w.expModelFallbackTitle))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 20 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 20 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("max_tools: ")+w.expMaxTools.View())

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
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("plugins_override: ")+w.ccPluginsOverride.View())

	case sectionSisyphusAgent:
		lines = append(lines, renderCheckbox(0, "disabled", w.saDisabled))
		lines = append(lines, renderCheckbox(1, "default_builder_enabled", w.saDefaultBuilderEnabled))
		lines = append(lines, renderCheckbox(2, "planner_enabled", w.saPlannerEnabled))
		lines = append(lines, renderCheckbox(3, "replace_plan", w.saReplacePlan))
		lines = append(lines, renderCheckbox(4, "tdd", w.saTDD))

	case sectionRalphLoop:
		lines = append(lines, renderCheckbox(0, "enabled", w.rlEnabled))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("max_iterations: ")+w.rlDefaultMaxIterations.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("state_dir: ")+w.rlStateDir.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("default_strategy: ")+ralphLoopStrategies[w.rlDefaultStrategyIdx])

	case sectionBackgroundTask:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("provider_concurrency: ")+w.btProviderConcurrency.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("model_concurrency: ")+w.btModelConcurrency.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("default_concurrency: ")+w.btDefaultConcurrency.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("stale_timeout_ms: ")+w.btStaleTimeoutMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("message_staleness_timeout_ms: ")+w.btMessageStalenessTimeoutMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("sync_poll_timeout_ms: ")+w.btSyncPollTimeoutMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 6 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("max_depth: ")+w.btMaxDepth.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 7 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("max_descendants: ")+w.btMaxDescendants.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 8 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 8 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("task_ttl_ms: ")+w.btTaskTtlMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 9 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 9 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("session_gone_timeout_ms: ")+w.btSessionGoneTimeoutMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 10 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 10 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("max_tool_calls: ")+w.btMaxToolCalls.View())

		lines = append(lines, renderCheckbox(11, "circuit_breaker.enabled", w.btCircuitBreakerEnabled))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 12 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 12 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("circuit_breaker.max_tool_calls: ")+w.btCircuitBreakerMaxCalls.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 13 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 13 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("circuit_breaker.consecutive_threshold: ")+w.btCircuitBreakerConsecutive.View())

	case sectionNotification:
		lines = append(lines, renderCheckbox(0, "force_enable", w.notifForceEnable))

	case sectionGitMaster:
		lines = append(lines, renderCheckbox(0, "commit_footer", w.gmCommitFooter))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("commit_footer_text: ")+w.gmCommitFooterText.View())

		lines = append(lines, renderCheckbox(2, "include_co_authored_by", w.gmIncludeCoAuthoredBy))

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("git_env_prefix: ")+w.gmGitEnvPrefix.View())

	case sectionCommentChecker:
		lines = append(lines, indent+"  custom_prompt: "+w.ccCustomPrompt.View())

	case sectionBabysitting:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("timeout_ms: ")+w.babysittingTimeoutMs.View())

	case sectionBrowserAutomationEngine:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("provider: ")+browserProviders[w.browserProviderIdx])

	case sectionTmux:
		lines = append(lines, renderCheckbox(0, "enabled", w.tmuxEnabled))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("layout: ")+tmuxLayouts[w.tmuxLayoutIdx])

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("main_pane_size: ")+w.tmuxMainPaneSize.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("main_pane_min_width: ")+w.tmuxMainPaneMinWidth.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 4 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("agent_pane_min_width: ")+w.tmuxAgentPaneMinWidth.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 5 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("isolation: ")+tmuxIsolations[w.tmuxIsolationIdx])

	case sectionWebsearch:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("provider: ")+websearchProviders[w.websearchProviderIdx])

	case sectionSisyphus:
		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 0 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("tasks.storage_path: ")+w.sisyphusTasksStoragePath.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 1 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("tasks.task_list_id: ")+w.sisyphusTasksTaskListID.View())

		lines = append(lines, renderCheckbox(2, "tasks.claude_code_compat", w.sisyphusTasksClaudeCodeCompat))

	case sectionDefaultRunAgent:
		lines = append(lines, indent+"  value: "+w.defaultRunAgent.View())

	case sectionStartWork:
		lines = append(lines, renderCheckbox(0, "auto_commit", w.startWorkAutoCommit))

	case sectionModelCapabilities:
		lines = append(lines, renderCheckbox(0, "enabled", w.mcEnabled))
		lines = append(lines, renderCheckbox(1, "auto_refresh_on_start", w.mcAutoRefreshOnStart))

		cursor := "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			cursor = selectedStyle.Render("> ")
		}
		style := dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 2 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("refresh_timeout_ms: ")+w.mcRefreshTimeoutMs.View())

		cursor = "  "
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			cursor = selectedStyle.Render("> ")
		}
		style = dimStyle
		if w.inSubSection && w.currentSection == section && w.subCursor == 3 {
			style = wizOtherLabelStyle
		}
		lines = append(lines, indent+cursor+style.Render("source_url: ")+w.mcSourceURL.View())

	case sectionOpenclaw:
		lines = append(lines, indent+w.openclawEditor.View())

	case sectionRuntimeFallback:
		lines = append(lines, indent+w.runtimeFallbackEditor.View())

	case sectionSkillsJson:
		lines = append(lines, indent+w.skillsEditor.View())
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
