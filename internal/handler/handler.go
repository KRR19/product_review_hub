package handler

import (
	"context"
	"product_review_hub/internal/api"
	"product_review_hub/internal/models"

	"github.com/jmoiron/sqlx"
)

var _ api.ServerInterface = (*Handler)(nil)

// ProductRepository defines interface for product operations.
type ProductRepository interface {
	Create(ctx context.Context, params models.CreateProductParams) (*models.Product, error)
	GetByID(ctx context.Context, id int64) (*models.ProductWithRating, error)
	List(ctx context.Context, params models.ListProductsParams) ([]models.ProductWithRating, error)
	Update(ctx context.Context, id int64, params models.UpdateProductParams) (*models.Product, error)
	Delete(ctx context.Context, id int64) error
}

type Handler struct{
	DB              *sqlx.DB
	ProductRepo     ProductRepository
}

func New(db *sqlx.DB, productRepo ProductRepository) *Handler {
	return &Handler{
		DB:          db,
		ProductRepo: productRepo,
	}
}
