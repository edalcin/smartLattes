package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/ai"
	"github.com/edalcin/smartlattes/internal/store"
)

type AnalysisHandler struct {
	Store  *store.MongoDB
	Prompt string
}

func (h *AnalysisHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}
	if r.URL.Path == "/api/analysis/save" {
		h.handleSave(w, r)
		return
	}
	h.handleGenerate(w, r)
}

func (h *AnalysisHandler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LattesID string `json:"lattesId"`
		Provider string `json:"provider"`
		APIKey   string `json:"apiKey"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LattesID == "" || req.Provider == "" || req.APIKey == "" || req.Model == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesId, provider, apiKey e model são obrigatórios"})
		return
	}

	ctx := r.Context()

	cvData, err := h.Store.GetCV(ctx, req.LattesID)
	if err != nil {
		if err.Error() == "CV não encontrado" {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "CV não encontrado para o ID informado"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	count, err := h.Store.CountCVs(ctx)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}
	if count <= 1 {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "error": "Não há outros pesquisadores na base para comparação"})
		return
	}

	otherCVs, err := h.Store.GetAllCVSummaries(ctx, req.LattesID)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	userData, wasTruncated := ai.TruncateAnalysisData(cvData, otherCVs, 80000)

	provider, err := ai.NewProvider(req.Provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}

	analysis, err := provider.Generate(ctx, ai.GenerateRequest{
		APIKey:       req.APIKey,
		Model:        req.Model,
		SystemPrompt: h.Prompt,
		UserData:     userData,
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

	header := buildSummaryHeader(cvData, req.LattesID, req.Provider, req.Model)
	analysis = header + analysis

	// Salvar automaticamente no banco de dados
	if err := h.Store.UpsertAnalysis(r.Context(), req.LattesID, analysis, req.Provider, req.Model, len(otherCVs)); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "análise gerada mas erro ao salvar no banco de dados"})
		return
	}

	response := map[string]any{
		"success":             true,
		"analysis":            analysis,
		"provider":            req.Provider,
		"model":               req.Model,
		"researchersAnalyzed": len(otherCVs),
	}
	if wasTruncated {
		response["truncated"] = true
		response["truncationWarning"] = "Os dados dos pesquisadores foram truncados para caber no limite do modelo. Algumas informações podem estar ausentes na análise."
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AnalysisHandler) handleSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LattesID            string `json:"lattesId"`
		Analysis            string `json:"analysis"`
		Provider            string `json:"provider"`
		Model               string `json:"model"`
		ResearchersAnalyzed int    `json:"researchersAnalyzed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LattesID == "" || req.Analysis == "" || req.Provider == "" || req.Model == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesId, analysis, provider e model são obrigatórios"})
		return
	}

	if err := h.Store.UpsertAnalysis(r.Context(), req.LattesID, req.Analysis, req.Provider, req.Model, req.ResearchersAnalyzed); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao salvar análise"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Análise salva com sucesso"})
}
