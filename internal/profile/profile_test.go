package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	return func() {
		config.ResetBaseDir()
	}
}

func TestLoad(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Ensure directories exist
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Create a test profile
	cfg := config.Config{
		DisabledMCPs: []string{"test-mcp"},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	profilePath := filepath.Join(config.ProfilesDir(), "test-profile.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	// Test Load
	p, err := Load("test-profile")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if p.Name != "test-profile" {
		t.Errorf("Expected name 'test-profile', got '%s'", p.Name)
	}

	if len(p.Config.DisabledMCPs) != 1 || p.Config.DisabledMCPs[0] != "test-mcp" {
		t.Errorf("Config not loaded correctly")
	}
}

func TestLoadNonexistent(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := Load("nonexistent")
	if err == nil {
		t.Error("Expected error when loading nonexistent profile")
	}
}

func TestSave(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	p := &Profile{
		Name: "new-profile",
		Config: config.Config{
			DisabledAgents: []string{"agent1"},
		},
	}

	if err := Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(config.ProfilesDir(), "new-profile.json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Profile file not created at %s", expectedPath)
	}

	// Verify content
	data, _ := os.ReadFile(expectedPath)
	var loaded config.Config
	_ = json.Unmarshal(data, &loaded)

	if len(loaded.DisabledAgents) != 1 || loaded.DisabledAgents[0] != "agent1" {
		t.Error("Saved config doesn't match original")
	}
}

func TestDelete(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Create a profile to delete
	profilePath := filepath.Join(config.ProfilesDir(), "to-delete.json")
	_ = os.WriteFile(profilePath, []byte("{}"), 0644)

	if err := Delete("to-delete"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Error("Profile file should be deleted")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	err := Delete("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent profile")
	}
}

func TestList(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Create test profiles
	profilesDir := config.ProfilesDir()
	_ = os.WriteFile(filepath.Join(profilesDir, "profile1.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(profilesDir, "profile2.json"), []byte("{}"), 0644)
	_ = os.WriteFile(filepath.Join(profilesDir, "not-json.txt"), []byte("{}"), 0644)

	profiles, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

	// Check that both are present (order may vary)
	found1, found2 := false, false
	for _, p := range profiles {
		if p == "profile1" {
			found1 = true
		}
		if p == "profile2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("Expected profile1 and profile2, got %v", profiles)
	}
}

func TestListEmpty(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Don't create profiles directory - should return empty list
	profiles, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}
}

func TestExists(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	// Create a profile
	profilePath := filepath.Join(config.ProfilesDir(), "exists-test.json")
	_ = os.WriteFile(profilePath, []byte("{}"), 0644)

	if !Exists("exists-test") {
		t.Error("Expected profile to exist")
	}

	if Exists("nonexistent") {
		t.Error("Expected profile to not exist")
	}
}

func TestLoadWithLegacyFields(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	jsonWithLegacy := `{
		"disabled_mcps": ["test-mcp"],
		"unknownLegacyField": "some value",
		"anotherUnknown": 123
	}`

	profilePath := filepath.Join(config.ProfilesDir(), "legacy-profile.json")
	if err := os.WriteFile(profilePath, []byte(jsonWithLegacy), 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	p, err := Load("legacy-profile")
	if err != nil {
		t.Fatalf("Load should succeed even with legacy fields: %v", err)
	}

	if !p.HasLegacyFields {
		t.Error("Expected HasLegacyFields to be true")
	}

	if p.LegacyFieldsWarning == "" {
		t.Error("Expected LegacyFieldsWarning to contain a message")
	}

	if len(p.Config.DisabledMCPs) != 1 || p.Config.DisabledMCPs[0] != "test-mcp" {
		t.Error("Known fields should still be loaded correctly")
	}
}

func TestLoadWithoutLegacyFields(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	validJSON := `{"disabled_mcps": ["valid-mcp"]}`

	profilePath := filepath.Join(config.ProfilesDir(), "valid-profile.json")
	if err := os.WriteFile(profilePath, []byte(validJSON), 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	p, err := Load("valid-profile")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if p.HasLegacyFields {
		t.Error("Expected HasLegacyFields to be false for valid profile")
	}

	if p.LegacyFieldsWarning != "" {
		t.Errorf("Expected empty LegacyFieldsWarning, got: %s", p.LegacyFieldsWarning)
	}
}

func TestProfileLoadPreservesUnknownJSON(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	profileJSON := `{
		"disabled_mcps": ["test-mcp"],
		"customField": {"enabled": true},
		"anotherLegacy": [1, 2, 3]
	}`

	profilePath := filepath.Join(config.ProfilesDir(), "preserve-unknown.json")
	if err := os.WriteFile(profilePath, []byte(profileJSON), 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	p, err := Load("preserve-unknown")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(p.Config.DisabledMCPs) != 1 || p.Config.DisabledMCPs[0] != "test-mcp" {
		t.Fatalf("known config field not loaded correctly: %#v", p.Config.DisabledMCPs)
	}

	if _, ok := p.PreservedUnknown["customField"]; !ok {
		t.Fatal("expected customField to be preserved")
	}

	if _, ok := p.PreservedUnknown["anotherLegacy"]; !ok {
		t.Fatal("expected anotherLegacy to be preserved")
	}

	if len(p.PreservedUnknown) != 2 {
		t.Fatalf("expected 2 preserved unknown keys, got %d", len(p.PreservedUnknown))
	}

	var customField map[string]bool
	if err := json.Unmarshal(p.PreservedUnknown["customField"], &customField); err != nil {
		t.Fatalf("failed to decode preserved customField: %v", err)
	}

	if !customField["enabled"] {
		t.Fatal("expected preserved customField.enabled to be true")
	}
	var anotherLegacy []int
	if err := json.Unmarshal(p.PreservedUnknown["anotherLegacy"], &anotherLegacy); err != nil {
		t.Fatalf("failed to decode preserved anotherLegacy: %v", err)
	}

	if !reflect.DeepEqual(anotherLegacy, []int{1, 2, 3}) {
		t.Fatalf("unexpected preserved anotherLegacy value: %#v", anotherLegacy)
	}
}

func TestProfileLoadCapturesFieldPresence(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	profileJSON := `{
		"disabled_mcps": ["test-mcp"],
		"agents": {
			"worker": {"model": "gpt-5"}
		}
	}`

	profilePath := filepath.Join(config.ProfilesDir(), "field-presence.json")
	if err := os.WriteFile(profilePath, []byte(profileJSON), 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	p, err := Load("field-presence")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !p.FieldPresence["disabled_mcps"] {
		t.Fatal("expected disabled_mcps to be marked present")
	}

	if !p.FieldPresence["agents"] {
		t.Fatal("expected agents to be marked present")
	}

	if _, ok := p.FieldPresence["categories"]; ok {
		t.Fatal("expected categories to be absent from FieldPresence")
	}
}

func TestProfileSaveRoundTripsPreservedUnknownFragments(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	p := &Profile{
		Name: "roundtrip-unknown",
		Config: config.Config{
			DisabledAgents: []string{"agent1"},
		},
		PreservedUnknown: map[string]json.RawMessage{
			"customField":   json.RawMessage(`{"enabled":true}`),
			"anotherLegacy": json.RawMessage(`[1,2,3]`),
		},
	}

	if err := Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reloaded, err := Load("roundtrip-unknown")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(reloaded.Config.DisabledAgents) != 1 || reloaded.Config.DisabledAgents[0] != "agent1" {
		t.Fatalf("known config field not round-tripped correctly: %#v", reloaded.Config.DisabledAgents)
	}

	var customField map[string]bool
	if err := json.Unmarshal(reloaded.PreservedUnknown["customField"], &customField); err != nil {
		t.Fatalf("failed to decode preserved customField: %v", err)
	}

	if !customField["enabled"] {
		t.Fatal("expected preserved customField.enabled to be true after reload")
	}

	var anotherLegacy []int
	if err := json.Unmarshal(reloaded.PreservedUnknown["anotherLegacy"], &anotherLegacy); err != nil {
		t.Fatalf("failed to decode preserved anotherLegacy: %v", err)
	}

	if !reflect.DeepEqual(anotherLegacy, []int{1, 2, 3}) {
		t.Fatalf("unexpected preserved anotherLegacy after reload: %#v", anotherLegacy)
	}
}

func TestProfileLoadFailsOnMalformedJSON(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	profilePath := filepath.Join(config.ProfilesDir(), "malformed.json")
	if err := os.WriteFile(profilePath, []byte("}{invalid"), 0644); err != nil {
		t.Fatalf("Failed to create malformed profile: %v", err)
	}

	if _, err := Load("malformed"); err == nil {
		t.Fatal("expected malformed JSON load to fail")
	}
}

func TestRegressionSparsePersistenceContract(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	const profileName = "regression-sparse-contract"
	profilePath := filepath.Join(config.ProfilesDir(), profileName+".json")
	initialProfileJSON := `{
		"disabled_mcps": ["legacy-mcp"],
		"hashline_edit": true,
		"custom_bundle": {
			"enabled": true,
			"thresholds": {
				"high": 2,
				"low": 1
			}
		},
		"custom_flags": ["alpha", "beta"]
	}`

	if err := os.WriteFile(profilePath, []byte(initialProfileJSON), 0644); err != nil {
		t.Fatalf("Failed to create initial regression profile: %v", err)
	}

	p, err := Load(profileName)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !p.FieldPresence["disabled_mcps"] {
		t.Fatal("expected disabled_mcps to be marked present on initial load")
	}
	if !p.FieldPresence["hashline_edit"] {
		t.Fatal("expected hashline_edit to be marked present on initial load")
	}
	if len(p.PreservedUnknown) != 2 {
		t.Fatalf("expected 2 top-level preserved unknown fragments, got %d", len(p.PreservedUnknown))
	}

	p.Config.DisabledMCPs = []string{"omit-me"}
	p.Config.HashlineEdit = boolPtr(false)
	p.Config.DisabledHooks = []string{}
	p.Config.DefaultRunAgent = ""
	p.Config.Experimental = &config.ExperimentalConfig{
		TaskSystem: boolPtr(false),
		MaxTools:   int64Ptr(0),
	}
	p.Config.Agents = map[string]*config.AgentConfig{
		"builder": {Model: "gpt-5"},
	}
	p.PreservedUnknown["agents"] = json.RawMessage(`{"builder":{"model":"legacy-model","legacy":true},"legacy_agent":{"model":"legacy-only"}}`)
	p.PreservedUnknown["experimental"] = json.RawMessage(`{"legacy_flag":true,"task_system":true}`)

	selection := NewBlankSelection()
	for _, path := range []string{
		"hashline_edit",
		"disabled_hooks",
		"default_run_agent",
		"experimental.task_system",
		"experimental.max_tools",
		"agents.*.model",
	} {
		selection.SetSelected(path, true)
	}

	data, err := MarshalSparse(&p.Config, selection, p.PreservedUnknown)
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	if _, ok := decoded["disabled_mcps"]; ok {
		t.Fatalf("expected unchecked disabled_mcps to be omitted, got %#v", decoded["disabled_mcps"])
	}

	serializedChecks := []struct {
		name  string
		check func(*testing.T, map[string]any)
	}{
		{
			name: "selected zero values survive sparse JSON",
			check: func(t *testing.T, payload map[string]any) {
				t.Helper()

				if value, ok := payload["hashline_edit"].(bool); !ok || value {
					t.Fatalf("expected hashline_edit to be false, got %#v", payload["hashline_edit"])
				}

				if value, ok := payload["default_run_agent"].(string); !ok || value != "" {
					t.Fatalf("expected default_run_agent to be an explicit empty string, got %#v", payload["default_run_agent"])
				}

				hooks, ok := payload["disabled_hooks"].([]any)
				if !ok || len(hooks) != 0 {
					t.Fatalf("expected disabled_hooks to be an explicit empty array, got %#v", payload["disabled_hooks"])
				}

				experimental := decodedObject(t, payload["experimental"], "experimental")
				if taskSystem, ok := experimental["task_system"].(bool); !ok || taskSystem {
					t.Fatalf("expected experimental.task_system to be false, got %#v", experimental["task_system"])
				}
				if maxTools, ok := experimental["max_tools"].(float64); !ok || maxTools != 0 {
					t.Fatalf("expected experimental.max_tools to be 0, got %#v", experimental["max_tools"])
				}
			},
		},
		{
			name: "multiple preserved unknown fragments survive and known leaves win overlaps",
			check: func(t *testing.T, payload map[string]any) {
				t.Helper()

				customBundle := decodedObject(t, payload["custom_bundle"], "custom_bundle")
				if enabled, ok := customBundle["enabled"].(bool); !ok || !enabled {
					t.Fatalf("expected preserved custom_bundle.enabled to remain true, got %#v", customBundle["enabled"])
				}

				thresholds := decodedObject(t, customBundle["thresholds"], "custom_bundle.thresholds")
				if low, ok := thresholds["low"].(float64); !ok || low != 1 {
					t.Fatalf("expected preserved custom_bundle.thresholds.low to remain 1, got %#v", thresholds["low"])
				}
				if high, ok := thresholds["high"].(float64); !ok || high != 2 {
					t.Fatalf("expected preserved custom_bundle.thresholds.high to remain 2, got %#v", thresholds["high"])
				}

				customFlags, ok := payload["custom_flags"].([]any)
				if !ok || !reflect.DeepEqual(customFlags, []any{"alpha", "beta"}) {
					t.Fatalf("expected preserved custom_flags to remain [alpha beta], got %#v", payload["custom_flags"])
				}

				agents := decodedObject(t, payload["agents"], "agents")
				builder := decodedObject(t, agents["builder"], "agents.builder")
				if model, ok := builder["model"].(string); !ok || model != "gpt-5" {
					t.Fatalf("expected selected agents.builder.model to win, got %#v", builder["model"])
				}
				if legacy, ok := builder["legacy"].(bool); !ok || !legacy {
					t.Fatalf("expected preserved agents.builder.legacy sibling to remain, got %#v", builder["legacy"])
				}

				legacyAgent := decodedObject(t, agents["legacy_agent"], "agents.legacy_agent")
				if model, ok := legacyAgent["model"].(string); !ok || model != "legacy-only" {
					t.Fatalf("expected preserved legacy_agent model to remain, got %#v", legacyAgent["model"])
				}

				experimental := decodedObject(t, payload["experimental"], "experimental")
				if legacyFlag, ok := experimental["legacy_flag"].(bool); !ok || !legacyFlag {
					t.Fatalf("expected preserved experimental.legacy_flag sibling to remain, got %#v", experimental["legacy_flag"])
				}
			},
		},
	}

	for _, tc := range serializedChecks {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t, decoded)
		})
	}

	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("Failed to persist sparse regression profile: %v", err)
	}

	reloaded, err := Load(profileName)
	if err != nil {
		t.Fatalf("Reload after sparse save failed: %v", err)
	}

	if reloaded.FieldPresence["disabled_mcps"] {
		t.Fatal("expected unchecked disabled_mcps to stay omitted after reload")
	}
	for _, requiredKey := range []string{"hashline_edit", "disabled_hooks", "default_run_agent", "agents", "experimental"} {
		if !reloaded.FieldPresence[requiredKey] {
			t.Fatalf("expected %s to be marked present after sparse reload", requiredKey)
		}
	}

	if reloaded.Config.HashlineEdit == nil || *reloaded.Config.HashlineEdit {
		t.Fatalf("expected hashline_edit to reload as false, got %#v", reloaded.Config.HashlineEdit)
	}
	if !reflect.DeepEqual(reloaded.Config.DisabledHooks, []string{}) {
		t.Fatalf("expected disabled_hooks to reload as an explicit empty slice, got %#v", reloaded.Config.DisabledHooks)
	}
	if reloaded.Config.Experimental == nil {
		t.Fatal("expected experimental config to reload")
	}
	if reloaded.Config.Experimental.TaskSystem == nil || *reloaded.Config.Experimental.TaskSystem {
		t.Fatalf("expected experimental.task_system to reload as false, got %#v", reloaded.Config.Experimental.TaskSystem)
	}
	if reloaded.Config.Experimental.MaxTools == nil || *reloaded.Config.Experimental.MaxTools != 0 {
		t.Fatalf("expected experimental.max_tools to reload as 0, got %#v", reloaded.Config.Experimental.MaxTools)
	}
	if got := reloaded.Config.DefaultRunAgent; got != "" {
		t.Fatalf("expected default_run_agent to reload as empty string, got %q", got)
	}
	if got := reloaded.Config.Agents["builder"].Model; got != "gpt-5" {
		t.Fatalf("expected selected builder model to reload, got %q", got)
	}
	if got := reloaded.Config.Agents["legacy_agent"].Model; got != "legacy-only" {
		t.Fatalf("expected preserved legacy_agent model to reload, got %q", got)
	}

	if len(reloaded.PreservedUnknown) != 2 {
		t.Fatalf("expected only top-level unknown fragments to survive reload, got %d", len(reloaded.PreservedUnknown))
	}

	reloadedUnknownChecks := []struct {
		name  string
		check func(*testing.T)
	}{
		{
			name: "custom bundle survives round-trip",
			check: func(t *testing.T) {
				t.Helper()

				var customBundle map[string]any
				if err := json.Unmarshal(reloaded.PreservedUnknown["custom_bundle"], &customBundle); err != nil {
					t.Fatalf("failed to decode reloaded custom_bundle: %v", err)
				}
				if enabled, ok := customBundle["enabled"].(bool); !ok || !enabled {
					t.Fatalf("expected reloaded custom_bundle.enabled to remain true, got %#v", customBundle["enabled"])
				}
			},
		},
		{
			name: "custom flags survive round-trip",
			check: func(t *testing.T) {
				t.Helper()

				var customFlags []string
				if err := json.Unmarshal(reloaded.PreservedUnknown["custom_flags"], &customFlags); err != nil {
					t.Fatalf("failed to decode reloaded custom_flags: %v", err)
				}
				if !reflect.DeepEqual(customFlags, []string{"alpha", "beta"}) {
					t.Fatalf("expected reloaded custom_flags to remain [alpha beta], got %#v", customFlags)
				}
			},
		},
	}

	for _, tc := range reloadedUnknownChecks {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t)
		})
	}
}
