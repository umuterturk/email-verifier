// Package middleware provides HTTP middleware components for the email validator service
package middleware

import (
	"fmt"
	"net/http"
	"os"
)

// NewRapidAPIAuthMiddleware creates a new RapidAPI authentication middleware with the given secret
func NewRapidAPIAuthMiddleware(next http.Handler, proxySecret string) http.Handler {
	if proxySecret == "" {
		proxySecret = os.Getenv("X_RAPIDAPI_SECRET")
	}

	skipSecret := os.Getenv("RAPID_API_SKIP_SECRET")
	fmt.Println("skipSecret", skipSecret)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for bypass secret first
		if skipSecret != "" && r.Header.Get("X-API-Skip-Secret") == skipSecret {
			next.ServeHTTP(w, r)
			return
		}

		// Get the proxy secret from the request header
		headerProxySecret := r.Header.Get("X-RapidAPI-Secret")

		// Validate RapidAPI headers
		if headerProxySecret == "" || headerProxySecret != proxySecret {
			http.Error(w, "Invalid or missing RapidAPI proxy secret", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RapidAPIAuthMiddleware handles authentication for RapidAPI using environment variable
// Deprecated: Use NewRapidAPIAuthMiddleware instead
func RapidAPIAuthMiddleware(next http.Handler) http.Handler {
	return NewRapidAPIAuthMiddleware(next, "")
}
