package handler

import (
	"product_review_hub/internal/api"

	"github.com/jmoiron/sqlx"
)

var _ api.ServerInterface = (*Handler)(nil)

type Handler struct{
	DB *sqlx.DB
}

func New(db *sqlx.DB) *Handler {
	return &Handler{
		DB: db,
	}
}
