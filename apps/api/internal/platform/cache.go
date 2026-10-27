package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// CachePlatform defines the caching contract with anti-stampede and pattern invalidation.
type CachePlatform interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Remember(ctx context.Context, key string, ttl time.Duration, dest interface{}, fetchFunc func(ctx context.Context) (interface{}, error)) error
}

type RedisCache struct {
	client  *redis.Client
	logger  *slog.Logger
	sfGroup singleflight.Group
}

// NewRedisCache initializes a RedisCache instance.
func NewRedisCache(client *redis.Client, logger *slog.Logger) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
	}
}

// Get retrieves and deserializes cached JSON data. Returns (true, nil) on hit, (false, nil) on miss.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if c.client == nil {
		return false, nil
	}

	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		c.logger.Warn("redis cache read error; continuing without cache",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return false, nil
	}

	if err := json.Unmarshal(val, dest); err != nil {
		c.logger.Error("failed to unmarshal cached data",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return false, err
	}

	return true, nil
}

// Set serializes and stores value into Redis with random jitter added to TTL to avoid thundering herd expirations.
func (c *RedisCache) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	if c.client == nil {
		return nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// Apply 10% jitter to TTL
	jitteredTTL := addJitter(ttl)

	if err := c.client.Set(ctx, key, data, jitteredTTL).Err(); err != nil {
		c.logger.Warn("redis cache write error",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// Delete removes one or more keys from Redis.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if c.client == nil || len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// DeletePattern scans and deletes all keys matching a glob pattern using SCAN cursor to avoid blocking Redis.
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	if c.client == nil || pattern == "" {
		return nil
	}

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			c.logger.Warn("redis SCAN error during pattern delete",
				slog.String("pattern", pattern),
				slog.String("error", err.Error()),
			)
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				c.logger.Warn("redis DEL error during pattern delete",
					slog.String("error", err.Error()),
				)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// Remember coordinates cache lookup, singleflight anti-stampede coordination, and background population.
func (c *RedisCache) Remember(
	ctx context.Context,
	key string,
	ttl time.Duration,
	dest interface{},
	fetchFunc func(ctx context.Context) (interface{}, error),
) error {
	// 1. Try fast path L2 cache read
	hit, err := c.Get(ctx, key, dest)
	if err == nil && hit {
		return nil
	}

	// 2. Cache miss -> Singleflight anti-stampede execution
	res, sfErr, _ := c.sfGroup.Do(key, func() (interface{}, error) {
		// Double check cache within singleflight worker
		var doubleCheckDest interface{}
		hit, _ := c.Get(ctx, key, &doubleCheckDest)
		if hit && doubleCheckDest != nil {
			return doubleCheckDest, nil
		}

		// Execute actual database query
		fetchedData, err := fetchFunc(ctx)
		if err != nil {
			return nil, err
		}

		// Asynchronously or synchronously populate Redis cache
		_ = c.Set(ctx, key, fetchedData, ttl)

		return fetchedData, nil
	})

	if sfErr != nil {
		return sfErr
	}

	// Copy result into dest via JSON roundtrip for clean type decoupling
	rawBytes, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return json.Unmarshal(rawBytes, dest)
}

func addJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	// +/- 10% jitter
	jitterFactor := 0.9 + (rand.Float64() * 0.2)
	return time.Duration(float64(d) * jitterFactor)
}
