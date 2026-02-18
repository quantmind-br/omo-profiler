# INTERNAL/PROFILE

## OVERVIEW

Core business logic for profile persistence, state management, and switching. Handles CRUD operations and manages active configuration state.

## FILES

| File | Role |
|------|------|
| `profile.go` | `Profile` struct, `Load`, `Save`, `Delete`, `List`, `Exists` |
| `active.go` | `GetActive`, `SetActive`, `MatchesConfig`, sidecar state management |
| `naming.go` | `SanitizeName`, `ValidateName` — strict regex `^[a-zA-Z0-9_-]+$` |
| `profile_test.go` | CRUD + naming tests with `setupTestEnv` helper |
| `active_test.go` | Sidecar + content-scan fallback tests |

## KEY TYPES

- `Profile`: Wraps `config.Config` + metadata (`Name`, `Path`, `HasLegacyFields`, `LegacyFieldsWarning`)
- `ActiveConfig`: Result of `GetActive()` — the currently active profile info
- `activeState`: Internal sidecar file representation

## SWITCHING LOGIC

`SetActive(name)`:
1. Reads profile JSON from `profiles/<name>.json`
2. Overwrites `oh-my-opencode.json` with profile content (**COPY, not symlink**)
3. Updates `.active-profile` sidecar with profile name

`GetActive()`:
1. **Fast path**: Read `.active-profile` sidecar → O(1) lookup
2. **Fallback**: If sidecar stale/missing → O(N) content scan of all profiles
3. `MatchesConfig` normalizes data (strips `$schema`) before byte-for-byte comparison

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