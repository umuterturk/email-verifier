// Package integration contains integration tests for the API handlers
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"emailvalidator/internal/api"
	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/cache"
	"emailvalidator/pkg/validator"
)

var (
	testServer     *httptest.Server
	testServerOnce sync.Once
)

func getTestServer(t *testing.T) *httptest.Server {
	testServerOnce.Do(func() {
		// Create mock cache and validator
		mockCache := cache.NewMockCache()
		emailValidator := validator.NewEmailValidator()

		// Create a new service with mock dependencies
		emailService := service.NewEmailServiceWithDeps(mockCache, emailValidator)
		handler := api.NewHandler(emailService)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux)
		testServer = httptest.NewServer(mux)
	})
	return testServer
}

func TestHandleValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	server := getTestServer(t)

	tests := []struct {
		name       string
		email      string
		method     string
		wantStatus int
		wantScore  int
	}{
		{
			name:       "Valid email POST",
			email:      "user@example.com",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
			wantScore:  100,
		},
		{
			name:       "Invalid email POST",
			email:      "invalid-email",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
			wantScore:  0,
		},
		{
			name:       "Valid email GET",
			email:      "user@example.com",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantScore:  100,
		},
		{
			name:       "Missing email parameter GET",
			email:      "",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Method not allowed",
			email:      "user@example.com",
			method:     http.MethodPut,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				t.Logf("Subtest '%s' took %v", tt.name, time.Since(start))
			}()

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			var resp *http.Response
			var err error

			switch tt.method {
			case http.MethodPost:
				reqBody := model.EmailValidationRequest{Email: tt.email}
				jsonBody, _ := json.Marshal(reqBody)
				resp, err = client.Post(server.URL+"/validate", "application/json", bytes.NewBuffer(jsonBody))
			case http.MethodGet:
				req, reqErr := http.NewRequest(http.MethodGet, server.URL+"/validate", nil)
				if reqErr != nil {
					t.Fatalf("Failed to create request: %v", reqErr)
				}
				if tt.email != "" {
					q := req.URL.Query()
					q.Add("email", tt.email)
					req.URL.RawQuery = q.Encode()
				}
				resp, err = client.Do(req)
			default:
				req, reqErr := http.NewRequest(tt.method, server.URL+"/validate", nil)
				if reqErr != nil {
					t.Fatalf("Failed to create request: %v", reqErr)
				}
				resp, err = client.Do(req)
			}

			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if (tt.method == http.MethodPost || tt.method == http.MethodGet) && resp.StatusCode == http.StatusOK {
				var result model.EmailValidationResponse
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if result.Score != tt.wantScore {
					t.Errorf("got score %d, want %d", result.Score, tt.wantScore)
				}
			}
		})
	}
}

func TestHandleBatchValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	server := getTestServer(t)

	tests := []struct {
		name       string
		emails     []string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "Multiple valid emails",
			emails:     []string{"user1@example.com", "user2@example.com"},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "Empty list",
			emails:     []string{},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				t.Logf("Subtest '%s' took %v", tt.name, time.Since(start))
			}()

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			reqBody := model.BatchValidationRequest{Emails: tt.emails}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := client.Post(server.URL+"/validate/batch", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			var result model.BatchValidationResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(result.Results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result.Results), tt.wantCount)
			}
		})
	}
}

func TestHandleTypoSuggestions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	server := getTestServer(t)

	tests := []struct {
		name              string
		email             string
		wantStatus        int
		wantHasSuggestion bool
	}{
		{
			name:              "Email with typo",
			email:             "user@gmial.com",
			wantStatus:        http.StatusOK,
			wantHasSuggestion: true,
		},
		{
			name:              "Valid email",
			email:             "user@gmail.com",
			wantStatus:        http.StatusOK,
			wantHasSuggestion: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				t.Logf("Subtest '%s' took %v", tt.name, time.Since(start))
			}()

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			reqBody := model.TypoSuggestionRequest{Email: tt.email}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := client.Post(server.URL+"/typo-suggestions", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			var result model.TypoSuggestionResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			hasSuggestion := len(result.Suggestions) > 0
			if hasSuggestion != tt.wantHasSuggestion {
				t.Errorf("got suggestions = %v, want %v", hasSuggestion, tt.wantHasSuggestion)
			}
		})
	}
}

func TestHandleStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	start := time.Now()
	defer func() {
		t.Logf("Test 'TestHandleStatus' took %v", time.Since(start))
	}()

	server := getTestServer(t)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(server.URL + "/status")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result model.APIStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Status != "healthy" {
		t.Errorf("got status %q, want 'healthy'", result.Status)
	}
}

func TestInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()
	server := getTestServer(t)

	endpoints := []string{
		"/validate",
		"/validate/batch",
		"/typo-suggestions",
	}

	for _, endpoint := range endpoints {
		endpoint := endpoint
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				t.Logf("Subtest '%s' took %v", endpoint, time.Since(start))
			}()

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			resp, err := client.Post(server.URL+endpoint, "application/json", bytes.NewBufferString("invalid json"))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}
