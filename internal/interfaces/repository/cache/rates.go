package cache

import (
	"sync"
	"time"
)

type cachedRate struct {
	rate      float64
	expiresAt time.Time
}

type RatesCache struct {
	mu    sync.RWMutex
	rates map[string]cachedRate
	ttl   time.Duration
}

func NewRatesCache(ttlMinutes int) *RatesCache {
	return &RatesCache{
		rates: make(map[string]cachedRate),
		ttl:   time.Duration(ttlMinutes) * time.Minute,
	}
}

func (c *RatesCache) Get(from, to string) (float64, bool) {
	key := c.buildKey(from, to)

	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.rates[key]
	if !exists {
		return 0, false
	}

	if time.Now().After(cached.expiresAt) {
		return 0, false
	}

	return cached.rate, true
}

func (c *RatesCache) Set(from, to string, rate float64) {
	key := c.buildKey(from, to)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.rates[key] = cachedRate{
		rate:      rate,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *RatesCache) buildKey(from, to string) string {
	return from + "_" + to
}

func (c *RatesCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, cached := range c.rates {
		if now.After(cached.expiresAt) {
			delete(c.rates, key)
		}
	}
}
