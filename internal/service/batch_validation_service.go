package service

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"emailvalidator/internal/model"
	"emailvalidator/internal/utils"
)

// BatchValidationService handles batch email validation operations
type BatchValidationService struct {
	emailRuleValidator   EmailRuleValidator
	domainValidationSvc  DomainValidationService
	metricsCollector     MetricsCollector
	maxConcurrentWorkers int
}

// NewBatchValidationService creates a new instance of BatchValidationService
func NewBatchValidationService(
	ruleValidator EmailRuleValidator,
	domainValidationSvc DomainValidationService,
	metricsCollector MetricsCollector,
) *BatchValidationService {
	return &BatchValidationService{
		emailRuleValidator:   ruleValidator,
		domainValidationSvc:  domainValidationSvc,
		metricsCollector:     metricsCollector,
		maxConcurrentWorkers: runtime.NumCPU() * 4,
	}
}

// ValidateEmails performs validation on multiple email addresses concurrently
func (s *BatchValidationService) ValidateEmails(emails []string) model.BatchValidationResponse {
	if len(emails) == 0 {
		return model.BatchValidationResponse{Results: []model.EmailValidationResponse{}}
	}

	// Group emails by domain
	emailsByDomain := s.groupEmailsByDomain(emails)

	// Process domain validations
	domainResults := s.processDomainValidations(emailsByDomain)

	// Process individual emails
	response := s.processEmails(emails, emailsByDomain, domainResults)

	return response
}

func (s *BatchValidationService) groupEmailsByDomain(emails []string) map[string][]string {
	emailsByDomain := make(map[string][]string)
	for _, email := range emails {
		if email == "" {
			continue
		}

		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			continue
		}

		domain := parts[1]
		emailsByDomain[domain] = append(emailsByDomain[domain], email)
	}
	return emailsByDomain
}

func (s *BatchValidationService) processDomainValidations(emailsByDomain map[string][]string) map[string]struct {
	DomainExists bool
	MXRecords    bool
	IsDisposable bool
} {
	ctx := context.Background()
	domainResults := make(map[string]struct {
		DomainExists bool
		MXRecords    bool
		IsDisposable bool
	})

	var wg sync.WaitGroup
	resultChan := make(chan struct {
		domain       string
		domainExists bool
		hasMX        bool
		isDisposable bool
	}, len(emailsByDomain))

	// Process domains concurrently
	for domain := range emailsByDomain {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			exists, hasMX, isDisposable := s.domainValidationSvc.ValidateDomainConcurrently(ctx, d)
			resultChan <- struct {
				domain       string
				domainExists bool
				hasMX        bool
				isDisposable bool
			}{d, exists, hasMX, isDisposable}
		}(domain)
	}

	// Close results channel after all goroutines complete
	go func() {
		wg.Wait()
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

	return domainResults
}

func (s *BatchValidationService) processEmails(
	emails []string,
	emailsByDomain map[string][]string,
	domainResults map[string]struct {
		DomainExists bool
		MXRecords    bool
		IsDisposable bool
	},
) model.BatchValidationResponse {
	var response model.BatchValidationResponse
	resultsMap := make(map[string]model.EmailValidationResponse)

	jobs := make(chan string, len(emails))
	results := make(chan model.EmailValidationResponse, len(emails))

	// Start workers
	workerCount := utils.MinInt(len(emails), s.maxConcurrentWorkers)
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go s.emailValidationWorker(&wg, jobs, results, emailsByDomain, domainResults)
	}

	// Send jobs
	for _, email := range emails {
		jobs <- email
	}
	close(jobs)

	// Wait for completion and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		resultsMap[result.Email] = result
	}

	// Preserve original order
	for _, email := range emails {
		response.Results = append(response.Results, resultsMap[email])
	}

	return response
}

func (s *BatchValidationService) emailValidationWorker(
	wg *sync.WaitGroup,
	jobs <-chan string,
	results chan<- model.EmailValidationResponse,
	emailsByDomain map[string][]string,
	domainResults map[string]struct {
		DomainExists bool
		MXRecords    bool
		IsDisposable bool
	},
) {
	defer wg.Done()

	for email := range jobs {
		response := s.validateSingleEmail(email, domainResults)
		results <- response
	}
}

func (s *BatchValidationService) validateSingleEmail(
	email string,
	domainResults map[string]struct {
		DomainExists bool
		MXRecords    bool
		IsDisposable bool
	},
) model.EmailValidationResponse {
	response := model.EmailValidationResponse{
		Email:       email,
		Validations: model.ValidationResults{},
	}

	if email == "" {
		response.Status = model.ValidationStatusMissingEmail
		return response
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}

	domain := parts[1]
	response.Validations.Syntax = s.emailRuleValidator.ValidateSyntax(email)
	if !response.Validations.Syntax {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}

	// Get domain validation results
	domainValidation := domainResults[domain]
	response.Validations.DomainExists = domainValidation.DomainExists
	response.Validations.MXRecords = domainValidation.MXRecords
	response.Validations.IsDisposable = domainValidation.IsDisposable
	response.Validations.IsRoleBased = s.emailRuleValidator.IsRoleBased(email)
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
	response.Score = s.emailRuleValidator.CalculateScore(validationMap)

	// Record validation score
	s.metricsCollector.RecordValidationScore("overall", float64(response.Score))

	// Set status
	response.Status = s.determineValidationStatus(response)

	return response
}

func (s *BatchValidationService) determineValidationStatus(response model.EmailValidationResponse) model.ValidationStatus {
	switch {
	case !response.Validations.DomainExists:
		return model.ValidationStatusInvalidDomain
	case !response.Validations.MXRecords:
		response.Score = 40 // Override score for no MX records case
		return model.ValidationStatusNoMXRecords
	case response.Validations.IsDisposable:
		return model.ValidationStatusDisposable
	case response.Score >= 90:
		return model.ValidationStatusValid
	case response.Score >= 70:
		return model.ValidationStatusProbablyValid
	default:
		return model.ValidationStatusInvalid
	}
}
