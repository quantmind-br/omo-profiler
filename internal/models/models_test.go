package models

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	return func() {
		config.ResetBaseDir()
	}
}

func TestLoad(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// 1. Non-existent file
	reg, err := Load()
	if err != nil {
		t.Fatalf("Load non-existent failed: %v", err)
	}
	if len(reg.Models) != 0 {
		t.Errorf("Expected empty registry, got %d models", len(reg.Models))
	}

	// 2. Valid file
	model := RegisteredModel{
		DisplayName: "Test Model",
		ModelID:     "test-model",
		Provider:    "openai",
	}
	reg.Models = append(reg.Models, model)
	if err := reg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	reg2, err := Load()
	if err != nil {
		t.Fatalf("Load valid failed: %v", err)
	}
	if len(reg2.Models) != 1 || reg2.Models[0].ModelID != "test-model" {
		t.Errorf("Loaded registry mismatch")
	}

	// 3. Corrupted file
	path := config.ModelsFile()
	if err := os.WriteFile(path, []byte("{invalid-json"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	reg3, err := Load()
	if err != nil {
		t.Fatalf("Load corrupted failed (should return empty, not error): %v", err)
	}
	if len(reg3.Models) != 0 {
		t.Errorf("Expected empty registry for corrupted file, got %d models", len(reg3.Models))
	}

	// Verify backup
	if _, err := os.Stat(path + ".bak"); os.IsNotExist(err) {
		t.Error("Backup file .bak not created")
	}
}

func TestSave(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg := &ModelsRegistry{
		Models: []RegisteredModel{
			{DisplayName: "A", ModelID: "a", Provider: "p"},
		},
	}

	if err := reg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := config.ModelsFile()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var loaded ModelsRegistry
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to parse saved file: %v", err)
	}

	if len(loaded.Models) != 1 || loaded.Models[0].ModelID != "a" {
		t.Error("Saved data mismatch")
	}
}

func TestAdd(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	m1 := RegisteredModel{DisplayName: "M1", ModelID: "m1", Provider: "p1"}

	// Success
	if err := reg.Add(m1); err != nil {
		t.Errorf("Add failed: %v", err)
	}

	if len(reg.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(reg.Models))
	}

	// Duplicate error
	if err := reg.Add(m1); err == nil {
		t.Error("Expected error for duplicate add, got nil")
	} else if err.Error() != "model with ID 'm1' already exists" {
		t.Errorf("Unexpected error msg: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	m1 := RegisteredModel{DisplayName: "M1", ModelID: "m1", Provider: "p1"}
	reg.Add(m1)

	// Success update
	updated := m1
	updated.DisplayName = "M1 Updated"
	if err := reg.Update("m1", updated); err != nil {
		t.Errorf("Update failed: %v", err)
	}
	if reg.Get("m1").DisplayName != "M1 Updated" {
		t.Error("Update didn't persist")
	}

	// Not found
	if err := reg.Update("missing", m1); err == nil {
		t.Error("Expected error for missing update")
	}

	// Rename success
	renamed := m1
	renamed.ModelID = "m1-renamed"
	if err := reg.Update("m1", renamed); err != nil {
		t.Errorf("Rename failed: %v", err)
	}
	if reg.Get("m1") != nil {
		t.Error("Old ID should be gone")
	}
	if reg.Get("m1-renamed") == nil {
		t.Error("New ID should exist")
	}

	// Conflict
	reg.Add(RegisteredModel{ModelID: "m2"})
	conflict := RegisteredModel{ModelID: "m2"}
	if err := reg.Update("m1-renamed", conflict); err == nil {
		t.Error("Expected error for conflict update")
	}
}

func TestDelete(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	reg.Add(RegisteredModel{ModelID: "m1"})

	// Success
	if err := reg.Delete("m1"); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if len(reg.Models) != 0 {
		t.Error("Model not deleted")
	}

	// Not found
	if err := reg.Delete("m1"); err == nil {
		t.Error("Expected error for deleting missing model")
	}
}

func TestGet(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	reg.Add(RegisteredModel{ModelID: "m1"})

	if reg.Get("m1") == nil {
		t.Error("Get should return model")
	}
	if reg.Get("missing") != nil {
		t.Error("Get missing should return nil")
	}
}

func TestListByProvider(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	reg.Add(RegisteredModel{DisplayName: "B", ModelID: "1", Provider: "openai"})
	reg.Add(RegisteredModel{DisplayName: "A", ModelID: "2", Provider: "openai"})
	reg.Add(RegisteredModel{DisplayName: "C", ModelID: "3", Provider: "anthropic"})
	reg.Add(RegisteredModel{DisplayName: "D", ModelID: "4", Provider: ""}) // Empty provider

	groups := reg.ListByProvider()

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}

	// Check order: Anthropic, OpenAI, "" (empty last)
	if groups[0].Provider != "anthropic" {
		t.Errorf("Expected anthropic first, got %s", groups[0].Provider)
	}
	if groups[1].Provider != "openai" {
		t.Errorf("Expected openai second, got %s", groups[1].Provider)
	}
	if groups[2].Provider != "" {
		t.Errorf("Expected empty provider last, got %s", groups[2].Provider)
	}

	// Check sort within group (OpenAI A before B)
	openaiGroup := groups[1]
	if openaiGroup.Models[0].DisplayName != "A" {
		t.Error("Models within group not sorted")
	}
}

func TestExists(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	reg, _ := Load()
	reg.Add(RegisteredModel{ModelID: "m1"})

	if !Exists("m1") {
		t.Error("Exists returned false for existing model")
	}
	if Exists("missing") {
		t.Error("Exists returned true for missing model")
	}
}
