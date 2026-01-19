package models

import (
	"encoding/json"
	"testing"
)

func TestModelsDevModel_ToRegisteredModel(t *testing.T) {
	model := ModelsDevModel{
		ID:   "claude-sonnet-4-0",
		Name: "Claude Sonnet 4",
	}

	result := model.ToRegisteredModel("anthropic")

	if result.ModelID != "claude-sonnet-4-0" {
		t.Errorf("Expected ModelID 'claude-sonnet-4-0', got '%s'", result.ModelID)
	}
	if result.DisplayName != "Claude Sonnet 4" {
		t.Errorf("Expected DisplayName 'Claude Sonnet 4', got '%s'", result.DisplayName)
	}
	if result.Provider != "anthropic" {
		t.Errorf("Expected Provider 'anthropic', got '%s'", result.Provider)
	}
}

func TestModelsDevModel_FormatCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		model    ModelsDevModel
		expected string
	}{
		{
			name: "full capabilities",
			model: ModelsDevModel{
				Limit:      ModelsDevLimit{Context: 200000},
				Reasoning:  true,
				ToolCall:   true,
				Attachment: true,
			},
			expected: "(200k ctx, reasoning, tools, vision)",
		},
		{
			name:     "no capabilities",
			model:    ModelsDevModel{},
			expected: "",
		},
		{
			name: "only context",
			model: ModelsDevModel{
				Limit: ModelsDevLimit{Context: 8192},
			},
			expected: "(8k ctx)",
		},
		{
			name: "million tokens",
			model: ModelsDevModel{
				Limit: ModelsDevLimit{Context: 2000000},
			},
			expected: "(2m ctx)",
		},
		{
			name: "reasoning only",
			model: ModelsDevModel{
				Reasoning: true,
			},
			expected: "(reasoning)",
		},
		{
			name: "tools only",
			model: ModelsDevModel{
				ToolCall: true,
			},
			expected: "(tools)",
		},
		{
			name: "vision only",
			model: ModelsDevModel{
				Attachment: true,
			},
			expected: "(vision)",
		},
		{
			name: "context and reasoning",
			model: ModelsDevModel{
				Limit:     ModelsDevLimit{Context: 128000},
				Reasoning: true,
			},
			expected: "(128k ctx, reasoning)",
		},
		{
			name: "context and tools",
			model: ModelsDevModel{
				Limit:    ModelsDevLimit{Context: 128000},
				ToolCall: true,
			},
			expected: "(128k ctx, tools)",
		},
		{
			name: "context and vision",
			model: ModelsDevModel{
				Limit:      ModelsDevLimit{Context: 128000},
				Attachment: true,
			},
			expected: "(128k ctx, vision)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.FormatCapabilities()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestModelsDevResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"anthropic": {
			"id": "anthropic",
			"name": "Anthropic",
			"models": {
				"claude-sonnet-4": {
					"id": "claude-sonnet-4",
					"name": "Claude Sonnet 4",
					"family": "claude-sonnet",
					"reasoning": true,
					"tool_call": true,
					"attachment": true,
					"limit": {"context": 200000, "output": 32000}
				}
			}
		}
	}`

	var resp ModelsDevResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	provider, ok := resp["anthropic"]
	if !ok {
		t.Fatal("Expected 'anthropic' provider")
	}

	if provider.Name != "Anthropic" {
		t.Errorf("Expected provider name 'Anthropic', got '%s'", provider.Name)
	}

	model, ok := provider.Models["claude-sonnet-4"]
	if !ok {
		t.Fatal("Expected 'claude-sonnet-4' model")
	}

	if !model.Reasoning {
		t.Error("Expected Reasoning to be true")
	}
	if model.Limit.Context != 200000 {
		t.Errorf("Expected context 200000, got %d", model.Limit.Context)
	}
}

func TestModelsDevResponse_Unmarshal_Defaults(t *testing.T) {
	// Test that missing fields get default values
	jsonData := `{
		"provider": {
			"id": "provider",
			"name": "Test Provider",
			"models": {
				"model1": {
					"id": "model1",
					"name": "Model 1"
				}
			}
		}
	}`

	var resp ModelsDevResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	model := resp["provider"].Models["model1"]
	if model.Reasoning != false {
		t.Error("Expected Reasoning default to be false")
	}
	if model.ToolCall != false {
		t.Error("Expected ToolCall default to be false")
	}
	if model.Limit.Context != 0 {
		t.Error("Expected Context default to be 0")
	}
}

func TestModelsDevResponse_ListProviders(t *testing.T) {
	resp := ModelsDevResponse{
		"openai": ModelsDevProvider{
			ID:   "openai",
			Name: "OpenAI",
			Models: map[string]ModelsDevModel{
				"gpt-4": {ID: "gpt-4"},
				"gpt-3": {ID: "gpt-3"},
			},
		},
		"anthropic": ModelsDevProvider{
			ID:   "anthropic",
			Name: "Anthropic",
			Models: map[string]ModelsDevModel{
				"claude": {ID: "claude"},
			},
		},
	}

	providers := resp.ListProviders()

	if len(providers) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(providers))
	}

	// Should be sorted by name (Anthropic before OpenAI)
	if providers[0].Name != "Anthropic" {
		t.Errorf("Expected first provider to be 'Anthropic', got '%s'", providers[0].Name)
	}
	if providers[0].ModelCount != 1 {
		t.Errorf("Expected Anthropic to have 1 model, got %d", providers[0].ModelCount)
	}
	if providers[1].Name != "OpenAI" {
		t.Errorf("Expected second provider to be 'OpenAI', got '%s'", providers[1].Name)
	}
	if providers[1].ModelCount != 2 {
		t.Errorf("Expected OpenAI to have 2 models, got %d", providers[1].ModelCount)
	}
}

func TestModelsDevResponse_ListProviders_Nil(t *testing.T) {
	var resp *ModelsDevResponse
	providers := resp.ListProviders()
	if len(providers) != 0 {
		t.Error("Expected empty slice for nil response")
	}
}

func TestModelsDevResponse_GetProviderModels(t *testing.T) {
	resp := ModelsDevResponse{
		"anthropic": ModelsDevProvider{
			ID:   "anthropic",
			Name: "Anthropic",
			Models: map[string]ModelsDevModel{
				"claude-sonnet": {ID: "claude-sonnet", Name: "Claude Sonnet"},
				"claude-opus":   {ID: "claude-opus", Name: "Claude Opus"},
			},
		},
	}

	models := resp.GetProviderModels("anthropic")

	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	// Should be sorted by name
	if models[0].Name != "Claude Opus" {
		t.Errorf("Expected first model 'Claude Opus', got '%s'", models[0].Name)
	}
}

func TestModelsDevResponse_GetProviderModels_NotFound(t *testing.T) {
	resp := ModelsDevResponse{}
	models := resp.GetProviderModels("nonexistent")
	if len(models) != 0 {
		t.Error("Expected empty slice for nonexistent provider")
	}
}
