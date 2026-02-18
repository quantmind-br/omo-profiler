# MODELS KNOWLEDGE BASE

## OVERVIEW

LLM model registry with local JSON persistence and models.dev API integration for external model discovery.

## FILES

| File | Role |
|------|------|
| `models.go` | `ModelsRegistry` CRUD, `Load`/`Save`, provider-based grouping |
| `modelsdev.go` | `FetchModelsDevRegistry` API client, `ModelsDevResponse` mapping, capability formatting |
| `models_test.go` | Registry operations, persistence, provider grouping tests |

## KEY TYPES

| Type | Role |
|------|------|
| `RegisteredModel` | Internal model record: `DisplayName`, `ModelID` (primary key), `Provider` |
| `ModelsRegistry` | Root container with `Load()`, `Save()`, `Add()`, `Update()`, `Delete()`, `List()`, `ListByProvider()` |
| `ProviderGroup` | Flattened structure for TUI: provider name + sorted model slice |
| `ModelsDevResponse` | Map-based API response from `https://models.dev/api.json` |
| `ModelsDevModel` | Rich external metadata: limits, family, capabilities (reasoning, tools, vision) |

## PERSISTENCE

- Storage: `config.ModelsFile()` → `~/.config/opencode/models.json`
- Corruption recovery: auto-backup to `.bak` on JSON parse failure
- Thread safety: basic mutex protection on registry operations

## API CLIENT

`FetchModelsDevRegistry()` → HTTP GET `https://models.dev/api.json` → returns `ModelsDevResponse`
- `ListProviders()` → sorted `ProviderWithCount` slice
- `GetProviderModels(provider)` → filtered model list
- `ToRegisteredModel()` → converts external model to local `RegisteredModel`

## ANTI-PATTERNS

- **Upstream Divergence**: Don't change `RegisteredModel` JSON tags; must remain compatible with `models.json`
- **Manual Persistence**: Never `os.WriteFile` for models; use `registry.Save()`
- **Direct Slice Access**: Use `List()` to get a copy; prevents unintended side effects