package model

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
	Message     string            `json:"message"`
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
