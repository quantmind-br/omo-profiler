# INTERNAL/PROFILE

## OVERVIEW

Core business logic for profile persistence, state management, and switching. Handles CRUD operations for JSON profiles and manages the active configuration state.

## PERSISTENCE MODEL

- **Storage**: Profiles stored as individual JSON files in `~/.config/opencode/profiles/`
- **Format**: Standard `oh-my-opencode.json` schema structure
- **Lifecycle**: Full CRUD via `profile.go`
- **Safety**: `config.EnsureDirs()` called before writes

## SWITCHING LOGIC

- **Mechanism**: **COPY**, not symlink. `SetActive` overwrites `oh-my-opencode.json` with profile content
  - Reason: Ensures compatibility with external tools/fsnotify
- **State Tracking**:
  1. **Primary**: Fast O(1) lookup via hidden `.active-profile` sidecar (stores name)
  2. **Fallback**: If sidecar stale/missing, performs O(N) content scan of all profiles
- **Comparison**: `MatchesConfig` normalizes data (strips `$schema`) before byte-for-byte check

## VALIDATION RULES

- **Naming**: Strict regex `^[a-zA-Z0-9_-]+$` (alphanumeric, underscores, hyphens)
- **Sanitization**: `SanitizeName` strips invalid chars and trims separators
- **Constraints**: Name cannot be empty

## ANTI-PATTERNS

- **Symlinking**: DO NOT use symlinks for switching. Always copy content.
- **Manual State**: DO NOT manually edit `.active-profile`. Let `SetActive` manage it.
- **Raw File Access**: Avoid `os.Open` for profiles; always use `profile.Load()`.
