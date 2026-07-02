# Agent TUI Full Configuration Support

## TL;DR

> **Quick Summary**: Make all 23 agent schema properties fully configurable via the TUI wizard, fixing missing dropdown values, adding missing UI fields, and replacing read-only displays with interactive editors for `providerOptions`, `permission.bash`, and `fallback_models`.
> 
> **Deliverables**:
> - Fixed `reasoningEffort` dropdown with all 6 schema values (`none`, `minimal`, `low`, `medium`, `high`, `xhigh`)
> - New `allow_non_gpt_model` checkbox shown only for `hephaestus` agent
> - Editable key-value editor for `providerOptions` (replaces read-only display)
> - Per-command rule editor for `permission.bash` object format (replaces read-only display)
> - Structured fallback models editor with model selector (replaces raw JSON textinput)
> - Inline validation for `color` (hex pattern), `temperature` (0-2), `top_p` (0-1)
> - Round-trip tests for all new/modified fields
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 2 waves + sequential chain
> **Critical Path**: Task 1 → Task 3 → Task 4 → Task 5 → Task 6 → Final

---

## Context

### Original Request
User identified 6 gaps between the upstream JSON schema for `agents` and the TUI's editing capabilities. The schema defines 23 properties per agent, but the TUI has incomplete dropdown values (reasoningEffort), missing fields (allow_non_gpt_model), read-only displays (providerOptions, permission.bash), fragile raw JSON input (fallback_models), and no inline validation (color, temperature, top_p).

### Interview Summary
**Key Discussions**:
- Scope: All 6 improvements (P0 through P3) — user confirmed
- Test strategy: Basic tests for SetConfig/Apply round-trips — user confirmed
- Go types (`config/types.go`) already match schema 100% — no types.go changes needed
- All modifications stay within TUI layer (`internal/tui/views/`)

**Research Findings**:
- `effortLevels` is defined in `wizard_categories.go:21` and shared by both wizard_agents and wizard_categories steps (9 references across 2 files)
- `getLineForField` uses hardcoded offsets (lines 777-808) and a fixed form height of `43` (line 770) — must be updated when adding fields
- `providerOptions` currently preserved implicitly via struct reuse in Apply (line 558-559) — needs explicit Apply logic once editable
- `permission.bash` preserves objects via `originalBash` passthrough (line 630-632) — needs replacement with proper editor state
- `fallback_models` schema supports `string | []string | []ModelObject` where ModelObject has 7 fields
- Existing test patterns in `wizard_agents_test.go`: `TestAgentApplyPreservesExistingFields` (line 532), `TestAgentApplyPreservesProviderOptions` (line 715)

### Metis Review
**Identified Gaps** (addressed):
- `effortLevels` is shared — change in wizard_categories.go affects both steps (verified: both have dedicated tests)
- `getLineForField` fragility — every field addition must update both `fieldOffsets` map AND form height constant `43`
- `AllowNonGptModel` data loss on new profile creation — once field has UI control, explicit Apply logic needed
- Sub-editor cancel behavior — must define atomic commit/rollback for providerOptions/bash/fallback editors
- File size growth (~1386 → ~1800 lines) — within project convention (wizard_other.go is 2460 lines)

---

## Work Objectives

### Core Objective
Make all 23 agent schema properties fully configurable via the TUI wizard, eliminating all read-only fields and missing values.

### Concrete Deliverables
- Modified `internal/tui/views/wizard_categories.go` — updated `effortLevels` array
- Modified `internal/tui/views/wizard_agents.go` — all 6 improvements integrated
- New/updated tests in `internal/tui/views/wizard_agents_test.go`
- All existing tests pass (`make test`)
- Lint passes (`make lint`)

### Definition of Done
- [x] `make test` passes with 0 failures
- [x] `make lint` passes with 0 errors
- [x] All 23 agent schema properties are editable in the TUI wizard
- [x] Round-trip tests prove: config → SetConfig → Apply → config matches original

### Must Have
- `reasoningEffort` dropdown includes `none` and `minimal`
- `allow_non_gpt_model` checkbox visible only on `hephaestus`
- `providerOptions` is editable (add/remove/edit key-value pairs)
- `permission.bash` object rules are editable (add/remove/edit per-command permissions)
- `fallback_models` has structured model picker (not raw JSON only)
- `color`, `temperature`, `top_p` show inline validation feedback
- All new fields have SetConfig/Apply round-trip tests

### Must NOT Have (Guardrails)
- NO changes to `internal/config/types.go` — types are already complete
- NO changes to `internal/schema/schema.json` — schema is upstream authority
- NO validation added to `wizard_categories.go` fields (separate scope)
- NO left/right arrow key behavior changes — currently used for expand/collapse in navigation mode
- NO removal of the `""` (empty/not-set) entry from dropdown arrays — serves as "unset" default at index 0
- NO new external dependencies
- NO `//nolint` or type assertion shortcuts
- NO excessive comments or docstring bloat (AI slop)
- NO changes to wizard_hooks.go, wizard_other.go, or other wizard steps

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (`wizard_agents_test.go` — 814 lines, 7+ test functions)
- **Automated tests**: Tests-after (basic round-trip tests for new/modified fields)
- **Framework**: Go standard `testing` + `github.com/stretchr/testify` assertions

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **TUI Interaction**: Use `interactive_bash` (tmux) — build binary, launch TUI, navigate to agent step, verify field rendering
- **Tests**: Use Bash — `go test -run TestName -v ./internal/tui/views/`
- **Build**: Use Bash — `make test && make lint`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Parallel — different files):
├── Task 1: Fix effortLevels in wizard_categories.go [quick]
└── Task 2: Add allow_non_gpt_model + inline validations in wizard_agents.go [quick]

Wave 2 (Sequential — same file, each depends on previous):
├── Task 3: providerOptions KV editor (depends: 2) [unspecified-high]
├── Task 4: permission.bash granular editor (depends: 3) [unspecified-high]
└── Task 5: fallback_models structured editor (depends: 4) [deep]

Wave 3 (After Wave 2):
└── Task 6: Comprehensive tests + build verification (depends: 5) [quick]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: Task 2 → Task 3 → Task 4 → Task 5 → Task 6 → F1-F4 → user okay
Parallel Speedup: ~20% (limited by same-file constraint in Wave 2)
Max Concurrent: 2 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 6 | 1 |
| 2 | — | 3 | 1 |
| 3 | 2 | 4 | 2 |
| 4 | 3 | 5 | 2 |
| 5 | 4 | 6 | 2 |
| 6 | 1, 5 | F1-F4 | 3 |

### Agent Dispatch Summary

- **Wave 1**: 2 tasks — T1 → `quick`, T2 → `quick`
- **Wave 2**: 3 tasks — T3 → `unspecified-high`, T4 → `unspecified-high`, T5 → `deep`
- **Wave 3**: 1 task — T6 → `quick`
- **FINAL**: 4 tasks — F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Fix `reasoningEffort` dropdown — add missing schema values

  **What to do**:
  - Edit the shared `effortLevels` array in `internal/tui/views/wizard_categories.go:21`
  - Current: `var effortLevels = []string{"", "low", "medium", "high", "xhigh"}`
  - New: `var effortLevels = []string{"", "none", "minimal", "low", "medium", "high", "xhigh"}`
  - Insert `"none"` at index 1 and `"minimal"` at index 2 (push existing values right)
  - Verify all 9 existing references across `wizard_agents.go` (4 refs) and `wizard_categories.go` (5 refs) still work correctly with the new indices
  - Update any existing tests in `wizard_categories_test.go` and `wizard_agents_test.go` that hardcode `effortLevels` index values (these will break because indices shifted)
  - Add a test case verifying round-trip for `reasoningEffort = "none"` and `reasoningEffort = "minimal"` through SetConfig → Apply

  **Must NOT do**:
  - Do NOT remove the `""` entry at index 0 — it serves as "unset" default
  - Do NOT modify any other dropdown arrays (modes, thinkingTypes, permissionValues, verbosityLevels)
  - Do NOT touch wizard_agents.go — this task ONLY modifies wizard_categories.go and test files

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-line change + test index updates — trivial scope
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `git-master`: Not needed — simple file edit

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Task 6 (comprehensive tests)
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_categories.go:21` — Current `effortLevels` definition (THE line to change)
  - `internal/tui/views/wizard_categories.go:351` — SetConfig loop using `effortLevels` index mapping
  - `internal/tui/views/wizard_categories.go:447` — Apply writing `effortLevels[idx]` back to config
  - `internal/tui/views/wizard_categories.go:673` — Dropdown cycling `(idx+1) % len(effortLevels)`
  - `internal/tui/views/wizard_categories.go:911` — Render using `renderDropdown` with `effortLevels`
  - `internal/tui/views/wizard_agents.go:525` — SetConfig loop matching effort string to index
  - `internal/tui/views/wizard_agents.go:691` — Apply writing `effortLevels[idx]` to config
  - `internal/tui/views/wizard_agents.go:934` — Dropdown cycling
  - `internal/tui/views/wizard_agents.go:1187` — Render

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go:532-645` — `TestAgentApplyPreservesExistingFields` — uses `ReasoningEffort: "high"`, index must still map correctly
  - `internal/tui/views/wizard_categories_test.go` — Any test that sets `reasoningEffortIdx` or checks `ReasoningEffort` value

  **WHY Each Reference Matters**:
  - The `effortLevels` array is the SINGLE SOURCE — all consumers index into it. Inserting 2 values at positions 1-2 shifts `"low"` from index 1 to index 3, `"medium"` from 2 to 4, etc. Any test hardcoding old indices will fail.

  **Acceptance Criteria**:
  - [x] `effortLevels` array has exactly 7 entries: `["", "none", "minimal", "low", "medium", "high", "xhigh"]`
  - [x] `go test -run TestWizardCategories -v ./internal/tui/views/` → PASS
  - [x] `go test -run TestAgent -v ./internal/tui/views/` → PASS

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: ReasoningEffort round-trip with new value "none"
    Tool: Bash
    Preconditions: Project compiles successfully
    Steps:
      1. Run `go test -run TestAgentApplyPreservesExistingFields -v ./internal/tui/views/` — verify existing round-trip still works with shifted indices
      2. Verify new test exists that creates config with `ReasoningEffort: "none"`, calls SetConfig → Apply, asserts output == "none"
      3. Same for `ReasoningEffort: "minimal"`
    Expected Result: All tests pass with 0 failures
    Failure Indicators: Test fails with "expected 'high', got 'xhigh'" (index shift not handled)
    Evidence: .sisyphus/evidence/task-1-effort-roundtrip.txt

  Scenario: effortLevels array has correct length and order
    Tool: Bash
    Preconditions: Change applied to wizard_categories.go
    Steps:
      1. Run `grep -n 'effortLevels' internal/tui/views/wizard_categories.go` — verify line 21 has all 7 values
      2. Run `go vet ./internal/tui/views/` — no issues
    Expected Result: Array is `["", "none", "minimal", "low", "medium", "high", "xhigh"]`
    Failure Indicators: Missing value or wrong order
    Evidence: .sisyphus/evidence/task-1-effort-array.txt
  ```

  **Commit**: YES
  - Message: `fix(agents): add missing reasoningEffort values none and minimal`
  - Files: `internal/tui/views/wizard_categories.go`, test files
  - Pre-commit: `go test -run TestWizardCategories -v ./internal/tui/views/ && go test -run TestAgent -v ./internal/tui/views/`

---

- [x] 2. Add `allow_non_gpt_model` checkbox + inline field validation

  **What to do**:
  **Part A — allow_non_gpt_model:**
  - Add `fieldAllowNonGpt` to the `agentFormField` enum in `wizard_agents.go:97-130` (after `fieldCompactionVariant`)
  - Add `allowNonGpt bool` field to the `agentConfig` struct (line ~147)
  - In `SetConfig()` (line ~411-544): when loading hephaestus config, read `agentCfg.AllowNonGptModel` into `ac.allowNonGpt`
  - In `Apply()` (line ~546-701): when writing hephaestus config, write `ac.allowNonGpt` to `agentCfg.AllowNonGptModel` as `*bool` (only for hephaestus)
  - In `renderAgentForm()` (line ~1113-1232): add `renderBool("allow_non_gpt", fieldAllowNonGpt, ac.allowNonGpt)` ONLY when agent name == `"hephaestus"` — place after compactionVariant, before the closing empty line
  - In field navigation (focusedField++ / focusedField--): skip `fieldAllowNonGpt` when agent is NOT hephaestus
  - In `updateFieldFocus()` (line ~703): handle focus/blur for the new field
  - Update `getLineForField()` (line ~777-808): add `fieldAllowNonGpt: 36` entry; update expanded form height from `43` to `44` (line 770) for hephaestus; keep `43` for other agents OR use dynamic calculation

  **Part B — inline validation:**
  - Add validation feedback in `renderAgentForm()` for 3 fields:
    - `color`: If non-empty and doesn't match `^#[0-9A-Fa-f]{6}$` → append red " ✗ invalid hex" after field value
    - `temperature`: If non-empty and not a float in [0, 2] → append red " ✗ must be 0-2"
    - `top_p`: If non-empty and not a float in [0, 1] → append red " ✗ must be 0-1"
  - Validation is DISPLAY-ONLY — does not block form submission. Invalid values are still accepted (wizard_review does the hard validation against schema).
  - Use `styles.ErrorStyle` or inline `lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Red))` for red text

  **Must NOT do**:
  - Do NOT show `allow_non_gpt_model` on ANY agent other than hephaestus
  - Do NOT add validation to wizard_categories.go fields
  - Do NOT block form submission on validation errors (display-only hints)
  - Do NOT change field ordering of existing fields

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Adding one boolean field + 3 inline validation hints — small scope, clear pattern
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Task 3
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:97-130` — `agentFormField` enum — add `fieldAllowNonGpt` at end
  - `internal/tui/views/wizard_agents.go:133-179` — `agentConfig` struct — add `allowNonGpt bool` field
  - `internal/tui/views/wizard_agents.go:1150-1162` — `renderBool()` function — EXACT pattern to follow for checkbox rendering
  - `internal/tui/views/wizard_agents.go:1227-1228` — compactionVariant render (last field before closing empty line) — insert `allow_non_gpt` render after this, conditional on `name == "hephaestus"`
  - `internal/tui/views/wizard_agents.go:770` — form height constant `43` — increase for hephaestus
  - `internal/tui/views/wizard_agents.go:777-808` — `getLineForField` offsets — add new entry
  - `internal/tui/views/wizard_agents.go:900-950` — field cycling with left/right keys — add skip logic for non-hephaestus agents
  - `internal/tui/views/wizard_agents.go:1126-1134` — `renderField()` function — pattern for adding validation hint text

  **API/Type References**:
  - `internal/config/types.go:66` — `AllowNonGptModel *bool` field in AgentConfig
  - `internal/tui/styles.go` — Shared style palette (import `Red` color for validation errors)

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go:532-645` — `TestAgentApplyPreservesExistingFields` — follow this pattern for round-trip test

  **WHY Each Reference Matters**:
  - `renderBool` (line 1150): Exact function to call — don't reinvent checkbox rendering
  - Field enum must be at END to avoid shifting existing iota values
  - `getLineForField` MUST be updated or scrolling breaks for hephaestus form

  **Acceptance Criteria**:
  - [x] `go test -run TestAgent -v ./internal/tui/views/` → PASS
  - [x] New test: `AllowNonGptModel: true` on hephaestus → SetConfig → Apply → `*bool == true`
  - [x] New test: `AllowNonGptModel` on non-hephaestus agent → Apply → field is nil (not emitted)
  - [x] Validation test: color `"#fff"` renders with error hint, `"#FF6AC1"` renders without
  - [x] Validation test: temperature `"2.5"` renders with error hint, `"1.5"` renders without
  - [x] Validation test: top_p `"1.5"` renders with error hint, `"0.5"` renders without

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: allow_non_gpt_model round-trip for hephaestus
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Run test that creates config with `Agents["hephaestus"] = &AgentConfig{AllowNonGptModel: boolPtr(true)}`
      2. Call SetConfig → Apply on fresh config
      3. Assert `newCfg.Agents["hephaestus"].AllowNonGptModel != nil && *newCfg.Agents["hephaestus"].AllowNonGptModel == true`
    Expected Result: Test passes — field survives round-trip
    Failure Indicators: `AllowNonGptModel` is nil after Apply
    Evidence: .sisyphus/evidence/task-2-allow-non-gpt-roundtrip.txt

  Scenario: allow_non_gpt_model NOT emitted for non-hephaestus agents
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Run test that sets AllowNonGptModel on "build" agent
      2. Call SetConfig → Apply
      3. Assert `newCfg.Agents["build"].AllowNonGptModel == nil`
    Expected Result: Field is nil for non-hephaestus agents
    Failure Indicators: AllowNonGptModel is non-nil on "build"
    Evidence: .sisyphus/evidence/task-2-allow-non-gpt-non-heph.txt

  Scenario: Inline validation renders error hints
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create WizardAgents, set color to "#fff" on "build" agent
      2. Call `renderAgentForm("build", ac)` and check output contains "invalid hex" or error indicator
      3. Set color to "#FF6AC1" — verify no error indicator
      4. Same for temperature "3.0" (error) vs "1.5" (ok)
      5. Same for top_p "1.5" (error) vs "0.5" (ok)
    Expected Result: Error hints appear for invalid values, absent for valid values
    Failure Indicators: No validation text rendered, or validation on valid values
    Evidence: .sisyphus/evidence/task-2-validation-hints.txt
  ```

  **Commit**: YES
  - Message: `feat(agents): add allow_non_gpt_model checkbox and inline validation`
  - Files: `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_agents_test.go`
  - Pre-commit: `go test -run TestAgent -v ./internal/tui/views/`

- [x] 3. Replace `providerOptions` read-only display with editable key-value editor

  **What to do**:
  - **New state fields** in `agentConfig` struct:
    - Replace `providerOptionsDisplay string` with `providerOptions map[string]interface{}` (actual data)
    - Add `editingProviderOptions bool` — flag for sub-editor mode (follows `savingCustomModel` pattern)
    - Add `providerKVKeys []string` — ordered list of keys (maps are unordered in Go)
    - Add `providerKVValues []textinput.Model` — parallel array of value inputs
    - Add `providerKVKeyInputs []textinput.Model` — parallel array of key inputs
    - Add `providerKVFocusedIdx int` — which key-value pair is focused
    - Add `providerKVFocusedCol int` — 0=key, 1=value column
    - Add `providerKVNewKey textinput.Model` — input for adding new key

  - **SetConfig changes** (~line 530-542):
    - Store actual `ProviderOptions` map: `ac.providerOptions = agentCfg.ProviderOptions`
    - Initialize `providerKVKeys` from sorted map keys
    - Create `providerKVValues` textinputs initialized with formatted values (stringify non-string values via `fmt.Sprintf`)
    - Create `providerKVKeyInputs` textinputs initialized with key names

  - **Apply changes** (~line 694):
    - If `editingProviderOptions` was used, reconstruct `map[string]interface{}` from KV editor state
    - Parse values back to Go types: try `strconv.ParseFloat` → try `strconv.ParseBool` → keep as string
    - Write reconstructed map to `agentCfg.ProviderOptions`
    - If map is empty, set to nil

  - **Sub-editor UI** (rendered when `editingProviderOptions == true`):
    - Takes over the form area (follows `handleSaveCustomModel` pattern at line 1234)
    - Show list of key-value pairs with cursor navigation
    - Keybindings: `a` = add new pair, `d` = delete focused pair, `Enter` = edit value, `Esc` = exit editor
    - Each pair rendered as: `  key: value` with focused pair highlighted
    - Show help footer: `[a]dd [d]elete [Enter]edit [Esc]done`

  - **Entry point** in Update (when `fieldProviderOptions` focused + Enter pressed):
    - Set `editingProviderOptions = true`
    - Initialize KV editor state from current `providerOptions` map

  - **Render** in `renderAgentForm()` (replace line 1189):
    - When NOT editing: show summary like `"3 options set [Enter to edit]"` or `"(none) [Enter to edit]"`
    - When editing: render the KV editor inline within the form area

  **Must NOT do**:
  - Do NOT support nested object values — only string, number, boolean primitives. Complex values display as JSON string but can't be parsed back to objects
  - Do NOT change `ProviderOptions` type in `config/types.go` (remains `map[string]interface{}`)
  - Do NOT auto-commit on every keystroke — changes apply atomically when Esc exits editor

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Sub-editor component with state management, key-value CRUD, type coercion
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential)
  - **Blocks**: Task 4
  - **Blocked By**: Task 2

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:1234-1321` — `handleSaveCustomModel()` — **THE pattern** for sub-editor state management: boolean flag → take over View/Update → handle input → emit result. Follow this EXACTLY.
  - `internal/tui/views/wizard_agents.go:846-849` — `ModelSelectedMsg` handler — pattern for sub-editor completion (receive message → update state → clear flag)
  - `internal/tui/views/wizard_agents.go:1189` — Current read-only render to replace
  - `internal/tui/views/wizard_agents.go:558-559` — Current implicit preservation via struct reuse — MUST be replaced with explicit Apply logic

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go:715-753` — `TestAgentApplyPreservesProviderOptions` — MUST be updated: now providerOptions should survive round-trip through the KV editor (not just struct passthrough)

  **WHY Each Reference Matters**:
  - `handleSaveCustomModel` (1234): Establishes the exact sub-editor pattern: boolean flag controls which code path handles input in Update(). Following this prevents architectural deviation.
  - Line 558: Currently Apply preserves providerOptions by reusing existing struct. Once the field is editable, Apply MUST explicitly write the edited values back. The test at line 715 must verify this new behavior.

  **Acceptance Criteria**:
  - [x] `go test -run TestAgent -v ./internal/tui/views/` → PASS
  - [x] New test: `ProviderOptions: map[string]interface{}{"flag": true, "timeout": 30, "name": "test"}` → SetConfig → Apply on FRESH config → map restored with correct types
  - [x] New test: Empty providerOptions → SetConfig → Apply → `ProviderOptions == nil`
  - [x] Updated `TestAgentApplyPreservesProviderOptions`: now verifies round-trip through KV editor, not just struct passthrough

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: providerOptions round-trip with mixed types
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `ProviderOptions: map[string]interface{}{"flag": true, "timeout": float64(30), "name": "test"}`
      2. Call SetConfig → Apply on fresh config (no existing agent)
      3. Assert `ProviderOptions["flag"] == true` (bool)
      4. Assert `ProviderOptions["timeout"]` is numeric 30
      5. Assert `ProviderOptions["name"] == "test"` (string)
    Expected Result: All 3 entries preserved with correct types
    Failure Indicators: nil map, wrong types, missing entries
    Evidence: .sisyphus/evidence/task-3-prov-roundtrip.txt

  Scenario: providerOptions empty map results in nil
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with no ProviderOptions
      2. SetConfig → Apply on fresh config
      3. Assert `ProviderOptions == nil`
    Expected Result: nil (not empty map)
    Failure Indicators: Non-nil empty map `map[]`
    Evidence: .sisyphus/evidence/task-3-prov-empty.txt
  ```

  **Commit**: YES
  - Message: `feat(agents): add editable providerOptions key-value editor`
  - Files: `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_agents_test.go`
  - Pre-commit: `go test -run TestAgent -v ./internal/tui/views/`

---

- [x] 4. Replace `permission.bash` read-only display with per-command rule editor

  **What to do**:
  - **New state fields** in `agentConfig` struct:
    - Add `editingBashPerms bool` — sub-editor mode flag
    - Add `bashRules map[string]string` — tool→permission map (extracted from `originalBash` when it's an object)
    - Add `bashRuleKeys []string` — ordered list of tool names
    - Add `bashRuleFocusedIdx int` — which rule is focused
    - Add `bashRuleFocusedCol int` — 0=tool name, 1=permission value
    - Add `bashRuleNewTool textinput.Model` — input for adding new tool rule
    - Add `bashRulePermIdx []int` — parallel array of permission value indices (into permissionValues)

  - **SetConfig changes** (~line 505-520):
    - When `permission.Bash` is `map[string]interface{}`: parse into `bashRules` map, populate `bashRuleKeys` (sorted), set `permBashIdx = 0`
    - When `permission.Bash` is string: keep current behavior (set `permBashIdx`)
    - Always store `originalBash` for reference

  - **Apply changes** (~line 628-632):
    - If `bashRules` is non-empty (object mode): reconstruct `map[string]interface{}` from `bashRules` and write to `agentCfg.Permission.Bash`
    - If `bashRules` is empty AND `permBashIdx > 0` (string mode): write string value as before
    - If both empty: write nil

  - **Sub-editor UI** (rendered when `editingBashPerms == true`):
    - Takes over form area (same pattern as providerOptions editor)
    - Show list of tool→permission pairs: `  bash: allow`, `  rm: deny`, `  git: ask`
    - Keybindings: `a` = add new rule, `d` = delete focused rule, `←/→` = cycle permission (ask/allow/deny) on focused rule, `Esc` = exit
    - Help footer: `[a]dd [d]elete [←/→]cycle permission [Esc]done`

  - **Entry point** in Update:
    - When `fieldPermBash` focused + Enter pressed + `originalBash` is object OR user wants to create object rules:
      - Set `editingBashPerms = true`, initialize from `bashRules`
    - When `fieldPermBash` focused + left/right + originalBash is NOT object:
      - Keep current cycling behavior (string mode)

  - **Render** in `renderAgentForm()` (replace lines 1194-1216):
    - When object mode and NOT editing: show `"N rules set [Enter to edit]"`
    - When object mode and editing: render rule list inline
    - When string mode: keep current dropdown render with `[←/→]` hint

  - **Mode transition**: Add ability to SWITCH from string mode to object mode:
    - When `fieldPermBash` focused in string mode + special key (e.g., `Enter`): prompt "Convert to per-command rules? (y/n)"
    - If yes: create `bashRules` from current string value as default for all tools, enter editor mode

  **Must NOT do**:
  - Do NOT change `PermissionConfig.Bash` type in `config/types.go` — remains `interface{}`
  - Do NOT modify how other permission fields (edit, webfetch, task, doom_loop, external_directory) work
  - Do NOT force object mode — string mode must remain available

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Sub-editor with dual-mode (string vs object), state management, permission cycling
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential)
  - **Blocks**: Task 5
  - **Blocked By**: Task 3

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:1234-1321` — `handleSaveCustomModel()` — sub-editor state management pattern (same as Task 3)
  - `internal/tui/views/wizard_agents.go:1194-1216` — Current bash permission rendering — THE code to replace
  - `internal/tui/views/wizard_agents.go:628-632` — Current Apply logic for bash (string vs originalBash passthrough) — replace with explicit write
  - `internal/tui/views/wizard_agents.go:936-941` — Current cycling logic with object-mode blocking (`if _, isObj := ...; isObj { break }`)
  - `internal/tui/views/wizard_agents.go:168` — `originalBash interface{}` field — still needed for initial state detection

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go:647-684` — `TestAgentApplyPreservesBashObjectPermission` — MUST be updated to verify round-trip through editor, not just passthrough

  **WHY Each Reference Matters**:
  - Lines 936-941: Currently BLOCKS cycling when bash is object. The new editor REPLACES this block with "enter editor mode on Enter" behavior.
  - Lines 628-632: Currently uses `originalBash` passthrough. Must be replaced with explicit reconstruction from `bashRules` state.

  **Acceptance Criteria**:
  - [x] `go test -run TestAgent -v ./internal/tui/views/` → PASS
  - [x] New test: bash object `{"git": "allow", "rm": "deny"}` → SetConfig → Apply → object restored with exact rules
  - [x] New test: bash string `"ask"` → SetConfig → Apply → string preserved (not converted to object)
  - [x] Updated `TestAgentApplyPreservesBashObjectPermission`: verifies round-trip through editor state
  - [x] New test: add rule to bash object → Apply → new rule appears in output
  - [x] New test: delete rule from bash object → Apply → rule removed from output

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: bash object round-trip preserves rules
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `Permission.Bash = map[string]interface{}{"git": "allow", "rm": "deny", "docker": "ask"}`
      2. SetConfig → Apply on fresh config
      3. Assert result Bash is `map[string]interface{}` with 3 entries
      4. Assert `result["git"] == "allow"`, `result["rm"] == "deny"`, `result["docker"] == "ask"`
    Expected Result: All 3 rules preserved
    Failure Indicators: Rules lost, type changed to string, wrong values
    Evidence: .sisyphus/evidence/task-4-bash-obj-roundtrip.txt

  Scenario: bash string mode preserved when no object editing
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `Permission.Bash = "allow"` (string)
      2. SetConfig → Apply
      3. Assert result Bash is string `"allow"`
    Expected Result: String value preserved (not converted to object)
    Failure Indicators: Type is map instead of string
    Evidence: .sisyphus/evidence/task-4-bash-string-roundtrip.txt
  ```

  **Commit**: YES
  - Message: `feat(agents): add permission.bash per-command rule editor`
  - Files: `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_agents_test.go`
  - Pre-commit: `go test -run TestAgent -v ./internal/tui/views/`

---

- [x] 5. Replace `fallback_models` textinput with structured editor

  **What to do**:
  - **New state fields** in `agentConfig` struct:
    - Add `editingFallbackModels bool` — sub-editor mode flag
    - Add `fallbackModelEntries []fallbackModelEntry` — structured list of models
    - Add `fallbackFocusedIdx int` — which model entry is focused
    - Add `fallbackFocusedField int` — which field within the entry is focused (0=model, 1=variant, 2=reasoningEffort)
    - Keep existing `fallbackModels textinput.Model` as raw JSON escape hatch

  - **New type** `fallbackModelEntry` struct:
    ```go
    type fallbackModelEntry struct {
        model            string   // required — model ID
        modelDisplay     string   // display name from registry
        variant          string   // optional
        reasoningEffort  string   // optional — one of effortLevels
        // temperature, top_p, maxTokens, thinking are advanced — omit from structured UI, preserve via rawJSON
        rawJSON          string   // for complex ModelObjects that can't be represented in simple fields
        isRawJSON        bool     // true if this entry uses raw JSON mode
    }
    ```

  - **SetConfig changes**:
    - Parse `FallbackModels` (which is `interface{}`) into `fallbackModelEntries`:
      - If string: create 1 entry with `model = string`
      - If `[]interface{}` array: iterate, each element is either string (simple entry) or `map[string]interface{}` (ModelObject entry)
      - For ModelObject: extract `model`, `variant`, `reasoningEffort` into structured fields; if it has additional fields (temperature, top_p, maxTokens, thinking), set `isRawJSON = true` and store full JSON
    - Also populate the legacy `fallbackModels` textinput with JSON representation (for raw editing escape hatch)

  - **Apply changes**:
    - If `editingFallbackModels` was used (structured entries exist):
      - If 0 entries: set `FallbackModels = nil`
      - If 1 entry and it's a simple model string: set `FallbackModels = "model-id"` (string)
      - If multiple entries: build `[]interface{}` array — simple entries as strings, complex entries as `map[string]interface{}`
    - If raw textinput was used instead (legacy mode): keep current JSON parsing logic (lines 565-575)

  - **Sub-editor UI**:
    - List of model entries with compact display:
      ```
      1. claude-sonnet-4  variant: fast  effort: high
      2. gpt-4o           (simple)
      3. [raw JSON] {"model":"deepseek","temperature":0.3}
      ```
    - Keybindings:
      - `a` = add new model (opens ModelSelector to pick model ID)
      - `d` = delete focused entry
      - `Enter` = edit focused entry's fields (cycle through model/variant/effort)
      - `r` = toggle raw JSON mode for focused entry
      - `Esc` = exit editor
    - Model selection uses existing `ModelSelector` component (line 170: `modelSelector ModelSelector`)
    - Help footer: `[a]dd [d]elete [Enter]edit [r]aw JSON [Esc]done`

  - **Entry point** in Update:
    - When `fieldFallbackModels` focused + Enter: set `editingFallbackModels = true`, initialize from `fallbackModelEntries`
    - Legacy: if user types directly in textinput (when NOT in editor mode), raw JSON input still works

  - **Render** in `renderAgentForm()` (replace line 1224):
    - When NOT editing: show summary like `"3 models [Enter to edit]"` or `"(none) [Enter to edit]"`
    - When editing: render model list inline

  **Must NOT do**:
  - Do NOT change `FallbackModels interface{}` type in `config/types.go`
  - Do NOT implement full 7-field ModelObject editor (only model, variant, reasoningEffort are structured; rest via raw JSON)
  - Do NOT remove raw JSON fallback — users must be able to input arbitrary JSON for complex ModelObjects
  - Do NOT lose data: if a ModelObject has `temperature`, `thinking`, etc., these must survive round-trip via `rawJSON` field

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Complex sub-editor with polymorphic data parsing, ModelSelector integration, dual-mode (structured + raw JSON), type coercion
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (sequential)
  - **Blocks**: Task 6
  - **Blocked By**: Task 4

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents.go:1234-1321` — `handleSaveCustomModel()` — sub-editor pattern
  - `internal/tui/views/wizard_agents.go:565-575` — Current fallbackModels Apply logic (JSON parsing) — keep as raw mode fallback
  - `internal/tui/views/wizard_agents.go:169-177` — ModelSelector state fields in agentConfig — reuse this pattern for fallback model selection
  - `internal/tui/views/wizard_agents.go:840-849` — ModelSelectedMsg handler — pattern for receiving model selection result
  - `internal/tui/views/model_selector.go` — Reusable `ModelSelector` component — use for "add model" action

  **API/Type References**:
  - `internal/config/types.go:45` — `FallbackModels interface{}` — the polymorphic field
  - JSON schema `fallback_models`: `anyOf[string, array[string|ModelObject]]`
  - ModelObject schema: `model` (required), `variant`, `reasoningEffort`, `temperature` (0-2), `top_p` (0-1), `maxTokens`, `thinking` object

  **Test References**:
  - `internal/tui/views/wizard_agents_test.go:532` — `TestAgentApplyPreservesExistingFields` — tests `FallbackModels: ["claude-haiku", "gpt-4o"]`

  **WHY Each Reference Matters**:
  - Lines 565-575: Current raw JSON parsing MUST be preserved as fallback mode. New structured editor is an ADDITION, not a replacement.
  - `ModelSelector` (model_selector.go): Reuse existing fuzzy-search model picker for adding models — don't build a new one.
  - ModelObject schema: Only `model`, `variant`, `reasoningEffort` get structured UI. Other 4 fields are "advanced" and handled via raw JSON mode.

  **Acceptance Criteria**:
  - [x] `go test -run TestAgent -v ./internal/tui/views/` → PASS
  - [x] New test: `FallbackModels: "claude-haiku"` (string) → round-trip → string preserved
  - [x] New test: `FallbackModels: []interface{}{"claude-haiku", "gpt-4o"}` → round-trip → array preserved
  - [x] New test: `FallbackModels: []interface{}{map[string]interface{}{"model": "claude-haiku", "variant": "fast", "reasoningEffort": "high"}}` → round-trip → ModelObject fields preserved
  - [x] New test: ModelObject with extra fields (temperature, thinking) → round-trip → preserved via rawJSON
  - [x] New test: empty/nil fallbackModels → round-trip → nil

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: fallback_models string round-trip
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `FallbackModels: "claude-haiku"`
      2. SetConfig → Apply on fresh config
      3. Assert `FallbackModels` is string `"claude-haiku"`
    Expected Result: String value preserved
    Failure Indicators: Value is nil or converted to array
    Evidence: .sisyphus/evidence/task-5-fallback-string.txt

  Scenario: fallback_models ModelObject with extra fields preserved
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `FallbackModels: []interface{}{map[string]interface{}{"model": "deepseek", "temperature": 0.3, "thinking": map[string]interface{}{"type": "enabled", "budgetTokens": 4096}}}`
      2. SetConfig → Apply on fresh config
      3. Assert ModelObject has `model`, `temperature`, and `thinking` fields intact
    Expected Result: All fields including advanced ones preserved via rawJSON mode
    Failure Indicators: temperature or thinking lost
    Evidence: .sisyphus/evidence/task-5-fallback-complex.txt

  Scenario: fallback_models mixed array round-trip
    Tool: Bash
    Preconditions: Project compiles
    Steps:
      1. Create config with `FallbackModels: []interface{}{"claude-haiku", map[string]interface{}{"model": "gpt-4o", "variant": "fast"}}`
      2. SetConfig → Apply
      3. Assert array has 2 elements: string + map
    Expected Result: Mixed array type preserved
    Failure Indicators: All converted to strings or all to maps
    Evidence: .sisyphus/evidence/task-5-fallback-mixed.txt
  ```

  **Commit**: YES
  - Message: `feat(agents): add structured fallback_models editor with model selector`
  - Files: `internal/tui/views/wizard_agents.go`, `internal/tui/views/wizard_agents_test.go`
  - Pre-commit: `go test -run TestAgent -v ./internal/tui/views/`

---

- [x] 6. Comprehensive test suite + build verification

  **What to do**:
  - Run `make test` and `make lint` to verify ALL existing + new tests pass
  - Add any missing round-trip test coverage for edge cases:
    - All 7 `effortLevels` values (including `""` at index 0) round-trip correctly
    - `allow_non_gpt_model` with `false` value (not just `true`) on hephaestus
    - `providerOptions` with single entry, many entries (5+), and unicode keys
    - `permission.bash` with empty object `{}` → should result in nil (not empty map)
    - `fallback_models` with `[]interface{}{}` (empty array) → should result in nil
  - Verify `getLineForField` returns correct offsets for ALL fields including `fieldAllowNonGpt`
  - Verify form height constant is correct for both hephaestus (with extra field) and other agents
  - Run `go vet ./...` for any issues

  **Must NOT do**:
  - Do NOT add tests for wizard_categories, wizard_hooks, or other steps (separate scope)
  - Do NOT add integration tests requiring TUI rendering (keep unit-level)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Test writing + build verification — well-scoped
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (after all implementation)
  - **Blocks**: Final Verification Wave
  - **Blocked By**: Tasks 1, 5

  **References**:

  **Pattern References**:
  - `internal/tui/views/wizard_agents_test.go:532-645` — `TestAgentApplyPreservesExistingFields` — THE gold standard round-trip test pattern
  - `internal/tui/views/wizard_agents_test.go:755-814` — `TestAgentApplyPreservesUnmanagedFieldsOnEdit` — tests for field preservation during edit

  **Test References**:
  - `Makefile` — `make test` command (uses `go test -v -race ./...`)
  - `Makefile` — `make lint` command (uses `golangci-lint run`)

  **Acceptance Criteria**:
  - [x] `make test` → PASS (0 failures, race detector clean)
  - [x] `make lint` → PASS (0 errors)
  - [x] `make build` → binary builds successfully
  - [x] `go vet ./...` → 0 issues
  - [x] ALL edge case tests added and passing

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Full test suite passes
    Tool: Bash
    Preconditions: All implementation tasks complete
    Steps:
      1. Run `make test` — capture output
      2. Verify 0 failures in output
      3. Run `make lint` — capture output
      4. Verify 0 errors in output
      5. Run `make build` — verify binary builds
    Expected Result: All green — 0 failures, 0 lint errors, binary builds
    Failure Indicators: Any FAIL line in test output, any lint error, build failure
    Evidence: .sisyphus/evidence/task-6-full-suite.txt

  Scenario: Edge case tests exist and pass
    Tool: Bash
    Preconditions: Edge case tests written
    Steps:
      1. Run `go test -run "TestAgent.*EdgeCase\|TestAgent.*None\|TestAgent.*Minimal\|TestAgent.*AllowNonGpt" -v ./internal/tui/views/`
      2. Verify each edge case test is present and passes
    Expected Result: All edge case tests pass
    Failure Indicators: Test not found or fails
    Evidence: .sisyphus/evidence/task-6-edge-cases.txt
  ```

  **Commit**: YES
  - Message: `test(agents): add comprehensive round-trip tests for all agent fields`
  - Files: `internal/tui/views/wizard_agents_test.go`
  - Pre-commit: `make test && make lint`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run test). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `make lint` + `make test`. Review all changed files for: `as any`/type assertion shortcuts, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names (data/result/item/temp). Verify no `//nolint` suppressions added.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Build the binary (`make build`). Launch in tmux. Navigate to wizard agents step. For EACH improvement: verify the field renders correctly, accepts valid input, rejects invalid input (where applicable), and persists through Save→Reload cycle. Test edge cases: empty state, maximum values, special characters. Save screenshots to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance: no types.go changes, no schema.json changes, no other wizard step changes. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| # | Message | Files | Pre-commit |
|---|---------|-------|------------|
| 1 | `fix(agents): add missing reasoningEffort values none and minimal` | `wizard_categories.go` | `go test -run TestWizardCategories -v ./internal/tui/views/` |
| 2 | `feat(agents): add allow_non_gpt_model checkbox and inline validation` | `wizard_agents.go`, `wizard_agents_test.go` | `go test -run TestWizardAgents -v ./internal/tui/views/` |
| 3 | `feat(agents): add editable providerOptions key-value editor` | `wizard_agents.go`, `wizard_agents_test.go` | `go test -run TestAgent -v ./internal/tui/views/` |
| 4 | `feat(agents): add permission.bash per-command rule editor` | `wizard_agents.go`, `wizard_agents_test.go` | `go test -run TestAgent -v ./internal/tui/views/` |
| 5 | `feat(agents): add structured fallback_models editor with model selector` | `wizard_agents.go`, `wizard_agents_test.go` | `go test -run TestAgent -v ./internal/tui/views/` |
| 6 | `test(agents): add comprehensive round-trip tests for all agent fields` | `wizard_agents_test.go` | `make test && make lint` |

---

## Success Criteria

### Verification Commands
```bash
make test    # Expected: PASS (all tests green)
make lint    # Expected: 0 errors
make build   # Expected: binary builds successfully
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass
- [x] All 23 agent properties editable in TUI
- [x] effortLevels has 7 entries (including empty)
- [x] hephaestus shows allow_non_gpt_model, other agents don't
- [x] providerOptions supports add/edit/remove key-value pairs
- [x] permission.bash supports per-command rule editing
- [x] fallback_models uses structured model picker
- [x] color/temperature/top_p show validation feedback
