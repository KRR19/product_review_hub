package handler

import (
	"net/http"
	"product_review_hub/internal/api"
)

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request, params api.GetProductsParams) {
	panic("unimplemented")
}

func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request, productId string) {
	panic("unimplemented")
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request, productId string) {
	panic("unimplemented")
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request, productId string) {
	panic("unimplemented")
}
