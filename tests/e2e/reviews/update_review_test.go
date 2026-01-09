package reviews_test

import (
	"fmt"
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/require"
)

func TestUpdateProductReview(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	reviewFixtures := e2e.NewReviewFixtures()
	assertions := e2e.NewReviewAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should update review with new rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := reviewFixtures.ValidUpdateRequest()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertContentTypeJSON(resp)
			assertions.AssertReviewUpdated(resp, updateReq, productID, review.Id)
		})

		t.Run("should update review with all fields", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := reviewFixtures.ValidUpdateRequestFull()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertReviewUpdated(resp, updateReq, productID, review.Id)
		})

		t.Run("should update to minimum rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := api.ReviewUpdate{Rating: 1}
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertReviewUpdated(resp, updateReq, productID, review.Id)
		})

		t.Run("should update to maximum rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := api.ReviewUpdate{Rating: 5}
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertReviewUpdated(resp, updateReq, productID, review.Id)
		})

		t.Run("should preserve review ID after update", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			originalReview := e2e.CreateTestReview(t, client, productID)

			updateReq := reviewFixtures.ValidUpdateRequestFull()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, originalReview.Id),
				updateReq,
			)

			require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

			updatedReview := e2e.ParseJSON[api.Review](t, resp)
			if updatedReview.Id != originalReview.Id {
				t.Errorf("Review ID changed: expected %s, got %s", originalReview.Id, updatedReview.Id)
			}
		})
	})

	t.Run("Validation Errors", func(t *testing.T) {
		t.Run("should return 400 for zero rating", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := reviewFixtures.UpdateRequestWithZeroRating()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertBadRequestWithMessage(resp, "rating")
		})

		t.Run("should return 400 for rating above 5", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			updateReq := reviewFixtures.UpdateRequestWithRatingAboveMax()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				updateReq,
			)

			assertions.AssertBadRequestWithMessage(resp, "rating")
		})
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		t.Run("should return 404 for non-existent review", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			updateReq := reviewFixtures.ValidUpdateRequest()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/99999", productID),
				updateReq,
			)

			assertions.AssertNotFoundWithMessage(resp, "Review not found")
		})

		t.Run("should return 404 for review on wrong product", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			// Create another product
			productFixtures := e2e.NewProductFixtures()
			req := productFixtures.ValidCreateRequestWithName("Another Product")
			resp := client.Post("/api/v1/products", req)
			anotherProduct := e2e.ParseJSON[api.Product](t, resp)

			// Try to update review using wrong product ID
			updateReq := reviewFixtures.ValidUpdateRequest()
			resp = client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", anotherProduct.Id, review.Id),
				updateReq,
			)

			assertions.AssertNotFoundWithMessage(resp, "Review not found")
		})

		t.Run("should return 400 for invalid product ID", func(t *testing.T) {
			updateReq := reviewFixtures.ValidUpdateRequest()
			resp := client.Put("/api/v1/products/invalid/reviews/1", updateReq)

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for invalid review ID", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			updateReq := reviewFixtures.ValidUpdateRequest()
			resp := client.Put(
				fmt.Sprintf("/api/v1/products/%s/reviews/invalid", productID),
				updateReq,
			)

			assertions.AssertBadRequestWithMessage(resp, "Invalid review ID")
		})
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Run("should return 400 for invalid JSON", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			resp := client.PutRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				[]byte("{invalid json}"),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for empty body", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			resp := client.PutRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				[]byte(""),
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for wrong type in rating field", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			body := []byte(`{"rating": "not a number"}`)
			resp := client.PutRaw(
				fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id),
				body,
				e2e.WithContentType("application/json"),
			)

			assertions.AssertBadRequest(resp)
		})
	})
}

func TestUpdateProductReviewConcurrency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	// Create a product and review first
	client := e2e.NewHTTPClient(t, env)
	productFixtures := e2e.NewProductFixtures()
	reviewFixtures := e2e.NewReviewFixtures()
	env.CleanupProducts(t)

	req := productFixtures.ValidCreateRequest()
	resp := client.Post("/api/v1/products", req)
	product := e2e.ParseJSON[api.Product](t, resp)

	reviewReq := reviewFixtures.ValidCreateRequestFull()
	resp = client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id), reviewReq)
	review := e2e.ParseJSON[api.Review](t, resp)

	t.Run("should handle concurrent updates", func(t *testing.T) {
		const numUpdates = 10
		results := make(chan *http.Response, numUpdates)

		for i := 0; i < numUpdates; i++ {
			go func(rating int) {
				client := e2e.NewHTTPClient(t, env)
				updateReq := api.ReviewUpdate{Rating: (rating % 5) + 1}
				resp := client.Put(
					fmt.Sprintf("/api/v1/products/%s/reviews/%s", product.Id, review.Id),
					updateReq,
				)
				results <- resp
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numUpdates; i++ {
			resp := <-results
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
			resp.Body.Close()
		}

		if successCount != numUpdates {
			t.Errorf("Expected %d successful updates, got %d", numUpdates, successCount)
		}
	})
}
