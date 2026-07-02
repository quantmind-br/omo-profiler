# Problems

<!-- Append unresolved blockers here -->

## 2026-04-08 F2 code quality review
- Sparse edit/template flows are not semantically safe until selection seeding is moved from top-level presence tracking to leaf-level presence tracking.
- Validation remains incomplete until the review/save path validates the exact `MarshalSparse` output (including merged preserved-unknown fragments) with `ValidateJSONForSave`.

## 2026-04-08 F2 re-audit
- Final approval is still blocked by two runtime-path issues: review marks validator failures as valid, and the wizard save flow still performs a blocking typed-config validation pass before sparse-byte validation.
