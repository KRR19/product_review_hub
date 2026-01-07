package handler

import (
	"product_review_hub/internal/api"
)

var _ api.ServerInterface = (*Handler)(nil)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}
