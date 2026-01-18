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
