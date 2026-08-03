package config

import (
	"os"
	"path/filepath"
)

const (
	// OmoBasename is the canonical config file name.
	OmoBasename = "omo.json"
	// OmoBasenameJSONC is the JSONC variant, preferred by upstream when present.
	OmoBasenameJSONC = "omo.jsonc"
	// OmoDirname is the config directory name, on every platform.
	OmoDirname = ".omo"

	// LegacyOpenagentBasename is the pre-unification config file name.
	// Upstream reads it only through its migration engine.
	LegacyOpenagentBasename = "oh-my-openagent.json"
	// LegacyOpencodeBasename is the oldest config file name, migration-only.
	LegacyOpencodeBasename = "oh-my-opencode.json"
)

var baseDir string // empty = use os.UserHomeDir()

// SetBaseDir sets a custom base directory (for testing)
func SetBaseDir(path string) { baseDir = path }

// ResetBaseDir resets to using real home directory
func ResetBaseDir() { baseDir = "" }

// HomeDir returns the effective home directory.
func HomeDir() string {
	if baseDir != "" {
		return baseDir
	}
	home, _ := os.UserHomeDir()
	return home
}

// OmoDir returns ~/.omo/ — the user config layer directory.
func OmoDir() string {
	return filepath.Join(HomeDir(), OmoDirname)
}

// OmoFile returns the user-layer config file path.
// Upstream prefers omo.jsonc over omo.json; we must edit whichever the user
// actually has, otherwise our writes would be ignored. New installs get omo.json.
func OmoFile() string {
	dir := OmoDir()
	jsonc := filepath.Join(dir, OmoBasenameJSONC)
	if _, err := os.Stat(jsonc); err == nil {
		return jsonc
	}
	return filepath.Join(dir, OmoBasename)
}

// ModelsFile returns ~/.omo/models.json — omo-profiler's local model registry.
// This is omo-profiler state, not part of the upstream omo.json contract.
func ModelsFile() string {
	return filepath.Join(OmoDir(), "models.json")
}


// LegacyConfigDir returns ~/.config/opencode/ — the pre-unification location,
// kept for detecting configs that still need migrating.
func LegacyConfigDir() string {
	return filepath.Join(HomeDir(), ".config", "opencode")
}

// LegacyConfigFile returns the legacy flat config path if one exists, else "".
func LegacyConfigFile() string {
	dir := LegacyConfigDir()
	for _, name := range []string{LegacyOpenagentBasename, LegacyOpencodeBasename} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// LegacyProfilesDir returns ~/.config/opencode/profiles/ — the pre-unification
// file-per-profile directory, kept for migration only.
func LegacyProfilesDir() string {
	return filepath.Join(LegacyConfigDir(), "profiles")
}

// EnsureDirs creates the omo config directory if it doesn't exist.
func EnsureDirs() error {
	return os.MkdirAll(OmoDir(), 0755)
}

// DefaultSchema is the schema URL written into new omo.json documents.
const DefaultSchema = "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/omo.schema.json"
