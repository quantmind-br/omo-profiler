package views

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
)

func boolPtr(b bool) *bool { return &b }

func TestParseMapStringBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single pair true",
			input:    "key:true",
			expected: map[string]bool{"key": true},
		},
		{
			name:     "single pair false",
			input:    "key:false",
			expected: map[string]bool{"key": false},
		},
		{
			name:     "multiple pairs mixed",
			input:    "key1:true, key2:false, key3:true",
			expected: map[string]bool{"key1": true, "key2": false, "key3": true},
		},
		{
			name:     "with spaces",
			input:    "  key1 : true , key2 : false  ",
			expected: map[string]bool{"key1": true, "key2": false},
		},
		{
			name:     "ignores invalid pairs - no colon",
			input:    "key1, key2:true",
			expected: map[string]bool{"key2": true},
		},
		{
			name:     "ignores invalid pairs - empty key",
			input:    ":true, key2:false",
			expected: map[string]bool{"key2": false},
		},
		{
			name:     "non-true values are false",
			input:    "key1:false, key2:abc, key3:anything",
			expected: map[string]bool{"key1": false, "key2": false, "key3": false},
		},
		{
			name:     "case sensitive - only lowercase true",
			input:    "key:true, key2:True, key3:TRUE",
			expected: map[string]bool{"key": true, "key2": false, "key3": false},
		},
		{
			name:     "only invalid pairs returns nil",
			input:    "abc, :true",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMapStringBool(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}

			for k, expectedVal := range tt.expected {
				if val, ok := result[k]; !ok {
					t.Errorf("missing key %q in result", k)
				} else if val != expectedVal {
					t.Errorf("key %q: expected %t, got %t", k, expectedVal, val)
				}
			}
		})
	}
}

func TestSerializeMapStringBool(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]bool
		expected string
		contains []string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty map",
			input:    map[string]bool{},
			expected: "",
		},
		{
			name:     "single true",
			input:    map[string]bool{"key": true},
			expected: "key:true",
		},
		{
			name:     "single false",
			input:    map[string]bool{"key": false},
			expected: "key:false",
		},
		{
			name:     "multiple mixed",
			input:    map[string]bool{"key1": true, "key2": false, "key3": true},
			contains: []string{"key1:true", "key2:false", "key3:true"},
		},
		{
			name:     "all true",
			input:    map[string]bool{"key1": true, "key2": true, "key3": true},
			contains: []string{"key1:true", "key2:true", "key3:true"},
		},
		{
			name:     "all false",
			input:    map[string]bool{"key1": false, "key2": false},
			contains: []string{"key1:false", "key2:false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeMapStringBool(tt.input)

			// For exact match tests (nil, empty, single item)
			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
				return
			}

			// For contains tests (multiple items - map order is not guaranteed)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, result)
				}
			}

			// Check comma count (n items = n-1 commas), skip for nil/empty
			if len(tt.input) > 0 {
				expectedCommas := len(tt.input) - 1
				commaCount := strings.Count(result, ", ")
				if commaCount != expectedCommas {
					t.Errorf("expected %d commas, got %d in result %q", expectedCommas, commaCount, result)
				}
			}
		})
	}
}

func TestParseSerializeMapStringBoolRoundTrip(t *testing.T) {
	original := map[string]bool{
		"key1": true,
		"key2": false,
		"key3": true,
		"key4": false,
	}

	serialized := serializeMapStringBool(original)
	parsed := parseMapStringBool(serialized)

	if len(parsed) != len(original) {
		t.Errorf("round trip: expected %d items, got %d", len(original), len(parsed))
	}

	for k, v := range original {
		if parsed[k] != v {
			t.Errorf("round trip: key %q: expected %t, got %t", k, v, parsed[k])
		}
	}
}

func TestNewAgentConfig(t *testing.T) {
	cfg := newAgentConfig()

	if !cfg.enabled {
		// enabled is false by default
		t.Log("agent config created with enabled=false")
	}

	// description is not focused by default in newAgentConfig()
	if cfg.description.Focused() {
		t.Error("expected description to not be focused initially")
	}

	if cfg.temperature.Placeholder == "" {
		t.Error("expected temperature placeholder to be set")
	}

	if cfg.topP.Placeholder == "" {
		t.Error("expected topP placeholder to be set")
	}

	if cfg.variant.Placeholder == "" {
		t.Error("expected variant placeholder to be set")
	}

	if cfg.category.Placeholder == "" {
		t.Error("expected category placeholder to be set")
	}
}

func TestNewWizardAgents(t *testing.T) {
	wa := NewWizardAgents()

	if len(wa.agents) != len(allAgents) {
		t.Errorf("expected %d agents, got %d", len(allAgents), len(wa.agents))
	}

	if wa.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", wa.cursor)
	}

	if wa.inForm {
		t.Error("expected inForm to be false initially")
	}

	if wa.ready {
		t.Error("expected ready to be false initially")
	}

	// Check that all agents exist
	for _, name := range allAgents {
		if _, ok := wa.agents[name]; !ok {
			t.Errorf("expected agent %q to exist", name)
		}
	}

	// Check key bindings
	if wa.keys.Up.Help().Key == "" {
		t.Error("expected Up key to be initialized")
	}

	if wa.keys.Down.Help().Key == "" {
		t.Error("expected Down key to be initialized")
	}

	if wa.keys.Toggle.Help().Key == "" {
		t.Error("expected Toggle key to be initialized")
	}

	if wa.keys.Expand.Help().Key == "" {
		t.Error("expected Expand key to be initialized")
	}

	if wa.keys.Next.Help().Key == "" {
		t.Error("expected Next key to be initialized")
	}

	if wa.keys.Back.Help().Key == "" {
		t.Error("expected Back key to be initialized")
	}
}

func TestWizardAgentsInit(t *testing.T) {
	wa := NewWizardAgents()
	cmd := wa.Init()

	if cmd != nil {
		t.Error("expected nil command from Init")
	}
}

func TestWizardAgentsSetSize(t *testing.T) {
	wa := NewWizardAgents()

	wa.SetSize(100, 50)

	if wa.width != 100 {
		t.Errorf("expected width 100, got %d", wa.width)
	}

	if wa.height != 50 {
		t.Errorf("expected height 50, got %d", wa.height)
	}

	if !wa.ready {
		t.Error("expected ready to be true after SetSize")
	}

	// Call SetSize again to test the ready=true path
	wa.SetSize(80, 40)

	if wa.width != 80 {
		t.Errorf("expected width 80, got %d", wa.width)
	}

	if wa.height != 40 {
		t.Errorf("expected height 40, got %d", wa.height)
	}
}

func TestWizardAgentsSetConfig(t *testing.T) {
	wa := NewWizardAgents()

	temp := 0.7
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:       "claude-sonnet-4",
				Temperature: &temp,
				Skills:      []string{"coding", "testing"},
			},
			"plan": {
				Model: "gpt-4",
			},
		},
	}

	wa.SetConfig(cfg, nil)

	// Check that build agent is enabled
	if !wa.agents["build"].enabled {
		t.Error("expected build agent to be enabled")
	}

	if wa.agents["build"].modelValue != "claude-sonnet-4" {
		t.Errorf("expected model 'claude-sonnet-4', got %q", wa.agents["build"].modelValue)
	}

	if wa.agents["build"].temperature.Value() != "0.7" {
		t.Errorf("expected temperature '0.7', got %q", wa.agents["build"].temperature.Value())
	}

	if wa.agents["build"].skills.Value() != "coding, testing" {
		t.Errorf("expected skills 'coding, testing', got %q", wa.agents["build"].skills.Value())
	}

	// Check that plan agent is enabled
	if !wa.agents["plan"].enabled {
		t.Error("expected plan agent to be enabled")
	}

	if wa.agents["plan"].modelValue != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %q", wa.agents["plan"].modelValue)
	}
}

func TestWizardAgentsApply(t *testing.T) {
	wa := NewWizardAgents()

	// Enable and configure build agent
	wa.agents["build"].enabled = true
	wa.agents["build"].modelValue = "claude-sonnet-4"
	wa.agents["build"].variant.SetValue("v1")
	wa.agents["build"].category.SetValue("coding")
	wa.agents["build"].temperature.SetValue("0.7")
	wa.agents["build"].skills.SetValue("coding")

	cfg := &config.Config{}
	wa.Apply(cfg, nil)

	if cfg.Agents == nil {
		t.Fatal("expected Agents to be set")
	}

	if len(cfg.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(cfg.Agents))
	}

	agentCfg, ok := cfg.Agents["build"]
	if !ok {
		t.Fatal("expected 'build' agent to exist")
	}

	if agentCfg.Model != "claude-sonnet-4" {
		t.Errorf("expected model 'claude-sonnet-4', got %q", agentCfg.Model)
	}

	if agentCfg.Variant != "v1" {
		t.Errorf("expected variant 'v1', got %q", agentCfg.Variant)
	}

	if agentCfg.Category != "coding" {
		t.Errorf("expected category 'coding', got %q", agentCfg.Category)
	}

	if agentCfg.Temperature == nil {
		t.Error("expected Temperature to be set")
	} else if *agentCfg.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", *agentCfg.Temperature)
	}
}

func TestWizardAgentsUpdateToggleKey(t *testing.T) {
	wa := NewWizardAgents()

	// build agent should be disabled initially
	if wa.agents["build"].enabled {
		t.Error("expected build agent to be disabled initially")
	}

	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd := wa.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Space key")
	}

	if !updated.agents["build"].enabled {
		t.Error("expected build agent to be enabled after Space")
	}
}

func TestWizardAgentsUpdateExpandKey(t *testing.T) {
	wa := NewWizardAgents()
	wa.SetSize(80, 24)

	// Enable build agent first
	wa.agents["build"].enabled = true

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := wa.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Enter key")
	}

	if !updated.agents["build"].expanded {
		t.Error("expected agent to be expanded after Enter")
	}

	if !updated.inForm {
		t.Error("expected inForm to be true after expansion")
	}

	if updated.focusedField != fieldModel {
		t.Errorf("expected focusedField to be fieldModel, got %v", updated.focusedField)
	}
}

func TestWizardAgentsUpdateFormEsc(t *testing.T) {
	wa := NewWizardAgents()
	wa.SetSize(80, 24)

	// Enable and expand build agent
	wa.agents["build"].enabled = true
	wa.agents["build"].expanded = true
	wa.inForm = true
	wa.focusedField = fieldModel

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := wa.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc in form mode")
	}

	if updated.inForm {
		t.Error("expected inForm to be false after Esc")
	}

	if updated.agents["build"].expanded {
		t.Error("expected agent to be collapsed after Esc")
	}
}

func TestWizardAgentsUpdateModelSelectedMsg(t *testing.T) {
	wa := NewWizardAgents()

	msg := ModelSelectedMsg{
		ModelID:     "claude-sonnet-4",
		DisplayName: "Claude Sonnet 4",
	}

	updated, cmd := wa.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for ModelSelectedMsg")
	}

	if updated.agents["build"].modelValue != "claude-sonnet-4" {
		t.Errorf("expected modelValue 'claude-sonnet-4', got %q", updated.agents["build"].modelValue)
	}

	if updated.agents["build"].modelDisplay != "Claude Sonnet 4" {
		t.Errorf("expected modelDisplay 'Claude Sonnet 4', got %q", updated.agents["build"].modelDisplay)
	}

	if updated.agents["build"].selectingModel {
		t.Error("expected selectingModel to be false after selection")
	}
}

func TestWizardAgentsView(t *testing.T) {
	wa := NewWizardAgents()
	wa.SetSize(80, 24)
	wa.viewport.SetContent(wa.renderContent())

	view := wa.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Configure Agents") {
		t.Error("expected 'Configure Agents' in view")
	}

	if !contains(view, "Space to enable/disable") {
		t.Error("expected toggle help in view")
	}
}

func roundTripFallbackModels(t *testing.T, input interface{}) interface{} {
	t.Helper()

	wa := NewWizardAgents()
	wa.SetConfig(&config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:          "test-model",
				FallbackModels: input,
			},
		},
	}, nil)

	out := &config.Config{}
	wa.Apply(out, nil)
	if out.Agents == nil || out.Agents["build"] == nil {
		t.Fatal("expected build agent after round trip")
	}
	return out.Agents["build"].FallbackModels
}

func normalizeJSONValue(t *testing.T, value interface{}) interface{} {
	t.Helper()
	if value == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var normalized interface{}
	if err := json.Unmarshal(b, &normalized); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	return normalized
}

func TestFallbackModelsStringRoundTrip(t *testing.T) {
	got := roundTripFallbackModels(t, "claude-sonnet-4")
	if got != "claude-sonnet-4" {
		t.Fatalf("expected string fallback preserved, got %#v", got)
	}
}

func TestFallbackModelsArrayRoundTrip(t *testing.T) {
	input := []interface{}{"model1", "model2"}
	got := roundTripFallbackModels(t, input)
	if !reflect.DeepEqual(normalizeJSONValue(t, got), normalizeJSONValue(t, input)) {
		t.Fatalf("expected array fallback preserved, got %#v", got)
	}
}

func TestFallbackModelsModelObjectRoundTrip(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"model":           "model1",
			"variant":         "fast",
			"reasoningEffort": "high",
		},
	}
	got := roundTripFallbackModels(t, input)
	if !reflect.DeepEqual(normalizeJSONValue(t, got), normalizeJSONValue(t, input)) {
		t.Fatalf("expected model object fallback preserved, got %#v", got)
	}
}

func TestFallbackModelsComplexObjectPreserved(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"model":           "model1",
			"variant":         "fast",
			"reasoningEffort": "high",
			"temperature":     0.7,
			"top_p":           0.9,
			"maxTokens":       4096,
			"thinking":        map[string]interface{}{"type": "enabled", "budgetTokens": 2048},
		},
	}
	got := roundTripFallbackModels(t, input)
	if !reflect.DeepEqual(normalizeJSONValue(t, got), normalizeJSONValue(t, input)) {
		t.Fatalf("expected complex fallback preserved, got %#v", got)
	}
}

func TestFallbackModelsEmptyRoundTrip(t *testing.T) {
	got := roundTripFallbackModels(t, nil)
	if got != nil {
		t.Fatalf("expected nil fallback preserved, got %#v", got)
	}
}

// === DATA PRESERVATION TESTS ===
// These tests ensure data is NOT lost through SetConfig → Apply cycles

func TestAgentApplyPreservesExistingFields(t *testing.T) {
	temp := 0.7
	topP := 0.9
	disable := true
	maxTokens := float64(8192)
	budgetTokens := float64(10000)

	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:           "claude-sonnet-4-20250514",
				Variant:         "fast",
				Category:        "coding",
				Temperature:     &temp,
				TopP:            &topP,
				Skills:          []string{"coding", "testing", "refactoring"},
				Tools:           map[string]bool{"bash": true, "edit": false},
				Prompt:          "You are a code assistant",
				PromptAppend:    "Always use TDD",
				Disable:         &disable,
				Description:     "Main build agent",
				Mode:            "subagent",
				Color:           "#FF6AC1",
				MaxTokens:       &maxTokens,
				Thinking:        &config.ThinkingConfig{Type: "enabled", BudgetTokens: &budgetTokens},
				ReasoningEffort: "high",
				TextVerbosity:   "high",
				ProviderOptions: map[string]interface{}{"custom_flag": true, "timeout": 30},
				Permission: &config.PermissionConfig{
					Edit:              "allow",
					Bash:              "ask",
					Webfetch:          "deny",
					DoomLoop:          "allow",
					ExternalDirectory: "ask",
				},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{Agents: make(map[string]*config.AgentConfig)}
	wa.Apply(newCfg, nil)

	agentCfg, ok := newCfg.Agents["build"]
	if !ok {
		t.Fatal("expected 'build' agent to exist after Apply")
	}

	if agentCfg.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model: expected 'claude-sonnet-4-20250514', got %q", agentCfg.Model)
	}
	if agentCfg.Variant != "fast" {
		t.Errorf("Variant: expected 'fast', got %q", agentCfg.Variant)
	}
	if agentCfg.Category != "coding" {
		t.Errorf("Category: expected 'coding', got %q", agentCfg.Category)
	}
	if agentCfg.Temperature == nil || *agentCfg.Temperature != 0.7 {
		t.Errorf("Temperature: expected 0.7, got %v", agentCfg.Temperature)
	}
	if agentCfg.TopP == nil || *agentCfg.TopP != 0.9 {
		t.Errorf("TopP: expected 0.9, got %v", agentCfg.TopP)
	}

	if len(agentCfg.Skills) != 3 {
		t.Errorf("Skills: expected 3 skills, got %d", len(agentCfg.Skills))
	}

	if agentCfg.Tools == nil || agentCfg.Tools["bash"] != true || agentCfg.Tools["edit"] != false {
		t.Errorf("Tools: not preserved correctly, got %v", agentCfg.Tools)
	}

	if agentCfg.Prompt != "You are a code assistant" {
		t.Errorf("Prompt: expected 'You are a code assistant', got %q", agentCfg.Prompt)
	}
	if agentCfg.PromptAppend != "Always use TDD" {
		t.Errorf("PromptAppend: expected 'Always use TDD', got %q", agentCfg.PromptAppend)
	}

	if agentCfg.Disable == nil || *agentCfg.Disable != true {
		t.Errorf("Disable: expected true, got %v", agentCfg.Disable)
	}

	if agentCfg.Description != "Main build agent" {
		t.Errorf("Description: expected 'Main build agent', got %q", agentCfg.Description)
	}
	if agentCfg.Mode != "subagent" {
		t.Errorf("Mode: expected 'subagent', got %q", agentCfg.Mode)
	}
	if agentCfg.Color != "#FF6AC1" {
		t.Errorf("Color: expected '#FF6AC1', got %q", agentCfg.Color)
	}

	if agentCfg.Permission == nil {
		t.Fatal("Permission: expected non-nil")
	}
	if agentCfg.Permission.Edit != "allow" {
		t.Errorf("Permission.Edit: expected 'allow', got %q", agentCfg.Permission.Edit)
	}
	if bashStr, ok := agentCfg.Permission.Bash.(string); !ok || bashStr != "ask" {
		t.Errorf("Permission.Bash: expected 'ask', got %v", agentCfg.Permission.Bash)
	}
	if agentCfg.Permission.Webfetch != "deny" {
		t.Errorf("Permission.Webfetch: expected 'deny', got %q", agentCfg.Permission.Webfetch)
	}
	if agentCfg.Permission.DoomLoop != "allow" {
		t.Errorf("Permission.DoomLoop: expected 'allow', got %q", agentCfg.Permission.DoomLoop)
	}
	if agentCfg.Permission.ExternalDirectory != "ask" {
		t.Errorf("Permission.ExternalDirectory: expected 'ask', got %q", agentCfg.Permission.ExternalDirectory)
	}
}

func TestAgentApplyPreservesBashObjectPermission(t *testing.T) {
	bashObj := map[string]interface{}{"git": "allow", "rm": "deny"}
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "claude-sonnet-4",
				Permission: &config.PermissionConfig{
					Bash: bashObj,
				},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{Agents: make(map[string]*config.AgentConfig)}
	wa.Apply(newCfg, nil)

	agentCfg, ok := newCfg.Agents["build"]
	if !ok {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.Permission == nil {
		t.Fatal("Permission: expected non-nil")
	}

	preservedBash, ok := agentCfg.Permission.Bash.(map[string]interface{})
	if !ok {
		t.Fatalf("Permission.Bash: expected map[string]interface{}, got %T", agentCfg.Permission.Bash)
	}
	if preservedBash["git"] != "allow" {
		t.Errorf("Permission.Bash[git]: expected 'allow', got %v", preservedBash["git"])
	}
	if preservedBash["rm"] != "deny" {
		t.Errorf("Permission.Bash[rm]: expected 'deny', got %v", preservedBash["rm"])
	}
}

func TestAgentSetConfigPopulatesPromptTextareas(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Prompt:       "my custom prompt",
				PromptAppend: "appended text",
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{Agents: make(map[string]*config.AgentConfig)}
	wa.Apply(newCfg, nil)

	agentCfg, ok := newCfg.Agents["build"]
	if !ok {
		t.Fatal("expected 'build' agent to exist after Apply")
	}

	if agentCfg.Prompt != "my custom prompt" {
		t.Errorf("Prompt: expected 'my custom prompt', got %q", agentCfg.Prompt)
	}
	if agentCfg.PromptAppend != "appended text" {
		t.Errorf("PromptAppend: expected 'appended text', got %q", agentCfg.PromptAppend)
	}
}

func TestAgentApplyPreservesProviderOptions(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:           "claude-sonnet-4",
				ProviderOptions: map[string]interface{}{"custom_flag": true, "timeout": float64(30)},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{}
	wa.Apply(newCfg, nil)

	agentCfg, ok := newCfg.Agents["build"]
	if !ok {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.ProviderOptions == nil {
		t.Fatal("ProviderOptions: expected non-nil after Apply")
	}
	if agentCfg.ProviderOptions["custom_flag"] != true {
		t.Errorf("ProviderOptions[custom_flag]: expected true, got %v", agentCfg.ProviderOptions["custom_flag"])
	}
	if agentCfg.ProviderOptions["timeout"] != float64(30) {
		t.Errorf("ProviderOptions[timeout]: expected 30, got %v", agentCfg.ProviderOptions["timeout"])
	}
}

func TestAgentApplyProviderOptionsRoundTrip(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test-model",
				ProviderOptions: map[string]interface{}{
					"flag":    true,
					"timeout": float64(30),
					"name":    "test",
				},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	freshCfg := &config.Config{}
	wa.Apply(freshCfg, nil)

	agentCfg := freshCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.ProviderOptions == nil {
		t.Fatal("ProviderOptions: expected non-nil")
	}
	if agentCfg.ProviderOptions["flag"] != true {
		t.Errorf("ProviderOptions[flag]: expected true, got %v (type %T)", agentCfg.ProviderOptions["flag"], agentCfg.ProviderOptions["flag"])
	}
	if agentCfg.ProviderOptions["timeout"] != float64(30) {
		t.Errorf("ProviderOptions[timeout]: expected 30, got %v (type %T)", agentCfg.ProviderOptions["timeout"], agentCfg.ProviderOptions["timeout"])
	}
	if agentCfg.ProviderOptions["name"] != "test" {
		t.Errorf("ProviderOptions[name]: expected 'test', got %v", agentCfg.ProviderOptions["name"])
	}
}

func TestAgentApplyProviderOptionsEmpty(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test-model",
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	freshCfg := &config.Config{}
	wa.Apply(freshCfg, nil)

	agentCfg := freshCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.ProviderOptions != nil {
		t.Errorf("ProviderOptions: expected nil when no options set, got %v", agentCfg.ProviderOptions)
	}
}

func TestAgentApplyPreservesUnmanagedFieldsOnEdit(t *testing.T) {
	maxTokens := float64(8192)
	budgetTokens := float64(10000)

	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:           "claude-sonnet-4-20250514",
				MaxTokens:       &maxTokens,
				Thinking:        &config.ThinkingConfig{Type: "enabled", BudgetTokens: &budgetTokens},
				ReasoningEffort: "high",
				TextVerbosity:   "high",
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	existingCfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				MaxTokens:       &maxTokens,
				Thinking:        &config.ThinkingConfig{Type: "enabled", BudgetTokens: &budgetTokens},
				ReasoningEffort: "high",
				TextVerbosity:   "high",
			},
		},
	}
	wa.Apply(existingCfg, nil)

	agentCfg := existingCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected 'build' agent to exist after Apply")
	}

	if agentCfg.MaxTokens == nil {
		t.Error("MaxTokens: expected to be preserved, got nil")
	} else if *agentCfg.MaxTokens != 8192 {
		t.Errorf("MaxTokens: expected 8192, got %f", *agentCfg.MaxTokens)
	}

	if agentCfg.Thinking == nil {
		t.Error("Thinking: expected to be preserved, got nil")
	} else {
		if agentCfg.Thinking.Type != "enabled" {
			t.Errorf("Thinking.Type: expected 'enabled', got %q", agentCfg.Thinking.Type)
		}
		if agentCfg.Thinking.BudgetTokens == nil || *agentCfg.Thinking.BudgetTokens != 10000 {
			t.Errorf("Thinking.BudgetTokens: expected 10000, got %v", agentCfg.Thinking.BudgetTokens)
		}
	}

	if agentCfg.ReasoningEffort != "high" {
		t.Errorf("ReasoningEffort: expected 'high', got %q", agentCfg.ReasoningEffort)
	}
	if agentCfg.TextVerbosity != "high" {
		t.Errorf("TextVerbosity: expected 'high', got %q", agentCfg.TextVerbosity)
	}
}

func TestAgentApplyAllowNonGptModelHephaestus(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"hephaestus": {
				Model:            "test",
				AllowNonGptModel: boolPtr(true),
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{Agents: make(map[string]*config.AgentConfig)}
	wa.Apply(newCfg, nil)

	agentCfg := newCfg.Agents["hephaestus"]
	if agentCfg == nil {
		t.Fatal("expected hephaestus agent to exist after Apply")
	}
	if agentCfg.AllowNonGptModel == nil || !*agentCfg.AllowNonGptModel {
		t.Fatalf("expected allow_non_gpt_model=true, got %v", agentCfg.AllowNonGptModel)
	}
}

func TestAgentApplyAllowNonGptModelNonHephaestus(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:            "test",
				AllowNonGptModel: boolPtr(true),
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	newCfg := &config.Config{Agents: make(map[string]*config.AgentConfig)}
	wa.Apply(newCfg, nil)

	agentCfg := newCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected build agent to exist after Apply")
	}
	if agentCfg.AllowNonGptModel != nil {
		t.Fatalf("expected allow_non_gpt_model=nil for non-hephaestus, got %v", agentCfg.AllowNonGptModel)
	}
}

func TestReasoningEffortNewValues(t *testing.T) {
	tests := []struct {
		name   string
		effort string
		idx    int
	}{
		{name: "none", effort: "none", idx: 1},
		{name: "minimal", effort: "minimal", idx: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Agents: map[string]*config.AgentConfig{
					"build": {ReasoningEffort: tt.effort},
				},
			}

			wa := NewWizardAgents()
			wa.SetConfig(cfg, nil)

			if got := wa.agents["build"].reasoningEffortIdx; got != tt.idx {
				t.Fatalf("reasoningEffortIdx: expected %d for %q, got %d", tt.idx, tt.effort, got)
			}

			out := &config.Config{}
			wa.Apply(out, nil)

			agentCfg := out.Agents["build"]
			if agentCfg == nil {
				t.Fatalf("expected build agent to exist after Apply")
			}

			if agentCfg.ReasoningEffort != tt.effort {
				t.Fatalf("ReasoningEffort: expected %q, got %q", tt.effort, agentCfg.ReasoningEffort)
			}
		})
	}
}

func TestAgentApplyBashObjectRoundTrip(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test-model",
				Permission: &config.PermissionConfig{
					Bash: map[string]interface{}{"git": "allow", "rm": "deny"},
				},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	freshCfg := &config.Config{}
	wa.Apply(freshCfg, nil)

	agentCfg := freshCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.Permission == nil {
		t.Fatal("Permission: expected non-nil")
	}
	bashMap, ok := agentCfg.Permission.Bash.(map[string]interface{})
	if !ok {
		t.Fatalf("Permission.Bash: expected map[string]interface{}, got %T", agentCfg.Permission.Bash)
	}
	if bashMap["git"] != "allow" {
		t.Errorf("Permission.Bash[git]: expected 'allow', got %v", bashMap["git"])
	}
	if bashMap["rm"] != "deny" {
		t.Errorf("Permission.Bash[rm]: expected 'deny', got %v", bashMap["rm"])
	}
}

func TestAgentApplyBashStringPreserved(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test-model",
				Permission: &config.PermissionConfig{
					Bash: "ask",
				},
			},
		},
	}

	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	freshCfg := &config.Config{}
	wa.Apply(freshCfg, nil)

	agentCfg := freshCfg.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected 'build' agent to exist after Apply")
	}
	if agentCfg.Permission == nil {
		t.Fatal("Permission: expected non-nil")
	}
	bashStr, ok := agentCfg.Permission.Bash.(string)
	if !ok {
		t.Fatalf("Permission.Bash: expected string, got %T", agentCfg.Permission.Bash)
	}
	if bashStr != "ask" {
		t.Errorf("Permission.Bash: expected 'ask', got %q", bashStr)
	}
}

func TestAllEffortLevelsRoundTrip(t *testing.T) {
	for _, effort := range effortLevels {
		t.Run(effort, func(t *testing.T) {
			cfg := &config.Config{
				Agents: map[string]*config.AgentConfig{
					"build": {
						Model:           "test",
						ReasoningEffort: effort,
					},
				},
			}
			wa := NewWizardAgents()
			wa.SetConfig(cfg, nil)

			out := &config.Config{}
			wa.Apply(out, nil)

			agentCfg := out.Agents["build"]
			if agentCfg == nil {
				t.Fatal("expected build agent")
			}
			if agentCfg.ReasoningEffort != effort {
				t.Errorf("expected %q, got %q", effort, agentCfg.ReasoningEffort)
			}
		})
	}
}

func TestAllowNonGptModelFalseOnHephaestus(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"hephaestus": {
				Model:            "test",
				AllowNonGptModel: boolPtr(false),
			},
		},
	}
	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	out := &config.Config{}
	wa.Apply(out, nil)

	agentCfg := out.Agents["hephaestus"]
	if agentCfg == nil {
		t.Fatal("expected hephaestus agent")
	}
	if agentCfg.AllowNonGptModel == nil {
		t.Fatal("expected AllowNonGptModel to be set for hephaestus, got nil")
	}
	if *agentCfg.AllowNonGptModel != false {
		t.Errorf("expected AllowNonGptModel=false, got %v", *agentCfg.AllowNonGptModel)
	}
}

func TestProviderOptionsUnicodeKeys(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test",
				ProviderOptions: map[string]interface{}{
					"日本語":    "value1",
					"emoji🎉": true,
					"café":   float64(42),
				},
			},
		},
	}
	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	out := &config.Config{}
	wa.Apply(out, nil)

	agentCfg := out.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected build agent")
	}
	if agentCfg.ProviderOptions == nil {
		t.Fatal("expected non-nil ProviderOptions")
	}
	if agentCfg.ProviderOptions["日本語"] != "value1" {
		t.Errorf("expected 'value1' for unicode key, got %v", agentCfg.ProviderOptions["日本語"])
	}
	if agentCfg.ProviderOptions["emoji🎉"] != true {
		t.Errorf("expected true for emoji key, got %v", agentCfg.ProviderOptions["emoji🎉"])
	}
	if agentCfg.ProviderOptions["café"] != float64(42) {
		t.Errorf("expected 42 for accented key, got %v", agentCfg.ProviderOptions["café"])
	}
}

func TestFallbackModelsEmptyArrayBecomesNil(t *testing.T) {
	wa := NewWizardAgents()
	wa.agents["build"].enabled = true
	wa.agents["build"].modelValue = "test"
	wa.agents["build"].fallbackEntries = []fallbackModelEntry{}
	wa.agents["build"].fallbackModels.SetValue("")

	out := &config.Config{}
	wa.Apply(out, nil)

	agentCfg := out.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected build agent")
	}
	if agentCfg.FallbackModels != nil {
		t.Errorf("expected nil for empty fallback entries, got %v", agentCfg.FallbackModels)
	}
}

func TestBashPermissionObjectWithThreeRules(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model: "test",
				Permission: &config.PermissionConfig{
					Bash: map[string]interface{}{
						"git":    "allow",
						"rm":     "deny",
						"docker": "ask",
					},
				},
			},
		},
	}
	wa := NewWizardAgents()
	wa.SetConfig(cfg, nil)

	out := &config.Config{}
	wa.Apply(out, nil)

	agentCfg := out.Agents["build"]
	if agentCfg == nil {
		t.Fatal("expected build agent")
	}
	if agentCfg.Permission == nil {
		t.Fatal("expected non-nil Permission")
	}
	bashMap, ok := agentCfg.Permission.Bash.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", agentCfg.Permission.Bash)
	}
	if len(bashMap) != 3 {
		t.Errorf("expected 3 rules, got %d", len(bashMap))
	}
	if bashMap["git"] != "allow" {
		t.Errorf("expected git=allow, got %v", bashMap["git"])
	}
	if bashMap["rm"] != "deny" {
		t.Errorf("expected rm=deny, got %v", bashMap["rm"])
	}
	if bashMap["docker"] != "ask" {
		t.Errorf("expected docker=ask, got %v", bashMap["docker"])
	}
}
