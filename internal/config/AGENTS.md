# INTERNAL/CONFIG

## OVERVIEW

Omo document helpers, `[opencode]` type authority, and path resolution. **Source of Truth** for the flat `config.Config` shape (the `[opencode]` harness block) and for reading/writing the unified `~/.omo/omo.json` / `omo.jsonc` document via `Document`.

## FILES

| File | Role |
|------|------|
| `types.go` | `Config` struct + nested structs (46 top-level fields = the `[opencode]` block) |
| `document.go` | `Document` — load/save omo.json(c), profile block CRUD, `Mutate`/`MutateWithPreSave` transactions, `WriteFileAtomic` |
| `jsonc.go` | `StripJSONC` — JSONC → JSON (comments/trailing commas blanked) |
| `paths.go` | `OmoDir`, `OmoFile`, `ModelsFile`, legacy migration helpers, `EnsureDirs` |
| `paths_test.go` | Path resolution + `SetBaseDir` isolation tests |
| `types_test.go` | Schema compliance + round-trip serialization tests |

## SCHEMA SAFETY

`types.go` is CRITICAL:

1. **JSON Tags**: Must match the `[opencode]` sub-schema keys exactly
2. **Pointers**: Use `*bool` to distinguish `false` from "missing"
3. **No Logic**: Structs must remain pure data containers; no methods
4. **Synchronization**: Fields must stay in sync with upstream `[opencode]` schema
5. **omitempty**: All JSON tags require `omitempty` to avoid dirty config files

`Config` is **not** the whole omo file — it models only the `[opencode]` block (and each profile's `profiles.<name>.[opencode]` override).

## KEY TYPES

- `Config`: Flat `[opencode]` container. 46 top-level fields including `Agents map[string]*AgentConfig`, `Categories map[string]*CategoryConfig`, `AgentOrder []string`
- `Document`: Parsed omo.json/omo.jsonc; keeps every top-level key as raw JSON so harness blocks (`[senpi]`, `[codex]`), shared typed keys, and future schema additions round-trip. Comments in `.jsonc` are **not** preserved across write (re-serialized as canonical JSON).
- `AgentConfig`: 25 fields — model, reasoning (canonical), variant/reasoningEffort (deprecated), prompt, tools, permissions, thinking config, displayName
- `CategoryConfig`: 20 fields — model, models (canonical chain), reasoning, max_tokens, provider_options, warn_unavailable; legacy maxTokens/variant/reasoningEffort/fallback_models retained
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
- `CodegraphConfig`: 8 fields — code-graph indexing subsystem (auto-init, auto-provision, daemon, excluded roots, install dir, telemetry, watch debounce) — daemon/excluded_roots added v4.19.2
- `TuiConfig`: 1 nested field — oh-my-openagent TUI sidebar toggle (`TuiSidebarConfig.Enabled`) — added v4.11.0
- `GoalConfig`: 3 fields — goal subsystem (`enabled`, `auto_start`, `default_max_iterations`) — added v4.19.0, replaces the old typed `ralph_loop`
- `Config.RalphLoop`: `json.RawMessage` — deprecated upstream compatibility shim (free-form record, migrated to `goal` by the plugin); preserved verbatim, not editable in the wizard

## DOCUMENT KEYS

```go
SchemaKey   = "$schema"
ProfilesKey = "profiles"
OpenCodeKey = "[opencode]"
SenpiKey    = "[senpi]"
CodexKey    = "[codex]"
```

## PATH RESOLUTION

```go
OmoDir()             → ~/.omo/
OmoFile()            → ~/.omo/omo.jsonc if present, else ~/.omo/omo.json
ModelsFile()         → ~/.omo/models.json          // omo-profiler local state
EnsureDirs()         → mkdir -p ~/.omo
LegacyConfigDir()    → pre-unification OpenCode config dir // migration detection only
LegacyConfigFile()   → legacy flat file if present, else ""
LegacyProfilesDir()  → pre-unification profiles dir         // migration detection only
DefaultSchema        // assets/omo.schema.json on the upstream monorepo
```

Removed (do not reference): `ConfigDir`, `ConfigFile`, `ProfilesDir`, `ConfigBasename`, `LegacyConfigBasename`.

`SetBaseDir(dir)` / `ResetBaseDir()` — test-only hooks that redirect ALL paths under `<tmp>/.omo/`.

## ANTI-PATTERNS

- **Hardcoded Paths**: `"/home/user/..."` → use `OmoDir()` / `OmoFile()` / `ModelsFile()`
- **Treating Config as the whole file**: `Config` is the `[opencode]` block only; use `Document` for the omo file
- **File-per-profile**: Profiles live at `profiles.<name>` inside the document — there is no profiles directory
- **Direct Struct Access**: Mutating `Config` fields outside `profile` package
- **Missing Tags**: Omitting `json:"...,omitempty"` creates dirty config files
- **Logic in Types**: Adding validation methods to `Config` (keep it pure data)
- **Schema Drift**: Adding fields that don't exist in upstream `[opencode]` schema
