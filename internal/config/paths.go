package config

import (
	"os"
	"path/filepath"
)

const (
	// ConfigBasename is the canonical config file name (oh-my-openagent)
	ConfigBasename = "oh-my-openagent.json"
	// LegacyConfigBasename is the legacy config file name (oh-my-opencode)
	LegacyConfigBasename = "oh-my-opencode.json"
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

// ConfigFile returns the active config file path.
// Checks for oh-my-openagent.json (canonical) first, then falls back to
// oh-my-opencode.json (legacy). Defaults to oh-my-openagent.json for new installs.
func ConfigFile() string {
	dir := ConfigDir()
	canonical := filepath.Join(dir, ConfigBasename)
	legacy := filepath.Join(dir, LegacyConfigBasename)

	if _, err := os.Stat(canonical); err == nil {
		return canonical
	}
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return canonical
}

// ModelsFile returns ~/.config/opencode/models.json
func ModelsFile() string {
	return filepath.Join(ConfigDir(), "models.json")
}

// EnsureDirs creates config and profiles directories if they don't exist
func EnsureDirs() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(ProfilesDir(), 0755)
}

// DefaultSchema is the schema URL to add when creating new profiles
const DefaultSchema = "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json"
