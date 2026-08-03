package views

import (
	"errors"
	"testing"

	"github.com/diogenes/omo-profiler/internal/profile"
)

// The wizard validates the profile name when the user types it, but the save
// lands much later, in a tea.Cmd goroutine. A web request handled in that gap
// must not be silently overwritten, so the destination is re-checked inside
// the write transaction. (That transaction is process-local, so this closes
// the in-process window; a second omo-profiler process only narrows it.)
func TestWizardRenameRejectsNameClaimedDuringEdit(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeProfile(t, "dev", `{"telemetry": false}`)

	loaded, err := profile.Load("dev")
	if err != nil {
		t.Fatalf("load dev: %v", err)
	}

	w := NewWizardForEdit(loaded)
	w.profileName = "prod"

	err = runWizardSaveExpectingError(t, &w, func() {
		// The gap: someone else creates `prod` while the save is in flight.
		writeProfile(t, "prod", `{"telemetry": true}`)
	})

	var exists *profile.ExistsError
	if !errors.As(err, &exists) {
		t.Fatalf("expected *profile.ExistsError, got %v", err)
	}

	// Neither profile may be damaged by the rejected rename.
	prod, err := profile.Load("prod")
	if err != nil {
		t.Fatalf("load prod: %v", err)
	}
	if prod.Config.Telemetry == nil || !*prod.Config.Telemetry {
		t.Fatal("the concurrently created profile was overwritten")
	}
	if !profile.Exists("dev") {
		t.Fatal("a rejected rename must leave the source profile in place")
	}
}

// A brand-new profile has no original name and no synchronous precheck at all,
// so the transactional guard is its only protection.
func TestWizardCreateRejectsNameClaimedDuringWizard(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	w := NewWizard()
	w.profileName = "prod"

	err := runWizardSaveExpectingError(t, &w, func() {
		writeProfile(t, "prod", `{"telemetry": true}`)
	})

	var exists *profile.ExistsError
	if !errors.As(err, &exists) {
		t.Fatalf("expected *profile.ExistsError, got %v", err)
	}

	prod, err := profile.Load("prod")
	if err != nil {
		t.Fatalf("load prod: %v", err)
	}
	if prod.Config.Telemetry == nil || !*prod.Config.Telemetry {
		t.Fatal("the concurrently created profile was overwritten")
	}
}

// The mirror race: the source disappears while the save is in flight. A rename
// whose source is gone must fail like profile.Rename does, not resurrect the
// deleted profile under the new name.
func TestWizardRenameRejectsDeletedSource(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeProfile(t, "dev", `{"telemetry": false}`)

	loaded, err := profile.Load("dev")
	if err != nil {
		t.Fatalf("load dev: %v", err)
	}

	w := NewWizardForEdit(loaded)
	w.profileName = "prod"

	err = runWizardSaveExpectingError(t, &w, func() {
		if err := profile.Delete("dev"); err != nil {
			t.Fatalf("delete dev: %v", err)
		}
	})

	var notFound *profile.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *profile.NotFoundError, got %v", err)
	}

	if profile.Exists("prod") {
		t.Fatal("a rename with a deleted source must not resurrect it under the new name")
	}
}

// Edit mode asserts the profile exists. An in-place edit racing a delete must
// report the deletion rather than silently undoing it — the rename path already
// did, and the two must not disagree.
func TestWizardInPlaceEditRejectsDeletedProfile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeProfile(t, "dev", `{"telemetry": false}`)

	loaded, err := profile.Load("dev")
	if err != nil {
		t.Fatalf("load dev: %v", err)
	}

	w := NewWizardForEdit(loaded)

	err = runWizardSaveExpectingError(t, &w, func() {
		if err := profile.Delete("dev"); err != nil {
			t.Fatalf("delete dev: %v", err)
		}
	})

	var notFound *profile.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected *profile.NotFoundError, got %v", err)
	}
	if profile.Exists("dev") {
		t.Fatal("an in-place edit must not resurrect a deleted profile")
	}
}

// Editing a profile in place still overwrites it — that is the whole point.
func TestWizardInPlaceEditStillOverwrites(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeProfile(t, "dev", `{"telemetry": false}`)

	loaded, err := profile.Load("dev")
	if err != nil {
		t.Fatalf("load dev: %v", err)
	}

	w := NewWizardForEdit(loaded)
	w.selection.SetSelected("telemetry", true)
	enabled := true
	w.config.Telemetry = &enabled

	if saved, _ := saveWizardAndRead(t, &w); saved == "" {
		t.Fatal("expected the in-place edit to save")
	}
}

// runWizardSaveExpectingError claims the name in the true race window: after
// the review step and the wizard's own synchronous precheck have both passed,
// while the save command is in flight on its goroutine.
func runWizardSaveExpectingError(t *testing.T, w *Wizard, claimName func()) error {
	t.Helper()

	w.step = StepReview
	w.reviewStep.SetConfig(w.profileName, &w.config, w.selection, w.preservedUnknown)
	if !w.reviewStep.IsValid() {
		t.Fatalf("expected the name to validate before the race: %#v", w.reviewStep.GetErrors())
	}

	_, cmd := w.Update(WizardNextMsg{})
	if cmd == nil {
		t.Fatalf("expected save command from review step, wizard error: %v", w.err)
	}

	claimName()

	result := cmd()
	saveDone, ok := result.(wizardSaveDoneMsg)
	if !ok {
		t.Fatalf("expected wizardSaveDoneMsg, got %T", result)
	}
	if saveDone.err == nil {
		t.Fatal("expected the save to be rejected, got success")
	}
	return saveDone.err
}
