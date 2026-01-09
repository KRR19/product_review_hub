package products_test

import (
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateProduct(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should update product with valid data", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update product
			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			updatedProduct := assertions.AssertProductUpdated(resp, updateReq, createdProduct.Id)
			assert.Equal(t, updateReq.Name, updatedProduct.Name)
			assert.Equal(t, updateReq.Description, updatedProduct.Description)
			assert.InDelta(t, updateReq.Price, updatedProduct.Price, 0.01)
		})

		t.Run("should update all fields", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with completely different data
			updateReq := fixtures.ValidUpdateRequestWithData(
				"Completely New Name",
				"Completely new description",
				999.99,
			)
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			updatedProduct := assertions.AssertProductUpdated(resp, updateReq, createdProduct.Id)
			assert.Equal(t, "Completely New Name", updatedProduct.Name)
			assert.Equal(t, "Completely new description", updatedProduct.Description)
			assert.InDelta(t, 999.99, updatedProduct.Price, 0.01)
		})

		t.Run("should preserve average_rating after update", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Add reviews
			e2e.CreateTestReviewWithRating(t, client, productID, 4)
			e2e.CreateTestReviewWithRating(t, client, productID, 5)

			// Update product
			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/"+productID, updateReq)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			updatedProduct := e2e.ParseJSON[api.Product](t, resp)

			// average_rating should still be calculated from reviews
			require.NotNil(t, updatedProduct.AverageRating)
			assert.InDelta(t, 4.5, *updatedProduct.AverageRating, 0.01)
		})

		t.Run("should update product with unicode characters", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with unicode
			updateReq := fixtures.ValidUpdateRequestWithData(
				"Продукт 产品 🎉",
				"Описание продукта с эмодзи 😀",
				199.99,
			)
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			updatedProduct := assertions.AssertProductUpdated(resp, updateReq, createdProduct.Id)
			assert.Equal(t, "Продукт 产品 🎉", updatedProduct.Name)
		})

		t.Run("should update product with special characters", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with special chars
			updateReq := fixtures.ValidUpdateRequestWithData(
				"Product with 'quotes' & <special> \"chars\"",
				"Description with special chars",
				99.99,
			)
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			updatedProduct := assertions.AssertProductUpdated(resp, updateReq, createdProduct.Id)
			assert.Equal(t, "Product with 'quotes' & <special> \"chars\"", updatedProduct.Name)
		})
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Run("should return 404 for non-existent product", func(t *testing.T) {
			env.CleanupProducts(t)

			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/999999", updateReq)

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})

		t.Run("should return 404 for deleted product", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create and delete product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			deleteResp := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

			// Try to update deleted product
			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})
	})

	t.Run("Validation Errors", func(t *testing.T) {
		t.Run("should return 400 for empty name", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with empty name
			updateReq := fixtures.UpdateRequestWithEmptyName()
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			assertions.AssertBadRequestWithMessage(resp, "name")
		})

		t.Run("should return 400 for zero price", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with zero price
			updateReq := fixtures.UpdateRequestWithZeroPrice()
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			assertions.AssertBadRequestWithMessage(resp, "price")
		})

		t.Run("should return 400 for negative price", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with negative price
			updateReq := fixtures.UpdateRequestWithNegativePrice()
			resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

			assertions.AssertBadRequestWithMessage(resp, "price")
		})
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Run("should return 400 for invalid JSON", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with invalid JSON
			resp := client.PutRaw(
				productsEndpoint+"/"+createdProduct.Id,
				[]byte("{invalid json}"),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for empty body", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with empty body
			resp := client.PutRaw(
				productsEndpoint+"/"+createdProduct.Id,
				[]byte(""),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for null body", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with null body
			resp := client.PutRaw(
				productsEndpoint+"/"+createdProduct.Id,
				[]byte("null"),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for wrong type in price field", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Update with wrong type
			body := []byte(`{"name": "Test", "description": "Test", "price": "not a number"}`)
			resp := client.PutRaw(
				productsEndpoint+"/"+createdProduct.Id,
				body,
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})
	})

	t.Run("Invalid ID", func(t *testing.T) {
		t.Run("should return 400 for non-numeric ID", func(t *testing.T) {
			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/invalid", updateReq)

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for floating point ID", func(t *testing.T) {
			updateReq := fixtures.ValidUpdateRequest()
			resp := client.Put(productsEndpoint+"/1.5", updateReq)

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})
	})
}

func TestUpdateProductContentType(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("should return JSON content type", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product
		createReq := fixtures.ValidCreateRequest()
		createResp := client.Post(productsEndpoint, createReq)
		createdProduct := assertions.AssertProductCreated(createResp, createReq)

		// Update product
		updateReq := fixtures.ValidUpdateRequest()
		resp := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)

		assertions.AssertContentTypeJSON(resp)
	})
}

func TestUpdateProductIdempotency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("should return same result for repeated updates", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product
		createReq := fixtures.ValidCreateRequest()
		createResp := client.Post(productsEndpoint, createReq)
		createdProduct := assertions.AssertProductCreated(createResp, createReq)

		updateReq := fixtures.ValidUpdateRequest()

		// First update
		resp1 := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)
		product1 := assertions.AssertProductUpdated(resp1, updateReq, createdProduct.Id)

		// Second update with same data
		resp2 := client.Put(productsEndpoint+"/"+createdProduct.Id, updateReq)
		product2 := assertions.AssertProductUpdated(resp2, updateReq, createdProduct.Id)

		// Results should be the same
		assert.Equal(t, product1.Name, product2.Name)
		assert.Equal(t, product1.Description, product2.Description)
		assert.Equal(t, product1.Price, product2.Price)
	})
}
