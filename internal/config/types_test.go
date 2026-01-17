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
