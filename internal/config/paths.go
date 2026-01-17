package config

import (
	"os"
	"path/filepath"
)

var baseDir string // empty = use os.UserHomeDir()

// SetBaseDir sets a custom base directory (for testing)
func SetBaseDir(path string) { baseDir = path }

// ResetBaseDir resets to using real home directory
func ResetBaseDir() { baseDir = "" }

// ConfigDir returns ~/.config/opencode/
func ConfigDir() string {
	if baseDir != "" {
		return filepath.Join(baseDir, ".config", "opencode")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
}

// ProfilesDir returns ~/.config/opencode/profiles/
func ProfilesDir() string {
	return filepath.Join(ConfigDir(), "profiles")
}

// ConfigFile returns ~/.config/opencode/oh-my-opencode.json
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "oh-my-opencode.json")
}

// EnsureDirs creates config and profiles directories if they don't exist
func EnsureDirs() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(ProfilesDir(), 0755)
}

// DefaultSchema is the schema URL to add when creating new profiles
const DefaultSchema = "https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/master/assets/oh-my-opencode.schema.json"
