# INTERNAL/CONFIG

## OVERVIEW

Schema authority and path resolution. **Source of Truth** for `oh-my-opencode.json` structure — changes here affect persistence, UI rendering, and upstream compatibility.

## FILES

| File | Role |
|------|------|
| `types.go` | Root `Config` struct + 20 nested structs (~25 top-level fields) |
| `paths.go` | `ConfigDir`, `ProfilesDir`, `ConfigFile`, `ModelsFile`, `EnsureDirs` |
| `paths_test.go` | Path resolution + `SetBaseDir` isolation tests |
| `types_test.go` | Schema compliance + round-trip serialization tests |

## SCHEMA SAFETY

`types.go` is CRITICAL:

1. **JSON Tags**: Must match `oh-my-opencode` schema keys exactly
2. **Pointers**: Use `*bool` to distinguish `false` from "missing"
3. **No Logic**: Structs must remain pure data containers; no methods
4. **Synchronization**: Fields must stay in sync with upstream schema
5. **omitempty**: All JSON tags require `omitempty` to avoid dirty config files

## KEY TYPES

- `Config`: Root container. ~25 top-level fields including `Agents map[string]*AgentConfig`, `Categories map[string]*CategoryConfig`
- `AgentConfig`: 19 fields — model, variant, prompt, tools, permissions, thinking config
- `CategoryConfig`: 12 fields — model settings for task categories
- `ThinkingConfig`: Nested in agents/categories for reasoning budget control
- `ExperimentalConfig`: Feature flags with deeply nested `DynamicContextPruningConfig`

## PATH RESOLUTION

```go
ConfigDir()    → ~/.config/opencode/
ProfilesDir()  → ~/.config/opencode/profiles/
ConfigFile()   → ~/.config/opencode/oh-my-opencode.json
ModelsFile()   → ~/.config/opencode/models.json
```

`SetBaseDir(dir)` / `ResetBaseDir()` — test-only hooks that redirect ALL paths to temp dir.

## ANTI-PATTERNS

- **Hardcoded Paths**: `"/home/user/..."` → use `ConfigDir()` / `ProfilesDir()`
- **Direct Struct Access**: Mutating `Config` fields outside `profile` package
- **Missing Tags**: Omitting `json:"...,omitempty"` creates dirty config files
- **Logic in Types**: Adding validation methods to `Config` (keep it pure data)
- **Schema Drift**: Adding fields that don't exist in upstream `oh-my-opencode-schema.json`