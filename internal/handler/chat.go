package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/ai"
	"github.com/edalcin/smartlattes/internal/store"
)

type ChatHandler struct {
	Store  *store.MongoDB
	Prompt string
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	var req struct {
		Provider string           `json:"provider"`
		APIKey   string           `json:"apiKey"`
		Model    string           `json:"model"`
		Messages []ai.ChatMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" || req.APIKey == "" || req.Model == "" || len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "provider, apiKey, model e messages são obrigatórios"})
		return
	}

	ctx := r.Context()

	cvs, err := h.Store.GetAllCVsForChat(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	if len(cvs) == 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "error": "Não há currículos na base de dados. Envie pelo menos um CV antes de usar o chat."})
		return
	}

	cvJSON, err := json.Marshal(cvs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "erro ao processar dados"})
		return
	}

	systemPrompt := strings.Replace(h.Prompt, "{{DATA}}", string(cvJSON), 1)

	provider, err := ai.NewProvider(req.Provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}

	response, err := provider.Chat(ctx, ai.ChatRequest{
		APIKey:       req.APIKey,
		Model:        req.Model,
		SystemPrompt: systemPrompt,
		Messages:     req.Messages,
		MaxTokens:    4096,
	})
	if err != nil {
		if errors.Is(err, ai.ErrInvalidKey) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"success": false, "error": "Chave de API inválida ou sem permissão para este provedor"})
			return
		}
		if errors.Is(err, ai.ErrTimeout) {
			writeJSON(w, http.StatusGatewayTimeout, map[string]any{"success": false, "error": "Tempo limite excedido (120s). Tente um modelo menor ou tente novamente."})
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

	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"response": response,
	})
}
