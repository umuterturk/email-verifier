// Package model defines the data structures used throughout the email validator service.
// It defines the request/response models for the API endpoints and internal data representations.
package model

// ValidationStatus represents the status of an email validation
type ValidationStatus string

// Possible validation statuses
const (
	ValidationStatusValid         ValidationStatus = "VALID"
	ValidationStatusProbablyValid ValidationStatus = "PROBABLY_VALID"
	ValidationStatusInvalid       ValidationStatus = "INVALID"
	ValidationStatusMissingEmail  ValidationStatus = "MISSING_EMAIL"
	ValidationStatusInvalidFormat ValidationStatus = "INVALID_FORMAT"
	ValidationStatusInvalidDomain ValidationStatus = "INVALID_DOMAIN"
	ValidationStatusNoMXRecords   ValidationStatus = "NO_MX_RECORDS"
	ValidationStatusDisposable    ValidationStatus = "DISPOSABLE"
)

// ValidationResults represents the results of various validation checks
type ValidationResults struct {
	Syntax        bool `json:"syntax"`
	DomainExists  bool `json:"domain_exists"`
	MXRecords     bool `json:"mx_records"`
	MailboxExists bool `json:"mailbox_exists"`
	IsDisposable  bool `json:"is_disposable"`
	IsRoleBased   bool `json:"is_role_based"`
}

// EmailValidationRequest represents a request to validate a single email
type EmailValidationRequest struct {
	Email string `json:"email"`
}

// EmailValidationResponse represents the response for email validation
type EmailValidationResponse struct {
	Email          string            `json:"email"`
	Validations    ValidationResults `json:"validations"`
	Score          int               `json:"score"`
	Status         ValidationStatus  `json:"status"`
	AliasOf        string            `json:"aliasOf,omitempty"`        // Optional field to indicate if email is an alias
	TypoSuggestion string            `json:"typoSuggestion,omitempty"` // Optional field for typo suggestion
}

// BatchValidationRequest represents a request to validate multiple emails
type BatchValidationRequest struct {
	Emails []string `json:"emails"`
}

// BatchValidationResponse represents the response for batch email validation
type BatchValidationResponse struct {
	Results []EmailValidationResponse `json:"results"`
}

// TypoSuggestionRequest represents a request for email typo suggestions
type TypoSuggestionRequest struct {
	Email string `json:"email"`
}

// TypoSuggestionResponse represents the response for email typo suggestions
type TypoSuggestionResponse struct {
	Email          string `json:"email"`
	TypoSuggestion string `json:"typoSuggestion,omitempty"`
}

// APIStatus represents the current status of the API
type APIStatus struct {
	Status            string  `json:"status"`
	Uptime            string  `json:"uptime"`
	RequestsHandled   int64   `json:"requests_handled"`
	AvgResponseTimeMs float64 `json:"average_response_time_ms"`
}

// CreditInfo represents the credit information for an API key
type CreditInfo struct {
	RemainingCredits int `json:"remaining_credits"`
	TotalCredits     int `json:"total_credits"`
}
