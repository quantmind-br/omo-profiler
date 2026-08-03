package web

import (
	"errors"
	"net/http"

	"github.com/diogenes/omo-profiler/internal/models"
)

// GET /api/models
func handleListModels(w http.ResponseWriter, r *http.Request) {
	reg, err := models.Load()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	type modelJSON struct {
		DisplayName string `json:"displayName"`
		ModelID     string `json:"modelId"`
		Provider    string `json:"provider"`
	}
	type groupJSON struct {
		Provider string      `json:"provider"`
		Models   []modelJSON `json:"models"`
	}

	groups := reg.ListByProvider()
	out := make([]groupJSON, 0, len(groups))
	for _, g := range groups {
		ms := make([]modelJSON, 0, len(g.Models))
		for _, m := range g.Models {
			ms = append(ms, modelJSON{DisplayName: m.DisplayName, ModelID: m.ModelID, Provider: m.Provider})
		}
		out = append(out, groupJSON{Provider: g.Provider, Models: ms})
	}

	writeJSON(w, http.StatusOK, map[string]any{"groups": out})
}

// POST /api/models
func handleCreateModel(w http.ResponseWriter, r *http.Request) {
	var m models.RegisteredModel
	if err := decodeBody(r, &m); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if m.ModelID == "" {
		writeErr(w, http.StatusBadRequest, "modelId is required")
		return
	}

	if err := models.Add(m); err != nil {
		writeModelErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// PUT /api/models/{provider}/{modelId}
func handleUpdateModel(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	modelID := r.PathValue("modelId")

	var m models.RegisteredModel
	if err := decodeBody(r, &m); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if m.ModelID == "" {
		writeErr(w, http.StatusBadRequest, "modelId is required")
		return
	}

	if err := models.Update(provider, modelID, m); err != nil {
		writeModelErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// DELETE /api/models/{provider}/{modelId}
func handleDeleteModel(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	modelID := r.PathValue("modelId")

	if err := models.Delete(provider, modelID); err != nil {
		writeModelErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/models/catalog
func handleModelsCatalog(w http.ResponseWriter, r *http.Request) {
	resp, err := models.FetchModelsDevRegistry()
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	type catModel struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Family       string `json:"family"`
		Reasoning    bool   `json:"reasoning"`
		ToolCall     bool   `json:"toolCall"`
		Attachment   bool   `json:"attachment"`
		Context      int    `json:"context"`
		Output       int    `json:"output"`
		Capabilities string `json:"capabilities"`
	}
	type catProvider struct {
		ID     string     `json:"id"`
		Name   string     `json:"name"`
		Models []catModel `json:"models"`
	}

	providers := resp.ListProviders()
	out := make([]catProvider, 0, len(providers))
	for _, p := range providers {
		ms := resp.GetProviderModels(p.ID)
		cms := make([]catModel, 0, len(ms))
		for _, m := range ms {
			cms = append(cms, catModel{
				ID:           m.ID,
				Name:         m.Name,
				Family:       m.Family,
				Reasoning:    m.Reasoning,
				ToolCall:     m.ToolCall,
				Attachment:   m.Attachment,
				Context:      m.Limit.Context,
				Output:       m.Limit.Output,
				Capabilities: m.FormatCapabilities(),
			})
		}
		out = append(out, catProvider{ID: p.ID, Name: p.Name, Models: cms})
	}

	writeJSON(w, http.StatusOK, map[string]any{"providers": out})
}

// writeModelErr maps registry failures to status codes: a renamed-onto-existing
// identity is a conflict, a missing model is 404, and anything else (a failed
// load or write) is a server error rather than a misleading 404.
func writeModelErr(w http.ResponseWriter, err error) {
	var exists *models.ModelExistsError
	var notFound *models.ModelNotFoundError
	switch {
	case errors.As(err, &exists):
		writeErr(w, http.StatusConflict, err.Error())
	case errors.As(err, &notFound):
		writeErr(w, http.StatusNotFound, err.Error())
	default:
		writeErr(w, http.StatusInternalServerError, err.Error())
	}
}
