# MODELS KNOWLEDGE BASE

## OVERVIEW
Registry of LLM models with local persistence and models.dev API integration for discovery and capability tracking.

## STRUCTURE
- `models.go`: Local registry CRUD, JSON persistence, and provider-based grouping.
- `modelsdev.go`: models.dev API client, response mapping, and capability formatting.

## RESPONSIBILITIES
- **Local Persistence**: Manages `models.json` via `config.ModelsFile()`; handles backups if JSON is corrupted.
- **Model CRUD**: Thread-safe-ish (local) operations with validation for unique `ModelID`.
- **UI Presentation**: Groups and sorts models by provider (case-insensitive, empty last) for Bubble Tea lists.
- **Discovery**: Fetches external model registry from `https://models.dev/api.json`.
- **Capability Metadata**: Tracks context windows, output limits, and features (reasoning, tools, vision).

## KEY TYPES
- `RegisteredModel`: The internal representation used by profiles. `ModelID` is the primary key.
- `ModelsRegistry`: Root container. `Load()` returns a registry with disk-syncing capabilities.
- `ProviderGroup`: Flattened structure for TUI menus (Provider name + sorted slice of models).
- `ModelsDevResponse`: Map-based API response from models.dev.
- `ModelsDevModel`: Rich metadata for external models (Limits, Family, Capabilities).

## ANTI-PATTERNS
- **Upstream Divergence**: Don't change `RegisteredModel` tags; they must remain compatible with `models.json`.
- **Manual Persistence**: Never call `os.WriteFile` for models; use `registry.Save()`.
- **Direct Slice Access**: Use `List()` to get a copy of models to prevent unintended side effects.
- **Ignore Errors**: Corruption in `models.json` must be handled (automatic backup to `.bak` is expected).
