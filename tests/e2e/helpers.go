package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"product_review_hub/internal/api"
)

// RequestOption is a function that modifies an HTTP request.
type RequestOption func(*http.Request)

// WithHeader adds a header to the request.
func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

// WithContentType sets the Content-Type header.
func WithContentType(contentType string) RequestOption {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", contentType)
	}
}

// HTTPClient provides helper methods for making HTTP requests in tests.
type HTTPClient struct {
	t       *testing.T
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new test HTTP client.
func NewHTTPClient(t *testing.T, env *TestEnv) *HTTPClient {
	return &HTTPClient{
		t:       t,
		client:  env.Client,
		baseURL: env.BaseURL,
	}
}

// Post sends a POST request with JSON body.
func (c *HTTPClient) Post(path string, body interface{}, opts ...RequestOption) *http.Response {
	c.t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(c.t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewBuffer(jsonBody))
	require.NoError(c.t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// PostRaw sends a POST request with raw body.
func (c *HTTPClient) PostRaw(path string, body []byte, opts ...RequestOption) *http.Response {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewBuffer(body))
	require.NoError(c.t, err, "Failed to create request")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// Get sends a GET request.
func (c *HTTPClient) Get(path string, opts ...RequestOption) *http.Response {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	require.NoError(c.t, err, "Failed to create request")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// Put sends a PUT request with JSON body.
func (c *HTTPClient) Put(path string, body interface{}, opts ...RequestOption) *http.Response {
	c.t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(c.t, err, "Failed to marshal request body")

	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, bytes.NewBuffer(jsonBody))
	require.NoError(c.t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// PutRaw sends a PUT request with raw body.
func (c *HTTPClient) PutRaw(path string, body []byte, opts ...RequestOption) *http.Response {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, bytes.NewBuffer(body))
	require.NoError(c.t, err, "Failed to create request")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// Delete sends a DELETE request.
func (c *HTTPClient) Delete(path string, opts ...RequestOption) *http.Response {
	c.t.Helper()

	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	require.NoError(c.t, err, "Failed to create request")

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "Failed to send request")

	return resp
}

// ParseJSON parses the response body as JSON into the provided struct.
func ParseJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result T
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse JSON response: %s", string(body))

	return result
}

// ReadBody reads and returns the response body as string.
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	return string(body)
}

// CreateTestProduct creates a product and returns its ID.
// This is a common helper for review tests.
func CreateTestProduct(t *testing.T, env *TestEnv, client *HTTPClient) string {
	t.Helper()
	env.CleanupProducts(t)
	productFixtures := NewProductFixtures()
	req := productFixtures.ValidCreateRequest()
	resp := client.Post("/api/v1/products", req)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create product")
	product := ParseJSON[api.Product](t, resp)
	return product.Id
}

// createTestReviewWithRequest is an internal helper that creates a review with the provided request.
func createTestReviewWithRequest(t *testing.T, client *HTTPClient, productID string, req interface{}) api.Review {
	t.Helper()
	resp := client.Post("/api/v1/products/"+productID+"/reviews", req)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create review")
	return ParseJSON[api.Review](t, resp)
}

// CreateTestReview creates a review with full data and returns it.
func CreateTestReview(t *testing.T, client *HTTPClient, productID string) api.Review {
	t.Helper()
	reviewFixtures := NewReviewFixtures()
	req := reviewFixtures.ValidCreateRequestFull()
	return createTestReviewWithRequest(t, client, productID, req)
}

// CreateTestReviewWithRating creates a review with specified rating and returns it.
func CreateTestReviewWithRating(t *testing.T, client *HTTPClient, productID string, rating int) api.Review {
	t.Helper()
	reviewFixtures := NewReviewFixtures()
	req := reviewFixtures.ValidCreateRequestWithRating(rating)
	return createTestReviewWithRequest(t, client, productID, req)
}

// ProductResponse is an alias for api.Product for use in tests.
type ProductResponse = api.Product

// CreateTestProductWithoutCleanup creates a product without cleaning up existing products.
// Useful when you need to create multiple products.
func CreateTestProductWithoutCleanup(t *testing.T, client *HTTPClient) string {
	t.Helper()
	productFixtures := NewProductFixtures()
	req := productFixtures.ValidCreateRequest()
	resp := client.Post("/api/v1/products", req)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create product")
	product := ParseJSON[api.Product](t, resp)
	return product.Id
}
