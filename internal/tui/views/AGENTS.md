# Views Knowledge Base

## OVERVIEW
Collection of Bubble Tea view components, primarily centered around a 6-step configuration wizard.

## WIZARD ARCHITECTURE
The `Wizard` struct (`wizard.go`) acts as a stateful orchestrator for 6 steps:
1. `StepName` (WizardName): Profile naming.
2. `StepCategories` (WizardCategories): Model categorization.
3. `StepAgents` (WizardAgents): Detailed agent configuration (longest step).
4. `StepHooks` (WizardHooks): Event hooks setup.
5. `StepOther` (WizardOther): Misc settings.
6. `StepReview` (WizardReview): Final validation and save.

Data flow: `Wizard` holds `config.Config`. Steps receive copy via `SetConfig()`, mutate internal state, and write back via `Apply()` on navigation.

## MESSAGE PATTERNS
| Message | Source | Purpose |
|---------|--------|---------|
| `WizardNextMsg` | Step View | Advance to next step (triggers `Apply`) |
| `WizardBackMsg` | Step View | Return to previous step |
| `WizardSaveMsg` | Review View | Finalize creation/edit (payload: `*profile.Profile`) |
| `WizardCancelMsg` | Any View | Abort wizard and return to dashboard |
| `NavTo*Msg` | Dashboard | Global navigation request |

## CONVENTIONS
- **Responsive Layout**: All views implement `SetSize(w, h)` to adjust viewports/lists.
- **Config Mutation**: Steps never mutate shared config directly; only on `Apply()`.
- **Keybindings**: Consistent usage of `WizardKeyMap` (Tab/Enter=Next, Shift+Tab/Esc=Back).
- **Edit Mode**: `NewWizardForEdit` pre-fills all steps; name changes trigger rename logic.

## KNOWN ISSUES
- `wizard_hooks.go`: incomplete implementation (search "todo-continuation-enforcer").
- `keybindings_test.go`: context-dependent Enter key behavior requires careful testing.
