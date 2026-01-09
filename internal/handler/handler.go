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
	Exists(ctx context.Context, id int64) (bool, error)
}

// ReviewRepository defines interface for review operations.
type ReviewRepository interface {
	Create(ctx context.Context, params models.CreateReviewParams) (*models.Review, error)
	GetByIDAndProductID(ctx context.Context, id, productID int64) (*models.Review, error)
	ListByProductID(ctx context.Context, params models.ListReviewsParams) ([]models.Review, error)
	UpdateByIDAndProductID(ctx context.Context, id, productID int64, params models.UpdateReviewParams) (*models.Review, error)
	DeleteByIDAndProductID(ctx context.Context, id, productID int64) error
	HasReviewsByProductID(ctx context.Context, productID int64) (bool, error)
}

// Handler implements all API handlers.
type Handler struct {
	DB          *sqlx.DB
	ProductRepo ProductRepository
	ReviewRepo  ReviewRepository
}

// New creates a new Handler instance.
func New(db *sqlx.DB, productRepo ProductRepository, reviewRepo ReviewRepository) *Handler {
	return &Handler{
		DB:          db,
		ProductRepo: productRepo,
		ReviewRepo:  reviewRepo,
	}
}
