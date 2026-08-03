# Architecture & Patterns

## TUI Architecture (Bubble Tea MVU)
The TUI follows the Model-View-Update pattern:

### State Machine (10 states)
- **Centralized State**: `App` struct in `internal/tui/app.go` (~840 lines) holds all state
- **State Enum**: `appState` iota defines all view states:
  ```
  stateDashboard | stateList | stateWizard | stateDiff
  stateImport | stateExport | stateModels | stateModelImport
  stateTemplateSelect | stateSchemaCheck
  ```
- **Transitions**: Via `navigateTo(state)` which re-initializes target views
- **Views re-created on every navigation** (no persisted state between transitions)

### View Composition (App struct fields)
```go
dashboard      views.Dashboard
list           views.List
wizard         views.Wizard
diff           views.Diff
modelRegistry  views.ModelRegistry
modelImport    views.ModelImport
importView     views.Import
exportView     views.Export
templateSelect views.TemplateSelect
schemaCheck    views.SchemaCheck
```

### Message-Driven Navigation
Views NEVER mutate App state directly. They emit messages:
```go
// View emits message
case key.Matches(msg, Keys.Enter):
    return m, func() tea.Msg { return NavToWizardMsg{} }

// App intercepts and handles navigation
case views.NavToWizardMsg:
    a.wizard = views.NewWizard()
    return a.navigateTo(stateWizard)
```

**Navigation messages by view:**
| View | Messages Emitted |
|------|-----------------|
| Dashboard | `NavToListMsg`, `NavToWizardMsg`, `NavToEditorMsg`, `NavToDiffMsg`, `NavToImportMsg`, `NavToExportMsg`, `NavToModelsMsg`, `NavToTemplateSelectMsg` |
| List | `SwitchProfileMsg`, `EditProfileMsg`, `DeleteProfileMsg`, `NavigateToWizardMsg`, `NavigateToDashboardMsg` |
| Wizard | `WizardNextMsg`, `WizardBackMsg`, `WizardSaveMsg`, `WizardCancelMsg` |
| Import | `ImportDoneMsg`, `ImportCancelMsg` |
| Export | `ExportDoneMsg`, `ExportCancelMsg` |
| ModelRegistry | `ModelRegistryBackMsg`, `ModelSavedMsg`, `ModelDeletedMsg`, `NavToModelImportMsg` |
| ModelImport | `ModelImportBackMsg`, `ModelImportDoneMsg` |
| TemplateSelect | `NavToWizardFromTemplateMsg`, `TemplateSelectCancelMsg` |
| SchemaCheck | `SchemaCheckBackMsg`, `NavToSchemaCheckMsg` |
| ModelSelector | `ModelSelectedMsg`, `ModelSelectorCancelMsg`, `PromptSaveCustomMsg` |

### Async Operations (tea.Cmd factories on App)
- `doSwitchProfile` → `switchProfileDoneMsg{name, err}` (emits export command; no document mutation)
- `doDeleteProfile` → `deleteProfileDoneMsg{name, err}`
- `doImportProfile` → `importProfileDoneMsg{profileName, hadCollision, err}`
- `doExportProfile` → `exportProfileDoneMsg{path, err}`

### Global UI Components
- **Toast System**: `showToast(text, type, duration)` → auto-clears via `tea.Tick` + `clearToastMsg`
- **Help Bubble**: Context-aware `renderShortHelp()` / `renderFullHelp()` per state
- **Loading Spinner**: `App.loading = true` replaces content with loading overlay
- **Min-Size Guard**: `belowMinSize` blocks rendering if terminal too small

### Layout Constants (`internal/tui/layout/layout.go`)
- `MinTerminalWidth`: 40, `MinTerminalHeight`: 12
- `MaxFieldWidth`: 120
- `IsCompact(width)`: `< 60` — views should use simpler layouts
- `IsShort(height)`: `< 20` — views should reduce vertical spacing
- `HelpBarHeight(height)`: 1 if `< 16`, otherwise 2

## Wizard Pattern
Multi-step form with 6 sequential steps and implicit interface:

**WizardStep interface (`internal/tui/views/step.go`):**
```go
type WizardStep interface {
    Init() tea.Cmd
    SetSize(width, height int)
    View() string
}
```
**Implicit methods** (called by orchestrator, not in interface):
- `SetConfig(*config.Config)` — load state from config (except WizardName: uses `SetName`/`GetName`)
- `Apply(*config.Config)` — save state to config

**Step sequence:**
1. `StepName` → `wizard_name.go` (Profile naming & rename logic)
2. `StepCategories` → `wizard_categories.go` (Category CRUD)
3. `StepAgents` → `wizard_agents.go` (Agent configuration forms)
4. `StepHooks` → `wizard_hooks.go` (Event trigger config)
5. `StepOther` → `wizard_other.go` (Large-scale settings, ~2460 lines, 50+ fields)
6. `StepReview` → `wizard_review.go` (Final JSON validation & save confirmation)

## Config Schema Safety
`internal/config/types.go` models the flat `[opencode]` block (`config.Config`):
- JSON tags must match the `[opencode]` sub-schema exactly
- Use pointers (`*bool`) to distinguish false from missing
- Keep structs pure — no methods, only data
- Use `json.RawMessage` for flexible fields like `skills`
- All JSON tags require `omitempty` to avoid dirty config files
- The whole omo file is `config.Document`; forms use `schema.GetOpenCodeSchema()`

## Testing Patterns
- Tests co-located with source (`*_test.go`)
- **MANDATORY**: Use `config.SetBaseDir(t.TempDir())` for filesystem isolation
- Always call `defer config.ResetBaseDir()` or use cleanup helper
- Seed profiles into the omo document (`SetProfileBlock` + `Save`), not as separate files
- Example:
```go
func setupTestEnv(t *testing.T) func() {
    tmpDir := t.TempDir()
    config.SetBaseDir(tmpDir)
    return func() { config.ResetBaseDir() }
}
```
- Assertions via `github.com/stretchr/testify`

## Profile Persistence
- Storage: `profiles.<name>` inside `~/.omo/omo.json` (or `omo.jsonc` when present)
- Editable payload: `profiles.<name>.[opencode]` (`config.Config`)
- Activation: env-driven (`Apply` + `ActiveName` root comparison)
- `Apply` substitutes profile keys into the document root (snapshotting the previous root if unmatched) (UI hint only — not authoritative)
- Comparison: `ActiveName` detects the applied profile via `canonicalJSON` comparison

## Backup System (`internal/backup/`)
- `Create(config.OmoFile())` → timestamped backup (`omo.json.bak.YYYY-MM-DD-HHMMSS`) before mutating writes
- Switch does not mutate the document and needs no backup
- `List()` → all backups sorted most recent first
- `Restore(backupPath)` → overwrites current omo file with backup
- `Clean(keepLast)` → rotation, removes old backups beyond threshold

## Anti-Patterns (NEVER DO)
- Direct state mutation in views (always use messages)
- Blocking I/O in Update() (use tea.Cmd)
- Hardcoded paths (use `config.OmoDir()` / `config.OmoFile()` / `config.ModelsFile()`)
- Raw hex colors in views (use `internal/tui/styles.go`)
- Modifying config.Config directly in wizard steps (use Apply())
- Copying or symlinking to "activate" a profile (activation is env-driven)
- Treating `.omo-profiler-selection` / any `.active-profile` as authoritative activation
- Type suppression (`as any`, `//nolint`, etc.)

## Sparse profile writes

Wizard/editor sparse saves: `profile.MarshalSparse` → `profile.WriteOpenCodeBlockInto` / `SaveOpenCodeBlock` into `profiles.<name>.[opencode]`. Do not use `Profile.Save` for that path (`omitempty` drops explicit zeros). Activation remains env-only via `Apply` → `export profile via `Apply`<name>`.
