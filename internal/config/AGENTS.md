# INTERNAL/CONFIG

## OVERVIEW

Schema authority and path resolution. **Source of Truth** for `oh-my-openagent.json` structure — changes here affect persistence, UI rendering, and upstream compatibility.

## FILES

| File | Role |
|------|------|
| `types.go` | Root `Config` struct + nested structs (44 top-level fields) |
| `paths.go` | `ConfigDir`, `ProfilesDir`, `ConfigFile`, `ModelsFile`, `EnsureDirs` |
| `paths_test.go` | Path resolution + `SetBaseDir` isolation tests |
| `types_test.go` | Schema compliance + round-trip serialization tests |

## SCHEMA SAFETY

`types.go` is CRITICAL:

1. **JSON Tags**: Must match `oh-my-openagent` schema keys exactly
2. **Pointers**: Use `*bool` to distinguish `false` from "missing"
3. **No Logic**: Structs must remain pure data containers; no methods
4. **Synchronization**: Fields must stay in sync with upstream schema
5. **omitempty**: All JSON tags require `omitempty` to avoid dirty config files

## KEY TYPES

- `Config`: Root container. 44 top-level fields including `Agents map[string]*AgentConfig`, `Categories map[string]*CategoryConfig`, `AgentOrder []string`
- `AgentConfig`: 24 fields — model, variant, prompt, tools, permissions, thinking config, displayName
- `CategoryConfig`: 15 fields — model settings for task categories
- `ThinkingConfig`: Nested in agents/categories for reasoning budget control
- `ExperimentalConfig`: 12 fields — feature flags with deeply nested `DynamicContextPruningConfig`
- `BackgroundTaskConfig`: 12 fields — circuit breaker, tool limits, depth controls, cleanup delay
- `TeamModeConfig`: 11 fields — multi-agent team mode (parallelism, message/wall-clock limits, mailbox)
- `GitMasterConfig`: 3 fields — commit footer, co-authored-by, env prefix (required at root level)
- `TmuxConfig`: 6 fields — layout, isolation, pane sizing
- `SisyphusAgentConfig`: 5 fields — TDD mode, replace plan, staleness control
- `KeywordDetectorConfig`: 2 fields — `enabled_expansions` allowlist + `disabled_keywords` for the keyword-detector hook
- `ClaudeCodeConfig`: 8 fields — mcp, commands, skills, agents, hooks, plugins, plugins_override, anthropic_provider
- `WebsearchConfig`: 1 field — provider selection
- `MonitorConfig`: 11 fields — output/log monitor subsystem (live mode, batch/ring buffers, runtime limits) — added v4.11.0
- `CodegraphConfig`: 5 fields — code-graph indexing subsystem (auto-provision, install dir, telemetry, watch debounce) — added v4.11.0
- `TuiConfig`: 1 nested field — oh-my-openagent TUI sidebar toggle (`TuiSidebarConfig.Enabled`) — added v4.11.0

## PATH RESOLUTION

```go
ConfigDir()    → ~/.config/opencode/
ProfilesDir()  → ~/.config/opencode/profiles/
ConfigFile()   → ~/.config/opencode/oh-my-openagent.json (detects legacy oh-my-opencode.json)
ModelsFile()   → ~/.config/opencode/models.json
```

`SetBaseDir(dir)` / `ResetBaseDir()` — test-only hooks that redirect ALL paths to temp dir.

## ANTI-PATTERNS

- **Hardcoded Paths**: `"/home/user/..."` → use `ConfigDir()` / `ProfilesDir()`
- **Direct Struct Access**: Mutating `Config` fields outside `profile` package
- **Missing Tags**: Omitting `json:"...,omitempty"` creates dirty config files
- **Logic in Types**: Adding validation methods to `Config` (keep it pure data)
- **Schema Drift**: Adding fields that don't exist in upstream `oh-my-openagent` schema