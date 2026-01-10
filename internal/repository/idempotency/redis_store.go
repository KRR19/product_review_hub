package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "idempotency:"

// RedisStore implements Store interface using Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore instance.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

// Get retrieves a cached response by idempotency key.
func (s *RedisStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	data, err := s.client.Get(ctx, keyPrefix+key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var resp CachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Set stores a cached response with the given TTL.
func (s *RedisStore) Set(ctx context.Context, key string, resp *CachedResponse, ttl time.Duration) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, keyPrefix+key, data, ttl).Err()
}
