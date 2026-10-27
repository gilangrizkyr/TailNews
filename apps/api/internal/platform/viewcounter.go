package platform

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"news-portal/apps/api/internal/domain/article"
)

const (
	PendingViewsHashKey = "news:v1:views:pending"
	TrendingViewsZSet   = "news:v1:views:zset:today"
	DedupKeyPrefix      = "news:v1:dedup:"
)

type ViewCounter interface {
	RecordView(ctx context.Context, articleID uint64, ip string, ua string) (bool, error)
	Flush(ctx context.Context) (int, error)
	StartWorker(ctx context.Context, interval time.Duration)
}

type RedisViewCounter struct {
	client      *redis.Client
	articleRepo article.Repository
	logger      *slog.Logger
	wg          sync.WaitGroup
}

// NewRedisViewCounter constructs a RedisViewCounter instance.
func NewRedisViewCounter(client *redis.Client, articleRepo article.Repository, logger *slog.Logger) *RedisViewCounter {
	return &RedisViewCounter{
		client:      client,
		articleRepo: articleRepo,
		logger:      logger,
	}
}

// RecordView tracks an article read with hourly deduplication per IP and User-Agent.
// Returns (true, nil) if recorded, (false, nil) if deduplicated.
func (vc *RedisViewCounter) RecordView(ctx context.Context, articleID uint64, ip string, ua string) (bool, error) {
	if vc.client == nil {
		// Fallback direct increment if redis not available
		return true, vc.articleRepo.IncrementViews(ctx, articleID, 1)
	}

	// Hourly deduplication hash bucket
	hourBucket := time.Now().UTC().Format("2006-01-02-15")
	rawStr := fmt.Sprintf("%s|%s|%d|%s", ip, ua, articleID, hourBucket)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawStr)))
	dedupKey := DedupKeyPrefix + hash

	// SETNX with 1 hour TTL
	recorded, err := vc.client.SetNX(ctx, dedupKey, "1", 1*time.Hour).Result()
	if err != nil {
		vc.logger.Warn("redis dedup check failed; continuing to record view",
			slog.String("error", err.Error()),
		)
	} else if !recorded {
		// Already recorded within this hour window
		return false, nil
	}

	// Increment pending atomic hash and trending sorted set
	pipe := vc.client.Pipeline()
	pipe.HIncrBy(ctx, PendingViewsHashKey, strconv.FormatUint(articleID, 10), 1)
	pipe.ZIncrBy(ctx, TrendingViewsZSet, 1, strconv.FormatUint(articleID, 10))
	_, err = pipe.Exec(ctx)
	if err != nil {
		vc.logger.Error("failed to pipeline increment view counters in redis",
			slog.Uint64("article_id", articleID),
			slog.String("error", err.Error()),
		)
		return false, err
	}

	return true, nil
}

// Flush atomically extracts all accumulated views from Redis and flushes them in a batch to MySQL.
func (vc *RedisViewCounter) Flush(ctx context.Context) (int, error) {
	if vc.client == nil {
		return 0, nil
	}

	// Atomically rotate pending key to isolated flush key
	flushKey := fmt.Sprintf("%s:flush:%d", PendingViewsHashKey, time.Now().UnixNano())
	err := vc.client.Rename(ctx, PendingViewsHashKey, flushKey).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// No pending views accumulated
			return 0, nil
		}
		// In Redis, if key does not exist Rename returns ERR no such key
		if err.Error() == "ERR no such key" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to rename pending views key: %w", err)
	}

	viewsMap, err := vc.client.HGetAll(ctx, flushKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch renamed views map: %w", err)
	}

	if len(viewsMap) == 0 {
		_ = vc.client.Del(ctx, flushKey).Err()
		return 0, nil
	}

	flushedCount := 0
	today := time.Now().UTC()

	for artIDStr, countStr := range viewsMap {
		artID, err := strconv.ParseUint(artIDStr, 10, 64)
		if err != nil {
			continue
		}
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil || count <= 0 {
			continue
		}

		// Update total views in articles table
		if err := vc.articleRepo.IncrementViews(ctx, artID, count); err != nil {
			vc.logger.Error("failed to update article view count in mysql",
				slog.Uint64("article_id", artID),
				slog.Int64("count", count),
				slog.String("error", err.Error()),
			)
		}

		// Upsert daily breakdown views
		if err := vc.articleRepo.RecordDailyViews(ctx, artID, today, count); err != nil {
			vc.logger.Error("failed to upsert article daily views in mysql",
				slog.Uint64("article_id", artID),
				slog.String("error", err.Error()),
			)
		}

		flushedCount++
	}

	// Clean up temporary flush key
	_ = vc.client.Del(ctx, flushKey).Err()

	if flushedCount > 0 {
		vc.logger.Debug("flushed view counts batch to mysql",
			slog.Int("articles_updated", flushedCount),
		)
	}

	return flushedCount, nil
}

// StartWorker kicks off the periodic view flushing background goroutine.
func (vc *RedisViewCounter) StartWorker(ctx context.Context, interval time.Duration) {
	vc.wg.Add(1)
	go func() {
		defer vc.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Final flush on shutdown with fresh background context
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = vc.Flush(shutdownCtx)
				cancel()
				return
			case <-ticker.C:
				_, _ = vc.Flush(ctx)
			}
		}
	}()
}
