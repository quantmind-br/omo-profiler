package profile

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

func TestSparseSerializerOmitsUncheckedFields(t *testing.T) {
	cfg := &config.Config{
		DisabledHooks: []string{"pre-commit"},
		Agents: map[string]*config.AgentConfig{
			"builder": {Model: "gpt-5"},
		},
	}

	selection := NewBlankSelection()
	selection.SetSelected("disabled_hooks", true)

	data, err := MarshalSparse(cfg, selection, nil)
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	if _, ok := decoded["disabled_hooks"]; !ok {
		t.Fatal("expected disabled_hooks to be present")
	}
	if _, ok := decoded["agents"]; ok {
		t.Fatal("expected agents to be omitted")
	}
}

func TestSparseSerializerKeepsExplicitZeroValuesWhenSelected(t *testing.T) {
	cfg := &config.Config{
		HashlineEdit:  boolPtr(false),
		ModelFallback: boolPtr(false),
		DisabledHooks: []string{},
		Agents: map[string]*config.AgentConfig{
			"builder": {Tools: map[string]bool{}},
		},
	}

	selection := NewBlankSelection()
	selection.SetSelected("hashline_edit", true)
	selection.SetSelected("model_fallback", true)
	selection.SetSelected("disabled_hooks", true)
	selection.SetSelected("agents.*.tools", true)

	data, err := MarshalSparse(cfg, selection, nil)
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)

	if value, ok := decoded["hashline_edit"].(bool); !ok || value {
		t.Fatalf("expected hashline_edit to be false, got %#v", decoded["hashline_edit"])
	}

	if value, ok := decoded["model_fallback"].(bool); !ok || value {
		t.Fatalf("expected model_fallback to be false, got %#v", decoded["model_fallback"])
	}

	hooks, ok := decoded["disabled_hooks"].([]interface{})
	if !ok {
		t.Fatalf("expected disabled_hooks to be an empty array, got %#v", decoded["disabled_hooks"])
	}
	if len(hooks) != 0 {
		t.Fatalf("expected disabled_hooks to be empty, got %#v", hooks)
	}

	agents, ok := decoded["agents"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected agents object, got %#v", decoded["agents"])
	}
	builder, ok := agents["builder"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected builder agent object, got %#v", agents["builder"])
	}
	tools, ok := builder["tools"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tools object, got %#v", builder["tools"])
	}
	if len(tools) != 0 {
		t.Fatalf("expected tools to be empty, got %#v", tools)
	}
}

func TestSparseSerializerMergesPreservedUnknownFragments(t *testing.T) {
	cfg := &config.Config{
		DisabledHooks: []string{"pre-commit"},
	}

	selection := NewBlankSelection()
	selection.SetSelected("disabled_hooks", true)

	data, err := MarshalSparse(cfg, selection, map[string]json.RawMessage{
		"customField": json.RawMessage(`{"enabled":true}`),
	})
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	if _, ok := decoded["disabled_hooks"]; !ok {
		t.Fatal("expected disabled_hooks to be present")
	}

	customField, ok := decoded["customField"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected customField object, got %#v", decoded["customField"])
	}
	if enabled, ok := customField["enabled"].(bool); !ok || !enabled {
		t.Fatalf("expected customField.enabled to be true, got %#v", customField["enabled"])
	}
}

func TestSparseSerializerKnownPathsOverridePreservedUnknownFragments(t *testing.T) {
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"builder": {Model: "gpt-5"},
		},
		Experimental: &config.ExperimentalConfig{
			TaskSystem: boolPtr(false),
		},
	}

	selection := NewBlankSelection()
	selection.SetSelected("agents.*.model", true)
	selection.SetSelected("experimental.task_system", true)

	data, err := MarshalSparse(cfg, selection, map[string]json.RawMessage{
		"agents":       json.RawMessage(`{"builder":{"model":"legacy-model","legacy":true},"legacy_agent":{"model":"legacy-only"}}`),
		"experimental": json.RawMessage(`{"task_system":true,"legacy_flag":true}`),
	})
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	agents := decodedObject(t, decoded["agents"], "agents")
	builder := decodedObject(t, agents["builder"], "agents.builder")

	if model, ok := builder["model"].(string); !ok || model != "gpt-5" {
		t.Fatalf("expected selected builder model to win, got %#v", builder["model"])
	}
	if legacy, ok := builder["legacy"].(bool); !ok || !legacy {
		t.Fatalf("expected preserved sibling field on builder to remain, got %#v", builder["legacy"])
	}

	legacyAgent := decodedObject(t, agents["legacy_agent"], "agents.legacy_agent")
	if model, ok := legacyAgent["model"].(string); !ok || model != "legacy-only" {
		t.Fatalf("expected preserved legacy_agent to remain, got %#v", legacyAgent["model"])
	}

	experimental := decodedObject(t, decoded["experimental"], "experimental")
	if taskSystem, ok := experimental["task_system"].(bool); !ok || taskSystem {
		t.Fatalf("expected selected experimental.task_system to win, got %#v", experimental["task_system"])
	}
	if legacyFlag, ok := experimental["legacy_flag"].(bool); !ok || !legacyFlag {
		t.Fatalf("expected preserved experimental.legacy_flag to remain, got %#v", experimental["legacy_flag"])
	}
}

func TestSparseSerializerProducesStablePrettyJSON(t *testing.T) {
	cfg := &config.Config{
		DefaultRunAgent: "",
		DisabledHooks:   []string{},
		Agents: map[string]*config.AgentConfig{
			"zeta":  {Model: "gpt-5"},
			"alpha": {Model: "gpt-4"},
		},
	}

	selection := NewBlankSelection()
	selection.SetSelected("default_run_agent", true)
	selection.SetSelected("disabled_hooks", true)
	selection.SetSelected("agents.*.model", true)

	preservedUnknown := map[string]json.RawMessage{
		"custom_field": json.RawMessage(`{"z":1,"a":2}`),
	}

	first, err := MarshalSparse(cfg, selection, preservedUnknown)
	if err != nil {
		t.Fatalf("first MarshalSparse failed: %v", err)
	}
	second, err := MarshalSparse(cfg, selection, preservedUnknown)
	if err != nil {
		t.Fatalf("second MarshalSparse failed: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("expected stable JSON output\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	want := "{\n  \"agents\": {\n    \"alpha\": {\n      \"model\": \"gpt-4\"\n    },\n    \"zeta\": {\n      \"model\": \"gpt-5\"\n    }\n  },\n  \"custom_field\": {\n    \"a\": 2,\n    \"z\": 1\n  },\n  \"default_run_agent\": \"\",\n  \"disabled_hooks\": []\n}"
	if string(first) != want {
		t.Fatalf("unexpected pretty JSON\n got:\n%s\nwant:\n%s", first, want)
	}
}

func TestSparseSerializerOmitsEmptyParentObjects(t *testing.T) {
	cfg := &config.Config{
		Experimental: &config.ExperimentalConfig{
			TaskSystem: boolPtr(false),
			MaxTools:   int64Ptr(0),
		},
	}

	selection := NewBlankSelection()
	selection.SetSelected("experimental.task_system", true)

	data, err := MarshalSparse(cfg, selection, nil)
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	if len(decoded) != 1 {
		t.Fatalf("expected only experimental at the top level, got %#v", decoded)
	}

	experimental := decodedObject(t, decoded["experimental"], "experimental")
	want := map[string]interface{}{"task_system": false}
	if !reflect.DeepEqual(experimental, want) {
		t.Fatalf("unexpected experimental payload: got %#v want %#v", experimental, want)
	}
}

func decodeSparseJSON(t *testing.T, data []byte) map[string]interface{} {
	t.Helper()

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode sparse JSON: %v\n%s", err, data)
	}
	return decoded
}

func decodedObject(t *testing.T, value interface{}, name string) map[string]interface{} {
	t.Helper()

	decoded, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("expected %s to be an object, got %#v", name, value)
	}
	return decoded
}

func boolPtr(v bool) *bool {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
