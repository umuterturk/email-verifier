// Package service implements the core business logic of the email validator service.
// It provides email validation, batch processing, and typo suggestion functionality.
package service

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"emailvalidator/internal/model"
	"emailvalidator/pkg/monitoring"
	"emailvalidator/pkg/validator"
)

// EmailService handles email validation operations
type EmailService struct {
	validator *validator.EmailValidator
	startTime time.Time
	requests  int64
	workers   int
}

// NewEmailService creates a new instance of EmailService
func NewEmailService() *EmailService {
	return &EmailService{
		validator: validator.NewEmailValidator(),
		startTime: time.Now(),
		workers:   runtime.NumCPU(), // Use number of CPU cores for worker count
	}
}

// ValidateEmail performs all validation checks on a single email
func (s *EmailService) ValidateEmail(email string) model.EmailValidationResponse {
	atomic.AddInt64(&s.requests, 1)

	response := model.EmailValidationResponse{
		Email:       email,
		Validations: model.ValidationResults{},
	}

	if email == "" {
		response.Status = model.ValidationStatusMissingEmail
		return response
	}

	// Split email into local part and domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}
	domain := parts[1]

	// Perform all validations
	response.Validations.Syntax = s.validator.ValidateSyntax(email)
	if !response.Validations.Syntax {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}

	// Perform domain validations
	response.Validations.DomainExists = s.validator.ValidateDomain(domain)
	response.Validations.MXRecords = s.validator.ValidateMXRecords(domain)
	response.Validations.IsDisposable = s.validator.IsDisposable(domain)
	response.Validations.IsRoleBased = s.validator.IsRoleBased(email)

	// Calculate mailbox existence based on MX records
	response.Validations.MailboxExists = response.Validations.MXRecords

	// Calculate score
	validationMap := map[string]bool{
		"syntax":         response.Validations.Syntax,
		"domain_exists":  response.Validations.DomainExists,
		"mx_records":     response.Validations.MXRecords,
		"mailbox_exists": response.Validations.MailboxExists,
		"is_disposable":  response.Validations.IsDisposable,
		"is_role_based":  response.Validations.IsRoleBased,
	}
	response.Score = s.validator.CalculateScore(validationMap)

	// Record validation score
	monitoring.RecordValidationScore("overall", float64(response.Score))

	// Set appropriate status based on validations
	switch {
	case !response.Validations.DomainExists:
		response.Status = model.ValidationStatusInvalidDomain
	case !response.Validations.MXRecords:
		response.Status = model.ValidationStatusNoMXRecords
	case response.Validations.IsDisposable:
		response.Status = model.ValidationStatusDisposable
	case response.Score >= 90:
		response.Status = model.ValidationStatusValid
	case response.Score >= 70:
		response.Status = model.ValidationStatusProbablyValid
	default:
		response.Status = model.ValidationStatusInvalid
	}

	return response
}

// worker processes emails from the jobs channel and sends results to the results channel
func (s *EmailService) worker(jobs <-chan string, results chan<- model.EmailValidationResponse, wg *sync.WaitGroup) {
	defer wg.Done()
	for email := range jobs {
		results <- s.ValidateEmail(email)
	}
}

// ValidateEmails performs validation on multiple email addresses concurrently
func (s *EmailService) ValidateEmails(emails []string) model.BatchValidationResponse {
	if len(emails) == 0 {
		return model.BatchValidationResponse{Results: []model.EmailValidationResponse{}}
	}

	// Create channels for jobs and results
	jobs := make(chan string, len(emails))
	results := make(chan model.EmailValidationResponse, len(emails))

	// Create a wait group to track workers
	var wg sync.WaitGroup

	// Determine number of workers (use min of emails count and available workers)
	workerCount := minInt(s.workers, len(emails))
	wg.Add(workerCount)

	// Record goroutine metrics
	monitoring.UpdateGoroutineCount(float64(runtime.NumGoroutine()))

	// Start workers
	for i := 0; i < workerCount; i++ {
		go s.worker(jobs, results, &wg)
	}

	// Send jobs
	for _, email := range emails {
		jobs <- email
	}
	close(jobs)

	// Wait for all workers to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	var response model.BatchValidationResponse
	resultsMap := make(map[string]model.EmailValidationResponse)
	for result := range results {
		resultsMap[result.Email] = result
	}

	// Preserve original order
	for _, email := range emails {
		response.Results = append(response.Results, resultsMap[email])
	}

	return response
}

// GetTypoSuggestions returns suggestions for possible email typos
func (s *EmailService) GetTypoSuggestions(email string) model.TypoSuggestionResponse {
	atomic.AddInt64(&s.requests, 1)
	suggestions := s.validator.GetTypoSuggestions(email)
	return model.TypoSuggestionResponse{
		Email:       email,
		Suggestions: suggestions,
	}
}

// GetAPIStatus returns the current status of the API
func (s *EmailService) GetAPIStatus() model.APIStatus {
	uptime := time.Since(s.startTime)

	// Update memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	monitoring.UpdateMemoryUsage(float64(m.HeapInuse), float64(m.StackInuse))

	return model.APIStatus{
		Status:            "healthy",
		Uptime:            uptime.String(),
		RequestsHandled:   atomic.LoadInt64(&s.requests),
		AvgResponseTimeMs: 25.0, // This should be calculated based on actual metrics
	}
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
