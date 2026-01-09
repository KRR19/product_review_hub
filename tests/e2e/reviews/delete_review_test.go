package reviews_test

import (
	"fmt"
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/require"
)

func TestDeleteProductReview(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	reviewFixtures := e2e.NewReviewFixtures()
	assertions := e2e.NewReviewAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should delete review successfully", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			resp := client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))

			assertions.AssertNoContent(resp)
			resp.Body.Close()
		})

		t.Run("should remove review from list after deletion", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			// Verify review exists
			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))
			reviews := e2e.ParseJSON[[]api.Review](t, resp)
			require.Len(t, reviews, 1, "Expected 1 review")

			// Delete the review
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))
			assertions.AssertNoContent(resp)
			resp.Body.Close()

			// Verify review is gone
			resp = client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))
			reviews = e2e.ParseJSON[[]api.Review](t, resp)
			if len(reviews) != 0 {
				t.Errorf("Expected 0 reviews after deletion, got %d", len(reviews))
			}
		})

		t.Run("should only delete specified review", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create 3 reviews
			review1 := e2e.CreateTestReview(t, client, productID)
			e2e.CreateTestReview(t, client, productID)
			e2e.CreateTestReview(t, client, productID)

			// Verify all 3 reviews exist
			resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))
			reviews := e2e.ParseJSON[[]api.Review](t, resp)
			require.Len(t, reviews, 3, "Expected 3 reviews")

			// Delete the first review
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review1.Id))
			assertions.AssertNoContent(resp)
			resp.Body.Close()

			// Verify only 2 reviews remain
			resp = client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", productID))
			reviews = e2e.ParseJSON[[]api.Review](t, resp)
			if len(reviews) != 2 {
				t.Errorf("Expected 2 reviews after deletion, got %d", len(reviews))
			}

			// Verify deleted review is not in the list
			for _, r := range reviews {
				if r.Id == review1.Id {
					t.Errorf("Deleted review %s still in list", review1.Id)
				}
			}
		})
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		t.Run("should return 404 for non-existent review", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			resp := client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/99999", productID))

			assertions.AssertNotFoundWithMessage(resp, "Review not found")
		})

		t.Run("should return 404 for already deleted review", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)
			review := e2e.CreateTestReview(t, client, productID)

			// Delete the review
			resp := client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))
			assertions.AssertNoContent(resp)
			resp.Body.Close()

			// Try to delete again
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))
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

			// Try to delete review using wrong product ID
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", anotherProduct.Id, review.Id))

			assertions.AssertNotFoundWithMessage(resp, "Review not found")
		})

		t.Run("should return 400 for invalid product ID", func(t *testing.T) {
			resp := client.Delete("/api/v1/products/invalid/reviews/1")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for invalid review ID", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			resp := client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/invalid", productID))

			assertions.AssertBadRequestWithMessage(resp, "Invalid review ID")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("should handle deletion of review with special characters in data", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create review with special characters
			req := reviewFixtures.CreateRequestWithSpecialChars()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)
			review := e2e.ParseJSON[api.Review](t, resp)

			// Delete the review
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))

			assertions.AssertNoContent(resp)
			resp.Body.Close()
		})

		t.Run("should handle deletion of review with unicode data", func(t *testing.T) {
			productID := e2e.CreateTestProduct(t, env, client)

			// Create review with unicode
			req := reviewFixtures.CreateRequestWithUnicode()
			resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", productID), req)
			review := e2e.ParseJSON[api.Review](t, resp)

			// Delete the review
			resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", productID, review.Id))

			assertions.AssertNoContent(resp)
			resp.Body.Close()
		})
	})
}

func TestDeleteProductReviewConcurrency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	productFixtures := e2e.NewProductFixtures()
	reviewFixtures := e2e.NewReviewFixtures()
	env.CleanupProducts(t)

	// Create a product
	req := productFixtures.ValidCreateRequest()
	resp := client.Post("/api/v1/products", req)
	product := e2e.ParseJSON[api.Product](t, resp)

	// Create multiple reviews
	const numReviews = 10
	reviewIDs := make([]string, numReviews)
	for i := 0; i < numReviews; i++ {
		req := reviewFixtures.ValidCreateRequestWithRating((i % 5) + 1)
		resp := client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id), req)
		review := e2e.ParseJSON[api.Review](t, resp)
		reviewIDs[i] = review.Id
	}

	t.Run("should handle concurrent deletions", func(t *testing.T) {
		results := make(chan *http.Response, numReviews)

		for _, reviewID := range reviewIDs {
			go func(id string) {
				client := e2e.NewHTTPClient(t, env)
				resp := client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", product.Id, id))
				results <- resp
			}(reviewID)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numReviews; i++ {
			resp := <-results
			if resp.StatusCode == http.StatusNoContent {
				successCount++
			}
			resp.Body.Close()
		}

		if successCount != numReviews {
			t.Errorf("Expected %d successful deletions, got %d", numReviews, successCount)
		}

		// Verify all reviews are deleted
		resp := client.Get(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id))
		reviews := e2e.ParseJSON[[]api.Review](t, resp)
		if len(reviews) != 0 {
			t.Errorf("Expected 0 reviews after deletions, got %d", len(reviews))
		}
	})
}

func TestDeleteReviewIdempotency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	productFixtures := e2e.NewProductFixtures()
	reviewFixtures := e2e.NewReviewFixtures()
	assertions := e2e.NewReviewAssertions(t)

	t.Run("should return 404 on repeated delete attempts", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product and review
		req := productFixtures.ValidCreateRequest()
		resp := client.Post("/api/v1/products", req)
		product := e2e.ParseJSON[api.Product](t, resp)

		reviewReq := reviewFixtures.ValidCreateRequest()
		resp = client.Post(fmt.Sprintf("/api/v1/products/%s/reviews", product.Id), reviewReq)
		review := e2e.ParseJSON[api.Review](t, resp)

		// First delete should succeed
		resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", product.Id, review.Id))
		assertions.AssertNoContent(resp)
		resp.Body.Close()

		// Second delete should return 404
		resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", product.Id, review.Id))
		assertions.AssertNotFoundWithMessage(resp, "Review not found")

		// Third delete should also return 404
		resp = client.Delete(fmt.Sprintf("/api/v1/products/%s/reviews/%s", product.Id, review.Id))
		assertions.AssertNotFoundWithMessage(resp, "Review not found")
	})
}
