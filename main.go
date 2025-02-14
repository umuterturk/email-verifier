// Package main is the entry point for the email validator service.
// It initializes and starts the HTTP server, sets up monitoring, and manages the service lifecycle.
package main

import (
	"log"
	"net/http"
	"time"

	"emailvalidator/internal/api"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/monitoring"
)

func main() {
	// Create service instances
	emailService := service.NewEmailService()

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", monitoring.PrometheusHandler())

	// Wrap all routes with monitoring middleware
	monitoredHandler := monitoring.MetricsMiddleware(mux)

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s", port)
	log.Printf("Metrics available at http://localhost%s/metrics", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           monitoredHandler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
