package models

import (
	"time"
)

// Review represents a review entity in the database.
type Review struct {
	ID        int64     `db:"id"`
	ProductID int64     `db:"product_id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Rating    int       `db:"rating"`
	Comment   *string   `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CreateReviewParams contains parameters for creating a new review.
type CreateReviewParams struct {
	ProductID int64
	FirstName string
	LastName  string
	Rating    int
	Comment   *string
}

// UpdateReviewParams contains parameters for updating a review.
type UpdateReviewParams struct {
	FirstName string
	LastName  string
	Rating    int
	Comment   *string
}

// ListReviewsParams contains parameters for listing reviews.
type ListReviewsParams struct {
	ProductID int64
	Limit     int
	Offset    int
}
