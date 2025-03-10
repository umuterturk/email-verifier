// Package service implements the core business logic of the email validator service.
// It provides email validation, batch processing, and typo suggestion functionality.
package service

import (
	"context"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"emailvalidator/internal/model"
	"emailvalidator/pkg/validator"
)

// EmailService handles email validation operations
type EmailService struct {
	emailRuleValidator  EmailRuleValidator
	domainValidator     DomainValidator
	domainValidationSvc DomainValidationService
	batchValidationSvc  *BatchValidationService
	metricsCollector    MetricsCollector
	startTime           time.Time
	requests            int64
}

// NewEmailService creates a new instance of EmailService
func NewEmailService() (*EmailService, error) {
	emailValidator, err := validator.NewEmailValidator()
	if err != nil {
		return nil, err
	}

	metricsAdapter := NewMetricsAdapter()
	domainValidationSvc := NewConcurrentDomainValidationService(emailValidator)
	batchValidationSvc := NewBatchValidationService(emailValidator, domainValidationSvc, metricsAdapter)

	return &EmailService{
		emailRuleValidator:  emailValidator,
		domainValidator:     emailValidator,
		domainValidationSvc: domainValidationSvc,
		batchValidationSvc:  batchValidationSvc,
		metricsCollector:    metricsAdapter,
		startTime:           time.Now(),
	}, nil
}

// NewEmailServiceWithDeps creates a new instance of EmailService with custom dependencies
// This is primarily used for testing
func NewEmailServiceWithDeps(validator interface{}) *EmailService {
	// Type assertion to get the required interfaces
	var emailRuleValidator EmailRuleValidator
	var domainValidator DomainValidator

	// Try to cast to the required interfaces
	if v, ok := validator.(EmailRuleValidator); ok {
		emailRuleValidator = v
	}
	if v, ok := validator.(DomainValidator); ok {
		domainValidator = v
	}

	metricsAdapter := NewMetricsAdapter()
	domainValidationSvc := NewConcurrentDomainValidationService(domainValidator)
	batchValidationSvc := NewBatchValidationService(emailRuleValidator, domainValidationSvc, metricsAdapter)

	return &EmailService{
		emailRuleValidator:  emailRuleValidator,
		domainValidator:     domainValidator,
		domainValidationSvc: domainValidationSvc,
		batchValidationSvc:  batchValidationSvc,
		metricsCollector:    metricsAdapter,
		startTime:           time.Now(),
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

	// Validate syntax first
	response.Validations.Syntax = s.emailRuleValidator.ValidateSyntax(email)
	if !response.Validations.Syntax {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}

	// Extract domain and validate
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		response.Status = model.ValidationStatusInvalidFormat
		return response
	}
	domain := parts[1]

	// Perform domain validations concurrently
	exists, hasMX, isDisposable := s.domainValidationSvc.ValidateDomainConcurrently(context.Background(), domain)

	// Set validation results
	response.Validations.DomainExists = exists
	response.Validations.MXRecords = hasMX
	response.Validations.IsDisposable = isDisposable
	response.Validations.IsRoleBased = s.emailRuleValidator.IsRoleBased(email)
	response.Validations.MailboxExists = hasMX

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

	// Set status based on validations
	switch {
	case !response.Validations.DomainExists:
		response.Status = model.ValidationStatusInvalidDomain
	case !response.Validations.MXRecords:
		response.Status = model.ValidationStatusNoMXRecords
		response.Score = 40 // Override score for no MX records case
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

// ValidateEmails performs validation on multiple email addresses concurrently
func (s *EmailService) ValidateEmails(emails []string) model.BatchValidationResponse {
	atomic.AddInt64(&s.requests, 1)
	return s.batchValidationSvc.ValidateEmails(emails)
}

// GetTypoSuggestions returns suggestions for possible email typos
func (s *EmailService) GetTypoSuggestions(email string) model.TypoSuggestionResponse {
	atomic.AddInt64(&s.requests, 1)
	suggestions := s.emailRuleValidator.GetTypoSuggestions(email)
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
	s.metricsCollector.UpdateMemoryUsage(float64(m.HeapInuse), float64(m.StackInuse))

	return model.APIStatus{
		Status:            "healthy",
		Uptime:            uptime.String(),
		RequestsHandled:   atomic.LoadInt64(&s.requests),
		AvgResponseTimeMs: 25.0, // This should be calculated based on actual metrics
	}
}

// SetDomainValidationService sets the domain validation service (for testing)
func (s *EmailService) SetDomainValidationService(svc DomainValidationService) {
	s.domainValidationSvc = svc
}

// SetMetricsCollector sets the metrics collector (for testing)
func (s *EmailService) SetMetricsCollector(collector MetricsCollector) {
	s.metricsCollector = collector
}

// SetBatchValidationService sets the batch validation service (for testing)
func (s *EmailService) SetBatchValidationService(svc *BatchValidationService) {
	s.batchValidationSvc = svc
}

// SetEmailRuleValidator sets the email rule validator (for testing)
func (s *EmailService) SetEmailRuleValidator(validator EmailRuleValidator) {
	s.emailRuleValidator = validator
}

// SetDomainValidator sets the domain validator (for testing)
func (s *EmailService) SetDomainValidator(validator DomainValidator) {
	s.domainValidator = validator
}
