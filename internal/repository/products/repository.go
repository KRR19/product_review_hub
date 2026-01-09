// Package products provides repository for managing products in the database.
package products

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
	ErrNotFound = errors.New("product not found")
)

// Repository provides methods for managing products in the database.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new products repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new product into the database.
func (r *Repository) Create(ctx context.Context, params models.CreateProductParams) (*models.Product, error) {
	query := `
		INSERT INTO products (name, description, price)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, price, created_at, updated_at
	`

	var product models.Product
	err := r.db.QueryRowxContext(ctx, query, params.Name, params.Description, params.Price).
		StructScan(&product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

// GetByID retrieves a product by its ID with average rating.
func (r *Repository) GetByID(ctx context.Context, id int64) (*models.ProductWithRating, error) {
	//this query is not efficient because it joins the reviews table on every product but for this simple project it is fine
	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
			AVG(r.rating)::FLOAT AS average_rating
		FROM products p
		LEFT JOIN reviews r ON p.id = r.product_id
		WHERE p.id = $1
		GROUP BY p.id
	`

	var product models.ProductWithRating
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&product)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves a list of products with pagination.
func (r *Repository) List(ctx context.Context, params models.ListProductsParams) ([]models.ProductWithRating, error) {
	//this query is not efficient because it joins the reviews table on every product but for this simple project it is fine
	query := `
		SELECT 
			p.id, p.name, p.description, p.price, p.created_at, p.updated_at,
			AVG(r.rating)::FLOAT AS average_rating
		FROM products p
		LEFT JOIN reviews r ON p.id = r.product_id
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	var products []models.ProductWithRating
	err := r.db.SelectContext(ctx, &products, query, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// Update updates an existing product.
func (r *Repository) Update(ctx context.Context, id int64, params models.UpdateProductParams) (*models.Product, error) {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, name, description, price, created_at, updated_at
	`

	var product models.Product
	err := r.db.QueryRowxContext(ctx, query, params.Name, params.Description, params.Price, id).
		StructScan(&product)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

// Delete removes a product from the database.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
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

// Exists checks if a product with the given ID exists.
func (r *Repository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowxContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}
