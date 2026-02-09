package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/export"
	"github.com/edalcin/smartlattes/internal/store"
)

type AnalysisDownloadHandler struct {
	Store *store.MongoDB
}

func (h *AnalysisDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/analysis/download/")
	lattesID := strings.TrimSuffix(path, "/")
	if lattesID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesID é obrigatório"})
		return
	}

	format := r.URL.Query().Get("format")
	if format != "md" && format != "docx" && format != "pdf" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "formato deve ser md, docx ou pdf"})
		return
	}

	doc, err := h.Store.GetAnalysis(r.Context(), lattesID)
	if err != nil {
		if err.Error() == "análise não encontrada" {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "análise não encontrada"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	filename := fmt.Sprintf("analise-%s", lattesID)

	switch format {
	case "md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.md", filename))
		w.Write(export.ToMarkdown(doc.Analise))
	case "pdf":
		writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "error": "exportação PDF ainda não implementada"})
	case "docx":
		writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "error": "exportação DOCX ainda não implementada"})
	}
}
