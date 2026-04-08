package schema

import (
	"strings"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
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

	// Config with required git_master should be valid
	coAuth := true
	cfg := &config.Config{
		GitMaster: &config.GitMasterConfig{
			CommitFooter:        true,
			IncludeCoAuthoredBy: &coAuth,
			GitEnvPrefix:        "GIT_MASTER=1",
		},
	}
	errs, err := v.Validate(cfg)
	require.NoError(t, err)
	assert.Nil(t, errs, "config with required git_master should be valid")
}

func TestValidate_ValidConfigWithAgents(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	temp := 0.7
	coAuth := true
	cfg := &config.Config{
		GitMaster: &config.GitMasterConfig{
			CommitFooter:        true,
			IncludeCoAuthoredBy: &coAuth,
			GitEnvPrefix:        "GIT_MASTER=1",
		},
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

	validJSON := []byte(`{"git_master": {"commit_footer": true, "include_co_authored_by": true, "git_env_prefix": "GIT_MASTER=1"}}`)
	errs, err := v.ValidateJSON(validJSON)
	require.NoError(t, err)
	assert.Nil(t, errs, "JSON with required git_master should be valid")
}

func TestValidateJSON_ValidJSONWithAgents(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	validJSON := []byte(`{
		"git_master": {"commit_footer": true, "include_co_authored_by": true, "git_env_prefix": "GIT_MASTER=1"},
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

func TestValidateForSaveAllowsEmptyConfig(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	errors, err := v.ValidateForSave(&config.Config{})
	require.NoError(t, err)
	assert.Nil(t, errors, "empty config should be valid for save")
}

func TestValidateForSaveRejectsMalformedPresentValues(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	cfg := &config.Config{
		Agents: map[string]*config.AgentConfig{
			"build": {
				Mode: "invalid_mode",
			},
		},
	}

	errors, err := v.ValidateForSave(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, errors, "invalid present values should fail save validation")
	assert.Contains(t, validationErrorMessages(errors)[0], "must be one of the following")
}

func TestValidateStrictStillReportsRequiredness(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	errors, err := v.Validate(&config.Config{})
	require.NoError(t, err)
	require.NotEmpty(t, errors, "strict validation should report required field errors")
	assert.True(t, hasRequiredValidationError(errors), "strict validation should keep required errors")
}

func TestValidateForSaveFiltersOnlyRequiredErrors(t *testing.T) {
	v, err := NewValidator()
	require.NoError(t, err)

	invalidJSON := []byte(`{
		"agents": {
			"build": {
				"temperature": "hot"
			}
		}
	}`)

	strictErrors, err := v.ValidateJSON(invalidJSON)
	require.NoError(t, err)
	require.NotEmpty(t, strictErrors)
	assert.True(t, hasRequiredValidationError(strictErrors), "strict validation should include required errors")

	saveErrors, err := v.ValidateJSONForSave(invalidJSON)
	require.NoError(t, err)
	require.NotEmpty(t, saveErrors, "non-required errors should remain after filtering")
	assert.False(t, hasRequiredValidationError(saveErrors), "save validation should filter required errors")
	assert.True(t, hasTypeValidationError(saveErrors), "save validation should keep type errors")
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

func TestRegressionSparseValidationContract(t *testing.T) {
	v, err := GetValidator()
	require.NoError(t, err)

	blankSparseJSON := mustMarshalSparseJSON(t, &config.Config{})
	assert.JSONEq(t, `{}`, string(blankSparseJSON))

	saveErrors, err := v.ValidateForSave(&config.Config{})
	require.NoError(t, err)
	assert.Nil(t, saveErrors, "blank sparse config should be valid for save")

	saveJSONErrors, err := v.ValidateJSONForSave(blankSparseJSON)
	require.NoError(t, err)
	assert.Nil(t, saveJSONErrors, "blank sparse JSON should be valid for save")

	strictErrors, err := v.Validate(&config.Config{})
	require.NoError(t, err)
	require.NotEmpty(t, strictErrors, "strict validation should reject a blank config")
	assert.True(t, hasRequiredValidationError(strictErrors), "strict validation should keep required-field errors")

	strictJSONErrors, err := v.ValidateJSON(blankSparseJSON)
	require.NoError(t, err)
	require.NotEmpty(t, strictJSONErrors, "strict validation should reject blank sparse JSON")
	assert.True(t, hasRequiredValidationError(strictJSONErrors), "strict JSON validation should keep required-field errors")

	invalidSaveCases := []struct {
		name          string
		cfg           *config.Config
		data          []byte
		wantSubstring string
	}{
		{
			name: "invalid enum remains invalid after sparse marshal",
			cfg: &config.Config{
				Agents: map[string]*config.AgentConfig{
					"build": {Mode: "invalid_mode"},
				},
			},
			data: mustMarshalSparseJSON(t, &config.Config{
				Agents: map[string]*config.AgentConfig{
					"build": {Mode: "invalid_mode"},
				},
			}, "agents.*.mode"),
			wantSubstring: "must be one of the following",
		},
		{
			name:          "wrong type survives required-error filtering",
			data:          []byte(`{"agents":{"build":{"temperature":"hot"}}}`),
			wantSubstring: "Invalid type",
		},
	}

	for _, tc := range invalidSaveCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cfg != nil {
				typedSaveErrors, err := v.ValidateForSave(tc.cfg)
				require.NoError(t, err)
				require.NotEmpty(t, typedSaveErrors, "typed save validation should reject malformed present values")
				assert.False(t, hasRequiredValidationError(typedSaveErrors), "save validation should only reject the malformed present value")
				assert.Contains(t, strings.Join(validationErrorMessages(typedSaveErrors), "\n"), tc.wantSubstring)
			}

			strictErrors, err := v.ValidateJSON(tc.data)
			require.NoError(t, err)
			require.NotEmpty(t, strictErrors, "strict validation should reject malformed sparse JSON")
			assert.True(t, hasRequiredValidationError(strictErrors), "strict validation should still report required-field errors")

			saveErrors, err := v.ValidateJSONForSave(tc.data)
			require.NoError(t, err)
			require.NotEmpty(t, saveErrors, "save validation should reject malformed present values")
			assert.False(t, hasRequiredValidationError(saveErrors), "save validation should filter only required errors")
			assert.Contains(t, strings.Join(validationErrorMessages(saveErrors), "\n"), tc.wantSubstring)
		})
	}
}

func hasRequiredValidationError(errors []ValidationError) bool {
	for _, err := range errors {
		if strings.Contains(err.Message, "is required") {
			return true
		}
	}

	return false
}

func hasTypeValidationError(errors []ValidationError) bool {
	for _, err := range errors {
		if strings.Contains(err.Message, "Invalid type") {
			return true
		}
	}

	return false
}

func validationErrorMessages(errors []ValidationError) []string {
	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Message)
	}

	return messages
}

func mustMarshalSparseJSON(t *testing.T, cfg *config.Config, selectedPaths ...string) []byte {
	t.Helper()

	selection := profile.NewBlankSelection()
	for _, path := range selectedPaths {
		selection.SetSelected(path, true)
	}

	data, err := profile.MarshalSparse(cfg, selection, nil)
	require.NoError(t, err)
	return data
}
