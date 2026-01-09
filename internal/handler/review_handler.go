package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"product_review_hub/internal/api"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/reviews"
)

const (
	defaultReviewLimit = 10
	maxReviewLimit     = 100

	minRating = 1
	maxRating = 5
)

// CreateProductReview creates a new review for a product.
func (h *Handler) CreateProductReview(w http.ResponseWriter, r *http.Request, productId string) {
	// Parse product ID
	prodID, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Decode request body
	var req api.ReviewCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateRating(req.Rating); err != nil {
		responseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Check if product exists
	exists, err := h.ProductRepo.Exists(r.Context(), tx, prodID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to check product existence")
		return
	}
	if !exists {
		responseError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Prepare create params
	params := models.CreateReviewParams{
		ProductID: prodID,
		Rating:    req.Rating,
		Author:    getStringPtrValue(req.Author),
		Comment:   req.Comment,
	}

	// Create review
	review, err := h.ReviewRepo.Create(r.Context(), tx, params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to create review")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Return created review
	responseJSON(w, http.StatusCreated, reviewToResponse(review))
}

// GetProductReviews returns a paginated list of reviews for a product.
func (h *Handler) GetProductReviews(w http.ResponseWriter, r *http.Request, productId string, params api.GetProductReviewsParams) {
	// Parse product ID
	prodID, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Check if product exists
	exists, err := h.ProductRepo.Exists(r.Context(), tx, prodID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to check product existence")
		return
	}
	if !exists {
		responseError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Apply pagination defaults
	limit := defaultReviewLimit
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
		if limit < 1 {
			limit = 1
		}
		if limit > maxReviewLimit {
			limit = maxReviewLimit
		}
	}

	if params.Offset != nil {
		offset = *params.Offset
		if offset < 0 {
			offset = 0
		}
	}

	// Fetch reviews
	reviewList, err := h.ReviewRepo.ListByProductID(r.Context(), tx, models.ListReviewsParams{
		ProductID: prodID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to fetch reviews")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Convert to API response
	response := make([]api.Review, len(reviewList))
	for i, rev := range reviewList {
		response[i] = reviewToResponse(&rev)
	}

	responseJSON(w, http.StatusOK, response)
}

// UpdateProductReview updates an existing review.
func (h *Handler) UpdateProductReview(w http.ResponseWriter, r *http.Request, productId string, reviewId string) {
	// Parse IDs
	prodID, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	revID, err := parseID(reviewId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid review ID")
		return
	}

	// Decode request body
	var req api.ReviewUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateRating(req.Rating); err != nil {
		responseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Prepare update params
	params := models.UpdateReviewParams{
		Rating:  req.Rating,
		Author:  getStringPtrValue(req.Author),
		Comment: req.Comment,
	}

	// Begin transaction
	tx, err := h.ReviewRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Update review
	review, err := h.ReviewRepo.UpdateByIDAndProductID(r.Context(), tx, revID, prodID, params)
	if err != nil {
		if errors.Is(err, reviews.ErrNotFound) {
			responseError(w, http.StatusNotFound, "Review not found")
			return
		}
		responseError(w, http.StatusInternalServerError, "Failed to update review")
		return
	}

	// Commit transaction
	if err := h.ReviewRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	responseJSON(w, http.StatusOK, reviewToResponse(review))
}

// DeleteProductReview deletes a review.
func (h *Handler) DeleteProductReview(w http.ResponseWriter, r *http.Request, productId string, reviewId string) {
	// Parse IDs
	prodID, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	revID, err := parseID(reviewId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid review ID")
		return
	}

	// Begin transaction
	tx, err := h.ReviewRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Delete review
	err = h.ReviewRepo.DeleteByIDAndProductID(r.Context(), tx, revID, prodID)
	if err != nil {
		if errors.Is(err, reviews.ErrNotFound) {
			responseError(w, http.StatusNotFound, "Review not found")
			return
		}
		responseError(w, http.StatusInternalServerError, "Failed to delete review")
		return
	}

	// Commit transaction
	if err := h.ReviewRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// validateRating validates that the rating is within the allowed range.
func validateRating(rating int) error {
	if rating < minRating || rating > maxRating {
		return errValidation("rating", "rating must be between 1 and 5")
	}
	return nil
}

// parseID parses a string ID to int64.
func parseID(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

// reviewToResponse converts a Review model to API response.
func reviewToResponse(review *models.Review) api.Review {
	return api.Review{
		Id:        strconv.FormatInt(review.ID, 10),
		ProductId: strconv.FormatInt(review.ProductID, 10),
		Rating:    review.Rating,
		Author:    &review.Author,
		Comment:   review.Comment,
	}
}

// getStringPtrValue returns the string value from a pointer or empty string if nil.
func getStringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
