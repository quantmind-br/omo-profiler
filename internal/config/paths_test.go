package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetBaseDir(t *testing.T) {
	// Save and restore original
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	expected := filepath.Join(tmpDir, ".config", "opencode")
	if got := ConfigDir(); got != expected {
		t.Errorf("ConfigDir() = %s, want %s", got, expected)
	}
}

func TestResetBaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)
	ResetBaseDir()

	// After reset, should use real home
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "opencode")
	if got := ConfigDir(); got != expected {
		t.Errorf("ConfigDir() = %s, want %s", got, expected)
	}
}

func TestConfigDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := ConfigDir()
	expected := filepath.Join(tmpDir, ".config", "opencode")
	if got != expected {
		t.Errorf("ConfigDir() = %s, want %s", got, expected)
	}
}

func TestProfilesDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := ProfilesDir()
	expected := filepath.Join(tmpDir, ".config", "opencode", "profiles")
	if got != expected {
		t.Errorf("ProfilesDir() = %s, want %s", got, expected)
	}
}

func TestConfigFile(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := ConfigFile()
	expected := filepath.Join(tmpDir, ".config", "opencode", "oh-my-opencode.json")
	if got != expected {
		t.Errorf("ConfigFile() = %s, want %s", got, expected)
	}
}

func TestEnsureDirs(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	// Dirs should not exist initially
	configDir := ConfigDir()
	profilesDir := ProfilesDir()

	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("configDir should not exist before EnsureDirs")
	}

	// Call EnsureDirs
	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() failed: %v", err)
	}

	// Both should exist now
	info, err := os.Stat(configDir)
	if err != nil {
		t.Errorf("configDir should exist after EnsureDirs: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("configDir should be a directory")
	}

	info, err = os.Stat(profilesDir)
	if err != nil {
		t.Errorf("profilesDir should exist after EnsureDirs: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("profilesDir should be a directory")
	}

	// Calling EnsureDirs again should be idempotent
	if err := EnsureDirs(); err != nil {
		t.Errorf("EnsureDirs() should be idempotent: %v", err)
	}
}

func TestDefaultSchema(t *testing.T) {
	expected := "https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/master/assets/oh-my-opencode.schema.json"
	if DefaultSchema != expected {
		t.Errorf("DefaultSchema = %s, want %s", DefaultSchema, expected)
	}
}
