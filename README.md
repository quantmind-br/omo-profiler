# omo-profiler

TUI profile manager for oh-my-openagent

## Installation

```bash
go install github.com/diogenes/omo-profiler/cmd/omo-profiler@latest
```

## Quick Start

```bash
# Launch TUI
omo-profiler

# Launch web UI (recommended) — http://127.0.0.1:4747
omo-profiler web

# List profiles
omo-profiler list

# Show current profile
omo-profiler current

# Switch profile (emits the shell command that activates it)
omo-profiler switch <profile-name>

# Import profile
omo-profiler import <file.json>

# Export profile
omo-profiler export <profile-name> <output.json>
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `omo-profiler` | Launch TUI |
| `omo-profiler web` | Launch the web UI (recommended) |
| `omo-profiler list` | List all profiles |
| `omo-profiler current` | Show active profile |
| `omo-profiler switch <name>` | Apply profile by substituting its keys into `~/.omo/omo.json` |
| `omo-profiler import <file>` | Import profile from JSON |
| `omo-profiler export <name> <path>` | Export profile to file |

## Web UI

`omo-profiler web` starts a local web server (default `http://127.0.0.1:4747`) with a
browser UI that reaches parity with every TUI screen: dashboard, profiles
(switch/create/clone/rename/import/export/delete), a schema-driven editor with a
validated raw-JSON tab, side-by-side diff, the model registry with models.dev
import, and the schema drift check.

```bash
omo-profiler web                 # bind 127.0.0.1:4747, open the browser
omo-profiler web --port 8080     # custom port
omo-profiler web --host 0.0.0.0  # override bind (loopback by default)
omo-profiler web --no-open       # do not launch a browser
```

The web UI is the recommended interface; the TUI remains the default
no-subcommand run but is slated for future deprecation.

The server binds loopback by default because it edits your local config. The
editor is **schema-driven** — it renders forms from the `[opencode]` sub-schema
(`schema.GetOpenCodeSchema()`), so upstream field additions appear automatically
after re-syncing `internal/schema/schema.json` and rebuilding.

Building the UI requires Node. `make build-web` builds the frontend and then the
binary with the SPA embedded; `make install` does the same before installing. A
plain `make build` stays Node-free and serves a "Web UI not built" placeholder
until you run `make build-web`.

## Features

- Dashboard with active profile overview
- Create profiles with step-by-step wizard
- Edit existing profiles
- Compare profiles side-by-side
- Import/export profiles
- Schema validation against oh-my-openagent (`omo.schema.json`)
- Automatic backups before mutating writes to `~/.omo/omo.json`

## Config Location

Unified document: `~/.omo/omo.json` (or `omo.jsonc` when present — preferred).

A profile is `profiles.<name>` inside that document; its editable payload is
`profiles.<name>.[opencode]`. There is no `profiles/` directory and no
file-per-profile.

Activation is in-document: `omo-profiler switch` substitutes the profile's keys directly into the document root. The profile is live as soon as the command returns — no environment variable, no shell command. If the root matches no profile, the previous configuration is snapshotted as `profiles.base` before being overwritten.

Model registry (omo-profiler local state): `~/.omo/models.json`
