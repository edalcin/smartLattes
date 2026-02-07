package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/edalcin/smartlattes/internal/parser"
	"github.com/edalcin/smartlattes/internal/store"
)

type UploadHandler struct {
	Store         *store.MongoDB
	MaxUploadSize int64
}

type uploadResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Updated bool        `json:"updated,omitempty"`
	Data    *uploadData `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type uploadData struct {
	LattesID   string           `json:"lattesId"`
	Name       string           `json:"name"`
	LastUpdate string           `json:"lastUpdate"`
	Counts     *productionCount `json:"counts,omitempty"`
}

type productionCount struct {
	BibliographicProduction int `json:"bibliographicProduction"`
	TechnicalProduction     int `json:"technicalProduction"`
	OtherProduction         int `json:"otherProduction"`
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Método não permitido")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.MaxUploadSize)

	if err := r.ParseMultipartForm(h.MaxUploadSize); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "Arquivo excede o tamanho máximo permitido de 10MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Nenhum arquivo enviado")
		return
	}
	defer file.Close()

	if header.Size == 0 {
		writeError(w, http.StatusBadRequest, "Arquivo vazio")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Erro ao ler o arquivo")
		return
	}

	result, err := parser.Parse(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Store == nil {
		writeError(w, http.StatusServiceUnavailable, "Banco de dados indisponível")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	upsertResult, err := h.Store.UpsertCV(ctx, result.Document, result.Summary.LattesID, header.Filename, header.Size)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Erro ao salvar no banco de dados")
		return
	}

	resp := uploadResponse{
		Success: true,
		Message: "CV importado com sucesso",
		Updated: upsertResult.Updated,
		Data: &uploadData{
			LattesID:   result.Summary.LattesID,
			Name:       result.Summary.Name,
			LastUpdate: result.Summary.LastUpdate,
			Counts: &productionCount{
				BibliographicProduction: result.Summary.Counts.BibliographicProduction,
				TechnicalProduction:     result.Summary.Counts.TechnicalProduction,
				OtherProduction:         result.Summary.Counts.OtherProduction,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(uploadResponse{
		Success: false,
		Error:   message,
	})
}
