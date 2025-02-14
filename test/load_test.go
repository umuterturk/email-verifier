package test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"sync"
	"testing"
	"time"

	"emailvalidator/internal/model"
)

var (
	duration        = flag.Duration("duration", 30*time.Second, "Duration for the load test")
	short           = flag.Bool("short", true, "Run short version of test")
	concurrentUsers = flag.Int("concurrent-users", 1, "Number of concurrent users")
)

type TestMetrics struct {
	sync.Mutex
	RequestCount   int
	SuccessCount   int
	FailureCount   int
	TotalDuration  time.Duration
	MaxDuration    time.Duration
	MinDuration    time.Duration
	StatusCodes    map[int]int
	EndpointCounts map[string]int
}

func NewTestMetrics() *TestMetrics {
	return &TestMetrics{
		StatusCodes:    make(map[int]int),
		EndpointCounts: make(map[string]int),
		MinDuration:    time.Hour, // Start with a large value
	}
}

func (m *TestMetrics) RecordRequest(endpoint string, duration time.Duration, statusCode int) {
	m.Lock()
	defer m.Unlock()

	m.RequestCount++
	m.TotalDuration += duration
	m.StatusCodes[statusCode]++
	m.EndpointCounts[endpoint]++

	if duration > m.MaxDuration {
		m.MaxDuration = duration
	}
	if duration < m.MinDuration {
		m.MinDuration = duration
	}

	if statusCode >= 200 && statusCode < 300 {
		m.SuccessCount++
	} else {
		m.FailureCount++
	}
}

func TestLoadGeneration(t *testing.T) {
	flag.Parse()

	if *short {
		t.Log("Running short test (30s)")
		*duration = 30 * time.Second
	}

	metrics := NewTestMetrics()

	// Test emails with different characteristics
	emails := []string{
		"valid@example.com",
		"invalid-email",
		"user@nonexistent.com",
		"admin@company.com",
		"test@disposable.com",
		"user@gmial.com", // Typo
	}

	t.Logf("Running load test for %v with %d concurrent users", *duration, *concurrentUsers)
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < *concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			runUserWorkload(ctx, t, emails, metrics, userID)
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()

	// Log results
	t.Logf("Load Test Results:")
	t.Logf("Total Requests: %d", metrics.RequestCount)
	t.Logf("Success Rate: %.2f%%", float64(metrics.SuccessCount)/float64(metrics.RequestCount)*100)
	t.Logf("Average Duration: %v", metrics.TotalDuration/time.Duration(metrics.RequestCount))
	t.Logf("Max Duration: %v", metrics.MaxDuration)
	t.Logf("Min Duration: %v", metrics.MinDuration)

	t.Logf("\nEndpoint Distribution:")
	for endpoint, count := range metrics.EndpointCounts {
		t.Logf("- %s: %d requests", endpoint, count)
	}

	t.Logf("\nStatus Code Distribution:")
	for code, count := range metrics.StatusCodes {
		t.Logf("- %d: %d requests", code, count)
	}
}

func runUserWorkload(ctx context.Context, t *testing.T, emails []string, metrics *TestMetrics, userID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, email := range emails {
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					// Test single validation
					start := time.Now()
					code := validateSingle(t, email)
					metrics.RecordRequest("/validate", time.Since(start), code)

					// Test batch validation
					start = time.Now()
					code = validateBatch(t, []string{email, "another@example.com"})
					metrics.RecordRequest("/validate/batch", time.Since(start), code)

					// Test typo suggestions
					start = time.Now()
					code = getTypoSuggestions(t, email)
					metrics.RecordRequest("/typo-suggestions", time.Since(start), code)

					// Check API status
					start = time.Now()
					code = getStatus(t)
					metrics.RecordRequest("/status", time.Since(start), code)
				}
			}
		}
	}
}

func validateSingle(t *testing.T, email string) int {
	reqBody := model.EmailValidationRequest{Email: email}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:8080/validate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Logf("Failed to validate email: %v", err)
		return http.StatusInternalServerError
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func validateBatch(t *testing.T, emails []string) int {
	reqBody := model.BatchValidationRequest{Emails: emails}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:8080/validate/batch", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Logf("Failed to validate batch: %v", err)
		return http.StatusInternalServerError
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func getTypoSuggestions(t *testing.T, email string) int {
	reqBody := model.TypoSuggestionRequest{Email: email}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:8080/typo-suggestions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Logf("Failed to get typo suggestions: %v", err)
		return http.StatusInternalServerError
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func getStatus(t *testing.T) int {
	resp, err := http.Get("http://127.0.0.1:8080/status")
	if err != nil {
		t.Logf("Failed to get status: %v", err)
		return http.StatusInternalServerError
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
