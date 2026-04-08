# TUI/VIEWS

## OVERVIEW

18 view files + 10 test files implementing Bubble Tea sub-views. Dominated by the 6-step Profile Wizard. `Wizard` (`wizard.go`) orchestrates sequential configuration steps.

## FILE MAP

| File | Lines | Role |
|------|-------|------|
| `wizard.go` | 499 | Orchestrator: step transitions, `NewWizard`/`NewWizardForEdit`/`NewWizardFromTemplate` |
| `step.go` | — | `WizardStep` interface definition (`Init`/`SetSize`/`View`) |
| `wizard_name.go` | — | Step 1: Profile naming + validation |
| `wizard_categories.go` | 1289 | Step 2: Category CRUD with dynamic form injection |
| `wizard_categories_defaults.go` | — | Category default prompts/model presets |
| `wizard_agents.go` | 2699 | Step 3: Agent config forms with nested viewport scrolling |
| `wizard_hooks.go` | 372 | Step 4: Event trigger configuration |
| `wizard_other.go` | 3786 | Step 5: Catch-all settings (50+ fields, 21 collapsible sections) |
| `wizard_review.go` | — | Step 6: JSON validation + persistence |
| `dashboard.go` | — | Main menu with active profile overview |
| `list.go` | — | Profile list with filtering, switch/edit/delete actions |
| `diff.go` | 447 | Side-by-side profile comparison with dual viewports |
| `import.go` | — | File-based profile import |
| `export.go` | — | Profile export to file |
| `model_selector.go` | 540 | Reusable searchable model dropdown (fuzzy, skip headers) |
| `model_registry.go` | 655 | Local model CRUD with in-place form swapping |
| `model_import.go` | 654 | Async models.dev fetcher with fuzzy filtering + multi-select |
| `template_select.go` | — | Profile template picker for wizard initialization |
| `schema_check.go` | — | Upstream schema diff viewer with save-to-file |

## WIZARD STEP INTERFACE

Explicit (`WizardStep` in `step.go`):
- `Init() tea.Cmd`
- `SetSize(w, h int)`
- `View() string`

Implicit (called by `Wizard` orchestrator):
- `SetConfig(*config.Config)` — load state from config before step activates
- `Apply(*config.Config)` — write local state back to config on step exit

## WIZARD DATA FLOW

```
Wizard holds config.Config
  → Step activates: Wizard calls step.SetConfig(&config)
  → User edits form fields (local state only)
  → Step exits: Wizard calls step.Apply(&config)
  → Next step activates with updated config
```

## MESSAGE PROTOCOL

| Message | Trigger | Action |
|---------|---------|--------|
| `WizardNextMsg` | Enter/Tab | `Apply()` current → increment step → `Init()` next |
| `WizardBackMsg` | Esc/Shift+Tab | Decrement step → `Init()` prev |
| `WizardSaveMsg` | Review confirm | Persist profile to disk → return to Dashboard |
| `WizardCancelMsg` | Ctrl+C | Discard → Dashboard |
| `NavTo*Msg` | Menu selection | Emitted by dashboard/list → intercepted by `App` |

## COMPLEXITY HOTSPOTS

- **wizard_other.go** (3786L): Manual focus management across 21 sections. Every upstream `Config` field change requires boilerplate in both `SetConfig` and `Apply`.
- **wizard_agents.go** (2699L): Nested "form-in-list" pattern with manual scroll offset calculations (`getLineForField`).
- **wizard_categories.go** (1289L): Dynamic form list — user can add/delete categories, requiring viewport rebuild.
- **model_selector.go**: Reused by both agents and categories steps. Heterogeneous list with non-selectable headers (`findNextSelectable`).

## ANTI-PATTERNS

- **Direct Config Mutation**: Steps must NOT modify `config.Config` directly; use `Apply()` only
- **Blocking Operations**: Disk I/O must be wrapped in `tea.Cmd`
- **Hardcoded Dimensions**: Always use `SetSize()` provided by parent
- **Local Style Definitions**: Import from `internal/tui/styles.go`; don't redeclare hex colors
- **Duplicated Model Logic**: `wizard_agents.go` and `wizard_categories.go` share custom-model-save logic that should be abstracted
