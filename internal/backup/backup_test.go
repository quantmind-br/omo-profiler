package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/diogenes/omo-profiler/internal/config"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	t.Cleanup(config.ResetBaseDir)

	// Create omo config directory
	if err := os.MkdirAll(config.OmoDir(), 0755); err != nil {
		t.Fatalf("failed to create omo dir: %v", err)
	}
	return tmpDir
}

func TestCreate(t *testing.T) {
	setupTestDir(t)

	// Create a test config file
	configPath := config.OmoFile()
	testContent := []byte(`{"test": "data"}`)
	if err := os.WriteFile(configPath, testContent, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Create backup
	backupPath, err := Create(configPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("backup file does not exist: %s", backupPath)
	}

	// Verify backup content matches original
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if string(backupContent) != string(testContent) {
		t.Errorf("backup content = %q, want %q", backupContent, testContent)
	}

	filename := filepath.Base(backupPath)
	expectedPrefix := config.OmoBasename + ".bak."
	if len(filename) < len(expectedPrefix+"2006-01-02-150405") {
		t.Errorf("backup filename too short: %s", filename)
	}
	if filename[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("backup filename prefix = %s, want prefix %s", filename, expectedPrefix)
	}
}

func TestCreate_SourceNotExists(t *testing.T) {
	setupTestDir(t)

	_, err := Create("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Create() should fail for nonexistent source")
	}
}

func TestList_Empty(t *testing.T) {
	setupTestDir(t)

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("List() returned %d backups, want 0", len(backups))
	}
}

func TestList_SortsByDateDescending(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	// Create backups with different timestamps
	timestamps := []string{
		"2025-01-15-100000",
		"2025-01-16-120000",
		"2025-01-14-080000",
	}
	for _, ts := range timestamps {
		name := config.OmoBasename + ".bak." + ts
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test backup: %v", err)
		}
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 3 {
		t.Fatalf("List() returned %d backups, want 3", len(backups))
	}

	expectedOrder := []string{
		config.OmoBasename + ".bak.2025-01-16-120000",
		config.OmoBasename + ".bak.2025-01-15-100000",
		config.OmoBasename + ".bak.2025-01-14-080000",
	}
	for i, backup := range backups {
		if backup.Name != expectedOrder[i] {
			t.Errorf("backups[%d].Name = %s, want %s", i, backup.Name, expectedOrder[i])
		}
	}
}

func TestList_IgnoresOtherFiles(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	validName := config.OmoBasename + ".bak.2025-01-16-120000"
	if err := os.WriteFile(filepath.Join(dir, validName), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create test backup: %v", err)
	}

	ignoredFiles := []string{
		config.OmoBasename,              // main config
		"other-file.bak.2025-01-16",     // different prefix
		config.OmoBasename + ".bak.bad", // bad timestamp
	}
	for _, name := range ignoredFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("List() returned %d backups, want 1", len(backups))
	}
	if len(backups) > 0 && backups[0].Name != validName {
		t.Errorf("backups[0].Name = %s, want %s", backups[0].Name, validName)
	}
}

func TestList_RecognisesLegacyPrefixes(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	// Migration-safety: pre-migration backups under legacy basenames must still list.
	files := []string{
		config.OmoBasename + ".bak.2025-01-16-120000",
		config.OmoBasenameJSONC + ".bak.2025-01-16-110000",
		config.LegacyOpenagentBasename + ".bak.2025-01-15-100000",
		config.LegacyOpencodeBasename + ".bak.2025-01-14-090000",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create test backup: %v", err)
		}
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 4 {
		t.Fatalf("List() returned %d backups, want 4 (omo, omo.jsonc, and both legacy prefixes)", len(backups))
	}

	got := map[string]bool{}
	for _, b := range backups {
		got[b.Name] = true
	}
	for _, name := range files {
		if !got[name] {
			t.Errorf("List() missing backup %s", name)
		}
	}
}

func TestRestore(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	backupContent := []byte(`{"restored": true}`)
	backupPath := filepath.Join(dir, config.OmoBasename+".bak.2025-01-16-120000")
	if err := os.WriteFile(backupPath, backupContent, 0644); err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Create original config with different content
	configPath := config.OmoFile()
	if err := os.WriteFile(configPath, []byte(`{"original": true}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Restore
	if err := Restore(backupPath); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// Verify config was restored
	restoredContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(restoredContent) != string(backupContent) {
		t.Errorf("restored content = %q, want %q", restoredContent, backupContent)
	}
}

func TestRestore_BackupNotExists(t *testing.T) {
	setupTestDir(t)

	err := Restore("/nonexistent/backup.json")
	if err == nil {
		t.Error("Restore() should fail for nonexistent backup")
	}
}

func TestClean(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	// Create 7 backups
	timestamps := []string{
		"2025-01-10-100000",
		"2025-01-11-100000",
		"2025-01-12-100000",
		"2025-01-13-100000",
		"2025-01-14-100000",
		"2025-01-15-100000",
		"2025-01-16-100000",
	}
	for _, ts := range timestamps {
		name := config.OmoBasename + ".bak." + ts
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create backup: %v", err)
		}
	}

	// Clean keeping only 5
	if err := Clean(5); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 5 {
		t.Errorf("after Clean(5), got %d backups, want 5", len(backups))
	}

	// Verify oldest backups were removed
	for _, backup := range backups {
		ts := backup.Timestamp
		if ts.Before(time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC)) {
			t.Errorf("old backup should have been removed: %s", backup.Name)
		}
	}
}

func TestClean_FewerThanKeep(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	// Create only 2 backups
	timestamps := []string{
		"2025-01-15-100000",
		"2025-01-16-100000",
	}
	for _, ts := range timestamps {
		name := config.OmoBasename + ".bak." + ts
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create backup: %v", err)
		}
	}

	// Clean keeping 5 (more than we have)
	if err := Clean(5); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("after Clean(5) with 2 backups, got %d, want 2", len(backups))
	}
}

// Backups written before the nanosecond format must stay listable, restorable
// and correctly ordered against new ones.
func TestList_MixesLegacyAndNanosecondNames(t *testing.T) {
	setupTestDir(t)
	dir := config.OmoDir()

	names := []string{
		config.OmoBasename + ".bak.2025-01-16-120000",           // legacy
		config.OmoBasename + ".bak.2025-01-16-120001.000000001", // new, 1s later
		config.OmoBasename + ".bak.2025-01-16-120001.000000002", // new, 1ns later
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o600); err != nil {
			t.Fatalf("failed to create test backup: %v", err)
		}
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(backups) != len(names) {
		t.Fatalf("got %d backups, want %d — a name format was dropped", len(backups), len(names))
	}

	// Most recent first, so the nanosecond ordering must be honoured.
	want := []string{names[2], names[1], names[0]}
	for i, b := range backups {
		if b.Name != filepath.Base(want[i]) {
			t.Fatalf("position %d: got %s, want %s", i, b.Name, filepath.Base(want[i]))
		}
	}
}

// A backup name must never be reused, even when the clock hands out the same
// reading twice — a coarse clock source or a second omo-profiler process can
// do exactly that, and a truncating write would destroy the pre-image the
// other writer depends on. Freezing the clock makes the collision certain;
// wall-clock concurrency cannot test this reliably.
func TestCreate_FrozenClockStillYieldsDistinctBackups(t *testing.T) {
	setupTestDir(t)
	configPath := filepath.Join(config.OmoDir(), config.OmoBasename)

	frozen := time.Date(2025, 1, 16, 12, 0, 1, 0, time.UTC)
	restore := now
	now = func() time.Time { return frozen }
	defer func() { now = restore }()

	const writes = 8
	paths := make(map[string]bool, writes)
	for i := 0; i < writes; i++ {
		content := fmt.Sprintf(`{"gen":%d}`, i)
		if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
			t.Fatalf("seed config: %v", err)
		}

		path, err := Create(configPath)
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		if paths[path] {
			t.Fatalf("Create #%d reused the name %s — a pre-image was overwritten", i, filepath.Base(path))
		}
		paths[path] = true

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read back #%d: %v", i, err)
		}
		if string(got) != content {
			t.Fatalf("backup #%d holds %q, want %q", i, got, content)
		}
	}

	// Every snapshot must still be on disk and readable through List, in the
	// order the writes happened.
	backups, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(backups) != writes {
		t.Fatalf("got %d backups, want %d — names collided", len(backups), writes)
	}
	for i, b := range backups {
		want := fmt.Sprintf(`{"gen":%d}`, writes-1-i) // List is most-recent-first
		got, err := os.ReadFile(b.Path)
		if err != nil {
			t.Fatalf("read %s: %v", b.Name, err)
		}
		if string(got) != want {
			t.Fatalf("List position %d holds %q, want %q — ordering broke", i, got, want)
		}
	}
}

// The omo document can hold API tokens, so a backup must never be more
// permissive than the file it copies. The mode is set explicitly rather than
// left to the umask, which would narrow an 0644 source under a strict umask.
func TestCreate_PreservesSourceMode(t *testing.T) {
	for _, perm := range []os.FileMode{0o600, 0o644} {
		t.Run(fmt.Sprintf("%o", perm), func(t *testing.T) {
			setupTestDir(t)
			configPath := filepath.Join(config.OmoDir(), config.OmoBasename)
			if err := os.WriteFile(configPath, []byte(`{}`), perm); err != nil {
				t.Fatalf("seed config: %v", err)
			}
			if err := os.Chmod(configPath, perm); err != nil { // defeat the umask on the source too
				t.Fatalf("chmod source: %v", err)
			}

			path, err := Create(configPath)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat backup: %v", err)
			}
			if got := info.Mode().Perm(); got != perm {
				t.Fatalf("backup mode %o, want %o — a backup must not differ from its source", got, perm)
			}
		})
	}
}
