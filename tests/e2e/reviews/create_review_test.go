package reviews_test

import (
	"fmt"
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/require"
)

func TestCreateProductReview(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	reviewFixtures := e2e.NewReviewFixtures()
	assertions := e2e.NewReviewAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should create review with minimal data (rating only)", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.ValidCreateRequest()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertContentTypeJSON(resp)
			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create review with all fields", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.ValidCreateRequestFull()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create review with rating 1 (minimum)", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.ValidCreateRequestWithRating(1)
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create review with rating 5 (maximum)", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.ValidCreateRequestWithRating(5)
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create review with unicode characters", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.CreateRequestWithUnicode()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create review with special characters", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.CreateRequestWithSpecialChars()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertReviewCreated(resp, req, productID)
		})

		t.Run("should create multiple reviews for same product", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			for i := 1; i <= 5; i++ {
				req := reviewFixtures.ValidCreateRequestWithRating(i)
				resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

				review := assertions.AssertReviewCreated(resp, req, productID)
				if review.Id == "" {
					t.Errorf("Review %d should have a valid ID", i)
				}
			}
		})
	})

	t.Run("Validation Errors", func(t *testing.T) {
		t.Run("should return 400 for zero rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.CreateRequestWithZeroRating()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertBadRequestWithMessage(resp, "rating")
		})

		t.Run("should return 400 for negative rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.CreateRequestWithNegativeRating()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertBadRequestWithMessage(resp, "rating")
		})

		t.Run("should return 400 for rating above 5", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			req := reviewFixtures.CreateRequestWithRatingAboveMax()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)

			assertions.AssertBadRequestWithMessage(resp, "rating")
		})
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		t.Run("should return 404 for non-existent product", func(t *testing.T) {
			env.CleanupProducts(t)

			req := reviewFixtures.ValidCreateRequest()
			resp := client.Post("/api/v1/products/99999/reviews", req)

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})

		t.Run("should return 400 for invalid product ID", func(t *testing.T) {
			req := reviewFixtures.ValidCreateRequest()
			resp := client.Post("/api/v1/products/invalid/reviews", req)

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Run("should return 400 for invalid JSON", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			resp := client.PostRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews", productID),
				[]byte("{invalid json}"),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for empty body", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			resp := client.PostRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews", productID),
				[]byte(""),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for wrong type in rating field", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			body := []byte(`{"rating": "not a number"}`)
			resp := client.PostRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews", productID),
				body,
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})
	})
}

func TestCreateProductReviewConcurrency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	// Create a product first
	client := e2e.NewHTTPClient(t, env)
	productFixtures := e2e.NewProductFixtures()
	env.CleanupProducts(t)

	req := productFixtures.ValidCreateRequest()
	resp := client.Post("/api/v1/products", req)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create product")
	product := e2e.ParseJSON[api.Product](t, resp)
	productID := product.Id

	t.Run("should handle concurrent review creation", func(t *testing.T) {
		const numReviews = 10
		results := make(chan *http.Response, numReviews)
		reviewFixtures := e2e.NewReviewFixtures()

		for i := 0; i < numReviews; i++ {
			go func(rating int) {
				client := e2e.NewHTTPClient(t, env)
				req := reviewFixtures.ValidCreateRequestWithRating((rating % 5) + 1)
				resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)
				results <- resp
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numReviews; i++ {
			resp := <-results
			if resp.StatusCode == http.StatusCreated {
				successCount++
			}
			resp.Body.Close()
		}

		if successCount != numReviews {
			t.Errorf("Expected %d successful creations, got %d", numReviews, successCount)
		}
	})
}
