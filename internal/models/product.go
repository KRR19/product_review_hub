// Package models defines domain models for the application.
package models

import (
	"time"
)

// Product represents a product entity in the database.
type Product struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// ProductWithRating represents a product with its average rating.
type ProductWithRating struct {
	Product
	AverageRating *float64 `db:"average_rating"`
}

// CreateProductParams contains parameters for creating a new product.
type CreateProductParams struct {
	Name        string
	Description *string
	Price       float64
}

// UpdateProductParams contains parameters for updating a product.
type UpdateProductParams struct {
	Name        string
	Description *string
	Price       float64
}

// ListProductsParams contains parameters for listing products.
type ListProductsParams struct {
	Limit  int
	Offset int
}
