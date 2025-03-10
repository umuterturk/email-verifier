// Package service implements the core business logic of the email validator service.
// It provides email validation, batch processing, and typo suggestion functionality.
package service

import (
	"context"
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
func NewEmailService() (*EmailService, error) {
	validator, err := validator.NewEmailValidator()
	if err != nil {
		return nil, err
	}

	return &EmailService{
		validator: validator,
		startTime: time.Now(),
		workers:   runtime.NumCPU(), // Use number of CPU cores for worker count
	}, nil
}

// NewEmailServiceWithDeps creates a new instance of EmailService with custom dependencies
func NewEmailServiceWithDeps(emailValidator *validator.EmailValidator) *EmailService {
	return &EmailService{
		validator: emailValidator,
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
		// Override score to 40 for no MX records case
		response.Score = 40
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

// performDomainValidations runs domain validation checks concurrently
func (s *EmailService) performDomainValidations(ctx context.Context, domain string) (exists, hasMX, isDisposable bool) {
	// Create a context with shorter timeout for domain validations
	domainCtx, domainCancel := context.WithTimeout(ctx, 5*time.Second)
	defer domainCancel()

	// Perform domain validations concurrently
	var wgDomain sync.WaitGroup
	wgDomain.Add(3)

	// Channel for collecting validation results
	domainResults := make(chan struct {
		validationType string
		isValid        bool
	}, 3)

	go func() {
		defer wgDomain.Done()
		select {
		case <-domainCtx.Done():
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"domain_exists", false}
		default:
			isValid := s.validator.ValidateDomain(domain)
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"domain_exists", isValid}
		}
	}()

	go func() {
		defer wgDomain.Done()
		select {
		case <-domainCtx.Done():
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"mx_records", false}
		default:
			isValid := s.validator.ValidateMXRecords(domain)
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"mx_records", isValid}
		}
	}()

	go func() {
		defer wgDomain.Done()
		select {
		case <-domainCtx.Done():
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"is_disposable", false}
		default:
			isValid := s.validator.IsDisposable(domain)
			domainResults <- struct {
				validationType string
				isValid        bool
			}{"is_disposable", isValid}
		}
	}()

	go func() {
		wgDomain.Wait()
		close(domainResults)
	}()

	// Collect validation results
	for result := range domainResults {
		switch result.validationType {
		case "domain_exists":
			exists = result.isValid
		case "mx_records":
			hasMX = result.isValid
		case "is_disposable":
			isDisposable = result.isValid
		}
	}

	return exists, hasMX, isDisposable
}

// ValidateEmails performs validation on multiple email addresses concurrently
func (s *EmailService) ValidateEmails(emails []string) model.BatchValidationResponse {
	if len(emails) == 0 {
		return model.BatchValidationResponse{Results: []model.EmailValidationResponse{}}
	}

	// Group emails by domain to avoid redundant domain checks
	emailsByDomain := make(map[string][]string)
	emailsWithInvalidFormat := make(map[string]bool)

	for _, email := range emails {
		if email == "" {
			// Handle empty emails separately
			emailsWithInvalidFormat[email] = true
			continue
		}

		// Split email into local part and domain
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			// Handle invalid format emails separately
			emailsWithInvalidFormat[email] = true
			continue
		}

		domain := parts[1]
		emailsByDomain[domain] = append(emailsByDomain[domain], email)
	}

	// Process domain validations once per domain
	domainResults := make(map[string]struct {
		DomainExists bool
		MXRecords    bool
		IsDisposable bool
	})

	// Create a context with timeout for domain validations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Process each domain concurrently
	var wgDomains sync.WaitGroup
	domainChan := make(chan string, len(emailsByDomain))
	resultChan := make(chan struct {
		domain       string
		domainExists bool
		hasMX        bool
		isDisposable bool
	}, len(emailsByDomain))

	// Determine optimal number of workers for domain validation
	domainWorkerCount := minInt(len(emailsByDomain), runtime.NumCPU()*2)
	wgDomains.Add(domainWorkerCount)

	// Start domain validation workers
	for i := 0; i < domainWorkerCount; i++ {
		go func() {
			defer wgDomains.Done()
			for domain := range domainChan {
				exists, hasMX, isDisposable := s.performDomainValidations(ctx, domain)
				resultChan <- struct {
					domain       string
					domainExists bool
					hasMX        bool
					isDisposable bool
				}{domain, exists, hasMX, isDisposable}
			}
		}()
	}

	// Send domains to be processed
	for domain := range emailsByDomain {
		domainChan <- domain
	}
	close(domainChan)

	// Wait for domain validations to complete in a separate goroutine
	go func() {
		wgDomains.Wait()
		close(resultChan)
	}()

	// Collect domain validation results
	for result := range resultChan {
		domainResults[result.domain] = struct {
			DomainExists bool
			MXRecords    bool
			IsDisposable bool
		}{result.domainExists, result.hasMX, result.isDisposable}
	}

	// Process individual emails using the domain validation results
	var wg sync.WaitGroup
	results := make(chan model.EmailValidationResponse, len(emails))

	// Create a channel for emails to be processed
	jobs := make(chan string, len(emails))

	// Determine number of workers for email processing
	workerCount := minInt(len(emails), runtime.NumCPU()*4)
	wg.Add(workerCount)

	// Start workers for email processing
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for email := range jobs {
				// Handle empty emails
				if email == "" {
					results <- model.EmailValidationResponse{
						Email:  email,
						Status: model.ValidationStatusMissingEmail,
					}
					continue
				}

				// Handle invalid format emails
				if emailsWithInvalidFormat[email] {
					results <- model.EmailValidationResponse{
						Email:  email,
						Status: model.ValidationStatusInvalidFormat,
					}
					continue
				}

				// Process valid format emails
				parts := strings.Split(email, "@")
				domain := parts[1]

				// Create response with domain validation results
				response := model.EmailValidationResponse{
					Email:       email,
					Validations: model.ValidationResults{},
				}

				// Check syntax
				response.Validations.Syntax = s.validator.ValidateSyntax(email)
				if !response.Validations.Syntax {
					response.Status = model.ValidationStatusInvalidFormat
					results <- response
					continue
				}

				// Get domain validation results
				domainValidation := domainResults[domain]
				response.Validations.DomainExists = domainValidation.DomainExists
				response.Validations.MXRecords = domainValidation.MXRecords
				response.Validations.IsDisposable = domainValidation.IsDisposable

				// Check if role-based (this is per-email, not per-domain)
				response.Validations.IsRoleBased = s.validator.IsRoleBased(email)
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
					// Override score to 40 for no MX records case
					response.Score = 40
				case response.Validations.IsDisposable:
					response.Status = model.ValidationStatusDisposable
				case response.Score >= 90:
					response.Status = model.ValidationStatusValid
				case response.Score >= 70:
					response.Status = model.ValidationStatusProbablyValid
				default:
					response.Status = model.ValidationStatusInvalid
				}

				results <- response
			}
		}()
	}

	// Send emails to be processed
	for _, email := range emails {
		jobs <- email
	}
	close(jobs)

	// Wait for all workers to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results while preserving order
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
