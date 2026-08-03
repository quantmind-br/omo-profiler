# INTERNAL/PROFILE

## OVERVIEW

Core business logic for profile persistence inside the unified omo document, and for resolving/emitting activation. Handles CRUD on `profiles.<name>` and reports what the environment currently activates — it never copies or symlinks config files.

## FILES

| File | Role |
|------|------|
| `profile.go` | `Profile` struct, `Load`, `Save`, `Delete`, `List`, `Exists`, `WriteInto`, `WriteOpenCodeBlockInto`, `SaveOpenCodeBlock` |
| `active.go` | `Apply`, `ActiveName`, `GetActive`, `canonicalJSON` |
| `sparse.go` | `MarshalSparse` — emit selected `[opencode]` fields (keeps explicit zeros) |
| `selection.go` | `FieldSelection` — wizard/editor field-path selection with wildcards |
| `naming.go` | `SanitizeName`, `ValidateName` — strict regex `^[a-zA-Z0-9_-]+$` |
| `profile_test.go` | CRUD + naming tests with `setupTestEnv` helper |
| `active_test.go` | In-document activation tests — `Apply` substitution, snapshot, `ActiveName` detection, malformed-JSON error propagation |
| `sparse_test.go` / `selection_test.go` | Sparse marshal + selection path tests |

## KEY TYPES

- `Profile`: `profiles.<name>` entry. `Config` mirrors `profiles.<name>.[opencode]`. `Path` is the omo document path (informational). `PreservedUnknown` keeps unknown keys **inside** `[opencode]`; `PreservedBlock` keeps sibling keys of `[opencode]` in the profile block. Also: `FieldPresence`, `HasLegacyFields`, `LegacyFieldsWarning`.
- `NotFoundError`: missing `profiles.<name>`; unwraps to `fs.ErrNotExist`, so test it with `errors.Is(err, fs.ErrNotExist)`. `os.IsNotExist` does **not** detect it — that helper predates error wrapping and only unwraps `*PathError`/`*LinkError`/`*SyscallError`.
- `ActiveConfig`: result of `GetActive()` — root `[opencode]` block, `ProfileName` (detected by root comparison), `Modified` (root matches no profile).
- `Applied`: result of `Apply()` — `Name` (profile applied) and `Snapshot` (profile name capturing the previous root, empty when the root already matched).

## PERSISTENCE

Profiles are **not files**. They are blocks inside `~/.omo/omo.json` (or `omo.jsonc`):

```
profiles.<name>.[opencode]   ← editable flat config (config.Config)
profiles.<name>.<siblings>   ← round-tripped via PreservedBlock
```

| Function | Behavior |
|----------|----------|
| `Load(name)` / `LoadFromDocument` | Read `profiles.<name>.[opencode]`, preserve unknowns/siblings, build `FieldPresence` |
| `Save` / `(p).Save` | Read-modify-write the omo document; other profiles/harness keys untouched. Marshals `Config` with `omitempty` — **not** for sparse wizard saves |
| `WriteInto(doc)` | Stage into a document without persisting |
| `WriteOpenCodeBlockInto(doc, name, openCode)` | Stage a pre-marshalled `[opencode]` payload; preserve profile sibling keys |
| `SaveOpenCodeBlock(name, openCode)` | Persist a pre-marshalled `[opencode]` payload (use after `MarshalSparse`) |
| `Delete(name)` | Remove `profiles.<name>` from the document |
| `List()` | Sorted profile names from the document |
| `Exists(name)` | `doc.HasProfile(name)` |
| `Create(name, cfg)` | Create, `*ExistsError` if taken — check and write in one transaction |
| `CreateFrom(name, fromName)` | Clone the whole block **verbatim** (no `omitempty` round-trip) |
| `ExportOpenCode(name)` | Read the stored `[opencode]` block **verbatim**, pretty-printed — the only export path (CLI, TUI, web) |
| `UpdateOpenCodeBlock(name, raw)` | Replace an existing profile's block; `*NotFoundError` if gone — **never creates** |
| `CreateAvailable(base, openCode)` | Claim `base` or the first free `base-N` **inside** the transaction; used by every import |
| `Rename(oldName, newName)` | Move the block in one write |

## VERBATIM I/O

Anything that copies a profile whole — export, import, clone, the web editor's
PUT — moves **raw block bytes**, never a re-marshalled `config.Config`. Every
field is `omitempty`, so a typed round-trip silently drops explicitly present
zero values (`"disabled_mcps": []`, `"default_run_agent": ""`) and any key this
build's struct does not model. Typed unmarshalling is still used to *validate*
a payload; it must not be what gets stored.

Writes **replace** the block rather than merging into it. Merging made unknown
keys sticky — the editor could never delete one. Clients that replace (the web
editor GETs the verbatim block and PUTs it back) round-trip losslessly.

## MUTATION CONTRACT

Every profile lives in one document, so a write rewrites the whole file. All
read-modify-write cycles MUST go through `config.Mutate` — directly, or via the
functions above. NEVER `LoadDocument` → edit → `Save` outside it: concurrent
cycles silently drop updates (the web server serves requests on separate
goroutines).

Check-then-write across two transactions is the same bug: `Exists()` followed by
`Save()` can lose the name in between. That is why `Create`, `CreateFrom` and
`CreateAvailable` exist.

Scope: one process; the pre-write backup is the recovery path for the rest.

Editing an existing profile asserts it still exists: the wizard and PUT both
fail with `*NotFoundError` when it was deleted during the edit, instead of
resurrecting it.

## BACKUPS

Every mutating entry point snapshots the document first, via
`config.MutateWithPreSave(backup.CreateOmoIfPresent, fn)` — one implementation,
no inline `Stat` + `backup.Create`, and **never** a bare `config.Mutate` for a
mutating path.

The snapshot runs **inside** the transaction lock, immediately before the write.
Taking it earlier is the bug it replaced: concurrent writers all captured the
same state, so intermediate versions were unrecoverable. Backup names carry
nanosecond precision for the same reason — second-precision names collided and
overwrote each other under load. Precision is not the guarantee, though:
`backup.Create` claims each name with `O_CREATE|O_EXCL` and advances a
nanosecond on collision, so two writers reading the same clock (coarse clock
sources, or a second omo-profiler process) cannot clobber each other's
pre-image.

### Sparse writes

Wizard/editor sparse saves must not go through `Profile.Save` / `WriteInto` (those drop explicit zero values via `omitempty`). Path:

1. `MarshalSparse(cfg, selection, preservedUnknown)` → `[opencode]` JSON
2. `WriteOpenCodeBlockInto` / `SaveOpenCodeBlock` → write into `profiles.<name>`

## ACTIVATION (in-document substitution)

Activation is in-document: `Apply(name)` copies every key in `profiles.<name>` over the matching document root key — verbatim, no merging. The active profile is detected by comparing the root against stored profiles (`ActiveName`), so the root *is* the effective configuration.

If the root matches no profile, it is snapshotted as `profiles.base` (colliding to `base-1`, `base-2`, …) before being overwritten, so a hand-edited or never-saved config is never destroyed.

`Apply(name) (Applied, error)`:
1. Loads `profiles.<name>` via `doc.ProfileBlock(name)`
2. Calls `ActiveName(doc)` — if it returns `""` and a declared key is present/non-empty at root, snapshots those root keys as `profiles.base`
3. Copies each declared key over the root via `doc.SetRaw(key, value)`
4. `doc.EnsureSchema()`

`GetActive()`:
1. Loads the document
2. `ActiveName(doc)` — first profile whose every declared key canonicalizes equal to the root
3. Decodes the root `[opencode]` straight into `Config` (no merge)
4. `Modified = ProfileName == "" && root [opencode] is present and non-empty`

## NAMING VALIDATION

- Regex: `^[a-zA-Z0-9_-]+$` (alphanumeric, underscores, hyphens only)
- `SanitizeName`: Strips invalid chars, trims leading/trailing separators
- Empty names rejected

## LEGACY FIELD DETECTION

`detectLegacyFields`: Scans the `[opencode]` JSON for fields not in `config.Config`. Sets `HasLegacyFields` + human-readable `LegacyFieldsWarning` on load.

## ANTI-PATTERNS

- **Copy/symlink switching**: DO NOT copy profile content over a live file or use symlinks — activation is in-document via `Apply`
- **Claiming activation**: DO NOT tell the user a profile is active after `Apply`; the document is mutated, so the profile is already live — no shell command to run
- **Merging on apply**: DO NOT deep-merge profile keys onto the root — `Apply` substitutes verbatim; a sparse profile yields a sparse root, and the snapshot step is what makes that recoverable
- **Raw File Access**: Avoid `os.Open` for profiles; use `profile.Load()` / `config.Document`
- **Skip EnsureDirs**: Always call `config.EnsureDirs()` before writes
- **Sparse via Profile.Save**: DO NOT persist `MarshalSparse` output through `Profile.Save`/`WriteInto`; use `WriteOpenCodeBlockInto` / `SaveOpenCodeBlock`
