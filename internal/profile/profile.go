package profile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/diogenes/omo-profiler/internal/config"
)

type Profile struct {
	Name                string
	Config              config.Config
	Path                string
	PreservedUnknown    map[string]json.RawMessage `json:"-"`
	FieldPresence       map[string]bool            `json:"-"`
	HasLegacyFields     bool                       `json:"-"`
	LegacyFieldsWarning string                     `json:"-"`
}

var knownConfigTags = []string{
	"$schema",
	"disabled_mcps",
	"disabled_agents",
	"disabled_skills",
	"disabled_hooks",
	"disabled_commands",
	"hashline_edit",
	"model_fallback",
	"agents",
	"categories",
	"claude_code",
	"sisyphus_agent",
	"comment_checker",
	"experimental",
	"auto_update",
	"skills",
	"ralph_loop",
	"runtime_fallback",
	"background_task",
	"notification",
	"git_master",
	"new_task_system_enabled",
	"disabled_tools",
	"babysitting",
	"browser_automation_engine",
	"tmux",
	"websearch",
	"sisyphus",
	"default_run_agent",
	"start_work",
	"openclaw",
	"model_capabilities",
	"_migrations",
}

func knownTags() map[string]struct{} {
	tags := make(map[string]struct{}, len(knownConfigTags))
	for _, tag := range knownConfigTags {
		tags[tag] = struct{}{}
	}
	return tags
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

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	preservedUnknown := make(map[string]json.RawMessage)
	fieldPresence := make(map[string]bool)
	tags := knownTags()
	for key, value := range raw {
		if _, ok := tags[key]; ok {
			fieldPresence[key] = true
			continue
		}
		preservedUnknown[key] = value
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
		PreservedUnknown:    preservedUnknown,
		FieldPresence:       fieldPresence,
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
	data, err := json.Marshal(p.Config)
	if err != nil {
		return err
	}

	merged := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &merged); err != nil {
		return err
	}

	for key, value := range p.PreservedUnknown {
		if _, exists := merged[key]; exists {
			continue
		}
		merged[key] = value
	}

	data, err = marshalSortedJSONObject(merged)
	if err != nil {
		return err
	}

	p.Path = path
	return os.WriteFile(path, data, 0644)
}

func marshalSortedJSONObject(values map[string]json.RawMessage) ([]byte, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, key := range keys {
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}

		buf.WriteString("  ")
		buf.Write(encodedKey)
		buf.WriteString(": ")
		buf.Write(values[key])
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
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
		if trimmedName, ok := strings.CutSuffix(name, ".json"); ok {
			names = append(names, trimmedName)
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
