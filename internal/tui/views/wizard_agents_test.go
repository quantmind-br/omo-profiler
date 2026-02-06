package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
)

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
				Model:        "claude-sonnet-4",
				Temperature: &temp,
				Skills:       []string{"coding", "testing"},
			},
			"plan": {
				Model: "gpt-4",
			},
		},
	}

	wa.SetConfig(cfg)

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
	wa.Apply(cfg)

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
