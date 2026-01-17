package profile

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	t.Cleanup(config.ResetBaseDir)
}

func TestGetActive_NoConfig(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	active, err := GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if active.Exists {
		t.Errorf("expected Exists=false when no config")
	}
}

func TestGetActive_MatchingProfile(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create a profile
	testConfig := config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "test-model"},
		},
	}
	p := &Profile{Name: "dev", Config: testConfig}
	if err := Save(p); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// Set it as active
	if err := SetActive("dev"); err != nil {
		t.Fatalf("failed to set active: %v", err)
	}

	// Get active
	active, err := GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if !active.Exists {
		t.Errorf("expected Exists=true")
	}
	if active.IsOrphan {
		t.Errorf("expected IsOrphan=false")
	}
	if active.ProfileName != "dev" {
		t.Errorf("expected ProfileName='dev', got '%s'", active.ProfileName)
	}
}

func TestGetActive_OrphanConfig(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create a config file directly (not from a profile)
	orphanConfig := config.Config{
		Agents: map[string]*config.AgentConfig{
			"oracle": {Model: "custom-model"},
		},
	}
	data, _ := json.MarshalIndent(orphanConfig, "", "  ")
	if err := os.WriteFile(config.ConfigFile(), data, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Get active
	active, err := GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if !active.Exists {
		t.Errorf("expected Exists=true")
	}
	if !active.IsOrphan {
		t.Errorf("expected IsOrphan=true")
	}
	if active.ProfileName != "(custom)" {
		t.Errorf("expected ProfileName='(custom)', got '%s'", active.ProfileName)
	}
}

func TestSetActive(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create a profile
	testConfig := config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "my-model"},
		},
	}
	p := &Profile{Name: "production", Config: testConfig}
	if err := Save(p); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// Set it active
	if err := SetActive("production"); err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Verify config file exists and matches
	data, err := os.ReadFile(config.ConfigFile())
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loaded config.Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if loaded.Agents["build"].Model != "my-model" {
		t.Errorf("expected model 'my-model', got '%s'", loaded.Agents["build"].Model)
	}
}

func TestSetActive_NotFound(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	err := SetActive("nonexistent")
	if err == nil {
		t.Errorf("expected error for non-existent profile")
	}
}

func TestMatchesConfig_SchemaDifference(t *testing.T) {
	setupTestDir(t)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Create a profile without schema
	cfg := config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "test"},
		},
	}
	p := &Profile{Name: "test", Config: cfg}

	// Config with schema should still match
	cfgWithSchema := cfg
	cfgWithSchema.Schema = "https://example.com/schema.json"

	if !p.MatchesConfig(&cfgWithSchema) {
		t.Errorf("configs should match (schema is ignored)")
	}
}

func TestMatchesProfile(t *testing.T) {
	setupTestDir(t)

	p := &Profile{
		Name: "test",
		Config: config.Config{
			Agents: map[string]*config.AgentConfig{
				"build": {Model: "model-a"},
			},
		},
	}

	// Matching config
	matching := config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "model-a"},
		},
	}

	if !p.MatchesConfig(&matching) {
		t.Errorf("expected configs to match")
	}

	// Non-matching config
	different := config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "model-b"},
		},
	}

	if p.MatchesConfig(&different) {
		t.Errorf("expected configs to not match")
	}
}

func TestNormalizeForComparison(t *testing.T) {
	// Test that $schema is removed
	cfg := &config.Config{
		Schema: "https://example.com/schema.json",
		Agents: map[string]*config.AgentConfig{
			"build": {Model: "test"},
		},
	}

	data, err := normalizeForComparison(cfg)
	if err != nil {
		t.Fatalf("normalizeForComparison failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := result["$schema"]; ok {
		t.Errorf("$schema should be removed in normalized output")
	}
}
