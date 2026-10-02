// Package main is the entry point for the email validator service.
// It initializes and starts the HTTP server, sets up monitoring, and manages the service lifecycle.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"emailvalidator/internal/api"
	"emailvalidator/internal/service"
	"emailvalidator/pkg/cache"
	"emailvalidator/pkg/monitoring"
)

func main() {
	// Initialize Redis cache if REDIS_URL is provided
	var redisCache cache.Cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		var err error
		redisCache, err = cache.NewRedisCache(redisURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis at %s: %v. Using in-memory cache instead.", redisURL, err)
			redisCache = nil
		} else {
			log.Printf("Connected to Redis at %s", redisURL)
		}
	} else {
		log.Println("REDIS_URL not set, using in-memory cache")
	}

	// Create service instances
	emailService, err := service.NewEmailServiceWithCache(redisCache)
	if err != nil {
		if redisCache != nil {
			_ = redisCache.Close()
		}
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	// Defer Redis cleanup after all fatal error checks
	if redisCache != nil {
		defer func() {
			if err := redisCache.Close(); err != nil {
				log.Printf("Warning: Failed to close Redis connection: %v", err)
			}
		}()
	}

	// Create and configure HTTP handler
	handler := api.NewHandler(emailService)

	// Create final mux for all routes
	finalMux := http.NewServeMux()

	// Register static file server first
	fs := http.FileServer(http.Dir("static"))
	finalMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve the OpenAPI document so the site docs match the spec in the repo.
	finalMux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "openapi.yaml")
	})

	// Serve index.html at the root
	finalMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Register API endpoints with monitoring
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/validate", handler.HandleValidate)
	apiMux.HandleFunc("/validate/batch", handler.HandleBatchValidate)
	apiMux.HandleFunc("/typo-suggestions", handler.HandleTypoSuggestions)
	apiMux.HandleFunc("/status", handler.HandleStatus)

	// Wrap API routes with monitoring
	monitoredHandler := monitoring.MetricsMiddleware(apiMux)
	finalMux.Handle("/api/", http.StripPrefix("/api", monitoredHandler))

	// Register metrics endpoint
	finalMux.Handle("/metrics", monitoring.MetricsMiddleware(monitoring.PrometheusHandler()))

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
		log.Printf("Server failed: %v", err)
		// Exit with error code - defers will run before exit
	}
}
