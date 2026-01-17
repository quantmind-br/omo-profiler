# omo-profiler

TUI profile manager for oh-my-opencode

## Installation

```bash
go install github.com/diogenes/omo-profiler/cmd/omo-profiler@latest
```

## Quick Start

```bash
# Launch TUI
omo-profiler

# List profiles
omo-profiler list

# Show current profile
omo-profiler current

# Switch profile
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
| `omo-profiler list` | List all profiles |
| `omo-profiler current` | Show active profile |
| `omo-profiler switch <name>` | Activate profile |
| `omo-profiler import <file>` | Import profile from JSON |
| `omo-profiler export <name> <path>` | Export profile to file |

## Features

- Dashboard with active profile overview
- Create profiles with step-by-step wizard
- Edit existing profiles
- Compare profiles side-by-side
- Import/export profiles
- Schema validation against oh-my-opencode
- Automatic backups on profile switch

## Profile Location

Profiles are stored in `~/.config/opencode/profiles/`

Active config: `~/.config/opencode/oh-my-opencode.json`
