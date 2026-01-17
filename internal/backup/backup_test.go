package backup

import (
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

	// Create config directory
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	return tmpDir
}

func TestCreate(t *testing.T) {
	setupTestDir(t)

	// Create a test config file
	configPath := config.ConfigFile()
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

	// Verify filename format
	filename := filepath.Base(backupPath)
	if len(filename) < len("oh-my-opencode.json.bak.2006-01-02-150405") {
		t.Errorf("backup filename too short: %s", filename)
	}
	if filename[:24] != "oh-my-opencode.json.bak." {
		t.Errorf("backup filename prefix wrong: %s", filename)
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
	dir := config.ConfigDir()

	// Create backups with different timestamps
	timestamps := []string{
		"2025-01-15-100000",
		"2025-01-16-120000",
		"2025-01-14-080000",
	}
	for _, ts := range timestamps {
		name := "oh-my-opencode.json.bak." + ts
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

	// Should be sorted descending (most recent first)
	expectedOrder := []string{
		"oh-my-opencode.json.bak.2025-01-16-120000",
		"oh-my-opencode.json.bak.2025-01-15-100000",
		"oh-my-opencode.json.bak.2025-01-14-080000",
	}
	for i, backup := range backups {
		if backup.Name != expectedOrder[i] {
			t.Errorf("backups[%d].Name = %s, want %s", i, backup.Name, expectedOrder[i])
		}
	}
}

func TestList_IgnoresOtherFiles(t *testing.T) {
	setupTestDir(t)
	dir := config.ConfigDir()

	// Create a valid backup
	validName := "oh-my-opencode.json.bak.2025-01-16-120000"
	if err := os.WriteFile(filepath.Join(dir, validName), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create test backup: %v", err)
	}

	// Create files that should be ignored
	ignoredFiles := []string{
		"oh-my-opencode.json",         // main config
		"other-file.bak.2025-01-16",   // different prefix
		"oh-my-opencode.json.bak.bad", // bad timestamp
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

func TestRestore(t *testing.T) {
	setupTestDir(t)
	dir := config.ConfigDir()

	// Create a backup file
	backupContent := []byte(`{"restored": true}`)
	backupPath := filepath.Join(dir, "oh-my-opencode.json.bak.2025-01-16-120000")
	if err := os.WriteFile(backupPath, backupContent, 0644); err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Create original config with different content
	configPath := config.ConfigFile()
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
	dir := config.ConfigDir()

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
		name := "oh-my-opencode.json.bak." + ts
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
	dir := config.ConfigDir()

	// Create only 2 backups
	timestamps := []string{
		"2025-01-15-100000",
		"2025-01-16-100000",
	}
	for _, ts := range timestamps {
		name := "oh-my-opencode.json.bak." + ts
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
