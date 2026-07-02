# Learnings

## 2026-04-08 Session Start
- Plan: optional-config-selection
- 12 implementation tasks + 4 final verification tasks
- Wave 1 (Tasks 1-5): Foundation layer
- Wave 2 (Tasks 6-9): Per-step checkbox retrofits (parallel)
- Wave 3 (Tasks 10-12): Integration + regression tests
- Critical path: T1 → T2 → T3/T4 (parallel) → T5 → T6-T9 (parallel) → T10 → T11/T12 (parallel) → F1-F4

## 2026-04-08 Task 2
- `FieldSelection` is fully value-independent: selection toggles only mutate canonical path membership and never touch config values.
- Seeding from `FieldPresence` works cleanly by expanding a present top-level key to every canonical subpath whose first segment matches that key.
- Wildcard lookup can stay generic by replacing each non-terminal segment with `*`, which covers `agents.*`, `categories.*`, `openclaw.gateways.*`, and similar map-backed paths.

## 2026-04-08 Task 3
- Wizard step plumbing can carry `*profile.FieldSelection` without widening `WizardStep`; the orchestrator already owns the implicit `SetConfig`/`Apply` contract.
- `prevStep()` already preserves selection automatically because it only re-inits views and never rebuilds `Wizard.selection`.
- Existing wizard step tests only needed signature updates with `nil` selection because this task stores selection state but does not consume it yet.

## 2026-04-08 Task 4
- Sparse serialization needs a canonical path layer independent from raw JSON tags because selection uses snake_case paths even when nested schema tags use camelCase (`maxTokens`, `reasoningEffort`, `replyListener`, etc.).
- Building the sparse object from reflected selected fields preserves explicit zero-like values (`false`, `""`, `[]`, `{}`) without depending on `omitempty` behavior.
- Preserved raw fragments can be merged safely by decoding them to generic maps and recursively overlaying selected known values so known leaves win while preserved sibling keys survive.

## 2026-04-08 Task 5
- Save-path schema validation can reuse the strict `gojsonschema` execution and filter only `ResultError.Type() == "required"`, preserving enum/type/pattern failures for present fields.
- Returning `nil, nil` after filtering all errors lets `{}` pass sparse save validation without changing strict `Validate` / `ValidateJSON` behavior.

## 2026-04-08 Task 8
- `wizard_hooks` only persists the top-level `disabled_hooks` array, so the new opt-in control can stay as a single inclusion row while leaving per-hook enable/disable editing untouched.
- Treating the inclusion row as cursor position `0` keeps existing up/down/page navigation simple and lets Space toggle either field inclusion or the currently focused hook based on row.
- `Apply` must build `disabled_hooks` from the per-hook state only when `selection.IsSelected("disabled_hooks")`; using `make([]string, 0)` preserves an explicitly selected empty array instead of collapsing it to `nil`.

## 2026-04-08 Categories checkbox retrofit
- `wizard_categories` can keep its existing single-row focus model by intercepting Space on persisted rows to toggle inclusion while leaving the current field value untouched.
- Category sparse persistence should be driven by canonical selection paths (for example `categories.*.model`, `categories.*.max_tokens`, `categories.*.thinking.budget_tokens`) because `FieldSelection` and `MarshalSparse` both normalize persisted paths to snake_case.
- Omitting a category should depend on whether any persisted subfield is selected, not on whether the current values are non-zero; selected empty/default values must still survive sparse serialization.

## 2026-04-08 Task 7
- `wizard_agents` can mirror the categories-step checkbox retrofit by keeping canonical snake_case selection paths for sparse serialization while normalizing camelCase aliases like `agents.*.maxTokens` and `agents.*.providerOptions` on entry.
- A `nil` selection still needs a legacy apply path in `wizard_agents` so older tests and non-selection callers keep the pre-opt-in behavior, while non-nil selection omits whole agents when no persisted subfields are selected.
- Parent-level opt-in rows are enough for nested editors like `providerOptions`, `permission.bash`, and `fallback_models`; Space only flips inclusion and leaves the nested editor state/value untouched.

## 2026-04-08 Review/save sparse parity
- The review step must call `profile.MarshalSparse` directly with both `selection` and `PreservedUnknown`; otherwise the preview shows omitted fields and unknown fragments differently from what the wizard actually writes.
- `ValidateForSave` is the right review/save gate for sparse profiles because it allows `{}` while still rejecting malformed selected values; using strict `Validate` in the review/save path reintroduces required-field failures for intentionally blank profiles.
- Wiring `PreservedUnknown` through `Wizard` into `WizardReview` keeps edit/template flows aligned: preview and save both include only selected known keys plus preserved unknown siblings, with byte-for-byte identical pretty JSON.

## 2026-04-08 Task 12
- Wizard regression coverage is easiest to keep stable by driving `WizardReview.SetConfig` plus the `Wizard.Update(WizardNextMsg{})` → `wizardSaveDoneMsg` save path, then asserting on decoded JSON and reloaded profiles instead of viewport text.
- For validation-safe explicit zero coverage, `background_task.providerConcurrency` accepts `0` while still round-tripping through sparse save/reload, unlike schema-constrained numeric fields such as `experimental.max_tools`.
- Manual raw JSON fixtures in `profiles/<name>.json` are the simplest way to model sparse edit/template/import-adjacent inputs with presence-seeded selection and preserved unknown keys intact.

## 2026-04-08 Task 11
- Cross-package regression coverage is easiest to keep future-proof by round-tripping sparse JSON through `profile.Load`/`MarshalSparse`/manual write/`Load` again, while asserting merged JSON separately for overlapping preserved fragments that `Load` intentionally only tracks at the top level.
- Blank sparse output should be asserted as literal `{}` from `profile.MarshalSparse` and then validated through both `ValidateForSave`/`ValidateJSONForSave` and strict `Validate`/`ValidateJSON` so the save-path exception stays scoped to `required` errors only.
- Table-driven save-validation fixtures work best when they mix typed invalid enums (to exercise `ValidateForSave`) with raw JSON type mismatches (to exercise `ValidateJSONForSave`), because Go types prevent many malformed-value cases from being represented in `config.Config` directly.

## 2026-04-08 F3 manual QA
- Command-driven QA is sufficient for this feature because the sparse-save behavior is already exercised end-to-end through wizard review/save tests plus profile/schema regression tests; no interactive terminal session is required to verify the JSON contract.
- The highest-signal regression points for optional selection are: blank review save (`{}`), explicit false preservation (`hashline_edit=false`), explicit empty slice preservation (`disabled_hooks=[]`), and unknown-fragment round-tripping during edit/save.

## 2026-04-08 F4 scope fidelity
- The safest scope-fidelity check is the combination of `git log <range> -- <guardrail-file>`, `git diff --name-only <range>`, and targeted symbol reads: it quickly proves whether a guardrail file or non-persisted UI surface was touched at all.
- `internal/config/types.go` remaining untouched is enough to establish that flexible/raw-field fidelity for `skills`, `runtime_fallback`, `fallback_models`, `permission.bash`, and `git_master.commit_footer` stayed intact for this plan.
- The existing wizard apply-path tests are strong evidence against scope creep: they prove omitted fields stay omitted and that explicit zero-like values only persist when a field is selected.

## 2026-04-08 F2 code quality review
- The current edit/template seeding strategy is too coarse: `FieldPresence` is top-level only, but `FieldSelection` is leaf-level. Expanding a present top-level object into every descendant path makes sparse edit/template flows over-select nested fields.
- `MarshalSparse` intentionally preserves selected nil values as explicit zero/empty JSON, so any over-selected field can become a semantic config change (`0`, `false`, `""`, `{}`, `[]`) instead of staying omitted.
- Preview/save byte parity is correct, but validation still needs to run against the exact sparse JSON payload rather than only the typed `config.Config`.

- 2026-04-08 F1 audit learning: top-level presence is insufficient for sparse edit UIs. Existing JSON needs canonical nested path presence (or equivalent) to avoid re-materializing absent agent/category/other fields on edit/save.

## 2026-04-08 leaf presence fix
- `profile.Load` can derive canonical sparse-presence paths by walking raw JSON and matching against `allFieldPaths`; exact canonical leaf matches should short-circuit descent so object-valued leaf fields like `provider_options` still count as present.
- Prefix traversal for map-backed config objects needs wildcard-aware matching on non-root segments, including the current tail segment (`agents.build` → `agents.*`), otherwise nested leaf presence never gets discovered.
- Once `FieldPresence` stores canonical leaf paths, `NewSelectionFromPresence` should only check direct presence membership; wildcard lookup stays in `FieldSelection.IsSelected` for concrete instance paths like `agents.build.model`.

## 2026-04-08 F1 re-audit
- The F1 reject state cleared once three things lined up together: import flows switched to `ValidateJSONForSave`, `FieldPresence` became leaf-level, and wizard review/save validated the exact sparse JSON bytes produced by `MarshalSparse`.
- The quickest high-signal proof for the nested-selection fix is the combination of `collectFieldPresence` recursion, direct-path `NewSelectionFromPresence`, and the new leaf-only tests in `profile_test.go` and `selection_test.go`.
- `internal/config/types.go` staying untouched remains the decisive schema-contract guardrail for this plan; both workspace diff checks and feature-range git checks should stay empty.

## 2026-04-08 F2 re-audit
- The leaf-presence repair held up under re-audit: `collectFieldPresence` now recurses to canonical leaf paths, `NewSelectionFromPresence` only seeds direct matches, and the leaf-only regression tests cover the old over-selection failure mode.
- The remaining review/save quality gap is no longer sparse-preview generation; it is error handling and gating. `WizardReview.validateAndPreview` still treats validator/runtime failures as valid, and `Wizard.nextStep` still keeps a typed-config validation gate ahead of sparse-byte validation.
