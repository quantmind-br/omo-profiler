# Optional Config Selection for TUI Profiles

## TL;DR
> **Summary**: Rework omo-profiler so every JSON-persisted TUI field is explicitly opt-in, unchecked fields are omitted from saved JSON, existing sparse profiles round-trip without losing presence semantics, and oh-my-openagent supplies runtime defaults for omissions.
> **Deliverables**:
> - Presence-aware profile load/edit/save pipeline
> - Per-field opt-in checkboxes across wizard config screens
> - Save-path validation that removes requiredness without losing type safety
> - Regression coverage for sparse create/edit/template/import round-trips
> **Effort**: XL
> **Parallel**: YES - 3 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Tasks 6-9 → Task 10 → Tasks 11-12

## Context
### Original Request
Add a feature so all settings managed by omo-profiler become optional. The user must choose exactly what to configure. Unconfigured options must not be written to the file. The generated file must contain only selected settings. Add a checkbox beside every TUI-configurable field so the user can decide what to configure. Remove requiredness validation so omitted settings rely on oh-my-openagent hardcoded defaults. Ground the plan in the code of both omo-profiler and the consumer app in `oh-my-openagent/`.

### Interview Summary
- Scope covers fields that persist into the final JSON config; profile name and auxiliary navigation controls stay outside the checkbox system.
- Unknown/legacy JSON fields in existing profiles must be preserved exactly through edit/save.
- Validation change is scoped to removing requiredness from the save/edit path; structural/type validation remains for values that are present.
- Test strategy is **tests-after**.
- Default applied: template-driven profiles start with checkboxes selected for any fields materialized by the chosen template; a brand-new blank profile may save as `{}`.

### Metis Review (gaps addressed)
- Preserve presence/absence independently from value, especially for explicit `false`/`0`/empty-like values in existing profiles.
- Do **not** mutate `internal/config/types.go` away from the upstream schema contract; keep opt-in state in a separate presence model owned by the TUI/profile pipeline.
- Prevent data loss for unknown/legacy keys by preserving raw JSON fragments through load/edit/save.
- Keep checkbox implementation minimal and consistent with current text-rendered patterns instead of turning this into a UI component rewrite.

## Work Objectives
### Core Objective
Deliver a sparse-config editing flow where the wizard saves only explicitly selected config fields, while preserving unknown fields and allowing oh-my-openagent to apply its own defaults for omissions.

### Deliverables
- A presence-aware representation of persisted config state for new, edited, imported, and template-based profiles.
- Wizard steps (categories, agents, hooks, other) updated with per-field opt-in checkboxes beside every JSON-persisted field they expose.
- Review/save pipeline that previews and writes sparse JSON only.
- Validation split between save-path sparse validation and strict schema-check behavior.
- Automated regression coverage for create/edit/template/import round-trips and unknown-field preservation.

### Definition of Done (verifiable conditions with commands)
- `go test ./internal/profile/... ./internal/tui/views/... ./internal/schema/... -v`
- `go test ./... -v`
- `make test`
- A fixture/profile with unknown keys survives load → edit/save → reload without losing unknown content.
- Saving a blank profile produces `{}` (or equivalent empty JSON object) with no materialized TUI defaults.

### Must Have
- Per-field opt-in state for every persisted TUI field.
- Existing profiles initialize checkbox state from actual JSON key presence, not inferred zero values.
- Unknown/legacy keys preserved exactly on save.
- Unchecked fields omitted from preview and persisted JSON.
- Omitted fields no longer fail save-path validation due to requiredness.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No schema divergence in `internal/config/types.go` (do not convert schema fields to UI-specific wrapper types).
- No loss of `json.RawMessage` / flexible field fidelity for skills, runtime fallback, fallback models, bash permission, or commit footer.
- No checkbox added to profile name, wizard navigation, template selection controls, or non-persisted UI helpers.
- No blanket removal of type/enum validation for present values.
- No save path that re-materializes zero/default values just because a view was visited.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after with Go stdlib + testify
- QA policy: Every task includes agent-executed scenarios using Go tests or deterministic command output
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = under-splitting.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: foundation for sparse persistence and validation (Tasks 1-5)
Wave 2: per-step checkbox retrofits (Tasks 6-9)
Wave 3: review/save integration and regression coverage (Tasks 10-12)

### Dependency Matrix (full, all tasks)
| Task | Depends On | Blocks |
|------|------------|--------|
| 1 | - | 2, 3, 4, 10, 11, 12 |
| 2 | 1 | 3, 4, 6, 7, 8, 9, 10, 11, 12 |
| 3 | 1, 2 | 6, 7, 8, 9, 10, 12 |
| 4 | 1, 2 | 5, 10, 11, 12 |
| 5 | 4 | 10, 11 |
| 6 | 2, 3 | 10, 12 |
| 7 | 2, 3 | 10, 12 |
| 8 | 2, 3 | 10, 12 |
| 9 | 2, 3 | 10, 12 |
| 10 | 3, 4, 5, 6, 7, 8, 9 | 11, 12 |
| 11 | 1, 4, 5, 10 | F1-F4 |
| 12 | 3, 6, 7, 8, 9, 10 | F1-F4 |

### Agent Dispatch Summary (wave → task count → categories)
- Wave 1 → 5 tasks → `deep`, `ultrabrain`, `unspecified-high`
- Wave 2 → 4 tasks → `visual-engineering`, `unspecified-high`
- Wave 3 → 3 tasks → `deep`, `unspecified-high`

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Preserve unknown/legacy JSON during profile load/save

  **What to do**: Add a non-schema persistence layer in `internal/profile/` that captures raw JSON for unknown/legacy keys and the original presence map for supported keys when a profile is loaded. Keep the typed `config.Config` as the editable value container, but store preserved raw fragments separately so edit/save can round-trip unsupported content without loss. Update every `profile.Load(...)` caller to receive the richer loaded state needed by the wizard and future sparse save logic.
  **Must NOT do**: Do not add UI-only fields to `internal/config/types.go`. Do not silently drop unknown nested objects, raw JSON blobs, or flexible fields such as skills/runtime fallback/fallback models.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: touches persistence boundaries and has to preserve data fidelity across load/save.
  - Skills: `[]` - No extra skill is needed beyond repo-native Go patterns.
  - Omitted: [`frontend-ui-ux`] - Reason: no visual redesign is involved.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 3, 4, 10, 11, 12 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/profile/profile.go:41-81` - current `Load`/`Save` entry points and the place where unknown fields are currently lost.
  - Pattern: `internal/tui/app.go:225-242` - active-profile edit path that must receive the richer loaded profile state.
  - Pattern: `internal/tui/app.go:400-407` - named-profile edit path that must receive the richer loaded profile state.
  - API/Type: `internal/config/types.go` - schema-constrained config types; keep them free of TUI presence metadata.
  - Test: `internal/profile/profile_test.go:12-19` - `setupTestEnv` isolation pattern for filesystem-backed profile tests.
  - External: `oh-my-openagent/src/plugin-config.ts` - consumer applies defaults when fields are absent, so sparse output is valid.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/profile -run 'Test(ProfileLoadPreservesUnknownJSON|ProfileLoadCapturesSupportedFieldPresence)' -v`
  - [x] `go test ./internal/profile -run 'Test(ProfileSaveRoundTripsPreservedUnknownFragments)' -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Preserve unknown nested JSON on load/save
    Tool: Bash
    Steps: run `go test ./internal/profile -run TestProfileSaveRoundTripsPreservedUnknownFragments -v | tee .sisyphus/evidence/task-1-preserve-unknown.log`
    Expected: test passes and confirms unknown nested keys/raw fragments survive load → save → reload unchanged
    Evidence: .sisyphus/evidence/task-1-preserve-unknown.log

  Scenario: Reject malformed source JSON cleanly
    Tool: Bash
    Steps: run `go test ./internal/profile -run TestProfileLoadFailsOnMalformedJSON -v | tee .sisyphus/evidence/task-1-preserve-unknown-error.log`
    Expected: test passes by asserting malformed JSON returns an error and does not create partial preserved state
    Evidence: .sisyphus/evidence/task-1-preserve-unknown-error.log
  ```

  **Commit**: YES | Message: `refactor(profile): preserve unknown json on sparse edits` | Files: `internal/profile/*`, `internal/tui/app.go`, related tests

- [x] 2. Introduce an explicit field-selection model for sparse configs

  **What to do**: Create a dedicated selection/presence structure outside `config.Config` (for example in `internal/profile/` or `internal/tui/views/`) keyed by canonical JSON paths for every field the wizard can persist. Provide constructors for three sources: blank new profile (all unchecked), loaded existing profile (checked from actual JSON presence), and template-based profile (checked for each field materialized by the template). Preserve value state independently from selection so unchecking a field omits it from output without erasing its in-memory value during the same wizard session.
  **Must NOT do**: Do not infer selection from zero values. Do not couple selection state to `omitempty`. Do not clear typed field values just because a checkbox is toggled off.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: this becomes the contract shared by load/edit/save and every wizard step.
  - Skills: `[]` - No extra skill required.
  - Omitted: [`frontend-ui-ux`] - Reason: logic-first infrastructure task.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 3, 4, 5, 6, 7, 8, 9, 10, 11, 12 | Blocked By: 1

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard.go:96-130` - `NewWizard`, `NewWizardForEdit`, and `NewWizardFromTemplate` are the bootstrap points for selection state.
  - Pattern: `internal/tui/views/wizard.go:251-341` - `nextStep` / `prevStep` define when values are committed and must also commit selection state.
  - Pattern: `internal/tui/views/wizard_categories.go:287-379` - existing SetConfig/Apply flow for nested category fields.
  - Pattern: `internal/tui/views/wizard_agents.go:625-786` - existing SetConfig/Apply flow for agent fields.
  - Pattern: `internal/tui/views/wizard_other.go:692-1577` - current field hydration/sentinel logic for the broadest surface area.
  - External: `oh-my-openagent/src/plugin-config.ts` - absent keys are safe; explicit selection is the only signal for persistence.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/profile ./internal/tui/views -run 'Test(SelectionStateFromBlankProfile|SelectionStateFromExistingJSON|SelectionStateFromTemplate)' -v`
  - [x] `go test ./internal/tui/views -run 'TestSelectionToggleRetainsFieldValueUntilSave' -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Existing profile seeds checkbox state from actual JSON presence
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestSelectionStateFromExistingJSON -v | tee .sisyphus/evidence/task-2-selection-state.log`
    Expected: test passes and proves explicit `false`/`0`/empty-like present values are still marked selected when the JSON key exists
    Evidence: .sisyphus/evidence/task-2-selection-state.log

  Scenario: Unchecking does not erase in-session value
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestSelectionToggleRetainsFieldValueUntilSave -v | tee .sisyphus/evidence/task-2-selection-state-error.log`
    Expected: test passes and proves a field can be unchecked, rechecked, and recover the prior typed value within the same wizard session
    Evidence: .sisyphus/evidence/task-2-selection-state-error.log
  ```

  **Commit**: YES | Message: `feat(wizard): add explicit config field selection state` | Files: new selection-state helpers, wizard bootstrap wiring, related tests

- [x] 3. Plumb selection-aware wizard state through every step

  **What to do**: Introduce a wizard-owned state envelope that carries `config.Config`, preserved raw JSON context, and the explicit selection model together. Update the concrete wizard step contracts so `SetConfig` and `Apply` receive both value state and selection state. Ensure `nextStep` commits both values and selection, while `prevStep` preserves the already-applied wizard state without silently re-materializing omitted fields. Wire new/edit/template entry points through the same state envelope.
  **Must NOT do**: Do not use globals/singletons for selection state. Do not let any step mutate selection for unrelated paths. Do not leave any step with an outdated signature that silently ignores selection.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: cross-cutting refactor with many call sites but bounded logic.
  - Skills: `[]` - Standard repo patterns are sufficient.
  - Omitted: [`frontend-ui-ux`] - Reason: plumbing, not presentation.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 5, 6, 7, 8, 9, 10, 12 | Blocked By: 1, 2

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/step.go:9-13` - baseline wizard-step interface shape.
  - Pattern: `internal/tui/views/wizard.go:73-458` - root wizard orchestrator and state transitions.
  - Pattern: `internal/tui/views/wizard_review.go:110-238` - review step currently rehydrates preview solely from `config.Config`.
  - Pattern: `internal/tui/views/wizard_categories.go:287-379` - example of current implicit `SetConfig` / `Apply` step contract.
  - Pattern: `internal/tui/views/wizard_agents.go:625-786` - example of step-local config hydration and write-back.
  - Test: `internal/tui/views/wizard_test.go` - place to add root-wizard state propagation tests.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizard(InitializesSelectionAwareStateForNewEditAndTemplate|NextStepCommitsSelectionAndValues|PrevStepKeepsAppliedSparseState)' -v`
  - [x] `go test ./internal/tui/views -run 'TestWizardAllStepsReceiveSelectionContext' -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Wizard bootstrap covers new, edit, and template flows
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardInitializesSelectionAwareStateForNewEditAndTemplate -v | tee .sisyphus/evidence/task-3-wizard-state.log`
    Expected: test passes and proves every entry point creates the same selection-aware state envelope with the correct initial selections
    Evidence: .sisyphus/evidence/task-3-wizard-state.log

  Scenario: No step silently drops selection context
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardAllStepsReceiveSelectionContext -v | tee .sisyphus/evidence/task-3-wizard-state-error.log`
    Expected: test passes and proves each step can round-trip selection-aware Apply/SetConfig without nil dereference or ignored paths
    Evidence: .sisyphus/evidence/task-3-wizard-state-error.log
  ```

  **Commit**: YES | Message: `refactor(wizard): carry selection state through all steps` | Files: `internal/tui/views/wizard*.go`, `internal/tui/views/step.go`, related tests

- [x] 4. Build the sparse JSON serializer used by preview and save

  **What to do**: Implement a deterministic sparse builder that creates the final JSON object from the typed config values plus the explicit selection model. Selected fields must serialize even when the value looks like a zero/default (`false`, `0`, `""`, empty list/object), while unchecked fields must be omitted entirely. After composing the selected known paths, deep-merge preserved unknown raw fragments, with explicitly selected known paths winning if they overlap. Expose the builder for both review preview and profile save so they share identical output.
  **Must NOT do**: Do not rely on `json.Marshal(config.Config)` alone for omission semantics. Do not let unknown preserved fragments overwrite user-edited known fields. Do not emit empty parent objects unless at least one child path is selected or preserved.

  **Recommended Agent Profile**:
  - Category: `ultrabrain` - Reason: requires careful handling of zero values, deep-merge precedence, and stable serialization semantics.
  - Skills: `[]` - No additional skill necessary.
  - Omitted: [`frontend-ui-ux`] - Reason: pure serialization logic.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 5, 10, 11, 12 | Blocked By: 1, 2

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/profile/profile.go:68-81` - current save path uses `json.MarshalIndent` directly and must be replaced/augmented.
  - Pattern: `internal/tui/views/wizard_review.go:116-162` - current preview builder marshals the full config and must switch to sparse output.
  - Pattern: `internal/config/types.go` - value source for typed known fields.
  - External: `oh-my-openagent/src/plugin-config.ts` - omitted keys are filled by consumer defaults, so sparse output is the intended contract.
  - External: `oh-my-openagent/src/plugin-handlers/category-config-resolver.ts` - missing categories fall back to `DEFAULT_CATEGORIES`.
  - External: `oh-my-openagent/src/config/schema/experimental.ts` - `task_system` is optional in the consumer schema, so omission is a first-class runtime state.
  - External: `oh-my-openagent/src/hooks/tasks-todowrite-disabler/hook.ts` - real consumer call site that exercises omitted `experimental.task_system` behavior.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/profile -run 'TestSparseSerializer(OmitsUncheckedFields|KeepsExplicitZeroValuesWhenSelected|MergesPreservedUnknownFragments)' -v`
  - [x] `go test ./internal/profile -run TestSparseSerializerProducesStablePrettyJSON -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Selected zero/default values remain present in sparse output
    Tool: Bash
    Steps: run `go test ./internal/profile -run TestSparseSerializerKeepsExplicitZeroValuesWhenSelected -v | tee .sisyphus/evidence/task-4-sparse-serializer.log`
    Expected: test passes and proves selected `false`, `0`, and empty-string values are still serialized
    Evidence: .sisyphus/evidence/task-4-sparse-serializer.log

  Scenario: Preserved unknown fragments cannot clobber managed paths
    Tool: Bash
    Steps: run `go test ./internal/profile -run TestSparseSerializerKnownPathsOverridePreservedUnknownFragments -v | tee .sisyphus/evidence/task-4-sparse-serializer-error.log`
    Expected: test passes and proves overlap resolution favors current known selected fields while leaving unrelated unknown fragments intact
    Evidence: .sisyphus/evidence/task-4-sparse-serializer-error.log
  ```

  **Commit**: YES | Message: `feat(profile): serialize sparse selected config output` | Files: `internal/profile/*`, `internal/tui/views/wizard_review.go`, related tests

- [x] 5. Split validation into sparse-save mode and strict mode

  **What to do**: Keep the embedded schema unchanged, but introduce a save-path validation entry point that ignores schema requiredness errors while still reporting type/enum/shape violations for fields that are present. Wire wizard review/save and any profile-import validation path to the sparse-save validator. Keep a strict validator entry point available for schema-focused checks/tests so the project does not lose upstream-schema visibility.
  **Must NOT do**: Do not edit the embedded schema to remove `required` keys. Do not suppress non-required validation failures. Do not change schema comparison/upstream diff behavior.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: validation semantics must change without creating schema drift.
  - Skills: `[]` - No extra skill required.
  - Omitted: [`frontend-ui-ux`] - Reason: no UI work.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 10, 11 | Blocked By: 4

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/schema/validator.go:64-89` - current validation API and error-shaping behavior.
  - Pattern: `internal/tui/views/wizard.go:251-341` - `nextStep` currently blocks progression on full validation failures.
  - Pattern: `internal/tui/views/wizard_review.go:116-162` - preview/validation coupling for review step.
  - Pattern: `internal/tui/views/import.go` - import flow should use the same sparse-save validation semantics if it validates profile content before save.
  - External: `oh-my-openagent/src/config/schema/oh-my-opencode-config.ts` - consumer treats top-level fields as optional and supplies defaults for omissions.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/schema -run 'Test(ValidateForSaveAllowsEmptyConfig|ValidateForSaveRejectsMalformedPresentValues)' -v`
  - [x] `go test ./internal/schema -run TestValidateStrictStillReportsRequiredness -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Empty sparse config is valid for save flow
    Tool: Bash
    Steps: run `go test ./internal/schema -run TestValidateForSaveAllowsEmptyConfig -v | tee .sisyphus/evidence/task-5-sparse-validation.log`
    Expected: test passes and proves `{}` no longer fails save-path validation
    Evidence: .sisyphus/evidence/task-5-sparse-validation.log

  Scenario: Present malformed values still fail validation
    Tool: Bash
    Steps: run `go test ./internal/schema -run TestValidateForSaveRejectsMalformedPresentValues -v | tee .sisyphus/evidence/task-5-sparse-validation-error.log`
    Expected: test passes and proves wrong-type or invalid enum values remain blocking when a field is selected/present
    Evidence: .sisyphus/evidence/task-5-sparse-validation-error.log
  ```

  **Commit**: YES | Message: `feat(schema): allow sparse save validation` | Files: `internal/schema/*`, wizard/import call sites, related tests

- [x] 6. Retrofit the categories step with per-field opt-in checkboxes

  **What to do**: Update `wizard_categories` so every persisted category field gets an adjacent inclusion checkbox driven by the shared selection model. Existing category entries loaded from JSON must start with checkboxes checked only for keys that were actually present. Keep checkbox state separate from field value state, and make `Apply` write only selected subfields for each category. If a category has zero selected persisted fields after apply, omit that category from sparse output entirely unless preserved unknown fragments require it to remain.
  **Must NOT do**: Do not treat zero values as “unchecked”. Do not overload the existing business booleans (`disable`, `isUnstable`) to mean field inclusion. Do not emit empty category objects.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: terminal-form UX change with substantial per-field rendering work.
  - Skills: `[]` - Existing repo patterns are enough.
  - Omitted: [`frontend-ui-ux`] - Reason: no visual redesign beyond current terminal style.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 10, 12 | Blocked By: 2, 3

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard_categories.go:287-379` - current SetConfig/Apply lifecycle for category configs.
  - Pattern: `internal/tui/views/wizard_categories.go:660-666` - current Enter/Space toggle handling for boolean category fields.
  - Pattern: `internal/tui/views/wizard_categories.go:879-889` - current `[ ]` / `[✓]` rendering pattern to extend for inclusion controls.
  - Test: `internal/tui/views/wizard_categories_test.go` - existing exhaustive step test suite to extend rather than replace.
  - External: `oh-my-openagent/src/plugin-handlers/category-config-resolver.ts` - missing categories safely fall back to defaults.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizardCategories(LoadsCheckboxStateFromJSONPresence|ApplyWritesOnlySelectedFields|OmitEmptyCategoryObjects)' -v`
  - [x] `go test ./internal/tui/views -run TestWizardCategoriesCheckboxToggleRetainsFormValues -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Present category keys load as checked and sparse save omits unchecked keys
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardCategoriesApplyWritesOnlySelectedFields -v | tee .sisyphus/evidence/task-6-categories.log`
    Expected: test passes and proves category JSON contains only the checked subfields after Apply/save
    Evidence: .sisyphus/evidence/task-6-categories.log

  Scenario: Unchecking a category field removes only that key
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardCategoriesOmitEmptyCategoryObjects -v | tee .sisyphus/evidence/task-6-categories-error.log`
    Expected: test passes and proves unchecked fields disappear without deleting unrelated selected sibling fields
    Evidence: .sisyphus/evidence/task-6-categories-error.log
  ```

  **Commit**: YES | Message: `feat(tui): add category field opt-in controls` | Files: `internal/tui/views/wizard_categories.go`, related tests

- [x] 7. Retrofit the agents step with per-field opt-in checkboxes

  **What to do**: Update `wizard_agents` so every persisted agent field gets its own inclusion checkbox. Explicitly cover the fields currently written unconditionally (`model`, `variant`, `category`, `description`, `color`, `prompt`, `promptAppend`, `reasoningEffort`, `textVerbosity`, `mode`, and any equivalent always-write fields found during implementation). Make `SetConfig` seed checkbox state from actual JSON presence and make `Apply` emit only selected subfields under each agent. Preserve business-level enable/disable semantics separately from inclusion selection.
  **Must NOT do**: Do not overload the agent list checkbox/expand affordance to mean inclusion. Do not keep unconditional writes for empty strings or default dropdown indexes. Do not drop selected sibling fields when one field is unchecked.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: largest TUI form surface after `wizard_other`, with complex per-field interactions.
  - Skills: `[]` - Repo-native patterns suffice.
  - Omitted: [`frontend-ui-ux`] - Reason: keep styling aligned to existing terminal UI.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 10, 12 | Blocked By: 2, 3

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard_agents.go:625-786` - current SetConfig/Apply flow for agent configs.
  - Pattern: `internal/tui/views/wizard_agents.go` - current list/form rendering and checkbox indicators around agent rows.
  - Pattern: `internal/tui/views/wizard_agents.go` - current unconditional string/dropdown writes that collapse presence and must become selection-aware.
  - Test: `internal/tui/views/wizard_agents_test.go` - extend or create agent-step tests in the co-located test file.
  - External: `oh-my-openagent/src/plugin-handlers/agent-config-handler.ts` - missing agent subfields fall back to builtin agent defaults.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizardAgents(LoadsCheckboxStateFromJSONPresence|ApplyWritesOnlySelectedFields|DoesNotMaterializeEmptyStrings)' -v`
  - [x] `go test ./internal/tui/views -run TestWizardAgentsCheckboxToggleRetainsAgentFieldValues -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Agent fields serialize only when explicitly selected
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardAgentsApplyWritesOnlySelectedFields -v | tee .sisyphus/evidence/task-7-agents.log`
    Expected: test passes and proves unchecked agent strings/dropdowns are omitted while selected zero/default-looking values remain present
    Evidence: .sisyphus/evidence/task-7-agents.log

  Scenario: Removing one agent subfield keeps other selected subfields intact
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardAgentsDoesNotMaterializeEmptyStrings -v | tee .sisyphus/evidence/task-7-agents-error.log`
    Expected: test passes and proves the step stops writing empty-string/default-index artifacts when the checkbox is off
    Evidence: .sisyphus/evidence/task-7-agents-error.log
  ```

  **Commit**: YES | Message: `feat(tui): add agent field opt-in controls` | Files: `internal/tui/views/wizard_agents.go`, related tests

- [x] 8. Retrofit the hooks step around the real `disabled_hooks` surface

  **What to do**: Update `wizard_hooks` around the actual persisted surface it exposes today: the single top-level `disabled_hooks` array. Add one explicit inclusion checkbox for the `disabled_hooks` field itself, and keep the existing per-hook enabled/disabled list as the value editor that determines which hook names populate that array. When the inclusion checkbox is off, omit `disabled_hooks` entirely. When it is on, serialize the current disabled-hook list exactly as edited — including an explicit empty array if the user intentionally selects the field but leaves every hook enabled.
  **Must NOT do**: Do not describe or implement nonexistent per-hook nested settings. Do not conflate “hook enabled/disabled inside the list” with “omit the entire `disabled_hooks` field”. Do not force omission when the user explicitly selected `disabled_hooks` but produced an empty list.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: existing hook UI already uses checkbox affordances and now needs a second layer of meaning for the top-level field.
  - Skills: `[]` - Existing patterns are enough.
  - Omitted: [`frontend-ui-ux`] - Reason: no redesign beyond current terminal conventions.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 10, 12 | Blocked By: 2, 3

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard_hooks.go:124-192` - current local state plus `SetConfig` / `Apply`; only `cfg.DisabledHooks` is read/written today.
  - Pattern: `internal/tui/views/wizard_hooks.go:239-309` - current flat 48-hook list rendering with per-hook enabled/disabled markers.
  - API/Type: `internal/config/types.go` - `DisabledHooks []string` is the only hook-related persisted field exposed by this step.
  - Test: `internal/tui/views/wizard_hooks_test.go` - extend/add hook-step tests alongside the source file.
  - External: `oh-my-openagent/src/plugin/event.ts` - omitted hook-related fields remain consumer-default behavior; only explicit `disabled_hooks` should alter runtime behavior.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizardHooks(LoadsDisabledHooksSelectionFromJSONPresence|ApplyWritesDisabledHooksOnlyWhenSelected|SelectedEmptyDisabledHooksSerializesAsEmptyArray)' -v`
  - [x] `go test ./internal/tui/views -run TestWizardHooksSeparatesFieldInclusionFromPerHookToggleState -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: `disabled_hooks` inclusion and per-hook toggle state remain independent
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardHooksSeparatesFieldInclusionFromPerHookToggleState -v | tee .sisyphus/evidence/task-8-hooks.log`
    Expected: test passes and proves list item toggles only affect array contents while the top-level inclusion checkbox controls whether `disabled_hooks` exists at all
    Evidence: .sisyphus/evidence/task-8-hooks.log

  Scenario: Explicitly selected empty `disabled_hooks` stays present as `[]`
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardHooksSelectedEmptyDisabledHooksSerializesAsEmptyArray -v | tee .sisyphus/evidence/task-8-hooks-error.log`
    Expected: test passes and proves a selected `disabled_hooks` field can intentionally serialize as an empty array, while an unselected one is omitted
    Evidence: .sisyphus/evidence/task-8-hooks-error.log
  ```

  **Commit**: YES | Message: `feat(tui): add hook field opt-in controls` | Files: `internal/tui/views/wizard_hooks.go`, related tests

- [x] 9. Retrofit the “other” step with per-field opt-in checkboxes

  **What to do**: Update `wizard_other` so every persisted field in its 21 sections gets an explicit inclusion checkbox. Reuse the current section/field rendering patterns, but for boolean config values render two distinct controls: a left-side inclusion checkbox and a separate value toggle so omission is never confused with `true`/`false`. Make all existing `HasData`/sentinel object builders (`expHasData`, `ccHasData`, `tmuxHasData`, `btHasData`, and similar) depend on selected child paths rather than non-zero values. Ensure fields such as `defaultRunAgent` stop being materialized when not selected.
  **Must NOT do**: Do not collapse inclusion and value toggles into one control. Do not emit empty nested structs because a section was expanded/visited. Do not leave any always-write fields in `Apply`.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: largest TUI surface with dense layout and many boolean/value combinations.
  - Skills: `[]` - Existing terminal patterns are sufficient.
  - Omitted: [`frontend-ui-ux`] - Reason: keep scope to functional terminal controls.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 10, 12 | Blocked By: 2, 3

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard_other.go:692-1091` - current SetConfig hydration for pointer/non-pointer fields.
  - Pattern: `internal/tui/views/wizard_other.go:1092-1577` - current Apply logic with `HasData` sentinels that must become selection-aware.
  - Pattern: `internal/tui/views/wizard_other.go:3177-3194` - current checkbox rendering helper to extend for inclusion controls.
  - Pattern: `internal/tui/layout/layout.go` - responsive width helpers to keep the denser layouts readable.
  - Test: `internal/tui/views/wizard_other_test.go` - extend/add tests alongside the source file.
  - External: `oh-my-openagent/src/config/schema/tmux.ts` - source of tmux default values when the section is omitted or empty.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizardOther(LoadsCheckboxStateFromJSONPresence|ApplyWritesOnlySelectedFields|BooleanFieldsSeparateIncludeAndValueControls)' -v`
  - [x] `go test ./internal/tui/views -run TestWizardOtherOmitsEmptyNestedSections -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Boolean fields distinguish omission from false
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardOtherBooleanFieldsSeparateIncludeAndValueControls -v | tee .sisyphus/evidence/task-9-other.log`
    Expected: test passes and proves a selected `false` remains serialized while an unselected boolean is omitted entirely
    Evidence: .sisyphus/evidence/task-9-other.log

  Scenario: Visiting a section does not create empty nested config objects
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardOtherOmitsEmptyNestedSections -v | tee .sisyphus/evidence/task-9-other-error.log`
    Expected: test passes and proves untouched or fully unchecked sections do not appear in sparse JSON
    Evidence: .sisyphus/evidence/task-9-other-error.log
  ```

  **Commit**: YES | Message: `feat(tui): add other-step field opt-in controls` | Files: `internal/tui/views/wizard_other.go`, related tests

- [x] 10. Integrate sparse preview and save behavior into the review flow

  **What to do**: Rework the review step and wizard save path so the JSON preview is generated exclusively through the sparse serializer and exactly matches what will be written to disk. Ensure blank profiles preview/save as `{}`; existing profiles show only selected keys plus preserved unknown fragments; template-based profiles show only the template-backed selected keys unless the user unchecks them. Keep review errors aligned to the sparse-save validator. Update save success/error messaging only if needed to reflect sparse output accurately.
  **Must NOT do**: Do not show a preview built from raw `config.Config` marshaling. Do not let preview and disk output diverge. Do not special-case template/edit flows with different serialization rules.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: this is the integration point where persistence, validation, and UI preview must converge exactly.
  - Skills: `[]` - No extra skill required.
  - Omitted: [`frontend-ui-ux`] - Reason: functional parity matters more than presentation changes.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 11, 12 | Blocked By: 3, 4, 5, 6, 7, 8, 9

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `internal/tui/views/wizard_review.go:110-238` - current review-state hydration, validation, and viewport preview.
  - Pattern: `internal/tui/views/wizard.go:251-341` - review-step entry and save gating.
  - Pattern: `internal/profile/profile.go:68-81` - final save path to disk.
  - Test: `internal/tui/views/wizard_review_test.go` - extend/add tests alongside the review step.
  - Test: `internal/tui/views/wizard_test.go` - end-to-end wizard save behavior tests.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizardReview(PreviewMatchesSparseSaveOutput|BlankProfileSavesAsEmptyObject|TemplateFlowHonorsSelection)' -v`
  - [x] `go test ./internal/tui/views -run TestWizardSaveUsesSparseValidatorAndSerializer -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Review preview exactly matches saved file contents
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardReviewPreviewMatchesSparseSaveOutput -v | tee .sisyphus/evidence/task-10-review-save.log`
    Expected: test passes and proves the rendered preview JSON bytes match the bytes written by save
    Evidence: .sisyphus/evidence/task-10-review-save.log

  Scenario: Blank sparse profile saves successfully
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run TestWizardReviewBlankProfileSavesAsEmptyObject -v | tee .sisyphus/evidence/task-10-review-save-error.log`
    Expected: test passes and proves a fully unchecked profile saves as `{}` without validation failure
    Evidence: .sisyphus/evidence/task-10-review-save-error.log
  ```

  **Commit**: YES | Message: `feat(wizard): preview and save sparse selected output` | Files: `internal/tui/views/wizard.go`, `internal/tui/views/wizard_review.go`, `internal/profile/profile.go`, related tests

- [x] 11. Add backend regression coverage for sparse persistence and validation

  **What to do**: Expand `internal/profile` and `internal/schema` regression coverage beyond the task-local unit tests to lock down the cross-cutting contract: unknown-fragment preservation, explicit zero/default values when selected, omission of unchecked fields, blank-profile save validity, and strict-vs-save validation behavior. Use table-driven fixtures under `internal/testdata` or package-local fixtures so future contributors can add sparse cases without rebuilding full wizard state.
  **Must NOT do**: Do not rely only on UI tests for backend invariants. Do not create flaky tests tied to map iteration order or non-deterministic JSON formatting.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: broad regression hardening across multiple packages.
  - Skills: `[]` - No extra skill required.
  - Omitted: [`frontend-ui-ux`] - Reason: backend regression task.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: F1-F4 | Blocked By: 1, 4, 5, 10

  **References** (executor has NO interview context - be exhaustive):
  - Test: `internal/profile/profile_test.go` - filesystem-backed profile regression entry point.
  - Test: `internal/schema/validator_test.go` - existing validation assertion patterns.
  - Pattern: `internal/testdata/` - place for cross-package JSON fixtures if package-local tests become noisy.
  - External: `oh-my-openagent/src/plugin-config.ts` - source of defaulting assumptions that sparse fixtures should mirror.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/profile ./internal/schema -run 'Test(Sparse|Profile|Validate)' -v`
  - [x] `go test ./internal/profile ./internal/schema -count=1 -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Backend sparse contract holds across profile and schema packages
    Tool: Bash
    Steps: run `go test ./internal/profile ./internal/schema -count=1 -v | tee .sisyphus/evidence/task-11-backend-regressions.log`
    Expected: package suites pass and cover sparse persistence plus validation semantics without order-dependent failures
    Evidence: .sisyphus/evidence/task-11-backend-regressions.log

  Scenario: Regression fixtures protect explicit-zero and omitted-field behavior
    Tool: Bash
    Steps: run `go test ./internal/profile -run 'Test(ProfileSaveRoundTripsPreservedUnknownFragments|SparseSerializerKeepsExplicitZeroValuesWhenSelected)' -v | tee .sisyphus/evidence/task-11-backend-regressions-error.log`
    Expected: tests pass and prove both omission and explicit zero/default paths remain locked down
    Evidence: .sisyphus/evidence/task-11-backend-regressions-error.log
  ```

  **Commit**: YES | Message: `test(profile): harden sparse persistence regressions` | Files: `internal/profile/*_test.go`, `internal/schema/*_test.go`, optional fixtures

- [x] 12. Add wizard-level regression coverage for create/edit/template/import sparse flows

  **What to do**: Add end-to-end-ish view-layer tests that exercise the wizard across create, edit, template, and import-adjacent flows with the new selection semantics. Cover checked/unchecked transitions, explicit false/0/empty-string values, review preview parity, and preservation of unknown keys after editing an imported or existing sparse profile. Prefer state-driven tests around `Wizard`, `Update`, `Apply`, and saved JSON output rather than brittle string snapshots.
  **Must NOT do**: Do not stop at per-step tests only. Do not rely solely on view substring assertions when state/output assertions are available. Do not require human inspection of terminal rendering.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: broad integration testing across wizard entry points and save behavior.
  - Skills: `[]` - No extra skill required.
  - Omitted: [`frontend-ui-ux`] - Reason: testing focus, not design.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: F1-F4 | Blocked By: 3, 6, 7, 8, 9, 10

  **References** (executor has NO interview context - be exhaustive):
  - Test: `internal/tui/views/wizard_test.go` - root wizard integration test entry point.
  - Test: `internal/tui/views/wizard_categories_test.go` - example of state-first step testing patterns.
  - Test: `internal/tui/views/model_import_test.go` - existing checkbox/list assertion patterns for selection-heavy views.
  - Pattern: `internal/tui/views/wizard.go:96-130` - new/edit/template bootstrap points that must all honor sparse selection.
  - Pattern: `internal/tui/views/import.go` - import-adjacent flow that should accept sparse configs without requiredness failures if it validates on entry.

  **Acceptance Criteria** (agent-executable only):
  - [x] `go test ./internal/tui/views -run 'TestWizard(SparseCreateFlow|SparseEditRoundTrip|SparseTemplateFlow|SparseImportFlow)' -v`
  - [x] `go test ./internal/tui/views -count=1 -v`

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Wizard create/edit/template/import flows all honor sparse selection
    Tool: Bash
    Steps: run `go test ./internal/tui/views -run 'TestWizard(SparseCreateFlow|SparseEditRoundTrip|SparseTemplateFlow|SparseImportFlow)' -v | tee .sisyphus/evidence/task-12-wizard-regressions.log`
    Expected: tests pass and prove all major wizard entry flows preserve selection semantics and sparse save behavior
    Evidence: .sisyphus/evidence/task-12-wizard-regressions.log

  Scenario: Full view test suite stays green after checkbox rollout
    Tool: Bash
    Steps: run `go test ./internal/tui/views -count=1 -v | tee .sisyphus/evidence/task-12-wizard-regressions-error.log`
    Expected: package suite passes without flaky rendering/order assertions
    Evidence: .sisyphus/evidence/task-12-wizard-regressions-error.log
  ```

  **Commit**: YES | Message: `test(tui): cover sparse wizard flows` | Files: `internal/tui/views/*_test.go`, optional fixtures

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [x] F1. Plan Compliance Audit — oracle ✅ APPROVED
  - Tool: `task(subagent_type="oracle")`
  - Steps: review the final diff against this plan file; verify each completed task maps to implemented files/tests; confirm unknown-field preservation, sparse serializer, validation split, and per-step checkbox rollout all exist.
  - Expected: explicit `APPROVE` only if no planned task is skipped and no acceptance criterion is left unaddressed.
  - Evidence: `.sisyphus/evidence/f1-plan-compliance.md` → `.sisyphus/evidence/f1-plan-compliance-reaudit.md`
- [x] F2. Code Quality Review — unspecified-high ✅ APPROVED
  - Tool: `task(category="unspecified-high")`
  - Steps: inspect implementation for brittle state coupling, duplicated checkbox logic that should have been kept minimal, accidental schema divergence in `internal/config/types.go`, and save/preview drift.
  - Expected: explicit `APPROVE` only if the implementation is maintainable, non-duplicative enough for this scope, and free of obvious regressions or dead code.
  - Evidence: `.sisyphus/evidence/f2-code-quality.md` → `.sisyphus/evidence/f2-code-quality-reaudit.md`
- [x] F3. Real Manual QA — unspecified-high ✅ APPROVED
  - Tool: `task(category="unspecified-high")` plus `interactive_bash` if terminal interaction is needed
  - Steps: run `make test`, then exercise at least one create flow and one edit flow in the TUI smoke path (blank profile save, explicit-false selection, `disabled_hooks` selected-empty-array case) using scripted terminal interaction or equivalent deterministic harness.
  - Expected: explicit `APPROVE` only if automated tests pass and the smoke scenarios confirm the saved JSON matches sparse expectations in practice.
  - Evidence: `.sisyphus/evidence/f3-manual-qa.md`
- [x] F4. Scope Fidelity Check — deep ✅ APPROVED
  - Tool: `task(category="deep")`
  - Steps: inspect changed files and commits for out-of-scope work; confirm no profile-name checkboxing, no schema-file drift, no blanket validation removal, and no features beyond sparse config selection.
  - Expected: explicit `APPROVE` only if the diff stays inside the plan's IN scope and respects all Must NOT Have guardrails.
  - Evidence: `.sisyphus/evidence/f4-scope-fidelity.md`

## Commit Strategy
- Commit Task 1, 2, 3, and 4 separately because they redefine the data/validation contract.
- Commit Tasks 5-8 separately per wizard step to keep regressions isolated.
- Commit Task 9 and 10 separately if review/save integration becomes noisy; otherwise combine only if hooks/tests remain clean.
- Commit Task 11 separately as the regression hardening pass.
- Never amend old commits after review feedback; create new commits for fixes.

## Success Criteria
- Editing an existing sparse profile preserves explicit presence for all supported fields and preserves untouched unknown/legacy raw JSON fragments exactly when they are merged back into the saved output.
- Unchecking a field removes that JSON key on the next save without disturbing sibling keys.
- Opening and saving a profile without touching a supported checked field keeps that field present in the output.
- The review screen preview matches the exact sparse JSON that will be written to disk.
- Save-path validation allows omitted fields but still rejects malformed present values.
