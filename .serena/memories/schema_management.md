# Schema Management

## Overview
The schema system keeps omo-profiler in sync with the upstream oh-my-openagent **omo document** schema (`assets/omo.schema.json`). It validates whole `~/.omo/omo.json` / `omo.jsonc` documents and flat `[opencode]` blocks (including `profiles.<name>.[opencode]`), compares against upstream, and detects drift.

Upstream monorepo layout:
- Schema source: `packages/omo-config-core/src/schema/`
- Migration transform: `packages/omo-opencode/src/config-migration/`

## Package: `internal/schema/`

### Files
| File | Role |
|------|------|
| `validator.go` | Singleton validator; `GetEmbeddedSchema`, `GetOpenCodeSchema`, document + `[opencode]` validate paths |
| `compare.go` | Upstream schema fetching, comparison, and diff saving |
| `schema.json` | Embedded **omo document** schema used for validation (via `go:embed`) |
| `compare_test.go` | Tests for schema comparison logic |
| `validator_test.go` | Tests for validation logic |
| `accessor_test.go` | Tests for schema accessors |

Root `omo.schema.json` mirrors the embedded document schema. The old root `oh-my-opencode.schema.json` was deleted.

### Key Types & Functions

**Validator (singleton):**
```go
schema.GetValidator() (*Validator, error)       // Singleton access (sync.Once)
schema.GetEmbeddedSchema() []byte               // FULL omo document schema
schema.GetOpenCodeSchema() ([]byte, error)      // self-contained [opencode] sub-schema for forms
validator.Validate(cfg *config.Config) ([]ValidationError, error)
validator.ValidateJSON(data []byte) ([]ValidationError, error)
validator.ValidateForSave / ValidateJSONForSave // permissive save path
validator.ValidateDocument / ValidateDocumentForSave
```

**Schema Comparison:**
```go
schema.FetchUpstreamSchema(ctx) ([]byte, error)  // HTTP GET from upstream URL
schema.CompareSchemas() (*CompareResult, error)   // Compare embedded vs upstream
schema.SaveDiff(dir, diffContent) (string, error) // Save diff to timestamped file
```

**Types:**
```go
type ValidationError struct {
    Path    string  // JSON path to the error
    Message string  // Error message
}

type CompareResult struct {
    Identical bool    // true if schemas match
    Diff      string  // unified diff output (empty if identical)
}
```

## Upstream Schema URL
```
https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/omo.schema.json
```
Also available as `config.DefaultSchema` (written into new documents via `Document.EnsureSchema()`) and `schema.UpstreamSchemaURL`.

## Schema Sync Workflow
1. **Automatic check**: `omo-profiler schema-check` CLI command or TUI Schema Check view
2. `CompareSchemas()` fetches upstream via HTTP, compares against `go:embed`'d `schema.json`
3. If different, generates unified diff via `diff.ComputeUnifiedDiff`
4. User can save diff to a folder via `SaveDiff()` (timestamped: `schema-diff-YYYYMMDD-HHMMSS.diff`)
5. To actually update: replace `internal/schema/schema.json` and root `omo.schema.json` with upstream content, then rebuild

**Note**: The old `update-schema.sh` script no longer exists. Schema comparison is now fully in Go code.

## Config Schema Authority
`internal/config/types.go` (`Config`) is the **source of truth for the `[opencode]` block**:
- Must match the `[opencode]` sub-schema 1:1
- JSON tags must be exact matches
- Use `*bool` for optional booleans (distinguish `false` from missing)
- Use `json.RawMessage` for flexible fields
- All tags require `omitempty`
- No methods on data structs
- The whole omo file is `config.Document`; editor forms must use `GetOpenCodeSchema()`

## Anti-Patterns
- **Schema Divergence**: Never add `Config` fields without upstream `[opencode]` support
- **Manual Schema Edits**: Don't modify `schema.json` without verifying against upstream
- **Wrong schema for forms**: Don't drive editors with `GetEmbeddedSchema()` — use `GetOpenCodeSchema()`
- **Referring to `oh-my-opencode.schema.json`**: Deleted; project schema is `omo.schema.json`
- **Skipping Validation**: Always validate imported profiles before saving
- **Hardcoded Schema URL**: Use `config.DefaultSchema` or `schema.UpstreamSchemaURL` constants
