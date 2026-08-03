package profile

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/diogenes/omo-profiler/internal/backup"
	"github.com/diogenes/omo-profiler/internal/config"
)

// SnapshotBaseName is the profile name used when the root configuration has to
// be preserved before an apply overwrites it.
const SnapshotBaseName = "base"

// Applied reports the outcome of Apply.
type Applied struct {
	// Name is the profile that was applied.
	Name string
	// Snapshot names the profile that captured the previous root configuration,
	// empty when the root already matched a profile and nothing had to be saved.
	Snapshot string
}

// Apply makes a profile live by copying every key it declares over the matching
// document root key. Keys the profile does not declare are left untouched.
//
// Substitution is verbatim: the profile block is the configuration, not an
// override merged onto the root. When the current root matches no profile it is
// first saved as a new profile, so an apply never destroys a configuration that
// exists nowhere else.
func Apply(name string) (Applied, error) {
	var result Applied
	err := config.MutateWithPreSave(backup.CreateOmoIfPresent, func(doc *config.Document) error {
		block, ok, err := doc.ProfileBlock(name)
		if err != nil {
			return err
		}
		if !ok {
			return &NotFoundError{Name: name}
		}

		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(block, &fields); err != nil {
			return fmt.Errorf("parse profile %q: %w", name, err)
		}

		current, err := ActiveName(doc)
		if err != nil {
			return err
		}

		// When the root matches no profile, snapshot the root keys the profile
		// declares and that are present and non-empty, so they are never destroyed.
		if current == "" {
			snapshot := map[string]json.RawMessage{}
			for key := range fields {
				if root, ok := doc.Raw(key); ok && len(bytes.TrimSpace(root)) > 0 {
					snapshot[key] = root
				}
			}
			if len(snapshot) > 0 {
				snapshotName := SnapshotBaseName
				for i := 1; doc.HasProfile(snapshotName); i++ {
					snapshotName = fmt.Sprintf("%s-%d", SnapshotBaseName, i)
				}
				marshalled, err := json.Marshal(snapshot)
				if err != nil {
					return err
				}
				if err := doc.SetProfileBlock(snapshotName, marshalled); err != nil {
					return err
				}
				result.Snapshot = snapshotName
			}
		}

		for key, value := range fields {
			doc.SetRaw(key, value)
		}

		doc.EnsureSchema()
		return nil
	})
	if err != nil {
		return Applied{}, err
	}
	result.Name = name
	return result, nil
}

// ActiveName returns the profile whose every declared key equals the
// corresponding document root key, or "" when no profile matches. Names are
// checked in sorted order, so two identical profiles resolve deterministically
// to the first.
func ActiveName(doc *config.Document) (string, error) {
	names, err := doc.ProfileNames()
	if err != nil {
		return "", err
	}
	for _, name := range names {
		block, ok, err := doc.ProfileBlock(name)
		if err != nil {
			return "", err
		}
		if !ok {
			continue
		}
		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(block, &fields); err != nil {
			return "", fmt.Errorf("parse profile %q: %w", name, err)
		}
		if len(fields) == 0 {
			continue
		}
		match := true
		for key, value := range fields {
			root, ok := doc.Raw(key)
			if !ok {
				match = false
				break
			}
			canonValue, err := canonicalJSON(value)
			if err != nil {
				return "", fmt.Errorf("canonicalize profile %q key %q: %w", name, key, err)
			}
			canonRoot, err := canonicalJSON(root)
			if err != nil {
				return "", fmt.Errorf("canonicalize root key %q: %w", key, err)
			}
			if !bytes.Equal(canonValue, canonRoot) {
				match = false
				break
			}
		}
		if match {
			return name, nil
		}
	}
	return "", nil
}

// canonicalJSON renders raw in a form where equal values compare equal:
// encoding/json sorts map keys at every depth. $schema is dropped so a root
// block that carries one still matches a stored profile, which never does
// (WriteInto strips it).
func canonicalJSON(raw json.RawMessage) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return []byte("null"), nil
	}
	var v any
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return nil, err
	}
	if m, ok := v.(map[string]any); ok {
		delete(m, config.SchemaKey)
	}
	return json.Marshal(v)
}

// ActiveConfig describes the effective OpenCode configuration and whether a
// profile is currently applied.
type ActiveConfig struct {
	// Exists reports whether an omo document is present on disk.
	Exists bool
	// Config is the root `[opencode]` block — after substitution the root *is*
	// the effective configuration, so nothing is merged.
	Config config.Config
	// ProfileName is the profile matching the root, empty when none does.
	ProfileName string
	// Modified is true when a root `[opencode]` block exists but matches no
	// profile — a hand-edited or never-saved configuration.
	Modified bool
}

// GetActive resolves the effective configuration from the document.
func GetActive() (*ActiveConfig, error) {
	doc, err := config.LoadDocument()
	if err != nil {
		return nil, err
	}

	result := &ActiveConfig{
		Exists: doc.Exists,
	}

	profileName, err := ActiveName(doc)
	if err != nil {
		return nil, err
	}
	result.ProfileName = profileName

	if raw, ok := doc.Raw(config.OpenCodeKey); ok && len(bytes.TrimSpace(raw)) > 0 {
		var cfg config.Config
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse %s block: %w", config.OpenCodeKey, err)
		}
		result.Config = cfg
		result.Modified = profileName == ""
	}

	return result, nil
}
