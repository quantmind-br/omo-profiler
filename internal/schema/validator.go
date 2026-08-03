package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON []byte

var (
	validatorInstance *Validator
	validatorOnce     sync.Once
	validatorErr      error

	// openCodeSchemaJSON is the `[opencode]` sub-schema carved out of the omo
	// document schema. Populated by GetValidator.
	openCodeSchemaJSON []byte
)

// ValidationError represents a single validation error
type ValidationError struct {
	Path    string // JSON path to the error
	Message string // Error message
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// Validator validates omo documents and the OpenCode harness blocks inside them.
type Validator struct {
	// schema validates a flat `[opencode]` config, i.e. a config.Config.
	schema *gojsonschema.Schema
	// documentSchema validates a whole omo.json document.
	documentSchema *gojsonschema.Schema
}

// GetEmbeddedSchema returns the raw embedded omo document schema.
func GetEmbeddedSchema() []byte {
	return schemaJSON
}

// GetOpenCodeSchema returns the `[opencode]` sub-schema, which describes the
// flat configuration omo-profiler edits. It is self-contained (no $ref) and
// therefore usable standalone, e.g. to drive a schema-rendered form.
func GetOpenCodeSchema() ([]byte, error) {
	if _, err := GetValidator(); err != nil {
		return nil, err
	}
	return openCodeSchemaJSON, nil
}

// extractOpenCodeSchema pulls properties["[opencode]"] out of the document schema.
func extractOpenCodeSchema(document []byte) ([]byte, error) {
	var root struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(document, &root); err != nil {
		return nil, fmt.Errorf("parse embedded schema: %w", err)
	}
	sub, ok := root.Properties[config.OpenCodeKey]
	if !ok {
		return nil, fmt.Errorf("embedded schema has no %q property", config.OpenCodeKey)
	}
	return sub, nil
}

// GetValidator returns the singleton validator instance.
// The schema is parsed only once on first call.
func GetValidator() (*Validator, error) {
	validatorOnce.Do(func() {
		documentSchema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaJSON))
		if err != nil {
			validatorErr = err
			return
		}

		openCodeSchemaJSON, err = extractOpenCodeSchema(schemaJSON)
		if err != nil {
			validatorErr = err
			return
		}

		openCodeSchema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(openCodeSchemaJSON))
		if err != nil {
			validatorErr = err
			return
		}

		validatorInstance = &Validator{schema: openCodeSchema, documentSchema: documentSchema}
	})
	return validatorInstance, validatorErr
}

// NewValidator creates a new validator with the embedded schema.
// Deprecated: Use GetValidator() for singleton access.
func NewValidator() (*Validator, error) {
	return GetValidator()
}

// Validate validates a config against the schema
func (v *Validator) Validate(cfg *config.Config) ([]ValidationError, error) {
	// Marshal config to JSON for validation
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	loader := gojsonschema.NewBytesLoader(data)
	result, err := v.schema.Validate(loader)
	if err != nil {
		return nil, err
	}

	if result.Valid() {
		return nil, nil
	}

	var errors []ValidationError
	for _, e := range result.Errors() {
		errors = append(errors, ValidationError{
			Path:    e.Field(),
			Message: e.Description(),
		})
	}
	return errors, nil
}

// ValidateForSave validates config for the save path. Unlike Validate,
// this ignores schema "required" errors, allowing sparse configs where
// omitted fields rely on consumer defaults. Type/enum/shape violations
// for present fields are still reported.
func (v *Validator) ValidateForSave(cfg *config.Config) ([]ValidationError, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return v.ValidateJSONForSave(data)
}

// ValidateJSON validates raw `[opencode]` config bytes against the schema.
func (v *Validator) ValidateJSON(data []byte) ([]ValidationError, error) {
	return validateBytes(v.schema, data)
}

// ValidateJSONForSave validates raw `[opencode]` config bytes for the save
// path while ignoring missing required-field errors.
func (v *Validator) ValidateJSONForSave(data []byte) ([]ValidationError, error) {
	return validateBytesForSave(v.schema, data)
}

// ValidateDocument validates a whole omo.json document.
func (v *Validator) ValidateDocument(data []byte) ([]ValidationError, error) {
	return validateBytes(v.documentSchema, data)
}

// ValidateDocumentForSave validates a whole omo.json document for the save
// path, tolerating sparse layers that rely on consumer defaults.
func (v *Validator) ValidateDocumentForSave(data []byte) ([]ValidationError, error) {
	return validateBytesForSave(v.documentSchema, data)
}

func validateBytes(s *gojsonschema.Schema, data []byte) ([]ValidationError, error) {
	result, err := s.Validate(gojsonschema.NewBytesLoader(data))
	if err != nil {
		return nil, err
	}
	if result.Valid() {
		return nil, nil
	}

	var errors []ValidationError
	for _, e := range result.Errors() {
		errors = append(errors, ValidationError{
			Path:    e.Field(),
			Message: e.Description(),
		})
	}
	return errors, nil
}

func validateBytesForSave(s *gojsonschema.Schema, data []byte) ([]ValidationError, error) {
	var parsed any
	_ = json.Unmarshal(data, &parsed)

	result, err := s.Validate(gojsonschema.NewBytesLoader(data))
	if err != nil {
		return nil, err
	}
	if result.Valid() {
		return nil, nil
	}

	var errors []ValidationError
	for _, e := range result.Errors() {
		if isRequiredError(e) || isAdditionalPropertyError(e) || isMinimumErrorOnZeroValue(e, parsed) {
			continue
		}

		errors = append(errors, ValidationError{
			Path:    e.Field(),
			Message: e.Description(),
		})
	}

	if len(errors) == 0 {
		return nil, nil
	}

	return errors, nil
}

func isRequiredError(e gojsonschema.ResultError) bool {
	return e.Type() == "required"
}

func isAdditionalPropertyError(e gojsonschema.ResultError) bool {
	return strings.Contains(strings.ToLower(e.Description()), "additional property")
}

func isMinimumErrorOnZeroValue(e gojsonschema.ResultError, data any) bool {
	if data == nil {
		return false
	}
	if !strings.Contains(strings.ToLower(e.Description()), "greater than or equal to") {
		return false
	}

	value, ok := jsonValueAtPath(data, e.Field())
	if !ok {
		return false
	}

	switch v := value.(type) {
	case float64:
		return v == 0
	case int:
		return v == 0
	case int64:
		return v == 0
	case json.Number:
		return v.String() == "0" || v.String() == "0.0"
	default:
		return false
	}
}

func jsonValueAtPath(data any, field string) (any, bool) {
	if field == "(root)" || field == "" {
		return data, true
	}

	current := data
	for _, part := range strings.Split(field, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}
