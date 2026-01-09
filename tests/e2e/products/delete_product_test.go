package products_test

import (
	"net/http"
	"testing"

	"product_review_hub/tests/e2e"

	"github.com/stretchr/testify/require"
)

func TestDeleteProduct(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should delete product without reviews", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Delete product
			resp := client.Delete(productsEndpoint + "/" + createdProduct.Id)

			assertions.AssertNoContent(resp)
		})

		t.Run("should verify product is actually deleted", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Delete product
			deleteResp := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			assertions.AssertNoContent(deleteResp)

			// Try to get deleted product
			getResp := client.Get(productsEndpoint + "/" + createdProduct.Id)
			assertions.AssertNotFoundWithMessage(getResp, "Product not found")
		})

		t.Run("should not appear in product list after deletion", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Verify product is in list
			listResp1 := client.Get(productsEndpoint)
			products1 := assertions.AssertProductsListExact(listResp1, 1)
			require.Equal(t, createdProduct.Id, products1[0].Id)

			// Delete product
			deleteResp := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			assertions.AssertNoContent(deleteResp)

			// Verify product is not in list
			listResp2 := client.Get(productsEndpoint)
			assertions.AssertProductsListExact(listResp2, 0)
		})
	})

	t.Run("Conflict", func(t *testing.T) {
		t.Run("should return 409 when product has reviews", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create review
			e2e.CreateTestReview(t, client, productID)

			// Try to delete product with reviews
			resp := client.Delete(productsEndpoint + "/" + productID)

			assertions.AssertConflictWithMessage(resp, "Cannot delete product with existing reviews")
		})

		t.Run("should return 409 when product has multiple reviews", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create multiple reviews
			e2e.CreateTestReviewWithRating(t, client, productID, 5)
			e2e.CreateTestReviewWithRating(t, client, productID, 4)
			e2e.CreateTestReviewWithRating(t, client, productID, 3)

			// Try to delete product with reviews
			resp := client.Delete(productsEndpoint + "/" + productID)

			assertions.AssertConflictWithMessage(resp, "Cannot delete product with existing reviews")
		})

		t.Run("should allow deletion after all reviews are deleted", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			productID := e2e.CreateTestProduct(t, env, client)

			// Create review
			review := e2e.CreateTestReview(t, client, productID)

			// Try to delete product (should fail)
			failResp := client.Delete(productsEndpoint + "/" + productID)
			assertions.AssertConflictWithMessage(failResp, "Cannot delete product with existing reviews")

			// Delete review
			deleteReviewResp := client.Delete(productsEndpoint + "/" + productID + "/reviews/" + review.Id)
			require.Equal(t, http.StatusNoContent, deleteReviewResp.StatusCode)

			// Now delete product (should succeed)
			successResp := client.Delete(productsEndpoint + "/" + productID)
			assertions.AssertNoContent(successResp)
		})
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Run("should return 404 for non-existent product", func(t *testing.T) {
			env.CleanupProducts(t)

			resp := client.Delete(productsEndpoint + "/999999")

			assertions.AssertNotFoundWithMessage(resp, "Product not found")
		})

		t.Run("should return 404 for already deleted product", func(t *testing.T) {
			env.CleanupProducts(t)

			// Create product
			createReq := fixtures.ValidCreateRequest()
			createResp := client.Post(productsEndpoint, createReq)
			createdProduct := assertions.AssertProductCreated(createResp, createReq)

			// Delete product
			deleteResp1 := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			assertions.AssertNoContent(deleteResp1)

			// Try to delete again
			deleteResp2 := client.Delete(productsEndpoint + "/" + createdProduct.Id)
			assertions.AssertNotFoundWithMessage(deleteResp2, "Product not found")
		})
	})

	t.Run("Invalid ID", func(t *testing.T) {
		t.Run("should return 400 for non-numeric ID", func(t *testing.T) {
			resp := client.Delete(productsEndpoint + "/invalid")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for floating point ID", func(t *testing.T) {
			resp := client.Delete(productsEndpoint + "/1.5")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})

		t.Run("should return 400 for special characters in ID", func(t *testing.T) {
			resp := client.Delete(productsEndpoint + "/abc!@#")

			assertions.AssertBadRequestWithMessage(resp, "Invalid product ID")
		})
	})
}

func TestDeleteProductConcurrency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("should handle concurrent delete attempts", func(t *testing.T) {
		env.CleanupProducts(t)

		// Create product
		createReq := fixtures.ValidCreateRequest()
		createResp := client.Post(productsEndpoint, createReq)
		createdProduct := assertions.AssertProductCreated(createResp, createReq)

		const numAttempts = 5
		results := make(chan int, numAttempts)

		// Try to delete the same product concurrently
		for i := 0; i < numAttempts; i++ {
			go func() {
				c := e2e.NewHTTPClient(t, env)
				resp := c.Delete(productsEndpoint + "/" + createdProduct.Id)
				results <- resp.StatusCode
				resp.Body.Close()
			}()
		}

		// Collect results
		successCount := 0
		notFoundCount := 0
		for i := 0; i < numAttempts; i++ {
			status := <-results
			switch status {
			case http.StatusNoContent:
				successCount++
			case http.StatusNotFound:
				notFoundCount++
			}
		}

		// Exactly one should succeed, others should get 404
		require.Equal(t, 1, successCount, "Exactly one delete should succeed")
		require.Equal(t, numAttempts-1, notFoundCount, "Other deletes should get 404")
	})
}
