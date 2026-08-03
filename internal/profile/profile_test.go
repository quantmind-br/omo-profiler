package profile

import (
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	config.SetBaseDir(t.TempDir())
	t.Cleanup(config.ResetBaseDir)
}

// seedProfile writes profiles.<name>.[opencode] into the user-layer omo document
// via the Document persistence contract (replacing the old file-per-profile seed).
func seedProfile(t *testing.T, name string, openCodeJSON string) {
	t.Helper()
	seedProfileBlock(t, name, mustProfileBlock(t, json.RawMessage(openCodeJSON)))
}

// seedProfileBlock writes an arbitrary profiles.<name> block into the document.
func seedProfileBlock(t *testing.T, name string, block json.RawMessage) {
	t.Helper()
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if err := doc.SetProfileBlock(name, block); err != nil {
		t.Fatalf("SetProfileBlock(%q): %v", name, err)
	}
	if err := doc.Save(); err != nil {
		t.Fatalf("Save document: %v", err)
	}
}

// CloneAs backs `create --from` (CLI) and POST /api/profiles (web). A clone
// must carry the whole profile block, not just `[opencode]`.
func TestCloneAsPreservesSiblingAndUnknownKeys(t *testing.T) {
	setupTestEnv(t)
	seedProfileBlock(t, "dev", json.RawMessage(
		`{"[opencode]":{"telemetry":false,"future_key_xyz":{"a":1}},`+
			`"[senpi]":{"keep":"me"},"[codex]":{"s":1},"team":{"shared":true}}`,
	))

	src, err := Load("dev")
	if err != nil {
		t.Fatalf("Load(dev): %v", err)
	}

	if err := Save(src.CloneAs("copy")); err != nil {
		t.Fatalf("Save(clone): %v", err)
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	raw, ok, err := doc.ProfileBlock("copy")
	if err != nil || !ok {
		t.Fatalf("ProfileBlock(copy): ok=%v err=%v", ok, err)
	}
	var block map[string]json.RawMessage
	if err := json.Unmarshal(raw, &block); err != nil {
		t.Fatalf("unmarshal clone block: %v", err)
	}

	for key, want := range map[string]string{
		"[senpi]": `{"keep":"me"}`,
		"[codex]": `{"s":1}`,
		"team":    `{"shared":true}`,
	} {
		got, present := block[key]
		if !present {
			t.Fatalf("clone dropped sibling key %s", key)
		}
		if !jsonEqual(t, string(got), want) {
			t.Fatalf("clone %s = %s, want %s", key, got, want)
		}
	}

	var openCode map[string]json.RawMessage
	if err := json.Unmarshal(block[config.OpenCodeKey], &openCode); err != nil {
		t.Fatalf("unmarshal clone [opencode]: %v", err)
	}
	if _, present := openCode["future_key_xyz"]; !present {
		t.Fatal("clone dropped unknown key inside [opencode]")
	}

	// The source profile must be untouched by the clone.
	if _, ok, err := doc.ProfileBlock("dev"); err != nil || !ok {
		t.Fatalf("source profile lost: ok=%v err=%v", ok, err)
	}
}

func jsonEqual(t *testing.T, a, b string) bool {
	t.Helper()
	var av, bv any
	if err := json.Unmarshal([]byte(a), &av); err != nil {
		t.Fatalf("unmarshal %q: %v", a, err)
	}
	if err := json.Unmarshal([]byte(b), &bv); err != nil {
		t.Fatalf("unmarshal %q: %v", b, err)
	}
	return reflect.DeepEqual(av, bv)
}

func mustProfileBlock(t *testing.T, openCode json.RawMessage) json.RawMessage {
	t.Helper()
	if len(openCode) == 0 {
		openCode = json.RawMessage(`{}`)
	}
	block, err := json.Marshal(map[string]json.RawMessage{
		config.OpenCodeKey: openCode,
	})
	if err != nil {
		t.Fatalf("marshal profile block: %v", err)
	}
	return block
}

func TestLoad(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "test-profile", `{"disabled_mcps":["test-mcp"]}`)

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

	if p.Path != config.OmoFile() {
		t.Errorf("Path = %q, want omo document %q", p.Path, config.OmoFile())
	}
}

func TestLoadNonexistent(t *testing.T) {
	setupTestEnv(t)

	_, err := Load("nonexistent")
	if err == nil {
		t.Fatal("Expected error when loading nonexistent profile")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *NotFoundError, got %T (%v)", err, err)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected errors.Is(err, fs.ErrNotExist), got %v", err)
	}
	if notFound.Name != "nonexistent" {
		t.Fatalf("NotFoundError.Name = %q, want nonexistent", notFound.Name)
	}
}

func TestSave(t *testing.T) {
	setupTestEnv(t)

	p := &Profile{
		Name: "new-profile",
		Config: config.Config{
			DisabledAgents: []string{"agent1"},
		},
	}

	if err := Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if p.Path != config.OmoFile() {
		t.Errorf("Path = %q, want omo document %q", p.Path, config.OmoFile())
	}

	loaded, err := Load("new-profile")
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}

	if len(loaded.Config.DisabledAgents) != 1 || loaded.Config.DisabledAgents[0] != "agent1" {
		t.Error("Saved config doesn't match original")
	}
}

func TestSavePreservesOtherProfilesAndSiblingKeys(t *testing.T) {
	setupTestEnv(t)

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	doc := config.NewDocument()
	doc.SetRaw(config.SenpiKey, json.RawMessage(`{"mode":"senpi-only","n":1}`))
	doc.SetRaw(config.CodexKey, json.RawMessage(`{"instructions":"keep me"}`))
	doc.SetRaw("task", json.RawMessage(`{"enabled":true,"queue":["a","b"]}`))
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"shared":true,"telemetry":false}`))

	if err := doc.SetProfileBlock("keep-me", json.RawMessage(
		`{"[opencode]":{"telemetry":false,"disabled_mcps":["keep"]},"custom_block":{"v":1}}`,
	)); err != nil {
		t.Fatalf("SetProfileBlock keep-me: %v", err)
	}
	if err := doc.SetProfileBlock("edit-me", json.RawMessage(
		`{"[opencode]":{"telemetry":true},"profile_sibling":{"x":2}}`,
	)); err != nil {
		t.Fatalf("SetProfileBlock edit-me: %v", err)
	}
	if err := doc.Save(); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	p, err := Load("edit-me")
	if err != nil {
		t.Fatalf("Load edit-me: %v", err)
	}
	p.Config.DisabledAgents = []string{"edited"}
	if err := Save(p); err != nil {
		t.Fatalf("Save edit-me: %v", err)
	}

	roundtrip, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}

	for _, key := range []string{config.SenpiKey, config.CodexKey, "task", config.OpenCodeKey} {
		got, ok := roundtrip.Raw(key)
		if !ok {
			t.Fatalf("sibling key %q dropped after Save", key)
		}
		want, _ := doc.Raw(key)
		var gotV, wantV any
		if err := json.Unmarshal(got, &gotV); err != nil {
			t.Fatalf("unmarshal got %q: %v", key, err)
		}
		if err := json.Unmarshal(want, &wantV); err != nil {
			t.Fatalf("unmarshal want %q: %v", key, err)
		}
		if !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("sibling key %q changed\ngot:  %#v\nwant: %#v", key, gotV, wantV)
		}
	}

	if !roundtrip.HasProfile("keep-me") {
		t.Fatal("other profile keep-me was dropped")
	}
	keepBlock, ok, err := roundtrip.ProfileBlock("keep-me")
	if err != nil || !ok {
		t.Fatalf("keep-me ProfileBlock: ok=%v err=%v", ok, err)
	}
	var keepFields map[string]json.RawMessage
	if err := json.Unmarshal(keepBlock, &keepFields); err != nil {
		t.Fatalf("unmarshal keep-me: %v", err)
	}
	if _, ok := keepFields["custom_block"]; !ok {
		t.Fatal("keep-me custom_block sibling was dropped")
	}
	var keepOpen map[string]any
	if err := json.Unmarshal(keepFields[config.OpenCodeKey], &keepOpen); err != nil {
		t.Fatalf("unmarshal keep-me opencode: %v", err)
	}
	mcps, ok := keepOpen["disabled_mcps"].([]any)
	if !ok || len(mcps) != 1 || mcps[0] != "keep" {
		t.Fatalf("keep-me [opencode] altered: %#v", keepOpen)
	}

	edited, err := Load("edit-me")
	if err != nil {
		t.Fatalf("reload edit-me: %v", err)
	}
	if len(edited.Config.DisabledAgents) != 1 || edited.Config.DisabledAgents[0] != "edited" {
		t.Fatalf("edit-me update missing: %#v", edited.Config.DisabledAgents)
	}
	if _, ok := edited.PreservedBlock["profile_sibling"]; !ok {
		t.Fatal("edit-me profile_sibling was not preserved across Save")
	}
}

func TestDelete(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "to-delete", `{}`)

	if err := Delete("to-delete"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if Exists("to-delete") {
		t.Error("Profile should be deleted from the document")
	}
}

// Rename must move the whole block in one write: no lost siblings, and never
// both names present.
func TestRenameMovesBlockAndLeavesOneName(t *testing.T) {
	setupTestEnv(t)
	seedProfileBlock(t, "dev", json.RawMessage(
		`{"[opencode]":{"telemetry":false,"future_key_xyz":1},"[senpi]":{"keep":"me"},"team":{"shared":true}}`,
	))

	if err := Rename("dev", "dev2"); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if Exists("dev") {
		t.Fatal("old name still present after rename")
	}
	if !Exists("dev2") {
		t.Fatal("new name missing after rename")
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	raw, ok, err := doc.ProfileBlock("dev2")
	if err != nil || !ok {
		t.Fatalf("ProfileBlock(dev2): ok=%v err=%v", ok, err)
	}
	var block map[string]json.RawMessage
	if err := json.Unmarshal(raw, &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !jsonEqual(t, string(block["[senpi]"]), `{"keep":"me"}`) {
		t.Fatalf("[senpi] lost in rename: %s", block["[senpi]"])
	}
	if !jsonEqual(t, string(block["team"]), `{"shared":true}`) {
		t.Fatalf("shared key lost in rename: %s", block["team"])
	}
	if !jsonEqual(t, string(block[config.OpenCodeKey]), `{"telemetry":false,"future_key_xyz":1}`) {
		t.Fatalf("[opencode] altered in rename: %s", block[config.OpenCodeKey])
	}
}

func TestRenameRejectsMissingAndTakenNames(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{}`)
	seedProfile(t, "prod", `{}`)

	var notFound *NotFoundError
	if err := Rename("ghost", "x"); !errors.As(err, &notFound) {
		t.Fatalf("expected *NotFoundError, got %T (%v)", err, err)
	}

	var exists *ExistsError
	if err := Rename("dev", "prod"); !errors.As(err, &exists) {
		t.Fatalf("expected *ExistsError, got %T (%v)", err, err)
	}

	// The rejected rename must not have disturbed either profile.
	if !Exists("dev") || !Exists("prod") {
		t.Fatal("a rejected rename mutated the document")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	setupTestEnv(t)

	err := Delete("nonexistent")
	if err == nil {
		t.Fatal("Expected error when deleting nonexistent profile")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *NotFoundError, got %T (%v)", err, err)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected errors.Is(err, fs.ErrNotExist), got %v", err)
	}
	if notFound.Name != "nonexistent" {
		t.Fatalf("NotFoundError.Name = %q, want nonexistent", notFound.Name)
	}
}

func TestList(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "profile1", `{}`)
	seedProfile(t, "profile2", `{}`)

	profiles, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

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
	setupTestEnv(t)

	profiles, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}
}

func TestExists(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "exists-test", `{}`)

	if !Exists("exists-test") {
		t.Error("Expected profile to exist")
	}

	if Exists("nonexistent") {
		t.Error("Expected profile to not exist")
	}
}

func TestLoadWithLegacyFields(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "legacy-profile", `{
		"disabled_mcps": ["test-mcp"],
		"unknownLegacyField": "some value",
		"anotherUnknown": 123
	}`)

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
	setupTestEnv(t)

	seedProfile(t, "valid-profile", `{"disabled_mcps": ["valid-mcp"]}`)

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
	setupTestEnv(t)

	seedProfile(t, "preserve-unknown", `{
		"disabled_mcps": ["test-mcp"],
		"customField": {"enabled": true},
		"anotherLegacy": [1, 2, 3]
	}`)

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
	setupTestEnv(t)

	seedProfile(t, "field-presence", `{
		"disabled_mcps": ["test-mcp"],
		"agents": {
			"worker": {"model": "gpt-5"}
		}
	}`)

	p, err := Load("field-presence")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !p.FieldPresence["disabled_mcps"] {
		t.Fatal("expected disabled_mcps to be marked present")
	}

	if !p.FieldPresence["agents.*.model"] {
		t.Fatal("expected agents.*.model to be marked present")
	}

	if p.FieldPresence["agents.*.temperature"] {
		t.Fatal("expected agents.*.temperature to remain absent from FieldPresence")
	}

	if _, ok := p.FieldPresence["categories.*.model"]; ok {
		t.Fatal("expected categories.*.model to be absent from FieldPresence")
	}
}

func TestProfileLoadCapturesLeafPresenceForNestedKnownFields(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "leaf-presence", `{
		"agents": {
			"build": {"model": "gpt-4"}
		}
	}`)

	p, err := Load("leaf-presence")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !p.FieldPresence["agents.*.model"] {
		t.Fatal("expected agents.*.model to be marked present")
	}

	if p.FieldPresence["agents.*.temperature"] {
		t.Fatal("expected agents.*.temperature to remain absent")
	}
}

func TestProfileSaveRoundTripsPreservedUnknownFragments(t *testing.T) {
	setupTestEnv(t)

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
	setupTestEnv(t)

	// Valid document envelope, but profiles.<name> is not a JSON object — Load
	// must fail when decoding the profile block.
	seedProfileBlock(t, "malformed", json.RawMessage(`"[not-an-object]"`))

	if _, err := Load("malformed"); err == nil {
		t.Fatal("expected malformed profile block load to fail")
	}
}

func TestRegressionSparsePersistenceContract(t *testing.T) {
	setupTestEnv(t)

	const profileName = "regression-sparse-contract"
	initialProfileJSON := `{
		"disabled_mcps": ["legacy-mcp"],
		"hashline_edit": true,
		"custom_bundle": {
			"enabled": true,
			"thresholds": {
				"high": 2,
				"low": 1
			}
		},
		"custom_flags": ["alpha", "beta"]
	}`

	seedProfile(t, profileName, initialProfileJSON)

	p, err := Load(profileName)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !p.FieldPresence["disabled_mcps"] {
		t.Fatal("expected disabled_mcps to be marked present on initial load")
	}
	if !p.FieldPresence["hashline_edit"] {
		t.Fatal("expected hashline_edit to be marked present on initial load")
	}
	if len(p.PreservedUnknown) != 2 {
		t.Fatalf("expected 2 top-level preserved unknown fragments, got %d", len(p.PreservedUnknown))
	}

	p.Config.DisabledMCPs = []string{"omit-me"}
	p.Config.HashlineEdit = boolPtr(false)
	p.Config.DisabledHooks = []string{}
	p.Config.DefaultRunAgent = ""
	p.Config.Experimental = &config.ExperimentalConfig{
		TaskSystem: boolPtr(false),
		MaxTools:   int64Ptr(0),
	}
	p.Config.Agents = map[string]*config.AgentConfig{
		"builder": {Model: "gpt-5"},
	}
	p.PreservedUnknown["agents"] = json.RawMessage(`{"builder":{"model":"legacy-model","legacy":true},"legacy_agent":{"model":"legacy-only"}}`)
	p.PreservedUnknown["experimental"] = json.RawMessage(`{"legacy_flag":true,"task_system":true}`)

	selection := NewBlankSelection()
	for _, path := range []string{
		"hashline_edit",
		"disabled_hooks",
		"default_run_agent",
		"experimental.task_system",
		"experimental.max_tools",
		"agents.*.model",
	} {
		selection.SetSelected(path, true)
	}

	data, err := MarshalSparse(&p.Config, selection, p.PreservedUnknown)
	if err != nil {
		t.Fatalf("MarshalSparse failed: %v", err)
	}

	decoded := decodeSparseJSON(t, data)
	if _, ok := decoded["disabled_mcps"]; ok {
		t.Fatalf("expected unchecked disabled_mcps to be omitted, got %#v", decoded["disabled_mcps"])
	}

	serializedChecks := []struct {
		name  string
		check func(*testing.T, map[string]any)
	}{
		{
			name: "selected zero values survive sparse JSON",
			check: func(t *testing.T, payload map[string]any) {
				t.Helper()

				if value, ok := payload["hashline_edit"].(bool); !ok || value {
					t.Fatalf("expected hashline_edit to be false, got %#v", payload["hashline_edit"])
				}

				if value, ok := payload["default_run_agent"].(string); !ok || value != "" {
					t.Fatalf("expected default_run_agent to be an explicit empty string, got %#v", payload["default_run_agent"])
				}

				hooks, ok := payload["disabled_hooks"].([]any)
				if !ok || len(hooks) != 0 {
					t.Fatalf("expected disabled_hooks to be an explicit empty array, got %#v", payload["disabled_hooks"])
				}

				experimental := decodedObject(t, payload["experimental"], "experimental")
				if taskSystem, ok := experimental["task_system"].(bool); !ok || taskSystem {
					t.Fatalf("expected experimental.task_system to be false, got %#v", experimental["task_system"])
				}
				if maxTools, ok := experimental["max_tools"].(float64); !ok || maxTools != 0 {
					t.Fatalf("expected experimental.max_tools to be 0, got %#v", experimental["max_tools"])
				}
			},
		},
		{
			name: "multiple preserved unknown fragments survive and known leaves win overlaps",
			check: func(t *testing.T, payload map[string]any) {
				t.Helper()

				customBundle := decodedObject(t, payload["custom_bundle"], "custom_bundle")
				if enabled, ok := customBundle["enabled"].(bool); !ok || !enabled {
					t.Fatalf("expected preserved custom_bundle.enabled to remain true, got %#v", customBundle["enabled"])
				}

				thresholds := decodedObject(t, customBundle["thresholds"], "custom_bundle.thresholds")
				if low, ok := thresholds["low"].(float64); !ok || low != 1 {
					t.Fatalf("expected preserved custom_bundle.thresholds.low to remain 1, got %#v", thresholds["low"])
				}
				if high, ok := thresholds["high"].(float64); !ok || high != 2 {
					t.Fatalf("expected preserved custom_bundle.thresholds.high to remain 2, got %#v", thresholds["high"])
				}

				customFlags, ok := payload["custom_flags"].([]any)
				if !ok || !reflect.DeepEqual(customFlags, []any{"alpha", "beta"}) {
					t.Fatalf("expected preserved custom_flags to remain [alpha beta], got %#v", payload["custom_flags"])
				}

				agents := decodedObject(t, payload["agents"], "agents")
				builder := decodedObject(t, agents["builder"], "agents.builder")
				if model, ok := builder["model"].(string); !ok || model != "gpt-5" {
					t.Fatalf("expected selected agents.builder.model to win, got %#v", builder["model"])
				}
				if legacy, ok := builder["legacy"].(bool); !ok || !legacy {
					t.Fatalf("expected preserved agents.builder.legacy sibling to remain, got %#v", builder["legacy"])
				}

				legacyAgent := decodedObject(t, agents["legacy_agent"], "agents.legacy_agent")
				if model, ok := legacyAgent["model"].(string); !ok || model != "legacy-only" {
					t.Fatalf("expected preserved legacy_agent model to remain, got %#v", legacyAgent["model"])
				}

				experimental := decodedObject(t, payload["experimental"], "experimental")
				if legacyFlag, ok := experimental["legacy_flag"].(bool); !ok || !legacyFlag {
					t.Fatalf("expected preserved experimental.legacy_flag sibling to remain, got %#v", experimental["legacy_flag"])
				}
			},
		},
	}

	for _, tc := range serializedChecks {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t, decoded)
		})
	}

	seedProfile(t, profileName, string(data))

	reloaded, err := Load(profileName)
	if err != nil {
		t.Fatalf("Reload after sparse save failed: %v", err)
	}

	if reloaded.FieldPresence["disabled_mcps"] {
		t.Fatal("expected unchecked disabled_mcps to stay omitted after reload")
	}
	for _, requiredKey := range []string{"hashline_edit", "disabled_hooks", "default_run_agent", "agents.*.model", "experimental.task_system", "experimental.max_tools"} {
		if !reloaded.FieldPresence[requiredKey] {
			t.Fatalf("expected %s to be marked present after sparse reload", requiredKey)
		}
	}

	if reloaded.Config.HashlineEdit == nil || *reloaded.Config.HashlineEdit {
		t.Fatalf("expected hashline_edit to reload as false, got %#v", reloaded.Config.HashlineEdit)
	}
	if !reflect.DeepEqual(reloaded.Config.DisabledHooks, []string{}) {
		t.Fatalf("expected disabled_hooks to reload as an explicit empty slice, got %#v", reloaded.Config.DisabledHooks)
	}
	if reloaded.Config.Experimental == nil {
		t.Fatal("expected experimental config to reload")
	}
	if reloaded.Config.Experimental.TaskSystem == nil || *reloaded.Config.Experimental.TaskSystem {
		t.Fatalf("expected experimental.task_system to reload as false, got %#v", reloaded.Config.Experimental.TaskSystem)
	}
	if reloaded.Config.Experimental.MaxTools == nil || *reloaded.Config.Experimental.MaxTools != 0 {
		t.Fatalf("expected experimental.max_tools to reload as 0, got %#v", reloaded.Config.Experimental.MaxTools)
	}
	if got := reloaded.Config.DefaultRunAgent; got != "" {
		t.Fatalf("expected default_run_agent to reload as empty string, got %q", got)
	}
	if got := reloaded.Config.Agents["builder"].Model; got != "gpt-5" {
		t.Fatalf("expected selected builder model to reload, got %q", got)
	}
	if got := reloaded.Config.Agents["legacy_agent"].Model; got != "legacy-only" {
		t.Fatalf("expected preserved legacy_agent model to reload, got %q", got)
	}

	if len(reloaded.PreservedUnknown) != 2 {
		t.Fatalf("expected only top-level unknown fragments to survive reload, got %d", len(reloaded.PreservedUnknown))
	}

	reloadedUnknownChecks := []struct {
		name  string
		check func(*testing.T)
	}{
		{
			name: "custom bundle survives round-trip",
			check: func(t *testing.T) {
				t.Helper()

				var customBundle map[string]any
				if err := json.Unmarshal(reloaded.PreservedUnknown["custom_bundle"], &customBundle); err != nil {
					t.Fatalf("failed to decode reloaded custom_bundle: %v", err)
				}
				if enabled, ok := customBundle["enabled"].(bool); !ok || !enabled {
					t.Fatalf("expected reloaded custom_bundle.enabled to remain true, got %#v", customBundle["enabled"])
				}
			},
		},
		{
			name: "custom flags survive round-trip",
			check: func(t *testing.T) {
				t.Helper()

				var customFlags []string
				if err := json.Unmarshal(reloaded.PreservedUnknown["custom_flags"], &customFlags); err != nil {
					t.Fatalf("failed to decode reloaded custom_flags: %v", err)
				}
				if !reflect.DeepEqual(customFlags, []string{"alpha", "beta"}) {
					t.Fatalf("expected reloaded custom_flags to remain [alpha beta], got %#v", customFlags)
				}
			},
		},
	}

	for _, tc := range reloadedUnknownChecks {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t)
		})
	}
}
