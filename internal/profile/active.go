package profile

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/diogenes/omo-profiler/internal/config"
)

type ActiveConfig struct {
	Exists      bool
	Config      config.Config
	ProfileName string
	IsOrphan    bool
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

	return os.WriteFile(config.ConfigFile(), data, 0644)
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
