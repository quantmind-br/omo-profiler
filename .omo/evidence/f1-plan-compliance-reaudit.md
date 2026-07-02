# F1 Plan Compliance Re-Audit

Plan reviewed: `.sisyphus/plans/optional-config-selection.md`

## Re-audit scope
- Ran `go test ./... -count=1` → **PASS**
- Re-checked the exact files called out in the re-audit request
- Re-checked the prior F1 overall REJECT reasons from `.sisyphus/evidence/f1-plan-compliance.md`
- Verified `internal/config/types.go` remained untouched

## Previous REJECT reasons

### 1. Presence tracking was only top-level, so nested fields were over-selected
**Status: RESOLVED**

**Evidence**
- `internal/profile/profile.go:87-118` now walks raw JSON recursively via `collectFieldPresenceFromRaw(...)` and records canonical leaf paths only when they match known field paths.
- The same code stops on an exact known leaf match instead of broad top-level marking, so paths like `agents.*.model` can be present without also marking `agents.*.temperature`.
- `internal/profile/selection.go:174-181` now seeds selection by direct path membership only: it selects a path only if `presence[path]` is true.
- Repo-wide search for `topLevelPath` in `*.go` returned **no matches**, confirming the old top-level collapsing helper is gone.
- Regression tests now lock this behavior down:
  - `internal/profile/profile_test.go:326-399` (`TestProfileLoadCapturesFieldPresence`, `TestProfileLoadCapturesLeafPresenceForNestedKnownFields`)
  - `internal/profile/selection_test.go:45-65` (`TestNewSelectionFromPresenceOnlySelectsPresentLeafPaths`)

**Conclusion**
- The old top-level-only presence bug is fixed. Nested sparse selection now follows actual leaf-level JSON presence.

### 2. Import validation paths still used strict `ValidateJSON`
**Status: RESOLVED**

**Evidence**
- `internal/tui/app.go:555-560` now calls `validator.ValidateJSONForSave(data)`.
- `internal/cli/cmd/import.go:42-48` now calls `validator.ValidateJSONForSave(data)`.
- No strict `ValidateJSON(...)` call remains in those two import paths.

**Conclusion**
- Sparse profiles imported through either TUI or CLI now use save-path validation, matching the plan.

### 3. Regression coverage missed the nested-presence bug, leaving the core sparse contract unsafe
**Status: RESOLVED**

**Evidence**
- Leaf-presence and leaf-selection regression coverage now exists:
  - `internal/profile/profile_test.go:368-399` verifies nested known fields produce leaf-level presence only.
  - `internal/profile/selection_test.go:45-65` verifies selection seeding only marks present leaf paths.
- Review/save validation now runs against the actual sparse JSON bytes, not just typed config state:
  - `internal/tui/views/wizard.go:335-346` marshals sparse JSON with `profile.MarshalSparse(...)` and validates `data` via `ValidateJSONForSave(data)` before writing.
  - `internal/tui/views/wizard_review.go:121-146` builds preview bytes with `profile.MarshalSparse(...)` and validates those exact `jsonData` bytes via `ValidateJSONForSave(jsonData)`.
- Full-suite verification passed with `go test ./... -count=1`.

**Conclusion**
- The missing regression gap identified in the prior audit is now covered by targeted tests, and the review/save path validates the exact sparse payload that will be previewed and persisted.

## Guardrail check

### `internal/config/types.go` unchanged
**Status: PASS**

**Evidence**
- `GIT_MASTER=1 git status --short -- internal/config/types.go` returned no output.
- `GIT_MASTER=1 git diff -- internal/config/types.go` returned no output.
- `GIT_MASTER=1 git diff --name-only 4015a4a..HEAD -- internal/config/types.go` returned no output.
- `GIT_MASTER=1 git log --oneline 4015a4a..HEAD -- internal/config/types.go` returned no output.

## Overall verdict
All prior F1 REJECT reasons are resolved, the requested file-level checks pass, the schema guardrail remains intact, and the full Go test suite passes.

**VERDICT: APPROVE**
