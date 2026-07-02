# Code Quality Refactoring — Full Implementation

## TL;DR

> **Quick Summary**: Implement all 17 code quality refactoring items from IDEATION_CODE_QUALITY.md across the TUI wizard layer of omo-profiler. The work targets God Object anti-patterns, massive code duplication (165+ lines of exact copies, ~750 lines of boilerplate), and extreme function lengths (1,422-line Update()) following the recommended trivial→small→medium→large order.
>
> **Deliverables**:
> - Shared utilities: `pathutil.go`, `list_renderer.go`, textinput factory, centralized color usage
> - Deduplicated code: merged Apply/applyAllAgentFields, unified expandPath, typed errors, ListRenderer
> - Split files: `wizard_other.go` (3,768→5 files), `wizard_categories.go` (1,345→2 files)
> - Structural improvements: table-driven wizard steps, extracted handlers, help consolidation
> - Field path consolidation: single source of truth for hardcoded paths
>
> **Estimated Effort**: XL (17 items, ~8,000+ lines affected)
> **Parallel Execution**: YES - 6 waves
> **Critical Path**: Task 1 (baseline) → Tasks 2-4 (trivial shared utils) → Tasks 5-7 (small dedup) → Tasks 8-11 (medium extract) → Tasks 12-14 (large splits) → Tasks 15-17 (interface/cleanup)

---

## Context

### Original Request
Implement all 17 refactoring items from IDEATION_CODE_QUALITY.md in the omo-profiler codebase.

### Interview Summary
**Key Discussions**:
- Scope: ALL 17 items (2 critical + 5 major + 7 minor + 3 suggestions)
- Execution order: Follow IDEATION recommended order (trivial → small → medium → large)
- Test strategy: Tests-after (refactor first, then add regression tests)
- CQ-009 approach: go:generate → downgraded to "consolidate to single file" per Metis advice
- Additional duplication found: ~488 lines across 5 patterns (handleSaveCustomModel, validateField, etc.) — explicitly excluded from this scope

**Research Findings**:
- All file sizes and duplication patterns CONFIRMED by explore agents
- Test coverage is EXCELLENT: 28 test files, all target files covered
- All code is in `package views` (same package) — file splits within package won't break tests
- Colors in local vars match centralized styles.go hex values exactly (safe to replace)
- wizard_other.go has 160 viewport.SetContent calls (needs refreshView extraction)

### Metis Review
**Identified Gaps** (addressed):
- 5 additional duplication patterns (~488 lines) not in IDEATION report → excluded as future work (CQ-018+)
- CQ-001 should be 3-phase commit: (1) pure file split, (2) dispatch table, (3) refreshView
- CQ-009 go:generate is a rabbit hole → downgraded to file consolidation
- CQ-012 WizardStep interface already partially exists in step.go → redefined as table-driven integration with CQ-005
- CQ-010 depends on CQ-007 completing first (naming consistency follows ListRenderer extraction)
- Receiver type change on WizardAgents is safe (MVU pattern returns new value)
- All tests are same-package — file splits won't break imports

---

## Work Objectives

### Core Objective
Refactor the TUI wizard layer to eliminate God Object anti-patterns, reduce code duplication, and improve maintainability — without changing any external behavior or API.

### Concrete Deliverables
- `internal/tui/views/pathutil.go` — shared path expansion utility
- `internal/tui/views/list_renderer.go` — shared list pagination component
- `internal/tui/views/wizard_other.go` split into 5 files (3,768 → ~5 files of 200-600 lines each)
- `internal/tui/views/wizard_categories.go` split into 2 files
- `internal/profile/fields.go` — consolidated field paths (single source)
- Merged `Apply()`/`applyAllAgentFields()` in wizard_agents.go
- Data-driven textinput dispatch table replacing 30 identical handler blocks
- Centralized color usage from `styles.go` across 6+ view files
- Table-driven wizard step transitions
- Extracted handlers in wizard_agents.go and app.go

### Definition of Done
- [ ] `make test` passes after every commit
- [ ] `make lint` passes on final state
- [ ] `go build ./...` succeeds
- [ ] All 17 CQ items addressed
- [ ] No behavioral changes (same TUI behavior)

### Must Have
- ALL 17 CQ items implemented
- `make test` passes between EVERY atomic commit
- `pre-refactor-baseline` git tag before any changes
- All code stays in `package views` (no package splits)
- CQ-001 done as 3-phase commit (split → dispatch → refreshView)
- Zero semantic/behavioral changes

### Must NOT Have (Guardrails)
- NO new packages (everything stays in existing packages)
- NO behavioral changes (UI must look and behave identically)
- NO test file modifications except when function signatures change
- NO mixing of file splits with code changes (split first, then refactor)
- NO inclusion of 5 additional duplication patterns (future work)
- NO `go:generate` tool creation (downgraded to file consolidation)
- NO changes to non-TUI packages unless directly required
- NO WizardStep interface expansion beyond existing methods
- NO modification of config/types.go type definitions

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (28 test files, Go standard + testify)
- **Automated tests**: Tests-after (refactor first, add regression tests if gaps found)
- **Framework**: Go standard testing + testify assertions
- **Command**: `make test` (runs `go test -v ./...`)

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Go source**: Use Bash (`go test`, `go build`, `go vet`) — compile, test, lint
- **Refactoring verification**: Use Bash — diff test output before/after
- **Structural verification**: Use Bash (`wc -l`, `grep`) — verify file sizes and patterns

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 0 (Foundation - MUST complete first):
└── Task 0: Create pre-refactor-baseline git tag [quick]

Wave 1 (Trivial shared utilities - ALL parallel):
├── Task 1: CQ-015 - Consolidate triplicated expandPath (3→1) [quick]
├── Task 2: CQ-016 - Remove view-specific color redefinitions [quick]
├── Task 3: CQ-017 - Fix inconsistent receiver types in wizard_agents [quick]
└── Task 4: CQ-011 - Extract textinput factory [quick]

Wave 2 (Small deduplication - ALL parallel):
├── Task 5: CQ-002 - Merge duplicate Apply functions in wizard_agents [unspecified-high]
├── Task 6: CQ-008 - Replace string error checks with typed errors [quick]
└── Task 7: CQ-007 - Extract shared ListRenderer component [unspecified-high]

Wave 3 (Medium structural changes - partial parallel):
├── Task 8: CQ-005 - Table-driven wizard step transitions [deep]
├── Task 9: CQ-003 - Extract wizard_agents Update handlers [unspecified-high]
├── Task 10: CQ-004 - Extract app.go message handlers [unspecified-high]
└── Task 11: CQ-010 - Standardize scroll variable naming (depends: Task 7) [quick]

Wave 4 (Large splits - sequential within, parallel across):
├── Task 12: CQ-001 - Split wizard_other.go into 5 files (3-phase) [deep]
└── Task 13: CQ-006 - Split wizard_categories.go into 2 files [unspecified-high]

Wave 5 (Interface & cleanup - ALL parallel):
├── Task 14: CQ-012 - Use WizardStep interface in wizard.go [unspecified-high]
├── Task 15: CQ-013 - Consolidate help rendering in app.go [quick]
├── Task 16: CQ-009 - Consolidate hardcoded field paths [unspecified-high]
└── Task 17: CQ-014 - Split config/types.go by domain [quick]

Wave FINAL (Verification - ALL parallel):
├── Task F1: Plan compliance audit [oracle]
├── Task F2: Code quality review [unspecified-high]
├── Task F3: Build + test verification [unspecified-high]
└── Task F4: Scope fidelity check [deep]
→ Present results → Get explicit user okay

Critical Path: Task 0 → Task 1-4 → Task 5-7 → Task 8-11 → Task 12 → Task 13 → Task 14-17 → F1-F4
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 4 (Waves 1, 2, 3, 5)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 0 | - | 1-17 | 0 |
| 1 | 0 | - | 1 |
| 2 | 0 | - | 1 |
| 3 | 0 | - | 1 |
| 4 | 0 | 5 | 1 |
| 5 | 4 | - | 2 |
| 6 | 0 | - | 2 |
| 7 | 0 | 11 | 2 |
| 8 | 0 | 14 | 3 |
| 9 | 0 | - | 3 |
| 10 | 0 | - | 3 |
| 11 | 7 | - | 3 |
| 12 | 0 | 13 | 4 |
| 13 | 12 | - | 4 |
| 14 | 8 | - | 5 |
| 15 | 0 | - | 5 |
| 16 | 0 | - | 5 |
| 17 | 0 | - | 5 |

### Agent Dispatch Summary

- **Wave 0**: 1 task — T0 → `quick`
- **Wave 1**: 4 tasks — T1-T4 → `quick`
- **Wave 2**: 3 tasks — T5 → `unspecified-high`, T6 → `quick`, T7 → `unspecified-high`
- **Wave 3**: 4 tasks — T8 → `deep`, T9-T10 → `unspecified-high`, T11 → `quick`
- **Wave 4**: 2 tasks — T12 → `deep`, T13 → `unspecified-high`
- **Wave 5**: 4 tasks — T14 → `unspecified-high`, T15 → `quick`, T16 → `unspecified-high`, T17 → `quick`
- **FINAL**: 4 tasks — F1 → `oracle`, F2-F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 0. Create pre-refactor-baseline git tag

  **What to do**:
  - Create an annotated git tag `pre-refactor-baseline` on the current HEAD
  - Run `make test` and capture output as baseline reference
  - Record current test count and test names for later comparison

  **Must NOT do**:
  - Modify any code

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single git command + test run
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 0 (foundation)
  - **Blocks**: Tasks 1-17
  - **Blocked By**: None

  **References**:
  - `Makefile` — contains `test` target definition: `go test -v ./...`

  **Acceptance Criteria**:
  - [ ] Git tag `pre-refactor-baseline` exists
  - [ ] `make test` passes (baseline recorded)

  **QA Scenarios:**
  ```
  Scenario: Baseline tag created
    Tool: Bash
    Preconditions: Clean git state
    Steps:
      1. Run: git tag -l "pre-refactor-baseline"
      2. Assert: output contains "pre-refactor-baseline"
      3. Run: make test
      4. Assert: exit code 0
    Expected Result: Tag exists, all tests pass
    Evidence: .sisyphus/evidence/task-0-baseline.txt
  ```

  **Commit**: YES
  - Message: `chore: tag pre-refactor-baseline`
  - Files: (no files, just tag)

---

- [x] 1. CQ-015 — Consolidate triplicated expandPath functions

  **What to do**:
  - Create `internal/tui/views/pathutil.go` with a single `expandPath(path string) (string, error)` function
  - The function expands `~` to home directory and converts relative paths to absolute
  - Replace calls to `expandExportPath()` in `export.go`, `expandPath()` in `import.go`, and `expandSchemaPath()` in `schema_check.go` with the shared function
  - Delete the 3 local function definitions

  **Must NOT do**:
  - Change any path expansion behavior
  - Export the function (keep it unexported within `package views`)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple extract-and-replace, ~50 lines of new code
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/export.go:192-210` — `expandExportPath()`: expand ~ then abs path, most complete version with explicit error checks
  - `internal/tui/views/import.go:175-195` — `expandPath()`: same logic with comments
  - `internal/tui/views/schema_check.go:112-122` — `expandSchemaPath()`: slightly simpler (one-liner return)

  **API/Type References**:
  - All 3 functions are unexported (same package `views`), so no import changes needed

  **WHY Each Reference Matters**:
  - Use `export.go:192-210` as the canonical implementation (most explicit error handling)
  - The `schema_check.go` version uses `filepath.Abs` directly in return — this is equivalent and preferred

  **Acceptance Criteria**:
  - [ ] `internal/tui/views/pathutil.go` exists with single `expandPath` function
  - [ ] No function named `expandExportPath`, `expandPath` (old), or `expandSchemaPath` exists in any view file
  - [ ] `make test` passes
  - [ ] `grep -r "expandExportPath\|expandSchemaPath" internal/tui/views/` returns no results

  **QA Scenarios:**
  ```
  Scenario: Shared function replaces all 3 local variants
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -c "func expand" internal/tui/views/pathutil.go
      2. Assert: output is "1" (single function)
      3. Run: grep -r "expandExportPath\|expandSchemaPath" internal/tui/views/ --include="*.go"
      4. Assert: exit code 1 (no matches)
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: Single expandPath in pathutil.go, no old variants remain, tests pass
    Evidence: .sisyphus/evidence/task-1-expand-path-consolidated.txt

  Scenario: Path expansion behavior unchanged
    Tool: Bash
    Preconditions: pathutil.go created
    Steps:
      1. Run: grep -A 10 "func expandPath" internal/tui/views/pathutil.go
      2. Assert: contains "HasPrefix" for ~ and "filepath.Abs" or "filepath.IsAbs"
      3. Run: make test
      4. Assert: all import/export/schema_check tests pass
    Expected Result: Function handles ~ expansion and relative→absolute conversion
    Failure Indicators: Any test failure in import_test.go, export_test.go, or schema_check_test.go
    Evidence: .sisyphus/evidence/task-1-path-behavior.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): consolidate triplicated expandPath into pathutil.go`
  - Files: `internal/tui/views/pathutil.go` (new), `export.go`, `import.go`, `schema_check.go`
  - Pre-commit: `make test`

---

- [x] 2. CQ-016 — Remove view-specific color redefinitions

  **What to do**:
  - Remove local color variables (`exportWhite`, `importGray`, `diffPurple`, `wizAgentPurple`, `wizOtherPurple`, etc.) from all view files
  - Replace all usages with imports from centralized `internal/tui/styles.go` palette (`tui.Purple`, `tui.Gray`, `tui.White`, etc.)
  - Fix `model_selector.go` inline hex values (#7D56F4, #6C7086, etc.) to use centralized colors
  - Keep view-specific STYLE compositions (e.g., `titleStyle := lipgloss.NewStyle().Foreground(tui.Purple)`) — only replace the color SOURCE

  **Must NOT do**:
  - Change style compositions or layout
  - Change any visual output
  - Remove colors from `styles.go`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Find-and-replace of color variable references, no logic changes
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/styles.go:5-14` — Centralized palette: Purple (#7D56F4), Magenta (#FF6AC1), Cyan (#78DCE8), Green (#A6E3A1), Red (#F38BA8), Yellow (#F9E2AF), Gray (#6C7086), White (#CDD6F4)

  **API/Type References**:
  - `internal/tui/views/export.go:16-22` — `exportWhite`, `exportGray`, `exportRed`, `exportPurple`, `exportYellow`
  - `internal/tui/views/import.go:16-20` — `importWhite`, `importGray`, `importRed`
  - `internal/tui/views/diff.go:16-23` — `diffPurple`, `diffMagenta`, `diffGreen`, `diffRed`, `diffGray`, `diffWhite`
  - `internal/tui/views/wizard_agents.go:146-153` — `wizAgentPurple`, `wizAgentGray`, `wizAgentText`, `wizAgentRed`, `wizAgentGreen`, `wizAgentPink`
  - `internal/tui/views/wizard_categories.go:40-46` — `wizCatPurple`, `wizCatGray`, `wizCatText`, `wizCatRed`, `wizCatGreen`
  - `internal/tui/views/wizard_other.go:21-26` — `wizOtherPurple`, `wizOtherGreen`, `wizOtherGray`, `wizOtherWhite`
  - `internal/tui/views/model_selector.go:407-413` — inline hex values in `renderList()`

  **WHY Each Reference Matters**:
  - All local hex values match the centralized palette exactly (verified by explore agents). Safe 1:1 replacement.

  **Acceptance Criteria**:
  - [ ] No local color variables remain in any view file (no `exportWhite`, `diffPurple`, etc.)
  - [ ] All view files import `tui` package colors
  - [ ] `make test` passes
  - [ ] `grep -r "wizAgent\|wizCat\|wizOther\|exportWhite\|importGray\|diffPurple" internal/tui/views/` returns no results

  **QA Scenarios:**
  ```
  Scenario: All local color vars eliminated
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -r "exportWhite\|exportGray\|exportRed\|exportPurple\|exportYellow" internal/tui/views/ --include="*.go" -l
      2. Assert: no output (no files match)
      3. Run: grep -r "importWhite\|importGray\|importRed" internal/tui/views/ --include="*.go" -l
      4. Assert: no output
      5. Run: grep -r "diffPurple\|diffMagenta\|diffGreen\|diffRed\|diffGray\|diffWhite" internal/tui/views/ --include="*.go" -l
      6. Assert: no output
      7. Run: grep -r "wizAgent\|wizCat\|wizOther" internal/tui/views/ --include="*.go" -l
      8. Assert: no output
      9. Run: make test
      10. Assert: exit code 0
    Expected Result: Zero local color variables, all tests pass
    Evidence: .sisyphus/evidence/task-2-colors-deduplicated.txt

  Scenario: model_selector.go inline hex values replaced
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run: grep "#7D56F4\|#6C7086\|#CDD6F4\|#FF6AC1\|#F38BA8" internal/tui/views/model_selector.go
      2. Assert: exit code 1 (no inline hex values remain)
    Expected Result: No hardcoded hex colors in model_selector.go
    Evidence: .sisyphus/evidence/task-2-selector-hex-removed.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): remove view-specific color vars, use centralized styles`
  - Files: All view files with local color vars + `model_selector.go`
  - Pre-commit: `make test`

---

- [x] 3. CQ-017 — Fix inconsistent receiver types in wizard_agents

  **What to do**:
  - Audit ALL methods on `WizardAgents` struct in `wizard_agents.go`
  - Convert ALL methods to pointer receiver `(w *WizardAgents)` consistently
  - Exception: `Update()` must continue returning by value (Bubble Tea MVU pattern), but can use pointer receiver internally
  - `View()` also returns by value — keep this pattern

  **Must NOT do**:
  - Change receiver types on other wizard types (out of scope)
  - Change any method behavior

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical find-and-replace of receiver types
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go` — 14 functions use `(w WizardAgents)`, 8 use `(w *WizardAgents)`
  - The AGENTS.md notes: "Go convention recommends using pointer receivers consistently when any method mutates state"

  **WHY Each Reference Matters**:
  - Since WizardAgents has maps and slices (large struct header), value receivers cause unnecessary copies
  - Safe because Update() returns new value by convention (MVU pattern)

  **Acceptance Criteria**:
  - [ ] ALL methods on `WizardAgents` use `(w *WizardAgents)` receiver
  - [ ] `make test` passes (verifies no semantic change)
  - [ ] `grep -c "func (w WizardAgents)" internal/tui/views/wizard_agents.go` returns 0

  **QA Scenarios:**
  ```
  Scenario: All methods use pointer receivers
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -c "func (w WizardAgents)" internal/tui/views/wizard_agents.go
      2. Assert: output is "0"
      3. Run: grep -c "func (w \*WizardAgents)" internal/tui/views/wizard_agents.go
      4. Assert: output > 0
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: Zero value receivers, all pointer receivers, tests pass
    Evidence: .sisyphus/evidence/task-3-receivers-fixed.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): fix inconsistent receiver types in wizard_agents`
  - Files: `internal/tui/views/wizard_agents.go`
  - Pre-commit: `make test`

---

- [x] 4. CQ-011 — Extract textinput factory helper

  **What to do**:
  - Create `newTextInput(placeholder string, width int) textinput.Model` helper function in a shared location (e.g., `internal/tui/views/form_helpers.go`)
  - Optionally create `newNumericInput(placeholder string, width, charLimit int)` for numeric fields
  - Replace ALL ~37 `textinput.New()` + placeholder/width assignments across `wizard_other.go`, `wizard_agents.go`, `wizard_categories.go`
  - Preserve exact placeholder text and width values — only extract the boilerplate

  **Must NOT do**:
  - Change any placeholder text or width values
  - Change any textinput behavior (CharLimit, Focus, etc.)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Extract factory function + mechanical replacement of 37 instantiations
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3)
  - **Blocks**: Task 5 (needs consistent patterns)
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:276-364` — 17 textinput instantiations
  - `internal/tui/views/wizard_categories.go:170-237` — 10 textinput instantiations
  - `internal/tui/views/wizard_other.go:713-887` — 10+ textinput instantiations

  **WHY Each Reference Matters**:
  - All follow the same pattern: `ti := textinput.New(); ti.Placeholder = X; ti.Width = Y`
  - Some have additional config (CharLimit) — factory should accept optional params or the caller can set them after

  **Acceptance Criteria**:
  - [ ] `internal/tui/views/form_helpers.go` exists with `newTextInput` function
  - [ ] All 37+ textinput instantiations use the factory
  - [ ] `make test` passes
  - [ ] `grep -c "textinput.New()" internal/tui/views/wizard_other.go internal/tui/views/wizard_agents.go internal/tui/views/wizard_categories.go` shows significant reduction

  **QA Scenarios:**
  ```
  Scenario: Factory replaces raw textinput.New() calls
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -c "textinput.New()" internal/tui/views/wizard_other.go internal/tui/views/wizard_agents.go internal/tui/views/wizard_categories.go
      2. Assert: each file shows 0 or near-0 (only factory itself uses textinput.New())
      3. Run: grep -c "newTextInput(" internal/tui/views/wizard_other.go internal/tui/views/wizard_agents.go internal/tui/views/wizard_categories.go
      4. Assert: sum of counts >= 37
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: All raw textinput.New() replaced with factory, tests pass
    Evidence: .sisyphus/evidence/task-4-textinput-factory.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): extract textinput factory helper`
  - Files: `internal/tui/views/form_helpers.go` (new), `wizard_other.go`, `wizard_agents.go`, `wizard_categories.go`
  - Pre-commit: `make test`

---

- [x] 5. CQ-002 — Merge duplicate Apply functions in wizard_agents

  **What to do**:
  - Merge `Apply()` (lines 808-1023) and `applyAllAgentFields()` (lines 1025-1182) into a single `applyAgentFields(cfg *config.Config, sel *profile.FieldSelection)` function
  - When `sel == nil`, apply all fields unconditionally (current `applyAllAgentFields` behavior)
  - When `sel != nil`, check field selection before writing (current `Apply` behavior)
  - Preserve the `hasSelectedFields` tracking logic in selection mode
  - Extract `refreshView()` helper: `func (w *WizardAgents) refreshView() { w.viewport.SetContent(w.renderContent()) }`
  - Replace the 52+ `w.viewport.SetContent(w.renderContent())` calls with `w.refreshView()`

  **Must NOT do**:
  - Change how field selection filtering works
  - Change the `hasSelectedFields` tracking behavior
  - Remove the `Apply` public method signature (it's called from wizard.go)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires careful merge of two similar-but-not-identical functions with behavioral preservation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 6, 7)
  - **Blocks**: None
  - **Blocked By**: Task 4 (textinput factory should be in place first)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:808-1023` — `Apply()`: selective mode with `isAgentFieldSelected()` checks per field
  - `internal/tui/views/wizard_agents.go:1025-1182` — `applyAllAgentFields()`: unconditional mode, directly assigns all 30+ fields

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go` — ~50 tests including SetConfig/Apply round-trips, fallback models, permissions

  **WHY Each Reference Matters**:
  - The field ASSIGNMENT bodies are identical between the two functions — only the guard condition differs
  - The `hasSelectedFields` flag in `Apply()` controls whether the agent entry is added to `cfg.Agents` at all
  - `applyAllAgentFields()` always adds the entry — this difference must be preserved

  **Acceptance Criteria**:
  - [ ] Single `applyAgentFields(cfg, sel)` function exists where `sel == nil` means "apply all"
  - [ ] `Apply()` delegates to `applyAgentFields` with selection
  - [ ] No `applyAllAgentFields` function exists
  - [ ] `refreshView()` helper exists and replaces all `viewport.SetContent` calls
  - [ ] `make test` passes (all 50+ wizard_agents tests)

  **QA Scenarios:**
  ```
  Scenario: Merged function handles both selection and full-apply modes
    Tool: Bash
    Preconditions: Task 4 completed
    Steps:
      1. Run: grep -c "applyAllAgentFields" internal/tui/views/wizard_agents.go
      2. Assert: output is "0" (function removed)
      3. Run: grep "applyAgentFields" internal/tui/views/wizard_agents.go
      4. Assert: function definition exists
      5. Run: grep "refreshView" internal/tui/views/wizard_agents.go
      6. Assert: helper exists and is called multiple times
      7. Run: make test
      8. Assert: exit code 0, all wizard_agents tests pass
    Expected Result: Single function, no old variant, all tests pass
    Evidence: .sisyphus/evidence/task-5-apply-merged.txt

  Scenario: Both code paths produce identical output
    Tool: Bash
    Preconditions: Merged function implemented
    Steps:
      1. Run: go test -v -run "TestWizardAgents.*Apply\|TestWizardAgents.*Checkbox\|TestWizardAgents.*Field" ./internal/tui/views/
      2. Assert: all tests pass (verifies selection mode)
      3. Run: go test -v -run "TestWizardAgents" ./internal/tui/views/ 2>&1 | grep -c "PASS"
      4. Assert: count matches baseline from Task 0
    Expected Result: Both selection and full-apply paths verified by existing tests
    Failure Indicators: Any test failure mentioning Apply, FieldSelection, or agent config
    Evidence: .sisyphus/evidence/task-5-roundtrip.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): merge duplicate Apply functions in wizard_agents`
  - Files: `internal/tui/views/wizard_agents.go`
  - Pre-commit: `make test`

---

- [x] 6. CQ-008 — Replace string error checks with typed errors

  **What to do**:
  - Define `ModelExistsError` struct in `internal/models/` package (where model CRUD lives)
  - `ModelExistsError` should contain the model ID
  - Update model registry/import code to return `ModelExistsError` instead of generic errors with "already exists" text
  - Replace all `strings.Contains(err.Error(), "already exists")` checks with `errors.As(err, &ModelExistsError{})`

  **Must NOT do**:
  - Change error message text (tests may assert on them)
  - Change any error handling behavior

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Define one error type + 2-3 replacement sites
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 7)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/model_import.go:393` — `strings.Contains(err.Error(), "already exists")`
  - `internal/tui/views/model_registry.go:443,450` — same pattern
  - `internal/models/models.go` — model CRUD where the error originates

  **WHY Each Reference Matters**:
  - The models package is the source of truth for model errors
  - All three check sites need the same typed error

  **Acceptance Criteria**:
  - [ ] `ModelExistsError` type defined in `internal/models/`
  - [ ] No `strings.Contains(err.Error(), "already exists")` in any file
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Typed error replaces string checks
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -r "already exists" internal/tui/views/ --include="*.go"
      2. Assert: no string-based error checking remains
      3. Run: grep "ModelExistsError" internal/models/models.go
      4. Assert: type definition exists
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: Typed error used everywhere, no string matching
    Evidence: .sisyphus/evidence/task-6-typed-errors.txt
  ```

  **Commit**: YES
  - Message: `refactor(models): replace string error checks with typed ModelExistsError`
  - Files: `internal/models/models.go`, `model_import.go`, `model_registry.go`
  - Pre-commit: `make test`

---

- [x] 7. CQ-007 — Extract shared ListRenderer component

  **What to do**:
  - Create `internal/tui/views/list_renderer.go` with a `ListRenderer` struct
  - Implement: `Cursor()`, `SetCursor(int)`, `Offset()`, `ScrollUp()`, `ScrollDown()`, `EnsureVisible(totalItems, visibleHeight int)`, `RenderScrollIndicator(totalItems, visibleHeight int, style) string`
  - Replace duplicated pagination logic in `model_import.go`, `model_registry.go`, `model_selector.go`
  - Each view embeds or holds a `ListRenderer` instance instead of managing cursor/offset independently

  **Must NOT do**:
  - Change any visual output of list rendering
  - Change scroll behavior or cursor movement
  - Create a new package — keep in `package views`

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Extract shared component, replace in 3 files with slightly different patterns
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6)
  - **Blocks**: Task 11 (naming consistency)
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/model_import.go:78-80` — cursor, offset, providerOffset
  - `internal/tui/views/model_registry.go:83-84` — cursor, offset
  - `internal/tui/views/model_selector.go:68-69` — cursor, scrollOffset
  - All 3 files have: `visibleHeight := m.height - 10`, boundary checks, scroll indicators

  **Test References**:
  - `internal/tui/views/model_import_test.go` — 45+ tests for state transitions and filtering
  - `internal/tui/views/model_registry_test.go` — 35+ tests for CRUD
  - `internal/tui/views/model_selector_test.go` — 30+ tests for cursor navigation

  **WHY Each Reference Matters**:
  - Each file uses slightly different variable names but identical logic
  - `model_import.go` has TWO lists (providers and models) — needs two ListRenderer instances
  - model_selector.go uses `scrollOffset` while others use `offset` — naming should standardize to `scrollOffset`

  **Acceptance Criteria**:
  - [ ] `internal/tui/views/list_renderer.go` exists with `ListRenderer` struct
  - [ ] All 3 model view files use `ListRenderer` instead of manual cursor/offset management
  - [ ] `make test` passes (all 110+ model view tests)

  **QA Scenarios:**
  ```
  Scenario: ListRenderer replaces manual pagination in all 3 files
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep "ListRenderer" internal/tui/views/model_import.go internal/tui/views/model_registry.go internal/tui/views/model_selector.go
      2. Assert: all 3 files reference ListRenderer
      3. Run: make test
      4. Assert: all model_import, model_registry, model_selector tests pass
    Expected Result: Shared component used in all 3 files, tests pass
    Evidence: .sisyphus/evidence/task-7-list-renderer.txt

  Scenario: Scroll behavior unchanged
    Tool: Bash
    Preconditions: ListRenderer integrated
    Steps:
      1. Run: go test -v -run "TestModelSelector.*Cursor\|TestModelImport.*Scroll\|TestModelRegistry.*Navigation" ./internal/tui/views/
      2. Assert: all tests pass
    Expected Result: All navigation tests pass, proving scroll behavior preserved
    Evidence: .sisyphus/evidence/task-7-scroll-behavior.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): extract shared ListRenderer component`
  - Files: `internal/tui/views/list_renderer.go` (new), `model_import.go`, `model_registry.go`, `model_selector.go`
  - Pre-commit: `make test`

---

- [x] 8. CQ-005 — Table-driven wizard step transitions

  **What to do**:
  - Define a `stepOrder` slice/table mapping step numbers to their views and transition logic
  - Replace the 4 near-identical transition cases in `nextStep()` (lines 263-377) with a loop over the table
  - Replace `prevStep()` with similar table-driven approach
  - Extract validation logic from the save closure (lines ~299-315) into a `validateConfig()` method on Wizard
  - Preserve flashMsg patterns for step transitions

  **Must NOT do**:
  - Change step order or transition behavior
  - Change the wizard step numbering
  - Remove the flash message display during transitions

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires understanding the wizard step flow and validation logic deeply
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 10, 11)
  - **Blocks**: Task 14 (interface usage depends on this)
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard.go:263-414` — `nextStep()` and `prevStep()` with 4 identical transition cases
  - `internal/tui/views/wizard.go:299-315` — validation logic duplicated in save closure

  **API/Type References**:
  - `internal/tui/views/step.go` — existing `WizardStep` interface with `Init`, `SetSize`, `View`
  - `internal/tui/views/wizard.go:263-377` — step transition pattern: Apply current → SetConfig next → Init next

  **Test References**:
  - `internal/tui/views/wizard_test.go` — 15 tests for create/edit flows, profile rename

  **WHY Each Reference Matters**:
  - The step.go interface already defines partial step contract — extend with table-driven approach
  - Each transition follows: `currentStep.Apply(cfg, sel) → nextStep.SetConfig(cfg, sel) → nextStep.Init()`
  - Validation in save closure should be reusable outside closure context

  **Acceptance Criteria**:
  - [ ] `stepOrder` table/slice defined in wizard.go
  - [ ] `nextStep()` uses table-driven loop (not individual switch cases)
  - [ ] `prevStep()` uses table-driven approach
  - [ ] `validateConfig()` method extracted from save closure
  - [ ] `make test` passes (all wizard_test.go tests)

  **QA Scenarios:**
  ```
  Scenario: Table-driven transitions replace switch cases
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -c "case StepCategories\|case StepAgents\|case StepHooks\|case StepOther" internal/tui/views/wizard.go
      2. Assert: significantly reduced from baseline (4 per function → 0 in ideal)
      3. Run: grep "stepOrder\|stepTransition" internal/tui/views/wizard.go
      4. Assert: table definition exists
      5. Run: make test
      6. Assert: all wizard tests pass
    Expected Result: Table-driven transitions, no individual step cases
    Evidence: .sisyphus/evidence/task-8-table-driven.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): table-driven wizard step transitions`
  - Files: `internal/tui/views/wizard.go`
  - Pre-commit: `make test`

---

- [x] 9. CQ-003 — Extract wizard_agents Update handlers

  **What to do**:
  - Extract `handleFormNavigation()` from `wizard_agents.go Update()` for the main form key handling
  - Ensure existing sub-mode handlers (`handleModelSelection`, `handleSaveCustomModel`, `handleProviderOptsEditor`, `handleFallbackModelsEditor`, `handleBashPermsEditor`) are called via early-return pattern
  - The `Update()` function should become a thin dispatcher:
    ```go
    func (w *WizardAgents) Update(msg tea.Msg) (WizardAgents, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            ac := w.agents[w.currentAgent()]
            if ac.selectingModel    { return w.handleModelSelection(ac, msg) }
            if ac.savingCustomModel { return w.handleSaveCustomModel(ac, msg) }
            if ac.editingProviderOpts   { return w.handleProviderOptsEditor(ac, msg) }
            if ac.editingFallbackModels { return w.handleFallbackModelsEditor(ac, msg) }
            if ac.editingBashPerms      { return w.handleBashPermsEditor(ac, msg) }
            return w.handleFormNavigation(ac, msg)
        case tea.WindowSizeMsg: ...
        }
    }
    ```
  - Extract the main form key handling (~200 lines of switch cases) into `handleFormNavigation`

  **Must NOT do**:
  - Reorder sub-mode checks (model selection MUST check before custom model save)
  - Change any keyboard behavior
  - Change the Update return signature (Bubble Tea pattern)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Careful extraction from a complex Update function with 7-level nesting
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 10, 11)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:1615-1927` — `Update()` function at 312 lines with 7-level nesting
  - Already extracted handlers: `handleFallbackModelsEditor`, `handleProviderOptsEditor`, `handleBashPermsEditor`, `handleSaveCustomModel`

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go` — ~50 tests including Update behavior

  **WHY Each Reference Matters**:
  - The main form navigation (~200 lines) handles field focus, cursor movement, section switching
  - Sub-mode handlers are already extracted — only the dispatch and form navigation remain monolithic
  - The early-return pattern is already partially used for sub-modes

  **Acceptance Criteria**:
  - [ ] `handleFormNavigation()` method exists on WizardAgents
  - [ ] `Update()` is a thin dispatcher (< 50 lines)
  - [ ] All sub-mode handlers called via early-return
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Update() reduced to thin dispatcher
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep -c "func.*handleFormNavigation" internal/tui/views/wizard_agents.go
      2. Assert: count is 1 (method exists)
      3. Run: wc -l internal/tui/views/wizard_agents.go
      4. Assert: reduced from 2717 baseline
      5. Run: make test
      6. Assert: all wizard_agents tests pass
    Expected Result: Handler extracted, Update simplified, tests pass
    Evidence: .sisyphus/evidence/task-9-handlers-extracted.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): extract wizard_agents Update handlers`
  - Files: `internal/tui/views/wizard_agents.go`
  - Pre-commit: `make test`

---

- [x] 10. CQ-004 — Extract app.go message handlers

  **What to do**:
  - Extract navigation handler methods (`handleNavToWizard`, `handleNavToDiff`, etc.) from Update()
  - Each follows the same 3-step pattern: create view, SetSize, navigateTo
  - Extract view delegation into a helper method using the existing interface pattern
  - Consolidate error toast pattern (`showToast("...failed: "+err.Error(), toastError, 3*time.Second)`) into `showErrorToast(err)` helper

  **Must NOT do**:
  - Change message routing order
  - Change toast behavior or messages
  - Change navigation flow

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Refactoring central router with 10+ state delegations
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 11)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/app.go:141-490` — `Update()` at 350 lines, 20+ message types, 10 state delegations
  - `internal/tui/app.go:220-373` — Navigation handlers repeating create/SetSize/navigateTo pattern ~6 times
  - `internal/tui/app.go:447-487` — View delegation switch with 10 identical cases

  **Test References**:
  - `internal/tui/app_test.go` — 15 tests for state transitions, navigation, toast system

  **WHY Each Reference Matters**:
  - Navigation pattern is always: `a.view = NewView(); a.view.SetSize(w, h); a.navigateTo(stateView)`
  - View delegation pattern is always: `a.view, cmd = a.view.Update(msg); cmds = append(cmds, cmd)`
  - These can be replaced with helper methods and/or interface dispatch

  **Acceptance Criteria**:
  - [ ] Navigation helper methods extracted
  - [ ] View delegation simplified (not 10 identical cases)
  - [ ] `showErrorToast` helper exists
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: app.go Update simplified
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: wc -l internal/tui/app.go
      2. Assert: reduced from 878 baseline
      3. Run: grep -c "showErrorToast\|handleNav" internal/tui/app.go
      4. Assert: extracted helpers exist
      5. Run: make test
      6. Assert: all app tests pass
    Expected Result: Update simplified, helpers extracted, tests pass
    Evidence: .sisyphus/evidence/task-10-app-extracted.txt
  ```

  **Commit**: YES
  - Message: `refactor(tui): extract app.go message handlers`
  - Files: `internal/tui/app.go`
  - Pre-commit: `make test`

---

- [x] 11. CQ-010 — Standardize scroll variable naming

  **What to do**:
  - After CQ-007 (ListRenderer) is complete, verify all remaining scroll-related variables use consistent naming
  - If any files still have manual scroll variables outside ListRenderer, rename to `scrollOffset`
  - Check that `model_import.go` no longer has separate `providerOffset` (should use a second ListRenderer instance)

  **Must NOT do**:
  - Change scroll behavior
  - Introduce new variables

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Naming standardization, no logic changes
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (but must wait for Task 7)
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10) — but only starts after Task 7 completes
  - **Blocks**: None
  - **Blocked By**: Task 7

  **References**:

  **Pattern References**:
  - `internal/tui/views/model_import.go` — uses `offset`, `providerOffset`
  - `internal/tui/views/model_registry.go` — uses `offset`
  - `internal/tui/views/model_selector.go` — uses `scrollOffset`

  **WHY Each Reference Matters**:
  - After ListRenderer extraction, most manual variables should be gone
  - Any remaining should be standardized

  **Acceptance Criteria**:
  - [ ] All scroll variables use consistent naming (`scrollOffset` or managed by `ListRenderer`)
  - [ ] No file uses both `offset` and `scrollOffset` for the same concept
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Consistent scroll naming across model views
    Tool: Bash
    Preconditions: Task 7 completed
    Steps:
      1. Run: grep -n "offset\b" internal/tui/views/model_import.go internal/tui/views/model_registry.go internal/tui/views/model_selector.go
      2. Assert: any remaining manual offset variables use consistent naming
      3. Run: make test
      4. Assert: exit code 0
    Expected Result: Consistent naming, tests pass
    Evidence: .sisyphus/evidence/task-11-naming-standardized.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): standardize scroll variable naming`
  - Files: model view files
  - Pre-commit: `make test`

---

- [x] 12. CQ-001 — Split wizard_other.go into 5 files (3-phase)

  **What to do**:
  This is the largest refactoring item. It MUST be done in 3 separate phases/commits.

  **PHASE 1: Pure file split (NO code changes)**
  Split `wizard_other.go` (3,768 lines) into 5 files by moving functions as-is:

  | New File | Content | Target Lines |
  |----------|---------|-------------|
  | `wizard_other.go` | Struct definition, `WizardOther` type, `NewWizardOther()`, `Init()`, `View()`, constants, section data tables | ~400 |
  | `wizard_other_config.go` | `SetConfig()`, `Apply()` | ~900 |
  | `wizard_other_update.go` | `Update()`, `toggleSection()`, `toggleSubItem()` | ~1,600 |
  | `wizard_other_render.go` | `renderContent()`, `renderSubSection()` | ~260 |
  | `wizard_other_fields.go` | Field path constants, `*HasData()` helpers, `parseMapStringInt`, `serializeMapStringInt`, other helpers | ~600 |

  **Structural map of wizard_other.go (for split guidance)**:
  ```
  Lines    1-35:    Package/color/style vars
  Lines   37-75:   parseMapStringInt/serializeMapStringInt helpers
  Lines   77-116:  Data tables (disableableAgents, etc.)
  Lines  119-172:  Field path constants
  Lines  176-252:  otherSectionNames map + key bindings
  Lines  255-315:  Small helper functions
  Lines  317-560:  Method-like helpers (fieldSelected, hasData, etc.)
  Lines  569-934:  WizardOther struct + NewWizardOther constructor
  Lines  936-939:  Init()
  Lines  940-1000: SetSize()
  Lines 1002-1402: SetConfig()
  Lines 1403-1880: Apply()
  Lines 1881-3302: Update() ← THE BEAST
  Lines 3304-3341: toggleSection()
  Lines 3343-3485: toggleSubItem()
  Lines 3487-3546: renderContent()
  Lines 3547-3740: renderSubSection()
  Lines 3741-3768: View()
  ```

  **PHASE 2: Data-driven textinput dispatch**
  Replace the ~30 near-identical textinput handler blocks in `Update()` (~750 lines of boilerplate) with:
  ```go
  type fieldBinding struct {
      section  otherSection
      cursor   int
      input    *textinput.Model
  }
  // Single dispatch table + generic handler function (~30 lines + data table)
  ```
  **CRITICAL**: Some blocks have special cases:
  - `sectionOpenclaw` cursor 4 has different behavior
  - `sectionRuntimeFallback` cursor 1 has different behavior
  - `sectionSkillsJson` cursor 1 has different behavior
  These MUST be handled by the dispatch table with per-entry overrides.

  **PHASE 3: Extract refreshView helper**
  - Create `func (w *WizardOther) refreshView() { w.viewport.SetContent(w.renderContent()) }`
  - Replace all 160 `w.viewport.SetContent(w.renderContent())` calls with `w.refreshView()`

  **Must NOT do**:
  - Mix phases in one commit (each phase = separate commit)
  - Change any behavior during Phase 1 (pure file split only)
  - Miss edge cases in dispatch table (Phase 2)
  - Change any keyboard behavior or field handling

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Largest refactoring item, requires careful 3-phase execution with verification at each step
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (sequential phases)
  - **Parallel Group**: Wave 4
  - **Blocks**: Task 13
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_other.go` — THE ENTIRE FILE (3,768 lines)
  - Lines 1881-3302 — `Update()` with 30 textinput handler blocks
  - Lines 3304-3341 — `toggleSection()`
  - Lines 3343-3485 — `toggleSubItem()`

  **Test References**:
  - `internal/tui/views/wizard_other_test.go` — 1079 lines, 60+ tests covering:
    - Field inclusion/exclusion toggles
    - Boolean field behavior
    - Disabled lists (5 sections × 3 states)
    - SetConfig/Apply round-trips
    - JSON serialization

  **WHY Each Reference Matters**:
  - Phase 1 is purely moving code between files in same package — tests don't need changes
  - Phase 2 is the highest-risk change — the 30 handler blocks have subtle differences
  - Phase 3 is purely mechanical — refreshView extraction

  **Acceptance Criteria**:
  - [ ] Phase 1: `wizard_other*.go` exists as 5 files, no single file > 1000 lines
  - [ ] Phase 1: `make test` passes with IDENTICAL output to pre-split
  - [ ] Phase 2: `fieldBinding` dispatch table exists
  - [ ] Phase 2: `wizard_other_update.go` reduced by ~700+ lines
  - [ ] Phase 2: `make test` passes
  - [ ] Phase 3: `refreshView()` exists and replaces all 160 SetContent calls
  - [ ] Phase 3: `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Phase 1 — Pure split preserves test output
    Tool: Bash
    Preconditions: Task 0 baseline recorded
    Steps:
      1. Run: ls internal/tui/views/wizard_other*.go
      2. Assert: 5 files exist (wizard_other.go, wizard_other_config.go, wizard_other_update.go, wizard_other_render.go, wizard_other_fields.go)
      3. Run: wc -l internal/tui/views/wizard_other*.go
      4. Assert: no single file > 1000 lines
      5. Run: make test
      6. Assert: exit code 0, same number of tests as baseline
    Expected Result: 5 files, all < 1000 lines, all tests pass
    Failure Indicators: Any test failure, any file > 1000 lines
    Evidence: .sisyphus/evidence/task-12-phase1-split.txt

  Scenario: Phase 2 — Dispatch table replaces boilerplate
    Tool: Bash
    Preconditions: Phase 1 complete
    Steps:
      1. Run: grep -c "fieldBinding" internal/tui/views/wizard_other_update.go
      2. Assert: dispatch table exists (count > 0)
      3. Run: wc -l internal/tui/views/wizard_other_update.go
      4. Assert: significantly reduced from Phase 1 size
      5. Run: go test -v -run "TestWizardOther.*Update\|TestWizardOther.*Inclusion\|TestWizardOther.*Boolean" ./internal/tui/views/
      6. Assert: all tests pass
    Expected Result: Dispatch table implemented, tests pass
    Failure Indicators: Any test failure in wizard_other tests
    Evidence: .sisyphus/evidence/task-12-phase2-dispatch.txt

  Scenario: Phase 3 — refreshView helper extracted
    Tool: Bash
    Preconditions: Phase 2 complete
    Steps:
      1. Run: grep -c "func.*refreshView" internal/tui/views/wizard_other*.go
      2. Assert: helper exists
      3. Run: grep -c "viewport.SetContent(w.renderContent())" internal/tui/views/wizard_other*.go
      4. Assert: significantly reduced (ideally 0, or very few for special cases)
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: refreshView used everywhere, tests pass
    Evidence: .sisyphus/evidence/task-12-phase3-refresh.txt
  ```

  **Commit**: YES (3 commits)
  - Commit 12a: `refactor(views): split wizard_other.go into multiple files (pure split)`
  - Commit 12b: `refactor(views): data-driven textinput dispatch in wizard_other`
  - Commit 12c: `refactor(views): extract refreshView helper in wizard_other`
  - Pre-commit (each): `make test`

---

- [x] 13. CQ-006 — Split wizard_categories.go into 2 files

  **What to do**:
  Split `wizard_categories.go` (1,345 lines) into:
  | New File | Content | Target Lines |
  |----------|---------|-------------|
  | `wizard_categories.go` | Core logic, struct, constructor, Init, View, Update, render methods | ~900 |
  | `wizard_categories_config.go` | SetConfig, Apply, selection paths | ~450 |

  Follow the same split pattern as CQ-001 Phase 1 for consistency.

  **Must NOT do**:
  - Change any behavior (pure file split)
  - Create new packages

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: File split with careful boundary identification
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Task 12 patterns being established)
  - **Parallel Group**: Wave 4 (after Task 12)
  - **Blocks**: None
  - **Blocked By**: Task 12 (should follow same patterns)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_categories.go` — 1,345 lines total
  - Use the same split pattern as Task 12 Phase 1 (wizard_other_config.go for SetConfig/Apply)

  **Test References**:
  - `internal/tui/views/wizard_categories_test.go` — 1160 lines, 45+ tests

  **WHY Each Reference Matters**:
  - Split boundary should separate config management (SetConfig/Apply) from UI logic (Update/View/Render)
  - Following the same pattern as CQ-001 ensures consistency

  **Acceptance Criteria**:
  - [ ] `wizard_categories.go` and `wizard_categories_config.go` exist
  - [ ] Neither file > 1000 lines
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Categories split preserves behavior
    Tool: Bash
    Preconditions: Task 12 completed
    Steps:
      1. Run: ls internal/tui/views/wizard_categories*.go
      2. Assert: 2 files exist
      3. Run: wc -l internal/tui/views/wizard_categories*.go
      4. Assert: no file > 1000 lines
      5. Run: make test
      6. Assert: all wizard_categories tests pass
    Expected Result: 2 files, both < 1000 lines, all tests pass
    Evidence: .sisyphus/evidence/task-13-categories-split.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): split wizard_categories.go into multiple files`
  - Files: `internal/tui/views/wizard_categories.go`, `internal/tui/views/wizard_categories_config.go` (new)
  - Pre-commit: `make test`

---

- [x] 14. CQ-012 — Use WizardStep interface in wizard.go

  **What to do**:
  - `step.go` already defines a `WizardStep` interface with `Init`, `SetSize`, `View`
  - The real work is using this interface in wizard.go's table-driven step management (built in Task 8)
  - Verify the step transition table from Task 8 uses the interface properly
  - Add compile-time interface checks if not present: `var _ WizardStep = (*WizardOther)(nil)` etc.
  - Do NOT expand the interface to include `SetConfig/Apply` — they have inconsistent signatures across steps

  **Must NOT do**:
  - Expand the WizardStep interface beyond existing methods
  - Force `SetConfig/Apply` into the interface (different signatures per step)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Interface integration with table-driven approach
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with Tasks 15, 16, 17)
  - **Blocks**: None
  - **Blocked By**: Task 8 (table-driven step transitions must exist first)

  **References**:

  **Pattern References**:
  - `internal/tui/views/step.go` — existing `WizardStep` interface
  - `internal/tui/views/wizard.go` — table-driven approach from Task 8

  **WHY Each Reference Matters**:
  - step.go already has the interface — just needs to be used consistently
  - NameStep and ReviewStep don't have SetConfig/Apply — that's why they stay out of the interface

  **Acceptance Criteria**:
  - [ ] WizardStep interface used in step transition table
  - [ ] Compile-time checks exist for all step types
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Interface properly used
    Tool: Bash
    Preconditions: Task 8 completed
    Steps:
      1. Run: grep "WizardStep" internal/tui/views/wizard.go
      2. Assert: interface is referenced in transition logic
      3. Run: grep "var _.*WizardStep" internal/tui/views/*.go
      4. Assert: compile-time checks exist
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: Interface integrated, compile-time checks present
    Evidence: .sisyphus/evidence/task-14-interface.txt
  ```

  **Commit**: YES
  - Message: `refactor(views): use WizardStep interface in table-driven wizard.go`
  - Files: `internal/tui/views/wizard.go`, possibly `step.go`
  - Pre-commit: `make test`

---

- [x] 15. CQ-013 — Consolidate help rendering in app.go

  **What to do**:
  - Consolidate `renderShortHelp()` (47 lines) and `renderFullHelp()` (86 lines) into a state-to-hints map
  - Create a `helpData` map with `appState` keys mapping to hint strings
  - Single rendering function that looks up hints by state

  **Must NOT do**:
  - Change help text content
  - Change help display behavior

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Consolidate two similar functions into a map-driven approach
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with Tasks 14, 16, 17)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/tui/app.go` — `renderShortHelp()` (47 lines) and `renderFullHelp()` (86 lines)
  - Both share state-specific hint definitions that could be a single data structure

  **Acceptance Criteria**:
  - [ ] State-to-hints map exists
  - [ ] `renderShortHelp` and `renderFullHelp` use the map
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Help rendering consolidated
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: grep "helpData\|helpHints" internal/tui/app.go
      2. Assert: data structure exists
      3. Run: make test
      4. Assert: all app tests pass (help rendering unchanged)
    Expected Result: Map-driven help, tests pass
    Evidence: .sisyphus/evidence/task-15-help-consolidated.txt
  ```

  **Commit**: YES
  - Message: `refactor(tui): consolidate help rendering in app.go`
  - Files: `internal/tui/app.go`
  - Pre-commit: `make test`

---

- [x] 16. CQ-009 — Consolidate hardcoded field paths

  **What to do**:
  - Move `allFieldPaths` (156 entries) from `internal/profile/selection.go` into a dedicated file `internal/profile/fields.go`
  - Move `knownConfigTags`, `knownFieldPaths`, `knownFieldPathPrefixes` from `internal/profile/profile.go` into the same file
  - Add clear documentation that this is the single source of truth for field paths
  - Keep the derived maps (`knownFieldPaths`, `knownFieldPathPrefixes`) as they're performance optimizations
  - Do NOT write a go:generate tool — just consolidate into one well-documented file

  **Must NOT do**:
  - Write a code generator (future work)
  - Remove the derived lookup maps (performance optimization)
  - Change any field path values

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Moving 156+ string constants between files with careful documentation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with Tasks 14, 15, 17)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/profile/selection.go:8-164` — `allFieldPaths` with 156 hardcoded entries
  - `internal/profile/profile.go:24-77` — `knownConfigTags`, `knownFieldPaths`, `knownFieldPathPrefixes`

  **WHY Each Reference Matters**:
  - Currently adding a new config field requires updating multiple lists across different files
  - Consolidating to one file means one place to update

  **Acceptance Criteria**:
  - [ ] `internal/profile/fields.go` exists with all field path definitions
  - [ ] Clear documentation comments on the file
  - [ ] `knownConfigTags`, `knownFieldPaths`, `knownFieldPathPrefixes` remain functional
  - [ ] `make test` passes

  **QA Scenarios:**
  ```
  Scenario: Field paths consolidated to single file
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: ls internal/profile/fields.go
      2. Assert: file exists
      3. Run: grep -c "FieldPath\|fieldPath" internal/profile/fields.go
      4. Assert: contains the field path definitions (156+ entries)
      5. Run: make test
      6. Assert: all profile tests pass
    Expected Result: Single source file, all tests pass
    Evidence: .sisyphus/evidence/task-16-field-paths.txt
  ```

  **Commit**: YES
  - Message: `refactor(profile): consolidate hardcoded field paths into single file`
  - Files: `internal/profile/fields.go` (new), `selection.go`, `profile.go`
  - Pre-commit: `make test`

---

- [x] 17. CQ-014 — Split config/types.go by domain

  **What to do**:
  - Split `internal/config/types.go` (325 lines, 31 nested types) into domain-grouped files:
    - `types.go` — root `Config` struct + shared types
    - `types_agent.go` — `AgentConfig` and agent-related types
    - `types_experimental.go` — `ExperimentalConfig` and sub-types
    - `types_background_task.go` — `BackgroundTaskConfig` and sub-types
  - Pure file split — no code changes, just reorganize type definitions

  **Must NOT do**:
  - Change any type definitions
  - Change any struct tags
  - Change any field names

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Pure file split, no logic changes
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with Tasks 14, 15, 16)
  - **Blocks**: None
  - **Blocked By**: Task 0

  **References**:

  **Pattern References**:
  - `internal/config/types.go` — 325 lines with 31 nested types
  - **CRITICAL**: Must match upstream JSON schema 1:1 (noted in AGENTS.md anti-patterns)

  **Test References**:
  - `internal/config/types_test.go` — 1200 lines, 50+ round-trip serialization tests

  **WHY Each Reference Matters**:
  - AGENTS.md warns: "Never add Config fields without upstream schema support"
  - The split is purely organizational — same types, same tags, just different files

  **Acceptance Criteria**:
  - [ ] `types_agent.go`, `types_experimental.go`, `types_background_task.go` exist
  - [ ] Root `types.go` contains `Config` struct + shared types
  - [ ] `make test` passes (all 50+ round-trip tests)

  **QA Scenarios:**
  ```
  Scenario: Config types split preserves serialization
    Tool: Bash
    Preconditions: Task 0 completed
    Steps:
      1. Run: ls internal/config/types*.go
      2. Assert: multiple type files exist
      3. Run: go test -v ./internal/config/ 2>&1 | grep -c "PASS"
      4. Assert: all round-trip tests pass
      5. Run: make test
      6. Assert: exit code 0
    Expected Result: Types split across files, all serialization tests pass
    Evidence: .sisyphus/evidence/task-17-types-split.txt
  ```

  **Commit**: YES
  - Message: `refactor(config): split types.go by domain`
  - Files: `internal/config/types.go`, `internal/config/types_agent.go` (new), `internal/config/types_experimental.go` (new), `internal/config/types_background_task.go` (new)
  - Pre-commit: `make test`

---

## Final Verification Wave

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns. Check evidence files exist in `.sisyphus/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `make lint` + `make test`. Review all changed files for: `as any`/type assertion shortcuts, empty catches, commented-out code, unused imports. Check no AI slop introduced. Verify file sizes reduced (wizard_other no longer 3700+ lines).
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Build + Test Verification** — `unspecified-high`
  Start from clean state (`go build ./...` from scratch). Run `make test` — all tests must pass. Run `go vet ./...`. Compare test count and names against pre-refactor baseline to ensure no tests were lost.
  Output: `Build [PASS/FAIL] | Tests [N/N pass] | Vet [PASS/FAIL] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec. Check "Must NOT do" compliance. Verify no behavioral changes by comparing test baselines. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Behavior [PRESERVED/CHANGED] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| Commit # | Tasks | Message |
|----------|-------|---------|
| 0 | T0 | `chore: tag pre-refactor-baseline` |
| 1 | T1 | `refactor(views): consolidate triplicated expandPath into pathutil.go` |
| 2 | T2 | `refactor(views): remove view-specific color vars, use centralized styles` |
| 3 | T3 | `refactor(views): fix inconsistent receiver types in wizard_agents` |
| 4 | T4 | `refactor(views): extract textinput factory helper` |
| 5 | T5 | `refactor(views): merge duplicate Apply functions in wizard_agents` |
| 6 | T6 | `refactor(models): replace string error checks with typed ModelExistsError` |
| 7 | T7 | `refactor(views): extract shared ListRenderer component` |
| 8 | T8 | `refactor(views): table-driven wizard step transitions` |
| 9 | T9 | `refactor(views): extract wizard_agents Update handlers` |
| 10 | T10 | `refactor(tui): extract app.go message handlers` |
| 11 | T11 | `refactor(views): standardize scroll variable naming` |
| 12a | T12-P1 | `refactor(views): split wizard_other.go into multiple files (pure split)` |
| 12b | T12-P2 | `refactor(views): data-driven textinput dispatch in wizard_other` |
| 12c | T12-P3 | `refactor(views): extract refreshView helper in wizard_other` |
| 13 | T13 | `refactor(views): split wizard_categories.go into multiple files` |
| 14 | T14 | `refactor(views): use WizardStep interface in table-driven wizard.go` |
| 15 | T15 | `refactor(tui): consolidate help rendering in app.go` |
| 16 | T16 | `refactor(profile): consolidate hardcoded field paths into single file` |
| 17 | T17 | `refactor(config): split types.go by domain` |

---

## Success Criteria

### Verification Commands
```bash
make test       # Expected: all tests pass, 0 failures
make lint       # Expected: 0 errors
make build      # Expected: successful build
go vet ./...    # Expected: 0 issues
wc -l internal/tui/views/wizard_other*.go  # Expected: no single file > 1000 lines
```

### Final Checklist
- [ ] All 17 CQ items implemented
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass (`make test`)
- [ ] Lint passes (`make lint`)
- [ ] No behavioral changes
- [ ] wizard_other.go no longer a single 3700+ line file
- [ ] No new packages created
- [ ] No duplication from the 5 excluded patterns introduced
