package views

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

var (
	wizOtherPurple = lipgloss.Color("#7D56F4")
	wizOtherGreen  = lipgloss.Color("#A6E3A1")
	wizOtherGray   = lipgloss.Color("#6C7086")
	wizOtherWhite  = lipgloss.Color("#CDD6F4")
)

var (
	wizOtherSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(wizOtherPurple)
	wizOtherEnabledStyle  = lipgloss.NewStyle().Foreground(wizOtherGreen)
	wizOtherDimStyle      = lipgloss.NewStyle().Foreground(wizOtherGray)
	wizOtherLabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(wizOtherWhite)
	wizOtherHelpStyle     = lipgloss.NewStyle().Foreground(wizOtherGray)
)

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

// Category grouping for sections
type otherCategory int

const (
	categoryDisabledFeatures otherCategory = iota
	categoryGeneralSettings
	categoryClaudeCode
	categoryAgentsLoops
	categoryInfrastructure
	categoryAdvanced
	categoryCount // sentinel for iteration
)

var otherCategoryNames = []string{
	"Disabled Features",
	"General Settings",
	"Claude Code",
	"Agents & Loops",
	"Infrastructure",
	"Advanced",
}

var categorySections = [][]otherSection{
	// categoryDisabledFeatures
	{sectionDisabledMcps, sectionDisabledAgents, sectionDisabledSkills, sectionDisabledCommands, sectionDisabledTools},
	// categoryGeneralSettings
	{sectionAutoUpdate, sectionHashlineEdit, sectionModelFallback, sectionNewTaskSystemEnabled, sectionStartWork},
	// categoryClaudeCode
	{sectionClaudeCode, sectionModelCapabilities},
	// categoryAgentsLoops
	{sectionDefaultRunAgent, sectionSisyphusAgent, sectionSisyphus, sectionRalphLoop, sectionBabysitting, sectionCommentChecker},
	// categoryInfrastructure
	{sectionBackgroundTask, sectionTmux, sectionBrowserAutomationEngine, sectionWebsearch, sectionNotification, sectionGitMaster},
	// categoryAdvanced
	{sectionExperimental, sectionOpenclaw, sectionRuntimeFallback, sectionSkillsJson},
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

	// UI State — category navigation
	currentCategory  otherCategory
	categoryExpanded map[otherCategory]bool
	inCategory       bool // true when cursor is on a section within an expanded category

	// UI State — section navigation
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
		sectionExpanded:  sectionExpanded,
		categoryExpanded: make(map[otherCategory]bool),
		keys:             newWizardOtherKeyMap(),
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
	w.refreshView()
}

func (w *WizardOther) refreshView() {
	w.viewport.SetContent(w.renderContent())
}

func (w WizardOther) View() string {
	titleStyle := wizOtherLabelStyle
	helpStyle := wizOtherHelpStyle

	title := titleStyle.Render("Other Settings")
	desc := helpStyle.Render("[Enter/→] expand  [←] collapse  [Space] toggle  [Tab] next  [Esc] back")

	if w.inSubSection {
		desc = helpStyle.Render("Space/Enter to toggle • Esc to close section")
	} else if w.inCategory {
		desc = helpStyle.Render("[Enter/→] expand  [←] collapse  [Space] toggle  [Esc] back to category")
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
