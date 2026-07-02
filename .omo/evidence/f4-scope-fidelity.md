# F4 Scope Fidelity Check

## Guardrails

### 1. No schema divergence in `internal/config/types.go`
**PASS**

**Evidence**
- `GIT_MASTER=1 git log --oneline cf45700^..HEAD -- internal/config/types.go` returned no commits.
- `GIT_MASTER=1 git diff --name-only cf45700^..HEAD` lists 20 changed files and does not include `internal/config/types.go`.
- Full file history still shows only prior schema-alignment commits, not this plan range.

### 2. No loss of `json.RawMessage` / flexible field fidelity
**PASS**

**Evidence**
- `internal/config/types.go` still defines:
  - `Config.Skills json.RawMessage`
  - `Config.RuntimeFallback json.RawMessage`
  - `AgentConfig.FallbackModels interface{}`
  - `CategoryConfig.FallbackModels interface{}`
  - `PermissionConfig.Bash interface{}`
  - `GitMasterConfig.CommitFooter interface{}`
- Since `internal/config/types.go` was not touched in the plan range, these fidelity-preserving types were not regressed by the Optional Config Selection work.

### 3. No checkbox on profile name, wizard navigation, template selection controls, or non-persisted UI helpers
**PASS**

**Evidence**
- `internal/tui/views/wizard_name.go` remains a plain text-input step with validation only (`textinput.Model`, `SetName`, `validate`, `View`); no checkbox rendering or selection state exists.
- `internal/tui/views/template_select.go` remains a cursor-based list picker (`cursor`, up/down/select/cancel keymap); no checkbox rendering, inclusion toggles, or selection-state plumbing exists.
- `git diff --name-only cf45700^..HEAD` does not include either `wizard_name.go` or `template_select.go`.

### 4. No blanket removal of type/enum validation
**PASS**

**Evidence**
- `internal/schema/validator.go` still has strict `Validate` / `ValidateJSON` methods that report all schema errors.
- Sparse-save relaxation is isolated to `ValidateForSave` / `ValidateJSONForSave`, which only skip `required` errors via `isRequiredError`.
- `go test ./internal/schema/... -run TestValidateStrictStillReportsRequiredness -v` passed.
- `TestValidateStrictStillReportsRequiredness` asserts `Validate(&config.Config{})` still returns required-field errors.

### 5. No save path that re-materializes zero/default values just because a view was visited
**PASS**

**Evidence**
- `WizardCategories.Apply`, `WizardAgents.Apply`, `WizardHooks.Apply`, and `WizardOther.Apply` all gate writes on explicit selection checks (`isCategoryFieldSelected`, `isAgentFieldSelected`, `selection.IsSelected`, `fieldSelected` / `selectedWithPrefix`).
- Step tests confirm omission behavior:
  - `TestWizardCategoriesApplyWritesOnlySelectedFields`
  - `TestWizardAgentsApplyWritesOnlySelectedFields`
  - `TestWizardHooksApplyWritesDisabledHooksOnlyWhenSelected`
  - `TestWizardHooksSelectedEmptyDisabledHooksSerializesAsEmptyArray`
  - `TestWizardOtherApplyWritesOnlySelectedFields`
  - `TestWizardOtherUntouchedSectionsRemainOmitted`
  - `TestWizardCheckedUncheckedTransitions_PreserveExplicitEmptyStringValue`
- Executed regressions passed:
  - `go test ./internal/tui/views -run 'TestWizard(AgentsApplyWritesOnlySelectedFields|CategoriesApplyWritesOnlySelectedFields|OtherApplyWritesOnlySelectedFields|OtherUntouchedSectionsRemainOmitted)|TestWizardCheckedUncheckedTransitions_PreserveExplicitEmptyStringValue' -v`
  - `go test ./internal/tui/views -run 'TestWizardHooksApplyWritesDisabledHooksOnlyWhenSelected' -v`
  - `go test ./internal/tui/views -run 'TestWizardHooksSelectedEmptyDisabledHooksSerializesAsEmptyArray' -v`

## Out-of-scope changes detected
None.

The changed-file set in `git diff --stat cf45700^..HEAD` matches the expected implementation surface: profile persistence/selection, schema validator, wizard steps/review, and related regression tests.

## Overall verdict
**APPROVE**

The Optional Config Selection implementation stayed within the plan guardrails. I found no schema drift, no fidelity regression in flexible/raw fields, no scope creep into non-persisted UI controls, no blanket loss of strict validation, and no evidence that merely visiting steps materializes defaults into saved output.

**VERDICT: APPROVE**
