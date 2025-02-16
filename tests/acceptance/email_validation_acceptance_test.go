// Package acceptance contains end-to-end acceptance tests for the email validation service
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
	"emailvalidator/pkg/cache"
)

type acceptanceTestServer struct {
	server *http.Server
	url    string
}

func setupAcceptanceTestServer(t *testing.T) *acceptanceTestServer {
	// Create mock cache for testing
	mockCache := cache.NewMockCache()

	// Create service instances with mock cache
	emailService := service.NewEmailServiceWithCache(mockCache)

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Find available port
	port := 8081
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	return &acceptanceTestServer{
		server: server,
		url:    fmt.Sprintf("http://localhost:%d", port),
	}
}

func (ts *acceptanceTestServer) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ts.server.Shutdown(ctx)
}

func TestAcceptanceEmailValidation(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupAcceptanceTestServer(t)
	defer func() {
		if err := server.cleanup(); err != nil {
			t.Errorf("Failed to cleanup test server: %v", err)
		}
	}()

	t.Run("Complete validation workflow", func(t *testing.T) {
		// Step 1: Validate a single valid email using POST
		t.Log("Step 1: Validating a single valid email using POST")
		email := "user@example.com"
		reqBody := model.EmailValidationRequest{Email: email}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(server.url+"/validate", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to make POST request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var result model.EmailValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !result.Validations.Syntax {
			t.Error("Expected syntax validation to pass")
		}

		// Step 2: Validate the same email using GET
		t.Log("Step 2: Validating a single valid email using GET")
		resp, err = http.Get(server.url + "/validate?email=" + email)
		if err != nil {
			t.Fatalf("Failed to make GET request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var getResult model.EmailValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&getResult); err != nil {
			t.Fatalf("Failed to decode GET response: %v", err)
		}

		if !getResult.Validations.Syntax {
			t.Error("Expected syntax validation to pass for GET request")
		}

		// Step 3: Batch validation using POST
		t.Log("Step 3: Batch validating emails using POST")
		batchEmails := []string{"user1@example.com", "user2@example.com"}
		batchReqBody := model.BatchValidationRequest{Emails: batchEmails}
		batchJsonBody, _ := json.Marshal(batchReqBody)

		resp, err = http.Post(server.url+"/validate/batch", "application/json", bytes.NewBuffer(batchJsonBody))
		if err != nil {
			t.Fatalf("Failed to make batch POST request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var batchResult model.BatchValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&batchResult); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		if len(batchResult.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(batchResult.Results))
		}

		// Step 4: Batch validation using GET
		t.Log("Step 4: Batch validating emails using GET")
		resp, err = http.Get(server.url + "/validate/batch?email=user1@example.com&email=user2@example.com")
		if err != nil {
			t.Fatalf("Failed to make batch GET request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var batchGetResult model.BatchValidationResponse
		if err := json.NewDecoder(resp.Body).Decode(&batchGetResult); err != nil {
			t.Fatalf("Failed to decode batch GET response: %v", err)
		}

		if len(batchGetResult.Results) != 2 {
			t.Errorf("Expected 2 results for GET batch, got %d", len(batchGetResult.Results))
		}

		// Step 5: Typo suggestions using POST
		t.Log("Step 5: Getting typo suggestions using POST")
		typoEmail := "user@gmial.com"
		typoReqBody := model.TypoSuggestionRequest{Email: typoEmail}
		typoJsonBody, _ := json.Marshal(typoReqBody)

		resp, err = http.Post(server.url+"/typo-suggestions", "application/json", bytes.NewBuffer(typoJsonBody))
		if err != nil {
			t.Fatalf("Failed to make typo POST request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var typoResult model.TypoSuggestionResponse
		if err := json.NewDecoder(resp.Body).Decode(&typoResult); err != nil {
			t.Fatalf("Failed to decode typo response: %v", err)
		}

		// Step 6: Typo suggestions using GET
		t.Log("Step 6: Getting typo suggestions using GET")
		resp, err = http.Get(server.url + "/typo-suggestions?email=" + typoEmail)
		if err != nil {
			t.Fatalf("Failed to make typo GET request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		var typoGetResult model.TypoSuggestionResponse
		if err := json.NewDecoder(resp.Body).Decode(&typoGetResult); err != nil {
			t.Fatalf("Failed to decode typo GET response: %v", err)
		}
	})
}

func TestAcceptanceErrorScenarios(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupAcceptanceTestServer(t)
	defer func() {
		if err := server.cleanup(); err != nil {
			t.Errorf("Failed to cleanup test server: %v", err)
		}
	}()

	tests := []struct {
		name                 string
		endpoint             string
		method               string
		body                 string
		wantStatus           int
		wantErrorMatch       string
		wantValidationStatus model.ValidationStatus
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
			method:         http.MethodPut,
			body:           "",
			wantStatus:     http.StatusMethodNotAllowed,
			wantErrorMatch: "Method not allowed",
		},
		{
			name:           "Missing email parameter",
			endpoint:       "/validate",
			method:         http.MethodGet,
			body:           "",
			wantStatus:     http.StatusBadRequest,
			wantErrorMatch: "Email parameter is required",
		},
		{
			name:                 "Empty email",
			endpoint:             "/validate",
			method:               http.MethodPost,
			body:                 `{"email": ""}`,
			wantStatus:           http.StatusOK,
			wantErrorMatch:       "",
			wantValidationStatus: model.ValidationStatusMissingEmail,
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
				client := &http.Client{}
				req, reqErr := http.NewRequest(tt.method, server.url+tt.endpoint, nil)
				if reqErr != nil {
					t.Fatalf("Failed to create request: %v", reqErr)
				}
				resp, err = client.Do(req)
			}

			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			// Read the response body
			var body json.RawMessage
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
				// For validation responses, check the validation status
				var validationResp model.EmailValidationResponse
				if err := json.Unmarshal(body, &validationResp); err != nil {
					t.Fatalf("Failed to decode validation response: %v", err)
				}
				if validationResp.Email != "" {
					t.Errorf("Expected empty email to be preserved, got %q", validationResp.Email)
				}
				if tt.wantValidationStatus != "" && validationResp.Status != tt.wantValidationStatus {
					t.Errorf("got validation status %q, want %q", validationResp.Status, tt.wantValidationStatus)
				}
			}
		})
	}
}

func TestAcceptanceConcurrentRequests(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping acceptance tests in CI environment")
	}

	server := setupAcceptanceTestServer(t)
	defer func() {
		if err := server.cleanup(); err != nil {
			t.Errorf("Failed to cleanup test server: %v", err)
		}
	}()

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
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	var status model.APIStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if status.RequestsHandled < int64(concurrentRequests) {
		t.Errorf("Expected at least %d requests handled, got %d", concurrentRequests, status.RequestsHandled)
	}
}
