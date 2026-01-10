package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Database string `json:"database"`
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	dbStatus := "healthy"
	if err := h.DB.PingContext(ctx); err != nil {
		dbStatus = "unhealthy: " + err.Error()
		response := HealthResponse{
			Status:   "unhealthy",
			Message:  "Service is degraded",
			Database: dbStatus,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		//nolint:errcheck // Response write error cannot be handled meaningfully here
		json.NewEncoder(w).Encode(response)
		return
	}

	response := HealthResponse{
		Status:   "healthy",
		Message:  "Service is running",
		Database: dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
