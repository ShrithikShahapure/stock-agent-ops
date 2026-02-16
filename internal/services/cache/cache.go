package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/shrithkshahapure/stock-agent-ops/internal/metrics"
	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
)

// Cache provides prediction caching functionality
type Cache struct {
	redis   *redisclient.Client
	metrics *metrics.Metrics
	ttl     time.Duration
}

// NewCache creates a new cache service
func NewCache(redis *redisclient.Client, m *metrics.Metrics, ttl time.Duration) *Cache {
	return &Cache{
		redis:   redis,
		metrics: m,
		ttl:     ttl,
	}
}

// Get retrieves a cached value for a ticker
func (c *Cache) Get(ticker string) (map[string]interface{}, bool) {
	if c.redis == nil {
		return nil, false
	}

	ctx := context.Background()
	key := redisclient.CacheKey(strings.ToLower(ticker))

	val, err := c.redis.Get(ctx, key)
	if err != nil {
		// Cache miss
		if c.metrics != nil {
			c.metrics.CacheMiss.WithLabelValues(key).Inc()
		}
		return nil, false
	}

	// Cache hit
	if c.metrics != nil {
		c.metrics.CacheHit.WithLabelValues(key).Inc()
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		log.Printf("Failed to unmarshal cached value: %v", err)
		return nil, false
	}

	return data, true
}

// Set stores a value in the cache
func (c *Cache) Set(ticker string, data map[string]interface{}) error {
	if c.redis == nil {
		return nil
	}

	ctx := context.Background()
	key := redisclient.CacheKey(strings.ToLower(ticker))

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, string(jsonData), c.ttl)
}

// Delete removes a cached value
func (c *Cache) Delete(ticker string) error {
	if c.redis == nil {
		return nil
	}

	ctx := context.Background()
	key := redisclient.CacheKey(strings.ToLower(ticker))

	return c.redis.Del(ctx, key)
}

// GetCachedTickers returns a list of all cached tickers
func (c *Cache) GetCachedTickers() ([]string, error) {
	if c.redis == nil {
		return nil, nil
	}

	ctx := context.Background()
	pattern := "predict_child_*"

	keys, err := c.redis.Keys(ctx, pattern)
	if err != nil {
		return nil, err
	}

	// Extract ticker names from keys
	tickers := make([]string, 0, len(keys))
	prefix := "predict_child_"
	for _, key := range keys {
		ticker := strings.TrimPrefix(key, prefix)
		tickers = append(tickers, strings.ToUpper(ticker))
	}

	return tickers, nil
}

// GetForTicker retrieves the cached data for a specific ticker
func (c *Cache) GetForTicker(ticker string) (map[string]interface{}, error) {
	data, found := c.Get(ticker)
	if !found {
		return nil, nil
	}
	return data, nil
}
