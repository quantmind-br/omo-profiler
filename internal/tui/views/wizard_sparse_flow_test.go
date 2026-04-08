package views

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
)

func TestWizardSparseCreateFlow_SavesOnlyCheckedFields(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	w := NewWizard()
	if got := w.selection.SelectedPaths(); len(got) != 0 {
		t.Fatalf("expected blank selection for new wizard, got %v", got)
	}

	w.profileName = "sparse-create"
	w.config.HashlineEdit = boolPtr(false)
	w.config.BackgroundTask = &config.BackgroundTaskConfig{ProviderConcurrency: map[string]int{"openai": 0}}
	w.config.DefaultRunAgent = ""

	w.selection.SetSelected("hashline_edit", true)
	w.selection.SetSelected("background_task.provider_concurrency", true)

	preview := wizardReviewPreview(t, &w)
	decoded := decodeWizardJSON(t, preview)
	if len(decoded) != 2 {
		t.Fatalf("expected only checked keys in preview, got %#v", decoded)
	}
	if value, ok := decoded["hashline_edit"].(bool); !ok || value {
		t.Fatalf("expected hashline_edit=false, got %#v", decoded["hashline_edit"])
	}
	backgroundTask := wizardObject(t, decoded["background_task"], "background_task")
	providerConcurrency := wizardObject(t, backgroundTask["providerConcurrency"], "background_task.providerConcurrency")
	if value, ok := providerConcurrency["openai"].(float64); !ok || value != 0 {
		t.Fatalf("expected background_task.providerConcurrency.openai=0, got %#v", providerConcurrency["openai"])
	}
	if _, ok := decoded["default_run_agent"]; ok {
		t.Fatalf("expected unchecked empty string field to be omitted, got %#v", decoded)
	}

	saved, saveMsg := saveWizardAndRead(t, &w)
	if saveMsg.Profile == nil || saveMsg.Profile.Name != "sparse-create" {
		t.Fatalf("unexpected saved profile metadata: %#v", saveMsg.Profile)
	}
	if saved != preview {
		t.Fatalf("preview/save mismatch\npreview:\n%s\n\nsaved:\n%s", preview, saved)
	}

	reloaded, err := profile.Load("sparse-create")
	if err != nil {
		t.Fatalf("reload sparse-create: %v", err)
	}
	if reloaded.Config.HashlineEdit == nil || *reloaded.Config.HashlineEdit {
		t.Fatalf("expected reloaded hashline_edit=false, got %#v", reloaded.Config.HashlineEdit)
	}
	if reloaded.Config.BackgroundTask == nil || reloaded.Config.BackgroundTask.ProviderConcurrency["openai"] != 0 {
		t.Fatalf("expected reloaded background_task.providerConcurrency.openai=0, got %#v", reloaded.Config.BackgroundTask)
	}
	if reloaded.Config.DefaultRunAgent != "" {
		t.Fatalf("expected unchecked default_run_agent to remain omitted on reload, got %q", reloaded.Config.DefaultRunAgent)
	}
}

func TestWizardCheckedUncheckedTransitions_PreserveExplicitEmptyStringValue(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	w := NewWizard()
	w.profileName = "toggle-empty-string"
	w.config.DefaultRunAgent = ""

	w.selection.SetSelected("default_run_agent", true)
	selectedPreview := wizardReviewPreview(t, &w)
	selectedDecoded := decodeWizardJSON(t, selectedPreview)
	if value, ok := selectedDecoded["default_run_agent"].(string); !ok || value != "" {
		t.Fatalf("expected selected default_run_agent to be explicit empty string, got %#v", selectedDecoded["default_run_agent"])
	}

	w.selection.SetSelected("default_run_agent", false)
	uncheckedPreview := wizardReviewPreview(t, &w)
	uncheckedDecoded := decodeWizardJSON(t, uncheckedPreview)
	if _, ok := uncheckedDecoded["default_run_agent"]; ok {
		t.Fatalf("expected unchecked default_run_agent to be omitted, got %#v", uncheckedDecoded)
	}

	w.selection.SetSelected("default_run_agent", true)
	recheckedPreview := wizardReviewPreview(t, &w)
	if selectedPreview != recheckedPreview {
		t.Fatalf("expected rechecked preview to match original selection\nselected:\n%s\n\nrechecked:\n%s", selectedPreview, recheckedPreview)
	}

	recheckedDecoded := decodeWizardJSON(t, recheckedPreview)
	if value, ok := recheckedDecoded["default_run_agent"].(string); !ok || value != "" {
		t.Fatalf("expected rechecked default_run_agent to preserve empty string, got %#v", recheckedDecoded["default_run_agent"])
	}

	saved, _ := saveWizardAndRead(t, &w)
	if saved != recheckedPreview {
		t.Fatalf("preview/save mismatch after recheck\npreview:\n%s\n\nsaved:\n%s", recheckedPreview, saved)
	}
	if w.config.DefaultRunAgent != "" {
		t.Fatalf("expected wizard config value to survive toggle cycle, got %q", w.config.DefaultRunAgent)
	}
}

func TestWizardSparseEditRoundTrip_PreservesUnknownAndDropsUncheckedFields(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeRawProfileJSON(t, "existing-sparse", "{\n  \"custom\": \"data\",\n  \"disabled_mcps\": [\"mcp1\"],\n  \"hashline_edit\": false\n}")

	loaded, err := profile.Load("existing-sparse")
	if err != nil {
		t.Fatalf("load existing-sparse: %v", err)
	}

	w := NewWizardForEdit(loaded)
	if !w.selection.IsSelected("disabled_mcps") {
		t.Fatal("expected disabled_mcps to be pre-selected from field presence")
	}
	if !w.selection.IsSelected("hashline_edit") {
		t.Fatal("expected hashline_edit to be pre-selected from field presence")
	}
	if w.selection.IsSelected("default_run_agent") {
		t.Fatal("did not expect unrelated field to be selected")
	}

	w.selection.SetSelected("disabled_mcps", false)

	preview := wizardReviewPreview(t, &w)
	decoded := decodeWizardJSON(t, preview)
	if len(decoded) != 2 {
		t.Fatalf("expected custom + hashline_edit after unchecking disabled_mcps, got %#v", decoded)
	}
	if _, ok := decoded["disabled_mcps"]; ok {
		t.Fatalf("expected unchecked disabled_mcps to be omitted, got %#v", decoded)
	}
	if value, ok := decoded["hashline_edit"].(bool); !ok || value {
		t.Fatalf("expected hashline_edit=false, got %#v", decoded["hashline_edit"])
	}
	if value, ok := decoded["custom"].(string); !ok || value != "data" {
		t.Fatalf("expected unknown custom field to survive preview, got %#v", decoded["custom"])
	}

	saved, _ := saveWizardAndRead(t, &w)
	if saved != preview {
		t.Fatalf("preview/save mismatch\npreview:\n%s\n\nsaved:\n%s", preview, saved)
	}

	reloaded, err := profile.Load("existing-sparse")
	if err != nil {
		t.Fatalf("reload existing-sparse: %v", err)
	}
	if len(reloaded.Config.DisabledMCPs) != 0 {
		t.Fatalf("expected unchecked disabled_mcps to be removed, got %#v", reloaded.Config.DisabledMCPs)
	}
	if reloaded.Config.HashlineEdit == nil || *reloaded.Config.HashlineEdit {
		t.Fatalf("expected hashline_edit=false after round-trip, got %#v", reloaded.Config.HashlineEdit)
	}
	rawCustom, ok := reloaded.PreservedUnknown["custom"]
	if !ok {
		t.Fatal("expected custom unknown field to be preserved after save")
	}
	var custom string
	if err := json.Unmarshal(rawCustom, &custom); err != nil {
		t.Fatalf("decode preserved custom field: %v", err)
	}
	if custom != "data" {
		t.Fatalf("expected preserved custom field to equal %q, got %q", "data", custom)
	}
}

func TestWizardSparseTemplateFlow_UncheckedTemplateFieldsAreOmitted(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeRawProfileJSON(t, "template-sparse", "{\n  \"default_run_agent\": \"\",\n  \"disabled_hooks\": []\n}")

	w, err := NewWizardFromTemplate("template-sparse")
	if err != nil {
		t.Fatalf("NewWizardFromTemplate failed: %v", err)
	}
	if w.editMode {
		t.Fatal("template wizard should not be in edit mode")
	}
	if !w.selection.IsSelected("default_run_agent") {
		t.Fatal("expected template default_run_agent to be pre-selected")
	}
	if !w.selection.IsSelected("disabled_hooks") {
		t.Fatal("expected template disabled_hooks to be pre-selected")
	}

	w.profileName = "from-template"
	w.selection.SetSelected("disabled_hooks", false)

	preview := wizardReviewPreview(t, &w)
	decoded := decodeWizardJSON(t, preview)
	if len(decoded) != 1 {
		t.Fatalf("expected only checked template key in preview, got %#v", decoded)
	}
	if value, ok := decoded["default_run_agent"].(string); !ok || value != "" {
		t.Fatalf("expected default_run_agent to remain explicit empty string, got %#v", decoded["default_run_agent"])
	}
	if _, ok := decoded["disabled_hooks"]; ok {
		t.Fatalf("expected unchecked disabled_hooks to be omitted, got %#v", decoded)
	}

	saved, _ := saveWizardAndRead(t, &w)
	if saved != preview {
		t.Fatalf("preview/save mismatch\npreview:\n%s\n\nsaved:\n%s", preview, saved)
	}
}

func TestWizardSparseImportAdjacentFlow_PreviewMatchesSavedOutput(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	writeRawProfileJSON(t, "imported-sparse", "{\n  \"background_task\": {\n    \"providerConcurrency\": {\n      \"openai\": 0\n    }\n  },\n  \"custom_import\": {\"keep\": true},\n  \"default_run_agent\": \"\"\n}")

	loaded, err := profile.Load("imported-sparse")
	if err != nil {
		t.Fatalf("load imported-sparse: %v", err)
	}

	w := NewWizardForEdit(loaded)
	if !w.selection.IsSelected("default_run_agent") {
		t.Fatal("expected imported default_run_agent to be selected from presence")
	}
	if !w.selection.IsSelected("background_task.provider_concurrency") {
		t.Fatal("expected imported background_task.provider_concurrency to be selected from presence")
	}

	preview := wizardReviewPreview(t, &w)
	decoded := decodeWizardJSON(t, preview)
	customImport := wizardObject(t, decoded["custom_import"], "custom_import")
	if keep, ok := customImport["keep"].(bool); !ok || !keep {
		t.Fatalf("expected preserved imported custom_import.keep=true, got %#v", customImport["keep"])
	}
	if value, ok := decoded["default_run_agent"].(string); !ok || value != "" {
		t.Fatalf("expected imported default_run_agent to remain explicit empty string, got %#v", decoded["default_run_agent"])
	}
	backgroundTask := wizardObject(t, decoded["background_task"], "background_task")
	providerConcurrency := wizardObject(t, backgroundTask["providerConcurrency"], "background_task.providerConcurrency")
	if value, ok := providerConcurrency["openai"].(float64); !ok || value != 0 {
		t.Fatalf("expected imported background_task.providerConcurrency.openai=0, got %#v", providerConcurrency["openai"])
	}

	saved, saveMsg := saveWizardAndRead(t, &w)
	if saveMsg.Profile == nil || saveMsg.Profile.Name != "imported-sparse" {
		t.Fatalf("unexpected saved profile metadata: %#v", saveMsg.Profile)
	}
	if saved != preview {
		t.Fatalf("preview/save mismatch\npreview:\n%s\n\nsaved:\n%s", preview, saved)
	}
}

func wizardReviewPreview(t *testing.T, w *Wizard) string {
	t.Helper()

	w.step = StepReview
	w.reviewStep.SetConfig(w.profileName, &w.config, w.selection, w.preservedUnknown)
	if !w.reviewStep.IsValid() {
		t.Fatalf("expected valid review preview, got errors %#v", w.reviewStep.GetErrors())
	}
	return w.reviewStep.jsonPreview
}

func saveWizardAndRead(t *testing.T, w *Wizard) (string, WizardSaveMsg) {
	t.Helper()

	preview := wizardReviewPreview(t, w)

	updated, cmd := w.Update(WizardNextMsg{})
	if cmd == nil {
		t.Fatal("expected save command from review step")
	}

	result := cmd()
	saveDone, ok := result.(wizardSaveDoneMsg)
	if !ok {
		t.Fatalf("expected wizardSaveDoneMsg, got %T", result)
	}
	if saveDone.err != nil {
		t.Fatalf("wizard save failed: %v", saveDone.err)
	}

	updated, cmd = updated.Update(saveDone)
	if cmd == nil {
		t.Fatal("expected follow-up WizardSaveMsg command after save")
	}

	result = cmd()
	saveMsg, ok := result.(WizardSaveMsg)
	if !ok {
		t.Fatalf("expected WizardSaveMsg, got %T", result)
	}

	savedBytes, err := os.ReadFile(filepath.Join(config.ProfilesDir(), w.profileName+".json"))
	if err != nil {
		t.Fatalf("read saved profile: %v", err)
	}
	if string(savedBytes) != preview {
		t.Fatalf("preview/save mismatch\npreview:\n%s\n\nsaved:\n%s", preview, string(savedBytes))
	}

	*w = updated
	return string(savedBytes), saveMsg
}

func writeRawProfileJSON(t *testing.T, name, contents string) {
	t.Helper()

	path := filepath.Join(config.ProfilesDir(), name+".json")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write raw profile %s: %v", name, err)
	}
}

func decodeWizardJSON(t *testing.T, raw string) map[string]any {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("decode wizard JSON: %v\n%s", err, raw)
	}
	return decoded
}

func wizardObject(t *testing.T, value any, name string) map[string]any {
	t.Helper()

	decoded, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected %s to be an object, got %#v", name, value)
	}
	return decoded
}
