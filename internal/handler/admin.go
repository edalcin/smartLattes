package handler

import (
	"crypto/subtle"
	"net/http"

	"github.com/edalcin/smartlattes/internal/store"
)

type AdminResearchersHandler struct {
	Store    *store.MongoDB
	AdminPIN string
}

func (h *AdminResearchersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	if h.AdminPIN == "" {
		writeJSON(w, http.StatusForbidden, map[string]any{"success": false, "error": "Admin desabilitado"})
		return
	}

	pin := r.Header.Get("X-Admin-PIN")
	if subtle.ConstantTimeCompare([]byte(pin), []byte(h.AdminPIN)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "PIN inválido"})
		return
	}

	researchers, err := h.Store.GetAllResearchersAdmin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "erro ao buscar pesquisadores"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "researchers": researchers})
}
