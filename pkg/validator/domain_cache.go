package validator

import (
	"sync"
	"time"
)

// domainCache represents a cached domain lookup result
type domainCache struct {
	exists    bool
	timestamp time.Time
}

// DomainCacheManager handles caching of domain validation results
type DomainCacheManager struct {
	cache         map[string]domainCache
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
}

// NewDomainCacheManager creates a new instance of DomainCacheManager
func NewDomainCacheManager(duration time.Duration) *DomainCacheManager {
	return &DomainCacheManager{
		cache:         make(map[string]domainCache, 100), // Pre-allocate space for better performance
		cacheDuration: duration,
	}
}

// Get retrieves a cached domain validation result
func (m *DomainCacheManager) Get(domain string) (bool, bool) {
	m.cacheMutex.RLock()
	cache, ok := m.cache[domain]
	if !ok {
		m.cacheMutex.RUnlock()
		return false, false
	}

	// Check expiration without allocating time.Time
	if time.Since(cache.timestamp) > m.cacheDuration {
		m.cacheMutex.RUnlock()
		return false, false
	}

	m.cacheMutex.RUnlock()
	return cache.exists, true
}

// Set stores a domain validation result in the cache
func (m *DomainCacheManager) Set(domain string, exists bool) {
	m.cacheMutex.Lock()
	m.cache[domain] = domainCache{
		exists:    exists,
		timestamp: time.Now(),
	}
	m.cacheMutex.Unlock()
}

// ClearExpired removes expired entries from the cache
func (m *DomainCacheManager) ClearExpired() {
	m.cacheMutex.Lock()
	now := time.Now()
	for domain, cache := range m.cache {
		if now.Sub(cache.timestamp) > m.cacheDuration {
			delete(m.cache, domain)
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
