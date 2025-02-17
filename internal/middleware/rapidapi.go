// Package middleware provides HTTP middleware components for the email validator service
package middleware

import (
	"net/http"
	"os"
)

// NewRapidAPIAuthMiddleware creates a new RapidAPI authentication middleware with the given secret
func NewRapidAPIAuthMiddleware(next http.Handler, proxySecret string) http.Handler {
	if proxySecret == "" {
		proxySecret = os.Getenv("X_RAPIDAPI_PROXY_SECRET")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the proxy secret from the request header
		headerProxySecret := r.Header.Get("X-RapidAPI-Proxy-Secret")

		// Validate RapidAPI headers
		if headerProxySecret == "" || headerProxySecret != proxySecret {
			http.Error(w, "Invalid or missing RapidAPI proxy secret", http.StatusUnauthorized)
			return
		}

		// Check other RapidAPI headers
		rapidAPIKey := r.Header.Get("X-RapidAPI-Key")
		if rapidAPIKey == "" {
			http.Error(w, "Missing RapidAPI key", http.StatusUnauthorized)
			return
		}

		// You can add additional validation here if needed
		next.ServeHTTP(w, r)
	})
}

// RapidAPIAuthMiddleware handles authentication for RapidAPI using environment variable
// Deprecated: Use NewRapidAPIAuthMiddleware instead
func RapidAPIAuthMiddleware(next http.Handler) http.Handler {
	return NewRapidAPIAuthMiddleware(next, "")
}
