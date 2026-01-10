// Package cache provides caching functionality using Redis.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"product_review_hub/internal/models"

	"github.com/redis/go-redis/v9"
)

const (
	// DefaultTTL is the default cache expiration time.
	DefaultTTL = 5 * time.Minute

	// Key prefixes
	reviewsKeyPrefix = "reviews:product:"
	ratingKeyPrefix  = "rating:product:"
)

// Service provides caching operations.
type Service struct {
	client *redis.Client
	ttl    time.Duration
}

// NewService creates a new cache service.
func NewService(client *redis.Client) *Service {
	return &Service{
		client: client,
		ttl:    DefaultTTL,
	}
}

// NewServiceWithTTL creates a new cache service with custom TTL.
func NewServiceWithTTL(client *redis.Client, ttl time.Duration) *Service {
	return &Service{
		client: client,
		ttl:    ttl,
	}
}

// reviewsKey generates a cache key for reviews list.
func reviewsKey(productID int64, limit, offset int) string {
	return fmt.Sprintf("%s%d:limit:%d:offset:%d", reviewsKeyPrefix, productID, limit, offset)
}

// reviewsPatternKey generates a pattern key for all reviews of a product.
func reviewsPatternKey(productID int64) string {
	return fmt.Sprintf("%s%d:*", reviewsKeyPrefix, productID)
}

// ratingKey generates a cache key for product rating.
func ratingKey(productID int64) string {
	return fmt.Sprintf("%s%d", ratingKeyPrefix, productID)
}

// GetReviews retrieves reviews from cache.
func (s *Service) GetReviews(ctx context.Context, productID int64, limit, offset int) ([]models.Review, error) {
	key := reviewsKey(productID, limit, offset)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get reviews from cache: %w", err)
	}

	var reviews []models.Review
	if err := json.Unmarshal(data, &reviews); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reviews: %w", err)
	}

	return reviews, nil
}

// SetReviews stores reviews in cache.
func (s *Service) SetReviews(ctx context.Context, productID int64, limit, offset int, reviews []models.Review) error {
	key := reviewsKey(productID, limit, offset)
	data, err := json.Marshal(reviews)
	if err != nil {
		return fmt.Errorf("failed to marshal reviews: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set reviews in cache: %w", err)
	}

	return nil
}

// InvalidateReviews removes all cached reviews for a product.
func (s *Service) InvalidateReviews(ctx context.Context, productID int64) error {
	pattern := reviewsPatternKey(productID)
	return s.deleteByPattern(ctx, pattern)
}

// GetRating retrieves product rating from cache.
func (s *Service) GetRating(ctx context.Context, productID int64) (*float64, bool, error) {
	key := ratingKey(productID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get rating from cache: %w", err)
	}

	// Handle null rating (product has no reviews)
	if string(data) == "null" {
		return nil, true, nil
	}

	var rating float64
	if err := json.Unmarshal(data, &rating); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal rating: %w", err)
	}

	return &rating, true, nil
}

// SetRating stores product rating in cache.
func (s *Service) SetRating(ctx context.Context, productID int64, rating *float64) error {
	key := ratingKey(productID)
	data, err := json.Marshal(rating)
	if err != nil {
		return fmt.Errorf("failed to marshal rating: %w", err)
	}

	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set rating in cache: %w", err)
	}

	return nil
}

// InvalidateRating removes cached rating for a product.
func (s *Service) InvalidateRating(ctx context.Context, productID int64) error {
	key := ratingKey(productID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete rating from cache: %w", err)
	}
	return nil
}

// InvalidateProductCache removes all cached data for a product (reviews and rating).
func (s *Service) InvalidateProductCache(ctx context.Context, productID int64) {
	// Invalidate reviews cache
	if err := s.InvalidateReviews(ctx, productID); err != nil {
		log.Printf("Failed to invalidate reviews cache for product %d: %v", productID, err)
	}

	// Invalidate rating cache
	if err := s.InvalidateRating(ctx, productID); err != nil {
		log.Printf("Failed to invalidate rating cache for product %d: %v", productID, err)
	}
}

// deleteByPattern deletes all keys matching the pattern using SCAN.
func (s *Service) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
