# F1 Plan Compliance Audit

Plan reviewed: `.sisyphus/plans/optional-config-selection.md`

## Audit scope
- Reviewed commit history with `GIT_MASTER=1 git log --oneline`
- Verified implementation files and tests for all 12 plan tasks
- Ran `go test ./... -count=1` → PASS
- Verified `internal/config/types.go` was unchanged across the feature range (`GIT_MASTER=1 git diff --name-only 4015a4a..HEAD -- internal/config/types.go` returned no output; `git log` over the same path was empty)
- Verified no checkbox markers were added to profile-name, wizard navigation, review, or template selection views (`wizard_name.go`, `wizard.go`, `wizard_review.go`, `template_select.go` contain no `[ ]`, `[✓]`, or `checkbox` matches)

## Task-by-task compliance

### Task 1 — Preserve unknown/legacy JSON during profile load/save
**FAIL**

**Evidence**
- Unknown-field preservation is implemented: `Profile` now carries `PreservedUnknown` and `FieldPresence` in `internal/profile/profile.go:14-22`; `Load` preserves unknown top-level raw fragments in `internal/profile/profile.go:94-108`; `Save` merges `PreservedUnknown` back in `internal/profile/profile.go:143-155`.
- Tests for unknown round-tripping exist and pass: `TestProfileLoadPreservesUnknownJSON` (`internal/profile/profile_test.go:268-324`), `TestProfileSaveRoundTripsPreservedUnknownFragments` (`internal/profile/profile_test.go:364-409`), and malformed JSON rejection (`internal/profile/profile_test.go:411-427`).
- The planned presence-tracking requirement is not met: `FieldPresence` only records top-level keys (`internal/profile/profile.go:102-105`). There is no recursive supported-field presence map for nested persisted fields such as `agents.*.model` or `categories.*.description`.
- Coverage only asserts top-level presence (`TestProfileLoadCapturesFieldPresence`, `internal/profile/profile_test.go:326-362`).

**Reason for failure**
The task required preserving unknown JSON *and* capturing supported-field presence for the sparse edit flow. The implementation preserves unknown JSON, but the presence map is top-level only, so nested persisted field presence is not available to later tasks.

### Task 2 — Introduce an explicit field-selection model for sparse configs
**FAIL**

**Evidence**
- `FieldSelection` exists with wildcard matching and cloning in `internal/profile/selection.go:166-279`.
- `NewBlankSelection`, `NewSelectionFromPresence`, and `NewSelectionFromTemplate` exist in `internal/profile/selection.go:170-186`.
- The selection seeding logic is incorrect for existing JSON: `NewSelectionFromPresence` marks **every** path under a present top-level key as selected (`internal/profile/selection.go:174-181`). If `agents` is present, it selects `agents.*.model`, `agents.*.variant`, `agents.*.description`, etc., regardless of which keys actually existed in the JSON.
- Tests exist for blank selection, wildcard matching, and toggle retention (`internal/profile/selection_test.go:8-99`), but they do not validate nested-field presence from real JSON.

**Reason for failure**
The plan required selection state to be keyed by actual JSON path presence, not inferred from a top-level object existing. Current logic over-selects nested fields and breaks sparse edit semantics.

### Task 3 — Plumb selection-aware wizard state through every step
**PASS**

**Evidence**
- `Wizard` now carries `config`, `selection`, and `preservedUnknown` in `internal/tui/views/wizard.go:77-98`.
- New/edit/template entry points seed that shared state in `internal/tui/views/wizard.go:101-156`.
- Step contracts were updated in concrete step implementations: `SetConfig`/`Apply` accept `*profile.FieldSelection` in categories (`internal/tui/views/wizard_categories.go:309-403`), agents (`internal/tui/views/wizard_agents.go:627-790`), hooks (`internal/tui/views/wizard_hooks.go:174-188`), other (`internal/tui/views/wizard_other.go:1031-1432`), and review (`internal/tui/views/wizard_review.go:113-119`).
- `Wizard.nextStep()` commits values and selection step-by-step, then passes both into review in `internal/tui/views/wizard.go:262-295`.

**Notes**
Dedicated task-specific tests are lighter than the plan described, but the actual plumbing is present in the implementation.

### Task 4 — Build the sparse JSON serializer used by preview and save
**PASS**

**Evidence**
- `MarshalSparse` and its helpers are implemented in `internal/profile/sparse.go:14-378`.
- Omission logic is driven by selection, not by plain `json.Marshal(config.Config)` (`buildSelectedStruct` / `buildSelectedValue`, `internal/profile/sparse.go:35-171`).
- Explicit zero/default values are preserved through `marshalLeafValue` and `zeroJSONValue` (`internal/profile/sparse.go:173-276`).
- Preserved unknown fragments are deep-merged with known selected values winning (`mergePreservedUnknown` / `mergeKnownValue`, `internal/profile/sparse.go:278-335`).
- Tests exist and pass for omission, explicit zero/default values, preserved unknown merges, known-over-unknown precedence, stable formatting, and empty-parent omission (`internal/profile/sparse_test.go:12-233`).

### Task 5 — Split validation into sparse-save mode and strict mode
**FAIL**

**Evidence**
- The validator split exists: `ValidateForSave` and `ValidateJSONForSave` were added in `internal/schema/validator.go:91-156`.
- Tests exist and pass for empty sparse configs, malformed present values, strict requiredness, and filtering required errors only (`internal/schema/validator_test.go:223-282`).
- Wizard review/save uses the sparse validator in `internal/tui/views/wizard.go:299-315` and `internal/tui/views/wizard_review.go:137-153`.
- Import validation paths were **not** rewired: TUI import still calls strict `ValidateJSON` in `internal/tui/app.go:555-566`, and CLI import still calls strict `ValidateJSON` in `internal/cli/cmd/import.go:42-59`.

**Reason for failure**
The plan explicitly required sparse-save validation on save/review **and any profile-import validation path**. The validator exists, but import flows still reject sparse configs via strict validation.

### Task 6 — Retrofit the categories step with per-field opt-in checkboxes
**FAIL**

**Evidence**
- The categories form renders inclusion checkboxes and separate boolean value toggles (`internal/tui/views/wizard_categories.go:1041-1084`).
- `Apply` writes only selected category fields and omits the whole categories map when no fields are selected (`internal/tui/views/wizard_categories.go:403-506`).
- Tests exist for checkbox rendering, sparse apply, omitted empty categories, and toggle-retains-value behavior (`internal/tui/views/wizard_categories_test.go:251-359`).
- The planned “load checkbox state from actual JSON presence” behavior is not satisfied in edit flows. `NewWizardForEdit` builds selection from top-level `FieldPresence` only (`internal/tui/views/wizard.go:123-137`), and `NewSelectionFromPresence` turns any present `categories` object into all `categories.*` paths selected (`internal/profile/selection.go:174-181`).

**Reason for failure**
The step UI exists, but existing category profiles cannot seed per-field inclusion from real JSON key presence. Any existing `categories` object effectively preselects every managed category field.

### Task 7 — Retrofit the agents step with per-field opt-in checkboxes
**FAIL**

**Evidence**
- Agent field inclusion logic exists in `internal/tui/views/wizard_agents.go:790-1004`, and per-field checkbox rendering exists in `internal/tui/views/wizard_agents.go:1999-2037`.
- Tests exist for checkbox rendering, sparse apply, omission of unchecked fields, and checkbox-toggle value retention (`internal/tui/views/wizard_agents_test.go:423-569`).
- Existing JSON presence is still seeded from the top-level `agents` object only (`internal/profile/selection.go:174-181`, `internal/tui/views/wizard.go:123-137`).
- Because many agent fields write direct string/dropdown values when selected (`internal/tui/views/wizard_agents.go:863-887`, `920-930`), over-selection in edit mode can materialize empty strings/default dropdown values that were never present in the source JSON.

**Reason for failure**
The UI checkbox rollout is present, but the required actual-key presence behavior for existing agent JSON is not. That breaks the “must not materialize defaults on edit/save” contract.

### Task 8 — Retrofit the hooks step around `disabled_hooks`
**PASS**

**Evidence**
- The hooks step implements a dedicated inclusion checkbox for `disabled_hooks` in `internal/tui/views/wizard_hooks.go:205-337`.
- Per-hook enabled/disabled state remains separate from inclusion state (`internal/tui/views/wizard_hooks.go:188-203`, `236-244`).
- Selecting `disabled_hooks` with no disabled entries serializes an explicit empty array (`internal/tui/views/wizard_hooks.go:190-199`).
- Tests cover selection seeding, apply-only-when-selected, selected-empty-array serialization, and independence between top-level inclusion and per-hook toggles (`internal/tui/views/wizard_hooks_test.go:166-332`).

### Task 9 — Retrofit the “other” step with per-field opt-in checkboxes
**FAIL**

**Evidence**
- The other step clearly separates inclusion checkboxes from boolean value toggles (`internal/tui/views/wizard_other.go:3556-3606`).
- `Apply` uses `fieldSelected`/`selectedWithPrefix` and section `HasData` helpers to omit unselected fields and sections (`internal/tui/views/wizard_other.go:323-423`, `1432-1908`).
- Tests cover checkbox rendering, separate boolean inclusion/value controls, sparse apply, and omission of untouched sections (`internal/tui/views/wizard_other_test.go:200-289`).
- Edit-mode selection seeding is still top-level only. For example, if `experimental` exists with one child key, `NewSelectionFromPresence` selects every `experimental.*` path (`internal/profile/selection.go:174-181`), and section builders such as `expHasData`, `btHasData`, and `tmuxHasData` then treat the whole section as selected (`internal/tui/views/wizard_other.go:349-416`).

**Reason for failure**
The rendering and sparse-apply mechanics exist, but the required per-child selection state from existing JSON presence is missing, so edit/save can still re-materialize unrelated nested defaults.

### Task 10 — Integrate sparse preview and save behavior into the review flow
**FAIL**

**Evidence**
- Review preview uses `profile.MarshalSparse` in `internal/tui/views/wizard_review.go:121-136`.
- Save uses the same sparse serializer and writes the preview bytes to disk in `internal/tui/views/wizard.go:334-358`.
- Tests prove preview/save parity and blank-profile `{}` behavior (`internal/tui/views/wizard_review_test.go:393-481`).
- However, review/save correctness depends on the selection model. Because edit/template flows still seed nested selections from top-level presence only (`internal/tui/views/wizard.go:123-156`, `internal/profile/selection.go:174-181`), review/save does **not** reliably show only the keys actually present/selected for nested agent/category/other fields.

**Reason for failure**
The serializer integration is correct, but the review flow still inherits the broken nested selection state from earlier tasks, so the end-to-end plan requirement is not fully met.

### Task 11 — Add backend regression coverage for sparse persistence and validation
**FAIL**

**Evidence**
- Regression-style backend tests exist: `TestRegressionSparsePersistenceContract` in `internal/profile/profile_test.go:429-675` and `TestRegressionSparseValidationContract` in `internal/schema/validator_test.go:315-431`.
- Full package tests passed as part of `go test ./... -count=1`.
- The regression coverage does **not** lock down the nested presence contract that the plan called out as cross-cutting. There is no test proving that existing nested keys seed only their actual JSON paths, and that omission survives edit/save for nested categories/agents/other sections.

**Reason for failure**
Regression tests were added, but they missed the main contract break introduced by top-level-only presence tracking. The backend hardening is therefore incomplete.

### Task 12 — Add wizard-level regression coverage for create/edit/template/import sparse flows
**FAIL**

**Evidence**
- `internal/tui/views/wizard_sparse_flow_test.go` adds wizard-level create/edit/template/import-adjacent tests (`13-260`) and helper coverage (`263-343`).
- The full views test suite passed inside `go test ./... -count=1`.
- Coverage is still incomplete relative to the plan: the tests mostly exercise top-level/simple fields such as `hashline_edit`, `disabled_mcps`, `default_run_agent`, and `background_task.provider_concurrency`, but do not exercise nested category/agent/other presence seeding from loaded JSON.
- The import-adjacent test does not exercise the real import validation path that still uses strict validation (`internal/tui/app.go:555-566`, `internal/cli/cmd/import.go:42-59`).

**Reason for failure**
Wizard regression coverage was expanded, but it did not catch the nested selection bug or the still-strict import validation path, so the task is not complete against the plan’s intended coverage.

## Guardrail checks

### `internal/config/types.go` schema divergence guardrail
**PASS**
- `GIT_MASTER=1 git diff --name-only 4015a4a..HEAD -- internal/config/types.go` produced no output.
- `GIT_MASTER=1 git log --oneline 4015a4a..HEAD -- internal/config/types.go` produced no output.

### No checkbox added to profile name or navigation/template controls
**PASS**
- No checkbox markers were found in `internal/tui/views/wizard_name.go`, `internal/tui/views/wizard.go`, `internal/tui/views/wizard_review.go`, or `internal/tui/views/template_select.go`.

### Full test suite
**PASS**
- `go test ./... -count=1` completed successfully.

## Overall verdict
**REJECT**

### Specific issues requiring follow-up
1. Presence tracking is only top-level (`internal/profile/profile.go:102-105`), so edit/template flows cannot seed checkbox state from actual nested JSON path presence.
2. Import validation paths still use strict validation (`internal/tui/app.go:555-566`, `internal/cli/cmd/import.go:42-59`) instead of sparse-save validation.
3. Regression coverage did not catch the nested presence bug, so several task-level tests pass while the core plan contract is still broken.

**VERDICT: REJECT**
