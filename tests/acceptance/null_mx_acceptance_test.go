package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"emailvalidator/internal/model"
)

// TestNullMXRecordAcceptance verifies that domains with null MX records
// are properly identified and marked as invalid for email
func TestNullMXRecordAcceptance(t *testing.T) {
	// Setup test server
	ts := setupAcceptanceTestServer(t)
	defer func() {
		if err := ts.cleanup(); err != nil {
			t.Logf("Failed to cleanup test server: %v", err)
		}
	}()

	// Test cases
	testCases := []struct {
		name           string
		email          string
		expectedStatus model.ValidationStatus
	}{
		{
			name:           "Valid email domain with MX records",
			email:          "test@gmail.com",
			expectedStatus: model.ValidationStatusValid,
		},
		{
			name:           "Domain with null MX record",
			email:          "test@gmail.dk",
			expectedStatus: model.ValidationStatusNoMXRecords,
		},
	}

	// Execute tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test POST endpoint
			testPostValidation(t, ts.url, tc.email, tc.expectedStatus)

			// Test GET endpoint
			testGetValidation(t, ts.url, tc.email, tc.expectedStatus)
		})
	}
}

// TestBatchNullMXRecordAcceptance verifies batch validation with null MX domains
func TestBatchNullMXRecordAcceptance(t *testing.T) {
	// Setup test server
	ts := setupAcceptanceTestServer(t)
	defer func() {
		if err := ts.cleanup(); err != nil {
			t.Logf("Failed to cleanup test server: %v", err)
		}
	}()

	// Create a batch request
	emails := []string{
		"test@gmail.com",   // Valid MX
		"test@gmail.dk",    // Null MX
		"test@example.com", // Valid MX
	}

	// Execute batch test
	t.Run("Batch with mixed valid and null MX domains", func(t *testing.T) {
		testBatchValidation(t, ts.url, emails)
	})
}

// Helper function to test POST validation
func testPostValidation(t *testing.T, baseURL, email string, expectedStatus model.ValidationStatus) {
	// Create request body
	reqBody, err := json.Marshal(model.EmailValidationRequest{Email: email})
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Send POST request
	resp, err := http.Post(
		fmt.Sprintf("%s/api/validate", baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result model.EmailValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify status
	if result.Status != expectedStatus {
		t.Errorf("POST: Expected status %s, got %s for email %s",
			expectedStatus, result.Status, email)
	}
}

// Helper function to test GET validation
func testGetValidation(t *testing.T, baseURL, email string, expectedStatus model.ValidationStatus) {
	// Send GET request
	resp, err := http.Get(
		fmt.Sprintf("%s/api/validate?email=%s", baseURL, email),
	)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result model.EmailValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify status
	if result.Status != expectedStatus {
		t.Errorf("GET: Expected status %s, got %s for email %s",
			expectedStatus, result.Status, email)
	}
}

// Helper function to test batch validation
func testBatchValidation(t *testing.T, baseURL string, emails []string) {
	// Create request body
	reqBody, err := json.Marshal(model.BatchValidationRequest{Emails: emails})
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Send POST request
	resp, err := http.Post(
		fmt.Sprintf("%s/api/validate/batch", baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		t.Fatalf("Failed to send batch request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result model.BatchValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify each result
	if len(result.Results) != len(emails) {
		t.Fatalf("Expected %d results, got %d", len(emails), len(result.Results))
	}

	// Check specific statuses for domains we know about
	for _, res := range result.Results {
		if res.Email == "test@gmail.com" {
			if res.Status != model.ValidationStatusValid &&
				res.Status != model.ValidationStatusProbablyValid {
				t.Errorf("Expected gmail.com to be valid, got %s", res.Status)
			}
		}

		if res.Email == "test@gmail.dk" {
			if res.Status != model.ValidationStatusNoMXRecords {
				t.Errorf("Expected gmail.dk to have no MX records, got %s", res.Status)
			}
		}
	}
}
