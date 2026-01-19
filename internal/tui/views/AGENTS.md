# TUI Views

## OVERVIEW

15 view components following Bubble Tea Model-View-Update pattern.

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
| Editor | `editor.go` | Direct JSON editing |
| Diff | `diff.go` | Side-by-side profile comparison |
| ModelSelector | `model_selector.go` | Model picker popup |
| ModelRegistry | `model_registry.go` | Manage custom models |

## WIZARD FLOW

```
StepName → StepCategories → StepAgents → StepHooks → StepOther → StepReview
```

Each step: `WizardNextMsg` advances, `WizardBackMsg` retreats.

## MESSAGE PATTERNS

| Message | Emitted By | Handled By |
|---------|------------|------------|
| `SwitchProfileMsg` | List | app.go |
| `EditProfileMsg` | List | app.go |
| `DeleteProfileMsg` | List | app.go |
| `WizardSaveMsg` | WizardReview | app.go |
| `WizardCancelMsg` | Any wizard step | app.go |

## CONVENTIONS

- Use `SetSize(w, h)` for responsive layout
- Emit messages for actions, don't mutate parent state
- Use bubbles components: `list`, `textinput`, `viewport`, `spinner`
