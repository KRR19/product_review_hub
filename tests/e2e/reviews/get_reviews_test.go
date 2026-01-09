package reviews_test

import (
	"fmt"
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/require"
)

func TestGetProductReviews(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	reviewFixtures := e2e.NewReviewFixtures()
	assertions := e2e.NewReviewAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should return empty list for product without reviews", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))

			assertions.AssertContentTypeJSON(resp)
			assertions.AssertReviewsList(resp, 0)
		})

		t.Run("should return reviews for product", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 3 reviews
			for i := 1; i <= 3; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, i)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))

			reviews := assertions.AssertReviewsList(resp, 3)

			// Verify all reviews belong to the correct product
			for _, review := range reviews {
				if review.ProductId != productID {
					t.Errorf("Review product ID mismatch: expected %s, got %s", productID, review.ProductId)
				}
			}
		})

		t.Run("should return reviews with all fields", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create review with full data
			req := reviewFixtures.ValidCreateRequestFull()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)
			require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create review")

			resp = client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))

			reviews := assertions.AssertReviewsList(resp, 1)
			if reviews[0].Author == nil || *reviews[0].Author == "" {
				t.Error("Author should not be empty")
			}
			if reviews[0].Comment == nil || *reviews[0].Comment == "" {
				t.Error("Comment should not be empty")
			}
		})
	})

	t.Run("Pagination", func(t *testing.T) {
		t.Run("should respect limit parameter", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 5 reviews
			for i := 1; i <= 5; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, i)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews?limit=3", productID))

			assertions.AssertReviewsList(resp, 3)
		})

		t.Run("should respect offset parameter", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 5 reviews
			for i := 1; i <= 5; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, i)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews?offset=3", productID))

			assertions.AssertReviewsList(resp, 2)
		})

		t.Run("should handle limit and offset together", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 10 reviews
			for i := 1; i <= 10; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, (i%5)+1)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews?limit=3&offset=5", productID))

			assertions.AssertReviewsList(resp, 3)
		})

		t.Run("should return empty list when offset exceeds count", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 3 reviews
			for i := 1; i <= 3; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, i)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews?offset=10", productID))

			assertions.AssertReviewsList(resp, 0)
		})

		t.Run("should use default limit when not specified", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 15 reviews
			for i := 1; i <= 15; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, (i%5)+1)
			}

			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))

			// Default limit is 10
			assertions.AssertReviewsList(resp, 10)
		})

		t.Run("should cap limit at maximum", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 5 reviews
			for i := 1; i <= 5; i++ {
				e2e.CreateTestReviewWithRating(t, client, productID, i)
			}

			// Request limit higher than available
			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews?limit=1000", productID))

			// Should return all available (capped to max 100)
			assertions.AssertReviewsList(resp, 5)
		})
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		t.Run("should return 404 for non-existent product", func(t *testing.T) {
			env.CleanupProducts(t)

			resp := client.Get("/api/v1/products/99999/reviews")

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})

		t.Run("should return 400 for invalid product ID", func(t *testing.T) {
			resp := client.Get("/api/v1/products/invalid/reviews")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})
	})
}

func TestGetProductReviewsOrdering(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	productFixtures := e2e.NewProductFixtures()
	reviewFixtures := e2e.NewReviewFixtures()

	t.Run("should return reviews in consistent order", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product
		req := productFixtures.ValidCreateRequest()
		resp := client.Post("/api/v1/products", req)
		product := e2e.ParseJSON[api.Product](t, resp)

		// Create multiple reviews with different ratings
		ratings := []int{3, 1, 5, 2, 4}
		createdIDs := make([]string, len(ratings))
		for i, rating := range ratings {
			req := reviewFixtures.ValidCreateRequestWithRating(rating)
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id), req)
			review := e2e.ParseJSON[api.Review](t, resp)
			createdIDs[i] = review.Id
		}

		// Get reviews twice and verify order is consistent
		resp1 := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id))
		reviews1 := e2e.ParseJSON[[]api.Review](t, resp1)

		resp2 := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id))
		reviews2 := e2e.ParseJSON[[]api.Review](t, resp2)

		require.Equal(t, len(reviews1), len(reviews2), "Review counts should match")

		for i := range reviews1 {
			if reviews1[i].Id != reviews2[i].Id {
				t.Errorf("Order mismatch at index %d: %s vs %s", i, reviews1[i].Id, reviews2[i].Id)
			}
		}
	})
}
