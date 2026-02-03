# INTERNAL/CONFIG

## OVERVIEW

Schema authority and path resolution. This package is the **Source of Truth** for `oh-my-opencode.json` structure. Changes affect persistence, UI rendering, and upstream compatibility.

## RESPONSIBILITIES

- **Schema Authority**: Maps Go structs (`types.go`) to JSON schema 1:1
- **Path Resolution**: Manages `~/.config/opencode/` paths via `paths.go`
- **Environment Abstraction**: Isolates filesystem paths for testability
- **Serialization**: Controls `json:"..."` tags for correct file I/O

## SCHEMA SAFETY

`types.go` is CRITICAL:

1. **JSON Tags**: Must match `oh-my-opencode` schema keys exactly
2. **Pointers**: Use pointers (`*bool`) to distinguish `false` from "missing"
3. **No Logic**: Structs must remain pure data containers; no methods
4. **Synchronization**: Fields must stay in sync with upstream schema

## TESTING HOOKS

Filesystem isolation is mandatory for all tests:

```go
// IN TESTS ONLY:
config.SetBaseDir(t.TempDir())
defer config.ResetBaseDir()
```

This redirects `ConfigDir()` to a temp directory, protecting the real `~/.config`.

## ANTI-PATTERNS

- **Hardcoded Paths**: `"/home/user/..."` (use `paths.ConfigDir()`)
- **Direct Struct Access**: Mutating `Config` fields outside `profile` package
- **Missing Tags**: Omitting `json:"...,omitempty"` creates dirty config files
- **Logic in Types**: Adding validation methods to `Config` struct (keep it pure)
