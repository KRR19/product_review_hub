package products_test

import (
	"net/http"
	"testing"

	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProductById(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should return product by ID", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			req := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, req)
			createdProduct := assertions.AssertProductCreated(createResp, req)

			// Get product by ID
			resp := client.Get(productsEndpoint + "/" + createdProduct.Id)

			product := assertions.AssertProductByID(resp, createdProduct.Id)
			assert.Equal(t, req.Name, product.Name)
			assert.Equal(t, req.Description, product.Description)
			assert.InDelta(t, req.Price, product.Price, 0.01)
		})

		t.Run("should return product with nil average_rating when no reviews", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			req := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, req)
			createdProduct := assertions.AssertProductCreated(createResp, req)

			// Get product by ID
			resp := client.Get(productsEndpoint + "/" + createdProduct.Id)

			product := assertions.AssertProductByID(resp, createdProduct.Id)
			assert.Nil(t, product.AverageRating, "average_rating should be nil when no reviews")
		})

		t.Run("should return product with average_rating when reviews exist", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create reviews with different ratings
			e2e.CreateTestReviewWithRating(t, client, productID, 3)
			e2e.CreateTestReviewWithRating(t, client, productID, 5)

			// Get product by ID
			resp := client.Get(productsEndpoint + "/" + productID)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			product := e2e.ParseJSON[e2e.ProductResponse](t, resp)
			require.NotNil(t, product.AverageRating, "average_rating should not be nil when reviews exist")
			assert.InDelta(t, 4.0, *product.AverageRating, 0.01, "average_rating should be 4.0")
		})

		t.Run("should return product with correct average_rating for single review", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create single review
			e2e.CreateTestReviewWithRating(t, client, productID, 5)

			// Get product by ID
			resp := client.Get(productsEndpoint + "/" + productID)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			product := e2e.ParseJSON[e2e.ProductResponse](t, resp)
			require.NotNil(t, product.AverageRating)
			assert.InDelta(t, 5.0, *product.AverageRating, 0.01)
		})
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Run("should return 404 for non-existent product ID", func(t *testing.T) {
			env.CleanupProducts(t)

			resp := client.Get(productsEndpoint + "/999999")

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})

		t.Run("should return 404 for deleted product", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create and delete product
			req := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, req)
			createdProduct := assertions.AssertProductCreated(createResp, req)

			deleteResp := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

			// Try to get deleted product
			resp := client.Get(productsEndpoint + "/" + createdProduct.Id)

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})
	})

	t.Run("Invalid ID", func(t *testing.T) {
		t.Run("should return 400 for non-numeric ID", func(t *testing.T) {
			resp := client.Get(productsEndpoint + "/invalid")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for empty ID", func(t *testing.T) {
			// Note: Empty ID would match the list endpoint, so we test with spaces
			resp := client.Get(productsEndpoint + "/abc123")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for floating point ID", func(t *testing.T) {
			resp := client.Get(productsEndpoint + "/1.5")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})
	})
}

func TestGetProductByIdContentType(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("should return JSON content type", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product
		req := fixtures.ValidCreateRequest()
		createResp := client.Post(productsEndpoint, req)
		createdProduct := assertions.AssertProductCreated(createResp, req)

		// Get product by ID
		resp := client.Get(productsEndpoint + "/" + createdProduct.Id)

		assertions.AssertContentTypeJSON(resp)
	})
}
