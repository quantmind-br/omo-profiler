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
