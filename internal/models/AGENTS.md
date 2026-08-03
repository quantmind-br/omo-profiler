# MODELS KNOWLEDGE BASE

## OVERVIEW

LLM model registry with local JSON persistence and models.dev API integration for external model discovery.

## FILES

| File | Role |
|------|------|
| `models.go` | `ModelsRegistry` CRUD, `Load`/`Save`, `Mutate` transaction, provider-based grouping |
| `modelsdev.go` | `FetchModelsDevRegistry` API client, `ModelsDevResponse` mapping, capability formatting |
| `models_test.go` | Registry operations, persistence, provider grouping tests |

## KEY TYPES

| Type | Role |
|------|------|
| `RegisteredModel` | Internal model record: `DisplayName`, `ModelID` (primary key), `Provider` |
| `ModelsRegistry` | Root container with `Load()`, `Save()`, `List()`, `ListByProvider()`. Mutators are unexported (`add`/`update`/`remove`) and in-memory only |
| `ProviderGroup` | Flattened structure for TUI: provider name + sorted model slice |
| `ModelsDevResponse` | Map-based API response from `https://models.dev/api.json` |
| `ModelsDevModel` | Rich external metadata: limits, family, capabilities (reasoning, tools, vision) |

## PERSISTENCE

- Storage: `config.ModelsFile()` → `~/.omo/models.json` (omo-profiler local state, not part of the upstream omo document)
- Corruption recovery: auto-backup to `.bak` on JSON parse failure
- Thread safety: basic mutex protection on registry operations

## API CLIENT

`FetchModelsDevRegistry()` → HTTP GET `https://models.dev/api.json` → returns `ModelsDevResponse`
- `ListProviders()` → sorted `ProviderWithCount` slice
- `GetProviderModels(provider)` → filtered model list
- `ToRegisteredModel()` → converts external model to local `RegisteredModel`

## MUTATION CONTRACT

Every write is a serialized transaction: `Mutate(fn)` locks `regMutex`, loads,
applies `fn`, then saves atomically (temp file + rename via
`config.WriteFileAtomic`). `fn` must not call `Mutate` (not reentrant) or
`Save`; returning an error aborts before any write.

| Function | Behavior |
|----------|----------|
| `Add(m)` | Insert, `*ModelExistsError` if `(Provider, ModelID)` taken |
| `Update(provider, modelId, m)` | Replace/rename, `*ModelNotFoundError` or `*ModelExistsError` |
| `Delete(provider, modelId)` | Remove, `*ModelNotFoundError` if absent |
| `AddMany(list)` | Bulk import in **one** transaction; reports added/skipped |

Scope: one process. Cross-process races remain possible.

## ANTI-PATTERNS

- **Upstream Divergence**: Don't change `RegisteredModel` JSON tags; must remain compatible with `models.json`
- **Manual Persistence**: Never `os.WriteFile` for models; go through the package API
- **Unserialized read-modify-write**: NEVER `Load()` → edit → `Save()`. `models.json` is rewritten in full, so concurrent cycles lose updates. Use the package-level `Add`/`Update`/`Delete`/`AddMany`, or `Mutate(fn)` for anything else — they hold `regMutex` across load+edit+save
- **Stale views after a write**: the transaction does not touch a caller's in-memory `*ModelsRegistry`; re-`Load()` after mutating (see `ModelRegistry.reload`)
- **Direct Slice Access**: Use `List()` to get a copy; prevents unintended side effects
