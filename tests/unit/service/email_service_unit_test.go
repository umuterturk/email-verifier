// Package servicetest contains unit tests for the service package
package servicetest

import (
	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/validator"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// mockDNSResolver implements validator.DNSResolver interface
type mockDNSResolver struct {
	delay time.Duration
}

func (m *mockDNSResolver) LookupMX(domain string) ([]*net.MX, error) {
	// Simulate network latency
	time.Sleep(m.delay)
	return []*net.MX{{Host: "mail." + domain, Pref: 10}}, nil
}

func (m *mockDNSResolver) LookupHost(domain string) ([]string, error) {
	// Simulate network latency
	time.Sleep(m.delay)
	return []string{"192.0.2.1"}, nil
}

func TestServiceValidateEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		wantScore     int
		wantSyntax    bool
		wantRoleBased bool
		wantStatus    model.ValidationStatus
	}{
		{
			name:       "Valid email",
			email:      "user@example.com",
			wantScore:  100,
			wantSyntax: true,
			wantStatus: model.ValidationStatusValid,
		},
		{
			name:       "Invalid email format",
			email:      "invalid-email",
			wantScore:  0,
			wantSyntax: false,
			wantStatus: model.ValidationStatusInvalidFormat,
		},
		{
			name:          "Role-based email",
			email:         "admin@example.com",
			wantScore:     90,
			wantSyntax:    true,
			wantRoleBased: true,
			wantStatus:    model.ValidationStatusValid,
		},
		{
			name:       "Empty email",
			email:      "",
			wantScore:  0,
			wantSyntax: false,
			wantStatus: model.ValidationStatusMissingEmail,
		},
	}

	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailService.ValidateEmail(tt.email)

			if result.Score != tt.wantScore {
				t.Errorf("Score = %v, want %v", result.Score, tt.wantScore)
			}
			if result.Validations.Syntax != tt.wantSyntax {
				t.Errorf("Syntax validation = %v, want %v", result.Validations.Syntax, tt.wantSyntax)
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestServiceValidateEmails(t *testing.T) {
	tests := []struct {
		name      string
		emails    []string
		wantCount int
	}{
		{
			name:      "Multiple valid emails",
			emails:    []string{"user1@example.com", "user2@example.com"},
			wantCount: 2,
		},
		{
			name:      "Mixed valid and invalid emails",
			emails:    []string{"user@example.com", "invalid-email"},
			wantCount: 2,
		},
		{
			name:      "Empty list",
			emails:    []string{},
			wantCount: 0,
		},
	}

	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailService.ValidateEmails(tt.emails)

			if len(result.Results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result.Results), tt.wantCount)
			}
		})
	}
}

func TestServiceGetTypoSuggestions(t *testing.T) {
	tests := []struct {
		name               string
		email              string
		wantEmail          string
		wantHasSuggestions bool
	}{
		{
			name:               "Email with typo",
			email:              "user@gmial.com",
			wantEmail:          "user@gmial.com",
			wantHasSuggestions: true,
		},
		{
			name:               "Valid email",
			email:              "user@gmail.com",
			wantEmail:          "user@gmail.com",
			wantHasSuggestions: false,
		},
	}

	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailService.GetTypoSuggestions(tt.email)

			if result.Email != tt.wantEmail {
				t.Errorf("Email = %v, want %v", result.Email, tt.wantEmail)
			}

			hasSuggestions := len(result.Suggestions) > 0
			if hasSuggestions != tt.wantHasSuggestions {
				t.Errorf("HasSuggestions = %v, want %v", hasSuggestions, tt.wantHasSuggestions)
			}
		})
	}
}

func TestServiceGetAPIStatus(t *testing.T) {
	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	// Validate initial state
	status := emailService.GetAPIStatus()
	if status.Status != "healthy" {
		t.Errorf("Status = %v, want 'healthy'", status.Status)
	}
	if status.RequestsHandled != 0 {
		t.Errorf("RequestsHandled = %v, want 0", status.RequestsHandled)
	}

	// Make some requests and check counter
	emailService.ValidateEmail("test@example.com")
	emailService.ValidateEmail("another@example.com")

	status = emailService.GetAPIStatus()
	if status.RequestsHandled != 2 {
		t.Errorf("RequestsHandled = %v, want 2", status.RequestsHandled)
	}

	// Check uptime format
	if _, err := time.ParseDuration(status.Uptime); err != nil {
		t.Errorf("Invalid uptime format: %v", status.Uptime)
	}
}

func TestServiceParallelBatchValidation(t *testing.T) {
	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	// Create a larger batch of emails with mixed domains
	batchSize := 100 // Reduced batch size for faster testing
	emails := make([]string, batchSize)
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "example.com"}

	for i := 0; i < batchSize; i++ {
		domain := domains[i%len(domains)]
		emails[i] = fmt.Sprintf("user%d@%s", i, domain)
	}

	// Time the parallel execution (average of multiple runs)
	runs := 3
	var totalParallel, totalSequential time.Duration

	for run := 0; run < runs; run++ {
		start := time.Now()
		result := emailService.ValidateEmails(emails)
		parallelDuration := time.Since(start)
		totalParallel += parallelDuration

		if len(result.Results) != batchSize {
			t.Errorf("Expected %d results, got %d", batchSize, len(result.Results))
		}

		// Verify order preservation
		for i, res := range result.Results {
			domain := domains[i%len(domains)]
			expectedEmail := fmt.Sprintf("user%d@%s", i, domain)
			if res.Email != expectedEmail {
				t.Errorf("Result at position %d: expected email %s, got %s", i, expectedEmail, res.Email)
			}
		}

		// Time sequential execution
		start = time.Now()
		for _, email := range emails {
			emailService.ValidateEmail(email)
		}
		totalSequential += time.Since(start)
	}

	avgParallel := totalParallel / time.Duration(runs)
	avgSequential := totalSequential / time.Duration(runs)

	// The parallel execution should be significantly faster
	if avgParallel >= avgSequential {
		t.Logf("Average Parallel: %v, Average Sequential: %v", avgParallel, avgSequential)
		t.Error("Parallel execution was not faster than sequential execution")
	} else {
		speedup := float64(avgSequential) / float64(avgParallel)
		t.Logf("Parallel speedup: %.2fx (Avg Parallel: %v, Avg Sequential: %v)", speedup, avgParallel, avgSequential)
	}
}

func TestServiceBatchValidationEdgeCases(t *testing.T) {
	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	tests := []struct {
		name      string
		emails    []string
		wantCount int
	}{
		{
			name:      "Empty batch",
			emails:    []string{},
			wantCount: 0,
		},
		{
			name:      "Single email",
			emails:    []string{"user@example.com"},
			wantCount: 1,
		},
		{
			name: "Mixed valid and invalid",
			emails: []string{
				"valid@example.com",
				"invalid-email",
				"",
				"admin@example.com",
			},
			wantCount: 4,
		},
		{
			name: "All invalid",
			emails: []string{
				"invalid1",
				"invalid2",
				"@invalid3",
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailService.ValidateEmails(tt.emails)

			if len(result.Results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result.Results), tt.wantCount)
			}

			// Verify order preservation
			for i, res := range result.Results {
				if res.Email != tt.emails[i] {
					t.Errorf("Result at position %d: got email %s, want %s", i, res.Email, tt.emails[i])
				}
			}
		})
	}
}

func TestServiceEmailValidationErrorScenarios(t *testing.T) {
	mockResolver := &mockDNSResolver{
		delay: 10 * time.Millisecond, // Add a realistic network delay
	}
	emailValidator, err := validator.NewEmailValidatorWithResolver(mockResolver)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	emailService := service.NewEmailServiceWithDeps(emailValidator)

	tests := []struct {
		name        string
		email       string
		wantStatus  model.ValidationStatus
		wantScore   int
		setupFunc   func()
		cleanupFunc func()
	}{
		{
			name:       "Invalid UTF-8 in email",
			email:      "user@example.com\xfe\xff",
			wantStatus: model.ValidationStatusInvalidFormat,
			wantScore:  0,
		},
		{
			name:       "Very long email address",
			email:      fmt.Sprintf("%s@example.com", strings.Repeat("a", 256)),
			wantStatus: model.ValidationStatusInvalidFormat,
			wantScore:  0,
		},
		{
			name:       "Domain with invalid characters",
			email:      "user@exa\\mple.com",
			wantStatus: model.ValidationStatusInvalidFormat,
			wantScore:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			result := emailService.ValidateEmail(tt.email)

			if result.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", result.Status, tt.wantStatus)
			}
			if result.Score != tt.wantScore {
				t.Errorf("Score = %v, want %v", result.Score, tt.wantScore)
			}

			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}
		})
	}
}
