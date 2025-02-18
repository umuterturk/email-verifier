// Package validatortest contains unit tests for the validator package
package validatortest

import (
	"emailvalidator/pkg/validator"
	"net"
	"testing"
	"time"
)

func TestValidateSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "Valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "Invalid email - no @",
			email: "invalid-email",
			want:  false,
		},
		{
			name:  "Invalid email - empty",
			email: "",
			want:  false,
		},
		{
			name:  "Invalid email - multiple @",
			email: "user@domain@example.com",
			want:  false,
		},
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateSyntax(tt.email)
			if got != tt.want {
				t.Errorf("ValidateSyntax(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsRoleBased(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Regular email",
			email: "user@example.com",
			want:  false,
		},
		{
			name:  "Admin email",
			email: "admin@example.com",
			want:  true,
		},
		{
			name:  "Support email",
			email: "support@example.com",
			want:  true,
		},
		{
			name:  "Sales email",
			email: "sales@example.com",
			want:  true,
		},
		{
			name:  "Similar but not role email",
			email: "administrator@example.com",
			want:  false,
		},
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsRoleBased(tt.email)
			if got != tt.want {
				t.Errorf("IsRoleBased(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsDisposable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{
			name:   "Regular domain",
			domain: "gmail.com",
			want:   false,
		},
		{
			name:   "Disposable domain",
			domain: "10minutemail.com",
			want:   true,
		},
		{
			name:   "Another disposable domain",
			domain: "mytempmail.com",
			want:   true,
		},
		{
			name:   "Unknown domain",
			domain: "example.com",
			want:   false,
		},
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsDisposable(tt.domain)
			if got != tt.want {
				t.Errorf("IsDisposable(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestGetTypoSuggestions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		email   string
		want    []string
		wantLen int
	}{
		{
			name:    "Gmail typo",
			email:   "user@gmial.com",
			want:    []string{"user@gmail.com"},
			wantLen: 1,
		},
		{
			name:    "Yahoo typo",
			email:   "user@yaho.com",
			want:    []string{"user@yahoo.com"},
			wantLen: 1,
		},
		{
			name:    "No typo",
			email:   "user@gmail.com",
			want:    nil,
			wantLen: 0,
		},
		{
			name:    "Invalid email",
			email:   "invalid-email",
			want:    nil,
			wantLen: 0,
		},
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.GetTypoSuggestions(tt.email)
			if len(got) != tt.wantLen {
				t.Errorf("GetTypoSuggestions(%q) returned %d suggestions, want %d", tt.email, len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0] != tt.want[0] {
				t.Errorf("GetTypoSuggestions(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		validations map[string]bool
		want        int
	}{
		{
			name: "All validations pass",
			validations: map[string]bool{
				"syntax":         true,
				"domain_exists":  true,
				"mx_records":     true,
				"mailbox_exists": true,
				"is_disposable":  false,
				"is_role_based":  false,
			},
			want: 100,
		},
		{
			name: "Only syntax valid",
			validations: map[string]bool{
				"syntax":         true,
				"domain_exists":  false,
				"mx_records":     false,
				"mailbox_exists": false,
				"is_disposable":  true,
				"is_role_based":  true,
			},
			want: 20,
		},
		{
			name: "Role-based email",
			validations: map[string]bool{
				"syntax":         true,
				"domain_exists":  true,
				"mx_records":     true,
				"mailbox_exists": true,
				"is_disposable":  false,
				"is_role_based":  true,
			},
			want: 90,
		},
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.CalculateScore(tt.validations)
			if got != tt.want {
				t.Errorf("CalculateScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

// MockResolver implements DNSResolver for testing
type MockResolver struct {
	validDomains map[string]bool
	validMX      map[string]bool
	delay        time.Duration // for testing timeouts
}

func NewMockResolver() *MockResolver {
	return &MockResolver{
		validDomains: map[string]bool{
			"example.com":      true,
			"gmail.com":        true,
			"valid-domain.com": true,
		},
		validMX: map[string]bool{
			"example.com":      true,
			"gmail.com":        true,
			"valid-domain.com": true,
		},
	}
}

func (r *MockResolver) LookupHost(domain string) ([]string, error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	if r.validDomains[domain] {
		return []string{"192.0.2.1"}, nil
	}
	return nil, &net.DNSError{
		Err:        "no such host",
		Name:       domain,
		IsNotFound: true,
	}
}

func (r *MockResolver) LookupMX(domain string) ([]*net.MX, error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	if r.validMX[domain] {
		return []*net.MX{{Host: "mail." + domain, Pref: 10}}, nil
	}
	return nil, &net.DNSError{
		Err:        "no such host",
		Name:       domain,
		IsNotFound: true,
	}
}

func TestDomainValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		domain        string
		wantExists    bool
		wantMXRecords bool
		checkCache    bool
		setupDelay    time.Duration
	}{
		{
			name:          "Valid domain with MX",
			domain:        "example.com",
			wantExists:    true,
			wantMXRecords: true,
		},
		{
			name:          "Invalid domain",
			domain:        "invalid-domain.com",
			wantExists:    false,
			wantMXRecords: false,
		},
		{
			name:          "Domain with no MX records",
			domain:        "no-mx.com",
			wantExists:    false,
			wantMXRecords: false,
		},
		{
			name:          "Cache test",
			domain:        "example.com",
			wantExists:    true,
			wantMXRecords: true,
			checkCache:    true,
		},
		{
			name:          "Timeout test",
			domain:        "slow-domain.com",
			wantExists:    false,
			wantMXRecords: false,
			setupDelay:    time.Millisecond * 300,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator, err := validator.NewEmailValidator()
			if err != nil {
				t.Fatalf("Failed to create validator: %v", err)
			}
			mockResolver := NewMockResolver()
			validator.SetResolver(mockResolver)
			validator.SetCacheDuration(time.Millisecond * 100)

			mockResolver.delay = tt.setupDelay

			exists := validator.ValidateDomain(tt.domain)
			if exists != tt.wantExists {
				t.Errorf("ValidateDomain(%q) = %v, want %v", tt.domain, exists, tt.wantExists)
			}

			mxExists := validator.ValidateMXRecords(tt.domain)
			if mxExists != tt.wantMXRecords {
				t.Errorf("ValidateMXRecords(%q) = %v, want %v", tt.domain, mxExists, tt.wantMXRecords)
			}

			if tt.checkCache {
				// Force a small delay to ensure the first result is cached
				time.Sleep(time.Millisecond * 10)

				// Set a long delay that would fail the test if the cache wasn't used
				mockResolver.delay = time.Second

				start := time.Now()
				exists = validator.ValidateDomain(tt.domain)
				duration := time.Since(start)

				if duration > time.Millisecond*20 {
					t.Errorf("Cache lookup took too long: %v, should be under 20ms", duration)
				}
				if exists != tt.wantExists {
					t.Errorf("Cached ValidateDomain(%q) = %v, want %v", tt.domain, exists, tt.wantExists)
				}
			}
		})
	}
}

func TestCacheExpiration(t *testing.T) {
	t.Parallel()

	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	mockResolver := NewMockResolver()
	validator.SetResolver(mockResolver)
	validator.SetCacheDuration(time.Millisecond * 50)

	domain := "example.com"

	exists := validator.ValidateDomain(domain)
	if !exists {
		t.Errorf("First check failed: domain should exist")
	}

	time.Sleep(time.Millisecond * 100)

	mockResolver.validDomains[domain] = false

	exists = validator.ValidateDomain(domain)
	if exists {
		t.Error("Got cached result after expiration")
	}
}
