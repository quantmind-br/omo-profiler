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

func TestTmuxConfigRoundTrip(t *testing.T) {
	jsonData := `{
		"tmux": {
			"enabled": true,
			"layout": "main-horizontal",
			"main_pane_size": 60.0,
			"main_pane_min_width": 120.0,
			"agent_pane_min_width": 40.0
		}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Tmux == nil {
		t.Fatal("tmux is nil")
	}
	if cfg.Tmux.Enabled == nil || *cfg.Tmux.Enabled != true {
		t.Errorf("expected enabled=true, got %v", cfg.Tmux.Enabled)
	}
	if cfg.Tmux.Layout != "main-horizontal" {
		t.Errorf("expected layout=main-horizontal, got %s", cfg.Tmux.Layout)
	}
	if cfg.Tmux.MainPaneSize == nil || *cfg.Tmux.MainPaneSize != 60.0 {
		t.Errorf("expected main_pane_size=60.0, got %v", cfg.Tmux.MainPaneSize)
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

	if !reflect.DeepEqual(cfg.Tmux, roundtrip.Tmux) {
		t.Errorf("round-trip mismatch: original=%+v, roundtrip=%+v", cfg.Tmux, roundtrip.Tmux)
	}
}

func TestBrowserAutomationEngineRoundTrip(t *testing.T) {
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

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"browser_automation_engine"`) {
		t.Errorf("browser_automation_engine not in output: %s", marshaled)
	}
	if !strings.Contains(string(marshaled), `"provider":"playwright"`) {
		t.Errorf("provider not preserved: %s", marshaled)
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

func TestAgentConfigExtendedFields(t *testing.T) {
	jsonData := `{
		"agents": {
			"build": {
				"model": "claude-sonnet-4",
				"maxTokens": 8192,
				"thinking": {"type": "enabled", "budgetTokens": 4096},
				"reasoningEffort": "high",
				"textVerbosity": "medium",
				"providerOptions": {"stream": true, "cache": false}
			}
		}
	}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	agent := cfg.Agents["build"]
	if agent == nil {
		t.Fatal("build agent is nil")
	}

	// Verify MaxTokens
	if agent.MaxTokens == nil || *agent.MaxTokens != 8192 {
		t.Errorf("expected maxTokens=8192, got %v", agent.MaxTokens)
	}

	// Verify Thinking
	if agent.Thinking == nil {
		t.Fatal("thinking is nil")
	}
	if agent.Thinking.Type != "enabled" {
		t.Errorf("expected thinking.type=enabled, got %s", agent.Thinking.Type)
	}
	if agent.Thinking.BudgetTokens == nil || *agent.Thinking.BudgetTokens != 4096 {
		t.Errorf("expected thinking.budgetTokens=4096, got %v", agent.Thinking.BudgetTokens)
	}

	// Verify ReasoningEffort
	if agent.ReasoningEffort != "high" {
		t.Errorf("expected reasoningEffort=high, got %s", agent.ReasoningEffort)
	}

	// Verify TextVerbosity
	if agent.TextVerbosity != "medium" {
		t.Errorf("expected textVerbosity=medium, got %s", agent.TextVerbosity)
	}

	// Verify ProviderOptions
	if agent.ProviderOptions == nil {
		t.Fatal("providerOptions is nil")
	}
	if agent.ProviderOptions["stream"] != true {
		t.Errorf("expected providerOptions.stream=true, got %v", agent.ProviderOptions["stream"])
	}
	if agent.ProviderOptions["cache"] != false {
		t.Errorf("expected providerOptions.cache=false, got %v", agent.ProviderOptions["cache"])
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

	rtAgent := roundtrip.Agents["build"]
	if rtAgent == nil {
		t.Fatal("build agent not found after round-trip")
	}

	if !reflect.DeepEqual(agent.MaxTokens, rtAgent.MaxTokens) {
		t.Errorf("maxTokens mismatch after round-trip")
	}
	if !reflect.DeepEqual(agent.Thinking, rtAgent.Thinking) {
		t.Errorf("thinking mismatch after round-trip")
	}
	if agent.ReasoningEffort != rtAgent.ReasoningEffort {
		t.Errorf("reasoningEffort mismatch after round-trip")
	}
	if agent.TextVerbosity != rtAgent.TextVerbosity {
		t.Errorf("textVerbosity mismatch after round-trip")
	}
	if !reflect.DeepEqual(agent.ProviderOptions, rtAgent.ProviderOptions) {
		t.Errorf("providerOptions mismatch after round-trip")
	}
}
