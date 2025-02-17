// Package model provides the data structures and types used throughout the email validator service.
// It defines the request/response models for the API endpoints and internal data representations.
package model

// ValidationStatus represents the status of email validation
type ValidationStatus string

const (
	// ValidationStatusValid indicates the email is valid and deliverable
	ValidationStatusValid ValidationStatus = "VALID"
	// ValidationStatusProbablyValid indicates the email is probably valid but has some issues
	ValidationStatusProbablyValid ValidationStatus = "PROBABLY_VALID"
	// ValidationStatusInvalid indicates the email has significant validation issues
	ValidationStatusInvalid ValidationStatus = "INVALID"
	// ValidationStatusMissingEmail indicates no email was provided
	ValidationStatusMissingEmail ValidationStatus = "MISSING_EMAIL"
	// ValidationStatusInvalidFormat indicates invalid email format
	ValidationStatusInvalidFormat ValidationStatus = "INVALID_FORMAT"
	// ValidationStatusInvalidDomain indicates the domain does not exist
	ValidationStatusInvalidDomain ValidationStatus = "INVALID_DOMAIN"
	// ValidationStatusNoMXRecords indicates the domain cannot receive emails
	ValidationStatusNoMXRecords ValidationStatus = "NO_MX_RECORDS"
	// ValidationStatusDisposable indicates a disposable email address
	ValidationStatusDisposable ValidationStatus = "DISPOSABLE"
)

// ValidationResults represents the results of various email validation checks
type ValidationResults struct {
	Syntax        bool `json:"syntax"`
	DomainExists  bool `json:"domain_exists"`
	MXRecords     bool `json:"mx_records"`
	MailboxExists bool `json:"mailbox_exists"`
	IsDisposable  bool `json:"is_disposable"`
	IsRoleBased   bool `json:"is_role_based"`
}

// EmailValidationRequest represents a request to validate an email
type EmailValidationRequest struct {
	Email string `json:"email"`
}

// EmailValidationResponse represents the response for email validation
type EmailValidationResponse struct {
	Email       string            `json:"email"`
	Validations ValidationResults `json:"validations"`
	Score       int               `json:"score"`
	Status      ValidationStatus  `json:"status"`
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
	Email       string   `json:"email"`
	Suggestions []string `json:"suggestions"`
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

// RapidAPIHealth represents the health check response for RapidAPI
type RapidAPIHealth struct {
	Status string `json:"status"`
}
