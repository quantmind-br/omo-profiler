package profile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/diogenes/omo-profiler/internal/backup"
	"github.com/diogenes/omo-profiler/internal/config"
)

// Profile is a single `profiles.<name>` entry of the omo document.
//
// Config mirrors the profile's `[opencode]` harness block — the flat
// pre-unification config shape. Sibling keys of that block (shared typed keys,
// `[senpi]`, `[codex]`) are not modelled here; they round-trip via
// PreservedBlock so editing a profile never drops them.
type Profile struct {
	Name                string
	Config              config.Config
	Path                string
	PreservedUnknown    map[string]json.RawMessage `json:"-"`
	PreservedBlock      map[string]json.RawMessage `json:"-"`
	FieldPresence       map[string]bool            `json:"-"`
	HasLegacyFields     bool                       `json:"-"`
	LegacyFieldsWarning string                     `json:"-"`
}

// CloneAs returns a snapshot of p to be saved under name. It reproduces the
// whole profile block — `[opencode]`, its unknown keys, and the sibling blocks
// kept in PreservedBlock — so cloning never silently drops `[senpi]`,
// `[codex]`, or shared typed keys.
//
// The copy is shallow: Config's maps/slices/pointers and the json.RawMessage
// values are shared with p, and only the preserved maps themselves are fresh
// (so adding or removing a key on one side does not affect the other). This is
// a write-then-discard snapshot for Save; do not mutate it in place expecting p
// to be unaffected.
func (p *Profile) CloneAs(name string) *Profile {
	return &Profile{
		Name:             name,
		Config:           p.Config,
		PreservedUnknown: copyRawMap(p.PreservedUnknown),
		PreservedBlock:   copyRawMap(p.PreservedBlock),
	}
}

func copyRawMap(src map[string]json.RawMessage) map[string]json.RawMessage {
	if src == nil {
		return nil
	}
	dst := make(map[string]json.RawMessage, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// NotFoundError reports a missing `profiles.<name>` entry. It unwraps to
// fs.ErrNotExist, so callers test it with errors.Is(err, fs.ErrNotExist).
// Note os.IsNotExist does NOT work here: it predates error wrapping and only
// unwraps *PathError/*LinkError/*SyscallError.
type NotFoundError struct{ Name string }

func (e *NotFoundError) Error() string { return fmt.Sprintf("profile %q not found", e.Name) }

func (e *NotFoundError) Unwrap() error { return fs.ErrNotExist }

var knownConfigTags = []string{
	"$schema",
	"disabled_mcps",
	"disabled_agents",
	"disabled_skills",
	"disabled_hooks",
	"disabled_commands",
	"hashline_edit",
	"model_fallback",
	"agents",
	"categories",
	"claude_code",
	"sisyphus_agent",
	"comment_checker",
	"experimental",
	"auto_update",
	"skills",
	"goal",
	"ralph_loop",
	"runtime_fallback",
	"background_task",
	"notification",
	"git_master",
	"new_task_system_enabled",
	"disabled_tools",
	"mcp_env_allowlist",
	"agent_definitions",
	"babysitting",
	"browser_automation_engine",
	"tmux",
	"websearch",
	"sisyphus",
	"default_run_agent",
	"start_work",
	"openclaw",
	"model_capabilities",
	"agent_order",
	"keyword_detector",
	"team_mode",
	"telemetry",
	"disabled_providers",
	"i18n",
	"default_mode",
	"monitor",
	"codegraph",
	"tui",
	"_migrations",
}

var knownFieldPaths = func() map[string]struct{} {
	paths := make(map[string]struct{}, len(allFieldPaths))
	for _, path := range allFieldPaths {
		paths[path] = struct{}{}
	}
	return paths
}()

var knownFieldPathPrefixes = func() map[string]struct{} {
	prefixes := make(map[string]struct{}, len(allFieldPaths)*2)
	for _, path := range allFieldPaths {
		parts := strings.Split(path, ".")
		for i := 1; i < len(parts); i++ {
			prefixes[strings.Join(parts[:i], ".")] = struct{}{}
		}
	}
	return prefixes
}()

func knownTags() map[string]struct{} {
	tags := make(map[string]struct{}, len(knownConfigTags))
	for _, tag := range knownConfigTags {
		tags[tag] = struct{}{}
	}
	return tags
}

func collectFieldPresence(raw map[string]json.RawMessage) map[string]bool {
	presence := make(map[string]bool)
	tags := knownTags()
	for key, value := range raw {
		if _, ok := tags[key]; !ok {
			continue
		}
		collectFieldPresenceFromRaw(canonicalPathSegment(key), value, presence)
	}
	return presence
}

func collectFieldPresenceFromRaw(path string, raw json.RawMessage, presence map[string]bool) {
	for _, candidate := range selectionPathCandidates(path) {
		if _, ok := knownFieldPaths[candidate]; ok {
			presence[candidate] = true
			return
		}
	}

	if !hasKnownFieldPathPrefix(path) {
		return
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return
	}

	for key, value := range object {
		collectFieldPresenceFromRaw(joinSelectionPath(path, canonicalPathSegment(key)), value, presence)
	}
}

func hasKnownFieldPathPrefix(path string) bool {
	for _, candidate := range selectionPathPrefixCandidates(path) {
		if _, ok := knownFieldPathPrefixes[candidate]; ok {
			return true
		}
	}
	return false
}

func selectionPathPrefixCandidates(path string) []string {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	wildcardableCount := len(parts) - 1
	total := 1 << wildcardableCount
	seen := make(map[string]struct{}, total)
	candidates := make([]string, 0, total)
	for mask := range total {
		candidateParts := append([]string(nil), parts...)
		for i := range wildcardableCount {
			if mask&(1<<i) == 0 {
				continue
			}
			candidateParts[i+1] = "*"
		}

		candidate := strings.Join(candidateParts, ".")
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	return candidates
}

func selectionPathCandidates(path string) []string {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	middleCount := max(0, len(parts)-2)

	total := 1 << middleCount
	seen := make(map[string]struct{}, total)
	candidates := make([]string, 0, total)
	for mask := range total {
		candidateParts := append([]string(nil), parts...)
		for i := range middleCount {
			if mask&(1<<i) == 0 {
				continue
			}
			candidateParts[i+1] = "*"
		}

		candidate := strings.Join(candidateParts, ".")
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	return candidates
}

// detectLegacyFields checks if the JSON data contains unknown fields
// that are not part of the config.Config struct. Returns true and an
// error message if unknown fields are detected.
func detectLegacyFields(data []byte) (bool, string) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var cfg config.Config
	if err := dec.Decode(&cfg); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "unknown field") {
			return true, errStr
		}
		return false, ""
	}
	return false, ""
}

// Load reads `profiles.<name>.[opencode]` from the user-layer omo document.
func Load(name string) (*Profile, error) {
	doc, err := config.LoadDocument()
	if err != nil {
		return nil, err
	}
	return LoadFromDocument(doc, name)
}

// LoadFromDocument extracts a profile from an already-parsed document.
func LoadFromDocument(doc *config.Document, name string) (*Profile, error) {
	block, ok, err := doc.ProfileBlock(name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &NotFoundError{Name: name}
	}

	blockFields := map[string]json.RawMessage{}
	if err := json.Unmarshal(block, &blockFields); err != nil {
		return nil, fmt.Errorf("parse profile %q: %w", name, err)
	}

	preservedBlock := make(map[string]json.RawMessage, len(blockFields))
	for key, value := range blockFields {
		if key == config.OpenCodeKey {
			continue
		}
		preservedBlock[key] = value
	}

	openCode := blockFields[config.OpenCodeKey]
	if len(bytes.TrimSpace(openCode)) == 0 {
		openCode = json.RawMessage("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(openCode, &raw); err != nil {
		return nil, fmt.Errorf("parse profile %q %s block: %w", name, config.OpenCodeKey, err)
	}

	preservedUnknown := make(map[string]json.RawMessage)
	tags := knownTags()
	for key, value := range raw {
		if _, ok := tags[key]; ok {
			continue
		}
		preservedUnknown[key] = value
	}
	fieldPresence := collectFieldPresence(raw)

	var cfg config.Config
	if err := json.Unmarshal(openCode, &cfg); err != nil {
		return nil, err
	}

	hasLegacy, warning := detectLegacyFields(openCode)

	return &Profile{
		Name:                name,
		Config:              cfg,
		Path:                doc.Path,
		PreservedUnknown:    preservedUnknown,
		PreservedBlock:      preservedBlock,
		FieldPresence:       fieldPresence,
		HasLegacyFields:     hasLegacy,
		LegacyFieldsWarning: warning,
	}, nil
}

func Save(p *Profile) error {
	return p.Save()
}

// Save writes the profile back into `profiles.<name>` of the omo document,
// leaving every other profile, harness block and shared key untouched.
func (p *Profile) Save() error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if err := p.WriteInto(doc); err != nil {
			return err
		}
		doc.EnsureSchema()
		p.Path = doc.Path
		return nil
	})
}

// WriteInto stages the profile into a document without persisting it.
func (p *Profile) WriteInto(doc *config.Document) error {
	data, err := json.Marshal(p.Config)
	if err != nil {
		return err
	}

	merged := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &merged); err != nil {
		return err
	}

	for key, value := range p.PreservedUnknown {
		if _, exists := merged[key]; exists {
			continue
		}
		merged[key] = value
	}

	// $schema belongs to the document root, never to a harness block.
	delete(merged, config.SchemaKey)

	openCode, err := marshalSortedJSONObject(merged)
	if err != nil {
		return err
	}

	block := make(map[string]json.RawMessage, len(p.PreservedBlock)+1)
	for key, value := range p.PreservedBlock {
		block[key] = value
	}
	block[config.OpenCodeKey] = openCode

	encoded, err := marshalSortedJSONObject(block)
	if err != nil {
		return err
	}
	return doc.SetProfileBlock(p.Name, encoded)
}

// WriteOpenCodeBlockInto stages a pre-marshalled `[opencode]` payload into
// `profiles.<name>`, preserving that profile's sibling keys.
//
// Use this when the caller produces the payload itself — the wizard's sparse
// save, for instance, where marshalling the Config struct would drop fields the
// user explicitly set to a zero value.
func WriteOpenCodeBlockInto(doc *config.Document, name string, openCode json.RawMessage) error {
	block := map[string]json.RawMessage{}

	existing, ok, err := doc.ProfileBlock(name)
	if err != nil {
		return err
	}
	if ok {
		if err := json.Unmarshal(existing, &block); err != nil {
			return fmt.Errorf("parse profile %q: %w", name, err)
		}
	}

	if len(bytes.TrimSpace(openCode)) == 0 {
		openCode = json.RawMessage("{}")
	}
	block[config.OpenCodeKey] = openCode

	encoded, err := marshalSortedJSONObject(block)
	if err != nil {
		return err
	}
	return doc.SetProfileBlock(name, encoded)
}

// UpdateOpenCodeBlock replaces an existing profile's `[opencode]` payload,
// failing with *NotFoundError when the profile is gone. Unlike
// SaveOpenCodeBlock it never creates: a stale editor tab saving after a
// concurrent delete must report the deletion, not undo it.
//
// The existence check and the write share one transaction.
func UpdateOpenCodeBlock(name string, openCode json.RawMessage) error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if !doc.HasProfile(name) {
			return &NotFoundError{Name: name}
		}
		if err := WriteOpenCodeBlockInto(doc, name, openCode); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
}

// SaveOpenCodeBlock persists a pre-marshalled `[opencode]` payload for a
// profile, leaving every other profile and top-level key untouched.
func SaveOpenCodeBlock(name string, openCode json.RawMessage) error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if err := WriteOpenCodeBlockInto(doc, name, openCode); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
}

func marshalSortedJSONObject(values map[string]json.RawMessage) ([]byte, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, key := range keys {
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}

		buf.WriteString("  ")
		buf.Write(encodedKey)
		buf.WriteString(": ")
		buf.Write(values[key])
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}")

	return buf.Bytes(), nil
}

// Delete removes `profiles.<name>` from the omo document.
func Delete(name string) error {
	err := config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		removed, err := doc.DeleteProfileBlock(name)
		if err != nil {
			return err
		}
		if !removed {
			return &NotFoundError{Name: name}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ExportOpenCode returns the stored `profiles.<name>.[opencode]` payload,
// pretty-printed. It reads the raw block instead of re-marshalling a typed
// Config, so an export reproduces what is on disk: unknown keys and explicitly
// present zero values ("disabled_mcps": [], "default_run_agent": "") survive
// and an export/import round-trip is lossless.
func ExportOpenCode(name string) ([]byte, error) {
	doc, err := config.LoadDocument()
	if err != nil {
		return nil, err
	}
	block, ok, err := doc.ProfileBlock(name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, &NotFoundError{Name: name}
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(block, &fields); err != nil {
		return nil, fmt.Errorf("parse profile %q: %w", name, err)
	}
	raw, ok := fields[config.OpenCodeKey]
	if !ok || len(bytes.TrimSpace(raw)) == 0 {
		return []byte("{}"), nil
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return nil, fmt.Errorf("format profile %q: %w", name, err)
	}
	return pretty.Bytes(), nil
}

// Create writes a new profile, failing with *ExistsError when the name is
// taken. The check and the write share one transaction, so two concurrent
// creates cannot both succeed and clobber each other.
func Create(name string, cfg config.Config) error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if doc.HasProfile(name) {
			return &ExistsError{Name: name}
		}
		p := &Profile{Name: name, Config: cfg}
		if err := p.WriteInto(doc); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
}

// CreateWithOpenCodeBlock writes a new profile from a pre-marshalled
// `[opencode]` payload. Use this for seeds that must keep explicit zeros
// (`[]`, `false`, `{}`) — Create re-marshals through typed Config and
// omitempty would drop them.
func CreateWithOpenCodeBlock(name string, openCode json.RawMessage) error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if doc.HasProfile(name) {
			return &ExistsError{Name: name}
		}
		if err := WriteOpenCodeBlockInto(doc, name, openCode); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
}

// CreateAvailable writes openCode under base, or under the first free `base-N`
// when base is taken. It reports the name actually used and whether a collision
// was resolved.
//
// Picking the name and claiming it happen in one transaction: a caller that
// looped on Exists() first could have the winning name taken from under it by a
// concurrent import and then overwrite that profile.
//
// The block is written verbatim rather than through the typed Config, so an
// import reproduces its source file: explicitly present zero values
// ("disabled_mcps": [], "default_run_agent": "") survive instead of being
// dropped by omitempty.
func CreateAvailable(base string, openCode json.RawMessage) (string, bool, error) {
	var name string
	collided := false
	err := config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		name, collided = base, false
		for i := 1; doc.HasProfile(name); i++ {
			name = fmt.Sprintf("%s-%d", base, i)
			collided = true
		}
		if err := WriteOpenCodeBlockInto(doc, name, openCode); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
	if err != nil {
		return "", false, err
	}
	return name, collided, nil
}

// CreateFrom clones fromName into name in one transaction, carrying the whole
// profile block. Reading the source and writing the clone under a single lock
// means the source cannot be renamed or deleted in between.
func CreateFrom(name, fromName string) error {
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		if doc.HasProfile(name) {
			return &ExistsError{Name: name}
		}
		// Copy the stored block verbatim. Round-tripping through the typed
		// Config would re-marshal with omitempty and silently drop explicitly
		// present zero values ("disabled_mcps": [], "default_run_agent": "").
		block, ok, err := doc.ProfileBlock(fromName)
		if err != nil {
			return err
		}
		if !ok {
			return &NotFoundError{Name: fromName}
		}
		if err := doc.SetProfileBlock(name, block); err != nil {
			return err
		}
		doc.EnsureSchema()
		return nil
	})
}

// ExistsError reports that a profile name is already taken.
type ExistsError struct{ Name string }

func (e *ExistsError) Error() string { return fmt.Sprintf("profile %q already exists", e.Name) }

// Rename moves `profiles.<oldName>` to `profiles.<newName>` in a single
// document write.
//
// The block moves verbatim, so `[opencode]`, its unknown keys, and the sibling
// blocks ([senpi], [codex], shared typed keys) all survive untouched. Doing the
// add and the remove in one Document.Save — which itself writes atomically via
// temp file + rename — means a failure leaves the document exactly as it was,
// instead of stranding both names and making the retry collide with its own
// half-finished result.
// Renaming the applied profile needs no follow-up: the block content is
// unchanged, so comparison-based detection follows the new name automatically.
func Rename(oldName, newName string) error {
	if oldName == newName {
		return nil
	}
	return config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		block, ok, err := doc.ProfileBlock(oldName)
		if err != nil {
			return err
		}
		if !ok {
			return &NotFoundError{Name: oldName}
		}
		if doc.HasProfile(newName) {
			return &ExistsError{Name: newName}
		}
		if err := doc.SetProfileBlock(newName, block); err != nil {
			return err
		}
		_, err = doc.DeleteProfileBlock(oldName)
		return err
	})
}

// List returns the profile names declared in the omo document.
func List() ([]string, error) {
	doc, err := config.LoadDocument()
	if err != nil {
		return nil, err
	}
	names, err := doc.ProfileNames()
	if err != nil {
		return nil, err
	}
	if names == nil {
		return []string{}, nil
	}
	return names, nil
}

// Exists reports whether `profiles.<name>` is present.
func Exists(name string) bool {
	doc, err := config.LoadDocument()
	if err != nil {
		return false
	}
	return doc.HasProfile(name)
}
