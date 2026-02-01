package config

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	// Load valid-config.json
	data, err := os.ReadFile("../testdata/valid-config.json")
	if err != nil {
		t.Fatalf("failed to read valid-config.json: %v", err)
	}

	// Unmarshal into Config
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Marshal back to JSON
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Compare normalized JSON (unmarshal both to map and compare)
	var original, roundtrip map[string]interface{}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatalf("failed to unmarshal original: %v", err)
	}
	if err := json.Unmarshal(marshaled, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal roundtrip: %v", err)
	}

	if !reflect.DeepEqual(original, roundtrip) {
		origJSON, _ := json.MarshalIndent(original, "", "  ")
		rtJSON, _ := json.MarshalIndent(roundtrip, "", "  ")
		t.Errorf("round-trip mismatch:\noriginal:\n%s\n\nroundtrip:\n%s", origJSON, rtJSON)
	}
}

func TestMinimalConfig(t *testing.T) {
	data, err := os.ReadFile("../testdata/minimal-config.json")
	if err != nil {
		t.Fatalf("failed to read minimal-config.json: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify build agent exists with correct model
	if cfg.Agents == nil {
		t.Fatal("agents is nil")
	}
	build, ok := cfg.Agents["build"]
	if !ok {
		t.Fatal("build agent not found")
	}
	if build.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %s", build.Model)
	}
}

func TestSkillsPreservation_Array(t *testing.T) {
	// Test that json.RawMessage preserves array format
	jsonData := `{"skills": ["git-master", "playwright"]}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Skills == nil {
		t.Fatal("skills is nil")
	}

	// Re-marshal and check format is preserved
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(marshaled), `"skills":["git-master","playwright"]`) {
		t.Errorf("skills array not preserved: %s", marshaled)
	}
}

func TestSkillsPreservation_Object(t *testing.T) {
	// Test that json.RawMessage preserves object format
	jsonData := `{"skills": {"git-master": true, "playwright": {"description": "browser automation"}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Skills == nil {
		t.Fatal("skills is nil")
	}

	// Re-marshal and verify object structure is preserved
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should contain object format, not array
	if strings.Contains(string(marshaled), `"skills":[`) {
		t.Errorf("skills should be object, not array: %s", marshaled)
	}
	if !strings.Contains(string(marshaled), `"git-master":true`) {
		t.Errorf("skills object not preserved: %s", marshaled)
	}
}

func TestBashPermission_String(t *testing.T) {
	// Test bash permission as string
	jsonData := `{"agents": {"build": {"permission": {"bash": "allow"}}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	perm := cfg.Agents["build"].Permission
	if perm == nil {
		t.Fatal("permission is nil")
	}

	bash, ok := perm.Bash.(string)
	if !ok {
		t.Fatalf("bash should be string, got %T", perm.Bash)
	}
	if bash != "allow" {
		t.Errorf("expected bash=allow, got %s", bash)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"bash":"allow"`) {
		t.Errorf("bash string not preserved: %s", marshaled)
	}
}

func TestBashPermission_Object(t *testing.T) {
	// Test bash permission as object
	jsonData := `{"agents": {"build": {"permission": {"bash": {"git": "allow", "rm": "deny"}}}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	perm := cfg.Agents["build"].Permission
	if perm == nil {
		t.Fatal("permission is nil")
	}

	bashObj, ok := perm.Bash.(map[string]interface{})
	if !ok {
		t.Fatalf("bash should be map, got %T", perm.Bash)
	}
	if bashObj["git"] != "allow" {
		t.Errorf("expected git=allow, got %v", bashObj["git"])
	}
	if bashObj["rm"] != "deny" {
		t.Errorf("expected rm=deny, got %v", bashObj["rm"])
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"git":"allow"`) {
		t.Errorf("bash object not preserved: %s", marshaled)
	}
}

func TestEmptyConfigMarshal(t *testing.T) {
	// Empty Config should marshal to {}
	cfg := Config{}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal empty config: %v", err)
	}

	if string(marshaled) != "{}" {
		t.Errorf("empty config should marshal to {}, got: %s", marshaled)
	}
}

func TestOmitempty(t *testing.T) {
	// Config with only schema set should only have $schema in output
	cfg := Config{
		Schema: "https://example.com/schema.json",
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should only contain $schema
	expected := `{"$schema":"https://example.com/schema.json"}`
	if string(marshaled) != expected {
		t.Errorf("expected %s, got %s", expected, marshaled)
	}
}

func TestAgentSkillsRoundTrip(t *testing.T) {
	cfg := &Config{
		Agents: map[string]*AgentConfig{
			"build": {Skills: []string{"playwright", "git-master"}},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result Config
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.Agents == nil || result.Agents["build"] == nil {
		t.Fatal("build agent not found after round-trip")
	}

	skills := result.Agents["build"].Skills
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if skills[0] != "playwright" || skills[1] != "git-master" {
		t.Errorf("expected [playwright git-master], got %v", skills)
	}
}

func TestCategoryConfigDescriptionRoundTrip(t *testing.T) {
	jsonData := `{"categories": {"quick": {"model": "claude-haiku", "description": "Fast tasks"}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Categories == nil {
		t.Fatal("categories is nil")
	}
	quick, ok := cfg.Categories["quick"]
	if !ok {
		t.Fatal("quick category not found")
	}
	if quick.Description != "Fast tasks" {
		t.Errorf("expected description='Fast tasks', got %s", quick.Description)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"description":"Fast tasks"`) {
		t.Errorf("description not preserved: %s", marshaled)
	}
}

func TestCategoryConfigExtendedFields(t *testing.T) {
	jsonData := `{
		"categories": {
			"quick": {
				"model": "claude-sonnet-4",
				"maxTokens": 8192,
				"thinking": {"type": "enabled", "budgetTokens": 4096},
				"reasoningEffort": "high",
				"textVerbosity": "medium"
			}
		}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	cat := cfg.Categories["quick"]
	if cat == nil {
		t.Fatal("quick category is nil")
	}

	// Verify MaxTokens
	if cat.MaxTokens == nil || *cat.MaxTokens != 8192 {
		t.Errorf("expected maxTokens=8192, got %v", cat.MaxTokens)
	}

	// Verify Thinking
	if cat.Thinking == nil {
		t.Fatal("thinking is nil")
	}
	if cat.Thinking.Type != "enabled" {
		t.Errorf("expected thinking.type=enabled, got %s", cat.Thinking.Type)
	}
	if cat.Thinking.BudgetTokens == nil || *cat.Thinking.BudgetTokens != 4096 {
		t.Errorf("expected thinking.budgetTokens=4096, got %v", cat.Thinking.BudgetTokens)
	}

	// Verify ReasoningEffort
	if cat.ReasoningEffort != "high" {
		t.Errorf("expected reasoningEffort=high, got %s", cat.ReasoningEffort)
	}

	// Verify TextVerbosity
	if cat.TextVerbosity != "medium" {
		t.Errorf("expected textVerbosity=medium, got %s", cat.TextVerbosity)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var roundtrip Config
	if err := json.Unmarshal(marshaled, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal roundtrip: %v", err)
	}

	rtCat := roundtrip.Categories["quick"]
	if rtCat == nil {
		t.Fatal("roundtrip quick category is nil")
	}

	if rtCat.MaxTokens == nil || *rtCat.MaxTokens != *cat.MaxTokens {
		t.Errorf("roundtrip MaxTokens mismatch")
	}
	if !reflect.DeepEqual(cat.Thinking, rtCat.Thinking) {
		t.Errorf("roundtrip Thinking mismatch")
	}
	if cat.ReasoningEffort != rtCat.ReasoningEffort {
		t.Errorf("roundtrip ReasoningEffort mismatch")
	}
	if cat.TextVerbosity != rtCat.TextVerbosity {
		t.Errorf("roundtrip TextVerbosity mismatch")
	}
}

func TestBabysittingConfigRoundTrip(t *testing.T) {
	jsonData := `{"babysitting": {"timeout_ms": 120000}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Babysitting == nil {
		t.Fatal("babysitting is nil")
	}
	if cfg.Babysitting.TimeoutMs == nil || *cfg.Babysitting.TimeoutMs != 120000 {
		t.Errorf("expected timeout_ms=120000, got %v", cfg.Babysitting.TimeoutMs)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"timeout_ms":120000`) {
		t.Errorf("timeout_ms not preserved: %s", marshaled)
	}
}

func TestBrowserAutomationEngineConfigRoundTrip(t *testing.T) {
	jsonData := `{"browser_automation_engine": {"provider": "playwright"}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.BrowserAutomationEngine == nil {
		t.Fatal("browser_automation_engine is nil")
	}
	if cfg.BrowserAutomationEngine.Provider != "playwright" {
		t.Errorf("expected provider=playwright, got %s", cfg.BrowserAutomationEngine.Provider)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"provider":"playwright"`) {
		t.Errorf("provider not preserved: %s", marshaled)
	}
}

func TestTmuxConfigRoundTrip(t *testing.T) {
	jsonData := `{"tmux": {"enabled": true, "layout": "main-vertical", "main_pane_size": 60}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Tmux == nil {
		t.Fatal("tmux is nil")
	}
	if cfg.Tmux.Enabled == nil || !*cfg.Tmux.Enabled {
		t.Error("expected enabled=true")
	}
	if cfg.Tmux.Layout != "main-vertical" {
		t.Errorf("expected layout=main-vertical, got %s", cfg.Tmux.Layout)
	}
	if cfg.Tmux.MainPaneSize == nil || *cfg.Tmux.MainPaneSize != 60 {
		t.Errorf("expected main_pane_size=60, got %v", cfg.Tmux.MainPaneSize)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"enabled":true`) {
		t.Errorf("enabled not preserved: %s", marshaled)
	}
}

func TestSisyphusConfigRoundTrip(t *testing.T) {
	jsonData := `{"sisyphus": {"tasks": {"storage_path": ".sisyphus/tasks", "claude_code_compat": false}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Sisyphus == nil || cfg.Sisyphus.Tasks == nil {
		t.Fatal("sisyphus or tasks is nil")
	}
	if cfg.Sisyphus.Tasks.StoragePath != ".sisyphus/tasks" {
		t.Errorf("expected storage_path=.sisyphus/tasks, got %s", cfg.Sisyphus.Tasks.StoragePath)
	}
	if cfg.Sisyphus.Tasks.ClaudeCodeCompat == nil || *cfg.Sisyphus.Tasks.ClaudeCodeCompat != false {
		t.Errorf("expected claude_code_compat=false, got %v", cfg.Sisyphus.Tasks.ClaudeCodeCompat)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"storage_path":".sisyphus/tasks"`) {
		t.Errorf("storage_path not preserved: %s", marshaled)
	}
}

func TestAgentConfigExtendedFieldsRoundTrip(t *testing.T) {
	jsonData := `{
		"agents": {
			"build": {
				"model": "claude-sonnet",
				"maxTokens": 8192,
				"thinking": {"type": "enabled", "budgetTokens": 4096},
				"reasoningEffort": "medium",
				"textVerbosity": "low",
				"providerOptions": {"custom_option": true}
			}
		}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	build := cfg.Agents["build"]
	if build == nil {
		t.Fatal("build agent is nil")
	}

	if build.MaxTokens == nil || *build.MaxTokens != 8192 {
		t.Errorf("expected maxTokens=8192, got %v", build.MaxTokens)
	}
	if build.Thinking == nil || build.Thinking.Type != "enabled" {
		t.Error("expected thinking.type=enabled")
	}
	if build.ReasoningEffort != "medium" {
		t.Errorf("expected reasoningEffort=medium, got %s", build.ReasoningEffort)
	}
	if build.TextVerbosity != "low" {
		t.Errorf("expected textVerbosity=low, got %s", build.TextVerbosity)
	}
	if build.ProviderOptions == nil {
		t.Error("expected providerOptions to be set")
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"maxTokens":8192`) {
		t.Errorf("maxTokens not preserved: %s", marshaled)
	}
}
