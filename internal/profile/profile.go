package profile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/diogenes/omo-profiler/internal/config"
)

// Profile represents a saved configuration profile
type Profile struct {
	Name                string
	Config              config.Config
	Path                string
	HasLegacyFields     bool   `json:"-"`
	LegacyFieldsWarning string `json:"-"`
}

// detectLegacyFields checks if the JSON data contains unknown fields
// that are not part of the config.Config struct. Returns true and an
// error message if unknown fields are detected.
func detectLegacyFields(data []byte) (bool, string) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cfg config.Config
	if err := dec.Decode(&cfg); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "unknown field") {
			return true, errStr
		}
		return false, ""
	}
	return false, ""
}

// Load loads a profile by name from the profiles directory
func Load(name string) (*Profile, error) {
	path := filepath.Join(config.ProfilesDir(), name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	hasLegacy, warning := detectLegacyFields(data)

	return &Profile{
		Name:                name,
		Config:              cfg,
		Path:                path,
		HasLegacyFields:     hasLegacy,
		LegacyFieldsWarning: warning,
	}, nil
}

func Save(p *Profile) error {
	return p.Save()
}

func (p *Profile) Save() error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	path := filepath.Join(config.ProfilesDir(), p.Name+".json")
	data, err := json.MarshalIndent(p.Config, "", "  ")
	if err != nil {
		return err
	}

	p.Path = path
	return os.WriteFile(path, data, 0644)
}

// Delete removes the profile file
func Delete(name string) error {
	path := filepath.Join(config.ProfilesDir(), name+".json")
	return os.Remove(path)
}

// List returns names of all profiles (without .json extension)
func List() ([]string, error) {
	dir := config.ProfilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	return names, nil
}

// Exists checks if a profile exists
func Exists(name string) bool {
	path := filepath.Join(config.ProfilesDir(), name+".json")
	_, err := os.Stat(path)
	return err == nil
}
