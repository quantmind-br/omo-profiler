package models

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/diogenes/omo-profiler/internal/config"
)

type RegisteredModel struct {
	DisplayName string `json:"displayName"`
	ModelID     string `json:"modelId"` // unique key
	Provider    string `json:"provider"`
}

type ProviderGroup struct {
	Provider string            // Provider name, "" for no provider
	Models   []RegisteredModel // Sorted by DisplayName ascending
}

type ModelsRegistry struct {
	Models []RegisteredModel `json:"models"`
}

// Load loads the models registry from the models.json file.
func Load() (*ModelsRegistry, error) {
	path := config.ModelsFile()

	// File does not exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ModelsRegistry{Models: []RegisteredModel{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// File exists, empty content
	if len(data) == 0 {
		return &ModelsRegistry{Models: []RegisteredModel{}}, nil
	}

	var registry ModelsRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		// File exists, corrupted JSON
		// Backup to models.json.bak
		bakPath := path + ".bak"
		if wErr := os.WriteFile(bakPath, data, 0644); wErr != nil {
			fmt.Fprintf(os.Stderr, "failed to backup corrupted models file: %v\n", wErr)
		}

		fmt.Fprintf(os.Stderr, "models.json is corrupted, backed up to %s. Loading empty registry. Error: %v\n", bakPath, err)
		return &ModelsRegistry{Models: []RegisteredModel{}}, nil
	}

	// Ensure slice is not nil
	if registry.Models == nil {
		registry.Models = []RegisteredModel{}
	}

	return &registry, nil
}

// Save persists the registry to disk.
func (r *ModelsRegistry) Save() error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(config.ModelsFile(), data, 0644)
}

// Add adds a new model to the registry. Errors if duplicate ModelID.
func (r *ModelsRegistry) Add(m RegisteredModel) error {
	for _, existing := range r.Models {
		if existing.ModelID == m.ModelID {
			return fmt.Errorf("model with ID '%s' already exists", m.ModelID)
		}
	}
	r.Models = append(r.Models, m)
	return r.Save()
}

// Update updates an existing model. Support renaming ModelID.
func (r *ModelsRegistry) Update(modelId string, m RegisteredModel) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("model '%s' not found", modelId)
	}

	// Check for conflict if renaming
	if m.ModelID != modelId {
		for _, existing := range r.Models {
			if existing.ModelID == m.ModelID {
				return fmt.Errorf("model with ID '%s' already exists", m.ModelID)
			}
		}
	}

	r.Models[idx] = m
	return r.Save()
}

// Delete removes a model by ID.
func (r *ModelsRegistry) Delete(modelId string) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("model '%s' not found", modelId)
	}

	r.Models = append(r.Models[:idx], r.Models[idx+1:]...)
	return r.Save()
}

// Get retrieves a model by ID.
func (r *ModelsRegistry) Get(modelId string) *RegisteredModel {
	for i := range r.Models {
		if r.Models[i].ModelID == modelId {
			return &r.Models[i]
		}
	}
	return nil
}

// List returns all registered models.
func (r *ModelsRegistry) List() []RegisteredModel {
	// Return a copy to avoid external modification of internal slice
	result := make([]RegisteredModel, len(r.Models))
	copy(result, r.Models)
	return result
}

// ListByProvider returns models grouped by provider.
func (r *ModelsRegistry) ListByProvider() []ProviderGroup {
	groups := make(map[string][]RegisteredModel)
	for _, m := range r.Models {
		groups[m.Provider] = append(groups[m.Provider], m)
	}

	var result []ProviderGroup
	for provider, models := range groups {
		// Sort models by DisplayName
		sort.Slice(models, func(i, j int) bool {
			return models[i].DisplayName < models[j].DisplayName
		})
		result = append(result, ProviderGroup{
			Provider: provider,
			Models:   models,
		})
	}

	// Sort groups by Provider (case-insensitive), empty last
	sort.Slice(result, func(i, j int) bool {
		p1 := result[i].Provider
		p2 := result[j].Provider

		if p1 == "" && p2 != "" {
			return false // empty last
		}
		if p1 != "" && p2 == "" {
			return true // empty last
		}

		return strings.ToLower(p1) < strings.ToLower(p2)
	})

	return result
}

// Exists checks if a model ID exists in the registry.
func Exists(modelId string) bool {
	r, err := Load()
	if err != nil {
		return false
	}
	return r.Get(modelId) != nil
}
