package idempotency

import (
	"context"
	"time"
)

// CachedResponse represents a cached HTTP response.
type CachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// Store defines the interface for idempotency key storage.
type Store interface {
	// Get retrieves a cached response by idempotency key.
	// Returns nil, nil if key does not exist.
	Get(ctx context.Context, key string) (*CachedResponse, error)

	// Set stores a cached response with the given TTL.
	Set(ctx context.Context, key string, resp *CachedResponse, ttl time.Duration) error
}
