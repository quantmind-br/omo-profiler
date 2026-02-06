# UI/UX Improvements — Prioritized Implementation Plan

## Overview

**22 issues analyzed. 3 discarded. 4 adjusted. 15 kept.** Reordered by implementation phase: foundational infrastructure first, then quick wins, then incremental polish.

---

## Phase 1: Foundation (Do First)

These changes enable and de-risk everything else.

---

### UIUX-001: Consolidate Style Duplication

**Category:** visual | **Effort:** medium

**Problem:** 100+ inline style definitions across 14/20 view files. Every view re-creates identical styles using hardcoded hex colors instead of importing from `styles.go`. Changing a color in `styles.go` has zero effect on most views.

**Affected Components:**
- `internal/tui/views/dashboard.go` (9 inline styles)
- `internal/tui/views/diff.go` (6 duplicate colors + 9 duplicate styles — worst offender)
- `internal/tui/views/wizard_other.go` (30+ inline definitions)
- `internal/tui/views/wizard_categories.go` (8+ per function)
- `internal/tui/views/wizard_agents.go` (8+ per function)
- `internal/tui/views/wizard_hooks.go`, `wizard_name.go`, `wizard_review.go`, `wizard.go`
- `internal/tui/views/import.go`, `export.go`, `model_selector.go`, `list.go`

**Implementation:**
1. Add missing utility styles to `styles.go`: `DimStyle`, `LabelStyle`, `FormValueStyle`, `CheckboxEnabledStyle`, `CheckboxDisabledStyle`, `CodeStyle`, `SelectedItemStyle`, `CursorStyle`
2. Add missing colors: `DarkGray = "#45475A"`, `CodeBlue = "#89B4FA"`
3. Replace all inline definitions with imports from `styles.go`
4. Delete the `diff*` color variables from `diff.go`
5. Delete the package-level style vars from `dashboard.go`

**Why first:** Every visual change after this becomes a single-file edit instead of 14-file hunt.

---

### UIUX-009: Standardize Help Text Format

**Category:** visual | **Effort:** small

**Problem:** Three different help text formats coexist:
1. `"↑↓ navigate • enter select"` (separator format, global)
2. `"[enter] import  [esc] cancel"` (bracket format, import/export)
3. Embedded in views (template_select renders its own help as content)

**Affected Components:**
- `internal/tui/views/template_select.go`
- `internal/tui/views/import.go`
- `internal/tui/views/export.go`
- `internal/tui/views/wizard_hooks.go`
- `internal/tui/app.go` (global help bar)

**Implementation:**
Standardize ALL help text to the global separator format: `"key action • key action • key action"`. Remove view-embedded help text — let the global help system handle everything.

**Why first:** Doing this before adding new help text (from subsequent features) prevents introducing more inconsistency.

---

## Phase 2: Quick Wins (Trivial Effort, High Impact)

Each of these is under 20 lines of code and immediately improves usability.

---

### UIUX-006: Select All / Deselect All for Hooks and Model Import

**Category:** usability | **Effort:** trivial

**Problem:** Users must manually toggle each of 36 hooks one by one. No bulk action.

**Affected Components:**
- `internal/tui/views/wizard_hooks.go`
- `internal/tui/views/model_import.go`

**Implementation:**
Add keyboard shortcuts:
- `a` — Select all / Enable all
- `A` (Shift+A) — Deselect all / Disable all
Update help text: `"a all • A none • space toggle • ↑/↓ navigate"`

---

### UIUX-018: Show All Validation Errors

**Category:** usability | **Effort:** trivial

**Problem:** Wizard's save action shows only `validationErrors[0]`, forcing users into a fix-one-revalidate loop.

**Affected Components:**
- `internal/tui/views/wizard.go` (line 267)

**Implementation:**
Replace single error display with joined errors, or redirect to the review step's full error display.

---

### UIUX-003: Add Wizard Step Descriptions

**Category:** usability | **Effort:** trivial

**Problem:** Categories, agents, and other steps launch directly into forms with no explanatory text. Users unfamiliar with omo config don't know what they're configuring.

**Affected Components:**
- `internal/tui/views/wizard_categories.go`
- `internal/tui/views/wizard_agents.go`
- `internal/tui/views/wizard_other.go`

**Implementation:**
Add a brief description line below each step title:
- **Categories**: "Define model categories — groups of model settings that agents can reference"
- **Agents**: "Configure individual agent behavior — models, permissions, tools, and prompts"
- **Other Settings**: "Additional settings — disabled components, experimental features, and integrations"

---

### UIUX-019: Home/End Key Support for Long Lists

**Category:** usability | **Effort:** trivial

**Problem:** No list view supports Home/End keys. With 36 hooks or 14 agents, users must hold arrow keys.

**Affected Components:**
- `internal/tui/views/wizard_hooks.go`
- `internal/tui/views/wizard_categories.go`
- `internal/tui/views/wizard_agents.go`
- `internal/tui/views/wizard_other.go`
- `internal/tui/views/list.go`
- `internal/tui/views/model_registry.go`
- `internal/tui/views/model_import.go`

**Implementation:**
Add `Home` (go to first item), `End` (go to last item), and `g`/`G` (vim-style) key bindings to all list views.

---

### UIUX-014: Add Menu Item Icons

**Category:** visual | **Effort:** trivial

**Problem:** Menu items are plain text with no visual differentiation.

**Affected Components:**
- `internal/tui/views/dashboard.go` (menuItems)

**Implementation:**
Add Unicode symbols (avoid emoji for terminal compatibility):
```
  Switch Profile
  Create New
  Create from Template
  Edit Current
  Compare Profiles
  Manage Models
  Import Profile
  Export Profile
```

**Note:** Use simple Unicode symbols, not Nerd Font glyphs or emoji.

---

## Phase 3: Core Usability (Medium Effort, High Value)

---

### UIUX-004: Wizard Cancel Confirmation

**Category:** usability | **Effort:** medium

**Problem:** Ctrl+C or Esc on any wizard step immediately exits with NO confirmation. All progress is silently discarded.

**Affected Components:**
- `internal/tui/views/wizard.go` (line 178: Ctrl+C immediately cancels)
- `internal/tui/app.go` (line 412-413: WizardCancelMsg goes straight to dashboard)

**Implementation:**
Add confirmation dialog when canceling IF `step > 0` OR any field has been modified:
```
Discard unsaved changes?
Your profile configuration will be lost.

[Y] Yes, discard    [N] No, continue editing
```

On step 0 (Name) with empty input, exit immediately without confirmation.

---

### UIUX-005: Hook Grouping by Category

**Category:** usability | **Effort:** medium

**Problem:** 36 hooks presented as a flat, undifferentiated list. Hook names are technical and provide zero context.

**Affected Components:**
- `internal/tui/views/wizard_hooks.go`

**Implementation:**
Group hooks by category with section headers:
- **Core** (context-window-monitor, session-recovery, preemptive-compaction, ...)
- **Output Processing** (grep-output-truncator, tool-output-truncator, ...)
- **Injection** (directory-agents-injector, rules-injector, compaction-context-injector, ...)
- **Agent Behavior** (ralph-loop, category-skill-reminder, sisyphus-junior-notepad, ...)
- **Recovery** (edit-error-recovery, delegate-task-retry, anthropic-context-window-limit-recovery, ...)
- **Other** (auto-update-checker, startup-toast, keyword-detector, ...)

**Note:** Skip per-hook descriptions — they're a maintenance burden that drifts from upstream. The grouping alone provides 80% of the value by reducing cognitive load from 36 flat items to ~6 logical groups. Format hook names readably (replace hyphens with spaces, capitalize).

---

### UIUX-002: Dashboard Visual Framing

**Category:** visual | **Effort:** small

**Problem:** Dashboard renders as plain text — no borders, no panels, no visual hierarchy. It's the first screen users see and feels unfinished compared to other views.

**Affected Components:**
- `internal/tui/views/dashboard.go`

**Implementation:**
Add a bordered status card around the profile info area only (not the entire dashboard):
```
+---------------------------------+
|  omo-profiler                   |
|  Profile manager for omo        |
|                                 |
|  Active: my-profile             |
|  3 profiles available           |
+---------------------------------+

  > Switch Profile
    Create New
    Create from Template
    ...
```

Keep menu items frameless for breathing room.

---

### UIUX-021: Search/Filter in Hooks Step

**Category:** usability | **Effort:** small

**Problem:** 36 hooks with no search. Other list views (Model Registry, List, Model Import) all support `/` search.

**Affected Components:**
- `internal/tui/views/wizard_hooks.go`

**Implementation:**
Add `/` search mode to filter hooks by name substring, consistent with other list views. Pattern exists in model_registry.go and list.go.

---

## Phase 4: Polish (Low-Medium Effort, Medium Value)

---

### UIUX-008: Import/Export Path Defaults

**Category:** usability | **Effort:** small

**Problem:** Both views show an empty text input. Users must type full path from scratch.

**Affected Components:**
- `internal/tui/views/import.go`
- `internal/tui/views/export.go`

**Implementation:**
- **Export**: Pre-fill with `~/[profile-name].json` as default value (user can edit)
- **Import**: Show CWD as context above the input. Don't pre-fill — users import from varied locations.

---

### UIUX-011: Responsive Wizard Progress Bar

**Category:** visual | **Effort:** small

**Problem:** Progress bar is ~90 chars wide with no width check. Overflows on terminals < 80 columns.

**Affected Components:**
- `internal/tui/views/wizard.go` (renderHeader)

**Implementation:**
1. At width < 80: Abbreviated step names: `Name -> Cat -> Agents -> Hooks -> Other -> Review`
2. At width < 50: Step numbers only: `[1] -> [2] -> [3] -> [4] -> [5] -> [6]`
3. Always show current step name in full below the progress indicator

---

### UIUX-012: Template Selection Preview

**Category:** usability | **Effort:** medium

**Problem:** Template selection shows only profile names. Users can't see contents before choosing.

**Affected Components:**
- `internal/tui/views/template_select.go`

**Implementation:**
Show a one-line summary for the currently highlighted profile (loaded on cursor move, not all at once):
`"3 categories, 5 agents, 12 hooks enabled"`

---

### UIUX-015: Dashboard Number Shortcuts

**Category:** usability | **Effort:** small

**Problem:** Only 2/8 menu items (import, export) have keyboard shortcuts.

**Affected Components:**
- `internal/tui/views/dashboard.go`

**Implementation:**
Add numbered shortcuts `1-8` for direct menu access. Use numbers (not letters) to avoid conflicts with future search/filter.

---

### UIUX-016: Contextual Loading Text

**Category:** interaction | **Effort:** small

**Problem:** Loading spinner always shows generic "Loading..." regardless of operation.

**Affected Components:**
- `internal/tui/app.go` (line 602)

**Implementation:**
Pass context-specific text to the loading overlay:
- "Switching to dev-profile..."
- "Deleting profile..."
- "Importing profile..."
- "Exporting to ~/profile.json..."

---

### UIUX-007: Verify Diff View Esc Navigation

**Category:** usability | **Effort:** trivial

**Problem:** Report claims Esc traps users in diff view, but app.go lines 164-173 appear to handle Esc at app level. Needs verification.

**Affected Components:**
- `internal/tui/app.go`
- `internal/tui/views/diff.go`

**Implementation:**
1. Test actual behavior — does Esc return to dashboard from diff normal mode?
2. If already working: update help text to show "esc back"
3. If trapped: add Esc handler in diff view's normal mode to emit `NavigateToDashboardMsg`

---

### UIUX-013: Dropdown Cycle Indicator

**Category:** interaction | **Effort:** trivial

**Problem:** Dropdowns already show `[Enter]`/`[<-/->]` hints, but don't visually indicate "there are more options."

**Affected Components:**
- `internal/tui/views/wizard_categories.go`
- `internal/tui/views/wizard_agents.go`
- `internal/tui/views/wizard_other.go`

**Implementation:**
Add a triangle indicator after the current value: `reasoning_effort: low v [<-/->]`. This signals more options exist without listing them all inline.

---

### UIUX-022: Better Import Error Messages

**Category:** usability | **Effort:** small

**Problem:** Import validation failure shows generic "validation failed" message.

**Affected Components:**
- `internal/tui/views/import.go`
- `internal/tui/app.go` (doImportProfile)

**Implementation:**
On validation failure, show specific errors: "Invalid JSON: unexpected token at position X" or "Schema error: missing required field 'categories'". Skip the file extension warning — it's over-protective.

---

## Discarded Features

| ID | Feature | Reason |
|----|---------|--------|
| UIUX-010 | Profile switch highlight/flash effect | Toast + dashboard refresh is standard and adequate. Flash effects in terminals are jarring and unreliable across emulators. Adds animation state complexity for marginal benefit. |
| UIUX-017 | Diff view line numbers | This is a profile comparison tool, not a code review tool. JSON profiles are 50-200 lines. `+`/`-` markers with color are sufficient. Line numbers add width overhead in already space-constrained panes. |
| UIUX-020 | Diff empty panes on entry | Auto-selecting first two profiles is a GOOD default. Starting with empty panes forces extra steps. The user can change selections easily — friction added, value removed. |

---

## Summary

| Phase | Items | Effort | Impact |
|-------|-------|--------|--------|
| **1: Foundation** | 2 | Medium + Small | Enables everything else |
| **2: Quick Wins** | 5 | All trivial | Immediate usability boost |
| **3: Core Usability** | 4 | Medium + Small | Biggest user-facing improvements |
| **4: Polish** | 8 | Small-Medium | Refinement and consistency |

**Total: 19 features across 4 phases.**
**Removed: 3 features (UIUX-010, UIUX-017, UIUX-020).**
**Adjusted: 4 features (UIUX-005, UIUX-007, UIUX-013, UIUX-022).**
