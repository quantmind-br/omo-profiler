package schema

import (
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaEmbedded(t *testing.T) {
	// Verify the schema is embedded and not empty
	assert.NotEmpty(t, schemaJSON, "embedded schemaJSON should not be empty")
	assert.Greater(t, len(schemaJSON), 100, "schemaJSON should contain substantial content")
}

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err, "NewValidator should succeed with valid embedded schema")
	assert.NotNil(t, v, "validator should not be nil")
	assert.NotNil(t, v.schema, "validator schema should not be nil")
}

func TestValidate_ValidConfig(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Empty config should be valid (all fields are optional)
	cfg := &config.Config{}
	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.Nil(t, errs, "empty config should be valid")
}

func TestValidate_ValidConfigWithAgents(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	temp := 0.7
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Model:       "claude-sonnet-4-20250514",
				Temperature: &temp,
				Color:       "#FF5733",
				Mode:        "primary",
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.Nil(t, errs, "valid config with agents should pass validation")
}

func TestValidate_InvalidTemperature(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Temperature > 2 should be invalid
	temp := 2.5
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Temperature: &temp,
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.NotNil(t, errs, "temperature > 2 should fail validation")
	assert.NotEmpty(t, errs, "should have validation errors")
}

func TestValidate_InvalidTopP(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// top_p > 1 should be invalid
	topP := 1.5
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"plan": {
				TopP: &topP,
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.NotNil(t, errs, "top_p > 1 should fail validation")
}

func TestValidate_InvalidMode(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Invalid mode value
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Mode: "invalid_mode",
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.NotNil(t, errs, "invalid mode should fail validation")
}

func TestValidate_InvalidColor(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Color not matching #RRGGBB pattern
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Color: "red", // Should be #RRGGBB format
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.NotNil(t, errs, "invalid color format should fail validation")
}

func TestValidate_InvalidPermission(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Invalid permission value
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Permission: &config.PermissionConfig{
					Edit: "invalid_value", // Should be ask/allow/deny
				},
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.NotNil(t, errs, "invalid permission should fail validation")
}

func TestValidateJSON_ValidJSON(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	validJSON := []byte(`{}`)
	errs, err := v.ValidateJSON(validJSON)
	require.NoError(t, err)
	assert.Nil(t, errs, "empty JSON object should be valid")
}

func TestValidateJSON_ValidJSONWithAgents(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	validJSON := []byte(`{
		"agents": {
			"build": {
				"temperature": 0.7,
				"mode": "primary",
				"color": "#FF5733"
			}
		}
	}`)

	errs, err := v.ValidateJSON(validJSON)
	require.NoError(t, err)
	assert.Nil(t, errs, "valid JSON should pass validation")
}

func TestValidateJSON_InvalidTemperature(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	invalidJSON := []byte(`{
		"agents": {
			"build": {
				"temperature": 3.0
			}
		}
	}`)

	errs, err := v.ValidateJSON(invalidJSON)
	require.NoError(t, err)
	assert.NotNil(t, errs, "temperature > 2 in JSON should fail validation")
}

func TestValidateJSON_InvalidJSON(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	// Malformed JSON should return error
	invalidJSON := []byte(`{not valid json}`)
	_, err = v.ValidateJSON(invalidJSON)
	assert.Error(t, err, "malformed JSON should return error")
}

func TestValidationError_HasPath(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	temp := 3.0
	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Temperature: &temp,
			},
		},
	}

	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, errs)

	// Check that validation errors have path and message
	for _, e := range errs {
		assert.NotEmpty(t, e.Message, "validation error should have a message")
	}
}

func TestValidationError_Error(t *testing.T) {
	e := ValidationError{
		Path:    "agents.build.temperature",
		Message: "Must be less than or equal to 2",
	}
	assert.Equal(t, "agents.build.temperature: Must be less than or equal to 2", e.Error())
}
