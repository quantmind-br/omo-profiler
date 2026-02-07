package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmbeddedSchema_ReturnsNonEmpty(t *testing.T) {
	schema := GetEmbeddedSchema()
	assert.NotEmpty(t, schema, "embedded schema should not be empty")
	assert.Greater(t, len(schema), 100, "schema should contain substantial content")
}

func TestGetEmbeddedSchema_ReturnsValidJSON(t *testing.T) {
	schema := GetEmbeddedSchema()
	var js map[string]interface{}
	err := json.Unmarshal(schema, &js)
	assert.NoError(t, err, "embedded schema should be valid JSON")
}
