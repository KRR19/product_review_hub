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

// AssertProductUpdated verifies that a product was updated successfully.
func (a *ProductAssertions) AssertProductUpdated(resp *http.Response, expected api.ProductUpdate, productID string) api.Product {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	product := ParseJSON[api.Product](a.t, resp)

	assert.Equal(a.t, productID, product.Id, "Product ID mismatch")
	assert.Equal(a.t, expected.Name, product.Name, "Product name mismatch")
	assert.Equal(a.t, expected.Description, product.Description, "Product description mismatch")
	assert.InDelta(a.t, expected.Price, product.Price, 0.01, "Product price mismatch")

	return product
}

// AssertProductsList verifies the products list response.
func (a *ProductAssertions) AssertProductsList(resp *http.Response, expectedMinCount int) []api.Product {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	products := ParseJSON[[]api.Product](a.t, resp)
	assert.GreaterOrEqual(a.t, len(products), expectedMinCount, "Products count should be at least expected")

	return products
}

// AssertProductsListExact verifies the products list response with exact count.
func (a *ProductAssertions) AssertProductsListExact(resp *http.Response, expectedCount int) []api.Product {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	products := ParseJSON[[]api.Product](a.t, resp)
	assert.Len(a.t, products, expectedCount, "Products count mismatch")

	return products
}

// AssertNotFound verifies that the response is a 404 Not Found.
func (a *ProductAssertions) AssertNotFound(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertNotFoundWithMessage verifies that the response is a 404 Not Found with specific message.
func (a *ProductAssertions) AssertNotFoundWithMessage(resp *http.Response, expectedMessage string) {
	a.t.Helper()

	errorResp := a.AssertNotFound(resp)
	assert.Contains(a.t, errorResp.Error, expectedMessage, "Error message should contain expected text")
}

// AssertNoContent verifies that the response is a 204 No Content.
func (a *ProductAssertions) AssertNoContent(resp *http.Response) {
	a.t.Helper()

	require.Equal(a.t, http.StatusNoContent, resp.StatusCode, "Expected 204 No Content status")
}

// AssertConflict verifies that the response is a 409 Conflict.
func (a *ProductAssertions) AssertConflict(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusConflict, resp.StatusCode, "Expected 409 Conflict status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertConflictWithMessage verifies that the response is a 409 Conflict with specific message.
func (a *ProductAssertions) AssertConflictWithMessage(resp *http.Response, expectedMessage string) {
	a.t.Helper()

	errorResp := a.AssertConflict(resp)
	assert.Contains(a.t, errorResp.Error, expectedMessage, "Error message should contain expected text")
}

// AssertProductByID verifies a single product response by GET.
func (a *ProductAssertions) AssertProductByID(resp *http.Response, expectedID string) api.Product {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	product := ParseJSON[api.Product](a.t, resp)
	assert.Equal(a.t, expectedID, product.Id, "Product ID mismatch")

	return product
}

// ReviewAssertions provides assertion helpers for review-related tests.
type ReviewAssertions struct {
	t *testing.T
}

// NewReviewAssertions creates a new ReviewAssertions instance.
func NewReviewAssertions(t *testing.T) *ReviewAssertions {
	return &ReviewAssertions{t: t}
}

// AssertReviewCreated verifies that a review was created successfully.
func (a *ReviewAssertions) AssertReviewCreated(resp *http.Response, expected api.ReviewCreate, productID string) api.Review {
	a.t.Helper()

	require.Equal(a.t, http.StatusCreated, resp.StatusCode, "Expected 201 Created status")

	review := ParseJSON[api.Review](a.t, resp)

	assert.NotEmpty(a.t, review.Id, "Review ID should not be empty")
	assert.Equal(a.t, productID, review.ProductId, "Product ID mismatch")
	assert.Equal(a.t, expected.Rating, review.Rating, "Rating mismatch")

	if expected.Author != nil {
		require.NotNil(a.t, review.Author, "Author should not be nil")
		assert.Equal(a.t, *expected.Author, *review.Author, "Author mismatch")
	}

	if expected.Comment != nil {
		require.NotNil(a.t, review.Comment, "Comment should not be nil")
		assert.Equal(a.t, *expected.Comment, *review.Comment, "Comment mismatch")
	}

	return review
}

// AssertReviewUpdated verifies that a review was updated successfully.
func (a *ReviewAssertions) AssertReviewUpdated(resp *http.Response, expected api.ReviewUpdate, productID, reviewID string) api.Review {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	review := ParseJSON[api.Review](a.t, resp)

	assert.Equal(a.t, reviewID, review.Id, "Review ID mismatch")
	assert.Equal(a.t, productID, review.ProductId, "Product ID mismatch")
	assert.Equal(a.t, expected.Rating, review.Rating, "Rating mismatch")

	if expected.Author != nil {
		require.NotNil(a.t, review.Author, "Author should not be nil")
		assert.Equal(a.t, *expected.Author, *review.Author, "Author mismatch")
	}

	if expected.Comment != nil {
		require.NotNil(a.t, review.Comment, "Comment should not be nil")
		assert.Equal(a.t, *expected.Comment, *review.Comment, "Comment mismatch")
	}

	return review
}

// AssertReviewsList verifies the reviews list response.
func (a *ReviewAssertions) AssertReviewsList(resp *http.Response, expectedCount int) []api.Review {
	a.t.Helper()

	require.Equal(a.t, http.StatusOK, resp.StatusCode, "Expected 200 OK status")

	reviews := ParseJSON[[]api.Review](a.t, resp)
	assert.Len(a.t, reviews, expectedCount, "Reviews count mismatch")

	return reviews
}

// AssertBadRequest verifies that the response is a 400 Bad Request.
func (a *ReviewAssertions) AssertBadRequest(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusBadRequest, resp.StatusCode, "Expected 400 Bad Request status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertBadRequestWithMessage verifies that the response is a 400 Bad Request with specific message.
func (a *ReviewAssertions) AssertBadRequestWithMessage(resp *http.Response, expectedMessage string) {
	a.t.Helper()

	errorResp := a.AssertBadRequest(resp)
	assert.Contains(a.t, errorResp.Error, expectedMessage, "Error message should contain expected text")
}

// AssertNotFound verifies that the response is a 404 Not Found.
func (a *ReviewAssertions) AssertNotFound(resp *http.Response) api.ErrorResponse {
	a.t.Helper()

	require.Equal(a.t, http.StatusNotFound, resp.StatusCode, "Expected 404 Not Found status")

	errorResp := ParseJSON[api.ErrorResponse](a.t, resp)
	assert.NotEmpty(a.t, errorResp.Error, "Error message should not be empty")

	return errorResp
}

// AssertNotFoundWithMessage verifies that the response is a 404 Not Found with specific message.
func (a *ReviewAssertions) AssertNotFoundWithMessage(resp *http.Response, expectedMessage string) {
	a.t.Helper()

	errorResp := a.AssertNotFound(resp)
	assert.Contains(a.t, errorResp.Error, expectedMessage, "Error message should contain expected text")
}

// AssertNoContent verifies that the response is a 204 No Content.
func (a *ReviewAssertions) AssertNoContent(resp *http.Response) {
	a.t.Helper()

	require.Equal(a.t, http.StatusNoContent, resp.StatusCode, "Expected 204 No Content status")
}

// AssertContentTypeJSON verifies that the response has JSON content type.
func (a *ReviewAssertions) AssertContentTypeJSON(resp *http.Response) {
	a.t.Helper()

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(a.t, contentType, "application/json", "Content-Type should be application/json")
}

// AssertStatusCode verifies that the response has the expected status code.
func (a *ReviewAssertions) AssertStatusCode(resp *http.Response, expected int) {
	a.t.Helper()

	require.Equal(a.t, expected, resp.StatusCode, "Unexpected status code")
}
