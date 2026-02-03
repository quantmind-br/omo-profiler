# TUI/VIEWS

## OVERVIEW

16 files implementing Bubble Tea views. Dominated by Profile Wizard logic. `Wizard` acts as parent component orchestrating 6 sequential configuration steps.

## WIZARD FLOW

State machine driven by `wizard.go` `Update` loop:

1. **Name** (`StepName`): Profile naming & validation
2. **Categories** (`StepCategories`): Model groupings
3. **Agents** (`StepAgents`): Complex form (~1000 lines) for agent config
4. **Hooks** (`StepHooks`): Event triggers
5. **Other** (`StepOther`): Misc settings
6. **Review** (`StepReview`): Final JSON validation & persistence

## KEY COMPONENTS

### Wizard State Machine
- **Orchestrator**: `Wizard` struct holds `config.Config` and manages transitions
- **Data Flow**: Steps read config via `SetConfig()`, mutate local state, write back via `Apply()` on exit
- **Implicit Interface**: All steps implement:
  - `SetConfig(*config.Config)`: Load state
  - `Apply(*config.Config)`: Save state
  - `SetSize(w, h)`: Responsive layout
  - `Init() tea.Cmd`: Lifecycle hook

### Views
- **List**: Dashboard with filtering/actions (`list.go`)
- **Diff**: Side-by-side profile comparison (`diff.go`)

## MESSAGE PROTOCOL

| Message | Trigger | Action |
|---------|---------|--------|
| `WizardNextMsg` | `Enter`/`Tab` | Call `Apply()`, increment step, `Init()` next |
| `WizardBackMsg` | `Esc`/`Shift+Tab` | Decrement step, `Init()` prev |
| `WizardSaveMsg` | Review Step | Persist profile to disk, return to Dashboard |
| `WizardCancelMsg` | `Ctrl+C` | Discard changes, return to Dashboard |

## ANTI-PATTERNS

- **Direct Config Mutation**: Steps must NOT modify `config.Config` directly; use `Apply()` only
- **Blocking Operations**: Disk I/O (Save/Load) should be wrapped in `tea.Cmd`
- **Hardcoded Dimensions**: Always use `SetSize()` provided by parent
- **Incomplete Features**: `wizard_hooks.go` has known TODOs ("todo-continuation-enforcer")
