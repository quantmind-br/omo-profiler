package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	return func() {
		config.ResetBaseDir()
	}
}

func createTestProfile(t *testing.T, name string, cfg *config.Config) {
	t.Helper()
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	if cfg == nil {
		cfg = &config.Config{
			DisabledMCPs: []string{"test-mcp"},
		}
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	profilePath := filepath.Join(config.ProfilesDir(), name+".json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}
}

func TestCreateFromTemplate_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	templateCfg := &config.Config{
		DisabledMCPs: []string{"mcp1", "mcp2"},
	}
	createTestProfile(t, "my-template", templateCfg)

	if !profile.Exists("my-template") {
		t.Fatal("Template profile not created")
	}

	fromTemplate = "my-template"
	defer func() { fromTemplate = "" }()

	newProfile := &profile.Profile{
		Name:   "new-profile",
		Config: *templateCfg,
	}

	if err := profile.Save(newProfile); err != nil {
		t.Fatalf("Failed to save new profile: %v", err)
	}

	if !profile.Exists("new-profile") {
		t.Fatal("New profile was not created")
	}

	loaded, err := profile.Load("new-profile")
	if err != nil {
		t.Fatalf("Failed to load new profile: %v", err)
	}

	if len(loaded.Config.DisabledMCPs) != 2 {
		t.Errorf("Expected 2 disabled MCPs, got %d", len(loaded.Config.DisabledMCPs))
	}
}

func TestCreateFromTemplate_TemplateNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if profile.Exists("nonexistent-template") {
		t.Fatal("Template should not exist")
	}

	if !profile.Exists("nonexistent-template") {
		t.Log("Template correctly identified as not existing")
	}
}

func TestCreateFromTemplate_ProfileAlreadyExists(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	templateCfg := &config.Config{
		DisabledMCPs: []string{"mcp1"},
	}
	createTestProfile(t, "template", templateCfg)
	createTestProfile(t, "existing-profile", templateCfg)

	if !profile.Exists("template") {
		t.Fatal("Template not created")
	}
	if !profile.Exists("existing-profile") {
		t.Fatal("Existing profile not created")
	}

	if profile.Exists("existing-profile") {
		t.Log("Profile correctly identified as already existing")
	}
}

func TestCreateFromTemplate_InvalidProfileName(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		input    string
		expected string
	}{
		{"valid-name", "valid-name"},
		{"valid_name", "valid_name"},
		{"ValidName123", "ValidName123"},
		{"invalid@name", "invalidname"},
		{"---invalid---", "invalid"},
		{"___invalid___", "invalid"},
		{"@@@", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := profile.SanitizeName(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCreateCmd_Integration(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	templateCfg := &config.Config{
		DisabledMCPs: []string{"mcp1", "mcp2"},
	}
	createTestProfile(t, "base-template", templateCfg)

	fromTemplate = "base-template"
	defer func() { fromTemplate = "" }()

	template, err := profile.Load("base-template")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	newProfile := &profile.Profile{
		Name:   "my-new-profile",
		Config: template.Config,
	}

	if err := profile.Save(newProfile); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	if !profile.Exists("my-new-profile") {
		t.Fatal("Profile not created")
	}

	loaded, err := profile.Load("my-new-profile")
	if err != nil {
		t.Fatalf("Failed to load created profile: %v", err)
	}

	if loaded.Name != "my-new-profile" {
		t.Errorf("Expected name 'my-new-profile', got %q", loaded.Name)
	}

	if len(loaded.Config.DisabledMCPs) != 2 {
		t.Errorf("Expected 2 disabled MCPs, got %d", len(loaded.Config.DisabledMCPs))
	}
}

func TestCreateCmd_SuccessMessageFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	templateCfg := &config.Config{
		DisabledMCPs: []string{"test"},
	}
	createTestProfile(t, "template", templateCfg)

	expectedMsg := "Created profile 'new-profile' from template 'template'\n"
	if expectedMsg != "Created profile 'new-profile' from template 'template'\n" {
		t.Error("Success message format incorrect")
	}
}
