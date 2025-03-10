package validatortest

import (
	"testing"

	"emailvalidator/pkg/validator"
)

func TestGmailAliasDetection(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Plain email",
			email:    "example@gmail.com",
			expected: "",
		},
		{
			name:     "Email with dots",
			email:    "ex.ample@gmail.com",
			expected: "example@gmail.com",
		},
		{
			name:     "Email with plus",
			email:    "example+test@gmail.com",
			expected: "example@gmail.com",
		},
		{
			name:     "Email with dots and plus",
			email:    "ex.am.ple+test@gmail.com",
			expected: "example@gmail.com",
		},
		{
			name:     "Googlemail domain",
			email:    "ex.ample+test@googlemail.com",
			expected: "example@gmail.com",
		},
	}

	detector := validator.NewAliasDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectAlias(tt.email)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestYahooAliasDetection(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Plain email",
			email:    "example@yahoo.com",
			expected: "",
		},
		{
			name:     "Email with hyphen",
			email:    "example-test@yahoo.com",
			expected: "example@yahoo.com",
		},
	}

	detector := validator.NewAliasDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectAlias(tt.email)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestOutlookAliasDetection(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Plain email",
			email:    "example@outlook.com",
			expected: "",
		},
		{
			name:     "Email with plus",
			email:    "example+test@outlook.com",
			expected: "example@outlook.com",
		},
		{
			name:     "Hotmail domain",
			email:    "example+test@hotmail.com",
			expected: "example@hotmail.com",
		},
		{
			name:     "Live domain",
			email:    "example+test@live.com",
			expected: "example@live.com",
		},
	}

	detector := validator.NewAliasDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectAlias(tt.email)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInvalidEmailHandling(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Invalid email without @",
			email:    "example.com",
			expected: "",
		},
		{
			name:     "Empty email",
			email:    "",
			expected: "",
		},
		{
			name:     "Unknown provider",
			email:    "example+test@unknown.com",
			expected: "",
		},
	}

	detector := validator.NewAliasDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectAlias(tt.email)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEmailValidatorAliasDetection(t *testing.T) {
	validator, err := validator.NewEmailValidator()
	if err != nil {
		t.Fatalf("Failed to create email validator: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Gmail alias",
			email:    "example+test@gmail.com",
			expected: "example@gmail.com",
		},
		{
			name:     "Yahoo alias",
			email:    "example-test@yahoo.com",
			expected: "example@yahoo.com",
		},
		{
			name:     "Outlook alias",
			email:    "example+test@outlook.com",
			expected: "example@outlook.com",
		},
		{
			name:     "Not an alias",
			email:    "example@gmail.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.DetectAlias(tt.email)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
