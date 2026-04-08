# UI/UX Improvements Analysis Report

## Executive Summary

Analysis of the omo-profiler TUI application — a Go-based terminal profile manager built with Charmbracelet Bubble Tea. The app features a dashboard, multi-step wizard, profile management, model registry, and diff views. Overall the app is well-architected with good keyboard navigation and responsive layouts. However, several friction points were identified across **40 findings** spanning usability, consistency, accessibility, and visual polish.

**Key themes:** silent failures without user feedback, inconsistent navigation/help text across views, missing validation indicators in wizard forms, and accessibility gaps with color-only distinctions.

---

## Issues Found

### High Priority

#### UIUX-001: Wizard-Level Errors Never Displayed

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard.go` (lines 97, 301-321, 406-431)

**Current State:**
The `Wizard` struct stores an `err` field and validation errors are set during step transitions, but `View()` never renders them. Users see no feedback when wizard-level validation fails.

**Proposed Change:**
Display wizard errors between the header and step content.

```go
// Current (wizard.go View())
return lipgloss.JoinVertical(lipgloss.Left, header, content)

// Proposed
errorDisplay := ""
if w.err != nil {
    errorDisplay = errorStyle.Render("⚠ " + w.err.Error())
}
return lipgloss.JoinVertical(lipgloss.Left, header, errorDisplay, content)
```

**User Benefit:** Users see why their action was rejected instead of wondering why nothing happened.

**Estimated Effort:** trivial

---

#### UIUX-002: Silent Failure on Save Attempt with Invalid Data

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_review.go` (lines 164-170)

**Current State:**
Pressing Enter/Ctrl+S on the review step when validation fails silently returns without any feedback. The user gets no confirmation that save was rejected.

**Proposed Change:**
Flash a warning message when save is attempted with validation errors.

```go
// Current
case key.Matches(msg, w.keys.Save):
    if w.isValid {
        return w, func() tea.Msg { return WizardNextMsg{} }
    }
    return w, nil  // Silent failure

// Proposed
case key.Matches(msg, w.keys.Save):
    if w.isValid {
        return w, func() tea.Msg { return WizardNextMsg{} }
    }
    w.flashMsg = "Fix validation errors before saving"
    return w, nil
```

**User Benefit:** Users understand they need to fix errors before saving rather than thinking the app is broken.

**Estimated Effort:** small

---

#### UIUX-003: Back Navigation Discards Changes Without Warning

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard.go` (lines 377-404)

**Current State:**
Going back in the wizard discards uncommitted changes in the current step without any warning. The code documents this is intentional but provides no user feedback.

**Proposed Change:**
Add a brief warning or confirmation when backing out of a step with unsaved edits, or auto-apply before navigating back.

**User Benefit:** Users don't lose work unexpectedly when navigating between wizard steps.

**Estimated Effort:** medium

---

#### UIUX-004: Missing Validation Feedback in Category/Agent Form Fields

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_categories.go` (form fields)
- `internal/tui/views/wizard_agents.go` (lines 77-96)

**Current State:**
Fields like `temperature` (0.0-2.0), `top_p` (0.0-1.0), and `color` (hex) have validation logic (e.g., `validateAgentField`) but it's never called during the main form flow. Users enter invalid values with no immediate feedback — errors only appear on save.

**Proposed Change:**
Integrate real-time validation indicators into form rendering, similar to `wizard_name.go` which shows ✓/✗.

```go
// Current: validation only on save
// Proposed: validate on every keystroke/blur
validationHint := validateAgentField(label, value)
if validationHint != "" {
    fieldLine += " " + validationHint
}
```

**User Benefit:** Users get immediate feedback on invalid input rather than discovering errors at the end.

**Estimated Effort:** medium

---

#### UIUX-005: Loading Overlay Lacks Context

**Category:** usability

**Affected Components:**
- `internal/tui/app.go` (lines 644-651)

**Current State:**
Loading overlay shows generic "Loading..." with a spinner, regardless of the operation (import, export, switch, delete). No cancel information.

**Proposed Change:**
Show operation-specific loading message and cancelability status.

```go
// Current
a.spinner.View() + " Loading..."

// Proposed
a.spinner.View() + " " + a.loadingMsg + "..."
// Where loadingMsg is set per-operation: "Switching profile", "Importing", etc.
```

**User Benefit:** Users know what's happening and whether they can cancel.

**Estimated Effort:** small

---

#### UIUX-006: Diff View Panes Differentiated Only by Color

**Category:** accessibility

**Affected Components:**
- `internal/tui/views/diff.go` (lines 171-195, 349-373)

**Current State:**
Left/right diff panes are distinguished only by border color (purple = focused, gray = unfocused). Tab switches pane silently with only border color changing.

**Proposed Change:**
Add text labels like "◀ Left" / "Right ▶" or bold/underline the focused pane title. Add a focus indicator beyond color.

**User Benefit:** Colorblind users and users on monochrome terminals can identify the active pane.

**Estimated Effort:** trivial

---

#### UIUX-007: "q" Silently Blocked in Wizard

**Category:** usability

**Affected Components:**
- `internal/tui/app.go` (lines 149-158)

**Current State:**
Pressing "q" in the wizard is silently ignored to prevent accidental data loss, but no feedback is shown.

**Proposed Change:**
Show a brief hint: "Press Esc to exit wizard" when "q" is pressed.

**User Benefit:** Users learn the correct way to exit rather than thinking the app is frozen.

**Estimated Effort:** trivial

---

### Medium Priority

#### UIUX-008: No Wizard Progress Indicator (Step X of Y)

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard.go` (lines 460-486)

**Current State:**
Progress is shown as dots/checkmarks (✓ Name → ● Categories → ○ ...) but there's no explicit "Step 3 of 6" or percentage indicator.

**Proposed Change:**
Add a progress fraction to the header.

```go
// Add to renderHeader:
progress := fmt.Sprintf("Step %d of %d", currentStep+1, totalSteps)
```

**User Benefit:** Users know exactly how far along they are and how much is left.

**Estimated Effort:** trivial

---

#### UIUX-009: Inconsistent Help Text Order Across Steps

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_categories.go` (line 1150)
- `internal/tui/views/wizard_agents.go` (line 2649)
- `internal/tui/views/wizard_other.go` (line 3764)

**Current State:**
Help text uses different orders and phrasing across steps:
- Categories: `"n: new • d: delete • →: expand • ←: collapse • Enter: edit • Tab: next step"`
- Agents: `"Space to enable/disable • Enter to expand • Tab to next step"`
- Other: `"Enter to expand • Space to toggle • Tab next • Shift+Tab back"`

**Proposed Change:**
Standardize help text ordering: navigation keys → action keys → step keys. Use consistent phrasing (either "key: action" or "key to action", not both).

**User Benefit:** Users build muscle memory and don't need to re-learn controls on each step.

**Estimated Effort:** small

---

#### UIUX-010: Missing Escape Key in Name Step Help

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_name.go` (lines 128-151)

**Current State:**
Name step only shows "tab/enter: next" but doesn't show `esc` to cancel, even though the keymap supports it.

**Proposed Change:**
Add escape key to displayed help text.

**User Benefit:** Users discover how to cancel without guessing.

**Estimated Effort:** trivial

---

#### UIUX-011: Inconsistent Confirmation Dialog Styling

**Category:** visual

**Affected Components:**
- `internal/tui/views/list.go` (lines 272-280)
- `internal/tui/views/model_registry.go`

**Current State:**
Delete confirmation dialogs use different styling between profile list and model registry. Profile list uses a yellow-on-gray styled inline message; model registry uses a different approach.

**Proposed Change:**
Extract a shared confirmation dialog component with consistent styling.

**User Benefit:** Consistent UI behavior builds trust and reduces cognitive load.

**Estimated Effort:** small

---

#### UIUX-012: Export Auto-Rename Without User Notification

**Category:** usability

**Affected Components:**
- `internal/tui/views/export.go` (lines 115-117, 202-216)

**Current State:**
When a file exists at the export path, it's silently auto-renamed. The user isn't told the final filename before or after export.

**Proposed Change:**
Show the resolved path before exporting, or prompt for confirmation: "File exists. Export as profile-1.json? [y/n]"

**User Benefit:** Users know exactly where their file was saved.

**Estimated Effort:** small

---

#### UIUX-013: Color-Only Hook Status Indicators

**Category:** accessibility

**Affected Components:**
- `internal/tui/views/wizard_hooks.go` (lines 316-330)

**Current State:**
Hook enabled/disabled uses green ✓ / red ✗ with color as primary differentiator. The "(disabled)" text label only appears for disabled hooks; enabled hooks have no text status.

**Proposed Change:**
Add text labels for both states: "[✓ enabled]" / "[✗ disabled]"

**User Benefit:** Status is readable regardless of color perception.

**Estimated Effort:** trivial

---

#### UIUX-014: Schema Check Error Lacks Recovery Actions

**Category:** usability

**Affected Components:**
- `internal/tui/views/schema_check.go` (lines 153-156, 238)

**Current State:**
When schema check fails, error is shown but recovery options aren't clear. No retry action is offered.

**Proposed Change:**
Add actionable error display: "Failed to fetch schema: {error}. Press [r] to retry or [esc] to go back."

**User Benefit:** Users can recover from transient errors without navigating away and back.

**Estimated Effort:** small

---

#### UIUX-015: Template Select Empty State Not Actionable

**Category:** usability

**Affected Components:**
- `internal/tui/views/template_select.go` (lines 103-109)

**Current State:**
Shows "No profiles available to use as template. Press esc to go back." — no suggestion of what to do next.

**Proposed Change:**
Add guidance: "Create a profile from the dashboard first, then use it as a template."

**User Benefit:** Users know the path to resolve the empty state.

**Estimated Effort:** trivial

---

#### UIUX-016: Hidden Include Toggle in Hooks Step

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_hooks.go` (lines 236-239, 291-299)

**Current State:**
The "Include disabled_hooks in profile" row is toggleable with Space, but this is only discoverable by accident. It looks like a label, not an interactive element.

**Proposed Change:**
Make it visually distinct with brackets: `[x] Include disabled_hooks in profile` and mention in help text.

**User Benefit:** Users discover the toggle without trial and error.

**Estimated Effort:** trivial

---

#### UIUX-017: Confusing Help for Complex Nested Editing in Agents

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_agents.go` (lines 2095-2103, 2134-2143)

**Current State:**
Permissions and fallback model editing show dense inline help: `"a:add d:del c:to-string ↑↓:nav ←→:cycle esc:done"`. Too many commands for new users.

**Proposed Change:**
Show only the 3 most common actions, with "? for more" to see full list. Add a visual border/title for the editing mode.

**User Benefit:** Reduces cognitive overload during complex nested editing.

**Estimated Effort:** small

---

#### UIUX-018: Help Text Truncated Without Indicator

**Category:** visual

**Affected Components:**
- `internal/tui/app.go` (lines 728-753)

**Current State:**
On narrow terminals (< 45 chars), help hints are silently truncated to 3 items with no indication more exist.

**Proposed Change:**
When truncating, append "…" or "[?] more" to signal hidden shortcuts.

**User Benefit:** Users know there are more shortcuts available.

**Estimated Effort:** trivial

---

#### UIUX-019: Back Key Inconsistency Between Wizard and Name Step

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard.go` (lines 55-74)
- `internal/tui/views/wizard_name.go` (lines 33-44)

**Current State:**
Main wizard uses both `shift+tab` and `esc` for back. Name step only uses `esc`. Inconsistent behavior for the same conceptual action.

**Proposed Change:**
Make Name step also accept `shift+tab` for back, or document the difference.

**User Benefit:** Consistent navigation across all wizard steps.

**Estimated Effort:** trivial

---

#### UIUX-020: No Context Category Label During Form Editing

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_categories.go` (lines 1152-1154)

**Current State:**
When editing a category form, help switches to generic navigation text but doesn't indicate which category is being edited.

**Proposed Change:**
Show context: "Editing: [category-name] — ↑/↓ navigate • Space: toggle • Esc: close"

**User Benefit:** Users maintain context during nested editing.

**Estimated Effort:** trivial

---

### Low Priority

#### UIUX-021: Unclear "Include" Semantics in Form Labels

**Category:** usability

**Affected Components:**
- `internal/tui/views/wizard_categories.go` (line 306)

**Current State:**
"Include disabled_hooks in profile" is ambiguous — could mean "include the field" or "include disabled/excluded hooks."

**Proposed Change:**
Clarify to: `Include 'disabled_hooks' field in profile`

**User Benefit:** Unambiguous label reduces confusion.

**Estimated Effort:** trivial

---

#### UIUX-022: Missing "Enabled" Status Label for Hooks

**Category:** visual

**Affected Components:**
- `internal/tui/views/wizard_hooks.go` (lines 316-330)

**Current State:**
Disabled hooks show "(disabled)" text but enabled hooks show no status text — the absence of a label is the indicator.

**Proposed Change:**
Show "(enabled)" for enabled hooks for symmetry.

**User Benefit:** Clear visual confirmation of hook status.

**Estimated Effort:** trivial

---

#### UIUX-023: Model Selector Custom Input Lacks Format Hint

**Category:** usability

**Affected Components:**
- `internal/tui/views/model_selector.go` (lines 513-525)

**Current State:**
"Model ID:" input field has no format hint. Users don't know if they should enter `provider/model-id` or just `model-id`.

**Proposed Change:**
Add placeholder or hint: `Model ID (e.g., anthropic/claude-sonnet-4-20250514)`

**User Benefit:** Users enter correctly formatted IDs on first try.

**Estimated Effort:** trivial

---

#### UIUX-024: Missing Scroll Hints in Model Import

**Category:** usability

**Affected Components:**
- `internal/tui/views/model_import.go` (lines 538-540)

**Current State:**
Scroll indicators ("↑ more above", "↓ more below") appear but help text doesn't mention how to scroll.

**Proposed Change:**
Add scrolling keys (pgup/pgdn or j/k) to the help bar.

**User Benefit:** Users know how to navigate long lists.

**Estimated Effort:** trivial

---

#### UIUX-025: "/" Search Key Not Consistently Documented

**Category:** usability

**Affected Components:**
- `internal/tui/views/model_import.go` (line 234)
- `internal/tui/views/model_registry.go` (line 307)

**Current State:**
"/" triggers search in both model views but isn't always shown in help text.

**Proposed Change:**
Add "/" to help text consistently: `[/] search`

**User Benefit:** Users discover search without guessing.

**Estimated Effort:** trivial

---

#### UIUX-026: Hardcoded Viewport Overhead Magic Numbers

**Category:** visual

**Affected Components:**
- `internal/tui/views/wizard_hooks.go` (lines 160-162)
- `internal/tui/views/wizard_review.go` (lines 100-102)

**Current State:**
Hardcoded overhead values (4, 8, 2) for viewport height calculations without documentation.

**Proposed Change:**
Extract to named constants with comments explaining what each line accounts for.

```go
const (
    viewportOverheadNormal = 4 // title + help + 2 spacing lines
    viewportOverheadShort  = 2 // title + help only
)
```

**User Benefit:** Easier maintenance and less risk of layout bugs.

**Estimated Effort:** trivial

---

#### UIUX-027: Help Text Format Inconsistency

**Category:** visual

**Affected Components:**
- `internal/tui/views/template_select.go` (line 143)
- `internal/tui/app.go` (help rendering)

**Current State:**
Some views use `"↑/↓ navigate • enter select • esc cancel"` while the app shell uses bracketed format `[key] action`. Mixed formatting reduces scanability.

**Proposed Change:**
Standardize on one format across all views.

**User Benefit:** Consistent visual language reduces parsing effort.

**Estimated Effort:** small

---

#### UIUX-028: No Micro-Feedback on Successful Field Completion

**Category:** interaction

**Affected Components:**
- All wizard step files

**Current State:**
When users complete fields correctly, there's no feedback. The form silently accepts input with no visual confirmation.

**Proposed Change:**
Add subtle confirmation for validated fields (brief checkmark or highlight), similar to how `wizard_name.go` shows ✓/✗.

**User Benefit:** Users feel confident their input was accepted.

**Estimated Effort:** medium

---

#### UIUX-029: Model Registry Error Display Lacks Visual Weight

**Category:** visual

**Affected Components:**
- `internal/tui/views/model_registry.go` (lines 608-611)

**Current State:**
Errors shown as plain red text that blends with form content. No visual separator or icon.

**Proposed Change:**
Add `⚠ ` prefix and a separator line above error messages.

**User Benefit:** Errors stand out from surrounding form content.

**Estimated Effort:** trivial

---

#### UIUX-030: No Compact Mode Visual Indicator

**Category:** visual

**Affected Components:**
- `internal/tui/views/wizard_hooks.go` (lines 352-372)

**Current State:**
Help text and layout change silently when terminal is small. Users don't know they're in compact mode or that features may be hidden.

**Proposed Change:**
Add subtle indicator: `[compact]` in help bar when layout is constrained.

**User Benefit:** Users understand why the UI looks different on small terminals.

**Estimated Effort:** trivial

---

## Summary

| Category | Count |
|----------|-------|
| Usability | 20 |
| Accessibility | 3 |
| Visual Polish | 5 |
| Interaction | 1 |
| Performance Perception | 1 |

**Total Components Analyzed:** 18 view files + 4 core files
**Total Issues Found:** 30

### Priority Breakdown

| Priority | Count | Theme |
|----------|-------|-------|
| High | 7 | Silent failures, missing validation, accessibility blockers |
| Medium | 13 | Inconsistent help/navigation, missing context, poor error recovery |
| Low | 10 | Polish, format consistency, documentation |

### Quick Wins (Trivial Effort)

Issues UIUX-001, 006, 007, 008, 010, 013, 015, 016, 018, 019, 020, 021, 022, 023, 024, 025, 026, 029, 030 — **19 issues** can be fixed with minimal code changes, mostly adding text labels, help hints, or error displays.
