package profile

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"testing"
	"github.com/diogenes/omo-profiler/internal/config"
)

func TestGetActive_NoConfig(t *testing.T) {
	setupTestEnv(t)

	active, err := GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if active.Exists {
		t.Errorf("expected Exists=false when no omo document")
	}
	if active.ProfileName != "" {
		t.Errorf("expected empty ProfileName, got %q", active.ProfileName)
	}
}

func TestApply_NotFound(t *testing.T) {
	setupTestEnv(t)

	_, err := Apply("nonexistent")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected errors.Is(err, fs.ErrNotExist), got %v", err)
	}
}

func TestApply_SubstitutesDeclaredKeysOnly(t *testing.T) {
	setupTestEnv(t)

	// Root has both [opencode] and [senpi].
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true,"git_master":{"commit_footer":"base"}}`))
	doc.SetRaw(config.SenpiKey, json.RawMessage(`{"keep":"me"}`))
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Profile declares only [opencode].
	seedProfile(t, "dev", `{"telemetry":false}`)

	applied, err := Apply("dev")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.Snapshot != "base" {
		t.Fatalf("Snapshot = %q, want base", applied.Snapshot)
	}

	// Root [opencode] should equal the profile block, [senpi] untouched.
	doc, err = config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument after Apply: %v", err)
	}
	rootOC, _ := doc.Raw(config.OpenCodeKey)
	canonRoot, err := canonicalJSON(rootOC)
	if err != nil {
		t.Fatalf("canonicalJSON root: %v", err)
	}
	profileBlock, _, err := doc.ProfileBlock("dev")
	if err != nil {
		t.Fatalf("ProfileBlock: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(profileBlock, &fields); err != nil {
		t.Fatalf("Unmarshal profile block: %v", err)
	}
	canonProfile, err := canonicalJSON(fields[config.OpenCodeKey])
	if err != nil {
		t.Fatalf("canonicalJSON profile: %v", err)
	}
	if string(canonRoot) != string(canonProfile) {
		t.Fatalf("root [opencode] = %s, want profile block %s", canonRoot, canonProfile)
	}

	rootSenpi, _ := doc.Raw(config.SenpiKey)
	canonSenpi, err := canonicalJSON(rootSenpi)
	if err != nil {
		t.Fatalf("canonicalJSON senpi: %v", err)
	}
	if string(canonSenpi) != `{"keep":"me"}` {
		t.Fatalf("root [senpi] = %s, want {\"keep\":\"me\"}", canonSenpi)
	}
}

func TestApply_SnapshotsUnmatchedRoot(t *testing.T) {
	setupTestEnv(t)

	// Root [opencode] matches no profile.
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	seedProfile(t, "dev", `{"telemetry":false}`)

	applied, err := Apply("dev")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.Snapshot != "base" {
		t.Fatalf("Snapshot = %q, want base", applied.Snapshot)
	}

	// profiles.base.[opencode] should hold the old root value.
	doc, err = config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument after Apply: %v", err)
	}
	baseBlock, ok, err := doc.ProfileBlock("base")
	if err != nil {
		t.Fatalf("ProfileBlock: %v", err)
	}
	if !ok {
		t.Fatal("profiles.base not found")
	}
	var baseFields map[string]json.RawMessage
	if err := json.Unmarshal(baseBlock, &baseFields); err != nil {
		t.Fatalf("Unmarshal base block: %v", err)
	}
	canonBase, err := canonicalJSON(baseFields[config.OpenCodeKey])
	if err != nil {
		t.Fatalf("canonicalJSON base: %v", err)
	}
	if string(canonBase) != `{"telemetry":true}` {
		t.Fatalf("profiles.base.[opencode] = %s, want {\"telemetry\":true}", canonBase)
	}
}

func TestApply_NoSnapshotWhenRootMatches(t *testing.T) {
	setupTestEnv(t)

	// Seed an unmatched root so the first apply snapshots it.
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	seedProfile(t, "dev", `{"telemetry":false}`)
	// First apply: root matches no profile → snapshot.
	applied, err := Apply("dev")
	if err != nil {
		t.Fatalf("Apply first: %v", err)
	}
	if applied.Snapshot != "base" {
		t.Fatalf("first Snapshot = %q, want base", applied.Snapshot)
	}

	// Second apply: root now matches "dev" → no snapshot.
	applied, err = Apply("dev")
	if err != nil {
		t.Fatalf("Apply second: %v", err)
	}
	if applied.Snapshot != "" {
		t.Fatalf("second Snapshot = %q, want empty", applied.Snapshot)
	}

	// No extra profile should have been created.
	doc, err = config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	names, err := doc.ProfileNames()
	if err != nil {
		t.Fatalf("ProfileNames: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 profiles (dev, base), got %d: %v", len(names), names)
	}
}

func TestApply_SnapshotNameCollision(t *testing.T) {
	setupTestEnv(t)

	// Seed an unmatched root so the apply snapshots it.
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	seedProfile(t, "dev", `{"telemetry":false}`)
	seedProfile(t, "base", `{"telemetry":false,"git_master":{"commit_footer":"base"}}`)
	applied, err := Apply("dev")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if applied.Snapshot != "base-1" {
		t.Fatalf("Snapshot = %q, want base-1", applied.Snapshot)
	}
}

func TestActiveName_DetectsAndReportsNone(t *testing.T) {
	setupTestEnv(t)

	seedProfile(t, "dev", `{"telemetry":false}`)

	if _, err := Apply("dev"); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}

	name, err := ActiveName(doc)
	if err != nil {
		t.Fatalf("ActiveName: %v", err)
	}
	if name != "dev" {
		t.Fatalf("ActiveName = %q, want dev", name)
	}

	// Mutate the root.
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	doc, err = config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument after mutate: %v", err)
	}

	name, err = ActiveName(doc)
	if err != nil {
		t.Fatalf("ActiveName after mutate: %v", err)
	}
	if name != "" {
		t.Fatalf("ActiveName = %q, want empty", name)
	}

	active, err := GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if !active.Modified {
		t.Fatal("expected Modified=true after mutating root")
	}
}

func TestApply_AppliesAllDeclaredKeys(t *testing.T) {
	setupTestEnv(t)

	seedProfileBlock(t, "dev", json.RawMessage(
		`{"[opencode]":{"telemetry":false},"[senpi]":{"keep":"me"},"agents":{"build":{"model":"test"}}}`,
	))

	if _, err := Apply("dev"); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}

	for _, key := range []string{config.OpenCodeKey, config.SenpiKey, "agents"} {
		raw, ok := doc.Raw(key)
		if !ok {
			t.Errorf("root key %q missing after Apply", key)
			continue
		}
		if len(raw) == 0 {
			t.Errorf("root key %q is empty after Apply", key)
		}
	}
}

// Apply must surface a parse error for a profile block that is valid JSON but
// not a JSON object, rather than silently treating it as "no match".
func TestApply_MalformedProfileBlock(t *testing.T) {
	setupTestEnv(t)

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	// Profile block is a JSON string, not a JSON object.
	if err := doc.SetProfileBlock("dev", json.RawMessage(`"not an object"`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}
	doc.EnsureSchema()
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Capture the document before the failed apply.
	before, err := os.ReadFile(config.OmoFile())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	_, err = Apply("dev")
	if err == nil {
		t.Fatal("expected error for malformed profile block")
	}

	// The failed apply must not mutate the document.
	after, err := os.ReadFile(config.OmoFile())
	if err != nil {
		t.Fatalf("ReadFile after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("Apply mutated the document despite the parse error")
	}
}

// ActiveName must propagate canonicalization errors for corrupt root JSON
// rather than silently treating the profile as "no match".
func TestActiveName_MalformedRootJSON(t *testing.T) {
	setupTestEnv(t)

	doc := config.NewDocument()
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{telemetry: false}`)) // malformed JSON
	if err := doc.SetProfileBlock("dev", json.RawMessage(`{"[opencode]":{"telemetry":false}}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}

	_, err := ActiveName(doc)
	if err == nil {
		t.Fatal("expected error for malformed root JSON")
	}
}

// ActiveName must propagate unmarshal errors for a profile block that is not a
// JSON object, rather than silently skipping it.
func TestActiveName_MalformedProfileBlock(t *testing.T) {
	setupTestEnv(t)

	doc := config.NewDocument()
	if err := doc.SetProfileBlock("dev", json.RawMessage(`"not an object"`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}

	_, err := ActiveName(doc)
	if err == nil {
		t.Fatal("expected error for malformed profile block")
	}
}
