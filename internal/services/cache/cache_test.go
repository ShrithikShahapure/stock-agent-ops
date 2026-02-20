package cache

import (
	"testing"
)

// TestCacheGetWithNilRedis verifies that Get returns (nil, false) gracefully when Redis is nil.
// This covers the nil-safety guard in every cache method.
func TestCacheGetWithNilRedis(t *testing.T) {
	c := &Cache{redis: nil}
	got, ok := c.Get("AAPL")
	if ok {
		t.Error("Get(nil redis) ok = true, want false")
	}
	if got != nil {
		t.Errorf("Get(nil redis) = %v, want nil", got)
	}
}

func TestCacheSetWithNilRedis(t *testing.T) {
	c := &Cache{redis: nil}
	err := c.Set("AAPL", map[string]interface{}{"x": 1})
	if err != nil {
		t.Errorf("Set(nil redis) err = %v, want nil", err)
	}
}

func TestCacheDeleteWithNilRedis(t *testing.T) {
	c := &Cache{redis: nil}
	err := c.Delete("AAPL")
	if err != nil {
		t.Errorf("Delete(nil redis) err = %v, want nil", err)
	}
}

func TestCacheGetCachedTickersWithNilRedis(t *testing.T) {
	c := &Cache{redis: nil}
	tickers, err := c.GetCachedTickers()
	if err != nil {
		t.Errorf("GetCachedTickers(nil redis) err = %v, want nil", err)
	}
	if tickers != nil {
		t.Errorf("GetCachedTickers(nil redis) = %v, want nil", tickers)
	}
}

func TestCacheGetForTickerWithNilRedis(t *testing.T) {
	c := &Cache{redis: nil}
	data, err := c.GetForTicker("AAPL")
	if err != nil {
		t.Errorf("GetForTicker(nil redis) err = %v, want nil", err)
	}
	if data != nil {
		t.Errorf("GetForTicker(nil redis) = %v, want nil", data)
	}
}
