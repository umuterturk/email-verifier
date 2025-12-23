package service

import (
	"context"
	"net"

	"emailvalidator/internal/model"
	"emailvalidator/pkg/validator"
)

// EmailValidator defines the contract for email validation operations
type EmailValidator interface {
	ValidateEmail(email string) model.EmailValidationResponse
	ValidateEmails(emails []string) model.BatchValidationResponse
	GetTypoSuggestions(email string) model.TypoSuggestionResponse
}

// DomainValidator defines the contract for domain-specific validations
type DomainValidator interface {
	ValidateDomain(domain string) bool
	ValidateMXRecords(domain string) bool
	IsDisposable(domain string) bool
	GetMXRecords(domain string) ([]*net.MX, error)
}

// EmailRuleValidator defines the contract for email-specific rule validations
type EmailRuleValidator interface {
	ValidateSyntax(email string) bool
	IsRoleBased(email string) bool
	CalculateScore(validations map[string]bool) int
	GetTypoSuggestions(email string) []string
	DetectAlias(email string) string
}

// MetricsCollector defines the contract for collecting service metrics
type MetricsCollector interface {
	RecordValidationScore(name string, score float64)
	UpdateMemoryUsage(heapInUse, stackInUse float64)
}

// DomainValidationService defines the contract for concurrent domain validation operations
type DomainValidationService interface {
	ValidateDomainConcurrently(ctx context.Context, domain string) (exists, hasMX, isDisposable bool)
}

// AliasDetector defines the contract for detecting email aliases
type AliasDetector interface {
	// DetectAlias checks if the email is an alias and returns the canonical email if it is
	// Returns empty string if the email is not an alias
	DetectAlias(email string) string
}

// SMTPVerificationService defines the contract for SMTP mailbox verification
type SMTPVerificationService interface {
	// VerifyMailbox performs SMTP RCPT TO verification to check if mailbox exists
	VerifyMailbox(ctx context.Context, email, domain string, mxRecords []*net.MX) validator.SMTPVerificationResult
	// IsCatchAll checks if domain is a catch-all (accepts all addresses)
	IsCatchAll(ctx context.Context, domain string, mxRecords []*net.MX) (bool, error)
	// IsEnabled returns whether SMTP verification is enabled
	IsEnabled() bool
}
