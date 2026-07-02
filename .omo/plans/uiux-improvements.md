# UI/UX Improvements — 30 Issues from IDEATION_UI_UX.md

## TL;DR

> **Quick Summary**: Implement all 30 UI/UX improvements across the omo-profiler TUI — fix silent failures, add validation feedback, standardize help text, improve accessibility, and polish visual consistency.
> 
> **Deliverables**:
> - Error/warning displays in wizard and app-level views
> - Real-time and on-exit validation indicators in wizard forms
> - Consistent help text format across all views
> - Accessible status indicators beyond color-only distinctions
> - Shared confirmation dialog component
> - Loading context messages per-operation
> - Progress indicator in wizard header
> - Visual polish (pane labels, format hints, scroll indicators, etc.)
> 
> **Estimated Effort**: Medium (~200-300 lines added, ~50 modified across 17 files)
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Task 1 (shared styles) → Task 2 (app.go) → Tasks 3-8 (parallel views) → Task 9 (wizard) → Tasks 10-12 (medium views) → Tasks 13-16 (low views)

---

## Context

### Original Request
Implement all 30 UI/UX improvements described in `IDEATION_UI_UX.md` — 7 High, 13 Medium, 10 Low priority issues spanning usability, accessibility, visual polish, and interaction design.

### Interview Summary
**Key Discussions**:
- **Scope**: All 30 issues, no subset
- **Commits**: 3 commits grouped by priority (high, medium, low)
- **Tests**: No automated tests — validation via agent QA (tmux TUI interaction)
- **UIUX-003 Back Nav**: Auto-apply changes before navigating back (no confirmation dialog)
- **UIUX-004 Validation**: Simple fields validate on-keystroke; range fields (temperature, top_p) validate on field-exit
- **UIUX-011 Shared Dialog**: Add `RenderConfirmDialog()` to `styles.go` (no new files)
- **UIUX-014**: Scope reduced — retry already exists, only improve error visual weight
- **UIUX-023**: Scope reduced — placeholder already exists, only update to show provider/model format

**Research Findings**:
- `wizard_name.go` has ✓/✗ validation pattern (lines 137-144) — replicate across forms
- `styles.go` has 8 colors + 13 styles but no shared confirmation component
- Each wizard step duplicates colors locally — new code MUST use `styles.go` imports
- `schema_check.go` already has `[r] retry [esc] back` — UIUX-014 only needs visual improvement
- `model_selector.go` already has `"e.g., gpt-4o-mini"` placeholder — UIUX-023 only needs format update
- Delete confirmation pattern identical in `list.go` and `model_registry.go` with minor style differences
- Toast system exists at app level (`app.go:622-626`) — wizard-level feedback uses `flashMsg` field pattern

### Metis Review
**Identified Gaps** (addressed):
- Two issues partially outdated (UIUX-014, UIUX-023) — scope reduced appropriately
- UIUX-003 needs behavioral decision — resolved: auto-apply before back navigation
- UIUX-004 validation timing decision — resolved: on field-exit for range fields
- UIUX-011 shared component location — resolved: styles.go function
- Help text format standardization — resolved: use bracketed format `[key] action`
- Risk: `wizard_agents.go` is 2699 lines — mitigated by targeting specific line ranges

---

## Work Objectives

### Core Objective
Fix 30 UI/UX friction points across the TUI: silent failures, missing validation feedback, inconsistent help text, accessibility gaps, and visual polish.

### Concrete Deliverables
- Error display in `wizard.go` View() when `w.err != nil`
- Flash message in `wizard_review.go` when save attempted with invalid data
- Auto-apply + flash in `wizard.go` prevStep() when navigating back
- Real-time validation indicators in `wizard_agents.go` and `wizard_categories.go` forms
- Operation-specific loading messages in `app.go`
- Text labels on diff view panes (beyond color)
- "q" key hint in wizard mode
- "Step X of Y" progress in wizard header
- Standardized help text across all views (bracketed format)
- Shared `RenderConfirmDialog()` in `styles.go`
- Status labels for hooks (`[✓ enabled]` / `[✗ disabled]`)
- Visual toggle indicator for "Include disabled_hooks"
- Context category label during form editing
- Resolved export path notification
- Improved error visual weight in schema_check and model_registry
- Better empty state guidance in template_select
- Compact mode indicator
- Help truncation indicator ("…" when truncated)
- All other minor polish items from the Low priority group

### Definition of Done
- [x] `make build` succeeds with no errors
- [x] `make test` passes all existing tests (36 test files)
- [x] `make lint` passes with no new warnings
- [x] All 30 issues verified via tmux QA scenarios
- [x] 3 commits created (high, medium, low priority)

### Must Have
- Error display visible to user when wizard validation fails (UIUX-001)
- Flash message when save rejected in review step (UIUX-002)
- Some feedback when navigating back in wizard (UIUX-003)
- Validation indicators on agent/category form fields (UIUX-004)
- Operation-specific loading text (UIUX-005)
- Non-color focus indicators on diff panes (UIUX-006)
- Hint when "q" pressed in wizard (UIUX-007)
- All other 23 issues implemented as described

### Must NOT Have (Guardrails)
- **No new files** except the function added to `styles.go`
- **No new go.mod dependencies**
- **No changes to Config struct, profile package, schema package, or backup package**
- **No new appState values in the enum**
- **No changes to WizardStep interface**
- **No local color/style definitions** — always import from `styles.go`
- **No testdata fixture changes**
- **No business logic changes** — only View/Update rendering changes
- **No JSDoc/doc comments on functions that don't have them**
- **No feature flags or configuration options for UI elements**
- **No i18n/localization infrastructure**
- **No refactoring of unrelated code** (e.g., existing local style anti-patterns in wizard steps)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.
> Acceptance criteria requiring "user manually tests/confirms" are FORBIDDEN.

### Test Decision
- **Infrastructure exists**: YES (Go testing + testify, 36 test files)
- **Automated tests**: None — these are UI/UX cosmetic changes
- **Framework**: N/A for this work
- **Existing tests must pass**: YES — `make test` after each commit is MANDATORY

### QA Policy
Every task includes agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **TUI Verification**: Use `interactive_bash` (tmux) — build binary, launch, send keystrokes, verify output
- **Build Verification**: `make build` + `make test` + `make lint` after each task wave
- **Visual Verification**: tmux screenshot for key scenarios

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — shared utilities):
├── Task 1: Shared styles + confirmation dialog in styles.go [quick]

Wave 2 (High Priority — 7 issues, MAX PARALLEL):
├── Task 2: app.go fixes — loading context + "q" hint + help truncation (UIUX-005,007,018) [quick]
├── Task 3: wizard.go error display + progress indicator (UIUX-001,008) [quick]
├── Task 4: wizard_review.go flash message on invalid save (UIUX-002) [quick]
├── Task 5: diff.go pane labels beyond color (UIUX-006) [quick]
├── Task 6: wizard.go back-nav auto-apply + flash (UIUX-003) [quick]
├── Task 7: wizard_agents.go + wizard_categories.go validation indicators (UIUX-004) [unspecified-high]

Wave 3 (Medium Priority — 13 issues, MAX PARALLEL):
├── Task 8: Confirmation dialog refactor — list.go + model_registry.go (UIUX-011) [quick]
├── Task 9: wizard_name.go help fixes (UIUX-010,019) [quick]
├── Task 10: wizard help text standardization — categories, agents, other, hooks (UIUX-009,020) [quick]
├── Task 11: wizard_hooks.go status labels + toggle visual (UIUX-013,016,022) [quick]
├── Task 12: wizard_agents.go simplified help (UIUX-017) [quick]
├── Task 13: export.go resolved path notification (UIUX-012) [quick]
├── Task 14: schema_check.go error visual weight (UIUX-014) + template_select.go empty state (UIUX-015) [quick]

Wave 4 (Low Priority — 10 issues, MAX PARALLEL):
├── Task 15: Label clarity + format hints (UIUX-021,023,024,025) [quick]
├── Task 16: Visual polish — viewport constants + compact indicator + error weight (UIUX-026,029,030) [quick]
├── Task 17: Help text format standardization across all views (UIUX-027) [quick]
├── Task 18: Micro-feedback on field completion across wizard steps (UIUX-028) [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real QA via tmux (unspecified-high)
├── Task F4: Scope fidelity check (deep)
→ Present results → Get explicit user okay
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | - | 2,8 | 1 |
| 2 | 1 | - | 2 |
| 3 | - | 6 | 2 |
| 4 | - | - | 2 |
| 5 | - | - | 2 |
| 6 | 3 | - | 2 |
| 7 | - | 18 | 2 |
| 8 | 1 | - | 3 |
| 9 | - | - | 3 |
| 10 | - | 17 | 3 |
| 11 | - | - | 3 |
| 12 | - | - | 3 |
| 13 | - | - | 3 |
| 14 | - | - | 3 |
| 15 | - | - | 4 |
| 16 | - | - | 4 |
| 17 | 10 | - | 4 |
| 18 | 7 | - | 4 |

### Agent Dispatch Summary

- **Wave 1**: **1** — T1 → `quick`
- **Wave 2**: **6** — T2,T3,T4,T5,T6 → `quick`, T7 → `unspecified-high`
- **Wave 3**: **7** — T8,T9,T10,T11,T12,T13,T14 → `quick`
- **Wave 4**: **4** — T15,T16,T17 → `quick`, T18 → `unspecified-high`
- **FINAL**: **4** — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Shared Styles + Confirmation Dialog Helper

  **What to do**:
  - Add `RenderConfirmDialog(target, message string) string` function to `internal/tui/styles.go`
  - This function renders a styled confirmation dialog using the same yellow-on-gray pattern currently duplicated in `list.go` and `model_registry.go`
  - Use existing `Yellow` and `Gray` colors from styles.go — no new color definitions
  - Also add `ErrorIconStyle` (Red + "⚠ " prefix) for reuse across error displays (schema_check, model_registry, wizard)

  **Must NOT do**:
  - Do not create new files
  - Do not add new colors beyond what's in styles.go
  - Do not over-engineer — this is a single function, not a component framework

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple function addition to existing file, <20 lines
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `frontend-ui-ux`: Not frontend web — this is terminal UI

  **Parallelization**:
  - **Can Run In Parallel**: NO — foundation task
  - **Parallel Group**: Wave 1 (solo)
  - **Blocks**: Tasks 2, 8
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/list.go:272-279` — Current confirmation dialog style (yellow bold on gray background, `[y/n]` format). Extract this exact pattern into the shared function.
  - `internal/tui/views/model_registry.go:616-640` — Same confirmation pattern with minor differences (`(y/n)` vs `[y/n]`). Unify format.

  **API/Type References**:
  - `internal/tui/styles.go` — All shared styles and colors. Add new function here. Follow existing `lipgloss.Style` patterns.

  **WHY Each Reference Matters**:
  - `list.go:272-279`: This is the canonical confirmation dialog to extract. Copy the style block exactly.
  - `model_registry.go:616-640`: Shows the slight variation that needs unification.
  - `styles.go`: Target file for the new function. Follows existing pattern of style definitions.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: RenderConfirmDialog renders styled dialog
    Tool: Bash (go test)
    Preconditions: styles.go contains new function
    Steps:
      1. Create a minimal Go test file that imports styles and calls RenderConfirmDialog("test-profile", "Delete")
      2. Verify output contains "test-profile" and "[y/n]"
      3. Verify output is non-empty and styled (contains ANSI codes)
    Expected Result: Function returns styled string with target name and y/n prompt
    Failure Indicators: Empty string, missing target name, missing [y/n]
    Evidence: .sisyphus/evidence/task-1-confirm-dialog.txt

  Scenario: ErrorIconStyle renders with ⚠ prefix
    Tool: Bash (go test)
    Preconditions: styles.go contains new style
    Steps:
      1. Call ErrorIconStyle.Render("test error")
      2. Verify output contains "⚠" prefix
    Expected Result: Styled string with ⚠ prefix
    Evidence: .sisyphus/evidence/task-1-error-icon.txt
  ```

  **Commit**: NO (commits with Wave 2 — High Priority)
  - Groups with: Tasks 2-7

- [x] 2. App.go Fixes — Loading Context + "q" Hint + Help Truncation (UIUX-005, 007, 018)

  **What to do**:
  - **UIUX-005**: Add `loadingMsg string` field to `App` struct. Replace hardcoded `"Loading..."` with `a.spinner.View() + " " + a.loadingMsg + "..."`. Set `loadingMsg` contextually before each operation (e.g., "Switching profile", "Importing", "Deleting profile", "Exporting").
  - **UIUX-007**: When "q" is pressed in wizard mode (currently `break`/silently ignored), show a toast hint: `a.showToast("Press Esc to exit wizard", ToastInfo, 3*time.Second)`. Use existing toast system.
  - **UIUX-018**: When help hints are truncated (width < 45, limited to 3 items), append `"…"` or `"[?] more"` after the truncated hints. Current code at line 751-763.

  **Must NOT do**:
  - Do not change how the spinner works — only the message text
  - Do not add new key bindings — just show toast on existing blocked "q"
  - Do not change help rendering logic — only add truncation indicator

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Three small targeted changes in one file, each 3-10 lines
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 6, 7)
  - **Blocks**: None
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `internal/tui/app.go:644-651` — Loading overlay rendering. Replace `"Loading..."` with `a.loadingMsg`.
  - `internal/tui/app.go:148-158` — "q" key handling. Currently breaks silently in wizard mode. Add toast.
  - `internal/tui/app.go:751-763` — Help text truncation. Add "…" indicator.
  - `internal/tui/app.go:622-626` — `showToast()` method signature and usage pattern.

  **API/Type References**:
  - `internal/tui/app.go:App` struct — Add `loadingMsg string` field.

  **WHY Each Reference Matters**:
  - `app.go:644-651`: Exact location of loading overlay to modify.
  - `app.go:148-158`: Exact location of "q" blocking to add toast.
  - `app.go:751-763`: Exact location of truncation to add indicator.
  - `app.go:622-626`: Pattern for calling showToast correctly.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Loading overlay shows operation-specific message
    Tool: Bash (go test)
    Preconditions: App struct has loadingMsg field
    Steps:
      1. Create App with loadingMsg = "Switching profile"
      2. Set loading = true
      3. Call View()
      4. Grep output for "Switching profile..."
    Expected Result: View output contains "Switching profile..."
    Failure Indicators: Output still shows generic "Loading..."
    Evidence: .sisyphus/evidence/task-2-loading-msg.txt

  Scenario: "q" in wizard shows toast hint
    Tool: Bash (go test)
    Preconditions: App in wizard state
    Steps:
      1. Create App with state = stateWizard
      2. Send "q" key message
      3. Check that toast command is returned
    Expected Result: tea.Cmd for toast with "Press Esc to exit wizard"
    Failure Indicators: No cmd returned (silent ignore), or app exits
    Evidence: .sisyphus/evidence/task-2-q-hint.txt

  Scenario: Help truncation shows indicator
    Tool: Bash (go test)
    Preconditions: App width < 45
    Steps:
      1. Create App with width = 40
      2. Set help hints with 5+ items
      3. Call renderShortHelp()
      4. Grep output for "…"
    Expected Result: Truncated help ends with "…"
    Failure Indicators: Help truncated without any indicator
    Evidence: .sisyphus/evidence/task-2-help-truncation.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Message: `fix(uiux): resolve 7 high-priority usability and accessibility issues`
  - Files: `internal/tui/app.go`
  - Pre-commit: `make test`

- [x] 3. Wizard Error Display + Progress Indicator (UIUX-001, 008)

  **What to do**:
  - **UIUX-001**: In `wizard.go` `View()` method, between header and content rendering, add error display when `w.err != nil`:
    ```go
    errorDisplay := ""
    if w.err != nil {
        errorDisplay = styles.ErrorStyle.Render("⚠ " + w.err.Error())
    }
    return lipgloss.JoinVertical(lipgloss.Left, header, errorDisplay, content)
    ```
  - **UIUX-008**: In `wizard.go` `renderHeader()`, add progress fraction to the step progress display:
    ```go
    progress := fmt.Sprintf("Step %d of %d", w.currentStep+1, len(w.steps))
    ```
    Integrate into the existing header layout (both normal and compact modes).

  **Must NOT do**:
  - Do not change how `w.err` is set — only display it
  - Do not change step navigation logic
  - Do not create local styles — import from `styles.go`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two small additions to existing View() rendering
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 4, 5, 6, 7)
  - **Blocks**: Task 6
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard.go:406-431` — View() method. Currently renders `lipgloss.JoinVertical(header, content)`. Insert error display between them.
  - `internal/tui/views/wizard.go:434-486` — renderHeader() method. Add progress fraction here.
  - `internal/tui/views/wizard.go:97` — `w.err` field definition. Error is set but never displayed.
  - `internal/tui/views/wizard.go:301-321` — Where `w.err` is set during step transitions.

  **API/Type References**:
  - `internal/tui/styles.go:ErrorStyle` — Use this for error display styling.

  **WHY Each Reference Matters**:
  - `wizard.go:406-431`: Exact View() method to modify — the core rendering.
  - `wizard.go:434-486`: renderHeader() where progress fraction goes.
  - `wizard.go:97`: Confirms err field exists and is set.
  - `styles.go:ErrorStyle`: Shared style to use (no local color definitions).

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Wizard error displayed in View output
    Tool: Bash (go test)
    Preconditions: Wizard with err set
    Steps:
      1. Create Wizard, set w.err = fmt.Errorf("profile already exists")
      2. Call View()
      3. Grep output for "⚠" and "profile already exists"
    Expected Result: View output contains "⚠ profile already exists"
    Failure Indicators: No error text in output
    Evidence: .sisyphus/evidence/task-3-wizard-error.txt

  Scenario: Progress fraction shown in header
    Tool: Bash (go test)
    Preconditions: Wizard with multiple steps
    Steps:
      1. Create Wizard with currentStep = 2, len(steps) = 6
      2. Call renderHeader() or View()
      3. Grep output for "Step 3 of 6"
    Expected Result: Header contains "Step 3 of 6"
    Failure Indicators: No progress fraction visible
    Evidence: .sisyphus/evidence/task-3-wizard-progress.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Groups with: Tasks 1, 2, 4, 5, 6, 7

- [x] 4. Wizard Review Flash Message on Invalid Save (UIUX-002)

  **What to do**:
  - Add `flashMsg string` field to `WizardReview` struct
  - When save is attempted with invalid data (`w.isValid` is false), set `w.flashMsg = "Fix validation errors before saving"` and return
  - Render the flash message in `View()` with `styles.WarningStyle`
  - Add time-based auto-clear: include a `tea.Tick` command to clear the flash after 3 seconds
  - Clear flash on any subsequent key press (simpler than timer if preferred)

  **Must NOT do**:
  - Do not change validation logic — only add feedback
  - Do not use app-level toast — this is wizard-level feedback
  - Do not create local styles — use `styles.WarningStyle`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small addition to existing Update/View methods
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 5, 6, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_review.go:164-170` — Save key handling. Currently returns `nil` when invalid. Add flash message here.
  - `internal/tui/views/wizard_review.go:182` — View() method. Add flash rendering here.
  - `internal/tui/views/wizard_review.go:121-155` — `validateAndPreview()` sets `isValid` and validation errors.

  **API/Type References**:
  - `internal/tui/styles.go:WarningStyle` — Yellow style for warning flash.

  **WHY Each Reference Matters**:
  - `wizard_review.go:164-170`: Exact location of the silent failure to fix.
  - `wizard_review.go:182`: Where to render the flash in the view.
  - `styles.go:WarningStyle`: Shared style for consistency.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Flash message shown on invalid save attempt
    Tool: Bash (go test)
    Preconditions: Review step with isValid = false
    Steps:
      1. Create WizardReview with isValid = false
      2. Send Save key message
      3. Check flashMsg field is set
      4. Call View()
      5. Grep output for "Fix validation errors"
    Expected Result: View contains "Fix validation errors before saving"
    Failure Indicators: flashMsg empty, or View has no warning text
    Evidence: .sisyphus/evidence/task-4-review-flash.txt

  Scenario: Flash cleared on subsequent input
    Tool: Bash (go test)
    Preconditions: Review step with flashMsg set
    Steps:
      1. Create WizardReview with flashMsg = "Fix validation errors before saving"
      2. Send any key message (e.g., arrow key)
      3. Check flashMsg is cleared
    Expected Result: flashMsg is empty after key press
    Failure Indicators: flashMsg persists
    Evidence: .sisyphus/evidence/task-4-flash-clear.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Groups with: Tasks 1, 2, 3, 5, 6, 7

- [x] 5. Diff View Pane Labels Beyond Color (UIUX-006)

  **What to do**:
  - Add text labels to diff panes to indicate focus beyond color
  - Options: "◀ Left" / "Right ▶" or bold/underline the focused pane title
  - Modify `renderDiffPane()` or `View()` in `diff.go` to add pane title with direction indicator
  - The focused pane should have a clear visual indicator (e.g., bold title + "◀" / "▶")

  **Must NOT do**:
  - Do not remove color-based focus — add text labels as supplementary
  - Do not change border rendering logic
  - Do not add new colors — use existing Purple/Gray

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small rendering addition to diff view
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 6, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/diff.go:291-320` — `renderDiffPane()` renders diff lines with color. Add title label above pane content.
  - `internal/tui/views/diff.go:396-401` — `borderColor()` returns purple (focused) or gray (unfocused). Add text label that matches focus state.
  - `internal/tui/views/diff.go:403-447` — `renderSelector()` renders profile selectors. Pane labels go between selectors and content.

  **API/Type References**:
  - `internal/tui/styles.go` — `Purple`, `Gray` colors, `ActiveStyle`, `InactiveStyle`.

  **WHY Each Reference Matters**:
  - `diff.go:291-320`: Where pane content is rendered — add title label here.
  - `diff.go:396-401`: Focus indicator function — coordinate label style with this.
  - `styles.go`: Colors and styles to use.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Diff panes show directional labels
    Tool: Bash (go test)
    Preconditions: Diff view with two profiles
    Steps:
      1. Create Diff view with left profile focused
      2. Call View()
      3. Grep for "◀" or "Left" in output
    Expected Result: Output contains pane direction indicators
    Failure Indicators: No pane labels in output
    Evidence: .sisyphus/evidence/task-5-diff-labels.txt

  Scenario: Tab switches focus indicator
    Tool: Bash (go test)
    Preconditions: Diff view with left focused
    Steps:
      1. Create Diff with focused = focusLeft
      2. Send Tab key message
      3. Check focused changed to focusRight
      4. Call View() — verify "▶" or "Right" is prominent
    Expected Result: Focus indicator moves to right pane
    Failure Indicators: Labels don't change on Tab
    Evidence: .sisyphus/evidence/task-5-diff-focus-switch.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Groups with: Tasks 1, 2, 3, 4, 6, 7

- [x] 6. Wizard Back-Nav Auto-Apply + Flash (UIUX-003)

  **What to do**:
  - Modify `prevStep()` in `wizard.go` to call `Apply()` on the current step before navigating back (auto-apply)
  - After successful apply, set a flash message: "Changes applied"
  - If apply fails (validation error), still navigate back but show the existing error via UIUX-001 display
  - Add `flashMsg string` field to `Wizard` struct (same pattern as Task 4 for review step)
  - Render flash in wizard View() and clear on next key press

  **Must NOT do**:
  - Do not add confirmation dialog — auto-apply is the chosen approach
  - Do not prevent navigation — always go back, just apply first
  - Do not change step transition logic beyond prevStep()

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Modify one method + add flash display
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (but logically after Task 3 for shared error display)
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Task 3 (for error display rendering to also show apply errors)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard.go:377-404` — `prevStep()` method. Currently discards without applying. Add Apply() call here.
  - `internal/tui/views/wizard.go:262-297` — `nextStep()` method as reference — it already calls Apply(). Follow the same pattern.
  - `internal/tui/views/wizard.go:406-431` — View() where flash should render (same location as error from Task 3).

  **API/Type References**:
  - `internal/tui/styles.go:SuccessStyle` — For "Changes applied" flash.

  **WHY Each Reference Matters**:
  - `wizard.go:377-404`: The method to modify — add Apply() before navigating back.
  - `wizard.go:262-297`: Reference for how nextStep() calls Apply() — replicate this pattern.
  - `wizard.go:406-431`: Where flash renders in View().

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Back navigation applies changes before going back
    Tool: Bash (go test)
    Preconditions: Wizard on step 2+ with modified config
    Steps:
      1. Create Wizard on step 2
      2. Modify config via step
      3. Send back navigation message (Shift+Tab)
      4. Verify current step decremented AND config was applied
    Expected Result: Step decremented, Apply() was called, flash shows "Changes applied"
    Failure Indicators: Step decremented but config not applied (data loss)
    Evidence: .sisyphus/evidence/task-6-back-apply.txt

  Scenario: Back navigation still works when Apply fails
    Tool: Bash (go test)
    Preconditions: Wizard on step 2+ with invalid data
    Steps:
      1. Create Wizard on step 2
      2. Set invalid state on step
      3. Send back navigation message
      4. Verify step decremented (navigation not blocked)
      5. Verify error is shown (from UIUX-001)
    Expected Result: Navigation succeeds, error displayed
    Failure Indicators: Navigation blocked by validation error
    Evidence: .sisyphus/evidence/task-6-back-error.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Groups with: Tasks 1, 2, 3, 4, 5, 7

- [x] 7. Validation Indicators in Agent + Category Forms (UIUX-004)

  **What to do**:
  - **wizard_agents.go**: Integrate `validateAgentField()` into form field rendering so validation hints appear inline
    - Simple fields (color): validate on every keystroke
    - Range fields (temperature, top_p): validate on field-exit (Tab/Enter)
  - **wizard_categories.go**: Add validation hints to category form fields where applicable
  - Follow the pattern from `wizard_name.go` lines 137-144 (✓/✗ with styled rendering)
  - Use `styles.ErrorStyle` and `styles.SuccessStyle` — no local color definitions

  **Must NOT do**:
  - Do not validate on every keystroke for range fields (temperature, top_p) — too noisy
  - Do not create local style definitions — import from styles.go
  - Do not change the validation logic itself — only display results

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: wizard_agents.go is 2699 lines; requires careful integration of validation into form rendering
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 2, 3, 4, 5, 6)
  - **Blocks**: Task 18
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_name.go:137-144` — **CANONICAL ✓/✗ pattern** to replicate:
    ```go
    status = wizNameErrorStyle.Render(fmt.Sprintf("✗ %s", w.err.Error()))
    status = wizNameValidStyle.Render("✓ Valid name")
    ```
  - `internal/tui/views/wizard_agents.go:77-96` — `validateAgentField()` function. Already validates color, temperature, top_p but return value is not used in rendering.
  - `internal/tui/views/wizard_agents.go:1912+` — `renderContent()` where form fields are rendered. Add validation hints inline.
  - `internal/tui/views/wizard_categories.go:1043+` — `renderCategoryForm()` where category fields render.

  **API/Type References**:
  - `internal/tui/styles.go:ErrorStyle`, `styles.go:SuccessStyle` — For ✓/✗ rendering.

  **WHY Each Reference Matters**:
  - `wizard_name.go:137-144`: The exact pattern to replicate — this is the template.
  - `wizard_agents.go:77-96`: Existing validation function to wire into rendering.
  - `wizard_agents.go:1912+`: Where to add inline validation indicators.
  - `wizard_categories.go:1043+`: Where to add inline validation for category forms.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Invalid hex color shows immediate feedback
    Tool: Bash (go test)
    Preconditions: Agent form with color field focused
    Steps:
      1. Create agent form with color field active
      2. Set value to "invalid"
      3. Trigger keystroke update
      4. Call View()
      5. Grep for "✗" and "invalid hex"
    Expected Result: View shows "✗ invalid hex" inline with field
    Failure Indicators: No validation hint visible
    Evidence: .sisyphus/evidence/task-7-validation-color.txt

  Scenario: Temperature validated on field exit
    Tool: Bash (go test)
    Preconditions: Agent form with temperature field
    Steps:
      1. Set temperature value to "3.0" (out of range)
      2. Send Tab key (field exit)
      3. Call View()
      4. Grep for validation hint
    Expected Result: "✗" indicator shown after leaving field
    Failure Indicators: No validation shown, or validation shown on every keystroke
    Evidence: .sisyphus/evidence/task-7-validation-temp.txt
  ```

  **Commit**: YES (Commit 1 — High Priority)
  - Groups with: Tasks 1, 2, 3, 4, 5, 6

- [x] 8. Confirmation Dialog Refactor — list.go + model_registry.go (UIUX-011)

  **What to do**:
  - Refactor `list.go` and `model_registry.go` to use the shared `RenderConfirmDialog()` from `styles.go` (created in Task 1)
  - Replace inline `confirmStyle` blocks in both files with calls to the shared function
  - Unify the `[y/n]` vs `(y/n)` format to a single canonical format
  - Remove duplicate style definitions from both files

  **Must NOT do**:
  - Do not change confirmation dialog behavior — only styling
  - Do not add new confirmation features
  - Do not change key handling logic (y/n/esc)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical refactor to use shared function
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 9, 10, 11, 12, 13, 14)
  - **Blocks**: None
  - **Blocked By**: Task 1 (needs shared function from styles.go)

  **References**:

  **Pattern References**:
  - `internal/tui/views/list.go:272-279` — Current inline confirmation dialog. Replace with shared function call.
  - `internal/tui/views/model_registry.go:616-640` — Current inline confirmation dialog. Replace with shared function call.
  - `internal/tui/styles.go` — Target: `RenderConfirmDialog()` function (from Task 1).

  **WHY Each Reference Matters**:
  - Both files have the same pattern — mechanical replacement with shared function.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Both views produce identical confirmation output
    Tool: Bash (go test)
    Preconditions: Both files refactored to use shared function
    Steps:
      1. Call RenderConfirmDialog("test-profile", "Delete") from styles.go
      2. Verify list.go and model_registry.go no longer define local confirmStyle
      3. Grep both files for "confirmStyle" — should only find import from styles
    Expected Result: No local confirmStyle definitions, both use shared function
    Failure Indicators: Local confirmStyle still exists in either file
    Evidence: .sisyphus/evidence/task-8-confirm-refactor.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 9, 10, 11, 12, 13, 14

- [x] 9. Wizard Name Help Fixes (UIUX-010, 019)

  **What to do**:
  - **UIUX-010**: Add `esc` to the help text displayed in the name step. Currently only shows "tab/enter: next" — add "esc: cancel"
  - **UIUX-019**: Add `shift+tab` support to the name step for back navigation. Currently only `esc` cancels. Both `shift+tab` and `esc` should work for consistency with other wizard steps.

  **Must NOT do**:
  - Do not change navigation behavior — only add missing key support
  - Do not change the name validation logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two small additions to keymap and help text
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 10, 11, 12, 13, 14)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_name.go:33-44` — Key bindings definition. Add shift+tab binding.
  - `internal/tui/views/wizard_name.go:128-151` — View() with help text rendering. Add "esc" to help text.
  - `internal/tui/views/wizard.go:55-74` — Wizard key bindings showing both shift+tab and esc for back. Match this.

  **WHY Each Reference Matters**:
  - `wizard_name.go:33-44`: Where keymap is defined — add shift+tab.
  - `wizard_name.go:128-151`: Where help text is rendered — add esc.
  - `wizard.go:55-74`: Reference for consistent key binding pattern.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Help text includes esc key
    Tool: Bash (go test)
    Steps:
      1. Create WizardName, call View()
      2. Grep for "esc" in output
    Expected Result: Help text contains "esc" as option
    Failure Indicators: No "esc" in help text
    Evidence: .sisyphus/evidence/task-9-name-help.txt

  Scenario: Shift+Tab triggers back navigation
    Tool: Bash (go test)
    Steps:
      1. Create WizardName with valid input
      2. Send shift+tab key message
      3. Verify WizardBackMsg or cancel message returned
    Expected Result: Navigation back triggered
    Failure Indicators: Key ignored
    Evidence: .sisyphus/evidence/task-9-shift-tab.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 10, 11, 12, 13, 14

- [x] 10. Help Text Standardization — Categories, Agents, Other, Hooks (UIUX-009, 020)

  **What to do**:
  - **UIUX-009**: Standardize help text order across wizard steps to: navigation keys → action keys → step keys. Use bracketed format `[key] action`.
    - Categories: Change `"n: new • d: delete • →: expand • ←: collapse • Enter: edit • Tab: next step"` → `"[↑↓] navigate  [n] new  [d] delete  [→] expand  [←] collapse  [Enter] edit  [Tab] next step"`
    - Agents: Change `"Space to enable/disable • Enter to expand • Tab to next step"` → `"[Space] toggle  [Enter] expand  [Tab] next step"`
    - Other: Change `"Enter to expand • Space to toggle • Tab next • Shift+Tab back"` → `"[Enter] expand  [Space] toggle  [Tab] next  [Shift+Tab] back"`
  - **UIUX-020**: When editing a category form, show context label: `Editing: [category-name] — [↑↓] navigate  [Space] toggle  [Esc] close`

  **Must NOT do**:
  - Do not change key bindings — only text format
  - Do not add new functionality to help text
  - Do not touch other view files in this task (those come in Task 17)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical text changes across 4 files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 11, 12, 13, 14)
  - **Blocks**: Task 17
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_categories.go:1150` — Current help text. Rewrite in bracketed format.
  - `internal/tui/views/wizard_agents.go:2649` — Current help text. Rewrite in bracketed format.
  - `internal/tui/views/wizard_other.go:3764` — Current help text. Rewrite in bracketed format.
  - `internal/tui/views/wizard_hooks.go` — Help text in View(). Rewrite in bracketed format.
  - `internal/tui/views/wizard_categories.go:1152-1154` — Form mode help. Add context label.

  **WHY Each Reference Matters**:
  - All four files have inconsistent help formats. Standardize to single bracketed format.
  - `wizard_categories.go:1152-1154`: Where form editing help is rendered — add category name context.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Help text uses bracketed format consistently
    Tool: Bash (go test)
    Steps:
      1. Create each wizard step, call View()
      2. Grep each output for bracketed key format like "[↑↓]" or "[Tab]"
    Expected Result: All wizard step help uses bracketed format
    Failure Indicators: Any step still uses "key: action" or "key to action" format
    Evidence: .sisyphus/evidence/task-10-help-format.txt

  Scenario: Category form shows context label
    Tool: Bash (go test)
    Steps:
      1. Create WizardCategories with inForm = true, editing category "general"
      2. Call View()
      3. Grep for "Editing:" and "general"
    Expected Result: Output contains "Editing: general" context
    Failure Indicators: No context label in form mode
    Evidence: .sisyphus/evidence/task-10-context-label.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 9, 11, 12, 13, 14

- [x] 11. Wizard Hooks — Status Labels + Toggle Visual + Enabled Text (UIUX-013, 016, 022)

  **What to do**:
  - **UIUX-013**: Change hook status from color-only to text+color: `[✓] enabled` / `[✗] disabled` (currently disabled shows "(disabled)" but enabled shows nothing)
  - **UIUX-016**: Make the "Include disabled_hooks in profile" toggle visually distinct with checkbox style: `[x] Include disabled_hooks in profile` and add to help text
  - **UIUX-022**: Add "(enabled)" text label for enabled hooks for symmetry with "(disabled)"

  **Must NOT do**:
  - Do not change toggle logic — only visual appearance
  - Do not add new local color definitions

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Three small rendering changes in one file
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10, 12, 13, 14)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks.go:316-330` — Current hook status rendering. Add text labels.
  - `internal/tui/views/wizard_hooks.go:236-239` — Include toggle. Make visually distinct with brackets.
  - `internal/tui/views/wizard_hooks.go:291-299` — Toggle row rendering.

  **WHY Each Reference Matters**:
  - `wizard_hooks.go:316-330`: Where status is rendered — add text for both enabled and disabled states.
  - `wizard_hooks.go:236-239`: Where toggle is defined — add visual brackets.
  - `wizard_hooks.go:291-299`: Where toggle row renders — update visual style.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Hook status shows text labels for both states
    Tool: Bash (go test)
    Steps:
      1. Create WizardHooks with some enabled, some disabled hooks
      2. Call View()
      3. Grep for "enabled" and "disabled"
    Expected Result: Both "enabled" and "disabled" text visible
    Failure Indicators: Only "(disabled)" shown, no "(enabled)" for active hooks
    Evidence: .sisyphus/evidence/task-11-hook-status.txt

  Scenario: Include toggle is visually distinct
    Tool: Bash (go test)
    Steps:
      1. Create WizardHooks, call View()
      2. Grep for bracket-style checkbox "[x]" or "[ ]" on include toggle row
    Expected Result: Include toggle shows checkbox-style indicator
    Failure Indicators: Toggle looks like a plain text label
    Evidence: .sisyphus/evidence/task-11-toggle-visual.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 9, 10, 12, 13, 14

- [x] 12. Wizard Agents Simplified Help (UIUX-017)

  **What to do**:
  - Simplify the dense inline help in agent nested editing (permissions, fallback models)
  - Current: `"a:add d:del c:to-string ↑↓:nav ←→:cycle esc:done"` — too many commands
  - Proposed: Show only the 3 most common actions, with `"[?] more"` to expand
  - Add a visual border/title for the editing mode to differentiate from the list view

  **Must NOT do**:
  - Do not remove key bindings — only hide them from initial display
  - Do not change editing logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Help text simplification in one file
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10, 11, 13, 14)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:2095-2103` — Dense help text for permissions editing.
  - `internal/tui/views/wizard_agents.go:2134-2143` — Dense help text for fallback model editing.

  **WHY Each Reference Matters**:
  - Both locations have the dense help that needs simplification. Show 3 most common + "[?] more".

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Nested editing shows simplified help
    Tool: Bash (go test)
    Steps:
      1. Create WizardAgents in permission editing mode
      2. Call View()
      3. Verify help shows max 3 action keys visible
      4. Verify "[?]" or "more" indicator is present
    Expected Result: Help shows ≤3 actions + "more" indicator
    Failure Indicators: All 6+ actions shown at once
    Evidence: .sisyphus/evidence/task-12-agent-help.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 9, 10, 11, 13, 14

- [x] 13. Export Resolved Path Notification (UIUX-012)

  **What to do**:
  - When export is triggered and the file is auto-renamed by `autoRenameIfExists()`, show the resolved path to the user
  - Currently: `autoRenameIfExists()` returns a new path silently
  - Proposed: After resolving the path, display it in the export view before/during export: `"Exporting to: {resolved-path}"`
  - If the path was changed from the original, add a note: `"File exists. Exporting as: {resolved-path}"`

  **Must NOT do**:
  - Do not add a confirmation prompt — just show the resolved path
  - Do not change the auto-rename logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Add notification text for auto-renamed paths
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10, 11, 12, 14)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/export.go:115-117` — Where auto-rename is called.
  - `internal/tui/views/export.go:202-216` — `autoRenameIfExists()` function.
  - `internal/tui/views/export.go:154` — Error display pattern with `exportErrorStyle.Render("✗ "+e.err.Error())`.

  **WHY Each Reference Matters**:
  - `export.go:115-117`: Where path resolution happens — capture resolved path for display.
  - `export.go:202-216`: The auto-rename logic — understand what triggers the notification.
  - `export.go:154`: Pattern for displaying messages in export view.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Auto-renamed export shows notification
    Tool: Bash (go test)
    Steps:
      1. Create Export view with path pointing to existing file
      2. Trigger export
      3. Call View() after auto-rename
      4. Grep for "Exporting as:" and the new filename
    Expected Result: View shows "File exists. Exporting as: profile-1.json"
    Failure Indicators: No notification of renamed file
    Evidence: .sisyphus/evidence/task-13-export-path.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 9, 10, 11, 12, 14

- [x] 14. Schema Check Error Weight + Template Select Empty State (UIUX-014, 015)

  **What to do**:
  - **UIUX-014**: Improve error display visual weight in schema_check.go. The `[r] retry [esc] back` already exists but the error message itself is plain red text. Add `⚠ ` prefix using `styles.ErrorStyle` (or the new `ErrorIconStyle` from Task 1) and add a separator line above the error.
  - **UIUX-015**: Improve empty state message in template_select.go. Change from: `"No profiles available to use as template. Press esc to go back."` To: `"No profiles available to use as template.\n\nCreate a profile from the dashboard first, then use it as a template.\n\nPress esc to go back."`

  **Must NOT do**:
  - Do not change retry mechanism (UIUX-014) — only improve error visual
  - Do not add new navigation from template_select to dashboard (UIUX-015)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two trivial text/rendering changes in two files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8, 9, 10, 11, 12, 13)
  - **Blocks**: None
  - **Blocked By**: Task 1 (for ErrorIconStyle)

  **References**:

  **Pattern References**:
  - `internal/tui/views/schema_check.go:153-156,238` — Error rendering. Add ⚠ prefix and separator.
  - `internal/tui/views/template_select.go:103-109` — Empty state message. Add guidance text.

  **WHY Each Reference Matters**:
  - `schema_check.go:153-156`: Where error is rendered — add visual weight.
  - `template_select.go:103-109`: Where empty state is shown — add next-step guidance.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Schema check error has visual weight
    Tool: Bash (go test)
    Steps:
      1. Create SchemaCheck in error state with errorMsg = "connection failed"
      2. Call View()
      3. Grep for "⚠" and separator
    Expected Result: Error message prefixed with ⚠ and visually separated
    Failure Indicators: Plain red text with no icon or separator
    Evidence: .sisyphus/evidence/task-14-schema-error.txt

  Scenario: Template select empty state shows guidance
    Tool: Bash (go test)
    Steps:
      1. Create TemplateSelect with empty profiles list
      2. Call View()
      3. Grep for "Create a profile" or "dashboard"
    Expected Result: Message includes next-step guidance about creating profiles
    Failure Indicators: Only shows "Press esc to go back"
    Evidence: .sisyphus/evidence/task-14-template-empty.txt
  ```

  **Commit**: YES (Commit 2 — Medium Priority)
  - Groups with: Tasks 8, 9, 10, 11, 12, 13

- [x] 15. Label Clarity + Format Hints + Scroll/Search Keys (UIUX-021, 023, 024, 025)

  **What to do**:
  - **UIUX-021**: Change ambiguous "Include disabled_hooks in profile" label in `wizard_categories.go` to: `Include 'disabled_hooks' field in profile` (with quotes around field name)
  - **UIUX-023**: Update model_selector.go placeholder from `"e.g., gpt-4o-mini"` to `"e.g., anthropic/claude-sonnet-4-20250514"` to show provider/model-id format
  - **UIUX-024**: Add scroll keys to help bar in model_import.go: `[pgup/pgdn] scroll` or `[j/k] scroll`
  - **UIUX-025**: Add `[/] search` to help text in model_registry.go if not already present

  **Must NOT do**:
  - Do not change field logic or validation — only text
  - Do not add new key bindings for scroll (UIUX-024) — only document existing ones in help

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Four trivial text changes across four files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 16, 17, 18)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_categories.go:306` — "Include disabled_hooks" label. Add quotes.
  - `internal/tui/views/model_selector.go:513-525` — CustomInput placeholder. Update format example.
  - `internal/tui/views/model_import.go:538-540` — Scroll indicators present but help doesn't mention keys.
  - `internal/tui/views/model_registry.go:307` — "/" search binding exists but may not be in help text.

  **WHY Each Reference Matters**:
  - All four are simple text updates with exact line references.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Labels and hints updated
    Tool: Bash (go test)
    Steps:
      1. Grep wizard_categories.go for "Include 'disabled_hooks'" — should find quoted version
      2. Grep model_selector.go for "anthropic/" — should find provider/model format in placeholder
      3. Grep model_import.go help text for scroll-related keys
      4. Grep model_registry.go help text for "/" search key
    Expected Result: All four text changes present
    Failure Indicators: Any original text unchanged
    Evidence: .sisyphus/evidence/task-15-labels-hints.txt
  ```

  **Commit**: YES (Commit 3 — Low Priority)
  - Groups with: Tasks 16, 17, 18

- [x] 16. Visual Polish — Viewport Constants + Error Weight + Compact Indicator (UIUX-026, 029, 030)

  **What to do**:
  - **UIUX-026**: Extract hardcoded viewport overhead magic numbers to named constants:
    ```go
    const (
        viewportOverheadNormal = 4 // title + help + 2 spacing lines
        viewportOverheadShort  = 2 // title + help only (compact mode)
    )
    ```
    Apply in `wizard_hooks.go` and `wizard_review.go` (and any other files with hardcoded overhead values).
  - **UIUX-029**: Add `⚠ ` prefix to error display in model_registry.go. Currently plain `errorStyle.Render("Error: "+m.errorMsg)` — add prefix for visual weight.
  - **UIUX-030**: Add `[compact]` indicator in help bar when layout is constrained in `wizard_hooks.go`. Check `layout.IsShort(height)` and append indicator.

  **Must NOT do**:
  - Do not extract constants for values NOT mentioned in UIUX-026
  - Do not change viewport calculation logic — only name the magic numbers
  - Do not add new layout behavior

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Three small mechanical changes across three files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15, 17, 18)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_hooks.go:160-162` — Hardcoded `overhead := 4`. Extract to constant.
  - `internal/tui/views/wizard_review.go:100-102` — Hardcoded overhead values. Extract to constant.
  - `internal/tui/views/model_registry.go:608-611` — Error display. Add ⚠ prefix.
  - `internal/tui/views/wizard_hooks.go:352-372` — Compact mode. Add `[compact]` indicator.

  **WHY Each Reference Matters**:
  - `wizard_hooks.go:160-162`: Magic number location to extract.
  - `wizard_review.go:100-102`: Second magic number location.
  - `model_registry.go:608-611`: Error display to add prefix.
  - `wizard_hooks.go:352-372`: Where compact mode affects layout — add indicator.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: No hardcoded viewport overhead values
    Tool: Bash (grep)
    Steps:
      1. Grep wizard_hooks.go and wizard_review.go for "overhead := [0-9]"
      2. Should find references to named constants, not raw numbers
    Expected Result: Overhead uses named constants (viewportOverheadNormal/Short)
    Failure Indicators: Raw numeric literals still present
    Evidence: .sisyphus/evidence/task-16-constants.txt

  Scenario: Model registry error has visual weight
    Tool: Bash (go test)
    Steps:
      1. Create ModelRegistry with errorMsg set
      2. Call View()
      3. Grep for "⚠" prefix
    Expected Result: Error message prefixed with ⚠
    Failure Indicators: Plain error text without icon
    Evidence: .sisyphus/evidence/task-16-error-weight.txt

  Scenario: Compact mode shows indicator
    Tool: Bash (go test)
    Steps:
      1. Create WizardHooks with constrained height (layout.IsShort = true)
      2. Call View()
      3. Grep for "compact"
    Expected Result: Help bar shows "[compact]" indicator
    Failure Indicators: No indicator in compact mode
    Evidence: .sisyphus/evidence/task-16-compact-indicator.txt
  ```

  **Commit**: YES (Commit 3 — Low Priority)
  - Groups with: Tasks 15, 17, 18

- [x] 17. Help Text Format Standardization Across All Views (UIUX-027)

  **What to do**:
  - Standardize help text across ALL remaining views to bracketed format `[key] action`
  - This covers views not already handled in Task 10:
    - `template_select.go`: Change `"↑/↓ navigate • enter select • esc cancel"` → `"[↑↓] navigate  [Enter] select  [Esc] cancel"`
    - `model_import.go`: Already uses bracketed format partially — ensure full consistency
    - `model_registry.go`: Ensure consistent format in all help text
    - `app.go` help bar: Already uses bracketed format — verify consistency
  - This task should run AFTER Task 10 (which handles wizard steps) to avoid conflicts

  **Must NOT do**:
  - Do not change key bindings — only text format
  - Do not reformat help text that was already changed in Task 10
  - Do not add new help items

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Mechanical text format changes across 4-5 files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15, 16, 18)
  - **Blocks**: None
  - **Blocked By**: Task 10 (must not conflict with wizard step help text changes)

  **References**:

  **Pattern References**:
  - `internal/tui/views/template_select.go:143` — Help text to reformat.
  - `internal/tui/views/model_import.go:494` — Help text to verify/reformat.
  - `internal/tui/views/model_registry.go` — Help text to verify/reformat.
  - `internal/tui/app.go:721-766` — Help bar rendering to verify.

  **WHY Each Reference Matters**:
  - All remaining views with non-bracketed help text need standardization.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: All views use bracketed help format
    Tool: Bash (grep)
    Steps:
      1. Grep all view files for help text patterns
      2. Verify no "key: action" or "key to action" format remains (outside wizard steps from Task 10)
      3. Verify bracketed format [key] action used consistently
    Expected Result: All help text uses [key] action format
    Failure Indicators: Any non-bracketed help text remaining
    Evidence: .sisyphus/evidence/task-17-help-standard.txt
  ```

  **Commit**: YES (Commit 3 — Low Priority)
  - Groups with: Tasks 15, 16, 18

- [x] 18. Micro-Feedback on Field Completion Across Wizard Steps (UIUX-028)

  **What to do**:
  - Add subtle confirmation indicators for validated fields across wizard steps
  - Follow the pattern from `wizard_name.go` (✓ for valid, ✗ for invalid)
  - Scope: Only add ✓ to fields that have clear validation rules — do NOT add to every field
  - Specifically:
    - `wizard_categories.go`: Add ✓ to completed/valid category fields
    - `wizard_agents.go`: Build on Task 7 validation indicators — add ✓ for valid fields
    - `wizard_other.go`: Add ✓ to boolean fields that are toggled (brief flash or static indicator)
    - `wizard_hooks.go`: Already has ✓/✗ for hooks — verify consistency
  - Use `styles.SuccessStyle` for ✓ — no local style definitions

  **Must NOT do**:
  - Do not add validation indicators to fields without clear validation rules
  - Do not create local style definitions
  - Do not over-engineer — simple ✓ suffix after valid values is sufficient
  - Do not add animation or timer-based feedback — static indicator only

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Touches 4 wizard step files, requires careful integration with existing rendering
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 15, 16, 17)
  - **Blocks**: None
  - **Blocked By**: Task 7 (wizard_agents validation indicators from Wave 2)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_name.go:137-144` — **CANONICAL pattern**: ✓/✗ display. Replicate this.
  - `internal/tui/views/wizard_categories.go:1043+` — Category form rendering. Add ✓ for valid fields.
  - `internal/tui/views/wizard_agents.go` — Agent form rendering. Build on Task 7 validation indicators.
  - `internal/tui/views/wizard_other.go` — Boolean toggle fields. Add ✓ indicator.
  - `internal/tui/views/wizard_hooks.go:316-330` — Already has ✓/✗ for hooks. Verify consistency.

  **API/Type References**:
  - `internal/tui/styles.go:SuccessStyle` — For ✓ rendering.

  **WHY Each Reference Matters**:
  - `wizard_name.go:137-144`: The template to follow for all ✓ indicators.
  - Other files: Where to add the indicators in their respective rendering methods.

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Valid fields show checkmark indicator
    Tool: Bash (go test)
    Steps:
      1. Create each wizard step with valid data populated
      2. Call View() for each
      3. Grep for "✓" in each output
    Expected Result: Valid fields in each step show ✓ indicator
    Failure Indicators: No ✓ visible for completed/valid fields
    Evidence: .sisyphus/evidence/task-18-micro-feedback.txt

  Scenario: Does not add ✓ to fields without validation
    Tool: Bash (go test)
    Steps:
      1. Review all fields in wizard steps
      2. Verify ✓ only appears on fields with clear validation rules
    Expected Result: ✓ only on fields like name, color, temperature, etc. Not on arbitrary text fields
    Failure Indicators: ✓ on every single field regardless of validation
    Evidence: .sisyphus/evidence/task-18-feedback-scope.txt
  ```

  **Commit**: YES (Commit 3 — Low Priority)
  - Groups with: Tasks 15, 16, 17

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
>
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, grep for pattern). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `make build` + `make lint` + `make test`. Review all changed files for: new local color definitions (anti-pattern), commented-out code, unused imports, console.log/print statements. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real QA via tmux** — `unspecified-high`
  Build binary. Launch in tmux. Execute key QA scenarios from tasks: verify error display in wizard, flash messages, diff pane labels, help text format, confirmation dialogs, validation indicators. Capture screenshots to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git diff). Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Commit 1** (after Wave 2 — High Priority): `fix(uiux): resolve 7 high-priority usability and accessibility issues`
  - Files: `internal/tui/styles.go`, `internal/tui/app.go`, `internal/tui/views/wizard.go`, `internal/tui/views/wizard_review.go`, `internal/tui/views/diff.go`, `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_categories.go`
  - Pre-commit: `make test`

- **Commit 2** (after Wave 3 — Medium Priority): `fix(uiux): resolve 13 medium-priority consistency and navigation issues`
  - Files: `internal/tui/views/list.go`, `internal/tui/views/model_registry.go`, `internal/tui/views/wizard_name.go`, `internal/tui/views/wizard_categories.go`, `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_hooks.go`, `internal/tui/views/wizard_other.go`, `internal/tui/views/export.go`, `internal/tui/views/schema_check.go`, `internal/tui/views/template_select.go`
  - Pre-commit: `make test`

- **Commit 3** (after Wave 4 — Low Priority): `fix(uiux): resolve 10 low-priority polish and formatting issues`
  - Files: `internal/tui/views/wizard_categories.go`, `internal/tui/views/model_selector.go`, `internal/tui/views/model_import.go`, `internal/tui/views/model_registry.go`, `internal/tui/views/wizard_hooks.go`, `internal/tui/views/wizard_review.go`, `internal/tui/views/wizard_other.go`, `internal/tui/views/app.go`
  - Pre-commit: `make test`

---

## Success Criteria

### Verification Commands
```bash
make build  # Expected: binary built successfully
make test   # Expected: all tests pass (36 files, 0 failures)
make lint   # Expected: no new warnings
```

### Final Checklist
- [x] All 30 UIUX issues implemented
- [x] No new files created (except function addition to styles.go)
- [x] No new dependencies added
- [x] No business logic changed
- [x] All existing tests pass
- [x] 3 atomic commits created by priority
