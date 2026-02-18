# Development Workflow

## Build Commands
```bash
make build       # Build binary to ./omo-profiler
make install     # Install to ~/.local/bin/omo-profiler
make uninstall   # Remove from ~/.local/bin
make test        # Run all tests with -v flag (NO race detector by default)
make lint        # Run golangci-lint (no .golangci.yml config file exists)
make clean       # Remove build artifacts
make help        # Show available targets
```

## Direct Go Commands
```bash
go build -v -o omo-profiler ./cmd/omo-profiler
go test -v ./...
go test -race ./...  # With race detector (not in Makefile default)
```

## No CI/CD
- No `.github/workflows/`, `.gitlab-ci.yml`, or Docker files
- No `.golangci.yml` linter configuration
- No Goreleaser config
- `update-schema.sh` no longer exists — replaced by `schema.CompareSchemas()` in Go code

## Project Conventions

### Code Organization
- **Go Version**: Strictly 1.25.6
- **Package Structure**: Internal packages only (no exported API)
- **Error Handling**: Bubble Tea commands return `Msg` with error field; CLI uses Cobra `RunE`
- **Naming**: Follow Go conventions (CamelCase for exported, camelCase for internal)
- **JSON Tags**: All fields use `omitempty`; `*bool` for optional booleans

### Adding New Features

#### Add CLI Command
1. Create file in `internal/cli/cmd/<command>.go`
2. Export a `var <Name>Cmd` Cobra command
3. Register in `internal/cli/root.go` `init()` function with `rootCmd.AddCommand(cmd.<Name>Cmd)`
4. Current commands: ListCmd, CurrentCmd, ExportCmd, SwitchCmd, ImportCmd, ModelsCmd, CreateCmd, SchemaCheckCmd

#### Modify Config Schema
1. Update `internal/config/types.go`
2. **CRITICAL**: Must match `oh-my-opencode.json` upstream schema
3. Use `omitempty` tags to avoid dirty config files
4. Use pointers for optional boolean fields
5. Run `omo-profiler schema-check` to verify alignment with upstream

#### Add TUI View
1. Create file in `internal/tui/views/<view>.go`
2. Add state to `appState` enum in `internal/tui/app.go`
3. Add view field to `App` struct
4. Handle navigation message in `App.Update()`
5. Add case in `App.View()` switch
6. Add help text in `renderShortHelp()` and `renderFullHelp()`

#### Add Wizard Step
1. Create `internal/tui/views/wizard_<step>.go`
2. Implement `WizardStep` interface: `Init()`, `SetSize()`, `View()`
3. Implement implicit methods: `SetConfig()` (or `SetName`/`GetName` for name step), `Apply()`
4. Add step constant to `wizard.go` (iota enum)
5. Add field to `Wizard` struct
6. Initialize in `NewWizard()`
7. Add navigation case in `Wizard.Update()`
8. Add view case in `Wizard.View()`

## Testing Requirements
- All tests must use `config.SetBaseDir(t.TempDir())`
- No global side effects allowed
- Tests should clean up with `defer config.ResetBaseDir()`
- Use `t.TempDir()` for automatic cleanup
- Assertions: `github.com/stretchr/testify`

## Dependencies
Key external dependencies:
- `github.com/charmbracelet/bubbletea` v1.3.10 — TUI framework
- `github.com/charmbracelet/bubbles` v0.21.0 — Pre-built components
- `github.com/charmbracelet/lipgloss` v1.1.0 — Styling
- `github.com/spf13/cobra` v1.10.2 — CLI framework
- `github.com/xeipuuv/gojsonschema` v1.2.0 — JSON schema validation
- `github.com/sergi/go-diff` v1.4.0 — Diff generation
- `github.com/stretchr/testify` v1.11.1 — Test assertions
- `github.com/mattn/go-runewidth` (indirect) — Used in layout truncation
