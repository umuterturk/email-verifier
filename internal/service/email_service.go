// Package service implements the core business logic of the email validator service.
// It provides email validation, batch processing, and typo suggestion functionality.
package service

import (
	"context"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"emailvalidator/internal/model"
	"emailvalidator/pkg/cache"
	"emailvalidator/pkg/monitoring"
	"emailvalidator/pkg/validator"
)

// EmailService handles email validation operations
type EmailService struct {
	validator *validator.EmailValidator
	cache     cache.Cache
	startTime time.Time
	requests  int64
	workers   int
}

// NewEmailService creates a new instance of EmailService
func NewEmailService() (*EmailService, error) {
	return NewEmailServiceWithCache(nil)
}

// NewEmailServiceWithCache creates a new instance of EmailService with a specific cache implementation
func NewEmailServiceWithCache(cacheImpl cache.Cache) (*EmailService, error) {
	if cacheImpl == nil {
		redisCache, err := cache.NewRedisCache("localhost:6379")
		if err != nil {
			// Log the error but continue without cache
			log.Printf("Failed to initialize Redis cache: %v", err)
		}
		cacheImpl = redisCache
	}

	validator, err := validator.NewEmailValidator()
	if err != nil {
		return nil, err
	}

	return &EmailService{
		validator: validator,
		cache:     cacheImpl,
		startTime: time.Now(),
		workers:   runtime.NumCPU(), // Use number of CPU cores for worker count
	}, nil
}

// NewEmailServiceWithDeps creates a new instance of EmailService with custom dependencies
func NewEmailServiceWithDeps(cacheImpl cache.Cache, emailValidator *validator.EmailValidator) *EmailService {
	if cacheImpl == nil {
		redisCache, err := cache.NewRedisCache("localhost:6379")
		if err != nil {
			// Log the error but continue without cache
			log.Printf("Failed to initialize Redis cache: %v", err)
		}
		cacheImpl = redisCache
	}

	return &EmailService{
		validator: emailValidator,
		cache:     cacheImpl,
		startTime: time.Now(),
		workers:   runtime.NumCPU(), // Use number of CPU cores for worker count
	}
}

// ValidateEmail performs all validation checks on a single email
func (s *EmailService) ValidateEmail(email string) model.EmailValidationResponse {
	atomic.AddInt64(&s.requests, 1)

	// Try to get from cache first
	if s.cache != nil {
		var cachedResponse model.EmailValidationResponse
		err := s.cache.Get(context.Background(), "email:"+email, &cachedResponse)
		if err == nil {
			monitoring.RecordCacheHit("email_validation")
			return cachedResponse
		}
		monitoring.RecordCacheMiss("email_validation")
	}

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

	// Cache the result
	if s.cache != nil {
		// Cache for 24 hours
		err := s.cache.Set(context.Background(), "email:"+email, response, 24*time.Hour)
		if err != nil {
			log.Printf("Failed to cache validation result: %v", err)
		}
	}

	return response
}

// performDomainValidations runs domain validations concurrently with proper timeout handling
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

	// Wait for all domain validations to complete
	go func() {
		wgDomain.Wait()
		close(domainResults)
	}()

	// Collect domain validation results
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

// worker processes emails from the jobs channel and sends results to the results channel
func (s *EmailService) worker(jobs <-chan string, results chan<- model.EmailValidationResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create a context with timeout for each validation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for email := range jobs {
		// Try to get from cache first
		if s.cache != nil {
			var cachedResponse model.EmailValidationResponse
			err := s.cache.Get(ctx, "email:"+email, &cachedResponse)
			if err == nil {
				monitoring.RecordCacheHit("email_validation")
				results <- cachedResponse
				continue
			}
			monitoring.RecordCacheMiss("email_validation")
		}

		// Perform validation
		response := model.EmailValidationResponse{
			Email:       email,
			Validations: model.ValidationResults{},
		}

		if email == "" {
			response.Status = model.ValidationStatusMissingEmail
			results <- response
			continue
		}

		// Split email into local part and domain
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			response.Status = model.ValidationStatusInvalidFormat
			results <- response
			continue
		}
		domain := parts[1]

		// Perform all validations
		response.Validations.Syntax = s.validator.ValidateSyntax(email)
		if !response.Validations.Syntax {
			response.Status = model.ValidationStatusInvalidFormat
			results <- response
			continue
		}

		// Perform domain validations concurrently
		response.Validations.DomainExists, response.Validations.MXRecords, response.Validations.IsDisposable = s.performDomainValidations(ctx, domain)

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
		case response.Validations.IsDisposable:
			response.Status = model.ValidationStatusDisposable
		case response.Score >= 90:
			response.Status = model.ValidationStatusValid
		case response.Score >= 70:
			response.Status = model.ValidationStatusProbablyValid
		default:
			response.Status = model.ValidationStatusInvalid
		}

		// Cache the result asynchronously with timeout
		if s.cache != nil {
			go func(resp model.EmailValidationResponse) {
				cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cacheCancel()

				err := s.cache.Set(cacheCtx, "email:"+resp.Email, resp, 24*time.Hour)
				if err != nil {
					log.Printf("Failed to cache validation result: %v", err)
				}
			}(response)
		}

		results <- response
	}
}

// ValidateEmails performs validation on multiple email addresses concurrently
func (s *EmailService) ValidateEmails(emails []string) model.BatchValidationResponse {
	if len(emails) == 0 {
		return model.BatchValidationResponse{Results: []model.EmailValidationResponse{}}
	}

	// Create channels for jobs and results with appropriate buffer sizes
	jobs := make(chan string, len(emails))
	results := make(chan model.EmailValidationResponse, len(emails))

	// Create a wait group to track workers
	var wg sync.WaitGroup

	// Determine optimal number of workers based on workload type
	// For IO-bound operations (DNS, Redis), we can use more workers than CPU cores
	workerCount := minInt(len(emails), runtime.NumCPU()*4)
	wg.Add(workerCount)

	// Record goroutine metrics
	monitoring.UpdateGoroutineCount(float64(runtime.NumGoroutine()))

	// Start workers and send jobs concurrently
	go func() {
		for _, email := range emails {
			jobs <- email
		}
		close(jobs)
	}()

	// Start workers
	for i := 0; i < workerCount; i++ {
		go s.worker(jobs, results, &wg)
	}

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
