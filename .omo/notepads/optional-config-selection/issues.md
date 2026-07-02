# Issues

<!-- Append issues here as they arise -->

## 2026-04-08 F2 code quality review
- CRITICAL: `profile.Load` records only top-level field presence, but `NewSelectionFromPresence` expands that presence to every descendant selection path. Editing/template-loading a sparse profile can therefore mark absent nested fields as selected.
- CRITICAL: Because `MarshalSparse` serializes selected nil pointers/maps/slices as explicit zero/empty JSON, the over-selection above can write `0`, `false`, `""`, `{}`, or `[]` for fields that were previously omitted.
- WARNING: `WizardReview.validateAndPreview` treats validator/runtime failures as valid, while the save path blocks on the same failures.
- WARNING: Review/save validation checks typed config via `ValidateForSave`, not the final sparse JSON after preserved-unknown fragments are merged.
- WARNING: Checkbox rendering/selection logic is duplicated across categories, agents, hooks, and other, and the UI already shows styling drift between steps.

- 2026-04-08 F1 audit: `FieldPresence` only records top-level keys in `internal/profile/profile.go`, so nested category/agent/other checkbox state cannot be reconstructed from actual JSON presence.
- 2026-04-08 F1 audit: import paths still call strict `ValidateJSON` in `internal/tui/app.go` and `internal/cli/cmd/import.go`, which rejects sparse configs the plan said should be importable.

## 2026-04-08 F2 re-audit
- WARNING: `WizardReview.validateAndPreview` still clears validator/runtime failures and sets `isValid = true`, so the review screen can present a broken validator state as saveable.
- WARNING: `Wizard.nextStep` still blocks on `ValidateForSave(&w.config)` before `MarshalSparse`, so save is not yet governed solely by validation of the exact sparse JSON bytes.
