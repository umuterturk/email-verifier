// Package main is the entry point for the email validator service.
// It initializes and starts the HTTP server, sets up monitoring, and manages the service lifecycle.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"emailvalidator/internal/api"
	"emailvalidator/internal/middleware"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/monitoring"
)

func main() {
	// Create service instances
	emailService, err := service.NewEmailService()
	if err != nil {
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)

	// Create a new mux for authenticated routes
	authenticatedMux := http.NewServeMux()

	// Register routes that require authentication
	authenticatedMux.HandleFunc("/validate", handler.HandleValidate)
	authenticatedMux.HandleFunc("/validate/batch", handler.HandleBatchValidate)
	authenticatedMux.HandleFunc("/typo-suggestions", handler.HandleTypoSuggestions)

	// Wrap authenticated routes with monitoring middleware and RapidAPI authentication
	monitoredHandler := monitoring.MetricsMiddleware(authenticatedMux)
	authenticatedHandler := middleware.NewRapidAPIAuthMiddleware(monitoredHandler, "")

	// Create final mux that combines both authenticated and unauthenticated routes
	finalMux := http.NewServeMux()

	// Register static file server first (no authentication required)
	fs := http.FileServer(http.Dir("static"))
	finalMux.Handle("/static/", http.StripPrefix("/static/", fs))
	
	// Serve index.html at the root
	finalMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Register public endpoints
	finalMux.Handle("/rapidapi-health", monitoring.MetricsMiddleware(http.HandlerFunc(handler.HandleRapidAPIHealth)))
	finalMux.Handle("/status", monitoring.MetricsMiddleware(http.HandlerFunc(handler.HandleStatus)))
	finalMux.Handle("/metrics", monitoring.MetricsMiddleware(monitoring.PrometheusHandler()))

	// Register API routes with authentication
	// Mount all authenticated routes under /api
	apiMux := http.NewServeMux()
	apiMux.Handle("/", authenticatedHandler)
	finalMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           finalMux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
