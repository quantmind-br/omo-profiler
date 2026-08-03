package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadDocument_MissingFile(t *testing.T) {
	defer ResetBaseDir()
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	doc, err := LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument missing file: %v", err)
	}
	if doc.Exists {
		t.Errorf("Exists = true, want false")
	}
	if doc.Path != OmoFile() {
		t.Errorf("Path = %s, want %s", doc.Path, OmoFile())
	}
}

func TestParseDocument_JSONCWithComments(t *testing.T) {
	src := []byte(`{
  // user layer
  "$schema": "https://example.com/omo.schema.json",
  "profiles": {
    /* one profile */
    "dev": {"[opencode]": {"telemetry": false}},
  },
}`)

	doc, err := ParseDocument(src)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}

	raw, ok := doc.Raw(SchemaKey)
	if !ok {
		t.Fatal("$schema missing")
	}
	var schema string
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	if schema != "https://example.com/omo.schema.json" {
		t.Errorf("schema = %q", schema)
	}

	block, ok, err := doc.ProfileBlock("dev")
	if err != nil {
		t.Fatalf("ProfileBlock: %v", err)
	}
	if !ok {
		t.Fatal("dev profile missing")
	}
	var got map[string]any
	if err := json.Unmarshal(block, &got); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	opencode, ok := got[OpenCodeKey].(map[string]any)
	if !ok {
		t.Fatalf("profile missing %q: %#v", OpenCodeKey, got)
	}
	if opencode["telemetry"] != false {
		t.Errorf("telemetry = %#v, want false", opencode["telemetry"])
	}
}

func TestProfileAddReplaceDeleteRoundTrip(t *testing.T) {
	defer ResetBaseDir()
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	doc := NewDocument()

	dev := json.RawMessage(`{"[opencode]":{"telemetry":false}}`)
	if err := doc.SetProfileBlock("dev", dev); err != nil {
		t.Fatalf("SetProfileBlock add: %v", err)
	}
	if !doc.HasProfile("dev") {
		t.Fatal("HasProfile(dev) = false after add")
	}

	replaced := json.RawMessage(`{"[opencode]":{"telemetry":true,"agents":{}}}`)
	if err := doc.SetProfileBlock("dev", replaced); err != nil {
		t.Fatalf("SetProfileBlock replace: %v", err)
	}
	block, ok, err := doc.ProfileBlock("dev")
	if err != nil || !ok {
		t.Fatalf("ProfileBlock after replace: ok=%v err=%v", ok, err)
	}
	var got map[string]any
	if err := json.Unmarshal(block, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	opencode := got[OpenCodeKey].(map[string]any)
	if opencode["telemetry"] != true {
		t.Errorf("telemetry after replace = %#v", opencode["telemetry"])
	}

	prod := json.RawMessage(`{"[opencode]":{"telemetry":true}}`)
	if err := doc.SetProfileBlock("prod", prod); err != nil {
		t.Fatalf("SetProfileBlock prod: %v", err)
	}

	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if !reloaded.Exists {
		t.Fatal("reloaded Exists = false")
	}
	if !reloaded.HasProfile("dev") || !reloaded.HasProfile("prod") {
		t.Fatal("profiles missing after reload")
	}

	deleted, err := reloaded.DeleteProfileBlock("dev")
	if err != nil {
		t.Fatalf("DeleteProfileBlock: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteProfileBlock reported false for existing profile")
	}
	if reloaded.HasProfile("dev") {
		t.Fatal("dev still present after delete")
	}
	if !reloaded.HasProfile("prod") {
		t.Fatal("prod should remain after deleting dev")
	}

	if err := reloaded.Save(); err != nil {
		t.Fatalf("Save after delete: %v", err)
	}
	again, err := LoadDocument()
	if err != nil {
		t.Fatalf("reload after delete: %v", err)
	}
	if again.HasProfile("dev") {
		t.Fatal("dev reappeared after save")
	}
	if !again.HasProfile("prod") {
		t.Fatal("prod missing after save")
	}
}

func TestDeleteProfileBlock_Absent(t *testing.T) {
	doc := NewDocument()
	if err := doc.SetProfileBlock("only", json.RawMessage(`{}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}

	deleted, err := doc.DeleteProfileBlock("missing")
	if err != nil {
		t.Fatalf("DeleteProfileBlock: %v", err)
	}
	if deleted {
		t.Fatal("DeleteProfileBlock reported true for absent profile")
	}
	if !doc.HasProfile("only") {
		t.Fatal("existing profile was removed")
	}
}

func TestDeleteProfileBlock_RemovesProfilesKeyWhenEmpty(t *testing.T) {
	doc := NewDocument()
	if err := doc.SetProfileBlock("solo", json.RawMessage(`{"x":1}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}

	deleted, err := doc.DeleteProfileBlock("solo")
	if err != nil {
		t.Fatalf("DeleteProfileBlock: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete of last profile")
	}
	if _, ok := doc.Raw(ProfilesKey); ok {
		t.Fatal("profiles key should be removed when last profile is deleted")
	}

	data, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("unmarshal bytes: %v", err)
	}
	if _, ok := top[ProfilesKey]; ok {
		t.Fatalf("serialized document still has %q: %s", ProfilesKey, data)
	}
}

func TestProfileNames_Sorted(t *testing.T) {
	doc := NewDocument()
	for _, name := range []string{"zeta", "alpha", "mu"} {
		if err := doc.SetProfileBlock(name, json.RawMessage(`{}`)); err != nil {
			t.Fatalf("SetProfileBlock(%s): %v", name, err)
		}
	}

	names, err := doc.ProfileNames()
	if err != nil {
		t.Fatalf("ProfileNames: %v", err)
	}
	want := []string{"alpha", "mu", "zeta"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("ProfileNames() = %v, want %v", names, want)
	}
}

func TestEnsureSchema(t *testing.T) {
	t.Run("sets when absent", func(t *testing.T) {
		doc := NewDocument()
		doc.EnsureSchema()
		raw, ok := doc.Raw(SchemaKey)
		if !ok {
			t.Fatal("$schema not set")
		}
		var got string
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got != DefaultSchema {
			t.Errorf("schema = %q, want %q", got, DefaultSchema)
		}
	})

	t.Run("leaves existing alone", func(t *testing.T) {
		doc := NewDocument()
		custom, _ := json.Marshal("https://example.com/custom.schema.json")
		doc.SetRaw(SchemaKey, custom)
		doc.EnsureSchema()
		raw, _ := doc.Raw(SchemaKey)
		var got string
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got != "https://example.com/custom.schema.json" {
			t.Errorf("schema overwritten: %q", got)
		}
	})
}

func TestDocument_PreservesUnknownTopLevelKeys(t *testing.T) {
	defer ResetBaseDir()
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	// Seed a document that mixes modelled (profiles/$schema) keys with
	// harness blocks and free-form top-level keys the profiler does not own.
	original := map[string]json.RawMessage{
		SchemaKey:        json.RawMessage(`"https://example.com/omo.schema.json"`),
		OpenCodeKey:      json.RawMessage(`{"shared":true}`),
		SenpiKey:         json.RawMessage(`{"mode":"senpi-only","n":1}`),
		CodexKey:         json.RawMessage(`{"instructions":"keep me"}`),
		"task":           json.RawMessage(`{"enabled":true,"queue":["a","b"]}`),
		"teams":          json.RawMessage(`[{"name":"alpha"},{"name":"beta"}]`),
		"future_key_xyz": json.RawMessage(`{"nested":{"v":42},"arr":[1,2,3]}`),
	}

	doc := NewDocument()
	for key, value := range original {
		doc.SetRaw(key, value)
	}
	if err := doc.SetProfileBlock("dev", json.RawMessage(`{"[opencode]":{"telemetry":false}}`)); err != nil {
		t.Fatalf("SetProfileBlock seed: %v", err)
	}
	if err := doc.Save(); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	loaded, err := LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}

	// Touch a modelled field so a buggy document layer that rebuilds from a
	// typed struct would drop unknowns on the subsequent save.
	if err := loaded.SetProfileBlock("dev", json.RawMessage(`{"[opencode]":{"telemetry":true}}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}
	loaded.EnsureSchema() // must not wipe sibling keys either
	if err := loaded.Save(); err != nil {
		t.Fatalf("Save after edit: %v", err)
	}

	roundtrip, err := LoadDocument()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}

	unknownKeys := []string{SenpiKey, CodexKey, "task", "teams", "future_key_xyz", OpenCodeKey}
	for _, key := range unknownKeys {
		want := original[key]
		got, ok := roundtrip.Raw(key)
		if !ok {
			t.Fatalf("unknown key %q dropped after load/save round-trip", key)
		}
		if !jsonRawEqual(t, got, want) {
			t.Errorf("key %q changed\ngot:  %s\nwant: %s", key, got, want)
		}
	}

	// Profile edit must have stuck without affecting unknowns.
	block, ok, err := roundtrip.ProfileBlock("dev")
	if err != nil || !ok {
		t.Fatalf("dev profile missing after round-trip: ok=%v err=%v", ok, err)
	}
	var profile map[string]any
	if err := json.Unmarshal(block, &profile); err != nil {
		t.Fatalf("unmarshal profile: %v", err)
	}
	opencode := profile[OpenCodeKey].(map[string]any)
	if opencode["telemetry"] != true {
		t.Errorf("edited telemetry not preserved: %#v", opencode["telemetry"])
	}

	// On-disk bytes must still contain the unknown keys as literal JSON members.
	data, err := os.ReadFile(OmoFile())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	for _, key := range []string{`"[senpi]"`, `"[codex]"`, `"task"`, `"teams"`, `"future_key_xyz"`} {
		if !bytes.Contains(data, []byte(key)) {
			t.Errorf("on-disk file missing %s\n%s", key, data)
		}
	}
}

func TestLoadDocumentFrom_JSONCFile(t *testing.T) {
	defer ResetBaseDir()
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	path := filepath.Join(OmoDir(), OmoBasenameJSONC)
	content := []byte(`{
  // jsonc document on disk
  "profiles": {
    "dev": {}
  },
}`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	doc, err := LoadDocumentFrom(path)
	if err != nil {
		t.Fatalf("LoadDocumentFrom: %v", err)
	}
	if !doc.Exists {
		t.Fatal("Exists = false")
	}
	if !doc.HasProfile("dev") {
		t.Fatal("dev profile not parsed from jsonc")
	}
	// Preference rule: with omo.jsonc present, OmoFile must point at it.
	if OmoFile() != path {
		t.Errorf("OmoFile() = %s, want %s", OmoFile(), path)
	}
}

func jsonRawEqual(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatalf("unmarshal a: %v (%s)", err, a)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatalf("unmarshal b: %v (%s)", err, b)
	}
	return reflect.DeepEqual(x, y)
}
