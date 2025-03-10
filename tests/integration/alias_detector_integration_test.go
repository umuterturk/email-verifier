package integration

import (
	"testing"

	"emailvalidator/internal/service"
)

func TestEmailServiceAliasDetection(t *testing.T) {
	// Create a service with the real validator
	svc, err := service.NewEmailService()
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	tests := []struct {
		name            string
		email           string
		expectedAlias   string
		shouldHaveAlias bool
	}{
		{
			name:            "Gmail with plus sign",
			email:           "example+test@gmail.com",
			expectedAlias:   "example@gmail.com",
			shouldHaveAlias: true,
		},
		{
			name:            "Gmail with dots",
			email:           "ex.ample@gmail.com",
			expectedAlias:   "example@gmail.com",
			shouldHaveAlias: true,
		},
		{
			name:            "Yahoo with hyphen",
			email:           "example-test@yahoo.com",
			expectedAlias:   "example@yahoo.com",
			shouldHaveAlias: true,
		},
		{
			name:            "Not an alias",
			email:           "example@gmail.com",
			expectedAlias:   "",
			shouldHaveAlias: false,
		},
		{
			name:            "Invalid email",
			email:           "invalid-email",
			expectedAlias:   "",
			shouldHaveAlias: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the email
			result := svc.ValidateEmail(tt.email)

			// Check if AliasOf field is set correctly
			if tt.shouldHaveAlias {
				if result.AliasOf == "" {
					t.Errorf("Expected AliasOf field to be set for email %s, but it was empty", tt.email)
				} else if result.AliasOf != tt.expectedAlias {
					t.Errorf("Expected AliasOf to be %q, got %q", tt.expectedAlias, result.AliasOf)
				}
			} else {
				if result.AliasOf != "" {
					t.Errorf("Expected AliasOf field to be empty for email %s, but got %q", tt.email, result.AliasOf)
				}
			}
		})
	}
}

func TestBatchValidationAliasDetection(t *testing.T) {
	// Create a service
	svc, err := service.NewEmailService()
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	// Test batch validation with emails that should have aliases
	emails := []string{
		"example+test@gmail.com",
		"ex.ample@gmail.com",
		"example-test@yahoo.com",
		"example@gmail.com", // Not an alias
		"example+test@outlook.com",
	}

	// Validate the emails
	result := svc.ValidateEmails(emails)

	// Verify results
	if len(result.Results) != len(emails) {
		t.Fatalf("Expected %d results, got %d", len(emails), len(result.Results))
	}

	// Expected aliases for each email
	expectedAliases := []struct {
		email           string
		alias           string
		shouldHaveAlias bool
	}{
		{emails[0], "example@gmail.com", true},
		{emails[1], "example@gmail.com", true},
		{emails[2], "example@yahoo.com", true},
		{emails[3], "", false},
		{emails[4], "example@outlook.com", true},
	}

	for i, expected := range expectedAliases {
		response := result.Results[i]
		if expected.shouldHaveAlias {
			if response.AliasOf == "" {
				t.Errorf("Expected alias for email %s, but got none", expected.email)
			} else if response.AliasOf != expected.alias {
				t.Errorf("For email %s: expected alias %q, got %q", expected.email, expected.alias, response.AliasOf)
			}
		} else {
			if response.AliasOf != "" {
				t.Errorf("Expected no alias for email %s, but got %q", expected.email, response.AliasOf)
			}
		}
	}
}
