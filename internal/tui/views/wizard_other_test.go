package views

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMapStringInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single pair",
			input:    "key:1",
			expected: map[string]int{"key": 1},
		},
		{
			name:     "multiple pairs",
			input:    "key1:1, key2:2, key3:3",
			expected: map[string]int{"key1": 1, "key2": 2, "key3": 3},
		},
		{
			name:     "with spaces",
			input:    "  key1 : 1 , key2 : 2  ",
			expected: map[string]int{"key1": 1, "key2": 2},
		},
		{
			name:     "ignores invalid pairs - no colon",
			input:    "key1, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "ignores invalid pairs - empty key",
			input:    ":1, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "ignores invalid pairs - non-numeric value",
			input:    "key1:abc, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "zero value",
			input:    "key:0",
			expected: map[string]int{"key": 0},
		},
		{
			name:     "negative values",
			input:    "key1:-1, key2:-10",
			expected: map[string]int{"key1": -1, "key2": -10},
		},
		{
			name:     "only invalid pairs returns nil",
			input:    "abc, :1, key:abc",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMapStringInt(tt.input)

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
					t.Errorf("key %q: expected %d, got %d", k, expectedVal, val)
				}
			}
		})
	}
}

func TestSerializeMapStringInt(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
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
			input:    map[string]int{},
			expected: "",
		},
		{
			name:     "single item",
			input:    map[string]int{"key": 1},
			expected: "key:1",
		},
		{
			name:     "multiple items",
			input:    map[string]int{"key1": 1, "key2": 2, "key3": 3},
			contains: []string{"key1:1", "key2:2", "key3:3"},
		},
		{
			name:     "zero value",
			input:    map[string]int{"key": 0},
			expected: "key:0",
		},
		{
			name:     "negative values",
			input:    map[string]int{"key1": -1, "key2": -10},
			contains: []string{"key1:-1", "key2:-10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeMapStringInt(tt.input)

			// For exact match tests (nil, empty, single item)
			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
				return
			}

			// For empty/nil maps, just check result is empty
			if len(tt.input) == 0 {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
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

func TestParseSerializeMapStringIntRoundTrip(t *testing.T) {
	original := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}

	serialized := serializeMapStringInt(original)
	parsed := parseMapStringInt(serialized)

	if len(parsed) != len(original) {
		t.Errorf("round trip: expected %d items, got %d", len(original), len(parsed))
	}

	for k, v := range original {
		if parsed[k] != v {
			t.Errorf("round trip: key %q: expected %d, got %d", k, v, parsed[k])
		}
	}
}

func TestWizardOtherLoadsCheckboxStateFromJSONPresence(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewSelectionFromPresence(map[string]bool{"experimental.aggressive_truncation": true})
	cfg := &config.Config{Experimental: &config.ExperimentalConfig{AggressiveTruncation: boolPtr(true)}}

	w.SetConfig(cfg, selection)
	w.currentSection = sectionExperimental
	w.sectionExpanded[sectionExperimental] = true

	content := w.renderSubSection(sectionExperimental)
	joined := strings.Join(content, "\n")
	if !strings.Contains(joined, "aggressive_truncation: [on]") {
		t.Fatalf("expected boolean field value toggle, got %q", joined)
	}
	if !strings.Contains(joined, "[✓]") {
		t.Fatalf("expected selected checkbox, got %q", joined)
	}
}

func TestWizardOtherBooleanFieldSeparatesInclusionAndValue(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("experimental.aggressive_truncation", true)
	w.selection = selection
	w.currentSection = sectionExperimental
	w.sectionExpanded[sectionExperimental] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.expAggressiveTrunc {
		t.Fatal("expected inclusion toggle to leave boolean value unchanged")
	}
	if updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected inclusion toggle to deselect field")
	}

	updated.selection.SetSelected("experimental.aggressive_truncation", true)
	updated.subValueFocused = true
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.expAggressiveTrunc {
		t.Fatal("expected value toggle to update boolean value")
	}
	if !updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected value toggle to leave inclusion selected")
	}
}

func TestWizardOtherApplyWritesOnlySelectedFields(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("experimental.auto_resume", true)
	selection.SetSelected("tmux.layout", true)

	w.expAutoResume = false
	w.expAggressiveTrunc = true
	w.tmuxLayoutIdx = 2
	w.tmuxEnabled = true

	cfg := &config.Config{}
	w.Apply(cfg, selection)

	if cfg.Experimental == nil || cfg.Experimental.AutoResume == nil || *cfg.Experimental.AutoResume {
		t.Fatalf("expected selected false boolean to persist, got %#v", cfg.Experimental)
	}
	if cfg.Experimental.AggressiveTruncation != nil {
		t.Fatalf("expected unselected experimental field to be omitted, got %#v", cfg.Experimental)
	}
	if cfg.Tmux == nil || cfg.Tmux.Layout != tmuxLayouts[2] {
		t.Fatalf("expected selected tmux layout to persist, got %#v", cfg.Tmux)
	}
	if cfg.Tmux.Enabled != nil {
		t.Fatalf("expected unselected tmux.enabled to be omitted, got %#v", cfg.Tmux)
	}
}

func TestWizardOtherUntouchedSectionsRemainOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.expPluginLoadTimeoutMs.SetValue("30000")
	w.ccPluginsOverride.SetValue("serena:true")
	w.tmuxMainPaneSize.SetValue("0.75")

	cfg := &config.Config{}
	w.Apply(cfg, selection)

	if cfg.Experimental != nil || cfg.ClaudeCode != nil || cfg.Tmux != nil {
		t.Fatalf("expected untouched sections to remain omitted, got experimental=%#v claude_code=%#v tmux=%#v", cfg.Experimental, cfg.ClaudeCode, cfg.Tmux)
	}
}

func setupWizardOtherWithSelection(t *testing.T, paths ...string) WizardOther {
	t.Helper()
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	for _, path := range paths {
		selection.SetSelected(path, true)
	}
	w.selection = selection
	return w
}

func applyAndMarshal(t *testing.T, w WizardOther, selection *profile.FieldSelection) map[string]interface{} {
	t.Helper()
	cfg := &config.Config{}
	w.Apply(cfg, selection)
	data, err := profile.MarshalSparse(cfg, selection, nil)
	require.NoError(t, err, "MarshalSparse should not error")
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "JSON unmarshal should not error")
	return result
}

func assertJSONContains(t *testing.T, jsonMap map[string]interface{}, key string) {
	t.Helper()
	parts := strings.Split(key, ".")
	current := jsonMap
	for i, part := range parts {
		if i == len(parts)-1 {
			_, exists := current[part]
			assert.True(t, exists, "expected key %q to exist", key)
			return
		}
		next, ok := current[part].(map[string]interface{})
		if !ok {
			t.Fatalf("expected %q to be a nested object, got %T", part, current[part])
		}
		current = next
	}
}

func assertJSONOmits(t *testing.T, jsonMap map[string]interface{}, key string) {
	t.Helper()
	parts := strings.Split(key, ".")
	current := jsonMap
	for i, part := range parts {
		if i == len(parts)-1 {
			_, exists := current[part]
			assert.False(t, exists, "expected key %q to NOT exist", key)
			return
		}
		next, ok := current[part].(map[string]interface{})
		if !ok {
			return
		}
		current = next
	}
}

func assertJSONEquals(t *testing.T, jsonMap map[string]interface{}, key string, expected interface{}) {
	t.Helper()
	parts := strings.Split(key, ".")
	current := jsonMap
	for i, part := range parts {
		if i == len(parts)-1 {
			actual, exists := current[part]
			require.True(t, exists, "expected key %q to exist", key)
			assert.Equal(t, expected, actual, "expected key %q to equal %v", key, expected)
			return
		}
		next, ok := current[part].(map[string]interface{})
		if !ok {
			t.Fatalf("expected %q to be a nested object, got %T", part, current[part])
		}
		current = next
	}
}

func TestWizardOtherDisabledAgentsInclusionToggle(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.selection = selection
	w.currentSection = sectionDisabledAgents
	w.sectionExpanded[sectionDisabledAgents] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_agents") {
		t.Fatal("expected disabled_agents to be selected after Row 0 toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("disabled_agents") {
		t.Fatal("expected disabled_agents to be deselected after second toggle")
	}
}

func TestWizardOtherDisabledSkillsInclusionToggle(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.selection = selection
	w.currentSection = sectionDisabledSkills
	w.sectionExpanded[sectionDisabledSkills] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_skills") {
		t.Fatal("expected disabled_skills to be selected after Row 0 toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("disabled_skills") {
		t.Fatal("expected disabled_skills to be deselected after second toggle")
	}
}

func TestWizardOtherDisabledCommandsInclusionToggle(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.selection = selection
	w.currentSection = sectionDisabledCommands
	w.sectionExpanded[sectionDisabledCommands] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_commands") {
		t.Fatal("expected disabled_commands to be selected after Row 0 toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("disabled_commands") {
		t.Fatal("expected disabled_commands to be deselected after second toggle")
	}
}

func TestWizardOtherDisabledMcpsInclusionToggle(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.selection = selection
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_mcps") {
		t.Fatal("expected disabled_mcps to be selected after Row 0 toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("disabled_mcps") {
		t.Fatal("expected disabled_mcps to be deselected after second toggle")
	}
}

func TestWizardOtherDisabledToolsInclusionToggle(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	w.selection = selection
	w.currentSection = sectionDisabledTools
	w.sectionExpanded[sectionDisabledTools] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_tools") {
		t.Fatal("expected disabled_tools to be selected after Row 0 toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("disabled_tools") {
		t.Fatal("expected disabled_tools to be deselected after second toggle")
	}
}
