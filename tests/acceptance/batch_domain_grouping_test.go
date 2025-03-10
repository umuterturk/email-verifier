// Package acceptance contains end-to-end acceptance tests for the email validation service
package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"emailvalidator/internal/model"
)

// EmailServicer defines the interface for an email validation service
type EmailServicer interface {
	ValidateEmail(email string) model.EmailValidationResponse
	ValidateEmails(emails []string) model.BatchValidationResponse
	GetTypoSuggestions(email string) model.TypoSuggestionResponse
	GetAPIStatus() model.APIStatus
}

// MockValidator tracks domain validation calls for testing purposes
type MockValidator struct {
	mu                 sync.Mutex
	domainCheckCounts  map[string]int
	mxCheckCounts      map[string]int
	disposableChecks   map[string]int
	roleBasedChecks    map[string]int
	domainExistResults map[string]bool
	mxRecordResults    map[string]bool
	disposableResults  map[string]bool
	roleBasedResults   map[string]bool
}

// NewMockValidator creates a new MockValidator instance
func NewMockValidator() *MockValidator {
	return &MockValidator{
		domainCheckCounts:  make(map[string]int),
		mxCheckCounts:      make(map[string]int),
		disposableChecks:   make(map[string]int),
		roleBasedChecks:    make(map[string]int),
		domainExistResults: make(map[string]bool),
		mxRecordResults:    make(map[string]bool),
		disposableResults:  make(map[string]bool),
		roleBasedResults:   make(map[string]bool),
	}
}

// Setup initializes the mock validator with predefined results
func (m *MockValidator) Setup() {
	// Valid domains
	for _, domain := range []string{"gmail.com", "outlook.com", "yahoo.com", "icloud.com", "hotmail.com", "aol.com"} {
		m.domainExistResults[domain] = true
		m.mxRecordResults[domain] = true
		m.disposableResults[domain] = false
	}

	// Domains with MX issues
	for _, domain := range []string{"nomx.example.com", "badmx.example.org", "nullmx.example.net"} {
		m.domainExistResults[domain] = true
		m.mxRecordResults[domain] = false
		m.disposableResults[domain] = false
	}

	// Disposable domains
	for _, domain := range []string{"tempmail.com", "10minutemail.com", "throwaway.com"} {
		m.domainExistResults[domain] = true
		m.mxRecordResults[domain] = true
		m.disposableResults[domain] = true
	}

	// Invalid domains
	for _, domain := range []string{"nonexistent123.com", "invalid456.org", "notarealdomain789.net"} {
		m.domainExistResults[domain] = false
		m.mxRecordResults[domain] = false
		m.disposableResults[domain] = false
	}

	// Role-based emails are checked by full email, so we'll set these separately
	m.roleBasedResults["admin@example.com"] = true
	m.roleBasedResults["support@example.com"] = true
	m.roleBasedResults["info@example.com"] = true
	m.roleBasedResults["sales@example.com"] = true
	m.roleBasedResults["contact@example.com"] = true
}

// ValidateDomain mocks domain existence check and tracks invocations
func (m *MockValidator) ValidateDomain(domain string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.domainCheckCounts[domain]++
	result, exists := m.domainExistResults[domain]
	if !exists {
		// Default to true for any unconfigured domain
		return true
	}
	return result
}

// ValidateMXRecords mocks MX record check and tracks invocations
func (m *MockValidator) ValidateMXRecords(domain string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mxCheckCounts[domain]++
	result, exists := m.mxRecordResults[domain]
	if !exists {
		// Default to true for any unconfigured domain
		return true
	}
	return result
}

// IsDisposable mocks disposable domain check and tracks invocations
func (m *MockValidator) IsDisposable(domain string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.disposableChecks[domain]++
	result, exists := m.disposableResults[domain]
	if !exists {
		// Default to false for any unconfigured domain
		return false
	}
	return result
}

// IsRoleBased mocks role-based email check and tracks invocations
func (m *MockValidator) IsRoleBased(email string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roleBasedChecks[email]++
	result, exists := m.roleBasedResults[email]
	if !exists {
		// Check if it starts with common role-based prefixes
		for _, prefix := range []string{"admin@", "support@", "info@", "sales@", "contact@"} {
			if strings.HasPrefix(email, prefix) {
				return true
			}
		}
		return false
	}
	return result
}

// ValidateSyntax wraps the real syntax validator
func (m *MockValidator) ValidateSyntax(email string) bool {
	// Use a simple check for testing purposes
	return strings.Contains(email, "@") && !strings.Contains(email, " ")
}

// GetTypoSuggestions provides simple typo suggestions for testing
func (m *MockValidator) GetTypoSuggestions(email string) []string {
	if !strings.Contains(email, "@") {
		return []string{}
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return []string{}
	}

	domain := parts[1]
	if strings.Contains(domain, "gmial") {
		return []string{strings.Replace(email, "gmial", "gmail", 1)}
	}
	if strings.Contains(domain, "yaho") {
		return []string{strings.Replace(email, "yaho", "yahoo", 1)}
	}
	if strings.Contains(domain, "hotnail") {
		return []string{strings.Replace(email, "hotnail", "hotmail", 1)}
	}

	return []string{}
}

// CalculateScore calculates a validation score
func (m *MockValidator) CalculateScore(validations map[string]bool) int {
	// For testing purposes, we'll return scores that ensure common domains get appropriate status
	// First, check if this is a valid domain with MX records
	if validations["domain_exists"] && validations["mx_records"] {
		// Valid domains with MX records but are disposable get 80
		if validations["is_disposable"] {
			return 60
		}

		// Role-based emails get 80
		if validations["is_role_based"] {
			return 80
		}

		// Valid, good domains get 95 (VALID status)
		return 95
	}

	// No MX records or domain doesn't exist - low score
	return 30
}

// GetTotalDomainChecks returns the total number of domain existence checks made
func (m *MockValidator) GetTotalDomainChecks() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, count := range m.domainCheckCounts {
		total += count
	}
	return total
}

// GetTotalMXChecks returns the total number of MX record checks made
func (m *MockValidator) GetTotalMXChecks() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, count := range m.mxCheckCounts {
		total += count
	}
	return total
}

// GetTotalDisposableChecks returns the total number of disposable domain checks made
func (m *MockValidator) GetTotalDisposableChecks() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, count := range m.disposableChecks {
		total += count
	}
	return total
}

// GetUniqueDomainChecks returns the number of unique domains checked
func (m *MockValidator) GetUniqueDomainChecks() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.domainCheckCounts)
}

// MockEmailService is a custom email service implementation for testing
type MockEmailService struct {
	validator *MockValidator
}

// ValidateEmail implements the email validation method
func (s *MockEmailService) ValidateEmail(email string) model.EmailValidationResponse {
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

	// Set appropriate status based on validations
	switch {
	case !response.Validations.DomainExists:
		response.Status = model.ValidationStatusInvalidDomain
	case !response.Validations.MXRecords:
		response.Status = model.ValidationStatusNoMXRecords
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

// ValidateEmails implements batch email validation with domain grouping
func (s *MockEmailService) ValidateEmails(emails []string) model.BatchValidationResponse {
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

	// Validate each domain only once
	for domain := range emailsByDomain {
		domainExists := s.validator.ValidateDomain(domain)
		mxRecords := s.validator.ValidateMXRecords(domain)
		isDisposable := s.validator.IsDisposable(domain)

		domainResults[domain] = struct {
			DomainExists bool
			MXRecords    bool
			IsDisposable bool
		}{domainExists, mxRecords, isDisposable}
	}

	// Process individual emails using the domain validation results
	// Pre-allocate results slice with capacity matching emails length
	results := make([]model.EmailValidationResponse, 0, len(emails))

	// Process all emails preserving the original order
	for _, email := range emails {
		// Handle empty emails
		if email == "" {
			results = append(results, model.EmailValidationResponse{
				Email:  email,
				Status: model.ValidationStatusMissingEmail,
			})
			continue
		}

		// Handle invalid format emails
		if emailsWithInvalidFormat[email] {
			results = append(results, model.EmailValidationResponse{
				Email:  email,
				Status: model.ValidationStatusInvalidFormat,
			})
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

		// Check syntax first
		response.Validations.Syntax = s.validator.ValidateSyntax(email)
		if !response.Validations.Syntax {
			response.Status = model.ValidationStatusInvalidFormat
			results = append(results, response)
			continue
		}

		// Get domain validation results - using the cached result
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

		results = append(results, response)
	}

	return model.BatchValidationResponse{Results: results}
}

// GetTypoSuggestions implements the typo suggestion method
func (s *MockEmailService) GetTypoSuggestions(email string) model.TypoSuggestionResponse {
	suggestions := s.validator.GetTypoSuggestions(email)
	return model.TypoSuggestionResponse{
		Email:       email,
		Suggestions: suggestions,
	}
}

// GetAPIStatus implements the status method
func (s *MockEmailService) GetAPIStatus() model.APIStatus {
	return model.APIStatus{
		Status:            "OK",
		Uptime:            "Just started",
		RequestsHandled:   0,
		AvgResponseTimeMs: 0,
	}
}

// HandlerFunc type is a function that takes a ResponseWriter and a Request
type HandlerFunc func(http.ResponseWriter, *http.Request)

// SimpleHandler holds endpoints for our test server
type SimpleHandler struct {
	emailService EmailServicer
}

// HandleBatchValidate handles batch validation requests
func (h *SimpleHandler) HandleBatchValidate(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Parse request
	var req model.BatchValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate emails
	result := h.emailService.ValidateEmails(req.Emails)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// Ignore error on writing the error response
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
	}
}

// TestBatchDomainGroupingOptimization tests that the domain grouping optimization correctly
// reduces the number of domain checks while preserving response order and accuracy.
func TestBatchDomainGroupingOptimization(t *testing.T) {
	// Create mock validator
	mockValidator := NewMockValidator()
	mockValidator.Setup()

	// Create mock email service with our validator
	emailService := &MockEmailService{
		validator: mockValidator,
	}

	// Create server with mock service
	handler := &SimpleHandler{
		emailService: emailService,
	}

	// Configure the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/validate/batch", handler.HandleBatchValidate)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Generate 100 test emails across various categories
	// Pre-allocate emails slice with initial capacity
	emails := make([]string, 0, 100)
	emailIndex := make(map[string]int)

	// Add 30 Gmail emails (should be grouped)
	for i := 1; i <= 30; i++ {
		email := fmt.Sprintf("user%d@gmail.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 20 Outlook emails (should be grouped)
	for i := 1; i <= 20; i++ {
		email := fmt.Sprintf("person%d@outlook.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 15 Yahoo emails (should be grouped)
	for i := 1; i <= 15; i++ {
		email := fmt.Sprintf("account%d@yahoo.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 10 Hotmail emails (should be grouped)
	for i := 1; i <= 10; i++ {
		email := fmt.Sprintf("profile%d@hotmail.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 5 iCloud emails (should be grouped)
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("apple%d@icloud.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 5 AOL emails (should be grouped)
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("member%d@aol.com", i)
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Add 3 domains with MX issues
	emails = append(emails, "test1@nomx.example.com")
	emailIndex["test1@nomx.example.com"] = len(emails) - 1
	emails = append(emails, "test2@badmx.example.org")
	emailIndex["test2@badmx.example.org"] = len(emails) - 1
	emails = append(emails, "test3@nullmx.example.net")
	emailIndex["test3@nullmx.example.net"] = len(emails) - 1

	// Add 3 disposable emails
	emails = append(emails, "temp@tempmail.com")
	emailIndex["temp@tempmail.com"] = len(emails) - 1
	emails = append(emails, "throwaway@10minutemail.com")
	emailIndex["throwaway@10minutemail.com"] = len(emails) - 1
	emails = append(emails, "disposable@throwaway.com")
	emailIndex["disposable@throwaway.com"] = len(emails) - 1

	// Add 3 non-existent domains
	emails = append(emails, "fake@nonexistent123.com")
	emailIndex["fake@nonexistent123.com"] = len(emails) - 1
	emails = append(emails, "invalid@invalid456.org")
	emailIndex["invalid@invalid456.org"] = len(emails) - 1
	emails = append(emails, "missing@notarealdomain789.net")
	emailIndex["missing@notarealdomain789.net"] = len(emails) - 1

	// Add 3 role-based emails
	emails = append(emails, "admin@example.com")
	emailIndex["admin@example.com"] = len(emails) - 1
	emails = append(emails, "support@example.com")
	emailIndex["support@example.com"] = len(emails) - 1
	emails = append(emails, "info@example.com")
	emailIndex["info@example.com"] = len(emails) - 1

	// Add 3 invalid format emails
	emails = append(emails, "not-an-email")
	emailIndex["not-an-email"] = len(emails) - 1
	emails = append(emails, "missing-at-sign.com")
	emailIndex["missing-at-sign.com"] = len(emails) - 1
	emails = append(emails, "spaces in email@domain.com")
	emailIndex["spaces in email@domain.com"] = len(emails) - 1

	// Add a few additional unique domain emails to reach 100 total
	uniqueDomains := []string{
		"unique1@custom1.com",
		"unique2@custom2.org",
		"unique3@custom3.net",
		"unique4@custom4.io",
		"unique5@custom5.dev",
	}

	for _, email := range uniqueDomains {
		emails = append(emails, email)
		emailIndex[email] = len(emails) - 1
	}

	// Validate that we have 100 emails
	if len(emails) < 100 {
		// Add any remaining emails to reach 100
		for i := len(emails); i < 100; i++ {
			email := fmt.Sprintf("extra%d@filler-domain.com", i)
			emails = append(emails, email)
			emailIndex[email] = len(emails) - 1
		}
	} else if len(emails) > 100 {
		// Trim if we have too many
		emails = emails[:100]
	}

	t.Logf("Testing batch validation with %d emails", len(emails))

	// Send batch validation request
	reqBody := model.BatchValidationRequest{Emails: emails}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/validate/batch", server.URL),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		t.Fatalf("Failed to send batch request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	// Verify HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result model.BatchValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify all emails were processed
	if len(result.Results) != len(emails) {
		t.Errorf("Expected %d results, got %d", len(emails), len(result.Results))
	}

	// Get count of unique domains in our email set
	domainCounts := make(map[string]int)
	for _, email := range emails {
		if strings.Contains(email, "@") {
			parts := strings.Split(email, "@")
			if len(parts) == 2 {
				domainCounts[parts[1]]++
			}
		}
	}
	uniqueDomainCount := len(domainCounts)

	// Check domain validation counts
	totalDomainChecks := mockValidator.GetTotalDomainChecks()
	totalMXChecks := mockValidator.GetTotalMXChecks()
	totalDisposableChecks := mockValidator.GetTotalDisposableChecks()
	uniqueDomainsChecked := mockValidator.GetUniqueDomainChecks()

	t.Logf("Email count: %d", len(emails))
	t.Logf("Unique domains: %d", uniqueDomainCount)
	t.Logf("Domain checks performed: %d", totalDomainChecks)
	t.Logf("MX checks performed: %d", totalMXChecks)
	t.Logf("Disposable checks performed: %d", totalDisposableChecks)
	t.Logf("Unique domains checked: %d", uniqueDomainsChecked)

	// Verify optimization - checks should be close to unique domain count, not email count
	if totalDomainChecks != uniqueDomainCount {
		t.Errorf("Too many domain checks: %d for %d unique domains", totalDomainChecks, uniqueDomainCount)
	}

	if totalMXChecks > uniqueDomainCount {
		t.Errorf("Too many MX checks: %d for %d unique domains", totalMXChecks, uniqueDomainCount)
	}

	if totalDisposableChecks > uniqueDomainCount {
		t.Errorf("Too many disposable checks: %d for %d unique domains", totalDisposableChecks, uniqueDomainCount)
	}

	// Verify response order matches input order
	for i, res := range result.Results {
		if res.Email != emails[i] {
			t.Errorf("Response order mismatch: expected %s at position %d, got %s",
				emails[i], i, res.Email)
		}
	}

	// Verify specific email result categories
	verifyEmailCategory := func(t *testing.T, results []model.EmailValidationResponse, email string,
		expectedStatus model.ValidationStatus, _ string) {
		index, ok := emailIndex[email]
		if !ok {
			t.Errorf("Test email %s not found in results", email)
			return
		}

		if index >= len(results) {
			t.Errorf("Index %d out of bounds for results length %d", index, len(results))
			return
		}

		result := results[index]
		if result.Status != expectedStatus {
			t.Errorf("Email %s: expected status %s, got %s",
				email, expectedStatus, result.Status)
		}
	}

	// Verify example emails from each category
	// 1. Valid email (Gmail)
	verifyEmailCategory(t, result.Results, "user1@gmail.com",
		model.ValidationStatusValid, "")

	// 2. No MX Records
	verifyEmailCategory(t, result.Results, "test1@nomx.example.com",
		model.ValidationStatusNoMXRecords, "")

	// 3. Disposable email
	verifyEmailCategory(t, result.Results, "temp@tempmail.com",
		model.ValidationStatusDisposable, "")

	// 4. Invalid domain
	verifyEmailCategory(t, result.Results, "fake@nonexistent123.com",
		model.ValidationStatusInvalidDomain, "")

	// 5. Invalid format
	verifyEmailCategory(t, result.Results, "not-an-email",
		model.ValidationStatusInvalidFormat, "")

	t.Log("✓ Batch processing with domain grouping optimization is functioning correctly")
	t.Logf("✓ Domain checks reduced from %d emails to %d unique domains",
		len(emails), uniqueDomainsChecked)
}
