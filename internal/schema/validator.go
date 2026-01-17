package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON []byte

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

// NewValidator creates a new validator with the embedded schema
func NewValidator() (*Validator, error) {
	loader := gojsonschema.NewBytesLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, err
	}
	return &Validator{schema: schema}, nil
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
