package handler

import (
	"net/http"

	"github.com/edalcin/smartlattes/internal/store"
)

type StatsHandler struct {
	Store *store.MongoDB
}

func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.Store.CountCVs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao contar currículos"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "count": count})
}
