package validator

import (
	"sync"
	"time"
)

// SMTPCacheEntry represents a cached SMTP verification result
type SMTPCacheEntry struct {
	IsCatchAll bool
	Timestamp  time.Time
}

// SMTPCacheManager handles caching of SMTP-related results
// Only caches domain-level data (catch-all status), never individual emails
// This maintains GDPR/CCPA compliance by not storing email addresses
type SMTPCacheManager struct {
	catchAllCache map[string]SMTPCacheEntry
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
}

// NewSMTPCacheManager creates a new SMTP cache manager
func NewSMTPCacheManager(duration time.Duration) *SMTPCacheManager {
	return &SMTPCacheManager{
		catchAllCache: make(map[string]SMTPCacheEntry, 100),
		cacheDuration: duration,
	}
}

// GetCatchAll retrieves cached catch-all status for a domain
// Returns (isCatchAll, found)
func (m *SMTPCacheManager) GetCatchAll(domain string) (bool, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	entry, ok := m.catchAllCache[domain]
	if !ok {
		return false, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > m.cacheDuration {
		return false, false
	}

	return entry.IsCatchAll, true
}

// SetCatchAll stores catch-all status for a domain
func (m *SMTPCacheManager) SetCatchAll(domain string, isCatchAll bool) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.catchAllCache[domain] = SMTPCacheEntry{
		IsCatchAll: isCatchAll,
		Timestamp:  time.Now(),
	}
}

// ClearExpired removes expired entries from the cache
func (m *SMTPCacheManager) ClearExpired() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	now := time.Now()
	for domain, entry := range m.catchAllCache {
		if now.Sub(entry.Timestamp) > m.cacheDuration {
			delete(m.catchAllCache, domain)
		}
	}
}

// Clear removes all entries from the cache
func (m *SMTPCacheManager) Clear() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	m.catchAllCache = make(map[string]SMTPCacheEntry, 100)
}
