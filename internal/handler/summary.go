package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/edalcin/smartlattes/internal/ai"
	"github.com/edalcin/smartlattes/internal/store"
)

type SummaryHandler struct {
	Store  *store.MongoDB
	Prompt string
}

func (h *SummaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "método não permitido"})
		return
	}

	if r.URL.Path == "/api/summary/save" {
		h.handleSave(w, r)
		return
	}
	h.handleGenerate(w, r)
}

func (h *SummaryHandler) handleGenerate(w http.ResponseWriter, r *http.Request) {
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

	cvData, err := h.Store.GetCV(r.Context(), req.LattesID)
	if err != nil {
		if err.Error() == "CV não encontrado" {
			writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "CV não encontrado para o ID informado"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao acessar banco de dados"})
		return
	}

	cvJSON, err := json.Marshal(cvData)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "erro ao processar dados do CV"})
		return
	}

	truncatedData, wasTruncated := ai.TruncateCV(cvData, 20000)
	userData := string(cvJSON)
	if wasTruncated {
		truncatedJSON, _ := json.Marshal(truncatedData)
		userData = string(truncatedJSON)
	}

	provider, err := ai.NewProvider(req.Provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}

	summary, err := provider.Generate(r.Context(), ai.GenerateRequest{
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
	summary = header + summary

	response := map[string]any{
		"success":  true,
		"summary":  summary,
		"provider": req.Provider,
		"model":    req.Model,
	}
	if wasTruncated {
		response["truncated"] = true
		response["truncationWarning"] = "Os dados do CV foram truncados para caber no limite do modelo. Algumas informações podem estar ausentes no resumo."
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *SummaryHandler) handleSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LattesID string `json:"lattesId"`
		Summary  string `json:"summary"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LattesID == "" || req.Summary == "" || req.Provider == "" || req.Model == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "lattesId, summary, provider e model são obrigatórios"})
		return
	}

	if err := h.Store.UpsertSummary(r.Context(), req.LattesID, req.Summary, req.Provider, req.Model); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "erro ao salvar resumo"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Resumo salvo com sucesso"})
}

func buildSummaryHeader(cvData map[string]any, lattesID, provider, model string) string {
	name := "Pesquisador"
	lastUpdate := ""

	if cv, ok := cvData["curriculo-vitae"].(map[string]any); ok {
		if dg, ok := cv["dados-gerais"].(map[string]any); ok {
			if n, ok := dg["nome-completo"].(string); ok && n != "" {
				name = n
			}
		}
		if dt, ok := cv["data-atualizacao"].(string); ok && len(dt) == 8 {
			lastUpdate = dt[0:2] + "/" + dt[2:4] + "/" + dt[4:8]
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", name))
	sb.WriteString(fmt.Sprintf("**Acesse o Lattes em:** [http://lattes.cnpq.br/%s](http://lattes.cnpq.br/%s)\n\n", lattesID, lattesID))
	sb.WriteString(fmt.Sprintf("**ID Lattes:** %s\n\n", lattesID))
	if lastUpdate != "" {
		sb.WriteString(fmt.Sprintf("**Última Atualização:** %s\n\n", lastUpdate))
	}
	sb.WriteString(fmt.Sprintf("**Gerado por:** %s / %s\n\n", provider, model))
	sb.WriteString("---\n\n")

	return sb.String()
}
