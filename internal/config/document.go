package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Document keys defined by the upstream omo schema.
const (
	SchemaKey   = "$schema"
	ProfilesKey = "profiles"
	// OpenCodeKey is the OpenCode harness block. Its contents are exactly the
	// pre-unification flat config, i.e. the Config struct in this package.
	OpenCodeKey = "[opencode]"
	SenpiKey    = "[senpi]"
	CodexKey    = "[codex]"
)

// Document is a parsed omo.json/omo.jsonc file.
//
// It keeps every top-level key as raw JSON so that writing back preserves
// harness blocks, shared typed keys and future schema additions that
// omo-profiler does not model. Only the parts we edit are decoded.
//
// Comments in a .jsonc source are NOT preserved across a write; the file is
// re-serialized as canonical JSON.
type Document struct {
	// Path is the file this document was read from (or would be written to).
	Path string
	// Exists reports whether the file was present on disk.
	Exists bool

	raw map[string]json.RawMessage
}

// NewDocument returns an empty document targeting the user-layer omo file.
func NewDocument() *Document {
	return &Document{
		Path:   OmoFile(),
		Exists: false,
		raw:    map[string]json.RawMessage{},
	}
}

// LoadDocument reads the user-layer omo config file. A missing file yields an
// empty document with Exists=false rather than an error, so callers can treat
// "no config yet" as a normal state.
func LoadDocument() (*Document, error) {
	return LoadDocumentFrom(OmoFile())
}

// docMutex serializes read-modify-write cycles on the user-layer document.
//
// Every profile now lives in one file, so two concurrent mutations that each
// load, edit and save would silently drop one of the two changes — the writes
// do not conflict at the byte level, they overwrite whole documents. The web
// server handles requests on separate goroutines, which is where this actually
// bites.
//
// Scope: this guards one process. Two omo-profiler processes writing at the
// same instant can still lose an update; the pre-write backup is the recovery
// path for that much rarer case. That path only works because backup names are
// claimed with O_EXCL (see backup.Create) — otherwise the two processes could
// pick the same name and the surviving snapshot would not be the lost write's
// pre-image.
var docMutex sync.Mutex

// Mutate runs fn against the user-layer document as a serialized transaction:
// load, apply fn, save — with no other Mutate interleaving.
//
// fn must not call Mutate (the lock is not reentrant) and must not call
// Document.Save itself; returning nil performs the save. Returning an error
// aborts before any write, leaving the file exactly as it was.
func Mutate(fn func(*Document) error) error {
	return MutateWithPreSave(nil, fn)
}

// MutateWithPreSave is Mutate with a hook that runs inside the lock, after fn
// succeeds and immediately before the write.
//
// That placement is the whole point for backups: a snapshot taken by the caller
// before Mutate races other writers. Two requests would both capture state S,
// then serialize as S→A→B, and no snapshot of A would exist to undo B. Running
// it here makes every snapshot the exact pre-image of the write that follows.
//
// preSave must not call Mutate. A failing preSave aborts before any write.
func MutateWithPreSave(preSave func() error, fn func(*Document) error) error {
	docMutex.Lock()
	defer docMutex.Unlock()

	doc, err := LoadDocument()
	if err != nil {
		return err
	}
	if err := fn(doc); err != nil {
		return err
	}
	if preSave != nil {
		if err := preSave(); err != nil {
			return err
		}
	}
	return doc.Save()
}

// LoadDocumentFrom reads an omo config document from an explicit path.
func LoadDocumentFrom(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Document{Path: path, Exists: false, raw: map[string]json.RawMessage{}}, nil
		}
		return nil, err
	}

	doc := &Document{Path: path, Exists: true, raw: map[string]json.RawMessage{}}
	if len(bytes.TrimSpace(data)) == 0 {
		return doc, nil
	}
	if err := json.Unmarshal(StripJSONC(data), &doc.raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return doc, nil
}

// ParseDocument builds a document from raw bytes, without touching disk.
func ParseDocument(data []byte) (*Document, error) {
	doc := &Document{raw: map[string]json.RawMessage{}}
	if len(bytes.TrimSpace(data)) == 0 {
		return doc, nil
	}
	if err := json.Unmarshal(StripJSONC(data), &doc.raw); err != nil {
		return nil, err
	}
	return doc, nil
}

// Raw returns the raw value of a top-level key.
func (d *Document) Raw(key string) (json.RawMessage, bool) {
	v, ok := d.raw[key]
	return v, ok
}

// SetRaw sets a top-level key to a raw JSON value.
func (d *Document) SetRaw(key string, value json.RawMessage) {
	if d.raw == nil {
		d.raw = map[string]json.RawMessage{}
	}
	d.raw[key] = value
}

// EnsureSchema sets $schema to the canonical omo schema URL when absent.
func (d *Document) EnsureSchema() {
	if _, ok := d.raw[SchemaKey]; ok {
		return
	}
	encoded, err := json.Marshal(DefaultSchema)
	if err != nil {
		return
	}
	d.SetRaw(SchemaKey, encoded)
}

// profiles decodes the profiles map, returning an empty map when absent.
func (d *Document) profiles() (map[string]json.RawMessage, error) {
	rawProfiles, ok := d.raw[ProfilesKey]
	if !ok {
		return map[string]json.RawMessage{}, nil
	}
	profiles := map[string]json.RawMessage{}
	if err := json.Unmarshal(rawProfiles, &profiles); err != nil {
		return nil, fmt.Errorf("parse %q: %w", ProfilesKey, err)
	}
	return profiles, nil
}

func (d *Document) setProfiles(profiles map[string]json.RawMessage) error {
	if len(profiles) == 0 {
		delete(d.raw, ProfilesKey)
		return nil
	}
	encoded, err := marshalSortedObject(profiles)
	if err != nil {
		return err
	}
	d.SetRaw(ProfilesKey, encoded)
	return nil
}

// ProfileNames returns the profile keys in sorted order.
func (d *Document) ProfileNames() ([]string, error) {
	profiles, err := d.profiles()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// ProfileBlock returns the raw `profiles.<name>` object.
func (d *Document) ProfileBlock(name string) (json.RawMessage, bool, error) {
	profiles, err := d.profiles()
	if err != nil {
		return nil, false, err
	}
	block, ok := profiles[name]
	return block, ok, nil
}

// SetProfileBlock inserts or replaces `profiles.<name>`.
func (d *Document) SetProfileBlock(name string, block json.RawMessage) error {
	profiles, err := d.profiles()
	if err != nil {
		return err
	}
	profiles[name] = block
	return d.setProfiles(profiles)
}

// DeleteProfileBlock removes `profiles.<name>`, reporting whether it existed.
func (d *Document) DeleteProfileBlock(name string) (bool, error) {
	profiles, err := d.profiles()
	if err != nil {
		return false, err
	}
	if _, ok := profiles[name]; !ok {
		return false, nil
	}
	delete(profiles, name)
	return true, d.setProfiles(profiles)
}

// HasProfile reports whether `profiles.<name>` exists.
func (d *Document) HasProfile(name string) bool {
	_, ok, err := d.ProfileBlock(name)
	return err == nil && ok
}

// Bytes serializes the document as canonical, indented JSON with sorted keys.
func (d *Document) Bytes() ([]byte, error) {
	compact, err := marshalSortedObject(d.raw)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, compact, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// Save writes the document back to its path, creating ~/.omo if needed.
//
// The write is atomic: the bytes go to a temp file in the same directory, are
// fsynced, and are then renamed over the target. This document holds every
// profile, so a plain truncating write that fails midway would take the whole
// set with it.
func (d *Document) Save() error {
	if d.Path == "" {
		d.Path = OmoFile()
	}
	if err := EnsureDirs(); err != nil {
		return err
	}
	data, err := d.Bytes()
	if err != nil {
		return err
	}
	if err := WriteFileAtomic(d.Path, data, 0600); err != nil {
		return err
	}
	d.Exists = true
	return nil
}

// WriteFileAtomic replaces path with data via a same-directory temp file and a
// rename, so a reader never observes a partially written document. The temp
// file must share the directory: os.Rename is only atomic within a filesystem.
//
// An existing file keeps its current permissions — this document can hold
// secrets (bot tokens), so a user who tightened it to 0600 must not have it
// silently widened. newPerm applies only when creating the file.
func WriteFileAtomic(path string, data []byte, newPerm os.FileMode) error {
	perm := newPerm
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		// Both are no-ops once the rename succeeded, and on failure paths the
		// error that matters is already being returned. Explicitly discarded.
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	// fsync before rename: without it a crash can leave the renamed file empty.
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// marshalSortedObject encodes a raw-value map as a compact JSON object with
// deterministically ordered keys.
func marshalSortedObject(values map[string]json.RawMessage) ([]byte, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(encodedKey)
		buf.WriteByte(':')
		value := values[key]
		if len(bytes.TrimSpace(value)) == 0 {
			buf.WriteString("null")
			continue
		}
		buf.Write(value)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
