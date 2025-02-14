package main

import (
	"log"
	"net/http"

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

	if err := http.ListenAndServe(port, monitoredHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
