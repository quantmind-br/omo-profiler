# TUI ARCHITECTURE

## OVERVIEW

Bubble Tea MVU application. Root `App` model (`app.go`, 840 lines) acts as state container, router, and composition layer for all sub-views.

## FILES

| File | Lines | Role |
|------|-------|------|
| `app.go` | 840 | Root model: state machine, message router, global overlays |
| `app_test.go` | — | State transitions, message routing, toast system tests |
| `tui.go` | ~20 | `Run()` entry point: `NewApp()` → `tea.NewProgram()` |
| `styles.go` | — | **Shared palette**: `Purple`, `Magenta`, `Cyan`, `Green`, `Red`, `Yellow`, `Gray`, `White` + component styles |
| `layout/layout.go` | — | `MinTerminalWidth/Height`, `MaxFieldWidth`, responsive width helpers |

## STATE MACHINE

10 states via `appState` enum:

```
stateDashboard → stateList → stateWizard → stateDiff
                            → stateImport → stateExport
                            → stateModels → stateModelImport
                            → stateTemplateSelect → stateSchemaCheck
```

Transitions via `navigateTo(state)`: updates state → re-initializes target view → returns `Init()` cmd.

## GLOBAL OVERLAYS

Rendered in `App.View()` over active view content:

1. **Toast**: `showToast(text, type, duration)` → auto-clears via `tea.Tick` + `clearToastMsg`
2. **Help**: Context-aware `renderShortHelp()` / `renderFullHelp()` per state
3. **Spinner**: `App.loading = true` replaces content with loading overlay
4. **Min-Size Guard**: `belowMinSize` blocks rendering if terminal too small

## NAVIGATION PROTOCOL

Views are decoupled from routing:
1. View emits `NavTo*Msg` (e.g., `NavToWizardMsg`, `NavToDiffMsg`)
2. `App.Update` intercepts → calls `navigateTo(newState)`
3. Views re-created on every navigation (no persisted state)

## ASYNC OPERATIONS

Business logic wrapped in `tea.Cmd` factory methods on `App`:
- `doSwitchProfile` → `switchProfileDoneMsg{name, err}`
- `doDeleteProfile` → `deleteProfileDoneMsg{name, err}`
- `doImportProfile` → `importProfileDoneMsg{profileName, hadCollision, err}`
- `doExportProfile` → `exportProfileDoneMsg{path, err}`

## ANTI-PATTERNS

- **Direct State Mutation**: Views must NEVER modify `App` fields; always emit `tea.Msg`
- **Blocking in Update**: File I/O must happen in `tea.Cmd`, never in `Update()` or `View()`
- **Persisted View State**: Views are re-created on navigation; don't rely on state persistence
- **Raw Styles in Views**: Import colors/styles from `styles.go`; never define hex colors locally