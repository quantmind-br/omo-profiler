package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/diogenes/omo-profiler/internal/config"
)

type RegisteredModel struct {
	DisplayName string `json:"displayName"`
	ModelID     string `json:"modelId"`
	Provider    string `json:"provider"`
}

// ModelNotFoundError is returned when no model matches (Provider, ModelID).
type ModelNotFoundError struct {
	Provider string
	ModelID  string
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("model with provider '%s' and ID '%s' not found", e.Provider, e.ModelID)
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

	return config.WriteFileAtomic(config.ModelsFile(), data, 0644)
}

// add inserts m in memory. Errors if the (Provider, ModelID) pair already
// exists. Persisting is the caller's job — use the package-level Add.
func (r *ModelsRegistry) add(m RegisteredModel) error {
	for _, existing := range r.Models {
		if existing.ModelID == m.ModelID && existing.Provider == m.Provider {
			return &ModelExistsError{Provider: m.Provider, ModelID: m.ModelID}
		}
	}
	r.Models = append(r.Models, m)
	return nil
}

// update replaces an existing model in memory, identified by (provider,
// modelId). Supports renaming ModelID. Use the package-level Update to persist.
func (r *ModelsRegistry) update(provider, modelId string, m RegisteredModel) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId && existing.Provider == provider {
			idx = i
			break
		}
	}

	if idx == -1 {
		return &ModelNotFoundError{Provider: provider, ModelID: modelId}
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
	return nil
}

// remove drops a model in memory, identified by (provider, modelId). Use the
// package-level Delete to persist.
func (r *ModelsRegistry) remove(provider, modelId string) error {
	idx := -1
	for i, existing := range r.Models {
		if existing.ModelID == modelId && existing.Provider == provider {
			idx = i
			break
		}
	}

	if idx == -1 {
		return &ModelNotFoundError{Provider: provider, ModelID: modelId}
	}

	r.Models = append(r.Models[:idx], r.Models[idx+1:]...)
	return nil
}

// regMutex serializes read-modify-write cycles on models.json.
//
// The registry is one file rewritten in full, so two concurrent mutations that
// each load, edit and save would silently drop one of the changes. The web
// server handles requests on separate goroutines, which is where this bites.
//
// Scope: this guards one process, like the config document's lock.
var regMutex sync.Mutex

// Mutate runs fn against the registry as a serialized transaction: load, apply
// fn, save. Returning an error aborts before any write.
//
// fn must not call Mutate (the lock is not reentrant) and must not call Save.
func Mutate(fn func(*ModelsRegistry) error) error {
	regMutex.Lock()
	defer regMutex.Unlock()

	reg, err := Load()
	if err != nil {
		return err
	}
	if err := fn(reg); err != nil {
		return err
	}
	return reg.Save()
}

// Add registers a new model, failing with *ModelExistsError when the
// (Provider, ModelID) pair is taken. The check and the write share one
// transaction.
func Add(m RegisteredModel) error {
	return Mutate(func(r *ModelsRegistry) error { return r.add(m) })
}

// AddMany registers every model that is not already present, in one
// transaction, and reports how many were added and how many were skipped as
// duplicates. A bulk import is a single file write, not one per model.
func AddMany(list []RegisteredModel) (added, skipped int, err error) {
	err = Mutate(func(r *ModelsRegistry) error {
		added, skipped = 0, 0
		for _, m := range list {
			switch addErr := r.add(m); {
			case addErr == nil:
				added++
			case errors.As(addErr, new(*ModelExistsError)):
				skipped++
			default:
				return addErr
			}
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return added, skipped, nil
}

// Update replaces the model identified by (provider, modelId), optionally
// renaming it. Lookup and write share one transaction.
func Update(provider, modelId string, m RegisteredModel) error {
	return Mutate(func(r *ModelsRegistry) error { return r.update(provider, modelId, m) })
}

// Delete removes the model identified by (provider, modelId). Lookup and write
// share one transaction.
func Delete(provider, modelId string) error {
	return Mutate(func(r *ModelsRegistry) error { return r.remove(provider, modelId) })
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
