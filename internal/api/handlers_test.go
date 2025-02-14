package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
)

func setupTestServer() *httptest.Server {
	emailService := service.NewEmailService()
	handler := NewHandler(emailService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func TestHandleValidate(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

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
			name:       "Method not allowed",
			email:      "user@example.com",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if tt.method == http.MethodPost {
				reqBody := model.EmailValidationRequest{Email: tt.email}
				jsonBody, _ := json.Marshal(reqBody)
				resp, err = http.Post(server.URL+"/validate", "application/json", bytes.NewBuffer(jsonBody))
			} else {
				resp, err = http.Get(server.URL + "/validate")
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

			if tt.method == http.MethodPost && resp.StatusCode == http.StatusOK {
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
	server := setupTestServer()
	defer server.Close()

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
		t.Run(tt.name, func(t *testing.T) {
			reqBody := model.BatchValidationRequest{Emails: tt.emails}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(server.URL+"/validate/batch", "application/json", bytes.NewBuffer(jsonBody))
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
	server := setupTestServer()
	defer server.Close()

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
		t.Run(tt.name, func(t *testing.T) {
			reqBody := model.TypoSuggestionRequest{Email: tt.email}
			jsonBody, _ := json.Marshal(reqBody)

			resp, err := http.Post(server.URL+"/typo-suggestions", "application/json", bytes.NewBuffer(jsonBody))
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
	server := setupTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/status")
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
	server := setupTestServer()
	defer server.Close()

	endpoints := []string{
		"/validate",
		"/validate/batch",
		"/typo-suggestions",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Post(server.URL+endpoint, "application/json", bytes.NewBufferString("invalid json"))
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
