package validator

import (
	"testing"
)

func TestDisposableValidatorWithStaticReader(t *testing.T) {
	// Create a static reader with test domains
	testDomains := []string{"test.com", "disposable.com"}
	reader := NewStaticDomainReader(testDomains)

	// Create validator with the test reader
	validator, err := NewDisposableValidatorWithReader(reader)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test cases
	tests := []struct {
		domain string
		want   bool
	}{
		{"test.com", true},
		{"disposable.com", true},
		{"valid.com", false},
		{"gmail.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := validator.Validate(tt.domain)
			if got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestDisposableValidatorWithFileReader(t *testing.T) {
	// Create a file reader with the config file
	reader := NewFileDomainReader("../../config/disposable_domains.txt")

	// Create validator with the file reader
	validator, err := NewDisposableValidatorWithReader(reader)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test cases
	tests := []struct {
		domain string
		want   bool
	}{
		{"10minutemail.com", true},
		{"temp-mail.org", true},
		{"gmail.com", false},
		{"yahoo.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := validator.Validate(tt.domain)
			if got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

// MockDomainReader implements DomainReader interface for testing
type MockDomainReader struct {
	domains []string
	err     error
}

func NewMockDomainReader(domains []string, err error) *MockDomainReader {
	return &MockDomainReader{
		domains: domains,
		err:     err,
	}
}

func (r *MockDomainReader) ReadDomains() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.domains, nil
}

func TestDisposableValidatorWithMockReader(t *testing.T) {
	// Create a mock reader
	mockDomains := []string{"mock.com", "test.com"}
	mockReader := NewMockDomainReader(mockDomains, nil)

	// Create validator with the mock reader
	validator, err := NewDisposableValidatorWithReader(mockReader)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test cases
	tests := []struct {
		domain string
		want   bool
	}{
		{"mock.com", true},
		{"test.com", true},
		{"valid.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := validator.Validate(tt.domain)
			if got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}
