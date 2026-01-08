package products_test

import (
	"net/http"
	"testing"

	"product_review_hub/internal/api"
	"product_review_hub/tests/e2e"
)

const productsEndpoint = "/api/v1/products"

func TestCreateProduct(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("Success", func(t *testing.T) {
		t.Run("should create product with valid data", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.ValidCreateRequest()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertContentTypeJSON(resp)
			product := assertions.AssertProductCreated(resp, req)

			// Verify product ID is a valid numeric string
			if product.Id == "" || product.Id == "0" {
				t.Error("Product ID should be a valid non-zero value")
			}
		})

		t.Run("should create product with minimum valid price", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.CreateRequestWithMinPrice()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertProductCreated(resp, req)
		})

		t.Run("should create product with large price", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.CreateRequestWithMaxPrice()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertProductCreated(resp, req)
		})

		t.Run("should create product with special characters in name", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.CreateRequestWithSpecialChars()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertProductCreated(resp, req)
		})

		t.Run("should create product with unicode characters", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.CreateRequestWithUnicode()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertProductCreated(resp, req)
		})

		t.Run("should create multiple products", func(t *testing.T) {
			env.CleanupProducts(t)

			productNames := []string{"Product 1", "Product 2", "Product 3"}
			createdProducts := make([]api.Product, 0, len(productNames))

			for _, name := range productNames {
				req := fixtures.ValidCreateRequestWithName(name)
				resp := client.Post(productsEndpoint, req)

				product := assertions.AssertProductCreated(resp, req)
				createdProducts = append(createdProducts, product)
			}

			// Verify all products have unique IDs
			ids := make(map[string]bool)
			for _, p := range createdProducts {
				if ids[p.Id] {
					t.Errorf("Duplicate product ID found: %s", p.Id)
				}
				ids[p.Id] = true
			}
		})
	})

	t.Run("Validation Errors", func(t *testing.T) {
		t.Run("should return 400 for empty name", func(t *testing.T) {
			req := fixtures.CreateRequestWithEmptyName()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertBadRequestWithMessage(resp, "name")
		})

		t.Run("should return 400 for zero price", func(t *testing.T) {
			req := fixtures.CreateRequestWithZeroPrice()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertBadRequestWithMessage(resp, "price")
		})

		t.Run("should return 400 for negative price", func(t *testing.T) {
			req := fixtures.CreateRequestWithNegativePrice()
			resp := client.Post(productsEndpoint, req)

			assertions.AssertBadRequestWithMessage(resp, "price")
		})
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Run("should return 400 for invalid JSON", func(t *testing.T) {
			resp := client.PostRaw(productsEndpoint, []byte("{invalid json}"), e2e.WithContentType("application/json"))

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for empty body", func(t *testing.T) {
			resp := client.PostRaw(productsEndpoint, []byte(""), e2e.WithContentType("application/json"))

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for null body", func(t *testing.T) {
			resp := client.PostRaw(productsEndpoint, []byte("null"), e2e.WithContentType("application/json"))

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for wrong type in price field", func(t *testing.T) {
			body := []byte(`{"name": "Test", "description": "Test", "price": "not a number"}`)
			resp := client.PostRaw(productsEndpoint, body, e2e.WithContentType("application/json"))

			assertions.AssertBadRequest(resp)
		})

		t.Run("should return 400 for missing required fields", func(t *testing.T) {
			body := []byte(`{"description": "Only description"}`)
			resp := client.PostRaw(productsEndpoint, body, e2e.WithContentType("application/json"))

			assertions.AssertBadRequest(resp)
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("should handle very long product name", func(t *testing.T) {
			env.CleanupProducts(t)

			req := fixtures.CreateRequestWithLongName()
			resp := client.Post(productsEndpoint, req)

			// Depending on database constraints, this might succeed or fail
			// We just verify we get a valid response
			if resp.StatusCode == http.StatusCreated {
				assertions.AssertContentTypeJSON(resp)
			} else if resp.StatusCode == http.StatusBadRequest {
				assertions.AssertBadRequest(resp)
			} else if resp.StatusCode == http.StatusInternalServerError {
				// Database constraint violation
				assertions.AssertInternalServerError(resp)
			} else {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
		})

		t.Run("should handle product with empty description", func(t *testing.T) {
			env.CleanupProducts(t)

			req := api.ProductCreate{
				Name:        "Product without description",
				Description: "",
				Price:       50.00,
			}
			resp := client.Post(productsEndpoint, req)

			// Empty description is valid per the API spec
			assertions.AssertProductCreated(resp, req)
		})

		t.Run("should handle float precision in price", func(t *testing.T) {
			env.CleanupProducts(t)

			req := api.ProductCreate{
				Name:        "Precision Test Product",
				Description: "Testing float precision",
				Price:       123.456789,
			}
			resp := client.Post(productsEndpoint, req)

			if resp.StatusCode == http.StatusCreated {
				product := e2e.ParseJSON[api.Product](t, resp)
				// Price should be stored with reasonable precision
				if product.Price < 123.0 || product.Price > 124.0 {
					t.Errorf("Price was not preserved correctly: got %f", product.Price)
				}
			}
		})
	})
}

func TestCreateProductIdempotency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	client := e2e.NewHTTPClient(t, env)
	fixtures := e2e.NewProductFixtures()
	assertions := e2e.NewProductAssertions(t)

	t.Run("should create products with same name as separate entities", func(t *testing.T) {
		env.CleanupProducts(t)

		req := fixtures.ValidCreateRequest()

		// Create first product
		resp1 := client.Post(productsEndpoint, req)
		product1 := assertions.AssertProductCreated(resp1, req)

		// Create second product with same data
		resp2 := client.Post(productsEndpoint, req)
		product2 := assertions.AssertProductCreated(resp2, req)

		// They should have different IDs
		if product1.Id == product2.Id {
			t.Error("Two products with same name should have different IDs")
		}
	})
}

func TestCreateProductConcurrency(t *testing.T) {
	env := e2e.Setup(t)
	defer env.Teardown(t)

	t.Run("should handle concurrent product creation", func(t *testing.T) {
		env.CleanupProducts(t)

		const numProducts = 10
		results := make(chan *http.Response, numProducts)
		fixtures := e2e.NewProductFixtures()

		for i := 0; i < numProducts; i++ {
			go func(index int) {
				client := e2e.NewHTTPClient(t, env)
				req := fixtures.ValidCreateRequestWithPrice(float32(100 + index))
				resp := client.Post(productsEndpoint, req)
				results <- resp
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numProducts; i++ {
			resp := <-results
			if resp.StatusCode == http.StatusCreated {
				successCount++
			}
			resp.Body.Close()
		}

		if successCount != numProducts {
			t.Errorf("Expected %d successful creations, got %d", numProducts, successCount)
		}
	})
}
