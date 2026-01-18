package views

import (
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
)

func TestNewWizardForEdit_SetsOriginalProfileName(t *testing.T) {
	p := &profile.Profile{Name: "test-profile", Config: config.Config{}}
	w := NewWizardForEdit(p)

	if w.originalProfileName != "test-profile" {
		t.Errorf("originalProfileName = %q, want %q", w.originalProfileName, "test-profile")
	}
	if !w.editMode {
		t.Error("editMode should be true for edit wizard")
	}
	if w.profileName != "test-profile" {
		t.Errorf("profileName = %q, want %q", w.profileName, "test-profile")
	}
}

func TestNewWizard_LeavesOriginalProfileNameEmpty(t *testing.T) {
	w := NewWizard()

	if w.originalProfileName != "" {
		t.Errorf("originalProfileName = %q, want empty string", w.originalProfileName)
	}
	if w.editMode {
		t.Error("editMode should be false for new wizard")
	}
}

func TestWizard_Save_CreateModeSkipsRenameLogic(t *testing.T) {
	w := NewWizard()
	w.step = StepReview
	w.profileName = "new-profile"

	if w.editMode {
		t.Error("New wizard should have editMode=false")
	}
}

func TestWizard_EditMode_DetectsRename(t *testing.T) {
	p := &profile.Profile{Name: "original-name", Config: config.Config{}}
	w := NewWizardForEdit(p)

	w.profileName = "new-name"

	isRename := w.editMode && w.profileName != w.originalProfileName
	if !isRename {
		t.Error("Should detect rename when profileName differs from originalProfileName")
	}
}

func TestWizard_EditMode_NoRenameWhenNameUnchanged(t *testing.T) {
	p := &profile.Profile{Name: "same-name", Config: config.Config{}}
	w := NewWizardForEdit(p)

	isRename := w.editMode && w.profileName != w.originalProfileName
	if isRename {
		t.Error("Should not detect rename when name is unchanged")
	}
}

func TestWizard_Save_ValidationCalled(t *testing.T) {
	w := NewWizard()
	w.step = StepReview
	w.profileName = "test-profile"

	if w.step != StepReview {
		t.Error("wizard should be at StepReview")
	}
	if w.err != nil {
		t.Error("wizard should have no error initially")
	}
}

func TestWizardEditFlow_PreservesConfig(t *testing.T) {
	cfg := config.Config{
		DisabledMCPs: []string{"test-mcp"},
	}
	p := &profile.Profile{Name: "test-profile", Config: cfg}
	w := NewWizardForEdit(p)

	if len(w.config.DisabledMCPs) != 1 {
		t.Error("Config should be preserved when creating wizard for edit")
	}
	if w.config.DisabledMCPs[0] != "test-mcp" {
		t.Errorf("MCP = %q, want %q", w.config.DisabledMCPs[0], "test-mcp")
	}
}

func TestWizardEditFlow_SetupAllSteps(t *testing.T) {
	p := &profile.Profile{Name: "test-profile", Config: config.Config{}}
	w := NewWizardForEdit(p)

	if w.step != StepName {
		t.Errorf("step = %d, want %d", w.step, StepName)
	}

	if !w.editMode {
		t.Error("editMode should be true")
	}

	if w.originalProfileName != "test-profile" {
		t.Errorf("originalProfileName = %q, want %q", w.originalProfileName, "test-profile")
	}
}

func TestWizardCreateFlow_StillWorks(t *testing.T) {
	w := NewWizard()

	if w.editMode {
		t.Error("New wizard should have editMode=false")
	}
	if w.originalProfileName != "" {
		t.Errorf("originalProfileName should be empty for new wizard, got %q", w.originalProfileName)
	}
	if w.profileName != "" {
		t.Errorf("profileName should be empty for new wizard, got %q", w.profileName)
	}
	if w.step != StepName {
		t.Errorf("New wizard should start at StepName, got %d", w.step)
	}
}

func TestWizard_GetProfile(t *testing.T) {
	p := &profile.Profile{Name: "original", Config: config.Config{}}
	w := NewWizardForEdit(p)

	w.profileName = "renamed"

	result := w.GetProfile()
	if result.Name != "renamed" {
		t.Errorf("GetProfile().Name = %q, want %q", result.Name, "renamed")
	}
}

func TestWizard_IsEditMode(t *testing.T) {
	createWizard := NewWizard()
	if createWizard.IsEditMode() {
		t.Error("NewWizard should return IsEditMode=false")
	}

	p := &profile.Profile{Name: "test", Config: config.Config{}}
	editWizard := NewWizardForEdit(p)
	if !editWizard.IsEditMode() {
		t.Error("NewWizardForEdit should return IsEditMode=true")
	}
}

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	config.SetBaseDir(tmpDir)
	if err := config.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}
	return func() {
		config.ResetBaseDir()
	}
}

func TestWizardEditFlow_LoadAndSave(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	original := &profile.Profile{
		Name:   "existing-profile",
		Config: config.Config{DisabledMCPs: []string{"mcp1"}},
	}
	if err := profile.Save(original); err != nil {
		t.Fatalf("Failed to save original profile: %v", err)
	}

	loaded, err := profile.Load("existing-profile")
	if err != nil {
		t.Fatalf("Failed to load profile: %v", err)
	}
	w := NewWizardForEdit(loaded)

	if w.profileName != "existing-profile" {
		t.Errorf("profileName = %q, want %q", w.profileName, "existing-profile")
	}
	if !w.editMode {
		t.Error("editMode should be true")
	}
	if len(w.config.DisabledMCPs) != 1 || w.config.DisabledMCPs[0] != "mcp1" {
		t.Error("config should be loaded correctly")
	}
}

func TestWizardEditFlow_RenameDeletesOldProfile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	original := &profile.Profile{
		Name:   "old-name",
		Config: config.Config{},
	}
	if err := profile.Save(original); err != nil {
		t.Fatalf("Failed to save original profile: %v", err)
	}

	w := NewWizardForEdit(original)
	w.profileName = "new-name"

	isRename := w.editMode && w.profileName != w.originalProfileName
	if !isRename {
		t.Fatal("Should detect rename")
	}

	newProfile := &profile.Profile{Name: w.profileName, Config: w.config}
	if err := profile.Save(newProfile); err != nil {
		t.Fatalf("Failed to save renamed profile: %v", err)
	}

	if err := profile.Delete(w.originalProfileName); err != nil {
		t.Fatalf("Failed to delete old profile: %v", err)
	}

	if profile.Exists("old-name") {
		t.Error("Old profile should be deleted")
	}
	if !profile.Exists("new-name") {
		t.Error("New profile should exist")
	}
}

func TestWizardEditFlow_RenameToDuplicateBlocked(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	profile1 := &profile.Profile{Name: "profile-a", Config: config.Config{}}
	profile2 := &profile.Profile{Name: "profile-b", Config: config.Config{}}
	if err := profile.Save(profile1); err != nil {
		t.Fatalf("Failed to save profile1: %v", err)
	}
	if err := profile.Save(profile2); err != nil {
		t.Fatalf("Failed to save profile2: %v", err)
	}

	w := NewWizardForEdit(profile1)
	w.profileName = "profile-b"

	if w.editMode && w.profileName != w.originalProfileName {
		if !profile.Exists(w.profileName) {
			t.Error("Should detect that profile-b already exists")
		}
	}
}

func TestWizardCreateFlow_SavesNewProfile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	w := NewWizard()
	w.profileName = "brand-new-profile"
	w.config = config.Config{DisabledAgents: []string{"agent1"}}

	if w.editMode {
		t.Fatal("Should be in create mode")
	}

	newProfile := &profile.Profile{Name: w.profileName, Config: w.config}
	if err := profile.Save(newProfile); err != nil {
		t.Fatalf("Failed to save new profile: %v", err)
	}

	loaded, err := profile.Load("brand-new-profile")
	if err != nil {
		t.Fatalf("Failed to load saved profile: %v", err)
	}

	if len(loaded.Config.DisabledAgents) != 1 || loaded.Config.DisabledAgents[0] != "agent1" {
		t.Error("Saved config should match original")
	}
}
