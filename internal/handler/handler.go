package handler

import (
	"context"
	"product_review_hub/internal/api"
	"product_review_hub/internal/cache"
	"product_review_hub/internal/models"
	"product_review_hub/internal/rabbitmq"

	"github.com/jmoiron/sqlx"
)

var _ api.ServerInterface = (*Handler)(nil)

// ProductRepository defines interface for product operations.
type ProductRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error

	Create(ctx context.Context, tx *sqlx.Tx, params models.CreateProductParams) (*models.Product, error)
	GetByID(ctx context.Context, tx *sqlx.Tx, id int64) (*models.ProductWithRating, error)
	List(ctx context.Context, tx *sqlx.Tx, params models.ListProductsParams) ([]models.ProductWithRating, error)
	Update(ctx context.Context, tx *sqlx.Tx, id int64, params models.UpdateProductParams) (*models.Product, error)
	Delete(ctx context.Context, tx *sqlx.Tx, id int64) error
	Exists(ctx context.Context, tx *sqlx.Tx, id int64) (bool, error)
}

// ReviewRepository defines interface for review operations.
type ReviewRepository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error

	Create(ctx context.Context, tx *sqlx.Tx, params models.CreateReviewParams) (*models.Review, error)
	GetByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64) (*models.Review, error)
	ListByProductID(ctx context.Context, tx *sqlx.Tx, params models.ListReviewsParams) ([]models.Review, error)
	UpdateByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64, params models.UpdateReviewParams) (*models.Review, error)
	DeleteByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64) error
	HasReviewsByProductID(ctx context.Context, tx *sqlx.Tx, productID int64) (bool, error)
}

// Handler implements all API handlers.
type Handler struct {
	DB          *sqlx.DB
	ProductRepo ProductRepository
	ReviewRepo  ReviewRepository
	Publisher   *rabbitmq.Publisher
	Cache       *cache.Service
}

// New creates a new Handler instance.
func New(db *sqlx.DB, productRepo ProductRepository, reviewRepo ReviewRepository, publisher *rabbitmq.Publisher, cacheService *cache.Service) *Handler {
	return &Handler{
		DB:          db,
		ProductRepo: productRepo,
		ReviewRepo:  reviewRepo,
		Publisher:   publisher,
		Cache:       cacheService,
	}
}
