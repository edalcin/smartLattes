package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/edalcin/smartlattes/internal/store"
)

type HealthHandler struct {
	Store *store.MongoDB
}

type healthResponse struct {
	Status  string `json:"status"`
	MongoDB string `json:"mongodb"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.Store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(healthResponse{
			Status:  "unhealthy",
			MongoDB: "disconnected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.Store.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(healthResponse{
			Status:  "unhealthy",
			MongoDB: "disconnected",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse{
		Status:  "healthy",
		MongoDB: "connected",
	})
}
