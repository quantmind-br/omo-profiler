package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/diff"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
)

// validationErr mirrors schema.ValidationError for JSON responses.
type validationErr struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func mapValidationErrors(errs []schema.ValidationError) []validationErr {
	out := make([]validationErr, 0, len(errs))
	for _, e := range errs {
		out = append(out, validationErr{Path: e.Path, Message: e.Message})
	}
	return out
}

// nameError maps profile name validation errors to a 400 response. Returns
// true when it handled (wrote) an error.
func nameError(w http.ResponseWriter, name string) bool {
	if err := profile.ValidateName(name); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return true
	}
	return false
}

// activeJSON describes the applied profile for API responses. The active
// profile is detected by comparing the root against stored profiles, so there
// is no separate "hint" or "source": the root *is* the effective configuration.
func activeJSON(active *profile.ActiveConfig) map[string]any {
	return map[string]any{
		"documentExists": active.Exists,
		"profileName":    active.ProfileName,
		"modified":       active.Modified,
	}
}

// GET /api/profiles
func handleListProfiles(w http.ResponseWriter, r *http.Request) {
	names, err := profile.List()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	active, err := profile.GetActive()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	type profileEntry struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	entries := make([]profileEntry, 0, len(names))
	for _, n := range names {
		entries = append(entries, profileEntry{
			Name:   n,
			Active: active.ProfileName == n,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"active":   activeJSON(active),
		"profiles": entries,
	})
}

// GET /api/profiles/{name}
func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}

	p, err := profile.Load(name)
	if err != nil {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("profile not found: %s", name))
		return
	}

	// Lossless `[opencode]` block (preserves unknown keys).
	raw, err := profile.ExportOpenCode(name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":                name,
		"config":              json.RawMessage(raw),
		"fieldPresence":       p.FieldPresence,
		"hasLegacyFields":     p.HasLegacyFields,
		"legacyFieldsWarning": p.LegacyFieldsWarning,
	})
}

// PUT /api/profiles/{name}
func handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}

	body, err := readBody(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	validator, err := schema.GetValidator()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	errs, err := validator.ValidateJSONForSave(body)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":            "validation failed",
			"validationErrors": mapValidationErrors(errs),
		})
		return
	}

	// Type-check the payload, then store it verbatim. Re-marshalling a typed
	// Config would drop explicitly present zero values and any key the editor
	// sends that this build's struct does not model.
	if err := json.Unmarshal(body, &config.Config{}); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	// One transaction: reading the existing block (for its sibling blocks) and
	// writing the new one must not be split by a concurrent change.
	if err := profile.UpdateOpenCodeBlock(name, body); err != nil {
		var notFound *profile.NotFoundError
		if errors.As(err, &notFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/profiles
func handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		From string `json:"from"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	if nameError(w, req.Name) {
		return
	}
	if profile.Exists(req.Name) {
		writeErr(w, http.StatusConflict, fmt.Sprintf("profile already exists: %s", req.Name))
		return
	}

	// Blank create uses an empty typed Config. The default template is written
	// as raw `[opencode]` bytes so explicit zeros (`[]`, `false`, `{}`) survive —
	// Create re-marshals through config.Config and omitempty would drop them.
	var err error
	switch {
	case req.From == "__default__":
		err = profile.CreateWithOpenCodeBlock(req.Name, DefaultTemplate())
	case req.From == "":
		err = profile.Create(req.Name, config.Config{})
	default:
		err = profile.CreateFrom(req.Name, req.From)
	}
	if err != nil {
		var notFound *profile.NotFoundError
		var exists *profile.ExistsError
		switch {
		case errors.As(err, &notFound):
			writeErr(w, http.StatusNotFound, fmt.Sprintf("source profile not found: %s", req.From))
		case errors.As(err, &exists):
			writeErr(w, http.StatusConflict, err.Error())
		default:
			writeErr(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"name": req.Name})
}

// DELETE /api/profiles/{name}
func handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}
	if !profile.Exists(name) {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("profile not found: %s", name))
		return
	}

	if err := profile.Delete(name); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/profiles/{name}/rename
func handleRenameProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if nameError(w, req.NewName) {
		return
	}
	if profile.Exists(req.NewName) {
		writeErr(w, http.StatusConflict, fmt.Sprintf("profile already exists: %s", req.NewName))
		return
	}

	if !profile.Exists(name) {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("profile not found: %s", name))
		return
	}

	// One document write: a failure here leaves the document untouched rather
	// than stranding both names. The block content is unchanged, so
	// comparison-based detection follows the new name automatically.
	if err := profile.Rename(name, req.NewName); err != nil {
		var notFound *profile.NotFoundError
		var exists *profile.ExistsError
		switch {
		case errors.As(err, &notFound):
			writeErr(w, http.StatusNotFound, err.Error())
		case errors.As(err, &exists):
			writeErr(w, http.StatusConflict, err.Error())
		default:
			writeErr(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": req.NewName})
}

// POST /api/profiles/{name}/activate
func handleActivateProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}
	if !profile.Exists(name) {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("profile not found: %s", name))
		return
	}

	applied, err := profile.Apply(name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"name":   applied.Name,
		"snapshot": applied.Snapshot,
	})
}

// GET /api/profiles/{name}/export
func handleExportProfile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if nameError(w, name) {
		return
	}

	data, err := profile.ExportOpenCode(name)
	if err != nil {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("profile not found: %s", name))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name+".json"))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GET /api/active
func handleGetActive(w http.ResponseWriter, r *http.Request) {
	active, err := profile.GetActive()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload := activeJSON(active)
	payload["config"] = active.Config
	writeJSON(w, http.StatusOK, payload)
}

// GET /api/diff?left=&right=
func handleDiff(w http.ResponseWriter, r *http.Request) {
	left := r.URL.Query().Get("left")
	right := r.URL.Query().Get("right")
	if left == "" || right == "" {
		writeErr(w, http.StatusBadRequest, "left and right query params are required")
		return
	}

	leftBytes, err := resolveDiffSide(left)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	rightBytes, err := resolveDiffSide(right)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}

	res, err := diff.ComputeDiff(leftBytes, rightBytes)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"leftLabel":  left,
		"rightLabel": right,
		"left":       marshalDiffLines(res.Left),
		"right":      marshalDiffLines(res.Right),
	})
}

type diffLineJSON struct {
	Text    string `json:"text"`
	Type    int    `json:"type"`
	LineNum int    `json:"lineNum"`
}

func marshalDiffLines(lines []diff.DiffLine) []diffLineJSON {
	out := make([]diffLineJSON, 0, len(lines))
	for _, l := range lines {
		out = append(out, diffLineJSON{Text: l.Text, Type: int(l.Type), LineNum: l.LineNum})
	}
	return out
}

// resolveDiffSide returns the JSON bytes for a diff side. "__active__" resolves
// to the effective OpenCode config; any other value is a profile's `[opencode]` block.
func resolveDiffSide(label string) ([]byte, error) {
	if label == "__active__" {
		active, err := profile.GetActive()
		if err != nil {
			return nil, err
		}
		if !active.Exists {
			return nil, fmt.Errorf("no omo config found at %s", config.OmoFile())
		}
		return json.Marshal(active.Config)
	}

	data, err := profile.ExportOpenCode(label)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", label)
	}
	return data, nil
}

// POST /api/import
func handleImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string          `json:"name"`
		Config json.RawMessage `json:"config"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Config) == 0 {
		writeErr(w, http.StatusBadRequest, "config is required")
		return
	}

	validator, err := schema.GetValidator()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	errs, err := validator.ValidateJSONForSave(req.Config)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":            "validation failed",
			"validationErrors": mapValidationErrors(errs),
		})
		return
	}

	// Type-check the payload before storing it verbatim; schema validation
	// alone would let a well-shaped but wrongly typed value through.
	if err := json.Unmarshal(req.Config, &config.Config{}); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	base := profile.SanitizeName(req.Name)
	if base == "" {
		base = "imported"
	}

	// The suffix is chosen inside the transaction that claims it, so two
	// concurrent imports of the same name cannot settle on it and overwrite.
	finalName, hadCollision, err := profile.CreateAvailable(base, req.Config)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"name": finalName, "hadCollision": hadCollision})
}

// POST /api/validate?mode=strict|save
func handleValidate(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	validator, err := schema.GetValidator()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	var errs []schema.ValidationError
	if r.URL.Query().Get("mode") == "strict" {
		errs, err = validator.ValidateJSON(body)
	} else {
		errs, err = validator.ValidateJSONForSave(body)
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"valid":  len(errs) == 0,
		"errors": mapValidationErrors(errs),
	})
}

// GET /api/schema — the flat `[opencode]` schema that drives the editor form.
func handleSchema(w http.ResponseWriter, r *http.Request) {
	data, err := schema.GetOpenCodeSchema()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GET /api/document-schema — the full omo.json document schema.
func handleDocumentSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(schema.GetEmbeddedSchema())
}

// GET /api/schema-check
func handleSchemaCheck(w http.ResponseWriter, r *http.Request) {
	res, err := schema.CompareSchemas()
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"identical": res.Identical,
		"diff":      res.Diff,
	})
}

// readBody reads and returns the raw request body.
func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, errors.New("empty request body")
	}
	defer func() { _ = r.Body.Close() }()
	return io.ReadAll(r.Body)
}

// decodeBody decodes the JSON request body into v.
func decodeBody(r *http.Request, v any) error {
	body, err := readBody(r)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty request body")
	}
	return json.Unmarshal(body, v)
}
