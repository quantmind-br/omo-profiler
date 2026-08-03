# TUI Architecture

The TUI is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.10 and follows the standard **Model-View-Update (MVU)** pattern. The root `App` model (`internal/tui/app.go`) owns the state machine, global overlays, and delegation to sub-views.

## State Machine

The TUI has **10 states**, managed by `App.state` (`appState` enum):

```
stateDashboard (0)
  → stateList
  → stateWizard
  → stateDiff
  → stateImport
  → stateExport
  → stateModels
  → stateModelImport
  → stateTemplateSelect
  → stateSchemaCheck
```

All transitions flow through `navigateTo(state)`, which:
1. Saves the previous state in `prevState`
2. Sets the new `state`
3. Calls `SetSize()` then `Init()` on the target view
4. Returns the init `tea.Cmd`

**Views are recreated fresh on every navigation** — no view state is persisted between navigations.

## MVU Pattern

### Root App Model (`internal/tui/app.go`)

```
type App struct {
    state     appState
    prevState appState
    width     int
    height    int
    ready     bool
    loading   bool
    showHelp  bool
    // One sub-view field per state:
    dashboard views.Dashboard
    list      views.List
    wizard    views.Wizard
    diff      views.Diff
    importV   views.ImportView
    exportV   views.ExportView
    models    views.ModelRegistry
    modelImp  views.ModelImport
    tmplSel   views.TemplateSelect
    schemaChk views.SchemaCheck
    // Global overlays
    spinner   spinner.Model
    toast     toastMsg
    help      help.Model
}
```

### Message Flow

```
tea.KeyMsg / tea.WindowSizeMsg / custom Msg
  → App.Update(msg)
    → Phase 1: Global handling
      - tea.KeyMsg → quit/help/back (unless view is capturing keys)
      - tea.WindowSizeMsg → resize all sub-views
      - Navigation messages (NavTo*Msg, *DoneMsg, *BackMsg)
      - Toast messages
    → Phase 2: Delegate to current sub-view's Update
      - switch a.state { ... view.Update(msg) ... }
  → Returns (tea.Model, tea.Cmd)
```

### Async Operations

All I/O runs inside `tea.Cmd` closures:

| Command | Posts back |
|---------|-----------|
| `doSwitchProfile(name)` | `switchProfileDoneMsg{name, err}` |
| `doDeleteProfile(name)` | `deleteProfileDoneMsg{name, err}` |
| `doImportProfile(path)` | `importProfileDoneMsg{name, hadCollision, err}` |
| `doExportProfile(name, path)` | `exportProfileDoneMsg{path, err}` |

Never block inside `Update()` or `View()`.

## Views (`internal/tui/views/`)

The package contains **18 view files** and a shared `step.go` for wizard step constants:

| View File | State | Purpose |
|-----------|-------|---------|
| `dashboard.go` | `stateDashboard` | Main menu — 9 items (Switch, Create, Template, Edit, Compare, Models, Import, Export, Schema Check) |
| `list.go` | `stateList` | Profile list with filtering (`/`), switch/edit/delete/create |
| `wizard.go` | `stateWizard` | Multi-step form orchestrator (new/edit/template modes) |
| `wizard_name.go` | — | Step 1: Profile name input |
| `wizard_categories.go` | — | Step 2: Toggle categories on/off (tree view) |
| `wizard_agents.go` | — | Step 3: Configure agent-specific settings (tree view) |
| `wizard_hooks.go` | — | Step 4: Configure hook commands |
| `wizard_other*.go` | — | Step 5: Miscellaneous config fields (complex tree view, split across config/fields/render/update files) |
| `wizard_review.go` | — | Step 6: Final review + schema validation + async save |
| `diff.go` | `stateDiff` | Side-by-side profile comparison (dual viewport) |
| `model_registry.go` | `stateModels` | Browse/manage registered models with fuzzy search |
| `model_import.go` | `stateModelImport` | Import models from models.dev API |
| `model_search.go` | — | Model search helper |
| `model_selector.go` | — | Model selector sub-view |
| `import.go` | `stateImport` | Import profile from JSON file |
| `export.go` | `stateExport` | Export profile to disk |
| `template_select.go` | `stateTemplateSelect` | Pick a template for new profile |
| `schema_check.go` | `stateSchemaCheck` | Validate schema and check upstream drift (4 sub-states) |
| `step.go` | — | Step constants and type definitions |

### Complexity Hotspots

- **`wizard_other*.go`** (~100KB across 5 files) — The "Other Settings" step has a deeply nested configurable tree for all miscellaneous config fields. Split across config, fields, render, and update files for manageability.
- **`wizard_agents.go`** (~78KB) — Tree view for per-agent settings with model selection, fallbacks, permissions.
- **`wizard_categories.go`** (~39KB) — Tree view for category toggling with custom model support.

## Wizard Steps

The wizard is a linear 6-step flow:

```
StepName (1) → StepCategories (2) → StepAgents (3) → StepHooks (4) → StepOther (5) → StepReview (6)
```

Navigation: `Tab` / `Enter` → next step; `Shift+Tab` → previous step; `Ctrl+S` → save from any step.

**Step lifecycle** (defined in `wizard.go` and each step's update file):
- On activation: `SetConfig(&cfg, selection)` loads config into step-local state
- During step: user edits local state through keyboard interactions
- On exit: `Apply(&cfg, selection)` writes edits back to the config
- On save/review: validates against JSON schema (using `ValidateForSave` — permissive mode)

The wizard supports three modes:
1. **New** — start from defaults
2. **Edit** — load existing profile for editing
3. **Template** — start from a template profile

## Layout System (`internal/tui/layout/layout.go`)

- **Minimum size guard**: `MinTerminalWidth=40`, `MinTerminalHeight=14`. When below, a centered warning replaces the TUI.
- **Responsive helpers**: `IsCompact(width)` for `<60`, `IsShort(height)` for `<20`
- **Field width utilities**: `MediumFieldWidth`, `WideFieldWidth`, `FixedSmallWidth`
- **Truncation**: `TruncateWithEllipsis` using `go-runewidth`

## Keybinding Patterns (`internal/tui/keys.go`)

Global keymap (`Keys` struct):

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit (or cancel wizard with toast) |
| `?` | Toggle full help overlay |
| `Esc` | Back to dashboard |
| `Up`/`k`, `Down`/`j` | Navigate |
| `Enter` | Select/activate |
| `Tab` / `Shift+Tab` | Next/previous (wizard, diff panes) |

**View-specific keymaps** are defined in each view file (e.g., `dashboardKeyMap`, `listKeyMap`, `wizardKeyMap`).

**Global vs. local dispatch**: The root `App.Update` intercepts `Keys.Quit`, `Keys.Help`, and `Keys.Back` first. If the current view is in a "capturing" mode (filtering in list view, editing in model registry), it must signal `IsFiltering()` or `IsEditing()` so the root breaks and lets the view handle the key.

**Known issue**: The profile list's filter mode (`/`) does not properly guard against global `q`, `?`, and `Esc` — these keys are captured by the root handler before the list view processes them. Fix: add `stateList && list.IsFiltering()` guards to the Quit, Help, and Back cases in `app.go`. The Models view is the reference implementation that handles this correctly.

## Global Overlays

The root `App.View()` renders in priority order:

1. **Min-Size Guard** — when terminal is too small
2. **Loading Spinner** — while async operations are in flight
3. **Full Help** — toggled by `?`
4. **Active Sub-View** — the current state's view
5. **Toast Bar** — success/error/info notifications (auto-dismiss after duration)
6. **Short Help Bar** — always at the bottom (1-2 lines, context-dependent)

## Style Management

- All styles defined in `internal/tui/styles.go` using `lipgloss.NewStyle()`
- Views import styles from the central location; do not define raw hex colors inline
- The imported `github.com/charmbracelet/lipgloss` and `github.com/charmbracelet/bubbles` packages provide the styling primitives

## Change Guidance

When modifying the TUI:

1. **Add a new view state**: Add to `appState` enum, add a field to `App`, handle in `navigateTo()`, add dispatch cases in `Update()` and `View()`.
2. **Add a wizard step**: Update `wizardStep` constants in `step.go`, add step case in the wizard's update/render dispatch.
3. **Add a global keybinding**: Add to `Keys` in `keys.go`, handle in Phase 1 of `App.Update()`.
4. **Guard global keys in capturing views**: Check `IsFiltering()`/`IsEditing()` before intercepting Quit/Help/Back.
5. **Test I/O**: Use `tea.Cmd` closures — never block `Update()` or `View()`.
6. **Run wizard tests**: `go test -v ./internal/tui/views/ -run TestWizard` — tests are large and thorough.
7. **Keybinding change**: Update `keybindings_test.go` which inventories all 46+ bindings.
