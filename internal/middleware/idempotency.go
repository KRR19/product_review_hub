package middleware

import (
	"bytes"
	"net/http"
	"time"

	"product_review_hub/internal/repository/idempotency"
)

// IdempotencyKeyHeader is the HTTP header name for idempotency key.
const IdempotencyKeyHeader = "X-Idempotency-Key"

// mutatingMethods contains HTTP methods that should be checked for idempotency.
var mutatingMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodDelete: true,
}

// responseRecorder captures the response for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	headers    http.Header
}

// newResponseRecorder creates a new responseRecorder wrapping the given ResponseWriter.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
		headers:        make(http.Header),
	}
}

// WriteHeader captures the status code and writes it to the underlying ResponseWriter.
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the body and writes it to the underlying ResponseWriter.
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Header returns the header map.
func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// toCachedResponse converts the recorded response to CachedResponse.
func (r *responseRecorder) toCachedResponse() *idempotency.CachedResponse {
	headers := make(map[string]string)
	for k, v := range r.ResponseWriter.Header() {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &idempotency.CachedResponse{
		StatusCode: r.statusCode,
		Headers:    headers,
		Body:       r.body.Bytes(),
	}
}

// isMutatingMethod checks if the HTTP method should be checked for idempotency.
func isMutatingMethod(method string) bool {
	return mutatingMethods[method]
}

// writeCachedResponse writes a cached response to the ResponseWriter.
func writeCachedResponse(w http.ResponseWriter, cached *idempotency.CachedResponse) {
	for k, v := range cached.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(cached.StatusCode)
	w.Write(cached.Body)
}

// Idempotency returns a middleware that checks for idempotency keys and caches responses.
// It only processes POST, PUT, and DELETE requests that include the X-Idempotency-Key header.
func Idempotency(store idempotency.Store, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip non-mutating methods
			if !isMutatingMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Get idempotency key from header
			key := r.Header.Get(IdempotencyKeyHeader)
			if key == "" {
				// No idempotency key provided, proceed without caching
				next.ServeHTTP(w, r)
				return
			}

			// Check for cached response
			cached, err := store.Get(r.Context(), key)
			if err == nil && cached != nil {
				// Return cached response
				writeCachedResponse(w, cached)
				return
			}

			// Execute handler and capture response
			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			// Cache successful responses (2xx status codes)
			if rec.statusCode >= 200 && rec.statusCode < 300 {
				_ = store.Set(r.Context(), key, rec.toCachedResponse(), ttl)
			}
		})
	}
}
