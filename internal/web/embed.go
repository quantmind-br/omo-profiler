package web

import "embed"

// distFS holds the built frontend SPA. The all: prefix embeds dotfiles so a
// committed frontend/dist/.gitkeep makes this compile before the UI is built.
//
//go:embed all:frontend/dist
var distFS embed.FS

// defaultTemplate is the bundled seed for the "Default template" new-profile
// option. Keep it identical to repo-root template/opencode-profile.json: a flat
// `[opencode]` payload written into `profiles.<name>.[opencode]` (no
// document-root `$schema` — that lives on the omo.json document root).
//
//go:embed assets/default-template.json
var defaultTemplate []byte

// DefaultTemplate returns the bundled default profile template bytes.
func DefaultTemplate() []byte { return defaultTemplate }
