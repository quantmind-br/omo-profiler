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

func TestPreemptiveCompactionRoundTrip(t *testing.T) {
	// Test PreemptiveCompaction with true value
	jsonData := `{"experimental": {"preemptive_compaction": true}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.PreemptiveCompaction == nil || !*cfg.Experimental.PreemptiveCompaction {
		t.Errorf("expected preemptive_compaction=true, got %v", cfg.Experimental.PreemptiveCompaction)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"preemptive_compaction":true`) {
		t.Errorf("preemptive_compaction not preserved: %s", marshaled)
	}

	// Verify round-trip equality
	var roundtrip Config
	if err := json.Unmarshal(marshaled, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal roundtrip: %v", err)
	}
	if roundtrip.Experimental == nil || roundtrip.Experimental.PreemptiveCompaction == nil {
		t.Fatal("roundtrip experimental or preemptive_compaction is nil")
	}
	if *roundtrip.Experimental.PreemptiveCompaction != *cfg.Experimental.PreemptiveCompaction {
		t.Errorf("roundtrip mismatch: expected %v, got %v", *cfg.Experimental.PreemptiveCompaction, *roundtrip.Experimental.PreemptiveCompaction)
	}
}

func TestPreemptiveCompactionFalse(t *testing.T) {
	// Test PreemptiveCompaction with false value
	jsonData := `{"experimental": {"preemptive_compaction": false}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.PreemptiveCompaction == nil || *cfg.Experimental.PreemptiveCompaction {
		t.Errorf("expected preemptive_compaction=false, got %v", cfg.Experimental.PreemptiveCompaction)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"preemptive_compaction":false`) {
		t.Errorf("preemptive_compaction=false not preserved: %s", marshaled)
	}
}

func TestPreemptiveCompactionOmitempty(t *testing.T) {
	// Test that nil PreemptiveCompaction is omitted from JSON
	cfg := Config{
		Experimental: &ExperimentalConfig{
			PreemptiveCompaction: nil,
		},
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if strings.Contains(string(marshaled), "preemptive_compaction") {
		t.Errorf("nil preemptive_compaction should be omitted: %s", marshaled)
	}
}

func TestTaskSystemRoundTrip(t *testing.T) {
	// Test TaskSystem with true value
	jsonData := `{"experimental": {"task_system": true}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.TaskSystem == nil || !*cfg.Experimental.TaskSystem {
		t.Errorf("expected task_system=true, got %v", cfg.Experimental.TaskSystem)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"task_system":true`) {
		t.Errorf("task_system not preserved: %s", marshaled)
	}

	// Verify round-trip equality
	var roundtrip Config
	if err := json.Unmarshal(marshaled, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal roundtrip: %v", err)
	}
	if roundtrip.Experimental == nil || roundtrip.Experimental.TaskSystem == nil {
		t.Fatal("roundtrip experimental or task_system is nil")
	}
	if *roundtrip.Experimental.TaskSystem != *cfg.Experimental.TaskSystem {
		t.Errorf("roundtrip mismatch: expected %v, got %v", *cfg.Experimental.TaskSystem, *roundtrip.Experimental.TaskSystem)
	}
}

func TestTaskSystemFalse(t *testing.T) {
	// Test TaskSystem with false value
	jsonData := `{"experimental": {"task_system": false}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.TaskSystem == nil || *cfg.Experimental.TaskSystem {
		t.Errorf("expected task_system=false, got %v", cfg.Experimental.TaskSystem)
	}

	// Round-trip
	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"task_system":false`) {
		t.Errorf("task_system=false not preserved: %s", marshaled)
	}
}

func TestTaskSystemOmitempty(t *testing.T) {
	// Test that nil TaskSystem is omitted from JSON
	cfg := Config{
		Experimental: &ExperimentalConfig{
			TaskSystem: nil,
		},
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if strings.Contains(string(marshaled), "task_system") {
		t.Errorf("nil task_system should be omitted: %s", marshaled)
	}
}

func TestUltraworkConfigRoundTrip(t *testing.T) {
	jsonData := `{
		"agents": {
			"build": {
				"model": "claude-sonnet",
				"ultrawork": {"model": "claude-sonnet-4-20250514", "variant": "fast"}
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
	if build.Ultrawork == nil {
		t.Fatal("ultrawork is nil")
	}
	if build.Ultrawork.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected ultrawork.model=claude-sonnet-4-20250514, got %s", build.Ultrawork.Model)
	}
	if build.Ultrawork.Variant != "fast" {
		t.Errorf("expected ultrawork.variant=fast, got %s", build.Ultrawork.Variant)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"ultrawork"`) {
		t.Errorf("ultrawork not preserved: %s", marshaled)
	}
	if !strings.Contains(string(marshaled), `"model":"claude-sonnet-4-20250514"`) {
		t.Errorf("ultrawork.model not preserved: %s", marshaled)
	}
}

func TestUltraworkConfigOmitempty(t *testing.T) {
	cfg := Config{
		Agents: map[string]*AgentConfig{
			"build": {Model: "test"},
		},
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if strings.Contains(string(marshaled), "ultrawork") {
		t.Errorf("nil ultrawork should be omitted: %s", marshaled)
	}
}

func TestHashlineEditRoundTrip(t *testing.T) {
	jsonData := `{"experimental": {"hashline_edit": true}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.HashlineEdit == nil || !*cfg.Experimental.HashlineEdit {
		t.Errorf("expected hashline_edit=true, got %v", cfg.Experimental.HashlineEdit)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"hashline_edit":true`) {
		t.Errorf("hashline_edit not preserved: %s", marshaled)
	}
}

func TestHashlineEditFalse(t *testing.T) {
	jsonData := `{"experimental": {"hashline_edit": false}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if cfg.Experimental == nil {
		t.Fatal("experimental is nil")
	}
	if cfg.Experimental.HashlineEdit == nil || *cfg.Experimental.HashlineEdit {
		t.Errorf("expected hashline_edit=false, got %v", cfg.Experimental.HashlineEdit)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if !strings.Contains(string(marshaled), `"hashline_edit":false`) {
		t.Errorf("hashline_edit=false not preserved: %s", marshaled)
	}
}

func TestHashlineEditOmitempty(t *testing.T) {
	cfg := Config{
		Experimental: &ExperimentalConfig{
			HashlineEdit: nil,
		},
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if strings.Contains(string(marshaled), "hashline_edit") {
		t.Errorf("nil hashline_edit should be omitted: %s", marshaled)
	}
}

func TestTaskListIDOmitempty(t *testing.T) {
	// Test that empty TaskListID is omitted from JSON
	jsonData := `{"sisyphus": {"tasks": {"storage_path": ".sisyphus/tasks"}}}`

	var cfg Config
	if err := json.Unmarshal([]byte(jsonData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	marshaled, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Empty task_list_id should be omitted
	if strings.Contains(string(marshaled), "task_list_id") {
		t.Errorf("empty task_list_id should be omitted: %s", marshaled)
	}
}
