package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/ai"
)

type ModelsHandler struct{}

func (h *ModelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" || req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "provider e apiKey são obrigatórios"})
		return
	}

	provider, err := ai.NewProvider(req.Provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}

	models, err := provider.ListModels(r.Context(), req.APIKey)
	if err != nil {
		if errors.Is(err, ai.ErrInvalidKey) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Chave de API inválida ou sem permissão para este provedor"})
			return
		}
		if errors.Is(err, ai.ErrRateLimited) {
			detail := strings.TrimPrefix(err.Error(), ai.ErrRateLimited.Error()+": ")
			writeJSON(w, 429, map[string]any{"success": false, "error": "Limite de requisições atingido: " + detail})
			return
		}
		if errors.Is(err, ai.ErrProviderUnavailable) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "Provedor de IA indisponível. Tente novamente mais tarde."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "models": models})
}
