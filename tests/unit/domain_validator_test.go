package unit

import (
	"emailvalidator/pkg/validator"
	"net"
	"testing"
)

// MockResolver implements the validator.DNSResolver interface for testing
type MockResolver struct {
	HostResults map[string][]string
	MXResults   map[string][]*net.MX
	HostErrors  map[string]error
	MXErrors    map[string]error
}

func (r *MockResolver) LookupHost(domain string) ([]string, error) {
	if err, ok := r.HostErrors[domain]; ok {
		return nil, err
	}
	return r.HostResults[domain], nil
}

func (r *MockResolver) LookupMX(domain string) ([]*net.MX, error) {
	if err, ok := r.MXErrors[domain]; ok {
		return nil, err
	}
	return r.MXResults[domain], nil
}

func TestNullMXRecord(t *testing.T) {
	// Create a mock resolver with test data
	mockResolver := &MockResolver{
		HostResults: map[string][]string{
			"example.com":     {"192.0.2.1"},
			"gmail.com":       {"172.217.0.1"},
			"gmail.dk":        {"172.217.0.2"},
			"nonexistent.com": {},
		},
		MXResults: map[string][]*net.MX{
			"example.com": {
				{Host: "mail.example.com", Pref: 10},
			},
			"gmail.com": {
				{Host: "gmail-smtp-in.l.google.com", Pref: 5},
				{Host: "alt1.gmail-smtp-in.l.google.com", Pref: 10},
			},
			// Null MX record as seen in the DNS lookup for gmail.dk
			"gmail.dk": {
				{Host: ".", Pref: 0},
			},
			// Empty MX records
			"nonexistent.com": {},
		},
		HostErrors: map[string]error{},
		MXErrors:   map[string]error{},
	}

	// Create a cache manager and domain validator with our mock resolver
	cacheManager := validator.NewDomainCacheManager(0)
	domainValidator := validator.NewDomainValidator(mockResolver, cacheManager)

	// Test cases
	testCases := []struct {
		name           string
		domain         string
		expectedResult bool
	}{
		{
			name:           "Domain with valid MX records",
			domain:         "example.com",
			expectedResult: true,
		},
		{
			name:           "Gmail with valid MX records",
			domain:         "gmail.com",
			expectedResult: true,
		},
		{
			name:           "Domain with null MX record",
			domain:         "gmail.dk",
			expectedResult: false,
		},
		{
			name:           "Domain with no MX records",
			domain:         "nonexistent.com",
			expectedResult: false,
		},
	}

	// Run the tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := domainValidator.ValidateMX(tc.domain)
			if result != tc.expectedResult {
				t.Errorf("ValidateMX for %s returned %v, expected %v",
					tc.domain, result, tc.expectedResult)
			}
		})
	}
}
