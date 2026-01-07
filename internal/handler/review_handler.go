package handler

import (
	"net/http"
	"product_review_hub/internal/api"
)

func (h *Handler) CreateProductReview(w http.ResponseWriter, r *http.Request, productId string) {
	panic("unimplemented")
}

func (h *Handler) DeleteProductReview(w http.ResponseWriter, r *http.Request, productId string, reviewId string) {
	panic("unimplemented")
}

func (h *Handler) GetProductReviews(w http.ResponseWriter, r *http.Request, productId string, params api.GetProductReviewsParams) {
	panic("unimplemented")
}

func (h *Handler) UpdateProductReview(w http.ResponseWriter, r *http.Request, productId string, reviewId string) {
	panic("unimplemented")
}
