package views

import (
	"encoding/json"
	"sort"
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
	marshalSelection := selection
	if marshalSelection == nil {
		marshalSelection = profile.NewBlankSelection()
		if cfg.DisabledAgents != nil {
			marshalSelection.SetSelected("disabled_agents", true)
		}
		if cfg.AutoUpdate != nil {
			marshalSelection.SetSelected("auto_update", true)
		}
		if cfg.Experimental != nil && cfg.Experimental.AggressiveTruncation != nil {
			marshalSelection.SetSelected("experimental.aggressive_truncation", true)
		}
	}
	data, err := profile.MarshalSparse(cfg, marshalSelection, nil)
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

func TestWizardOtherEditFlow_LoadsFieldPresence(t *testing.T) {
	profileJSON := `{
		"disabled_agents": ["sisyphus"],
		"experimental": {
			"aggressive_truncation": true
		}
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(profileJSON), &cfg))

	selection := profile.NewSelectionFromPresence(map[string]bool{
		"disabled_agents":                    true,
		"experimental.aggressive_truncation": true,
	})

	w := NewWizardOther()
	w.SetConfig(&cfg, selection)

	require.True(t, w.selection.IsSelected("disabled_agents"))
	require.True(t, w.selection.IsSelected("experimental.aggressive_truncation"))
	assert.True(t, w.disabledAgents["sisyphus"], "expected sisyphus checkbox to load as disabled")
	assert.True(t, w.expAggressiveTrunc, "expected aggressive truncation to load from config")

	marshaled := applyAndMarshal(t, w, selection)

	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(profileJSON), &expected))
	assert.Equal(t, expected, marshaled)
}

func TestWizardOtherEditFlow_ModifyAndSave(t *testing.T) {
	profileJSON := `{
		"disabled_agents": ["sisyphus"]
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(profileJSON), &cfg))

	selection := profile.NewSelectionFromPresence(map[string]bool{
		"disabled_agents": true,
	})

	w := NewWizardOther()
	w.SetConfig(&cfg, selection)
	w.disabledAgents["prometheus"] = true

	marshaled := applyAndMarshal(t, w, selection)
	assertJSONContains(t, marshaled, "disabled_agents")

	rawAgents, ok := marshaled["disabled_agents"].([]interface{})
	require.True(t, ok, "expected disabled_agents array, got %T", marshaled["disabled_agents"])

	actual := make([]string, 0, len(rawAgents))
	for _, agent := range rawAgents {
		actual = append(actual, agent.(string))
	}
	sort.Strings(actual)
	assert.Equal(t, []string{"prometheus", "sisyphus"}, actual)
}

func TestWizardOtherEditFlow_RemoveField(t *testing.T) {
	profileJSON := `{
		"auto_update": true
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(profileJSON), &cfg))

	selection := profile.NewSelectionFromPresence(map[string]bool{
		"auto_update": true,
	})

	w := NewWizardOther()
	w.SetConfig(&cfg, selection)
	require.True(t, w.selection.IsSelected("auto_update"))
	require.True(t, w.autoUpdate)

	w.selection.SetSelected("auto_update", false)
	marshaled := applyAndMarshal(t, w, selection)
	assertJSONOmits(t, marshaled, "auto_update")
}

func TestWizardOtherTemplateFlow_PreservesSelection(t *testing.T) {
	templateJSON := `{
		"disabled_agents": ["sisyphus"],
		"auto_update": true,
		"experimental": {
			"aggressive_truncation": true
		}
	}`

	var cfg config.Config
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &cfg))

	selection := profile.NewSelectionFromPresence(map[string]bool{
		"disabled_agents":                    true,
		"auto_update":                        true,
		"experimental.aggressive_truncation": true,
	})

	w := NewWizardOther()
	w.SetConfig(&cfg, selection)

	assert.True(t, w.selection.IsSelected("disabled_agents"))
	assert.True(t, w.selection.IsSelected("auto_update"))
	assert.True(t, w.selection.IsSelected("experimental.aggressive_truncation"))

	marshaled := applyAndMarshal(t, w, selection)

	var expected map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &expected))
	assert.Equal(t, expected, marshaled)
}

func TestWizardOtherNewProfile_BlankSelection(t *testing.T) {
	w := NewWizardOther()
	w.autoUpdate = true
	w.disabledAgents["sisyphus"] = true
	w.expAggressiveTrunc = true

	marshaled := applyAndMarshal(t, w, nil)
	assertJSONEquals(t, marshaled, "auto_update", true)
	assertJSONEquals(t, marshaled, "experimental.aggressive_truncation", true)

	rawAgents, ok := marshaled["disabled_agents"].([]interface{})
	require.True(t, ok, "expected disabled_agents array, got %T", marshaled["disabled_agents"])
	require.Len(t, rawAgents, 1)
	assert.Equal(t, "sisyphus", rawAgents[0])
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

// --- Disabled-list omission tests (unselected → field absent from JSON) ---

func TestWizardOtherOmission_DisabledAgents(t *testing.T) {
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "disabled_agents")
}

func TestWizardOtherOmission_DisabledSkills(t *testing.T) {
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "disabled_skills")
}

func TestWizardOtherOmission_DisabledCommands(t *testing.T) {
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "disabled_commands")
}

func TestWizardOtherOmission_DisabledMcps(t *testing.T) {
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "disabled_mcps")
}

func TestWizardOtherOmission_DisabledTools(t *testing.T) {
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "disabled_tools")
}

// --- Disabled-list empty-array tests (selected, no items → field as []) ---

func TestWizardOtherEmptyArray_DisabledAgents(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_agents")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "disabled_agents", []interface{}{})
}

func TestWizardOtherEmptyArray_DisabledSkills(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_skills")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "disabled_skills", []interface{}{})
}

func TestWizardOtherEmptyArray_DisabledCommands(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_commands")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "disabled_commands", []interface{}{})
}

func TestWizardOtherEmptyArray_DisabledMcps(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_mcps")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "disabled_mcps", []interface{}{})
}

func TestWizardOtherEmptyArray_DisabledTools(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_tools")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "disabled_tools", []interface{}{})
}

// --- Disabled-list with-values tests (selected with items → field as ["item1", ...]) ---

func TestWizardOtherWithValues_DisabledAgents(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_agents")
	w.disabledAgents["sisyphus"] = true
	w.disabledAgents["oracle"] = true
	result := applyAndMarshal(t, w, w.selection)
	assertJSONContains(t, result, "disabled_agents")
	actual := result["disabled_agents"].([]interface{})
	assert.Equal(t, 2, len(actual), "expected 2 disabled agents")
	// Order follows disableableAgents slice
	assert.Equal(t, "sisyphus", actual[0])
	assert.Equal(t, "oracle", actual[1])
}

func TestWizardOtherWithValues_DisabledSkills(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_skills")
	w.disabledSkills["playwright"] = true
	w.disabledSkills["git-master"] = true
	result := applyAndMarshal(t, w, w.selection)
	assertJSONContains(t, result, "disabled_skills")
	actual := result["disabled_skills"].([]interface{})
	assert.Equal(t, 2, len(actual), "expected 2 disabled skills")
	assert.Equal(t, "playwright", actual[0])
	assert.Equal(t, "git-master", actual[1])
}

func TestWizardOtherWithValues_DisabledCommands(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_commands")
	w.disabledCommands["ralph-loop"] = true
	w.disabledCommands["refactor"] = true
	result := applyAndMarshal(t, w, w.selection)
	assertJSONContains(t, result, "disabled_commands")
	actual := result["disabled_commands"].([]interface{})
	assert.Equal(t, 2, len(actual), "expected 2 disabled commands")
	assert.Equal(t, "ralph-loop", actual[0])
	assert.Equal(t, "refactor", actual[1])
}

func TestWizardOtherWithValues_DisabledMcps(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_mcps")
	w.disabledMcps.SetValue("mcp-server-1, mcp-server-2")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONContains(t, result, "disabled_mcps")
	actual := result["disabled_mcps"].([]interface{})
	assert.Equal(t, 2, len(actual), "expected 2 disabled mcps")
	assert.Equal(t, "mcp-server-1", actual[0])
	assert.Equal(t, "mcp-server-2", actual[1])
}

func TestWizardOtherWithValues_DisabledTools(t *testing.T) {
	w := setupWizardOtherWithSelection(t, "disabled_tools")
	w.disabledTools.SetValue("tool-a, tool-b")
	result := applyAndMarshal(t, w, w.selection)
	assertJSONContains(t, result, "disabled_tools")
	actual := result["disabled_tools"].([]interface{})
	assert.Equal(t, 2, len(actual), "expected 2 disabled tools")
	assert.Equal(t, "tool-a", actual[0])
	assert.Equal(t, "tool-b", actual[1])
}

// --- Boolean field distinction tests (auto_update) ---

func TestWizardOtherOmission_AutoUpdate(t *testing.T) {
	// Unselected → field completely absent from JSON
	w := setupWizardOtherWithSelection(t)
	result := applyAndMarshal(t, w, w.selection)
	assertJSONOmits(t, result, "auto_update")
}

func TestWizardOtherAutoUpdate_SelectedFalse(t *testing.T) {
	// Selected with value false → "auto_update": false in JSON
	w := setupWizardOtherWithSelection(t, "auto_update")
	w.autoUpdate = false
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "auto_update", false)
}

func TestWizardOtherAutoUpdate_SelectedTrue(t *testing.T) {
	// Selected with value true → "auto_update": true in JSON
	w := setupWizardOtherWithSelection(t, "auto_update")
	w.autoUpdate = true
	result := applyAndMarshal(t, w, w.selection)
	assertJSONEquals(t, result, "auto_update", true)
}

func TestWizardOtherInclusionSeparateFromValue_DisabledAgents(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("disabled_agents", true)
	w.selection = selection
	w.currentSection = sectionDisabledAgents
	w.sectionExpanded[sectionDisabledAgents] = true
	w.inSubSection = true
	w.disabledAgents = map[string]bool{"sisyphus": true, "oracle": false}

	w.subCursor = 0
	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})

	if updated.selection.IsSelected("disabled_agents") {
		t.Fatal("expected disabled_agents inclusion to be toggled off")
	}
	if !updated.disabledAgents["sisyphus"] {
		t.Fatal("expected sisyphus disabled=true to be preserved when inclusion toggled off")
	}
	if updated.disabledAgents["oracle"] {
		t.Fatal("expected oracle disabled=false to be preserved when inclusion toggled off")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_agents") {
		t.Fatal("expected disabled_agents inclusion to be toggled back on")
	}
	if !updated.disabledAgents["sisyphus"] {
		t.Fatal("expected sisyphus disabled=true to survive inclusion round-trip")
	}
	if updated.disabledAgents["oracle"] {
		t.Fatal("expected oracle disabled=false to survive inclusion round-trip")
	}

	cfg := &config.Config{}
	updated.Apply(cfg, updated.selection)
	if cfg.DisabledAgents == nil {
		t.Fatal("expected DisabledAgents to be set after Apply")
	}
	require.Equal(t, []string{"sisyphus"}, cfg.DisabledAgents)
}

func TestWizardOtherInclusionSeparateFromValue_BoolField(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("experimental.aggressive_truncation", true)
	w.selection = selection
	w.currentSection = sectionExperimental
	w.sectionExpanded[sectionExperimental] = true
	w.inSubSection = true
	w.subCursor = 0

	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected inclusion to be toggled off")
	}
	if updated.expAggressiveTrunc {
		t.Fatal("expected boolean value to remain false after inclusion toggle")
	}

	updated.toggleFieldSelection("experimental.aggressive_truncation")
	if !updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected inclusion to be toggled back on")
	}
	if updated.expAggressiveTrunc {
		t.Fatal("expected boolean value to still be false after inclusion round-trip")
	}

	updated.subValueFocused = true
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.expAggressiveTrunc {
		t.Fatal("expected boolean value to be toggled to true")
	}
	if !updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected inclusion to remain selected after value toggle")
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if updated.expAggressiveTrunc {
		t.Fatal("expected boolean value to be toggled back to false")
	}
	if !updated.selection.IsSelected("experimental.aggressive_truncation") {
		t.Fatal("expected inclusion to remain selected after second value toggle")
	}
}

func TestWizardOtherInclusionSeparateFromValue_SliceField(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("disabled_mcps", true)
	w.selection = selection
	w.currentSection = sectionDisabledMcps
	w.sectionExpanded[sectionDisabledMcps] = true
	w.inSubSection = true
	w.disabledMcps.SetValue("mcp-server-1,mcp-server-2")

	w.subCursor = 0
	updated, _ := w.Update(tea.KeyMsg{Type: tea.KeySpace})

	if updated.selection.IsSelected("disabled_mcps") {
		t.Fatal("expected disabled_mcps inclusion to be toggled off")
	}
	if updated.disabledMcps.Value() != "mcp-server-1,mcp-server-2" {
		t.Fatalf("expected text input value preserved, got %q", updated.disabledMcps.Value())
	}

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !updated.selection.IsSelected("disabled_mcps") {
		t.Fatal("expected disabled_mcps inclusion to be toggled back on")
	}
	if updated.disabledMcps.Value() != "mcp-server-1,mcp-server-2" {
		t.Fatalf("expected text input value to survive round-trip, got %q", updated.disabledMcps.Value())
	}

	cfg := &config.Config{}
	updated.Apply(cfg, updated.selection)
	if cfg.DisabledMCPs == nil || len(cfg.DisabledMCPs) != 2 {
		t.Fatalf("expected 2 disabled mcps after Apply, got %v", cfg.DisabledMCPs)
	}
}

func TestWizardOtherInclusionSeparateFromValue_StringField(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("default_run_agent", true)
	w.selection = selection
	w.currentSection = sectionDefaultRunAgent
	w.defaultRunAgent.SetValue("build")

	w.toggleFieldSelection("default_run_agent")
	if w.selection.IsSelected("default_run_agent") {
		t.Fatal("expected default_run_agent inclusion to be toggled off")
	}
	if w.defaultRunAgent.Value() != "build" {
		t.Fatalf("expected text input value preserved, got %q", w.defaultRunAgent.Value())
	}

	w.toggleFieldSelection("default_run_agent")
	if !w.selection.IsSelected("default_run_agent") {
		t.Fatal("expected default_run_agent inclusion to be toggled back on")
	}
	if w.defaultRunAgent.Value() != "build" {
		t.Fatalf("expected text input value to survive round-trip, got %q", w.defaultRunAgent.Value())
	}

	cfg := &config.Config{}
	w.Apply(cfg, w.selection)
	if cfg.DefaultRunAgent != "build" {
		t.Fatalf("expected default_run_agent='build' after Apply, got %q", cfg.DefaultRunAgent)
	}
}

// --- Round-trip tests: disabled-list sections (5 sections × 3 states) ---
// Each test verifies the full data path: set wizard state → Apply → MarshalSparse → verify JSON output.

// Round-trip: disabled_agents

func TestWizardOtherRoundTrip_DisabledAgents(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledAgentsFieldPath)
	w.disabledAgents["sisyphus"] = true
	w.disabledAgents["oracle"] = true

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_agents")
	assertJSONEquals(t, result, "disabled_agents", []interface{}{"sisyphus", "oracle"})
}

func TestWizardOtherRoundTrip_DisabledAgentsEmpty(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledAgentsFieldPath)
	// No agents toggled on → should serialize as []

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_agents")
	assertJSONEquals(t, result, "disabled_agents", []interface{}{})
}

func TestWizardOtherRoundTrip_DisabledAgentsOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	// disabled_agents NOT selected → key must be absent from JSON

	result := applyAndMarshal(t, w, selection)

	assertJSONOmits(t, result, "disabled_agents")
}

// Round-trip: disabled_skills

func TestWizardOtherRoundTrip_DisabledSkills(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledSkillsFieldPath)
	w.disabledSkills["playwright"] = true
	w.disabledSkills["git-master"] = true

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_skills")
	assertJSONEquals(t, result, "disabled_skills", []interface{}{"playwright", "git-master"})
}

func TestWizardOtherRoundTrip_DisabledSkillsEmpty(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledSkillsFieldPath)
	// No skills toggled on → should serialize as []

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_skills")
	assertJSONEquals(t, result, "disabled_skills", []interface{}{})
}

func TestWizardOtherRoundTrip_DisabledSkillsOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()

	result := applyAndMarshal(t, w, selection)

	assertJSONOmits(t, result, "disabled_skills")
}

// Round-trip: disabled_commands

func TestWizardOtherRoundTrip_DisabledCommands(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledCommandsFieldPath)
	w.disabledCommands["ralph-loop"] = true
	w.disabledCommands["refactor"] = true

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_commands")
	assertJSONEquals(t, result, "disabled_commands", []interface{}{"ralph-loop", "refactor"})
}

func TestWizardOtherRoundTrip_DisabledCommandsEmpty(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledCommandsFieldPath)
	// No commands toggled on → should serialize as []

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_commands")
	assertJSONEquals(t, result, "disabled_commands", []interface{}{})
}

func TestWizardOtherRoundTrip_DisabledCommandsOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()

	result := applyAndMarshal(t, w, selection)

	assertJSONOmits(t, result, "disabled_commands")
}

// Round-trip: disabled_mcps

func TestWizardOtherRoundTrip_DisabledMcps(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledMcpsFieldPath)
	w.disabledMcps.SetValue("mcp-server-1, mcp-server-2")

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_mcps")
	assertJSONEquals(t, result, "disabled_mcps", []interface{}{"mcp-server-1", "mcp-server-2"})
}

func TestWizardOtherRoundTrip_DisabledMcpsEmpty(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledMcpsFieldPath)
	// No MCPs entered (empty text input) → should serialize as []

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_mcps")
	assertJSONEquals(t, result, "disabled_mcps", []interface{}{})
}

func TestWizardOtherRoundTrip_DisabledMcpsOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()

	result := applyAndMarshal(t, w, selection)

	assertJSONOmits(t, result, "disabled_mcps")
}

// Round-trip: disabled_tools

func TestWizardOtherRoundTrip_DisabledTools(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledToolsFieldPath)
	w.disabledTools.SetValue("tool-a, tool-b")

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_tools")
	assertJSONEquals(t, result, "disabled_tools", []interface{}{"tool-a", "tool-b"})
}

func TestWizardOtherRoundTrip_DisabledToolsEmpty(t *testing.T) {
	w := setupWizardOtherWithSelection(t, disabledToolsFieldPath)
	// No tools entered (empty text input) → should serialize as []

	result := applyAndMarshal(t, w, w.selection)

	assertJSONContains(t, result, "disabled_tools")
	assertJSONEquals(t, result, "disabled_tools", []interface{}{})
}

func TestWizardOtherRoundTrip_DisabledToolsOmitted(t *testing.T) {
	w := NewWizardOther()
	selection := profile.NewBlankSelection()

	result := applyAndMarshal(t, w, selection)

	assertJSONOmits(t, result, "disabled_tools")
}

// --- Checkmark indicator tests (Bug 1 + 1b fixes) ---
// These tests verify that the ✓ indicator reflects field selection state,
// not the boolean field value.

func TestWizardOtherCheckmarkReflectsSelection_NotValue_SimpleBoolean(t *testing.T) {
	// Test: selected + value=false → ✓ present
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("auto_update", true)
	w.selection = selection
	w.autoUpdate = false
	w.currentSection = sectionAutoUpdate

	content := w.renderContent()
	if !strings.Contains(content, "[off]") {
		t.Fatal("expected [off] in render output")
	}
	if !strings.Contains(content, "✓") {
		t.Fatal("expected ✓ indicator when field is selected, even with value=false")
	}
}

func TestWizardOtherCheckmarkReflectsSelection_NotValue_Unselected(t *testing.T) {
	// Test: NOT selected + value=true → ✓ absent
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	// auto_update NOT selected
	w.selection = selection
	w.autoUpdate = true
	w.currentSection = sectionAutoUpdate

	content := w.renderContent()
	if !strings.Contains(content, "[on]") {
		t.Fatal("expected [on] in render output")
	}
	// The checkbox [✓] should NOT be present (field not selected)
	// But the value ✓ after [on] should also NOT be present
	// We need to check that the line doesn't have the green ✓ after the value
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Auto Update") {
			// Should have [on] but NOT the green ✓ after it
			if strings.Contains(line, "[on] ✓") {
				t.Fatalf("expected NO ✓ after [on] when field not selected, got: %s", line)
			}
			// Should have [ ] checkbox, not [✓]
			if strings.Contains(line, "[✓]") {
				t.Fatalf("expected [ ] checkbox when field not selected, got: %s", line)
			}
		}
	}
}

func TestWizardOtherCheckmarkReflectsSelection_SubSectionBoolField(t *testing.T) {
	// Test: sub-section boolean selected + value=false → ✓ present
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	selection.SetSelected("experimental.aggressive_truncation", true)
	w.selection = selection
	w.expAggressiveTrunc = false
	w.currentSection = sectionExperimental
	w.sectionExpanded[sectionExperimental] = true

	content := w.renderSubSection(sectionExperimental)
	joined := strings.Join(content, "\n")
	if !strings.Contains(joined, "aggressive_truncation:") {
		t.Fatal("expected aggressive_truncation in render output")
	}
	if !strings.Contains(joined, "[off]") {
		t.Fatal("expected [off] in render output")
	}
	// The ✓ should be present because the field is selected, even though value=false
	if !strings.Contains(joined, "✓") {
		t.Fatalf("expected ✓ indicator when field is selected, even with value=false, got: %s", joined)
	}
}

func TestWizardOtherCheckmarkReflectsSelection_SubSectionBoolField_Unselected(t *testing.T) {
	// Test: sub-section boolean NOT selected + value=true → ✓ absent
	w := NewWizardOther()
	selection := profile.NewBlankSelection()
	// experimental.aggressive_truncation NOT selected
	w.selection = selection
	w.expAggressiveTrunc = true
	w.currentSection = sectionExperimental
	w.sectionExpanded[sectionExperimental] = true

	content := w.renderSubSection(sectionExperimental)
	joined := strings.Join(content, "\n")
	if !strings.Contains(joined, "aggressive_truncation:") {
		t.Fatal("expected aggressive_truncation in render output")
	}
	if !strings.Contains(joined, "[on]") {
		t.Fatal("expected [on] in render output")
	}
	// The ✓ after [on] should NOT be present because the field is not selected
	lines := strings.Split(joined, "\n")
	for _, line := range lines {
		if strings.Contains(line, "aggressive_truncation") {
			if strings.Contains(line, "[on] ✓") {
				t.Fatalf("expected NO ✓ after [on] when field not selected, got: %s", line)
			}
		}
	}
}
