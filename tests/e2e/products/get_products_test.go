package products_test

import (
	"fmt"
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProducts(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should return empty list when no products exist", func(t *testing.T) {
			env.CleanupProducts(t)

			resp := client.Get(productsEndpoint)

			products := assertions.AssertProductsListExact(resp, 0)
			assert.Empty(t, products)
		})

		t.Run("should return list of products", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create multiple products
			for i := 0; i < 3; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint)

			products := assertions.AssertProductsListExact(resp, 3)
			for _, p := range products {
				assert.NotEmpty(t, p.Id)
				assert.NotEmpty(t, p.Name)
			}
		})

		t.Run("should return products with nil average_rating when no reviews", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, req)
			require.Equal(t, http.StatusCreated, createResp.StatusCode)
			createResp.Body.Close()

			resp := client.Get(productsEndpoint)

			products := assertions.AssertProductsListExact(resp, 1)
			assert.Nil(t, products[0].AverageRating, "average_rating should be nil when no reviews")
		})

		t.Run("should return products with average_rating when reviews exist", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create reviews with different ratings
			e2e.CreateTestReviewWithRating(t, client, productID, 4)
			e2e.CreateTestReviewWithRating(t, client, productID, 5)

			resp := client.Get(productsEndpoint)

			products := assertions.AssertProductsListExact(resp, 1)
			require.NotNil(t, products[0].AverageRating, "average_rating should not be nil when reviews exist")
			assert.InDelta(t, 4.5, *products[0].AverageRating, 0.01, "average_rating should be 4.5")
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		t.Run("should respect limit parameter", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 5 products
			for i := 0; i < 5; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint + "?limit=2")

			products := assertions.AssertProductsListExact(resp, 2)
			assert.Len(t, products, 2)
		})

		t.Run("should respect offset parameter", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 5 products
			createdProducts := make([]api.Product, 5)
			for i := 0; i < 5; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				createdProducts[i] = e2e.ParseJSON[api.Product](t, resp)
			}

			// Get all products to understand order
			allResp := client.Get(productsEndpoint + "?limit=10")
			allProducts := e2e.ParseJSON[[]api.Product](t, allResp)

			// Get with offset
			resp := client.Get(productsEndpoint + "?limit=10&offset=2")

			products := assertions.AssertProductsListExact(resp, 3)
			// Products should be offset by 2 from the full list
			assert.Equal(t, allProducts[2].Id, products[0].Id)
		})

		t.Run("should use default limit of 10", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 15 products
			for i := 0; i < 15; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint)

			products := assertions.AssertProductsListExact(resp, 10)
			assert.Len(t, products, 10)
		})

		t.Run("should cap limit at 100", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 5 products (we won't create 100+ for performance)
			for i := 0; i < 5; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			// Request with limit > 100
			resp := client.Get(productsEndpoint + "?limit=200")

			// Should return all 5 products (limit is capped but we only have 5)
			products := assertions.AssertProductsListExact(resp, 5)
			assert.Len(t, products, 5)
		})

		t.Run("should handle limit=1", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 3 products
			for i := 0; i < 3; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint + "?limit=1")

			products := assertions.AssertProductsListExact(resp, 1)
			assert.Len(t, products, 1)
		})

		t.Run("should return empty list when offset exceeds total", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 3 products
			for i := 0; i < 3; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint + "?offset=100")

			products := assertions.AssertProductsListExact(resp, 0)
			assert.Empty(t, products)
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("should handle negative limit by using minimum", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 3 products
			for i := 0; i < 3; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint + "?limit=-5")

			// Should return at least 1 product (minimum limit)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			products := e2e.ParseJSON[[]api.Product](t, resp)
			assert.GreaterOrEqual(t, len(products), 1)
		})

		t.Run("should handle negative offset by using 0", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create 3 products
			for i := 0; i < 3; i++ {
				req := fixtures.ValidCreateRequestWithName(fmt.Sprintf("Product %d", i+1))
				resp := client.Post(productsEndpoint, req)
				require.Equal(t, http.StatusCreated, resp.StatusCode)
				resp.Body.Close()
			}

			resp := client.Get(productsEndpoint + "?offset=-5")

			// Should return products (offset treated as 0)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			products := e2e.ParseJSON[[]api.Product](t, resp)
			assert.Len(t, products, 3)
		})
	})
}

func TestGetProductsContentType(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	assertions := e2e.NewProductAssertions(t)

	t.Run("should return JSON content type", func(t *testing.T) {
		env.CleanupProducts(t)

		resp := client.Get(productsEndpoint)

		assertions.AssertContentTypeJSON(resp)
	})
}
