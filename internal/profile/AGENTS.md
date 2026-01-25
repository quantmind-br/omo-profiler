# INTERNAL/PROFILE KNOWLEDGE BASE

## OVERVIEW
Core business logic for profile persistence, state management, and switching mechanisms.

## CORE RESPONSIBILITIES
- **Persistence**: Handles full lifecycle (Create, Read, Update, Delete) of profile JSON files stored in `~/.config/opencode/profiles/`.
- **Directory Management**: Automatically ensures configuration directories exist via `config.EnsureDirs()` before any write operation.
- **State Management**: specific logic to determine the "active" profile, using a dual-strategy approach (cache vs. content match).
- **Switching**: Implements `SetActive` to atomically overwrite the main `oh-my-opencode.json` config with the selected profile's content.
- **Validation**: Enforces strict naming conventions (`^[a-zA-Z0-9_-]+$`) to ensure filesystem safety and consistent CLI usage.

## KEY FILES
| File | Purpose |
|---|---|
| `profile.go` | Defines `Profile` struct and implements filesystem persistence (`Load`, `Save`, `Delete`, `List`). |
| `active.go` | Logic for `GetActive`/`SetActive`, managing the `.active-profile` cache and content comparison. |
| `naming.go` | Regex-based validation (`ValidateName`) and sanitization (`SanitizeName`) for profile identifiers. |

## PATTERNS
- **Copy-over-Symlink**: `SetActive` copies JSON content to the target config location instead of using symlinks. This ensures compatibility with external tools that might overwrite the file or rely on `fsnotify` events on the real file.
- **Optimistic State Caching**: Maintains a hidden `.active-profile` sidecar file to enable O(1) lookup of the current profile name. This avoids the expensive O(N) operation of loading and comparing every profile against the active config on every CLI run.
- **Fallback Verification**: If the cache is stale or missing, `GetActive` falls back to scanning all profiles (`List()`) and comparing content byte-for-byte to identify the active configuration.
- **Config Normalization**: The `MatchesConfig` method specifically strips non-functional metadata (like `$schema`) before comparing JSON payloads, preventing false negatives when the schema URL changes but content remains identical.
