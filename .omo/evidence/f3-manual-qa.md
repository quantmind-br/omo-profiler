# F3 Real Manual QA

Date: 2026-04-08
Mode: command-driven QA via Go tests + build

## Commands run

```text
$ go test ./... -count=1 -v
$ make build
$ go test ./internal/tui/views/... -run 'TestWizard(SparseCreateFlow|SparseEditRoundTrip|SparseTemplateFlow|SparseImportAdjacentFlow)' -v
$ go test ./internal/profile/... -run 'TestRegressionSparsePersistenceContract' -v
$ go test ./internal/schema/... -run 'TestRegressionSparseValidationContract' -v
$ go test ./internal/tui/views/... -run 'TestWizardReview(PreviewMatchesSparseSaveOutput|BlankProfileSavesAsEmptyObject)' -v
```

## Scenario results

### 1. Automated tests pass — PASS

Evidence:

```text
$ go test ./... -count=1 -v
...
=== RUN   TestWizardCreateFlow_SavesNewProfile
--- PASS: TestWizardCreateFlow_SavesNewProfile (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/tui/views 0.133s
```

Additional sparse-contract runs also passed:

```text
$ go test ./internal/schema/... -run 'TestRegressionSparseValidationContract' -v
=== RUN   TestRegressionSparseValidationContract
=== RUN   TestRegressionSparseValidationContract/invalid_enum_remains_invalid_after_sparse_marshal
=== RUN   TestRegressionSparseValidationContract/wrong_type_survives_required-error_filtering
--- PASS: TestRegressionSparseValidationContract (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/schema 0.006s
```

### 2. Build succeeds — PASS

Evidence:

```text
$ make build
go build -v -o omo-profiler ./cmd/omo-profiler
github.com/diogenes/omo-profiler/cmd/omo-profiler
```

### 3. Blank profile save writes `{}` — PASS

Evidence:

```text
$ go test ./internal/tui/views/... -run 'TestWizardReview(PreviewMatchesSparseSaveOutput|BlankProfileSavesAsEmptyObject)' -v
=== RUN   TestWizardReviewPreviewMatchesSparseSaveOutput
--- PASS: TestWizardReviewPreviewMatchesSparseSaveOutput (0.01s)
=== RUN   TestWizardReviewBlankProfileSavesAsEmptyObject
--- PASS: TestWizardReviewBlankProfileSavesAsEmptyObject (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/tui/views 0.011s
```

Assertion backing this scenario: `internal/tui/views/wizard_review_test.go:446-447` requires preview `== "{}"`, and `:475-476` requires saved file `== "{}"`.

### 4. Selected `hashline_edit=false` survives save — PASS

Evidence:

```text
$ go test ./internal/tui/views/... -run 'TestWizard(SparseCreateFlow|SparseEditRoundTrip|SparseTemplateFlow|SparseImportAdjacentFlow)' -v
=== RUN   TestWizardSparseCreateFlow_SavesOnlyCheckedFields
--- PASS: TestWizardSparseCreateFlow_SavesOnlyCheckedFields (0.01s)
=== RUN   TestWizardSparseEditRoundTrip_PreservesUnknownAndDropsUncheckedFields
--- PASS: TestWizardSparseEditRoundTrip_PreservesUnknownAndDropsUncheckedFields (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/tui/views 0.020s
```

```text
$ go test ./internal/profile/... -run 'TestRegressionSparsePersistenceContract' -v
=== RUN   TestRegressionSparsePersistenceContract
=== RUN   TestRegressionSparsePersistenceContract/selected_zero_values_survive_sparse_JSON
=== RUN   TestRegressionSparsePersistenceContract/multiple_preserved_unknown_fragments_survive_and_known_leaves_win_overlaps
=== RUN   TestRegressionSparsePersistenceContract/custom_bundle_survives_round-trip
=== RUN   TestRegressionSparsePersistenceContract/custom_flags_survive_round-trip
--- PASS: TestRegressionSparsePersistenceContract (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/profile 0.002s
```

Assertion backing this scenario: `internal/tui/views/wizard_sparse_flow_test.go:35-36` and `internal/profile/profile_test.go:516-517` both require `hashline_edit` to serialize as `false` when explicitly selected.

### 5. Selected `disabled_hooks` with no hooks disabled saves `[]` — PASS

Evidence:

```text
$ go test ./internal/profile/... -run 'TestRegressionSparsePersistenceContract' -v
=== RUN   TestRegressionSparsePersistenceContract
=== RUN   TestRegressionSparsePersistenceContract/selected_zero_values_survive_sparse_JSON
--- PASS: TestRegressionSparsePersistenceContract (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/profile 0.002s
```

Assertion backing this scenario: `internal/profile/profile_test.go:524-526` requires `disabled_hooks` to be an explicit empty array, and `:610-611` requires it to reload as `[]string{}`.

### 6. Unknown key preservation survives edit/save — PASS

Evidence:

```text
$ go test ./internal/tui/views/... -run 'TestWizard(SparseCreateFlow|SparseEditRoundTrip|SparseTemplateFlow|SparseImportAdjacentFlow)' -v
=== RUN   TestWizardSparseEditRoundTrip_PreservesUnknownAndDropsUncheckedFields
--- PASS: TestWizardSparseEditRoundTrip_PreservesUnknownAndDropsUncheckedFields (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/tui/views 0.020s
```

```text
$ go test ./internal/profile/... -run 'TestRegressionSparsePersistenceContract' -v
=== RUN   TestRegressionSparsePersistenceContract/multiple_preserved_unknown_fragments_survive_and_known_leaves_win_overlaps
=== RUN   TestRegressionSparsePersistenceContract/custom_bundle_survives_round-trip
=== RUN   TestRegressionSparsePersistenceContract/custom_flags_survive_round-trip
--- PASS: TestRegressionSparsePersistenceContract (0.00s)
PASS
ok   github.com/diogenes/omo-profiler/internal/profile 0.002s
```

Assertion backing this scenario: `internal/tui/views/wizard_sparse_flow_test.go:147-175` preserves unknown `custom: "data"` through edit/save, and `internal/profile/profile_test.go:543-558` plus `:645-664` preserves unknown fragments across round-trip save/reload.

## Overall verdict

APPROVE

All requested verification commands passed. The sparse/optional-selection behavior is covered by targeted wizard, profile, and schema tests, and the build succeeded.

**VERDICT: APPROVE**
