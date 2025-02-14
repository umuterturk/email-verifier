package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"emailvalidator/internal/api"
	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
)

type testServer struct {
	server *http.Server
	url    string
}

func setupTestServer(t *testing.T) *testServer {
	// Create service instances
	emailService := service.NewEmailService()

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Find available port
	port := 8081
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	return &testServer{
		server: server,
		url:    fmt.Sprintf("http://localhost:%d", port),
	}
}

func (ts *testServer) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ts.server.Shutdown(ctx)
}

func TestEndToEndEmailValidation(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupTestServer(t)
	defer server.cleanup()

	t.Run("Complete validation workflow", func(t *testing.T) {
		// Step 1: Validate a single valid email
		t.Log("Step 1: Validating a single valid email")
		email := "user@example.com"
		reqBody := model.EmailValidationRequest{Email: email}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(server.url+"/validate", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result model.EmailValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !result.Validations.Syntax {
			t.Errorf("Expected valid syntax for %s", email)
		}
		if result.Score < 90 {
			t.Errorf("Expected high score for valid email, got %d", result.Score)
		}

		// Step 2: Batch validate multiple emails
		t.Log("Step 2: Batch validating multiple emails")
		batchReqBody := model.BatchValidationRequest{
			Emails: []string{
				"user1@example.com",
				"invalid-email",
				"admin@example.com",
			},
		}
		jsonBody, _ = json.Marshal(batchReqBody)

		resp, err = http.Post(server.url+"/validate/batch", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
		}
		defer resp.Body.Close()

		var batchResult model.BatchValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&batchResult); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		if len(batchResult.Results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(batchResult.Results))
		}

		// Step 3: Check typo suggestions
		t.Log("Step 3: Checking typo suggestions")
		typoReqBody := model.TypoSuggestionRequest{
			Email: "user@gmial.com",
		}
		jsonBody, _ = json.Marshal(typoReqBody)

		resp, err = http.Post(server.url+"/typo-suggestions", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make typo suggestion request: %v", err)
		}
		defer resp.Body.Close()

		var typoResult model.TypoSuggestionResponse
		if err := json.NewDecoder(resp.Body).Decode(&typoResult); err != nil {
			t.Fatalf("Failed to decode typo suggestion response: %v", err)
		}

		if len(typoResult.Suggestions) == 0 {
			t.Error("Expected typo suggestions for gmial.com")
		}

		// Step 4: Verify API status
		t.Log("Step 4: Verifying API status")
		resp, err = http.Get(server.url + "/status")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}
		defer resp.Body.Close()

		var statusResult model.APIStatus
		if err := json.NewDecoder(resp.Body).Decode(&statusResult); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		if statusResult.Status != "healthy" {
			t.Errorf("Expected API to be healthy, got %s", statusResult.Status)
		}
		if statusResult.RequestsHandled < 3 { // We made at least 3 requests
			t.Errorf("Expected at least 3 requests handled, got %d", statusResult.RequestsHandled)
		}
	})
}

func TestErrorScenarios(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupTestServer(t)
	defer server.cleanup()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		body           string
		wantStatus     int
		wantErrorMatch string
	}{
		{
			name:           "Invalid JSON",
			endpoint:       "/validate",
			method:         http.MethodPost,
			body:           `{"invalid json"`,
			wantStatus:     http.StatusBadRequest,
			wantErrorMatch: "Invalid request body",
		},
		{
			name:           "Method not allowed",
			endpoint:       "/validate",
			method:         http.MethodGet,
			body:           "",
			wantStatus:     http.StatusMethodNotAllowed,
			wantErrorMatch: "Method not allowed",
		},
		{
			name:           "Empty email",
			endpoint:       "/validate",
			method:         http.MethodPost,
			body:           `{"email": ""}`,
			wantStatus:     http.StatusOK,
			wantErrorMatch: "", // Not an error response, but a valid response with a message
		},
		{
			name:           "Empty batch request",
			endpoint:       "/validate/batch",
			method:         http.MethodPost,
			body:           `{"emails": []}`,
			wantStatus:     http.StatusOK,
			wantErrorMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			switch tt.method {
			case http.MethodPost:
				resp, err = http.Post(server.url+tt.endpoint, "application/json", bytes.NewBufferString(tt.body))
			case http.MethodGet:
				resp, err = http.Get(server.url + tt.endpoint)
			default:
				t.Fatalf("Unsupported method: %s", tt.method)
			}

			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			// Read the response body
			body, err := json.RawMessage{}, error(nil)
			if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.wantErrorMatch != "" {
				// For error responses, expect error field
				var errorResp map[string]string
				if err := json.Unmarshal(body, &errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if msg, ok := errorResp["error"]; !ok || msg != tt.wantErrorMatch {
					t.Errorf("got error message %q, want %q", msg, tt.wantErrorMatch)
				}
			} else if tt.endpoint == "/validate" && tt.method == http.MethodPost {
				// For validation responses, check the message field
				var validationResp model.EmailValidationResponse
				if err := json.Unmarshal(body, &validationResp); err != nil {
					t.Fatalf("Failed to decode validation response: %v", err)
				}
				if validationResp.Email != "" {
					t.Errorf("Expected empty email to be preserved, got %q", validationResp.Email)
				}
				if validationResp.Message != "Email address is required" {
					t.Errorf("got message %q, want %q", validationResp.Message, "Email address is required")
				}
			}
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupTestServer(t)
	defer server.cleanup()

	// Number of concurrent requests to make
	concurrentRequests := 10

	// Create a channel to collect results
	results := make(chan error, concurrentRequests)

	// Make concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		go func(i int) {
			email := fmt.Sprintf("user%d@example.com", i)
			reqBody := model.EmailValidationRequest{Email: email}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(server.url+"/validate", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				results <- fmt.Errorf("request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			var result model.EmailValidationResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				results <- fmt.Errorf("failed to decode response: %v", err)
				return
			}

			if !result.Validations.Syntax {
				results <- fmt.Errorf("invalid syntax validation for %s", email)
				return
			}

			results <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < concurrentRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request error: %v", err)
		}
	}

	// Verify API handled all requests
	resp, err := http.Get(server.url + "/status")
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	defer resp.Body.Close()

	var status model.APIStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if status.RequestsHandled < int64(concurrentRequests) {
		t.Errorf("Expected at least %d requests handled, got %d", concurrentRequests, status.RequestsHandled)
	}
}
