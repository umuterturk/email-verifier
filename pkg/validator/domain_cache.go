package validator

import (
	"context"
	"log"
	"sync"
	"time"

	"emailvalidator/pkg/cache"
)

// domainCache represents a cached domain lookup result
type domainCache struct {
	exists    bool
	timestamp time.Time
}

// mxCache represents a cached MX lookup result
type mxCache struct {
	hasMX     bool
	timestamp time.Time
}

// DomainCacheResult is the structure stored in Redis cache for domain existence
type DomainCacheResult struct {
	Exists bool `json:"exists"`
}

// MXCacheResult is the structure stored in Redis cache for MX records
type MXCacheResult struct {
	HasMX bool `json:"has_mx"`
}

// DomainCacheManager handles caching of domain validation results
type DomainCacheManager struct {
	localCache    map[string]domainCache
	localMXCache  map[string]mxCache
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
	redisCache    cache.Cache
}

// NewDomainCacheManager creates a new instance of DomainCacheManager with local cache only
func NewDomainCacheManager(duration time.Duration) *DomainCacheManager {
	return &DomainCacheManager{
		localCache:    make(map[string]domainCache, 100), // Pre-allocate space for better performance
		localMXCache:  make(map[string]mxCache, 100),
		cacheDuration: duration,
		redisCache:    nil,
	}
}

// NewDomainCacheManagerWithRedis creates a new instance of DomainCacheManager with Redis cache
func NewDomainCacheManagerWithRedis(duration time.Duration, redisCache cache.Cache) *DomainCacheManager {
	return &DomainCacheManager{
		localCache:    make(map[string]domainCache, 100),
		localMXCache:  make(map[string]mxCache, 100),
		cacheDuration: duration,
		redisCache:    redisCache,
	}
}

// Get retrieves a cached domain validation result
func (m *DomainCacheManager) Get(domain string) (bool, bool) {
	// L1: Check local in-memory cache first (fastest)
	m.cacheMutex.RLock()
	cached, ok := m.localCache[domain]
	if ok && time.Since(cached.timestamp) <= m.cacheDuration {
		m.cacheMutex.RUnlock()
		return cached.exists, true
	}
	m.cacheMutex.RUnlock()

	// L2: Fall back to Redis if available
	if m.redisCache != nil {
		var result DomainCacheResult
		err := m.redisCache.Get(context.Background(), "domain:"+domain, &result)
		if err == nil {
			// Populate L1 cache from L2 hit
			m.cacheMutex.Lock()
			m.localCache[domain] = domainCache{
				exists:    result.Exists,
				timestamp: time.Now(),
			}
			m.cacheMutex.Unlock()
			log.Printf("[Cache] Redis L2 HIT for domain:%s (exists=%v)", domain, result.Exists)
			return result.Exists, true
		}
	}

	return false, false
}

// Set stores a domain validation result in both L1 (local) and L2 (Redis) caches
func (m *DomainCacheManager) Set(domain string, exists bool) {
	// L1: Store in local in-memory cache
	m.cacheMutex.Lock()
	m.localCache[domain] = domainCache{
		exists:    exists,
		timestamp: time.Now(),
	}
	m.cacheMutex.Unlock()

	// L2: Store in Redis if available
	if m.redisCache != nil {
		result := DomainCacheResult{Exists: exists}
		if err := m.redisCache.Set(context.Background(), "domain:"+domain, result, m.cacheDuration); err != nil {
			log.Printf("[Cache] ERROR: Failed to write domain:%s to Redis: %v", domain, err)
		} else {
			log.Printf("[Cache] Wrote domain:%s to Redis (exists=%v, ttl=%v)", domain, exists, m.cacheDuration)
		}
	}
}

// GetMX retrieves a cached MX validation result
func (m *DomainCacheManager) GetMX(domain string) (bool, bool) {
	// L1: Check local in-memory cache first
	m.cacheMutex.RLock()
	cached, ok := m.localMXCache[domain]
	if ok && time.Since(cached.timestamp) <= m.cacheDuration {
		m.cacheMutex.RUnlock()
		return cached.hasMX, true
	}
	m.cacheMutex.RUnlock()

	// L2: Fall back to Redis if available
	if m.redisCache != nil {
		var result MXCacheResult
		err := m.redisCache.Get(context.Background(), "mx:"+domain, &result)
		if err == nil {
			// Populate L1 cache from L2 hit
			m.cacheMutex.Lock()
			m.localMXCache[domain] = mxCache{
				hasMX:     result.HasMX,
				timestamp: time.Now(),
			}
			m.cacheMutex.Unlock()
			log.Printf("[Cache] Redis L2 HIT for mx:%s (hasMX=%v)", domain, result.HasMX)
			return result.HasMX, true
		}
	}

	return false, false
}

// SetMX stores an MX validation result in both L1 (local) and L2 (Redis) caches
func (m *DomainCacheManager) SetMX(domain string, hasMX bool) {
	// L1: Store in local in-memory cache
	m.cacheMutex.Lock()
	m.localMXCache[domain] = mxCache{
		hasMX:     hasMX,
		timestamp: time.Now(),
	}
	m.cacheMutex.Unlock()

	// L2: Store in Redis if available
	if m.redisCache != nil {
		result := MXCacheResult{HasMX: hasMX}
		if err := m.redisCache.Set(context.Background(), "mx:"+domain, result, m.cacheDuration); err != nil {
			log.Printf("[Cache] ERROR: Failed to write mx:%s to Redis: %v", domain, err)
		} else {
			log.Printf("[Cache] Wrote mx:%s to Redis (hasMX=%v, ttl=%v)", domain, hasMX, m.cacheDuration)
		}
	}
}

// ClearExpired removes expired entries from the local cache
// Note: Redis handles its own TTL expiration
func (m *DomainCacheManager) ClearExpired() {
	m.cacheMutex.Lock()
	now := time.Now()
	for domain, cached := range m.localCache {
		if now.Sub(cached.timestamp) > m.cacheDuration {
			delete(m.localCache, domain)
		}
	}
	for domain, cached := range m.localMXCache {
		if now.Sub(cached.timestamp) > m.cacheDuration {
			delete(m.localMXCache, domain)
		}
	}
	m.cacheMutex.Unlock()
}

// SetDuration updates the cache duration
func (m *DomainCacheManager) SetDuration(duration time.Duration) {
	m.cacheMutex.Lock()
	m.cacheDuration = duration
	m.cacheMutex.Unlock()
}

// CacheDuration returns the current cache duration
func (m *DomainCacheManager) CacheDuration() time.Duration {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	return m.cacheDuration
}

// SetRedisCache sets the Redis cache backend
func (m *DomainCacheManager) SetRedisCache(redisCache cache.Cache) {
	m.redisCache = redisCache
}

// HasRedis returns true if Redis cache is configured
func (m *DomainCacheManager) HasRedis() bool {
	return m.redisCache != nil
}

// Close closes the Redis connection if available
func (m *DomainCacheManager) Close() error {
	if m.redisCache != nil {
		return m.redisCache.Close()
	}
	return nil
}
