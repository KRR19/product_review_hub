package e2e

import (
	"net/http"
	"testing"

	"product_review_hub/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ProductAssertions provides assertion helpers for product-related tests.
type ProductAssertions struct {
	t *testing.T
}

// NewProductAssertions creates a new ProductAssertions instance.
func NewProductAssertions(t *testing.T) *ProductAssertions {
	return &ProductAssertions{t: t}
}

// AssertProductCreated verifies that a product was created successfully.
func (a *ProductAssertions) AssertProductCreated(resp *http.Response, expected api.ProductCreate) api.Product {
	a.t.Helper()

	require.Equal(a.t, http.StatusCreated, resp.StatusCode, "Expected 201 Created status")

	product := ParseJSON[api.Product](a.t, resp)

	assert.NotEmpty(a.t, product.Id, "Product ID should not be empty")
	assert.Equal(a.t, expected.Name, product.Name, "Product name mismatch")
	assert.Equal(a.t, expected.Description, product.Description, "Product description mismatch")
	assert.InDelta(a.t, expected.Price, product.Price, 0.01, "Product price mismatch")

	return product
}

// AssertBadRequest verifies that the response is a 400 Bad Request.
func (a *ProductAssertions) AssertBadRequest(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusBadRequest, resp.StatusCode, "Expected 400 Bad Request status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertBadRequestWithMessage verifies that the response is a 400 Bad Request with specific message.
func (a *ProductAssertions) AssertBadRequestWithMessage(resp *http.Response, expectedMessage string) {
	a.t.Helper()

	errorResp := a.AssertBadRequest(resp)
	assert.Contains(a.t, errorResp.Error, expectedMessage, "Error message should contain expected text")
}

// AssertInternalServerError verifies that the response is a 500 Internal Server Error.
func (a *ProductAssertions) AssertInternalServerError(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusInternalServerError, resp.StatusCode, "Expected 500 Internal Server Error status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertContentTypeJSON verifies that the response has JSON content type.
func (a *ProductAssertions) AssertContentTypeJSON(resp *http.Response) {
	a.t.Helper()

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(a.t, contentType, "application/json", "Content-Type should be application/json")
}

// AssertStatusCode verifies that the response has the expected status code.
func (a *ProductAssertions) AssertStatusCode(resp *http.Response, expected int) {
	a.t.Helper()

	require.Equal(a.t, expected, resp.StatusCode, "Unexpected status code")
}
