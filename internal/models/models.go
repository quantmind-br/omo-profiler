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
	ModelID     string `json:"modelId"`
	Provider    string `json:"provider"`
}

// ModelExistsError is returned when a model with the same (Provider, ModelID) already exists.
type ModelExistsError struct {
	Provider string
	ModelID  string
}

func (e *ModelExistsError) Error() string {
	if e.Provider == "" {
		return fmt.Sprintf("model with ID '%s' (no provider) already exists", e.ModelID)
	}
	return fmt.Sprintf("model with provider '%s' and ID '%s' already exists", e.Provider, e.ModelID)
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

// Add adds a new model to the registry. Errors if the (Provider, ModelID) pair already exists.
func (r *ModelsRegistry) Add(m RegisteredModel) error {
	for _, existing := range r.Models {
		if existing.ModelID == m.ModelID && existing.Provider == m.Provider {
			return &ModelExistsError{Provider: m.Provider, ModelID: m.ModelID}
		}
	}
	r.Models = append(r.Models, m)
	return r.Save()
}

// Update updates an existing model identified by (provider, modelId). Supports renaming ModelID.
func (r *ModelsRegistry) Update(provider, modelId string, m RegisteredModel) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId && existing.Provider == provider {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("model with provider '%s' and ID '%s' not found", provider, modelId)
	}

	// Check for conflict if renaming
	if m.ModelID != modelId || m.Provider != provider {
		for _, existing := range r.Models {
			if existing.ModelID == m.ModelID && existing.Provider == m.Provider {
				return &ModelExistsError{Provider: m.Provider, ModelID: m.ModelID}
			}
		}
	}

	r.Models[idx] = m
	return r.Save()
}

// Delete removes a model identified by (provider, modelId).
func (r *ModelsRegistry) Delete(provider, modelId string) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId && existing.Provider == provider {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("model with provider '%s' and ID '%s' not found", provider, modelId)
	}

	r.Models = append(r.Models[:idx], r.Models[idx+1:]...)
	return r.Save()
}

// Get retrieves a model by (provider, modelId).
func (r *ModelsRegistry) Get(provider, modelId string) *RegisteredModel {
	for i := range r.Models {
		if r.Models[i].ModelID == modelId && r.Models[i].Provider == provider {
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

// Exists checks if a model with the given (provider, modelId) exists in the registry.
func Exists(provider, modelId string) bool {
	r, err := Load()
	if err != nil {
		return false
	}
	return r.Get(provider, modelId) != nil
}
