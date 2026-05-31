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
	// Check cache first
	if hasMX, found := v.cacheManager.GetMX(domain); found {
		monitoring.RecordCacheOperation("mx_lookup", "hit")
		return hasMX
	}
	monitoring.RecordCacheOperation("mx_lookup", "miss")

	start := time.Now()
	mxRecords, err := v.resolver.LookupMX(domain)
	monitoring.RecordDNSLookup("mx", time.Since(start))

	var hasMX bool

	// If there's an error in lookup, the domain doesn't have valid MX records
	if err != nil {
		hasMX = false
	} else if len(mxRecords) == 0 {
		// No MX records means the domain doesn't accept email
		hasMX = false
	} else if len(mxRecords) == 1 && mxRecords[0].Host == "." {
		// Check for null MX record (RFC 7505)
		// A single MX record with "." as the host indicates the domain doesn't accept email
		hasMX = false
	} else {
		// Otherwise, the domain has valid MX records
		hasMX = true
	}

	// Update cache
	v.cacheManager.SetMX(domain, hasMX)

	// Periodically clean up expired cache entries
	go v.cacheManager.ClearExpired()

	return hasMX
}
