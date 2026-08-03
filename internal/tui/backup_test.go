package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
)

// Every mutating entry point snapshots the document first. The TUI import and
// delete paths ran without one while the CLI and web paths backed up, so a bad
// TUI write had no recovery path.
func TestTUIMutationsBackUpFirst(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(t *testing.T, a App) tea.Msg
	}{
		{"delete", func(t *testing.T, a App) tea.Msg {
			return a.doDeleteProfile("dev")()
		}},
		{"import", func(t *testing.T, a App) tea.Msg {
			src := filepath.Join(t.TempDir(), "imported.json")
			if err := os.WriteFile(src, []byte(`{"telemetry":true}`), 0o644); err != nil {
				t.Fatalf("write import source: %v", err)
			}
			return a.doImportProfile(src)()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seedTUIDoc(t)

			if n := countBackups(t); n != 0 {
				t.Fatalf("expected no backups before the write, got %d", n)
			}

			switch msg := tc.run(t, NewApp()).(type) {
			case deleteProfileDoneMsg:
				if msg.err != nil {
					t.Fatalf("delete failed: %v", msg.err)
				}
			case importProfileDoneMsg:
				if msg.err != nil {
					t.Fatalf("import failed: %v", msg.err)
				}
			default:
				t.Fatalf("unexpected message %T", msg)
			}

			if n := countBackups(t); n != 1 {
				t.Fatalf("expected exactly 1 backup after the write, got %d", n)
			}
		})
	}
}

func seedTUIDoc(t *testing.T) {
	t.Helper()
	config.SetBaseDir(t.TempDir())
	t.Cleanup(config.ResetBaseDir)

	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	if err := doc.SetProfileBlock("dev", json.RawMessage(`{"[opencode]":{"telemetry":false}}`)); err != nil {
		t.Fatalf("SetProfileBlock: %v", err)
	}
	doc.EnsureSchema()
	if err := doc.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !profile.Exists("dev") {
		t.Fatal("seed failed")
	}
}

func countBackups(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir(config.OmoDir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	n := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak") {
			n++
		}
	}
	return n
}
