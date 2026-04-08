package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
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
)

// ValidationError represents a single validation error
type ValidationError struct {
	Path    string // JSON path to the error
	Message string // Error message
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// Validator validates configs against the embedded schema
type Validator struct {
	schema *gojsonschema.Schema
}

// GetEmbeddedSchema returns the raw embedded JSON schema bytes.
func GetEmbeddedSchema() []byte {
	return schemaJSON
}

// GetValidator returns the singleton validator instance.
// The schema is parsed only once on first call.
func GetValidator() (*Validator, error) {
	validatorOnce.Do(func() {
		loader := gojsonschema.NewBytesLoader(schemaJSON)
		schema, err := gojsonschema.NewSchema(loader)
		if err != nil {
			validatorErr = err
			return
		}
		validatorInstance = &Validator{schema: schema}
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

// ValidateJSON validates raw JSON bytes against the schema
func (v *Validator) ValidateJSON(data []byte) ([]ValidationError, error) {
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

// ValidateJSONForSave validates raw JSON bytes for the save path while
// ignoring missing required-field errors.
func (v *Validator) ValidateJSONForSave(data []byte) ([]ValidationError, error) {
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
		if isRequiredError(e) {
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
