# TUI Views

## OVERVIEW

16 view components following Bubble Tea Model-View-Update pattern.

## VIEW INVENTORY

| View | File | Purpose |
|------|------|---------|
| Dashboard | `dashboard.go` | Main menu, active profile display |
| List | `list.go` | Profile list with search, CRUD actions |
| Wizard | `wizard.go` | Multi-step profile creation/edit orchestrator |
| WizardName | `wizard_name.go` | Step 1: Profile name input |
| WizardCategories | `wizard_categories.go` | Step 2: Category model selection |
| WizardAgents | `wizard_agents.go` | Step 3: Agent configuration |
| WizardHooks | `wizard_hooks.go` | Step 4: Hook settings |
| WizardOther | `wizard_other.go` | Step 5: Misc settings |
| WizardReview | `wizard_review.go` | Step 6: Review and save |
| Diff | `diff.go` | Side-by-side profile comparison |
| ModelSelector | `model_selector.go` | Model picker popup |
| ModelRegistry | `model_registry.go` | Manage custom models |
| ModelImport | `model_import.go` | Import models from external sources |

## WIZARD ARCHITECTURE

```
StepName → StepCategories → StepAgents → StepHooks → StepOther → StepReview
```

- Orchestrator: `Wizard` struct in `wizard.go`
- Navigation: `WizardNextMsg` advances, `WizardBackMsg` retreats
- Each step is a sub-model with its own `Update`/`View`
- Fields: `step`, `profileName`, `config`, `editMode`, step sub-models

Constructors: `NewWizard()` (create), `NewWizardForEdit()` (edit existing)

## MESSAGE PATTERNS

| Message | Emitted By | Handled By |
|---------|------------|------------|
| `SwitchProfileMsg` | List | app.go |
| `EditProfileMsg` | List | app.go |
| `DeleteProfileMsg` | List | app.go |
| `WizardSaveMsg` | WizardReview | app.go |
| `WizardCancelMsg` | Any wizard step | app.go |
| `WizardNextMsg` | Step views | Wizard |
| `WizardBackMsg` | Step views | Wizard |

## CONVENTIONS

- Use `SetSize(w, h)` for responsive layout
- Emit messages for actions, don't mutate parent state
- Use bubbles components: `list`, `textinput`, `viewport`, `spinner`
- Wizard steps access shared `config` via orchestrator

## KNOWN ISSUES

- `wizard_hooks.go`: Contains TODO placeholder "todo-continuation-enforcer"
- `keybindings_test.go`: Enter key should NOT toggle in certain contexts
