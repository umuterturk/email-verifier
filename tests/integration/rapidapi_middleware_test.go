package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"emailvalidator/internal/middleware"
)

func TestRapidAPIAuthMiddleware(t *testing.T) {
	// Use a mock secret for testing
	//nolint:gosec // This is a mock secret for testing purposes
	const expectedSecret = "test-proxy-secret"

	// Create a simple test handler that always returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		rapidAPIKey    string
		proxySecret    string
		expectedStatus int
	}{
		{
			name:           "Valid headers",
			rapidAPIKey:    "valid-api-key",
			proxySecret:    expectedSecret,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing RapidAPI Key",
			rapidAPIKey:    "",
			proxySecret:    expectedSecret,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid proxy secret",
			rapidAPIKey:    "valid-api-key",
			proxySecret:    "invalid-secret",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing proxy secret",
			rapidAPIKey:    "valid-api-key",
			proxySecret:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest("GET", "/test", nil)

			// Set headers
			if tt.rapidAPIKey != "" {
				req.Header.Set("X-RapidAPI-Key", tt.rapidAPIKey)
			}
			if tt.proxySecret != "" {
				req.Header.Set("X-RapidAPI-Secret", tt.proxySecret)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create the middleware handler with the mock secret
			handler := middleware.NewRapidAPIAuthMiddleware(testHandler, expectedSecret)

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
