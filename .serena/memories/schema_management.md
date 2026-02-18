# Schema Management

## Overview
The schema system ensures omo-profiler stays in sync with the upstream `oh-my-opencode` JSON schema. It provides validation, upstream comparison, and drift detection.

## Package: `internal/schema/`

### Files
| File | Role |
|------|------|
| `validator.go` | Singleton JSON schema validator with `Validate` and `ValidateJSON` |
| `compare.go` | Upstream schema fetching, comparison, and diff saving |
| `schema.json` | Embedded JSON schema used for validation (via `go:embed`) |
| `oh-my-opencode.schema.json` | Reference copy of upstream schema |
| `compare_test.go` | Tests for schema comparison logic |
| `validator_test.go` | Tests for validation logic |
| `accessor_test.go` | Tests for `GetEmbeddedSchema` |

### Key Types & Functions

**Validator (singleton):**
```go
schema.GetValidator() (*Validator, error)       // Singleton access (sync.Once)
validator.Validate(cfg *config.Config) ([]ValidationError, error)  // From struct
validator.ValidateJSON(data []byte) ([]ValidationError, error)     // From raw JSON
schema.GetEmbeddedSchema() []byte               // Raw embedded schema bytes
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
https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/master/assets/oh-my-opencode.schema.json
```
Also available as `config.DefaultSchema` constant (used when creating new profiles).

## Schema Sync Workflow
1. **Automatic check**: `omo-profiler schema-check` CLI command or TUI Schema Check view
2. `CompareSchemas()` fetches upstream via HTTP, compares against `go:embed`'d `schema.json`
3. If different, generates unified diff via `diff.ComputeUnifiedDiff`
4. User can save diff to a folder via `SaveDiff()` (timestamped: `schema-diff-YYYYMMDD-HHMMSS.diff`)
5. To actually update: manually replace `internal/schema/schema.json` with upstream content, then rebuild

**Note**: The old `update-schema.sh` script no longer exists. Schema comparison is now fully in Go code.

## Config Schema Authority
`internal/config/types.go` is the **source of truth** for the Go struct representation:
- Must match upstream schema 1:1
- JSON tags must be exact matches
- Use `*bool` for optional booleans (distinguish `false` from missing)
- Use `json.RawMessage` for flexible fields
- All tags require `omitempty`
- No methods on data structs

## Anti-Patterns
- **Schema Divergence**: Never add `Config` fields without upstream schema support
- **Manual Schema Edits**: Don't modify `schema.json` without verifying against upstream
- **Skipping Validation**: Always validate imported profiles before saving
- **Hardcoded Schema URL**: Use `config.DefaultSchema` or `schema.UpstreamSchemaURL` constants
