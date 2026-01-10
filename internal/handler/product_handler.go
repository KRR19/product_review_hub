package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"product_review_hub/internal/api"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/products"
)

const (
	defaultProductLimit = 10
	maxProductLimit     = 100
)

// CreateProduct creates a new product.
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

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Create product in database
	product, err := h.ProductRepo.Create(r.Context(), tx, params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
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

// GetProducts returns a paginated list of products.
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request, params api.GetProductsParams) {
	// Apply pagination defaults
	limit := defaultProductLimit
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
		if limit < 1 {
			limit = 1
		}
		if limit > maxProductLimit {
			limit = maxProductLimit
		}
	}

	if params.Offset != nil {
		offset = *params.Offset
		if offset < 0 {
			offset = 0
		}
	}

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Fetch products from database
	productList, err := h.ProductRepo.List(r.Context(), tx, models.ListProductsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Cache ratings for products
	if h.Cache != nil {
		for i := range productList {
			// Try to get rating from cache
			cachedRating, ratingCached, cacheErr := h.Cache.GetRating(r.Context(), productList[i].ID)
			if cacheErr != nil {
				log.Printf("Failed to get rating from cache for product %d: %v", productList[i].ID, cacheErr)
				continue
			}

			if ratingCached {
				// Use cached rating
				productList[i].AverageRating = cachedRating
			} else {
				// Cache the DB rating
				if err := h.Cache.SetRating(r.Context(), productList[i].ID, productList[i].AverageRating); err != nil {
					log.Printf("Failed to cache rating for product %d: %v", productList[i].ID, err)
				}
			}
		}
	}

	// Convert to API response
	response := make([]api.Product, len(productList))
	for i := range productList {
		response[i] = productToResponse(&productList[i])
	}

	responseJSON(w, http.StatusOK, response)
}

// GetProductById returns a product by its ID.
func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request, productId string) {
	// Parse product ID
	id, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Try to get rating from cache
	var cachedRating *float64
	var ratingCached bool
	if h.Cache != nil {
		var cacheErr error
		cachedRating, ratingCached, cacheErr = h.Cache.GetRating(r.Context(), id)
		if cacheErr != nil {
			log.Printf("Failed to get rating from cache: %v", cacheErr)
		}
	}

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Fetch product from database
	product, err := h.ProductRepo.GetByID(r.Context(), tx, id)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			responseError(w, http.StatusNotFound, "Product not found")
			return
		}
		responseError(w, http.StatusInternalServerError, "Failed to fetch product")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// Use cached rating if available, otherwise cache the DB rating
	if ratingCached {
		product.AverageRating = cachedRating
	} else if h.Cache != nil {
		if err := h.Cache.SetRating(r.Context(), id, product.AverageRating); err != nil {
			log.Printf("Failed to cache rating: %v", err)
		}
	}

	responseJSON(w, http.StatusOK, productToResponse(product))
}

// UpdateProduct updates an existing product.
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request, productId string) {
	// Parse product ID
	id, err := parseID(productId)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Decode request body
	var req api.ProductUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateProductUpdate(req); err != nil {
		responseError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Prepare update params
	params := models.UpdateProductParams{
		Name:        req.Name,
		Description: &req.Description,
		Price:       float64(req.Price),
	}

	// Begin transaction
	tx, err := h.ProductRepo.BeginTx(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback()

	// Update product in database
	product, err := h.ProductRepo.Update(r.Context(), tx, id, params)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			responseError(w, http.StatusNotFound, "Product not found")
			return
		}
		responseError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	// Fetch product with rating for response
	productWithRating, err := h.ProductRepo.GetByID(r.Context(), tx, product.ID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to fetch updated product")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	responseJSON(w, http.StatusOK, productToResponse(productWithRating))
}

func validateProductUpdate(req api.ProductUpdate) error {
	if req.Name == "" {
		return errValidation("name", "name is required")
	}
	if req.Price <= 0 {
		return errValidation("price", "price must be greater than 0")
	}
	return nil
}

// DeleteProduct deletes a product by its ID.
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request, productId string) {
	// Parse product ID
	id, err := parseID(productId)
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
	exists, err := h.ProductRepo.Exists(r.Context(), tx, id)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to check product existence")
		return
	}
	if !exists {
		responseError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Check if product has reviews
	hasReviews, err := h.ReviewRepo.HasReviewsByProductID(r.Context(), tx, id)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to check product reviews")
		return
	}
	if hasReviews {
		responseError(w, http.StatusConflict, "Cannot delete product with existing reviews")
		return
	}

	// Delete product
	err = h.ProductRepo.Delete(r.Context(), tx, id)
	if err != nil {
		if errors.Is(err, products.ErrNotFound) {
			responseError(w, http.StatusNotFound, "Product not found")
			return
		}
		responseError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	// Commit transaction
	if err := h.ProductRepo.CommitTx(r.Context(), tx); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// productToResponse converts ProductWithRating model to API response.
func productToResponse(p *models.ProductWithRating) api.Product {
	var avgRating *float32
	if p.AverageRating != nil {
		rating := float32(*p.AverageRating)
		avgRating = &rating
	}

	return api.Product{
		Id:            strconv.FormatInt(p.ID, 10),
		Name:          p.Name,
		Description:   getStringValue(p.Description),
		Price:         float32(p.Price),
		AverageRating: avgRating,
	}
}
