// Package reviews provides repository for managing reviews in the database.
package reviews

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"product_review_hub/internal/models"

	"github.com/jmoiron/sqlx"
)

// Common errors.
var (
	ErrNotFound = errors.New("review not found")
)

// Repository provides methods for managing reviews in the database.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new reviews repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *Repository) CommitTx(ctx context.Context, tx *sqlx.Tx) error {
	return tx.Commit()
}

// Create inserts a new review into the database.
func (r *Repository) Create(ctx context.Context, tx *sqlx.Tx, params models.CreateReviewParams) (*models.Review, error) {
	query := `
		INSERT INTO reviews (product_id, author, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, author, rating, comment, created_at, updated_at
	`

	var review models.Review
	err := tx.QueryRowxContext(ctx, query, params.ProductID, params.Author, params.Rating, params.Comment).
		StructScan(&review)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return &review, nil
}

// GetByID retrieves a review by its ID.
func (r *Repository) GetByID(ctx context.Context, tx *sqlx.Tx, id int64) (*models.Review, error) {
	query := `
		SELECT id, product_id, author, rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = $1
	`

	var review models.Review
	err := tx.QueryRowxContext(ctx, query, id).StructScan(&review)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

// GetByIDAndProductID retrieves a review by its ID and product ID.
func (r *Repository) GetByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64) (*models.Review, error) {
	query := `
		SELECT id, product_id, author, rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = $1 AND product_id = $2
	`

	var review models.Review
	err := tx.QueryRowxContext(ctx, query, id, productID).StructScan(&review)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

// ListByProductID retrieves all reviews for a specific product.
func (r *Repository) ListByProductID(ctx context.Context, tx *sqlx.Tx, params models.ListReviewsParams) ([]models.Review, error) {
	query := `
		SELECT id, product_id, author, rating, comment, created_at, updated_at
		FROM reviews
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var reviews []models.Review
	err := tx.SelectContext(ctx, &reviews, query, params.ProductID, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	return reviews, nil
}

// Update updates an existing review.
func (r *Repository) Update(ctx context.Context, tx *sqlx.Tx, id int64, params models.UpdateReviewParams) (*models.Review, error) {
	query := `
		UPDATE reviews
		SET author = $1, rating = $2, comment = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, product_id, author, rating, comment, created_at, updated_at
	`

	var review models.Review
	err := tx.QueryRowxContext(ctx, query, params.Author, params.Rating, params.Comment, id).
		StructScan(&review)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	return &review, nil
}

// UpdateByIDAndProductID updates a review by its ID and product ID.
func (r *Repository) UpdateByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64, params models.UpdateReviewParams) (*models.Review, error) {
	query := `
		UPDATE reviews
		SET author = $1, rating = $2, comment = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND product_id = $5
		RETURNING id, product_id, author, rating, comment, created_at, updated_at
	`

	var review models.Review
	err := tx.QueryRowxContext(ctx, query, params.Author, params.Rating, params.Comment, id, productID).
		StructScan(&review)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	return &review, nil
}

// Delete removes a review from the database.
func (r *Repository) Delete(ctx context.Context, tx *sqlx.Tx, id int64) error {
	query := `DELETE FROM reviews WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteByIDAndProductID removes a review by its ID and product ID.
func (r *Repository) DeleteByIDAndProductID(ctx context.Context, tx *sqlx.Tx, id, productID int64) error {
	query := `DELETE FROM reviews WHERE id = $1 AND product_id = $2`

	result, err := tx.ExecContext(ctx, query, id, productID)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetAverageRatingByProductID calculates the average rating for a product.
func (r *Repository) GetAverageRatingByProductID(ctx context.Context, tx *sqlx.Tx, productID int64) (*float64, error) {
	query := `SELECT AVG(rating)::FLOAT FROM reviews WHERE product_id = $1`

	var avgRating *float64
	err := tx.QueryRowxContext(ctx, query, productID).Scan(&avgRating)
	if err != nil {
		return nil, fmt.Errorf("failed to get average rating: %w", err)
	}

	return avgRating, nil
}

// HasReviewsByProductID checks if a product has any reviews.
func (r *Repository) HasReviewsByProductID(ctx context.Context, tx *sqlx.Tx, productID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM reviews WHERE product_id = $1)`

	var exists bool
	err := tx.QueryRowxContext(ctx, query, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check reviews existence: %w", err)
	}

	return exists, nil
}
