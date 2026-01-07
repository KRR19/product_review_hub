package handler

import (
	"encoding/json"
	"net/http"
	"product_review_hub/internal/api"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

var _ api.ServerInterface = (*Handler)(nil)

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := api.HealthResponse{
		Status: "ok",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetReviews(w http.ResponseWriter, r *http.Request) {
	reviews := []api.Review{
		{
			Id:        "1",
			ProductId: "prod-100",
			Rating:    5,
			Comment:   strPtr("Excellent product!"),
		},
		{
			Id:        "2",
			ProductId: "prod-200",
			Rating:    4,
			Comment:   strPtr("Good quality"),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reviews)
}

func strPtr(s string) *string {
	return &s
}
