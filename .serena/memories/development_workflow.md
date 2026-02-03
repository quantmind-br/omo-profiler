# Development Workflow

## Build Commands
```bash
make build       # Build binary to ./omo-profiler
make install     # Install to ~/.local/bin/omo-profiler
make test        # Run all tests (race detector enabled)
make lint        # Run golangci-lint
make clean       # Remove build artifacts
make help        # Show available targets
```

## Direct Go Commands
```bash
go build -v -o omo-profiler ./cmd/omo-profiler
go test -v ./...
go test -race ./...  # With race detector
```

## Project Conventions

### Code Organization
- **Go Version**: Strictly 1.25.6
- **Package Structure**: Internal packages only (no exported API)
- **Error Handling**: Bubble Tea commands return `Msg` with error field
- **Naming**: Follow Go conventions (CamelCase for exported, camelCase for internal)

### Adding New Features

#### Add CLI Command
1. Create file in `internal/cli/cmd/<command>.go`
2. Register in `internal/cli/root.go` `init()` function

#### Modify Config Schema
1. Update `internal/config/types.go`
2. **CRITICAL**: Must match `oh-my-opencode.json` upstream schema
3. Use `omitempty` tags to avoid dirty config files
4. Use pointers for optional boolean fields

#### Add TUI View
1. Create file in `internal/tui/views/<view>.go`
2. Add state to `appState` enum in `internal/tui/app.go`
3. Add view field to `App` struct
4. Handle navigation message in `App.Update()`
5. Add case in `App.View()` switch
6. Add help text in `renderShortHelp()` and `renderFullHelp()`

#### Add Wizard Step
1. Create `internal/tui/views/wizard_<step>.go`
2. Implement implicit interface: `SetConfig()`, `Apply()`, `SetSize()`, `Init()`
3. Add step constant to `wizard.go`
4. Add field to `Wizard` struct
5. Initialize in `NewWizard()`
6. Add navigation case in `Wizard.Update()`
7. Add view case in `Wizard.View()`

## Testing Requirements
- All tests must use `config.SetBaseDir(t.TempDir())`
- No global side effects allowed
- Tests should clean up with `defer config.ResetBaseDir()`
- Use `t.TempDir()` for automatic cleanup

## Dependencies
Key external dependencies:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Pre-built components
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/spf13/cobra` - CLI framework
- `github.com/xeipuuv/gojsonschema` - JSON schema validation
- `github.com/sergi/go-diff` - Diff generation
- `github.com/stretchr/testify` - Test assertions
