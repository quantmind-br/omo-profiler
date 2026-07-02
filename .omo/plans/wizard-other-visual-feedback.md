# Other Settings Visual Feedback — Bug Fix Plan

## TL;DR

> **Quick Summary**: Fix 5 visual feedback bugs in the "Other Settings" wizard step where boolean fields lack proper cursor, selection, and value indication. The green ✓ indicator incorrectly reflects field value instead of field selection state, `simpleValueFocused` mode has no visual highlight, and `sectionStartWork` is inconsistent with other simple boolean sections.
> 
> **Deliverables**:
> - Correct ✓ indicator semantics (selection-based, not value-based) in both top-level and sub-section booleans
> - Visual highlight when editing simple boolean values (matching existing `subValueFocused` pattern)
> - Consistent `sectionStartWork` behavior as a simple boolean section
> - Tests covering all 5 bug fixes
> 
> **Estimated Effort**: Short
> **Parallel Execution**: YES — 2 waves + final verification
> **Critical Path**: Task 1 → Task 3 → F1-F4

---

## Context

### Original Request
Fix 5 visual feedback bugs in `wizard_other*.go` files where boolean fields don't provide adequate visual feedback for: (1) cursor position, (2) field selection state, and (3) current field value.

### Interview Summary
**Key Discussions**:
- User provided detailed bug spec with file/line references for all 5 bugs
- All bugs are in the render/update path of `WizardOther` step
- Scope is strictly limited to `internal/tui/views/wizard_other_*.go` files

**Research Findings**:
- `renderContent()` line 35-53: `valid` variable mirrors boolean value, not selection state
- `renderBoolField()` line 108-110: Same pattern — ✓ shows for `value=true`, not for `fieldSelected(path)=true`
- `renderSubSection()` has `subValueFocused` highlight pattern (lines 111-113) that simple booleans lack
- `isSimpleBooleanSection()` excludes `sectionStartWork` despite it being semantically identical
- `toggleSection()` has no case for `sectionStartWork`
- `topLevelFieldPath()` has no case for `sectionStartWork`

### Metis Review
**Identified Gaps** (addressed):
- `renderBoolField()` ✓ bug affects ~34 sub-section boolean fields — included in scope as Bug 1b
- `selection == nil` (new profile mode) makes `fieldSelected()` return true for all — this is correct behavior
- `sectionStartWork` conversion: `startWorkAutoCommitFieldPath` is nested (`start_work.auto_commit`) but this is already handled by `topLevelFieldPath()` for other nested paths
- Dead code at `update.go:245-256` after conversion — will be cleaned up
- `renderSubSection(sectionStartWork)` case at `render.go:239-240` becomes unreachable for the simple-boolean path — will be cleaned up

---

## Work Objectives

### Core Objective
Fix visual feedback in the "Other Settings" wizard step so users can clearly distinguish between field selection state (included in JSON), field value (on/off), and cursor focus (which field/mode is active).

### Concrete Deliverables
- Fixed ✓ indicator in `renderContent()` for 4 simple boolean sections
- Fixed ✓ indicator in `renderBoolField()` for ~34 sub-section boolean fields
- Visual highlight when `simpleValueFocused=true` (bold white value text)
- `sectionStartWork` as a consistent simple boolean section
- Removal of dead code paths after `sectionStartWork` conversion
- Tests for all fixes

### Definition of Done
- [x] `make test` passes with 0 failures
- [x] `make lint` passes
- [x] All 5 bugs verified fixed via automated tests

### Must Have
- ✓ indicator reflects `fieldSelected(path)`, not boolean value, everywhere
- Visual highlight (bold white via `labelStyle`) on value portion when `simpleValueFocused=true`
- `sectionStartWork` in `isSimpleBooleanSection()`, `topLevelFieldPath()`, and `toggleSection()`
- Tests for each bug fix

### Must NOT Have (Guardrails)
- **DO NOT add new style variables** — reuse `wizOtherLabelStyle`, `wizOtherEnabledStyle`, `wizOtherDimStyle` from `wizard_other.go:22-26`
- **DO NOT change the checkbox `[✓]` logic** — the checkbox already correctly reflects `fieldSelected()`; only the VALUE ✓ indicator needs fixing
- **DO NOT modify Update() key-handling logic** except: (a) adding `sectionStartWork` to `toggleSection()`, (b) removing dead inSubSection handling for `sectionStartWork`, and (c) guarding the generic `keys.Expand` handler to skip simple boolean sections (prevents them from entering expandable mode)
- **DO NOT modify files outside** `internal/tui/views/wizard_other_render.go`, `wizard_other_update.go`, `wizard_other_fields.go`, `wizard_other_test.go`
- **DO NOT refactor the rendering architecture** — surgical fixes only
- **DO NOT add keyboard shortcut hints** or help text changes
- **DO NOT touch `renderInclude()` or `renderValueField()`** closures — only `renderBoolField()` and the simple boolean path in `renderContent()`

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: YES (tests-after — add tests for each bug fix)
- **Framework**: `go test` via `make test`

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Library/Module**: Use Bash (`make test`) — Run tests, assert pass count
- **Code verification**: Use Grep/Read to verify code changes match spec

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Sequential — both edit renderContent() in the same block):
├── Task 1: Fix ✓ indicator semantics (Bug 1 + 1b) [quick]
└── Task 2: Add simpleValueFocused visual highlight (Bug 2 + 3) [quick] — AFTER Task 1

Wave 2 (After Wave 1 — depends on both fixes being in place):
└── Task 3: Make sectionStartWork a simple boolean (Bug 4 + 5) [unspecified-low]

Wave FINAL (After ALL tasks):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 1 → Task 2 → Task 3 → F1-F4 → user okay
Max Concurrent: 4 (Wave FINAL only)
```

> **⚠️ Momus Review Note**: Wave 1 was originally parallel but Momus identified that
> Tasks 1 and 2 both modify the same `renderContent()` block in `wizard_other_render.go`
> and both add tests to `wizard_other_test.go`. Running them in parallel would cause
> merge conflicts. Changed to sequential execution.

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 2, 3 | 1 |
| 2 | 1 | 3 | 1 |
| 3 | 1, 2 | F1-F4 | 2 |
| F1-F4 | 3 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **2 tasks sequential** — T1 → `quick`, then T2 → `quick`
- **Wave 2**: **1 task** — T3 → `unspecified-low`
- **FINAL**: **4 tasks parallel** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Fix ✓ indicator semantics (Bug 1 + 1b)

  **What to do**:
  - In `wizard_other_render.go` `renderContent()` (lines 35-53), change the `valid` variable assignment for all 4 simple boolean sections from mirroring the boolean value to using `w.fieldSelected(path)`. The `path` variable is already in scope from line 29's `w.topLevelFieldPath(section)` call. Replace:
    ```go
    case sectionAutoUpdate:
        value = onOff(w.autoUpdate)
        valid = w.autoUpdate          // BUG: mirrors value
    ```
    With:
    ```go
    case sectionAutoUpdate:
        value = onOff(w.autoUpdate)
        valid = w.fieldSelected(path)  // FIX: mirrors selection state
    ```
    Apply the same change for `sectionNewTaskSystemEnabled`, `sectionHashlineEdit`, `sectionModelFallback`.
  - In `wizard_other_render.go` `renderBoolField()` closure (line 108), change `if value {` to `if w.fieldSelected(path) {`. The `path` parameter is already passed to the closure (line 98 signature). This fixes the same ✓ bug for ~34 sub-section boolean fields across all expandable sections.
  - Add tests in `wizard_other_test.go` covering:
    - Simple boolean selected + value=false → ✓ present
    - Simple boolean selected + value=true → ✓ present
    - Simple boolean NOT selected + value=true → ✓ absent
    - Sub-section boolean selected + value=false → ✓ present (via `renderBoolField`)
    - Sub-section boolean selected + value=true → ✓ present
    - Sub-section boolean NOT selected + value=true → ✓ absent

  **Must NOT do**:
  - DO NOT change the checkbox `[✓]`/`[ ]` logic (lines 30-33) — it already correctly reflects `fieldSelected()`
  - DO NOT modify `renderInclude()` or `renderValueField()` closures
  - DO NOT add new style variables

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-file render logic change with clear before/after pattern. Small scope, straightforward.
  - **Skills**: `[]`
    - No specialized skills needed for Go code changes.
  - **Skills Evaluated but Omitted**:
    - `git-master`: Commits handled by orchestrator
    - `playwright`: No browser interaction needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Task 3
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/tui/views/wizard_other_render.go:29-57` — The simple boolean rendering block in `renderContent()`. Line 29 gets `path` via `topLevelFieldPath(section)`, lines 30-33 use `w.fieldSelected(path)` for checkbox — follow the same pattern for `valid`.
  - `internal/tui/views/wizard_other_render.go:98-115` — The `renderBoolField()` closure in `renderSubSection()`. Line 100 already uses `w.fieldSelected(path)` for checkbox. Line 108 `if value {` is the bug — change to `if w.fieldSelected(path) {`.

  **API/Type References**:
  - `internal/tui/views/wizard_other_fields.go:135-140` — `fieldSelected()` method: returns true if path is selected (or always true when `selection==nil` for new profile mode)

  **Test References**:
  - `internal/tui/views/wizard_other_test.go:204-221` — `TestWizardOtherLoadsCheckboxStateFromJSONPresence` — existing test pattern using `renderSubSection()` output and `strings.Contains()` assertions

  **WHY Each Reference Matters**:
  - render.go:29-57: This is the EXACT block being modified. The `path` variable at line 29 is already available — no new variables needed.
  - render.go:98-115: The `renderBoolField` closure is the second site of the same bug. The `path` parameter is already in the function signature.
  - fields.go:135-140: Understanding `fieldSelected()` behavior is critical — when `selection==nil` it returns true, which means ✓ will show for ALL fields in new-profile mode. This is correct and expected.
  - test.go:204-221: Shows the test pattern — create wizard, set state, call render, assert with `strings.Contains()`.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Simple boolean selected with value=false shows ✓
    Tool: Bash (go test)
    Preconditions: WizardOther with selection having "auto_update" selected, w.autoUpdate=false
    Steps:
      1. Create WizardOther, set selection with "auto_update" selected
      2. Set w.autoUpdate = false
      3. Call renderContent()
      4. Assert output contains "✓" on the Auto Update line
    Expected Result: The line for Auto Update contains both "[off]" and "✓"
    Failure Indicators: renderContent() output does NOT contain "✓" on Auto Update line when value=false
    Evidence: .sisyphus/evidence/task-1-checkmark-false-selected.txt

  Scenario: Simple boolean NOT selected with value=true does NOT show ✓
    Tool: Bash (go test)
    Preconditions: WizardOther with selection NOT having "auto_update" selected, w.autoUpdate=true
    Steps:
      1. Create WizardOther, set selection WITHOUT "auto_update" selected
      2. Set w.autoUpdate = true
      3. Call renderContent()
      4. Assert output does NOT contain "✓" on the Auto Update line (but still shows "[on]")
    Expected Result: The line for Auto Update contains "[on]" but NOT "✓"
    Failure Indicators: "✓" appears on Auto Update line despite field not being selected
    Evidence: .sisyphus/evidence/task-1-checkmark-true-unselected.txt

  Scenario: Sub-section renderBoolField selected with value=false shows ✓
    Tool: Bash (go test)
    Preconditions: WizardOther with selection having "experimental.aggressive_truncation" selected, w.expAggressiveTrunc=false
    Steps:
      1. Create WizardOther, set selection with "experimental.aggressive_truncation" selected
      2. Set w.expAggressiveTrunc = false
      3. Expand sectionExperimental, call renderSubSection(sectionExperimental)
      4. Assert output contains "✓" on the aggressive_truncation line
    Expected Result: aggressive_truncation line contains "[off]" AND "✓"
    Failure Indicators: No "✓" on aggressive_truncation line when selected but value=false
    Evidence: .sisyphus/evidence/task-1-subsection-checkmark-false-selected.txt
  ```

  **Evidence to Capture:**
  - [x] `task-1-checkmark-false-selected.txt` — test output showing ✓ with value=false
  - [x] `task-1-checkmark-true-unselected.txt` — test output showing no ✓ with value=true but unselected
  - [x] `task-1-subsection-checkmark-false-selected.txt` — sub-section ✓ test output

  **Commit**: YES
  - Message: `fix(tui): correct ✓ indicator to reflect field selection state`
  - Files: `internal/tui/views/wizard_other_render.go`, `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `make test`

- [x] 2. Add simpleValueFocused visual highlight (Bug 2 + 3)

  **What to do**:
  - In `wizard_other_render.go` `renderContent()`, after building the `value` string (around lines 50-53, after the switch block), add a condition to highlight the value when in value-editing mode. Insert BEFORE the `line := fmt.Sprintf(...)` at line 54:
    ```go
    if w.simpleValueFocused && section == w.currentSection && !w.inSubSection {
        value = labelStyle.Render(value)
    }
    ```
    This matches the pattern used in `renderBoolField()` at lines 111-113 where `subValueFocused` triggers `labelStyle.Render(valueRender)`.
  - Add tests in `wizard_other_test.go` covering:
    - `simpleValueFocused=true` on current section → value portion has bold/white styling
    - `simpleValueFocused=false` → value portion does NOT have bold/white styling
    - `simpleValueFocused=true` on a DIFFERENT section → no highlight on non-current sections

  **Must NOT do**:
  - DO NOT add new style variables — use existing `wizOtherLabelStyle`
  - DO NOT modify the cursor (`> `) logic
  - DO NOT change Update() key-handling

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small render logic addition, single conditional block, following an existing pattern.
  - **Skills**: `[]`
    - No specialized skills needed.
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: This is a Go TUI, not web frontend

  **Parallelization**:
  - **Can Run In Parallel**: NO (edits same `renderContent()` block as Task 1)
  - **Parallel Group**: Wave 1 (sequential after Task 1)
  - **Blocks**: Task 3
  - **Blocked By**: Task 1 (both modify `renderContent()` in `wizard_other_render.go`)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/tui/views/wizard_other_render.go:111-113` — The `subValueFocused` highlight in `renderBoolField()`. When `w.subValueFocused` is true for the current section/cursor, the value render gets wrapped in `labelStyle.Render()`. This is the EXACT pattern to replicate for `simpleValueFocused`.
  - `internal/tui/views/wizard_other_render.go:127-129` — Same pattern in `renderValueField()` for text input fields.

  **API/Type References**:
  - `internal/tui/views/wizard_other.go:320` — `simpleValueFocused bool` field in `WizardOther` struct
  - `internal/tui/views/wizard_other.go:22-26` — Style variables: `wizOtherLabelStyle` is bold white, `wizOtherDimStyle` is gray

  **WHY Each Reference Matters**:
  - render.go:111-113: This is the pattern being replicated. It shows exactly how to conditionally wrap a value render with `labelStyle` when a focus flag is active.
  - wizard_other.go:320: The `simpleValueFocused` field is set by `keys.Right` (update.go:341-343) and cleared by `keys.Left` (update.go:352-354) and Up/Down navigation (update.go:322, 327). Understanding its lifecycle ensures the highlight appears/disappears correctly.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Value highlighted when simpleValueFocused=true
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionAutoUpdate, simpleValueFocused=true, selection has "auto_update" selected
    Steps:
      1. Create WizardOther, set currentSection to sectionAutoUpdate
      2. Set simpleValueFocused = true
      3. Call renderContent()
      4. Find the Auto Update line in the output
      5. Assert the value portion "[off]" or "[on]" is rendered differently than when simpleValueFocused=false (wrapped in labelStyle which adds bold+white ANSI codes)
    Expected Result: The value portion contains bold/white ANSI styling when simpleValueFocused=true
    Failure Indicators: Value portion renders identically whether simpleValueFocused is true or false
    Evidence: .sisyphus/evidence/task-2-value-highlighted.txt

  Scenario: Value NOT highlighted when simpleValueFocused=false
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionAutoUpdate, simpleValueFocused=false
    Steps:
      1. Create WizardOther, set currentSection to sectionAutoUpdate
      2. Set simpleValueFocused = false
      3. Call renderContent()
      4. Find the Auto Update line
      5. Assert the value portion does NOT have bold/white styling
    Expected Result: Value rendered in default/dim style
    Failure Indicators: Value has bold/white styling when it shouldn't
    Evidence: .sisyphus/evidence/task-2-value-not-highlighted.txt

  Scenario: Value NOT highlighted on non-current section
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionNewTaskSystemEnabled, simpleValueFocused=true
    Steps:
      1. Create WizardOther, set currentSection to sectionNewTaskSystemEnabled
      2. Set simpleValueFocused = true
      3. Call renderContent()
      4. Find the Auto Update line (NOT the current section)
      5. Assert Auto Update's value does NOT have bold/white styling
    Expected Result: Only the current section's value is highlighted
    Failure Indicators: Non-current section values get highlighted
    Evidence: .sisyphus/evidence/task-2-wrong-section-not-highlighted.txt
  ```

  **Evidence to Capture:**
  - [x] `task-2-value-highlighted.txt` — test output showing highlight applied
  - [x] `task-2-value-not-highlighted.txt` — test output showing no highlight
  - [x] `task-2-wrong-section-not-highlighted.txt` — test output for non-current section

  **Commit**: YES
  - Message: `fix(tui): add visual highlight for simple boolean value-editing mode`
  - Files: `internal/tui/views/wizard_other_render.go`, `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `make test`

- [x] 3. Make sectionStartWork a consistent simple boolean section (Bug 4 + 5)

  **What to do**:
  - **(a) `wizard_other_fields.go`** — Add `sectionStartWork` to `isSimpleBooleanSection()` (line 268-275):
    ```go
    case sectionAutoUpdate, sectionNewTaskSystemEnabled, sectionHashlineEdit, sectionModelFallback, sectionStartWork:
        return true
    ```
  - **(b) `wizard_other_fields.go`** — Add `sectionStartWork` case to `topLevelFieldPath()` (line 237-266):
    ```go
    case sectionStartWork:
        return startWorkAutoCommitFieldPath
    ```
    This makes `renderContent()` at line 29 get `path="start_work.auto_commit"` and enter the simple boolean rendering path instead of falling through to the expandable section path.
  - **(c) `wizard_other_render.go`** — Add `sectionStartWork` to the `switch` block in `renderContent()` (lines 36-49) that builds the `value` and `valid` variables:
    ```go
    case sectionStartWork:
        value = onOff(w.startWorkAutoCommit)
        valid = w.fieldSelected(path)
    ```
  - **(d) `wizard_other_update.go`** — Add `sectionStartWork` case to `toggleSection()` (line 373-410), following the exact pattern of the other 4 simple booleans:
    ```go
    case sectionStartWork:
        if w.simpleValueFocused {
            w.startWorkAutoCommit = !w.startWorkAutoCommit
        } else {
            w.toggleFieldSelection(startWorkAutoCommitFieldPath)
        }
    ```
  - **(e) `wizard_other_update.go`** — Guard the generic `keys.Expand` handler (lines 333-339) to skip simple boolean sections. Currently, pressing Enter on ANY top-level section sets `inSubSection=true` and expands it. Simple booleans must NOT enter this path. Add a guard:
    ```go
    case key.Matches(msg, w.keys.Expand):
        if w.isSimpleBooleanSection(w.currentSection) {
            // Simple booleans toggle value directly, don't expand
            w.toggleSection()
            break
        }
        w.sectionExpanded[w.currentSection] = !w.sectionExpanded[w.currentSection]
        // ... rest of existing code
    ```
    > **⚠️ Momus Review Note**: The original plan claimed `update.go:245-256` was dead code after
    > converting `sectionStartWork` to simple boolean. Momus identified this was FALSE — the generic
    > `keys.Expand` handler at lines 333-339 would still route `sectionStartWork` into `inSubSection`
    > mode, making the handler reachable. The fix is to guard the Expand handler, THEN the code
    > truly becomes dead and can be safely removed.
    After adding the Expand guard, remove the dead `sectionStartWork` inSubSection handling at lines 245-256 and the `renderSubSection(sectionStartWork)` case at `render.go:239-240`.
  - **(f) `wizard_other_test.go`** — Add tests covering:
    - `isSimpleBooleanSection(sectionStartWork)` returns true
    - `sectionStartWork` renders as a simple boolean line (no ▶/▼ icon) in `renderContent()`
    - `toggleSection()` with `sectionStartWork` + `simpleValueFocused=false` → toggles field selection
    - `toggleSection()` with `sectionStartWork` + `simpleValueFocused=true` → toggles `startWorkAutoCommit` value
    - Right arrow on `sectionStartWork` → sets `simpleValueFocused=true`
    - Apply still persists `startWorkAutoCommit` correctly after conversion

  **Must NOT do**:
  - DO NOT change the `Apply()` or `SetConfig()` logic for `sectionStartWork` — it already works correctly with `startWorkAutoCommitFieldPath`
  - DO NOT modify other section types
  - DO NOT add new section constants or field paths

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Changes span 3 files but each change is small and follows established patterns. No complex logic, just wiring.
  - **Skills**: `[]`
    - No specialized skills needed.
  - **Skills Evaluated but Omitted**:
    - `git-master`: Commits handled by orchestrator

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential after Wave 1)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 1 (needs correct ✓ semantics), Task 2 (needs visual highlight)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/tui/views/wizard_other_fields.go:268-275` — `isSimpleBooleanSection()` — add `sectionStartWork` to the existing case list
  - `internal/tui/views/wizard_other_fields.go:249-256` — `topLevelFieldPath()` cases for `sectionAutoUpdate` through `sectionModelFallback` — follow the same pattern for `sectionStartWork`
  - `internal/tui/views/wizard_other_update.go:375-398` — `toggleSection()` cases for the 4 existing simple booleans — copy the pattern for `sectionStartWork`
  - `internal/tui/views/wizard_other_render.go:36-49` — The `switch` block in `renderContent()` that builds `value`/`valid` for simple booleans — add `sectionStartWork` case

  **API/Type References**:
  - `internal/tui/views/wizard_other_fields.go:62` — `startWorkAutoCommitFieldPath = "start_work.auto_commit"` — the field path constant to use
  - `internal/tui/views/wizard_other.go:199` — `startWorkAutoCommit bool` — the boolean field in the struct
  - `internal/tui/views/wizard_other.go:98` — `sectionStartWork` — the section constant

  **Expand Guard Reference**:
  - `internal/tui/views/wizard_other_update.go:333-339` — The generic `keys.Expand` handler that sets `inSubSection=true` for ANY section. Must add `isSimpleBooleanSection()` guard to prevent simple booleans from entering expand mode.

  **Dead Code References** (to remove AFTER adding Expand guard):
  - `internal/tui/views/wizard_other_update.go:245-256` — Special `sectionStartWork` inSubSection handler — dead after Expand guard prevents simple booleans from entering inSubSection
  - `internal/tui/views/wizard_other_render.go:239-240` — `case sectionStartWork:` in `renderSubSection()` — unreachable after Expand guard

  **WHY Each Reference Matters**:
  - fields.go:268-275: The exact function to modify — adding one more constant to the case list
  - fields.go:249-256: The `topLevelFieldPath()` pattern — `sectionStartWork` must return `startWorkAutoCommitFieldPath` so `renderContent()` enters the simple boolean path
  - update.go:375-398: The exact toggle pattern to replicate — `if simpleValueFocused toggle value, else toggle selection`
  - render.go:36-49: The switch block that builds value/valid — must add `sectionStartWork` case or it won't render value/valid
  - update.go:333-339: The generic Expand handler — without a guard, pressing Enter on `sectionStartWork` would set `inSubSection=true` even after conversion, creating a broken invisible-focus state. The guard makes simple booleans route Enter through `toggleSection()` instead.
  - update.go:245-256: Only truly dead AFTER the Expand guard is in place. This was the inSubSection handler that let users toggle via Enter/Space within the expanded section. Now users toggle directly from the main list.
  - render.go:239-240: Only truly unreachable AFTER the Expand guard prevents simple booleans from expanding.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: sectionStartWork renders as simple boolean (no expand icon)
    Tool: Bash (go test)
    Preconditions: WizardOther with default state
    Steps:
      1. Create WizardOther with selection having "start_work.auto_commit" selected
      2. Call renderContent()
      3. Find the "Start Work" line
      4. Assert it contains "[✓]" and "[off]" (or "[on]") — simple boolean format
      5. Assert it does NOT contain "▶" or "▼" expand icons
    Expected Result: Start Work renders as `[✓] Start Work: [off] ✓` (not as expandable section)
    Failure Indicators: Line contains ▶ or ▼, or lacks [on]/[off] value
    Evidence: .sisyphus/evidence/task-3-startwork-simple-render.txt

  Scenario: Space toggles field selection when simpleValueFocused=false
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionStartWork, simpleValueFocused=false, selection has "start_work.auto_commit" selected
    Steps:
      1. Create WizardOther, set currentSection=sectionStartWork
      2. Set simpleValueFocused=false
      3. Press Space (send KeyMsg{Type: tea.KeySpace})
      4. Assert selection.IsSelected("start_work.auto_commit") is now false
      5. Assert w.startWorkAutoCommit is unchanged
    Expected Result: Field selection toggled, value unchanged
    Failure Indicators: Value toggled instead of selection, or nothing happened
    Evidence: .sisyphus/evidence/task-3-startwork-toggle-selection.txt

  Scenario: Space toggles value when simpleValueFocused=true
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionStartWork, simpleValueFocused=true
    Steps:
      1. Create WizardOther, set currentSection=sectionStartWork
      2. Set simpleValueFocused=true
      3. Press Space
      4. Assert w.startWorkAutoCommit is now true (was false)
      5. Assert selection state unchanged
    Expected Result: Boolean value toggled, selection unchanged
    Failure Indicators: Selection toggled instead of value, or nothing happened
    Evidence: .sisyphus/evidence/task-3-startwork-toggle-value.txt

  Scenario: Right arrow enters value focus mode
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionStartWork, simpleValueFocused=false
    Steps:
      1. Create WizardOther, set currentSection=sectionStartWork
      2. Press Right arrow
      3. Assert simpleValueFocused=true
    Expected Result: simpleValueFocused becomes true
    Failure Indicators: simpleValueFocused remains false, or section expands instead
    Evidence: .sisyphus/evidence/task-3-startwork-right-arrow.txt

  Scenario: Enter on sectionStartWork toggles value, does NOT expand
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionStartWork, simpleValueFocused=false
    Steps:
      1. Create WizardOther, set currentSection=sectionStartWork
      2. Press Enter (send KeyMsg{Type: tea.KeyEnter})
      3. Assert inSubSection is still false
      4. Assert sectionExpanded[sectionStartWork] is still false
      5. Assert toggleSection was invoked (field selection toggled)
    Expected Result: Enter triggers toggle, not expand
    Failure Indicators: Section enters inSubSection mode or shows expand icon
    Evidence: .sisyphus/evidence/task-3-enter-no-expand.txt

  Scenario: Enter on non-simple-boolean section still expands normally
    Tool: Bash (go test)
    Preconditions: WizardOther with currentSection=sectionExperimental
    Steps:
      1. Create WizardOther, set currentSection=sectionExperimental
      2. Press Enter
      3. Assert sectionExpanded[sectionExperimental] is true
      4. Assert inSubSection is true
    Expected Result: Expandable sections still work normally
    Failure Indicators: Expandable section fails to expand
    Evidence: .sisyphus/evidence/task-3-expand-still-works.txt

  Scenario: Dead code removed (update.go inSubSection handler)
    Tool: Grep
    Preconditions: After code changes applied
    Steps:
      1. Search wizard_other_update.go for "sectionStartWork" within an inSubSection block
      2. Assert the old handler (lines 245-256 pattern) is removed
    Expected Result: No inSubSection-specific handling for sectionStartWork remains
    Failure Indicators: Old dead code still present
    Evidence: .sisyphus/evidence/task-3-dead-code-removed.txt
  ```

  **Evidence to Capture:**
  - [x] `task-3-startwork-simple-render.txt` — render output showing simple boolean format
  - [x] `task-3-startwork-toggle-selection.txt` — test output for selection toggle
  - [x] `task-3-startwork-toggle-value.txt` — test output for value toggle
  - [x] `task-3-startwork-right-arrow.txt` — test output for Right arrow behavior
  - [x] `task-3-enter-no-expand.txt` — test output confirming Enter doesn't expand simple booleans
  - [x] `task-3-expand-still-works.txt` — test output confirming expandable sections still work
  - [x] `task-3-dead-code-removed.txt` — grep output confirming dead code removal

  **Commit**: YES
  - Message: `fix(tui): make sectionStartWork a consistent simple boolean section`
  - Files: `internal/tui/views/wizard_other_fields.go`, `internal/tui/views/wizard_other_update.go`, `internal/tui/views/wizard_other_render.go`, `internal/tui/views/wizard_other_test.go`
  - Pre-commit: `make test`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run `make test`). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.sisyphus/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `make test` + `make lint`. Review all changed files for: `as any`/`@ts-ignore` equivalent patterns, empty catches, commented-out code, unused imports. Check for AI slop: excessive comments, over-abstraction, generic variable names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (all 5 bugs fixed together). Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (`git diff`). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| # | Message | Files | Pre-commit |
|---|---------|-------|------------|
| 1 | `fix(tui): correct ✓ indicator to reflect field selection state` | `wizard_other_render.go`, `wizard_other_test.go` | `make test` |
| 2 | `fix(tui): add visual highlight for simple boolean value-editing mode` | `wizard_other_render.go`, `wizard_other_test.go` | `make test` |
| 3 | `fix(tui): make sectionStartWork a consistent simple boolean section` | `wizard_other_fields.go`, `wizard_other_update.go`, `wizard_other_render.go`, `wizard_other_test.go` | `make test` |

---

## Success Criteria

### Verification Commands
```bash
make test    # Expected: PASS, 0 failures
make lint    # Expected: 0 errors
```

### Final Checklist
- [x] Simple booleans show ✓ based on `fieldSelected()`, not value
- [x] Sub-section booleans show ✓ based on `fieldSelected()`, not value
- [x] `simpleValueFocused=true` has visible bold-white highlight on value
- [x] `sectionStartWork` renders and behaves like other simple booleans
- [x] Dead code from `sectionStartWork` conversion cleaned up
- [x] All "Must NOT Have" guardrails verified absent
- [x] All tests pass, no regressions
