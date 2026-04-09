# Code Quality & Refactoring Analysis Report

## Executive Summary

The omo-profiler codebase (84 Go files, ~33,500 lines) is a well-structured TUI application for managing oh-my-opencode profiles. The core packages (`config`, `profile`, `schema`, `diff`, `models`) are clean and well-organized. However, the TUI layer — specifically the wizard views — contains severe maintainability issues. Three files alone account for 7,830 lines (23% of the codebase) and exhibit God Object anti-patterns, massive code duplication, and extreme function lengths. The model-related views also share significant duplicated list rendering logic.

**Key Metrics:**
- 3 files over 1,000 lines (wizard_other: 3,768, wizard_agents: 2,717, wizard_categories: 1,345)
- 1 function at 1,422 lines (`wizard_other.Update()`)
- ~165 lines of exact duplicate code (`Apply()` vs `applyAllAgentFields()`)
- 8 nesting levels deep in `wizard_other.Update()`
- ~30 near-identical textinput handler blocks (~750 lines of boilerplate)

**Priority:** The wizard files are the clear bottleneck for maintainability. Everything else is in good shape.

---

## Critical & Major Issues

### CQ-001: wizard_other.go is a 3,768-line God Object

**Category:** large_files
**Severity:** critical
**Best Practice:** Single Responsibility Principle, Extract Class

**Affected Files:**
- `internal/tui/views/wizard_other.go` (3,768 lines)

**Current State:**
The file handles all "Other Settings" configuration — 28 config sections with 70+ struct fields. The `Update()` function alone is 1,422 lines with nesting up to 8 levels deep. `SetConfig()` (398 lines) and `Apply()` (477 lines) are mirror images of each other with repetitive nil-check patterns. ~30 textinput handler blocks each repeat 25-35 lines of nearly identical Focus/Blur/Update logic (~750 lines of boilerplate).

**Proposed Change:**
Split into 5 files and extract repetitive patterns:

```
wizard_other.go          → types, struct, Init, View (~300 lines)
wizard_other_config.go   → SetConfig, Apply (~500 lines, de-duplicated)
wizard_other_update.go   → Update with extracted handlers (~600 lines)
wizard_other_render.go   → renderContent, renderSubSection (~250 lines)
wizard_other_fields.go   → field handler map, textinput factory (~200 lines)
```

Extract textinput handling into a data-driven dispatch:
```go
// Current: 30 near-identical blocks of 25-35 lines each (~750 lines)
if w.currentSection == sectionX && w.subCursor == N {
    switch msg.String() {
    case "esc": w.fieldName.Blur(); w.inSubSection = false; return w, nil
    case "up", "k": ...
    // ... identical for every field
    }
}

// Proposed: single dispatch (~30 lines + data table)
type fieldBinding struct {
    section  otherSection
    cursor   int
    input    *textinput.Model
}
fields := []fieldBinding{
    {sectionDCP, 0, &w.dcpTurnProtTurns},
    {sectionDCP, 1, &w.dcpProtectedTools},
    // ...
}
```

**Breaking Change:** No (internal refactor)
**Prerequisites:** None
**Estimated Effort:** large

---

### CQ-002: wizard_agents.go has 165+ lines of exact duplicate code

**Category:** duplication
**Severity:** critical
**Best Practice:** DRY (Don't Repeat Yourself)

**Affected Files:**
- `internal/tui/views/wizard_agents.go` (2,717 lines)

**Current State:**
`Apply()` (lines 808-1023, 216 lines) and `applyAllAgentFields()` (lines 1025-1180, 156 lines) contain nearly identical logic. The only difference is that `Apply()` checks field selection before writing. Additionally, `viewport.SetContent(w.renderContent())` is called 50+ times throughout the file.

**Proposed Change:**
```go
// Current: two separate functions with identical field mapping
func (w *WizardAgents) Apply(cfg *config.Config, sel *profile.FieldSelection) { ... }
func (w *WizardAgents) applyAllAgentFields(cfg *config.Config) { ... }

// Proposed: single function with optional selection parameter
func (w *WizardAgents) applyAgentFields(cfg *config.Config, sel *profile.FieldSelection) {
    // sel == nil means apply all fields without filtering
}

// Also extract viewport refresh
func (w *WizardAgents) refreshView() {
    w.viewport.SetContent(w.renderContent())
}
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** small

---

### CQ-003: wizard_agents.go Update() at 312 lines with 7-level nesting

**Category:** complexity
**Severity:** major
**Best Practice:** Extract Method, Reduce Nesting

**Affected Files:**
- `internal/tui/views/wizard_agents.go` (lines 1615-1927)

**Current State:**
The `Update()` function handles 5 different editing sub-modes (model selection, custom model save, provider opts, fallback models, bash perms) plus the main form navigation — all in a single function with 7-level nesting and 60+ switch cases.

**Proposed Change:**
The editor handlers are already partially extracted (`handleFallbackModelsEditor`, `handleProviderOptsEditor`, `handleBashPermsEditor`, `handleSaveCustomModel`), but the dispatch and main form handling remain monolithic. Extract the main form key handling into `handleFormNavigation()` and early-return from sub-mode handlers:

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
    // ...
    }
}
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** medium

---

### CQ-004: app.go Update() is a 350-line monolith

**Category:** complexity
**Severity:** major
**Best Practice:** Extract Method, Table-Driven Design

**Affected Files:**
- `internal/tui/app.go` (lines 141-490, 350 lines)

**Current State:**
The main app `Update()` function handles 20+ message types and 10 state-based view delegations. Navigation handlers (lines 220-373) repeat the same 3-step pattern (create view, SetSize, navigateTo) ~6 times. The view delegation switch (lines 447-487) has 10 near-identical cases. Error toast pattern (`showToast("...failed: "+err.Error(), toastError, 3*time.Second)`) repeats ~10 times.

**Proposed Change:**
Extract message handlers into named methods and use a view manager pattern:

```go
// Current: 10 identical delegation cases
case stateDashboard:
    a.dashboard, cmd = a.dashboard.Update(msg)
    cmds = append(cmds, cmd)
case stateList:
    a.list, cmd = a.list.Update(msg)
    cmds = append(cmds, cmd)
// ... 8 more identical cases

// Proposed: interface-based delegation
type updatable interface {
    Update(tea.Msg) (tea.Model, tea.Cmd)
}
func (a *App) activeView() updatable { ... }
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** medium

---

### CQ-005: wizard.go nextStep()/prevStep() step transition duplication

**Category:** duplication
**Severity:** major
**Best Practice:** Table-Driven Design, DRY

**Affected Files:**
- `internal/tui/views/wizard.go` (lines 263-414)

**Current State:**
`nextStep()` (114 lines) contains 4 near-identical step transition cases (80% duplicate):
```go
case StepCategories: w.categoriesStep.Apply(...); w.step = StepAgents; w.agentsStep.SetConfig(...); return w, w.agentsStep.Init()
case StepAgents:     w.agentsStep.Apply(...);     w.step = StepHooks;  w.hooksStep.SetConfig(...);  return w, w.hooksStep.Init()
case StepHooks:      w.hooksStep.Apply(...);      w.step = StepOther;  w.otherStep.SetConfig(...);  return w, w.otherStep.Init()
case StepOther:      w.otherStep.Apply(...);       w.step = StepReview; ...
```
`prevStep()` (37 lines) mirrors this with identical Apply + flashMsg patterns. The save closure inside `nextStep()` (39 lines) duplicates the validation logic from lines 299-315.

**Proposed Change:**
Define a step interface and use a table-driven approach:

```go
type wizardStep interface {
    Apply(cfg *config.Config, sel *profile.FieldSelection)
    SetConfig(cfg *config.Config, sel *profile.FieldSelection)
    Init() tea.Cmd
}

var stepOrder = []struct {
    step    Step
    view    func(w *Wizard) wizardStep
}{
    {StepCategories, func(w *Wizard) wizardStep { return &w.categoriesStep }},
    {StepAgents,     func(w *Wizard) wizardStep { return &w.agentsStep }},
    // ...
}
```

Extract validation into `validateConfig()` method to eliminate the duplication between pre-save and in-closure validation.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** medium

---

### CQ-006: wizard_categories.go follows same patterns at 1,345 lines

**Category:** large_files
**Severity:** major
**Best Practice:** Single Responsibility Principle

**Affected Files:**
- `internal/tui/views/wizard_categories.go` (1,345 lines)

**Current State:**
Similar structural issues to wizard_agents.go: 18-field `categoryConfig` struct, repeated Focus/Blur switch blocks (14 cases each in `updateFieldFocus()`), hardcoded line-to-field mappings, and the same textinput initialization boilerplate repeated 11 times.

**Proposed Change:**
Split into `wizard_categories.go` (core logic + Update) and `wizard_categories_config.go` (SetConfig/Apply/selection paths). Extract shared textinput factory and field focus patterns — these are common across all three wizard files.

**Breaking Change:** No
**Prerequisites:** CQ-001 (extract shared patterns first)
**Estimated Effort:** medium

---

### CQ-007: Duplicated list rendering across model views

**Category:** duplication
**Severity:** major
**Best Practice:** DRY, Extract Component

**Affected Files:**
- `internal/tui/views/model_import.go` (654 lines)
- `internal/tui/views/model_registry.go` (649 lines)
- `internal/tui/views/model_selector.go` (540 lines)

**Current State:**
All three files implement near-identical list pagination logic: cursor/offset management, `visibleHeight` calculation, scroll indicator rendering, and filtered item retrieval. The scroll calculation pattern, offset boundary checks, and list item rendering loops are functionally identical with only variable names changed (`offset` vs `scrollOffset`).

**Proposed Change:**
Extract a shared `ListRenderer` component:

```go
// internal/tui/views/list_renderer.go
type ListRenderer struct {
    cursor       int
    offset       int
    visibleHeight int
}

func (lr *ListRenderer) EnsureVisible()
func (lr *ListRenderer) ScrollUp()
func (lr *ListRenderer) ScrollDown()
func (lr *ListRenderer) RenderScrollIndicator(style lipgloss.Style) string
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** medium

---

## Minor Issues

### CQ-008: String-based error checking in model views

**Category:** code_smells
**Severity:** minor
**Best Practice:** Use typed errors

**Affected Files:**
- `internal/tui/views/model_import.go` (line 393)
- `internal/tui/views/model_registry.go` (lines 443, 450)

**Current State:**
```go
if strings.Contains(err.Error(), "already exists") {
    // handle duplicate error
}
```

**Proposed Change:**
Define a custom error type in the models package:
```go
type ModelExistsError struct { ID string }
func (e *ModelExistsError) Error() string { ... }

// Usage:
if errors.As(err, &ModelExistsError{}) { ... }
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** small

---

### CQ-009: Hardcoded field path lists in profile package

**Category:** code_smells
**Severity:** minor
**Best Practice:** Generate from schema or struct tags

**Affected Files:**
- `internal/profile/selection.go` (lines 8-164): `allFieldPaths` with 156 hardcoded entries
- `internal/profile/profile.go` (lines 24-77): `knownConfigTags`, `knownFieldPaths`, `knownFieldPathPrefixes`

**Current State:**
156 field paths are manually maintained as string literals. Adding a new config field requires updating multiple hardcoded lists across files — error-prone and easy to forget.

**Proposed Change:**
Generate field paths from struct reflection or maintain a single source of truth that the lists derive from. Could use `go generate` with a custom tool that reads `config.Config` struct tags.

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** medium

---

### CQ-010: Inconsistent scroll variable naming

**Category:** naming
**Severity:** minor
**Best Practice:** Consistent naming conventions

**Affected Files:**
- `internal/tui/views/model_import.go`: uses `offset`, `providerOffset`
- `internal/tui/views/model_registry.go`: uses `offset`
- `internal/tui/views/model_selector.go`: uses `scrollOffset`

**Proposed Change:**
Standardize on `scrollOffset` across all components (or whichever name the shared `ListRenderer` uses after CQ-007).

**Breaking Change:** No
**Prerequisites:** CQ-007
**Estimated Effort:** trivial

---

### CQ-011: Repeated textinput initialization boilerplate

**Category:** duplication
**Severity:** minor
**Best Practice:** Factory Pattern

**Affected Files:**
- `internal/tui/views/wizard_other.go` (lines 712-933): 35+ textinput instantiations
- `internal/tui/views/wizard_agents.go` (lines 275-390): 35+ textinput instantiations
- `internal/tui/views/wizard_categories.go` (lines 169-238): 11+ textinput instantiations

**Current State:**
```go
field := textinput.New()
field.Placeholder = "some-value"
field.Width = 30
// repeated 80+ times across wizard files
```

**Proposed Change:**
```go
func newTextInput(placeholder string, width int) textinput.Model {
    ti := textinput.New()
    ti.Placeholder = placeholder
    ti.Width = width
    return ti
}
```

**Breaking Change:** No
**Prerequisites:** None
**Estimated Effort:** small

---

## Suggestions

### CQ-012: Shared wizard step interface

**Category:** structure
**Severity:** suggestion

All wizard step views (categories, agents, hooks, other) share the same method signatures: `Init()`, `SetSize()`, `SetConfig()`, `Apply()`, `Update()`, `View()`. Defining a formal `WizardStep` interface would enable table-driven step management and reduce boilerplate in `wizard.go`.

**Estimated Effort:** medium

---

### CQ-013: Extract help rendering from app.go

**Category:** structure
**Severity:** suggestion

`renderShortHelp()` (47 lines) and `renderFullHelp()` (86 lines) in app.go share state-specific hint definitions. Adding a new state requires updating both. Could consolidate into a single state-to-hints map.

**Estimated Effort:** small

---

### CQ-014: Split config/types.go by domain

**Category:** structure
**Severity:** suggestion

`internal/config/types.go` (325 lines) contains 31 nested types for the entire configuration schema. While not large enough to be urgent, grouping related types (agent config, experimental config, background task config) into separate files would improve navigability.

**Estimated Effort:** small

---

## Code Metrics

| Metric | Value | Status |
|--------|-------|--------|
| Files > 500 lines | 7 (source) + 6 (test) | Needs attention |
| Files > 1000 lines | 3 (source) + 3 (test) | Needs attention |
| Functions > 50 lines | ~15 | Needs attention |
| Functions > 200 lines | 4 | Needs attention |
| Functions > 1000 lines | 1 (Update at 1,422) | Critical |
| Max nesting depth | 8 levels | Critical |
| Duplicate code blocks | ~400+ lines | Needs attention |
| Functions > 5 parameters | 0 | Good |
| TODO/FIXME comments | 0 | Good |
| Unused imports | 0 | Good |

## Summary

| Severity | Count |
|----------|-------|
| Critical | 2 |
| Major | 5 |
| Minor | 4 |
| Suggestion | 3 |

| Category | Count |
|----------|-------|
| Large Files | 3 |
| Code Smells | 2 |
| Complexity | 2 |
| Duplication | 4 |
| Naming | 1 |
| Structure | 2 |
| Types | 0 |
| Dead Code | 0 |

**Total Files Analyzed:** 84
**Total Issues Found:** 14

## Recommended Refactoring Order

1. **CQ-011** (trivial) - Extract textinput factory — immediate win, shared across all wizard files
2. **CQ-002** (small) - Merge duplicate Apply functions in wizard_agents — exact duplication removal
3. **CQ-008** (small) - Replace string error checks — quick safety improvement
4. **CQ-007** (medium) - Extract shared ListRenderer — reduces 3 files at once
5. **CQ-005** (medium) - Table-driven wizard step transitions — reduces wizard.go complexity
6. **CQ-003** (medium) - Extract wizard_agents Update handlers — nesting reduction
7. **CQ-004** (medium) - Extract app.go message handlers — readability improvement
8. **CQ-001** (large) - Split wizard_other.go — the biggest payoff but highest effort
9. **CQ-006** (medium) - Split wizard_categories.go — benefits from patterns established in CQ-001
10. **CQ-012** (medium) - Formal WizardStep interface — enables further simplification
