# ✅ CONCLUÍDO - Sparse Field Inclusion: Hooks & Other Settings Consistency

> **Status**: ✅ COMPLETO  
> **Data de Conclusão**: 2026-04-08  
> **Commits**: 2 commits (0231b21, ffb9a49)  
> **Push**: Enviado para origin/main

## TL;DR

> **Quick Summary**: Fix a UI toggle bug where disabled-list sections in wizard_other cannot toggle field inclusion, and add comprehensive tests covering the full SetConfig→Apply→MarshalSparse round-trip for both wizard_hooks and wizard_other. The core serialization engine (FieldSelection, MarshalSparse, sparse.go) is already correct — only the UI toggle layer and test coverage need work.
> 
> **Deliverables**:
> - Fix `toggleSubItem()` to handle Row 0 (inclusion toggle) for disabled-list sections
> - Add disabled-list section inclusion toggling to `toggleSection()` (when not in sub-section)
> - Comprehensive round-trip tests for wizard_other (edit flow, template flow, new profile flow)
> - Empty-array serialization tests (selected-but-empty vs unselected)
> - Inclusion/value separation tests for all field types
> - Existing test regression guard
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 2 waves
> **Critical Path**: Task 1 (test helpers) → Tasks 2-3 (tests + fix in parallel) → Task 4 (integration tests)

---

## Context

### Original Request
Ensure consistent sparse field inclusion/omission across wizard_hooks and wizard_other. Reuse `FieldSelection` as source of truth. No schema changes. Add tests for loading, inclusion/value separation, JSON omission, persistence, and edit/template flows.

### Interview Summary
**Key Discussions**:
- wizard_hooks already fully implements the disabled_hooks inclusion/omission feature correctly
- wizard_other already has per-field inclusion for most sections (experimental, claude_code, tmux, etc.)
- "Disabled list" sections (agents/skills/commands/mcps/tools) use whole-section inclusion — user confirmed to keep this approach
- No label/nomenclature changes needed
- Tests should go in existing test files

**Research Findings**:
- **BUG FOUND** (Metis): For disabled-list sections in wizard_other, the inclusion checkbox rendered by `renderInclude()` is NOT toggleable. `toggleSubItem()` at Row 0 returns `""` from `subSectionFieldPath()` for disabled sections, so the field selection toggle never fires. `toggleSection()` also has no handler for disabled sections.
- wizard_hooks works correctly because it uses a separate `includeFocused` bool pattern
- The core serialization engine (MarshalSparse, FieldSelection, sparse.go) is correct
- The Apply() methods for both steps correctly gate on `fieldSelected()`

### Metis Review
**Identified Gaps** (addressed):
- Disabled-section inclusion toggle bug: Will fix in `toggleSubItem()` and `toggleSection()`
- Test coverage for wizard_other is minimal (7 tests, 289 lines) vs implementation (3766 lines): Will add comprehensive data-path tests
- No round-trip tests (SetConfig→Apply→MarshalSparse→verify JSON): Will add
- No edit/template flow tests for wizard_other: Will add

---

## Work Objectives

### Core Objective
Fix the disabled-section toggle bug in wizard_other and add comprehensive test coverage for the sparse field inclusion/omission pipeline across both wizard_hooks and wizard_other.

### Concrete Deliverables
- Fixed `toggleSubItem()` and `toggleSection()` in `wizard_other.go`
- ~20+ new tests in `wizard_other_test.go`
- All existing tests continue passing

### Definition of Done
- [x] `make test` passes with zero failures
- [x] `make lint` passes with zero new warnings (5 issues pré-existentes)
- [x] Disabled-list sections can toggle inclusion via UI (Space on Row 0)
- [x] Round-trip tests verify JSON output for each section type
- [x] Edit/template flow tests pass

### Must Have
- Fix for `toggleSubItem()` Row 0 bug for disabled-list sections
- Fix for `toggleSection()` for disabled-list sections when not expanded
- Round-trip tests: SetConfig → Apply → MarshalSparse → verify JSON
- Edit flow tests: Load with FieldPresence → SetConfig → modify → Apply → verify
- Empty-array tests: selected-but-empty → `[]`, unselected → absent
- Inclusion/value separation tests

### Must NOT Have (Guardrails)
- No changes to `config/types.go` (no schema changes)
- No changes to `selection.go` (no new field paths)
- No changes to `sparse.go` (serialization engine is correct)
- No changes to `wizard_hooks.go` (it's the reference implementation)
- No changes to `profile.go` (Load/Save is correct)
- No refactoring of `wizard_other.go` structure
- No rendering/UI layout changes beyond the toggle fix
- No per-item inclusion for disabled lists (whole-section only, per user decision)
- No label/nomenclature changes (per user decision)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Go test framework, testify assertions)
- **Automated tests**: YES (Tests-after — the feature is mostly a bug fix + test augmentation)
- **Framework**: Go standard `testing` + `github.com/stretchr/testify`
- **Pattern**: Co-located `*_test.go` files

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Unit tests**: Use Bash (`go test ./internal/... -v -run TestName`) — run tests, assert pass/fail
- **Integration tests**: Use Bash (`go test ./internal/... -v`) — full suite, assert zero failures
- **Lint**: Use Bash (`make lint`) — assert zero new warnings

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - test helpers + fix):
├── Task 1: Test helpers in wizard_other_test.go [quick]
└── Task 2: Fix toggleSubItem/toggleSection bug [quick]

Wave 2 (Tests - max parallel):
├── Task 3: Round-trip tests for disabled-list sections [unspecified-high]
├── Task 4: Round-trip tests for non-disabled sections [unspecified-high]
├── Task 5: Empty-array & omission tests [unspecified-high]
├── Task 6: Inclusion/value separation tests [unspecified-high]
└── Task 7: Edit/template flow integration tests [deep]

Wave FINAL (After ALL tasks — verification):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Full test suite QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
```

### Dependency Matrix

| Task | Depends On | Blocks |
|------|-----------|--------|
| 1 | - | 3, 4, 5, 6, 7 |
| 2 | - | F1-F4 |
| 3 | 1 | F1-F4 |
| 4 | 1 | F1-F4 |
| 5 | 1 | F1-F4 |
| 6 | 1 | F1-F4 |
| 7 | 1 | F1-F4 |

### Agent Dispatch Summary

- **Wave 1**: **2** — T1 → `quick`, T2 → `quick`
- **Wave 2**: **5** — T3 → `unspecified-high`, T4 → `unspecified-high`, T5 → `unspecified-high`, T6 → `unspecified-high`, T7 → `deep`
- **FINAL**: **4** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Test Helpers for wizard_other Sparse Field Tests

  **What to do**:
  - Add test helper functions to `wizard_other_test.go`:
    - `setupWizardOtherWithSelection(t *testing.T, paths ...string)` — creates `WizardOther` with a `FieldSelection` that has the given paths selected
    - `applyAndMarshal(t *testing.T, w *WizardOther, selection *profile.FieldSelection) map[string]interface{}` — calls `w.Apply(&cfg, selection)`, then `profile.MarshalSparse(&cfg, selection, nil)`, unmarshals result to `map[string]interface{}` for assertions
    - `assertJSONContains(t *testing.T, jsonMap map[string]interface{}, key string)` — asserts key exists in JSON output
    - `assertJSONOmits(t *testing.T, jsonMap map[string]interface{}, key string)` — asserts key does NOT exist in JSON output
    - `assertJSONEquals(t *testing.T, jsonMap map[string]interface{}, key string, expected interface{})` — asserts key exists with specific value
  - Follow the existing `setupTestEnv(t)` pattern from `wizard_hooks_test.go`
  - Import `profile.MarshalSparse` for JSON output verification

  **Must NOT do**:
  - Do not modify production code
  - Do not add helpers that are too specific to one test case

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Pure test helper additions, no production code changes
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `git-master`: No git operations needed

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 2)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 3, 4, 5, 6, 7
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks_test.go:166-187` — `setupTestEnv` helper pattern and `TestWizardHooksLoadsDisabledHooksSelectionFromJSONPresence` for selection-based test setup
  - `internal/tui/views/wizard_other_test.go:200-217` — Existing `TestWizardOtherLoadsCheckboxStateFromJSONPresence` for how tests currently create WizardOther with selection

  **API/Type References**:
  - `internal/profile/sparse.go:14-33` — `MarshalSparse(cfg, selection, preservedUnknown)` signature and return type
  - `internal/profile/selection.go:170-186` — `NewBlankSelection()`, `NewSelectionFromPresence()` constructors
  - `internal/config/types.go:6-40` — `Config` struct for creating test instances

  **WHY Each Reference Matters**:
  - `wizard_hooks_test.go:166` — Copy the `setupTestEnv` pattern exactly; it handles `config.SetBaseDir(t.TempDir())`
  - `sparse.go:14` — Need `MarshalSparse` to verify JSON output in tests; this is the function that actually decides what goes in the JSON
  - `selection.go:170` — `NewBlankSelection()` creates empty selection; `SetSelected()` to mark specific paths

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Test helpers compile and work
    Tool: Bash
    Preconditions: Go module intact
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOther -count=1 -v`
      2. Verify no compilation errors
      3. Verify existing tests still pass
    Expected Result: All existing wizard_other tests pass, helpers are importable
    Failure Indicators: Compilation error, test failure
    Evidence: .sisyphus/evidence/task-1-helpers-compile.txt

  Scenario: applyAndMarshal produces valid JSON
    Tool: Bash
    Preconditions: Test helpers added
    Steps:
      1. Write a small test that uses `applyAndMarshal` with a selected field
      2. Verify the returned map contains the expected key
      3. Run `go test ./internal/tui/views/ -run TestHelper -count=1 -v`
    Expected Result: Helper correctly round-trips through Apply→MarshalSparse
    Failure Indicators: Nil map, missing keys, JSON parse error
    Evidence: .sisyphus/evidence/task-1-helpers-roundtrip.txt
  ```

  **Commit**: YES
  - Message: `fix(tui): add test helpers for wizard_other sparse field tests`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -count=1`

- [x] 2. Fix Toggle Bug for Disabled-List Section Inclusion

  **What to do**:
  - In `wizard_other.go`, fix `toggleSubItem()` (line 3362):
    - At the TOP of the function, BEFORE the `subSectionFieldPath()` call (line 3363), add a check:
    - If `subCursor == 0` AND the current section is a disabled-list section (`sectionDisabledMcps`, `sectionDisabledAgents`, `sectionDisabledSkills`, `sectionDisabledCommands`, `sectionDisabledTools`), call `toggleFieldSelection()` with the section's top-level path from `topLevelFieldPath()` and return
  - In `wizard_other.go`, fix `toggleSection()` (line 3333):
    - Add cases for the disabled-list sections: `sectionDisabledMcps`, `sectionDisabledAgents`, `sectionDisabledSkills`, `sectionDisabledCommands`, `sectionDisabledTools`
    - Each case should call `w.toggleFieldSelection(w.topLevelFieldPath(section))`
    - These sections don't have `simpleValueFocused` (they use sub-section expansion), so no value-toggle branch needed
  - Add tests for the fix:
    - `TestWizardOtherDisabledAgentsInclusionToggle` — verify Space on Row 0 toggles the inclusion checkbox
    - `TestWizardOtherDisabledSkillsInclusionToggle` — same for skills
    - `TestWizardOtherDisabledCommandsInclusionToggle` — same for commands
    - `TestWizardOtherDisabledMcpsInclusionToggle` — same for mcps
    - `TestWizardOtherDisabledToolsInclusionToggle` — same for tools
    - Each test: create WizardOther with blank selection, expand section, set subCursor=0, send Space key, verify `selection.IsSelected(path)` toggled

  **Must NOT do**:
  - Do not change the Apply() logic for disabled sections (it's already correct)
  - Do not change rendering code
  - Do not add per-item inclusion paths
  - Do not modify `subSectionFieldPath()` or `topLevelFieldPath()` — use them as-is

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small targeted fix in two methods + simple test additions
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `git-master`: No git operations needed

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 1)
  - **Parallel Group**: Wave 1
  - **Blocks**: F1-F4
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_other.go:3333-3360` — `toggleSection()` — only handles 4 sections currently; add disabled-list cases following same pattern
  - `internal/tui/views/wizard_other.go:3362-3394` — `toggleSubItem()` — the function to fix; Row 0 for disabled sections currently falls through
  - `internal/tui/views/wizard_hooks.go:236-243` — Reference: how wizard_hooks handles the inclusion toggle via `includeFocused` + `selection.Toggle()`

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:425-454` — `topLevelFieldPath()` — already returns correct paths for disabled sections
  - `internal/tui/views/wizard_other.go:323-328` — `fieldSelected(path)` helper
  - `internal/tui/views/wizard_other.go:330-335` — `toggleFieldSelection(path)` helper

  **Test References**:
  - `internal/tui/views/wizard_other_test.go:219-246` — `TestWizardOtherBooleanFieldSeparatesInclusionAndValue` — pattern for testing Space key toggle behavior with selection

  **WHY Each Reference Matters**:
  - `toggleSection():3333` — This is WHERE to add the disabled-section cases. Follow the existing `sectionAutoUpdate` pattern.
  - `toggleSubItem():3362` — This is WHERE to add the Row 0 early return. The fix goes BEFORE line 3363.
  - `topLevelFieldPath():425` — Already returns `"disabled_agents"` etc. for disabled sections — use this to get the path.
  - `wizard_hooks.go:236` — Reference pattern: `selection.Toggle(disabledHooksFieldPath)` is the correct way to toggle inclusion.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Row 0 toggle in expanded disabled_agents section
    Tool: Bash
    Preconditions: WizardOther created, section expanded, subCursor=0
    Steps:
      1. Create WizardOther with blank selection
      2. Set currentSection=sectionDisabledAgents, inSubSection=true, subCursor=0
      3. Send Space key via Update()
      4. Verify selection.IsSelected("disabled_agents") == true
      5. Send Space again
      6. Verify selection.IsSelected("disabled_agents") == false
    Expected Result: Inclusion toggles on/off with each Space press
    Failure Indicators: Selection state unchanged after Space
    Evidence: .sisyphus/evidence/task-2-disabled-agents-toggle.txt

  Scenario: Row 1+ toggle individual agents (not affected by fix)
    Tool: Bash
    Preconditions: WizardOther created, section expanded, subCursor=1
    Steps:
      1. Create WizardOther, expand disabled_agents section
      2. Set subCursor=1 (first agent "sisyphus")
      3. Send Space key
      4. Verify disabledAgents["sisyphus"] toggled from false to true
      5. Verify selection state unchanged
    Expected Result: Individual agent value toggles, inclusion unchanged
    Failure Indicators: Inclusion also toggled, or agent value unchanged
    Evidence: .sisyphus/evidence/task-2-agent-value-toggle.txt

  Scenario: Section-level toggle (not expanded)
    Tool: Bash
    Preconditions: WizardOther created, section NOT expanded, cursor on disabled_agents
    Steps:
      1. Create WizardOther with blank selection
      2. Set currentSection=sectionDisabledAgents, inSubSection=false
      3. Call toggleSection() or send Space key
      4. Verify selection.IsSelected("disabled_agents") == true
    Expected Result: Section-level inclusion toggle works
    Failure Indicators: Selection state unchanged
    Evidence: .sisyphus/evidence/task-2-section-level-toggle.txt

  Scenario: Existing tests still pass after fix
    Tool: Bash
    Preconditions: Fix applied
    Steps:
      1. Run `make test`
    Expected Result: All tests pass, zero failures
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-2-regression.txt
  ```

  **Commit**: YES
  - Message: `fix(tui): enable inclusion toggle for disabled-list sections in wizard_other`
  - Files: `internal/tui/views/wizard_other.go`, `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -count=1`

- [x] 3. Round-Trip Tests for Disabled-List Sections

  **What to do**:
  - Add round-trip tests to `wizard_other_test.go` for each disabled-list section:
    - `TestWizardOtherRoundTrip_DisabledAgents` — select `disabled_agents`, disable some agents, Apply→MarshalSparse→verify JSON has `disabled_agents: ["agent1", "agent2"]`
    - `TestWizardOtherRoundTrip_DisabledAgentsEmpty` — select `disabled_agents`, disable nothing, verify JSON has `disabled_agents: []`
    - `TestWizardOtherRoundTrip_DisabledAgentsOmitted` — don't select `disabled_agents`, verify JSON has NO `disabled_agents` key
    - Same 3-test pattern for: `DisabledSkills`, `DisabledCommands`, `DisabledMcps`, `DisabledTools`
  - Each test uses the `applyAndMarshal` helper from Task 1
  - For `DisabledMcps` and `DisabledTools`: set text input value before Apply
  - For `DisabledAgents/Skills/Commands`: set individual toggle values in the map

  **Must NOT do**:
  - Do not modify production code
  - Do not test UI rendering — focus on data path only

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Multiple test cases requiring careful assertion logic, but all follow same pattern
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 6, 7)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (test helpers)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks_test.go:219-236` — `TestWizardHooksSelectedEmptyDisabledHooksSerializesAsEmptyArray` — exact pattern for empty-array test
  - `internal/tui/views/wizard_hooks_test.go:189-217` — `TestWizardHooksApplyWritesDisabledHooksOnlyWhenSelected` — pattern for selected-vs-omitted test

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:1432-1488` — Apply() for disabled sections (lines 1435-1488) — shows exactly which fields are set and how
  - `internal/tui/views/wizard_other.go:303-311` — `emptySliceIfSelected()` — explains nil vs empty-slice logic

  **Test References**:
  - `internal/tui/views/wizard_other_test.go:248-274` — `TestWizardOtherApplyWritesOnlySelectedFields` — existing pattern for testing Apply with selection

  **WHY Each Reference Matters**:
  - `wizard_hooks_test.go:219` — Copy this exact test structure for empty-array tests. It creates WizardHooks, sets selection, calls Apply, checks `cfg.DisabledHooks == nil` vs `len == 0`.
  - `wizard_other.go:1435-1444` — Shows the Apply code for disabled_agents: `cfg.DisabledAgents = emptySliceIfSelected(true, agents)` — the `true` is hardcoded because the outer `if` already checked `fieldSelected`. This means selected-but-empty → `[]`, unselected → `nil`.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All disabled-list round-trip tests pass
    Tool: Bash
    Preconditions: Test helpers available (Task 1)
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherRoundTrip_Disabled -v -count=1`
      2. Verify all tests pass
    Expected Result: 15 tests pass (3 per section × 5 sections)
    Failure Indicators: Any test failure, missing test
    Evidence: .sisyphus/evidence/task-3-disabled-roundtrip.txt

  Scenario: Round-trip matches expected JSON output
    Tool: Bash
    Preconditions: Tests written
    Steps:
      1. Run the round-trip tests
      2. Verify selected test outputs `"disabled_agents": ["sisyphus"]` (not null, not missing)
      3. Verify omitted test outputs no `disabled_agents` key
      4. Verify empty test outputs `"disabled_agents": []` (not null, not missing)
    Expected Result: Each variant produces correct JSON representation
    Failure Indicators: null instead of [], field present when it should be omitted
    Evidence: .sisyphus/evidence/task-3-json-variants.txt
  ```

  **Commit**: YES
  - Message: `test(tui): add round-trip tests for disabled-list sections`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -run TestWizardOtherRoundTrip_Disabled -count=1`

- [x] 4. Round-Trip Tests for Non-Disabled Sections

  **What to do**:
  - Add round-trip tests for each non-disabled section in `wizard_other_test.go`:
    - **Experimental**: Select individual fields (e.g., `experimental.aggressive_truncation`, `experimental.auto_resume`), set values, Apply→MarshalSparse→verify JSON structure. Test with some fields selected and others omitted. Include nested `dynamic_context_pruning` fields.
    - **ClaudeCode**: Select individual `claude_code.*` fields, set boolean values, verify JSON.
    - **SisyphusAgent**: Select individual `sisyphus_agent.*` fields, verify.
    - **RalphLoop**: Select `ralph_loop.enabled` + other fields, verify.
    - **BackgroundTask**: Select individual fields, verify.
    - **Notification**: Select `notification.force_enable`, verify.
    - **GitMaster**: Select individual `git_master.*` fields, verify.
    - **Tmux**: Select `tmux.enabled`, `tmux.layout`, verify.
    - **Simple booleans**: `auto_update`, `new_task_system_enabled`, `hashline_edit`, `model_fallback`.
    - **String fields**: `default_run_agent`.
    - **JSON RawMessage**: `runtime_fallback`, `skills`.
  - For each: test 3 states — selected with value, selected without value (if applicable), unselected.
  - Group tests by section to keep file organized.

  **Must NOT do**:
  - Do not modify production code
  - Do not test every single sub-field of every section — pick representative fields
  - Do not test rendering

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Many test cases but all follow the same pattern; high volume work
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 5, 6, 7)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (test helpers)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_other_test.go:248-274` — `TestWizardOtherApplyWritesOnlySelectedFields` — existing test that covers experimental + tmux. Extend this pattern.
  - `internal/profile/sparse_test.go` — Tests for MarshalSparse with selected/unselected fields

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:1432-1908` — Full Apply() method — reference for which fields are set for each section and how
  - `internal/config/types.go` — Config struct with all field names and JSON tags

  **Test References**:
  - `internal/profile/sparse_test.go` — `TestSparseSerializerOmitsUncheckedFields` — pattern for verifying field omission

  **WHY Each Reference Matters**:
  - `wizard_other.go:1432` — The entire Apply() method. Each section has its own block. The test must verify that each block correctly respects `fieldSelected()`.
  - `sparse_test.go` — Shows how to verify MarshalSparse output using JSON unmarshal and map assertions.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Non-disabled section round-trip tests pass
    Tool: Bash
    Preconditions: Test helpers available
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherRoundTrip -v -count=1`
      2. Verify all round-trip tests pass (both disabled and non-disabled)
    Expected Result: All tests pass, at least 1 test per section type
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-4-non-disabled-roundtrip.txt

  Scenario: Nested struct fields handled correctly
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherRoundTrip_Experimental -v`
      2. Verify experimental.dynamic_context_pruning.enabled appears in JSON when selected
      3. Verify experimental block omitted entirely when no experimental.* fields selected
    Expected Result: Nested structs correctly included/omitted based on selection
    Failure Indicators: Empty experimental: {} in JSON, or missing DCP fields
    Evidence: .sisyphus/evidence/task-4-nested-struct.txt
  ```

  **Commit**: YES
  - Message: `test(tui): add round-trip tests for non-disabled sections`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -run TestWizardOtherRoundTrip -count=1`

- [x] 5. Empty-Array and Field Omission Tests

  **What to do**:
  - Add tests for the three distinct JSON states of disabled-list fields:
    - **Unselected** (not in FieldSelection) → field completely absent from JSON
    - **Selected but empty** (in FieldSelection, no items disabled) → field present as `[]`
    - **Selected with values** (in FieldSelection, items disabled) → field present as `["item1", "item2"]`
  - Test each disabled-list section (agents, skills, commands, mcps, tools) for all three states
  - Test name pattern: `TestWizardOtherOmission_Disabled{Section}` and `TestWizardOtherEmptyArray_Disabled{Section}`
  - Additionally test that these work through the MarshalSparse pipeline:
    - Create Config, set selection, call MarshalSparse, unmarshal result to map, verify exact state
  - Also test boolean fields for similar distinction:
    - Unselected `auto_update` → absent from JSON
    - Selected `auto_update` with value `false` → `"auto_update": false` in JSON
    - Selected `auto_update` with value `true` → `"auto_update": true` in JSON

  **Must NOT do**:
  - Do not modify production code
  - Do not test rendering
  - Do not confuse "empty array" `[]` with `null` — both must be tested

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Many test cases requiring careful JSON output verification
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 6, 7)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (test helpers)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks_test.go:219-236` — `TestWizardHooksSelectedEmptyDisabledHooksSerializesAsEmptyArray` — EXACT pattern for empty-array test. This test creates hooks, selects the field, applies with nothing disabled, verifies `cfg.DisabledHooks` is `[]string{}` not nil.
  - `internal/profile/sparse_test.go` — `TestSparseSerializerKeepsExplicitZeroValuesWhenSelected` — pattern for verifying selected-but-false booleans persist

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:303-311` — `emptySliceIfSelected()` — the helper that produces `[]string{}` vs nil
  - `internal/profile/sparse.go:78-171` — `buildSelectedValue()` — how MarshalSparse decides to include leaf values

  **WHY Each Reference Matters**:
  - `wizard_hooks_test.go:219` — This is the EXACT test to replicate for each disabled section. Copy structure, change field names.
  - `emptySliceIfSelected():303` — The critical function: `selected=false → nil`, `selected=true && nil values → []string{}`. Tests must verify both paths.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Empty-array tests pass for all disabled sections
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run "TestWizardOtherEmptyArray|TestWizardOtherOmission" -v -count=1`
    Expected Result: All tests pass, 3 states tested per section (5 sections × 3 = 15 tests)
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-5-empty-omission.txt

  Scenario: Boolean false vs omitted distinction
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherOmission_AutoUpdate -v`
      2. Verify unselected → absent, selected false → "auto_update": false
    Expected Result: Three distinct JSON outputs for three states
    Failure Indicators: false treated as absent, or absent treated as null
    Evidence: .sisyphus/evidence/task-5-boolean-distinction.txt
  ```

  **Commit**: YES
  - Message: `test(tui): add empty-array and omission tests`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -run "TestWizardOtherEmptyArray|TestWizardOtherOmission" -count=1`

- [x] 6. Inclusion/Value Separation Tests

  **What to do**:
  - Add tests verifying that toggling inclusion does NOT change the field's value, and vice versa:
    - `TestWizardOtherInclusionSeparateFromValue_DisabledAgents` — toggle inclusion off, verify individual agent disabled states preserved; toggle back on, verify states still there
    - `TestWizardOtherInclusionSeparateFromValue_BoolField` — toggle inclusion of `experimental.aggressive_truncation` off, verify `expAggressiveTrunc` value unchanged; toggle value, verify inclusion unchanged
    - `TestWizardOtherInclusionSeparateFromValue_SliceField` — toggle inclusion of `disabled_mcps` off, verify text input value preserved
    - `TestWizardOtherInclusionSeparateFromValue_StringField` — toggle inclusion of `default_run_agent` off, verify text input value preserved
  - Pattern: For each test, set up a value, toggle the OTHER property (inclusion or value), verify the first is unchanged

  **Must NOT do**:
  - Do not modify production code
  - Do not test UI rendering

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Careful state-verification tests requiring precise assertions
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 7)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (test helpers)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_other_test.go:219-246` — `TestWizardOtherBooleanFieldSeparatesInclusionAndValue` — EXACT pattern to follow. Tests that Space on inclusion toggle doesn't change boolean value, and Space on value toggle doesn't change inclusion.
  - `internal/tui/views/wizard_hooks_test.go:313-332` — `TestWizardHooksSeparatesFieldInclusionFromPerHookToggleState` — same pattern for hooks

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:323-335` — `fieldSelected()` and `toggleFieldSelection()` — the helpers being tested
  - `internal/tui/views/wizard_other.go:3362-3394` — `toggleSubItem()` — where inclusion vs value branching happens

  **WHY Each Reference Matters**:
  - `wizard_other_test.go:219` — This test ALREADY exists for boolean fields. Replicate its structure for disabled lists, string fields, and slice fields. Key pattern: set up initial state, toggle one property, assert the other is unchanged.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Inclusion/value separation tests pass
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherInclusionSeparate -v -count=1`
    Expected Result: All separation tests pass (4+ tests)
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-6-separation.txt

  Scenario: Toggling inclusion preserves value state
    Tool: Bash
    Steps:
      1. In test: set agent "sisyphus" to disabled=true, select disabled_agents
      2. Toggle inclusion off via toggleFieldSelection
      3. Verify disabledAgents["sisyphus"] still == true
      4. Toggle inclusion back on
      5. Apply → verify "sisyphus" appears in cfg.DisabledAgents
    Expected Result: Value state survives inclusion toggling
    Failure Indicators: Value reset when inclusion toggled
    Evidence: .sisyphus/evidence/task-6-value-preservation.txt
  ```

  **Commit**: YES
  - Message: `test(tui): add inclusion/value separation tests`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `go test ./internal/tui/views/ -run TestWizardOtherInclusionSeparate -count=1`

- [x] 7. Edit/Template Flow Integration Tests

  **What to do**:
  - Add integration tests that exercise the full edit and template flows:
    - `TestWizardOtherEditFlow_LoadsFieldPresence` — simulate loading a profile that has `disabled_agents: ["sisyphus"]` and `experimental.aggressive_truncation: true`:
      1. Create a JSON profile string with specific fields
      2. Use `profile.NewSelectionFromPresence()` to create selection from the profile's field presence
      3. Call `SetConfig()` with the loaded config and selection
      4. Verify inclusion checkboxes reflect actual field presence (disabled_agents selected, auto_update not selected)
      5. Call `Apply()` → `MarshalSparse()`
      6. Verify JSON output matches original profile's field set (no extra fields added)
    - `TestWizardOtherEditFlow_ModifyAndSave` — same setup, but modify a value before Apply:
      1. Load profile with `disabled_agents: ["sisyphus"]`
      2. Toggle "prometheus" to also be disabled
      3. Apply → MarshalSparse
      4. Verify JSON has `disabled_agents: ["prometheus", "sisyphus"]` (sorted)
    - `TestWizardOtherEditFlow_RemoveField` — load profile with a field, deselect it:
      1. Load profile with `auto_update: true`
      2. Toggle `auto_update` inclusion OFF
      3. Apply → MarshalSparse
      4. Verify JSON does NOT contain `auto_update`
    - `TestWizardOtherTemplateFlow_PreservesSelection` — load from template, verify selection matches template's fields:
      1. Create template JSON with specific fields
      2. Use `NewSelectionFromPresence()` 
      3. SetConfig with template config and selection
      4. Apply → MarshalSparse
      5. Verify output has exactly the same fields as template
    - `TestWizardOtherNewProfile_BlankSelection` — new profile with blank selection:
      1. Create WizardOther with blank selection (new profile)
      2. Set some values
      3. Apply → MarshalSparse
      4. With `selection == nil` (blank), ALL fields should be included by default (per `fieldSelected()` returning true when selection is nil)

  **Must NOT do**:
  - Do not modify production code
  - Do not test file I/O — test the in-memory flow only (SetConfig→Apply→MarshalSparse)
  - Do not depend on actual profile files on disk — use in-memory configs

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Integration tests spanning multiple components (profile loading, selection, wizard steps, serialization)
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 6)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (test helpers)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks_test.go:166-187` — `TestWizardHooksLoadsDisabledHooksSelectionFromJSONPresence` — pattern for testing that SetConfig loads selection correctly from FieldPresence
  - `internal/profile/profile_test.go` — `TestProfileLoadCapturesFieldPresence` — pattern for testing field presence detection

  **API/Type References**:
  - `internal/profile/selection.go:174-182` — `NewSelectionFromPresence()` — creates selection from field presence map
  - `internal/profile/profile.go:87-118` — `collectFieldPresence()` — how presence is detected from raw JSON
  - `internal/tui/views/wizard_other.go:1031-1430` — `SetConfig()` — how wizard loads config into view state
  - `internal/tui/views/wizard_other.go:1432-1908` — `Apply()` — how wizard writes view state back to config

  **Test References**:
  - `internal/profile/sparse_test.go` — `TestSparseSerializerProducesStablePrettyJSON` — pattern for verifying JSON output structure
  - `internal/profile/profile_test.go` — `TestRegressionSparsePersistenceContract` — comprehensive round-trip test pattern

  **WHY Each Reference Matters**:
  - `NewSelectionFromPresence():174` — This is the bridge from "what was in the JSON" to "what the wizard shows as selected". The edit flow test must use this to simulate real editing.
  - `SetConfig():1031` — Shows how the wizard populates its internal state from Config + selection. Tests must call this before Apply to simulate real wizard flow.
  - `TestRegressionSparsePersistenceContract` — This test in profile_test.go tests the full pipeline. The new tests should be similar but focused on wizard_other specifically.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Edit flow integration test passes
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherEditFlow -v -count=1`
    Expected Result: All edit flow tests pass (4+ tests)
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-7-edit-flow.txt

  Scenario: Template flow integration test passes
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherTemplateFlow -v -count=1`
    Expected Result: Template flow test passes, output matches template fields
    Failure Indicators: Extra fields in output, missing template fields
    Evidence: .sisyphus/evidence/task-7-template-flow.txt

  Scenario: New profile with nil selection includes all fields
    Tool: Bash
    Steps:
      1. Run `go test ./internal/tui/views/ -run TestWizardOtherNewProfile -v -count=1`
    Expected Result: With nil selection, all set values appear in JSON
    Failure Indicators: Fields missing when they should be included
    Evidence: .sisyphus/evidence/task-7-new-profile.txt

  Scenario: Full suite regression check
    Tool: Bash
    Steps:
      1. Run `make test`
    Expected Result: ALL tests pass (existing + new), zero failures
    Failure Indicators: Any test failure
    Evidence: .sisyphus/evidence/task-7-regression.txt
  ```

  **Commit**: YES
  - Message: `test(tui): add edit/template flow integration tests`
  - Files: `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `make test`

---

## Final Verification Wave

- [x] F1. **Plan Compliance Audit** — `oracle`
  - Must Have [7/7] | Must NOT Have [7/7] | Tasks [7/7] | VERDICT: APPROVE
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run test). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.sisyphus/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  - Lint [PASS - 5 pre-existing issues] | Tests [ALL PASS] | Files [2 modified] | VERDICT: APPROVE
  Run `make lint` + `make test`. Review all changed files for: `as any`/type assertion shortcuts, empty catches, console.log equivalents, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Full Test Suite QA** — `unspecified-high`
  - Tests [ALL PASS] | Races [CLEAN] | Flakes [CLEAN] | VERDICT: APPROVE
  Run `make test` from clean state. Verify ALL tests pass including new ones. Check test output for panics, races, or flaky behavior. Run twice to confirm stability.
  Output: `Tests [N/N pass] | Races [CLEAN/N] | Flakes [CLEAN/N] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  - Tasks [7/7 compliant] | Contamination [CLEAN] | Unaccounted [CLEAN] | VERDICT: APPROVE
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N] | Unaccounted [CLEAN/N] | VERDICT`

---

## Commit Strategy

- **Commit 1**: `fix(tui): add test helpers for wizard_other sparse field tests` — `wizard_other_test.go`
- **Commit 2**: `fix(tui): enable inclusion toggle for disabled-list sections in wizard_other` — `wizard_other.go`, `wizard_other_test.go`
- **Commit 3**: `test(tui): add round-trip tests for disabled-list sections` — `wizard_other_test.go`
- **Commit 4**: `test(tui): add round-trip tests for non-disabled sections` — `wizard_other_test.go`
- **Commit 5**: `test(tui): add empty-array and omission tests` — `wizard_other_test.go`
- **Commit 6**: `test(tui): add inclusion/value separation tests` — `wizard_other_test.go`
- **Commit 7**: `test(tui): add edit/template flow integration tests` — `wizard_other_test.go`

---

## Success Criteria

### Verification Commands
```bash
make test     # Expected: all tests pass, 0 failures
make lint     # Expected: 0 new warnings
go test ./internal/tui/views/ -run TestWizardOther -v  # Expected: all new tests pass
go test ./internal/tui/views/ -run TestWizardHooks -v  # Expected: all existing tests still pass
go test ./internal/profile/ -v                          # Expected: all serialization tests pass
```

### Final Checklist
- [x] Disabled-list sections can toggle inclusion via Space on Row 0
- [x] All round-trip tests pass (SetConfig→Apply→MarshalSparse→JSON verified)
- [x] Empty-array serialization correct (selected-but-empty → `[]`, unselected → absent)
- [x] Inclusion/value separation verified for bool, string, and slice fields
- [x] Edit/template flow preserves sparse semantics
- [x] No changes to config types, selection paths, or sparse.go
- [x] All existing tests continue passing
