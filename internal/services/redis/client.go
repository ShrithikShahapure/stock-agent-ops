package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shrithkshahapure/stock-agent-ops/internal/config"
	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
)

// Client wraps the Redis client with helper methods
type Client struct {
	client  *redis.Client
	metrics *metrics.Metrics
}

// New creates a new Redis client with retry logic
func New(cfg *config.Config, m *metrics.Metrics) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       cfg.RedisDB,
		PoolSize: 10,
	})

	// Retry connection with backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.Ping(ctx).Result()
		cancel()

		if err == nil {
			log.Printf("Connected to Redis at %s", addr)
			if m != nil {
				m.RedisUp.Set(1)
			}
			return &Client{client: client, metrics: m}, nil
		}

		log.Printf("Redis connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	if m != nil {
		m.RedisUp.Set(0)
	}
	return nil, fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
}

// Ping checks if Redis is available
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		if c.metrics != nil {
			c.metrics.RedisUp.Set(0)
		}
		return err
	}
	if c.metrics != nil {
		c.metrics.RedisUp.Set(1)
	}
	return nil
}

// Get retrieves a value from Redis
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value in Redis with optional TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Incr increments a key
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Expire sets TTL on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

// Keys returns keys matching a pattern
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// FlushAll clears all keys
func (c *Client) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// DBSize returns the number of keys
func (c *Client) DBSize(ctx context.Context) (int64, error) {
	return c.client.DBSize(ctx).Result()
}

// Del deletes keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// IsConnected returns true if Redis is connected
func (c *Client) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.Ping(ctx) == nil
}

// UpdateKeyCount updates the Redis keys metric
func (c *Client) UpdateKeyCount(ctx context.Context) {
	if c.metrics != nil {
		if count, err := c.DBSize(ctx); err == nil {
			c.metrics.RedisKeys.Set(float64(count))
		}
	}
}

// TaskKey returns the Redis key for a task status
func TaskKey(taskID string) string {
	return fmt.Sprintf("task_status:%s", taskID)
}

// CacheKey returns the Redis key for a prediction cache
func CacheKey(ticker string) string {
	return fmt.Sprintf("predict_child_%s", ticker)
}

// RateLimitKey returns the Redis key for rate limiting
func RateLimitKey(prefix string, window int64) string {
	return fmt.Sprintf("rate_limit:%s:%d", prefix, window)
}
