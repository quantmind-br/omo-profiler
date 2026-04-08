# INTERNAL/PROFILE

## OVERVIEW

Core business logic for profile persistence, state management, sparse field serialization, and switching. Handles CRUD operations and manages active configuration state.

## FILES

| File | Lines | Role |
|------|-------|------|
| `profile.go` | 350 | `Profile` struct, `Load`, `Save`, `Delete`, `List`, `Exists`, legacy field detection |
| `active.go` | 151 | `GetActive`, `SetActive`, `MatchesConfig`, sidecar state management |
| `naming.go` | 30 | `SanitizeName`, `ValidateName` — strict regex `^[a-zA-Z0-9_-]+$` |
| `sparse.go` | 378 | `MarshalSparse` — selective field serialization preserving unknown keys |
| `selection.go` | 254 | `FieldSelection` — tracks which fields user selected for sparse output |
| `profile_test.go` | 712 | CRUD + naming tests with `setupTestEnv` helper |
| `active_test.go` | — | Sidecar + content-scan fallback tests |
| `sparse_test.go` | — | Sparse serialization round-trip tests |
| `selection_test.go` | — | Field selection toggle/clone tests |

## KEY TYPES

- `Profile`: Wraps `config.Config` + metadata (`Name`, `Path`, `PreservedUnknown`, `FieldPresence`, `HasLegacyFields`, `LegacyFieldsWarning`)
- `FieldSelection`: Tracks user-selected fields for sparse serialization (`IsSelected`/`SetSelected`/`Toggle`)
- `ActiveConfig`: Result of `GetActive()` — the currently active profile info
- `activeState`: Internal sidecar file representation

## SWITCHING LOGIC

`SetActive(name)`:
1. Reads profile JSON from `profiles/<name>.json`
2. Overwrites `oh-my-openagent.json` with profile content (**COPY, not symlink**)
3. Updates `.active-profile` sidecar with profile name

`GetActive()`:
1. **Fast path**: Read `.active-profile` sidecar → O(1) lookup
2. **Fallback**: If sidecar stale/missing → O(N) content scan of all profiles
3. `MatchesConfig` normalizes data (strips `$schema`) before byte-for-byte comparison

## SPARSE SERIALIZATION

`MarshalSparse(profile, selection)`:
- Writes only fields marked as selected in `FieldSelection`
- Preserves unknown JSON keys from upstream (round-trip safe)
- `FieldSelection` created from: `NewBlankSelection`, `NewSelectionFromPresence`, `NewSelectionFromTemplate`

## NAMING VALIDATION

- Regex: `^[a-zA-Z0-9_-]+$` (alphanumeric, underscores, hyphens only)
- `SanitizeName`: Strips invalid chars, trims leading/trailing separators
- Empty names rejected

## LEGACY FIELD DETECTION

`detectLegacyFields`: Scans profile for fields deprecated in upstream schema. Sets `HasLegacyFields` + human-readable `LegacyFieldsWarning` on load.

## ANTI-PATTERNS

- **Symlinking**: DO NOT use symlinks for switching — always copy content
- **Manual State**: DO NOT manually edit `.active-profile` — let `SetActive` manage it
- **Raw File Access**: Avoid `os.Open` for profiles; use `profile.Load()`
- **Skip EnsureDirs**: Always call `config.EnsureDirs()` before writes
- **Direct Sparse Bypass**: Always use `MarshalSparse` for writing; never serialize full config when selection exists
