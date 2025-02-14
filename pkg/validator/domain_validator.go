package validator

import (
	"time"

	"emailvalidator/pkg/monitoring"
)

// DomainValidator handles domain existence validation
type DomainValidator struct {
	resolver     DNSResolver
	cacheManager *DomainCacheManager
}

// NewDomainValidator creates a new instance of DomainValidator
func NewDomainValidator(resolver DNSResolver, cacheManager *DomainCacheManager) *DomainValidator {
	return &DomainValidator{
		resolver:     resolver,
		cacheManager: cacheManager,
	}
}

// Validate checks if the domain exists
func (v *DomainValidator) Validate(domain string) bool {
	// Check cache first
	if exists, found := v.cacheManager.Get(domain); found {
		monitoring.RecordCacheOperation("domain_lookup", "hit")
		return exists
	}
	monitoring.RecordCacheOperation("domain_lookup", "miss")

	// Perform lookup
	start := time.Now()
	_, err := v.resolver.LookupHost(domain)
	monitoring.RecordDNSLookup("host", time.Since(start))
	exists := err == nil

	// Update cache
	v.cacheManager.Set(domain, exists)

	// Periodically clean up expired cache entries
	go v.cacheManager.ClearExpired()

	return exists
}

// ValidateMX checks if the domain has valid MX records
func (v *DomainValidator) ValidateMX(domain string) bool {
	start := time.Now()
	_, err := v.resolver.LookupMX(domain)
	monitoring.RecordDNSLookup("mx", time.Since(start))
	return err == nil
}
