# Decisions

## 2026-04-08
- Plan execution order follows dependency matrix strictly
- Wave 2 tasks (6-9) are parallelizable after T2+T3 complete
- Wave 3 tasks (11-12) are parallelizable after T10 completes

## 2026-04-08 Task 2
- Canonical selection paths use normalized snake_case segments for wizard state, even when underlying schema tags mix snake_case and camelCase.
- `NewSelectionFromTemplate` delegates to `NewSelectionFromPresence` so profile loads and template materialization share identical checkbox-seeding behavior.

## 2026-04-08 Task 3
- `Wizard` is the sole owner of canonical selection state via a new `selection *profile.FieldSelection` field; steps receive the pointer for synchronization but do not interpret it yet.
- New profile creation starts from `profile.NewBlankSelection()`, while edit/template flows reseed from `FieldPresence` before step hydration so selection and config snapshots stay aligned.

## 2026-04-08 Task 4
- `MarshalSparse` lives in a dedicated `internal/profile/sparse.go` helper and builds selected known JSON directly from typed config values rather than trying to reuse `omitempty`-filtered output.
- Deterministic output relies on `json.MarshalIndent` over `map[string]any`, which keeps keys sorted at every object level and matches the repo's pretty JSON format.
- Known selected values take precedence during preserved-fragment merges, but preserved sibling keys under the same top-level object are retained via recursive map merging.

## 2026-04-08 Task 5
- `Validator` now has separate save-path entry points (`ValidateForSave`, `ValidateJSONForSave`) instead of weakening existing strict validation methods.
- Required-field filtering is centralized in `isRequiredError`, keeping the sparse-save rule explicit and limited to missing-field schema failures.

## 2026-04-08 F4 scope review
- Scope verdict is APPROVE because the implementation stayed inside the planned surface area: persistence/selection plumbing, sparse validation, wizard persisted-field steps, and regression tests only.
- `wizard_name.go`, `template_select.go`, and `internal/config/types.go` are treated as explicit guardrail files for this feature and should remain unchanged in follow-up polish unless the plan is reopened.

- 2026-04-08 F1 audit decision: reject plan compliance. The sparse serializer and hook flow are implemented, but the foundational nested presence model was not delivered, so dependent wizard steps and review/save behavior are not plan-complete.

## 2026-04-08 leaf presence fix
- `FieldPresence` is now defined as canonical leaf-level selection paths (for example `agents.*.model`, `experimental.task_system`) instead of top-level object keys.
- Presence discovery walks known raw JSON branches only, normalizes JSON tag names to canonical snake_case path segments, and uses wildcard-aware prefix matching for map-backed sections like `agents`, `categories`, `openclaw.gateways`, and `openclaw.hooks`.
- Selection seeding now uses direct leaf-path membership; the old top-level expansion behavior was removed to prevent sparse edit flows from materializing absent sibling fields.
