# Repository Guidelines

Read `openwiki/quickstart.md` first, then follow its links to the relevant architecture, workflow, domain, operation, and testing notes.

## Project Overview

`omo-profiler` manages named configuration profiles for **oh-my-openagent**. Three front-ends over one domain core: a Bubble Tea **TUI** (default when run bare), a Cobra **CLI** (9 subcommands), and a **web UI** (`omo-profiler web`) — a Go HTTP server with an embedded React SPA.

Profiles live as `profiles.<name>.[opencode]` blocks inside a **single** document at `~/.omo/omo.json` (or `omo.jsonc`). There is no file-per-profile layout. Activation is **in-document**: `switch` substitutes profile keys into the root, so the profile is live as soon as the command returns.

Module `github.com/diogenes/omo-profiler`, Go 1.25.6.

## Architecture & Data Flow

`cmd/omo-profiler/main.go` → `cli.Execute()` → root Cobra command. No subcommand → `tui.Run()`; otherwise dispatch to `internal/cli/cmd/`.

Layering is strict and downward-only: **`config`** (file I/O, JSONC strip, atomic write) → **`profile`** (CRUD, activation, sparse marshal) → front-ends (`tui`, `cli/cmd`, `web`). `schema` validates, and `models`/`backup`/`diff` are leaf services. Nothing below reaches upward.

**Two config granularities — do not conflate:**

- `config.Document` — the whole omo file, held as a raw `map[string]json.RawMessage`.
- `config.Config` — only the flat `[opencode]` harness block, 46 typed fields.

`config.OpenCodeKey` is `"[opencode]"`. A profile block may also carry **sibling harness keys** (`[senpi]`, `[codex]`, …) that `Config` does not model; clone, rename, and save must preserve them verbatim.

**Write path:** every mutation runs through a mutex-guarded load→apply→save transaction (`config.Mutate` / `MutateWithPreSave`), then `Document.Save()` → `WriteFileAtomic` (temp file in the same dir, `Sync`, `Chmod`, `Rename`). `backup.CreateOmoIfPresent` is the `preSave` hook.

**Activation:** `profile.Apply(name)` copies every key in `profiles.<name>` over the matching document root key — verbatim, no merging. The active profile is detected by comparing the root against stored profiles (`profile.ActiveName`), so the root *is* the effective configuration. If the root matches no profile, it is snapshotted as `profiles.base` before being overwritten, so a hand-edited or never-saved config is never destroyed. No environment variables, no hint file, no shell command.

## Key Directories

| Path | Owns |
|---|---|
| `internal/config/` | `Document`, `Config`, path helpers, JSONC, atomic write |
| `internal/profile/` | CRUD, activation, naming, `FieldSelection`, `MarshalSparse` |
| `internal/schema/` | Embedded `schema.json`, validator singleton, upstream compare |
| `internal/models/` | Model registry (`~/.omo/models.json`) + models.dev client |
| `internal/backup/` | Pre-write snapshots |
| `internal/diff/` | Side-by-side + unified diff |
| `internal/tui/` | Root `App` (10 states), styles, keys, layout |
| `internal/tui/views/` | 25 view/step files; the 6-step wizard dominates |
| `internal/web/` | HTTP mux, handlers, `//go:embed` of the SPA |
| `internal/web/frontend/` | React 18 + Vite 6 + TypeScript + Tailwind |
| `internal/testdata/` | JSON fixtures |
| `template/` | Starter `[opencode]` block |

Seven packages carry their own `AGENTS.md` (`config`, `profile`, `schema`, `models`, `cli/cmd`, `tui`, `tui/views`) — read the local one before editing that package.

## Development Commands

```bash
make build      # go build -v -o omo-profiler ./cmd/omo-profiler
make test       # go test -v ./...   (no -race)
make lint       # golangci-lint run ./...   (external tool; no .golangci.yml in repo)
make web-build  # npm run build in internal/web/frontend   (requires Node)
make build-web  # web-build then build → binary with SPA embedded
make install    # depends on build-web, so this needs Node; cp to ~/.local/bin
make clean
```

```bash
go test ./internal/profile/...                   # narrow loop
go test -run TestLoad ./internal/profile/
cd internal/web/frontend && npm run dev          # Vite dev server, proxies /api → :4747
./omo-profiler web --port 4747                   # Go API alongside Vite
```

CLI surface: `list`, `current`, `create [--from]`, `switch`, `import [--name]`, `export [-f]`, `models list|add|edit|delete`, `schema-check --output`, `web [--host --port --no-open]`.

## Code Conventions & Common Patterns

### Go

- **Errors:** always wrap — `fmt.Errorf("parse %s: %w", path, err)`. Sentinels `profile.ErrInvalidName`/`ErrEmptyName`; typed `*profile.NotFoundError` unwraps to `fs.ErrNotExist`, so test with `errors.Is(err, fs.ErrNotExist)`, never `os.IsNotExist`.
- **Never load-and-save a document ad hoc** — go through the transaction helper:

  ```go
  return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
      return profile.WriteOpenCodeBlockInto(doc, name, openCode)
  })
  ```

- **Pointer-for-optional:** every optional scalar in `Config` is `*bool`/`*int64`/`*float64` with `omitempty`, so absent is distinguishable from `false`/`0`. Every field needs a `json:"...,omitempty"` tag.
- **`json.RawMessage`** for unmodeled sub-trees (`Skills`, `RalphLoop`, `RuntimeFallback`) — they round-trip verbatim. **`interface{}`** for upstream union types (`FallbackModels`, `PermissionConfig.Bash`, `GitMasterConfig.CommitFooter`).
- **Paths:** never hardcode `~/.omo`. Use `config.OmoDir()`, `OmoFile()`, `ModelsFile()`, `EnsureDirs()`. The `Legacy*` helpers exist only for migration detection.
- **Validation:** always via the `schema.GetValidator()` singleton. Three targets (`Config`, `[opencode]` JSON, whole document) × two modes — strict (`Validate`, `ValidateJSON`, `ValidateDocument`) and permissive (`…ForSave`, which ignores `required`, `additionalProperties`, and minimum-on-zero). Wizard and profile saves use the permissive pair. Editor forms use `schema.GetOpenCodeSchema()`, not `GetEmbeddedSchema()`.

### TUI (Bubble Tea, pure MVU)

- `App.Update` is the only router. Views emit messages and never mutate `App`.
- **Navigation recreates the view:** a `NavTo*Msg` makes `App` build a fresh `New*()`, call `SetSize`, then `navigateTo(state)` → `Init()`. All view-local state (cursor, scroll, filter text) is intentionally discarded. Only the wizard carries state across steps, via `SetConfig`/`Apply` on transitions.
- **All I/O in `tea.Cmd` factories** on `App`, never inline in `Update`/`View`. Three legacy call sites in `app.go` still call `profile.Load`/`GetActive` synchronously in `Update` — don't copy that shape.
- **Wizard:** Name → Categories → Agents → Hooks → Other → Review. The `WizardStep` interface is only `Init`/`SetSize`/`View`; `SetConfig(*config.Config, *profile.FieldSelection)` and `Apply(...)` are called by the orchestrator on concrete types because `WizardName` uses `SetName`/`GetName` instead.
- **Text-capture guards:** before intercepting `q`, `?`, or `esc`, `App` consults per-state guards (`IsFiltering`, `IsEditing`, `IsFocused`, `IsCapturing`) so keys reach inputs. Adding a text-input view means adding a guard in three places in `app.go`.
- **Keys:** global `q`/`ctrl+c` quit, `?` help, `esc` back, arrows or `kjhl`, `enter` select. List-shaped views use `n`/`d`/`/`/`e`; collapsible editors use `+` add, `-` delete, `space` toggle, `enter` expand; the wizard uses `tab`/`shift+tab` and `ctrl+s` on Review.
- Import styles from `internal/tui/styles.go` and sizing from `internal/tui/layout/` — never define hex colors locally (the existing views violate this; do not extend it).

### Web

- Handlers are uniform `func handleXyz(w http.ResponseWriter, r *http.Request)`, responding via `writeJSON` / `writeErr`; route params via `r.PathValue(...)`. Routes are all registered in `newMux()` (`server.go`), with the SPA fallback last so `/api/*` wins.
- **Profiles are stored verbatim.** `PUT /api/profiles/{name}` validates, type-checks, then persists the raw request bytes without re-marshalling — so `[]`, `""`, `false` survive and omitting a key deletes it.
- `GET /api/schema` serves the **`[opencode]` sub-schema** (what the form renders); `/api/document-schema` serves the whole-document schema.
- React: server state through `@tanstack/react-query` (keys `['active']`, `['profile', name]`, `['schema']`, `['models']`); all fetches through the typed wrappers in `src/lib/api.ts`, never bare `fetch` in components; Tailwind classes merged with `cn()`.
- **The editor is schema-driven.** `SchemaForm.tsx` walks `schema.properties` and picks a widget per field. New `[opencode]` fields appear automatically once `internal/schema/schema.json` is re-embedded — never hand-write form fields.

## Important Files

| File | Why |
|---|---|
| `internal/cli/root.go` | Registers subcommands; launches the TUI when bare |
| `internal/config/types.go` | `Config` — must stay 1:1 with the `[opencode]` sub-schema |
| `internal/config/document.go` | `Document`, `Mutate`, `WriteFileAtomic`, `OpenCodeKey` |
| `internal/config/paths.go` | All path helpers, `SetBaseDir`/`ResetBaseDir` |
| `internal/profile/active.go` | `Apply`, `ActiveName`, `GetActive` — in-document substitution + detection |
| `internal/profile/sparse.go`, `selection.go` | `MarshalSparse`, `FieldSelection`, field-path tables |
| `internal/schema/validator.go` | Singleton + the six validation entry points |
| `internal/schema/compare.go` | Upstream URL, fetch, compare (`.upstream-sha` is the sync anchor) |
| `internal/tui/app.go` | State router, nav handlers, `tea.Cmd` factories |
| `internal/tui/views/wizard.go` | 6-step orchestrator and step lifecycle |
| `internal/web/server.go`, `embed.go` | Route table; `//go:embed all:frontend/dist` |
| `omo.schema.json` (root, 618 KB) | Full upstream schema copy — never read whole |

## Runtime/Tooling Preferences

- **Go ≥ 1.25.6.** Deps: bubbletea/bubbles/lipgloss, cobra, gojsonschema, go-diff, testify, sahilm/fuzzy.
- **npm** is the frontend package manager — `package-lock.json` is the only lockfile. Do not introduce bun/pnpm/yarn.
- Frontend scripts are exactly `dev`, `build` (`tsc && vite build`), `preview`. TypeScript is `strict`.
- Node is needed only for the web UI — and for `make install`, which depends on `build-web`.
- **No CI, no `.github/workflows`, no Dockerfile, no `.golangci.yml`.** `make lint` shells out to a globally installed `golangci-lint`. Quality gates are local only.
- Root `AGENTS.md` and `CLAUDE.md` are kept byte-identical — update both together.

## Testing & QA

Stdlib `testing`; use `testify/require` where the file already does and plain `t.Fatalf` where it doesn't. Tests are co-located. `cmd/omo-profiler` and root `internal/cli` have no tests, and there are **no frontend tests**. No `-race`, no build tags, no golden files.

**Any test touching the filesystem must redirect the base dir** — and clear activation env, or the suite leaks into it:

```go
func setupTestEnv(t *testing.T) {
    t.Helper()
    config.SetBaseDir(t.TempDir())
    t.Cleanup(config.ResetBaseDir)

    t.Setenv(EnvOmoProfile, "")
    t.Setenv(EnvOcxProfile, "")
    t.Setenv(EnvOpenCodeConfigDir, "")
}
```

**Seed profiles into the document**, never as standalone files:

```go
func seedProfile(t *testing.T, name, openCodeJSON string) {
    t.Helper()
    doc, err := config.LoadDocument()
    require.NoError(t, err)
    block, err := json.Marshal(map[string]json.RawMessage{
        config.OpenCodeKey: json.RawMessage(openCodeJSON),
    })
    require.NoError(t, err)
    require.NoError(t, doc.SetProfileBlock(name, block))
    doc.EnsureSchema()
    require.NoError(t, doc.Save())
}
```

Web handlers are exercised at the HTTP layer with `httptest` against `newMux()` — no live server. Fixtures in `internal/testdata/` cover a full config, a minimal one, the `json.RawMessage` preservation path (`skills-object.json`), and the `interface{}` union path (`complex-permissions.json`).

Run the narrowest package tests you touched, then `make test`. For TUI changes also run the binary and exercise the changed view — lost guards and blank viewports don't surface in unit tests.

## Landmines

1. **A fresh clone does not compile.** `dist/*` is gitignored and the un-ignored `dist/.gitkeep` is missing, so `//go:embed all:frontend/dist` has nothing to embed. Run `make web-build` (or create the placeholder) before `go build`.
2. **Two save paths — choose deliberately.** `Profile.Save()` marshals through typed `Config`, where `omitempty` silently drops explicit zero values. For payloads the caller already produced (wizard sparse save, import, web `PUT`) use `profile.SaveOpenCodeBlock` / `WriteOpenCodeBlockInto` / `UpdateOpenCodeBlock` with raw `json.RawMessage`.
3. **`Document.Save()` strips JSONC comments and sorts keys alphabetically.** A hand-annotated `omo.jsonc` is flattened on the next write.
4. **`Apply` substitutes verbatim** — it copies profile keys over the root with no merging, so a sparse profile yields a sparse root config. The snapshot step (`profiles.base`) is what makes that recoverable; don't "fix" it to merge.
5. **Changing `Config` needs three parallel edits:** the struct in `config/types.go`, `knownConfigTags` in `profile/profile.go`, and the field-path tables in `profile/selection.go`. Drift causes false validation errors and silently unsaved wizard fields.
6. **Backups never auto-rotate.** `CreateOmoIfPresent` snapshots before every mutation; nothing calls `backup.Clean`.

`ARCHITECTURE.md` and `TUI_AUDIT.md` predate current code — prefer `openwiki/` and the source.
