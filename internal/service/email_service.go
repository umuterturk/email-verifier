// Package service implements the core business logic of the email validator service.
// It provides email validation, batch processing, and typo suggestion functionality.
package service

import (
	"context"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"emailvalidator/internal/config"
	"emailvalidator/internal/model"
	"emailvalidator/pkg/validator"
)

// EmailService handles email validation operations
type EmailService struct {
	emailRuleValidator  EmailRuleValidator
	domainValidator     DomainValidator
	domainValidationSvc DomainValidationService
	smtpVerificationSvc SMTPVerificationService
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

	// Load SMTP config and create verification service
	smtpConfig := config.LoadSMTPVerifierConfig()
	smtpVerificationSvc := NewSMTPVerificationService(smtpConfig)

	batchValidationSvc := NewBatchValidationService(emailValidator, domainValidationSvc, metricsAdapter)

	return &EmailService{
		emailRuleValidator:  emailValidator,
		domainValidator:     emailValidator,
		domainValidationSvc: domainValidationSvc,
		smtpVerificationSvc: smtpVerificationSvc,
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
	response.Validations.MailboxExists = hasMX // Default to MX result, will be updated by SMTP

	// Perform SMTP verification if MX records exist and service is enabled
	if hasMX && s.smtpVerificationSvc != nil && s.smtpVerificationSvc.IsEnabled() {
		mxRecords, err := s.domainValidator.GetMXRecords(domain)
		if err == nil && len(mxRecords) > 0 {
			smtpResult := s.smtpVerificationSvc.VerifyMailbox(context.Background(), email, domain, mxRecords)

			if !smtpResult.Skipped {
				// Set SMTP verification results
				smtpVerified := smtpResult.Verified
				response.Validations.SMTPVerified = &smtpVerified

				isCatchAll := smtpResult.IsCatchAll
				response.Validations.IsCatchAll = &isCatchAll

				// Update MailboxExists based on SMTP result
				if smtpResult.IsCatchAll {
					// Catch-all domains always accept, so we can't truly verify
					response.Validations.MailboxExists = true
				} else if smtpResult.Verified {
					// SMTP confirmed mailbox exists
					response.Validations.MailboxExists = true
				} else if smtpResult.ResponseCode == 550 || smtpResult.ResponseCode == 553 || smtpResult.ResponseCode == 551 {
					// SMTP confirmed mailbox does NOT exist
					response.Validations.MailboxExists = false
				}
				// If error or timeout, keep MailboxExists = hasMX (graceful fallback)
			}
		}
	}

	// Always check for typo suggestions
	suggestions := s.emailRuleValidator.GetTypoSuggestions(email)
	if len(suggestions) > 0 {
		response.TypoSuggestion = suggestions[0]
	}

	// Detect if email is an alias
	if canonicalEmail := s.emailRuleValidator.DetectAlias(email); canonicalEmail != "" && canonicalEmail != email {
		response.AliasOf = canonicalEmail
	}

	// Calculate score
	validationMap := map[string]bool{
		"syntax":         response.Validations.Syntax,
		"domain_exists":  response.Validations.DomainExists,
		"mx_records":     response.Validations.MXRecords,
		"mailbox_exists": response.Validations.MailboxExists,
		"is_disposable":  response.Validations.IsDisposable,
		"is_role_based":  response.Validations.IsRoleBased,
	}

	// Add SMTP verification to score if available
	if response.Validations.SMTPVerified != nil {
		validationMap["smtp_verified"] = *response.Validations.SMTPVerified
	}

	response.Score = s.emailRuleValidator.CalculateScore(validationMap)

	// Apply catch-all penalty (verification is uncertain for catch-all domains)
	if response.Validations.IsCatchAll != nil && *response.Validations.IsCatchAll {
		response.Score = max(0, response.Score-10)
	}

	// Reduce score if there's a typo suggestion
	if response.TypoSuggestion != "" {
		response.Score = max(0, response.Score-20) // Ensure score doesn't go below 0
	}

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
	case response.Validations.IsCatchAll != nil && *response.Validations.IsCatchAll:
		response.Status = model.ValidationStatusCatchAll
	case !response.Validations.MailboxExists:
		// Mailbox doesn't exist (confirmed by SMTP)
		response.Status = model.ValidationStatusInvalid
		response.Score = max(0, response.Score-30) // Heavy penalty for non-existent mailbox
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
	response := model.TypoSuggestionResponse{
		Email: email,
	}
	if len(suggestions) > 0 {
		response.TypoSuggestion = suggestions[0]
	}
	return response
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

// SetSMTPVerificationService sets the SMTP verification service (for testing)
func (s *EmailService) SetSMTPVerificationService(svc SMTPVerificationService) {
	s.smtpVerificationSvc = svc
}
