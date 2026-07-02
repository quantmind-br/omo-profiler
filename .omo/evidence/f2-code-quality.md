# F2 Code Quality Review

## Verification
- `go test ./... -count=1` ✅
- `go vet ./...` ✅

## Findings

### Brittle state coupling
- **CRITICAL**: Edit/template flows only track **top-level** field presence, but saving is done at **leaf-path** granularity. `profile.Load` records presence by top-level key only (`internal/profile/profile.go:102-107`), then `NewSelectionFromPresence` expands any present top-level key to every descendant path in `allFieldPaths` (`internal/profile/selection.go:174-181`). `NewWizardForEdit` and `NewWizardFromTemplate` both seed selection from that broadened presence map (`internal/tui/views/wizard.go:123-156`).

  That means a profile containing only `background_task.providerConcurrency` is treated as if *all* `background_task.*` fields were selected. `WizardOther.Apply` then eagerly constructs selected siblings (`internal/tui/views/wizard_other.go:1677-1735`), and `MarshalSparse` serializes selected nil pointers/maps/slices as explicit zero values via `marshalLeafValue` + `zeroJSONValue` (`internal/profile/sparse.go:173-210`, `243-276`). Result: sparse edit/template round-trips can expand previously-absent fields into `0`, `false`, `""`, `{}`, or `[]`, changing config semantics.

- **INFO**: Selection state is passed explicitly through `Wizard` and step `SetConfig/Apply` calls. I found no global/singleton coupling for `FieldSelection` itself.

- **INFO**: `FieldSelection` is **not** concurrency-safe (`internal/profile/selection.go:166-252` uses a raw map with no synchronization). Current Bubble Tea usage is effectively single-threaded, so this is acceptable today, but it is not safe for future concurrent mutation.

### Duplicated checkbox logic
- **WARNING**: Checkbox selection logic is copy-pasted across steps instead of centralized. Categories and agents duplicate alias canonicalization, `is*FieldSelected`, and toggle logic (`internal/tui/views/wizard_categories.go:549-632`, `internal/tui/views/wizard_agents.go:1369-1465`). Rendering is also reimplemented separately in categories, agents, hooks, and other (`internal/tui/views/wizard_categories.go:1041-1084`, `internal/tui/views/wizard_agents.go:1967-2014`, `internal/tui/views/wizard_hooks.go:291-333`, `internal/tui/views/wizard_other.go:3564-3605`).

  The duplication is already causing UI drift: hooks/other color their include checkboxes, while categories/agents render raw `[✓]` strings without shared styling.

### Schema divergence
- **INFO**: `internal/config/types.go` was **not** modified by the optional-config-selection series. `git log --oneline 13b207a..HEAD -- internal/config/types.go` returned no commits. The last commit touching that file is `f907a33`, which predates this feature series.

### Save/preview drift
- **INFO**: Review preview and disk save use the same sparse serializer. `WizardReview.validateAndPreview` calls `profile.MarshalSparse(w.config, w.selection, w.preservedUnknown)` (`internal/tui/views/wizard_review.go:129`), and the save path does the same before `os.WriteFile` (`internal/tui/views/wizard.go:335-347`). Existing tests also assert preview/save parity (`internal/tui/views/wizard_sparse_flow_test.go`, `internal/tui/views/wizard_review_test.go`).

### Error handling
- **WARNING**: Review hides validator/runtime failures by marking the config valid. In `WizardReview.validateAndPreview`, validator acquisition/validation errors clear error state and set `isValid = true` (`internal/tui/views/wizard_review.go:138-149`). The actual save path does the opposite and aborts (`internal/tui/views/wizard.go:299-308`). Users can be told the config is valid and only discover the problem after pressing save.

- **WARNING**: Validation is performed against `w.config`, not the actual sparse JSON that will be written. Both review and save call `ValidateForSave` (`internal/tui/views/wizard_review.go:145`, `internal/tui/views/wizard.go:305`), which marshals the typed config and never validates the merged `preservedUnknown` fragments or the exact `MarshalSparse` output (`internal/schema/validator.go:95-101`, `126-155`). The serializer and validator are therefore not checking the same artifact.

### Dead code
- **INFO**: `profile.NewSelectionFromTemplate` currently has no production references (`internal/profile/selection.go:184-185`). It looks redundant with `NewSelectionFromPresence` in the current implementation.

## Overall verdict
- **REJECT**

## Must-fix issues
1. Fix edit/template selection seeding so it preserves **leaf-level** presence instead of broadening any present top-level object to every descendant field.
2. Validate the exact sparse JSON that will be written (`ValidateJSONForSave` on `MarshalSparse` output), including merged preserved-unknown fragments.
3. Stop treating validator/runtime failures in review as "valid"; surface them as blocking errors before save.

**VERDICT: REJECT** — sparse edit/template flows can over-select nested fields and write explicit default values that were not present before, which is a semantic regression for sparse configs.
