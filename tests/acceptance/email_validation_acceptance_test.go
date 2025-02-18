// Package acceptance contains end-to-end acceptance tests for the email validation service
package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"emailvalidator/internal/api"
	"emailvalidator/internal/middleware"
	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/monitoring"
)

const (
	testRapidAPISecret = "test-secret"
)

type acceptanceTestServer struct {
	server *httptest.Server
	url    string
}

func (ts *acceptanceTestServer) cleanup() error {
	ts.server.Close()
	return nil
}

func setupAcceptanceTestServer(t *testing.T) *acceptanceTestServer {
	// Create service instances
	emailService, err := service.NewEmailService()
	if err != nil {
		t.Fatalf("Failed to initialize email service: %v", err)
	}

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", monitoring.PrometheusHandler())

	// Create a new mux for authenticated routes
	authenticatedMux := http.NewServeMux()

	// Register routes that require authentication
	authenticatedMux.HandleFunc("/validate", handler.HandleValidate)
	authenticatedMux.HandleFunc("/validate/batch", handler.HandleBatchValidate)
	authenticatedMux.HandleFunc("/typo-suggestions", handler.HandleTypoSuggestions)

	// Wrap authenticated routes with monitoring middleware and RapidAPI authentication
	monitoredHandler := monitoring.MetricsMiddleware(authenticatedMux)
	authenticatedHandler := middleware.NewRapidAPIAuthMiddleware(monitoredHandler, testRapidAPISecret)

	// Create final mux that combines both authenticated and unauthenticated routes
	finalMux := http.NewServeMux()

	// Register public endpoints first
	finalMux.Handle("/rapidapi-health", monitoring.MetricsMiddleware(http.HandlerFunc(handler.HandleRapidAPIHealth)))
	finalMux.Handle("/status", monitoring.MetricsMiddleware(http.HandlerFunc(handler.HandleStatus)))
	finalMux.Handle("/metrics", monitoring.MetricsMiddleware(monitoring.PrometheusHandler()))

	// Register authenticated routes last (catch-all)
	finalMux.Handle("/", authenticatedHandler)

	server := httptest.NewServer(finalMux)
	return &acceptanceTestServer{
		server: server,
		url:    server.URL,
	}
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

		req, err := http.NewRequest(http.MethodPost, server.url+"/validate", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create POST request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err := http.DefaultClient.Do(req)
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
		req, err = http.NewRequest(http.MethodGet, server.url+"/validate?email="+email, nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err = http.DefaultClient.Do(req)
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

		// Step 3: Batch validating emails using POST
		t.Log("Step 3: Batch validating emails using POST")
		batchEmails := []string{email, "another@example.com"}
		batchReqBody := model.BatchValidationRequest{Emails: batchEmails}
		batchJsonBody, err := json.Marshal(batchReqBody)
		if err != nil {
			t.Fatalf("Failed to marshal batch request body: %v", err)
		}

		req, err = http.NewRequest(http.MethodPost, server.url+"/validate/batch", bytes.NewBuffer(batchJsonBody))
		if err != nil {
			t.Fatalf("Failed to create batch request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
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

		if len(batchResult.Results) != len(batchEmails) {
			t.Errorf("Expected %d results, got %d", len(batchEmails), len(batchResult.Results))
		}

		for i, result := range batchResult.Results {
			if !result.Validations.Syntax {
				t.Errorf("Expected syntax validation to pass for email %q", batchEmails[i])
			}
		}

		// Step 4: Batch validating emails using GET
		t.Log("Step 4: Batch validating emails using GET")
		var queryParams []string
		for _, email := range batchEmails {
			queryParams = append(queryParams, "email="+email)
		}
		req, err = http.NewRequest(http.MethodGet, server.url+"/validate/batch?"+strings.Join(queryParams, "&"), nil)
		if err != nil {
			t.Fatalf("Failed to create batch GET request: %v", err)
		}
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err = http.DefaultClient.Do(req)
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

		if len(batchGetResult.Results) != len(batchEmails) {
			t.Errorf("Expected %d results, got %d", len(batchEmails), len(batchGetResult.Results))
		}

		for i, result := range batchGetResult.Results {
			if !result.Validations.Syntax {
				t.Errorf("Expected syntax validation to pass for email %q", batchEmails[i])
			}
		}

		// Step 5: Typo suggestions using POST
		t.Log("Step 5: Getting typo suggestions using POST")
		typoEmail := "user@gmial.com"
		typoReqBody := model.TypoSuggestionRequest{Email: typoEmail}
		typoJsonBody, _ := json.Marshal(typoReqBody)

		req, err = http.NewRequest(http.MethodPost, server.url+"/typo-suggestions", bytes.NewBuffer(typoJsonBody))
		if err != nil {
			t.Fatalf("Failed to create typo POST request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err = http.DefaultClient.Do(req)
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
		req, err = http.NewRequest(http.MethodGet, server.url+"/typo-suggestions?email="+typoEmail, nil)
		if err != nil {
			t.Fatalf("Failed to create typo GET request: %v", err)
		}
		req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

		resp, err = http.DefaultClient.Do(req)
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

	testCases := []struct {
		name          string
		endpoint      string
		method        string
		body          interface{}
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Invalid JSON",
			endpoint:      "/validate",
			method:        http.MethodPost,
			body:          []byte(`{"invalid": json}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name:          "Method not allowed",
			endpoint:      "/validate",
			method:        http.MethodPut,
			body:          nil,
			expectedCode:  http.StatusMethodNotAllowed,
			expectedError: "Method not allowed",
		},
		{
			name:          "Missing email parameter",
			endpoint:      "/validate",
			method:        http.MethodGet,
			body:          nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email parameter is required",
		},
		{
			name:          "Empty email",
			endpoint:      "/validate",
			method:        http.MethodPost,
			body:          model.EmailValidationRequest{Email: ""},
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
		{
			name:          "Empty batch request",
			endpoint:      "/validate/batch",
			method:        http.MethodPost,
			body:          model.BatchValidationRequest{Emails: []string{}},
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if tc.body != nil {
				switch v := tc.body.(type) {
				case []byte:
					reqBody = v
				default:
					reqBody, err = json.Marshal(tc.body)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
				}
			}

			req, err := http.NewRequest(tc.method, server.url+tc.endpoint, bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("got status %d, want %d", resp.StatusCode, tc.expectedCode)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tc.expectedError != "" {
				if errMsg, ok := result["error"].(string); !ok || errMsg != tc.expectedError {
					t.Errorf("got error %q, want %q", errMsg, tc.expectedError)
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

	const numRequests = 10
	var wg sync.WaitGroup
	requestErrors := make(chan error, numRequests)
	successfulRequests := int32(0)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			email := fmt.Sprintf("user%d@example.com", i)
			reqBody := model.EmailValidationRequest{Email: email}
			jsonBody, err := json.Marshal(reqBody)
			if err != nil {
				requestErrors <- fmt.Errorf("failed to marshal request body: %v", err)
				return
			}

			req, err := http.NewRequest(http.MethodPost, server.url+"/validate", bytes.NewBuffer(jsonBody))
			if err != nil {
				requestErrors <- fmt.Errorf("failed to create request: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-RapidAPI-Secret", testRapidAPISecret)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				requestErrors <- fmt.Errorf("failed to make request: %v", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			var result model.EmailValidationResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				requestErrors <- fmt.Errorf("failed to decode response: %v", err)
				return
			}

			if result.Email != email {
				requestErrors <- fmt.Errorf("got email %q, want %q", result.Email, email)
				return
			}

			atomic.AddInt32(&successfulRequests, 1)
		}(i)
	}

	wg.Wait()
	close(requestErrors)

	for err := range requestErrors {
		t.Error("Concurrent request error:", err)
	}

	if atomic.LoadInt32(&successfulRequests) < numRequests {
		t.Errorf("Expected at least %d requests handled, got %d", numRequests, atomic.LoadInt32(&successfulRequests))
	}
}
