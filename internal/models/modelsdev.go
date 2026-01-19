package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ModelsDevLimit represents context and output token limits
type ModelsDevLimit struct {
	Context int `json:"context"`
	Output  int `json:"output"`
}

// ModelsDevModel represents a single model from models.dev API
type ModelsDevModel struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Family     string         `json:"family"`
	Reasoning  bool           `json:"reasoning"`
	ToolCall   bool           `json:"tool_call"`
	Attachment bool           `json:"attachment"`
	Limit      ModelsDevLimit `json:"limit"`
}

// ModelsDevProvider represents a provider with its models
type ModelsDevProvider struct {
	ID     string                    `json:"id"`
	Name   string                    `json:"name"`
	Models map[string]ModelsDevModel `json:"models"`
}

// ModelsDevResponse is the top-level API response
type ModelsDevResponse map[string]ModelsDevProvider

// FetchModelsDevRegistry fetches the models registry from models.dev API
func FetchModelsDevRegistry() (*ModelsDevResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get("https://models.dev/api.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models.dev API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models.dev API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result ModelsDevResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &result, nil
}

// ProviderWithCount is a provider with model count for display
type ProviderWithCount struct {
	ID         string
	Name       string
	ModelCount int
}

// ListProviders returns a sorted list of providers with model counts
func (r *ModelsDevResponse) ListProviders() []ProviderWithCount {
	if r == nil {
		return []ProviderWithCount{}
	}

	providers := make([]ProviderWithCount, 0, len(*r))
	for _, provider := range *r {
		providers = append(providers, ProviderWithCount{
			ID:         provider.ID,
			Name:       provider.Name,
			ModelCount: len(provider.Models),
		})
	}

	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name < providers[j].Name
	})

	return providers
}

// GetProviderModels returns models for a given provider ID, sorted by name
func (r *ModelsDevResponse) GetProviderModels(providerID string) []ModelsDevModel {
	if r == nil {
		return []ModelsDevModel{}
	}

	provider, ok := (*r)[providerID]
	if !ok {
		return []ModelsDevModel{}
	}

	models := make([]ModelsDevModel, 0, len(provider.Models))
	for _, model := range provider.Models {
		models = append(models, model)
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models
}

// ToRegisteredModel converts a ModelsDevModel to a RegisteredModel
func (m ModelsDevModel) ToRegisteredModel(provider string) RegisteredModel {
	return RegisteredModel{
		DisplayName: m.Name,
		ModelID:     m.ID,
		Provider:    provider,
	}
}

// FormatCapabilities returns a formatted string of model capabilities
// Example: "(200k ctx, reasoning, tools)"
func (m ModelsDevModel) FormatCapabilities() string {
	var parts []string

	if m.Limit.Context > 0 {
		ctx := formatTokenCount(m.Limit.Context)
		parts = append(parts, ctx+" ctx")
	}

	if m.Reasoning {
		parts = append(parts, "reasoning")
	}
	if m.ToolCall {
		parts = append(parts, "tools")
	}
	if m.Attachment {
		parts = append(parts, "vision")
	}

	if len(parts) == 0 {
		return ""
	}

	return "(" + strings.Join(parts, ", ") + ")"
}

// formatTokenCount formats large numbers into readable format (e.g., 200000 -> "200k")
func formatTokenCount(count int) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.0fm", float64(count)/1000000)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.0fk", float64(count)/1000)
	}
	return fmt.Sprintf("%d", count)
}
