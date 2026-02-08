package handler

import (
	"net/http"

	"github.com/edalcin/smartlattes/internal/store"
)

type SearchHandler struct {
	Store *store.MongoDB
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	q := r.URL.Query().Get("q")
	if len(q) < 3 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "busca deve ter pelo menos 3 caracteres"})
		return
	}

	results, err := h.Store.SearchCVs(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao buscar CVs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "results": results})
}
