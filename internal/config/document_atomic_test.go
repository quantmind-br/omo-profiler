package config

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Save must leave exactly one file behind: no temp litter, always parseable.
func TestSaveLeavesNoTempFiles(t *testing.T) {
	SetBaseDir(t.TempDir())
	t.Cleanup(ResetBaseDir)

	for range 25 {
		doc, err := LoadDocument()
		if err != nil {
			t.Fatalf("LoadDocument: %v", err)
		}
		if err := doc.SetProfileBlock("p", json.RawMessage(`{"[opencode]":{}}`)); err != nil {
			t.Fatalf("SetProfileBlock: %v", err)
		}
		if err := doc.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	entries, err := os.ReadDir(OmoDir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}

	data, err := os.ReadFile(OmoFile())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("document is not valid JSON after repeated saves: %v", err)
	}

	info, err := os.Stat(OmoFile())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("new document mode = %v, want 0600 (it can hold secrets)", perm)
	}
}

// A Save that cannot complete must leave the previous document untouched. A
// plain truncating write would have already destroyed it by this point.
func TestSaveFailureLeavesPreviousDocumentIntact(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	SetBaseDir(t.TempDir())
	t.Cleanup(ResetBaseDir)

	doc, err := LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if err := doc.SetProfileBlock("original", json.RawMessage(`{"[opencode]":{"telemetry":true}}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	before, err := os.ReadFile(OmoFile())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Block temp-file creation without making the target itself unwritable.
	if err := os.Chmod(OmoDir(), 0500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(OmoDir(), 0700) })

	doc2, err := LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if err := doc2.SetProfileBlock("replacement", json.RawMessage(`{"[opencode]":{}}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}
	if err := doc2.Save(); err == nil {
		t.Fatal("expected Save to fail when it cannot stage a temp file")
	}

	after, err := os.ReadFile(OmoFile())
	if err != nil {
		t.Fatalf("ReadFile after failed save: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("failed Save modified the document:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// A user who tightened (or loosened) the document's permissions must keep them:
// atomic replacement recreates the file, so the mode has to be carried over.
func TestSavePreservesExistingFileMode(t *testing.T) {
	for _, mode := range []os.FileMode{0600, 0640} {
		t.Run(mode.String(), func(t *testing.T) {
			SetBaseDir(t.TempDir())
			t.Cleanup(ResetBaseDir)

			doc, err := LoadDocument()
			if err != nil {
				t.Fatalf("LoadDocument: %v", err)
			}
			if err := doc.Save(); err != nil {
				t.Fatalf("Save: %v", err)
			}
			if err := os.Chmod(OmoFile(), mode); err != nil {
				t.Fatalf("Chmod: %v", err)
			}

			doc2, err := LoadDocument()
			if err != nil {
				t.Fatalf("LoadDocument: %v", err)
			}
			if err := doc2.SetProfileBlock("p", json.RawMessage(`{"[opencode]":{}}`)); err != nil {
				t.Fatalf("SetProfileBlock: %v", err)
			}
			if err := doc2.Save(); err != nil {
				t.Fatalf("Save: %v", err)
			}

			info, err := os.Stat(OmoFile())
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			if got := info.Mode().Perm(); got != mode {
				t.Fatalf("mode = %v, want %v (Save must not change permissions)", got, mode)
			}
		})
	}
}
