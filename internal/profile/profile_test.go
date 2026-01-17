package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	json.Unmarshal(data, &loaded)

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
	os.WriteFile(profilePath, []byte("{}"), 0644)

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
	os.WriteFile(filepath.Join(profilesDir, "profile1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(profilesDir, "profile2.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(profilesDir, "not-json.txt"), []byte("{}"), 0644)

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
	os.WriteFile(profilePath, []byte("{}"), 0644)

	if !Exists("exists-test") {
		t.Error("Expected profile to exist")
	}

	if Exists("nonexistent") {
		t.Error("Expected profile to not exist")
	}
}
