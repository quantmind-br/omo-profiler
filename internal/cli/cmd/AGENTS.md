# CLI COMMANDS KNOWLEDGE BASE

## OVERVIEW

Cobra CLI commands for headless profile management. Thin wrappers delegating to `internal/profile` and `internal/backup`.

## COMMANDS

| Command | File | Args | Action |
|---------|------|------|--------|
| `list` | `list.go` | none | Tabulates profiles from the omo document; marks the applied profile with `*` |
| `current` | `current.go` | none | Prints the profile matching the root of `~/.omo/omo.json` via `profile.GetActive()`; warns if the root matches no profile |
| `switch` | `switch.go` | `<name>` | `profile.Apply` — substitutes profile keys into the document root with a pre-write backup |
| `import` | `import.go` | `<file>` | Loads JSON → `profiles.<name>` in the omo document (backup `OmoFile` before mutate) |
| `export` | `export.go` | `<name> <path>` | Writes the profile's `[opencode]` payload to a JSON file |
| `create` | `create.go` | `<name>` | Headless clone via `--from` flag; wizard logic stays in TUI |
| `models` | `models.go` | none | Lists LLM providers and models from registry |
| `schema-check` | `schema_check.go` | none | Validates / compares against upstream `assets/omo.schema.json` |

## REGISTRATION

All commands exported as `var XxxCmd = &cobra.Command{...}` and registered in `internal/cli/root.go` `init()`:
```go
rootCmd.AddCommand(cmd.ListCmd)
rootCmd.AddCommand(cmd.SwitchCmd)
// etc.
```

## CONVENTIONS

- **Exported Vars**: Commands must be exported (e.g., `SwitchCmd`) for registration
- **Thin Wrappers**: Minimal logic; delegate to `internal/profile` or `internal/backup`
- **RunE**: Prefer `RunE` over `Run` for Cobra error propagation
- **Args Validation**: Use `cobra.ExactArgs(n)` or `cobra.MaximumNArgs(n)`
- **StdErr**: Print errors to `os.Stderr` and exit code 1

## ANTI-PATTERNS

- **Fat Commands**: No business logic or FS operations in `Run` — delegate to packages
- **Interactive Prompts**: CLI must be non-interactive; use TUI for wizards
- **Global State**: Avoid modifying globals; use flags and arguments
- **Raw Paths**: Never hardcode `~/.omo`; use `config.OmoDir` / `OmoFile` / `ModelsFile` helpers
- **Claiming switch activated**: `Apply` mutates the document, so the profile is live — no shell command to emit
