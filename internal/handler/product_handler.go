package handler

import (
	"encoding/json"
	"net/http"
	"product_review_hub/internal/api"
	"product_review_hub/internal/models"
	"strconv"
)

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req api.ProductCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateProductCreate(req); err != nil {
		responseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create product params
	params := models.CreateProductParams{
		Name:        req.Name,
		Description: &req.Description,
		Price:       float64(req.Price),
	}

	// Create product in database
	product, err := h.ProductRepo.Create(r.Context(), params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	// Convert to API response
	response := api.Product{
		Id:          strconv.FormatInt(product.ID, 10),
		Name:        product.Name,
		Description: getStringValue(product.Description),
		Price:       float32(product.Price),
	}

	// Return created product
	responseJSON(w, http.StatusCreated, response)
}

func validateProductCreate(req api.ProductCreate) error {
	if req.Name == "" {
		return errValidation("name", "name is required")
	}
	if req.Price <= 0 {
		return errValidation("price", "price must be greater than 0")
	}
	return nil
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
