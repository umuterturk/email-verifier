package service

import (
	"emailvalidator/internal/model"
	"fmt"
	"testing"
	"time"
)

func TestValidateEmail(t *testing.T) {
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

	service := NewEmailService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateEmail(tt.email)

			if result.Score != tt.wantScore {
				t.Errorf("Score = %v, want %v", result.Score, tt.wantScore)
			}
			if result.Validations.Syntax != tt.wantSyntax {
				t.Errorf("Syntax = %v, want %v", result.Validations.Syntax, tt.wantSyntax)
			}
			if result.Validations.IsRoleBased != tt.wantRoleBased {
				t.Errorf("IsRoleBased = %v, want %v", result.Validations.IsRoleBased, tt.wantRoleBased)
			}
			if result.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestValidateEmails(t *testing.T) {
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

	service := NewEmailService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateEmails(tt.emails)

			if len(result.Results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result.Results), tt.wantCount)
			}
		})
	}
}

func TestGetTypoSuggestions(t *testing.T) {
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

	service := NewEmailService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetTypoSuggestions(tt.email)

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

func TestGetAPIStatus(t *testing.T) {
	service := NewEmailService()

	// Validate initial state
	status := service.GetAPIStatus()
	if status.Status != "healthy" {
		t.Errorf("Status = %v, want 'healthy'", status.Status)
	}
	if status.RequestsHandled != 0 {
		t.Errorf("RequestsHandled = %v, want 0", status.RequestsHandled)
	}

	// Make some requests and check counter
	service.ValidateEmail("test@example.com")
	service.ValidateEmail("another@example.com")

	status = service.GetAPIStatus()
	if status.RequestsHandled != 2 {
		t.Errorf("RequestsHandled = %v, want 2", status.RequestsHandled)
	}

	// Check uptime format
	if _, err := time.ParseDuration(status.Uptime); err != nil {
		t.Errorf("Invalid uptime format: %v", status.Uptime)
	}
}

func TestParallelBatchValidation(t *testing.T) {
	service := NewEmailService()

	// Create a large batch of emails
	batchSize := 100
	emails := make([]string, batchSize)
	for i := 0; i < batchSize; i++ {
		emails[i] = fmt.Sprintf("user%d@example.com", i)
	}

	// Time the parallel execution
	start := time.Now()
	result := service.ValidateEmails(emails)
	parallelDuration := time.Since(start)

	// Verify results
	if len(result.Results) != batchSize {
		t.Errorf("Expected %d results, got %d", batchSize, len(result.Results))
	}

	// Verify order preservation
	for i, res := range result.Results {
		expectedEmail := fmt.Sprintf("user%d@example.com", i)
		if res.Email != expectedEmail {
			t.Errorf("Result at position %d: expected email %s, got %s", i, expectedEmail, res.Email)
		}
	}

	// Time sequential execution for comparison
	start = time.Now()
	for _, email := range emails {
		service.ValidateEmail(email)
	}
	sequentialDuration := time.Since(start)

	// The parallel execution should be significantly faster
	if parallelDuration >= sequentialDuration {
		t.Logf("Parallel: %v, Sequential: %v", parallelDuration, sequentialDuration)
		t.Error("Parallel execution was not faster than sequential execution")
	} else {
		speedup := float64(sequentialDuration) / float64(parallelDuration)
		t.Logf("Parallel speedup: %.2fx (Parallel: %v, Sequential: %v)", speedup, parallelDuration, sequentialDuration)
	}
}

func TestBatchValidationEdgeCases(t *testing.T) {
	service := NewEmailService()

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
			result := service.ValidateEmails(tt.emails)

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
