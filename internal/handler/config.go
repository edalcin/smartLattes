package handler

import "net/http"

type ConfigHandler struct {
	ShareBaseURL string
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"shareBaseUrl": h.ShareBaseURL})
}
