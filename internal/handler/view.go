package handler

import (
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/store"
)

type SummaryViewHandler struct {
	Store *store.MongoDB
}

func (h *SummaryViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	lattesID := strings.TrimPrefix(r.URL.Path, "/api/summary/view/")
	lattesID = strings.TrimSuffix(lattesID, "/")
	if lattesID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesID é obrigatório"})
		return
	}

	doc, err := h.Store.GetSummary(r.Context(), lattesID)
	if err != nil {
		if err.Error() == "resumo não encontrado" {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Nenhum resumo salvo para este pesquisador"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"summary":  doc.Resumo,
		"provider": doc.Metadata.Provider,
		"model":    doc.Metadata.Model,
		"generatedAt": doc.Metadata.GeneratedAt,
	})
}

type AnalysisViewHandler struct {
	Store *store.MongoDB
}

func (h *AnalysisViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	lattesID := strings.TrimPrefix(r.URL.Path, "/api/analysis/view/")
	lattesID = strings.TrimSuffix(lattesID, "/")
	if lattesID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesID é obrigatório"})
		return
	}

	doc, err := h.Store.GetAnalysis(r.Context(), lattesID)
	if err != nil {
		if err.Error() == "análise não encontrada" {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "Nenhuma análise salva para este pesquisador"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":             true,
		"analysis":            doc.Analise,
		"provider":            doc.Metadata.Provider,
		"model":               doc.Metadata.Model,
		"generatedAt":         doc.Metadata.GeneratedAt,
		"researchersAnalyzed": doc.Metadata.ResearchersAnalyzed,
	})
}
