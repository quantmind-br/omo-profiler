package profile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/diogenes/omo-profiler/internal/config"
)

type ActiveConfig struct {
	Exists      bool
	Config      config.Config
	ProfileName string
	IsOrphan    bool
}

type activeState struct {
	Name string `json:"name"`
}

func activeStateFile() string {
	return filepath.Join(config.ConfigDir(), ".active-profile")
}

func loadActiveState() (*activeState, error) {
	data, err := os.ReadFile(activeStateFile())
	if err != nil {
		return nil, err
	}
	var state activeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveActiveState(name string) error {
	state := activeState{Name: name}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(activeStateFile(), data, 0644)
}

func GetActive() (*ActiveConfig, error) {
	configPath := config.ConfigFile()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ActiveConfig{Exists: false}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Fast path: try cached active profile first (O(1) instead of O(N))
	if state, err := loadActiveState(); err == nil && state.Name != "" {
		if profile, err := Load(state.Name); err == nil {
			if profile.MatchesConfig(&cfg) {
				return &ActiveConfig{
					Exists:      true,
					Config:      cfg,
					ProfileName: state.Name,
					IsOrphan:    false,
				}, nil
			}
		}
	}

	// Fallback: scan all profiles (state file missing/stale)
	profiles, err := List()
	if err != nil {
		return nil, err
	}

	for _, name := range profiles {
		profile, err := Load(name)
		if err != nil {
			continue
		}
		if profile.MatchesConfig(&cfg) {
			_ = saveActiveState(name)
			return &ActiveConfig{
				Exists:      true,
				Config:      cfg,
				ProfileName: name,
				IsOrphan:    false,
			}, nil
		}
	}

	return &ActiveConfig{
		Exists:      true,
		Config:      cfg,
		ProfileName: "(custom)",
		IsOrphan:    true,
	}, nil
}

func SetActive(name string) error {
	profile, err := Load(name)
	if err != nil {
		return err
	}

	if err := config.EnsureDirs(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile.Config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(config.ConfigFile(), data, 0644); err != nil {
		return err
	}

	// Save active state for O(1) lookup
	// Ignore errors - state file is optimization only
	_ = saveActiveState(name)

	return nil
}

func (p *Profile) MatchesConfig(cfg *config.Config) bool {
	pBytes, err1 := normalizeForComparison(&p.Config)
	cBytes, err2 := normalizeForComparison(cfg)

	if err1 != nil || err2 != nil {
		return false
	}

	return bytes.Equal(pBytes, cBytes)
}

func normalizeForComparison(cfg *config.Config) ([]byte, error) {
	cfgCopy := *cfg
	cfgCopy.Schema = ""

	return json.Marshal(cfgCopy)
}
