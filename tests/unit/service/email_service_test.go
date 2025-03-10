// Package servicetest contains unit tests for the email service
package servicetest

import (
	"testing"

	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
	"emailvalidator/tests/unit/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEmailValidator is a mock implementation of validator.EmailValidator
type MockEmailValidator struct {
	mock.Mock
	*mocks.MockEmailRuleValidator
	*mocks.MockDomainValidator
}

// ValidateSyntax implements the validator.EmailValidator interface
func (m *MockEmailValidator) ValidateSyntax(email string) bool {
	return m.MockEmailRuleValidator.ValidateSyntax(email)
}

// IsRoleBased implements the validator.EmailValidator interface
func (m *MockEmailValidator) IsRoleBased(email string) bool {
	return m.MockEmailRuleValidator.IsRoleBased(email)
}

// CalculateScore implements the validator.EmailValidator interface
func (m *MockEmailValidator) CalculateScore(validations map[string]bool) int {
	return m.MockEmailRuleValidator.CalculateScore(validations)
}

// GetTypoSuggestions implements the validator.EmailValidator interface
func (m *MockEmailValidator) GetTypoSuggestions(email string) []string {
	return m.MockEmailRuleValidator.GetTypoSuggestions(email)
}

// DetectAlias implements the validator.EmailValidator interface
func (m *MockEmailValidator) DetectAlias(email string) string {
	return m.MockEmailRuleValidator.DetectAlias(email)
}

// ValidateDomain implements the validator.EmailValidator interface
func (m *MockEmailValidator) ValidateDomain(domain string) bool {
	return m.MockDomainValidator.ValidateDomain(domain)
}

// ValidateMXRecords implements the validator.EmailValidator interface
func (m *MockEmailValidator) ValidateMXRecords(domain string) bool {
	return m.MockDomainValidator.ValidateMXRecords(domain)
}

// IsDisposable implements the validator.EmailValidator interface
func (m *MockEmailValidator) IsDisposable(domain string) bool {
	return m.MockDomainValidator.IsDisposable(domain)
}

// Test cases
func TestEmailService_ValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		setup    func(*mocks.MockEmailRuleValidator, *mocks.MockDomainValidationService, *mocks.MockMetricsCollector)
		expected model.EmailValidationResponse
	}{
		{
			name:  "Empty email",
			email: "",
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
			},
			expected: model.EmailValidationResponse{
				Email:  "",
				Status: model.ValidationStatusMissingEmail,
			},
		},
		{
			name:  "Invalid syntax",
			email: "invalid@email",
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
				rv.On("ValidateSyntax", "invalid@email").Return(false)
			},
			expected: model.EmailValidationResponse{
				Email:  "invalid@email",
				Status: model.ValidationStatusInvalidFormat,
			},
		},
		{
			name:  "Valid email",
			email: "test@example.com",
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
				rv.On("ValidateSyntax", "test@example.com").Return(true)
				rv.On("IsRoleBased", "test@example.com").Return(false)
				rv.On("DetectAlias", "test@example.com").Return("")
				dv.On("ValidateDomainConcurrently", mock.Anything, "example.com").Return(true, true, false)
				rv.On("CalculateScore", mock.Anything).Return(95)
				mc.On("RecordValidationScore", "overall", float64(95))
			},
			expected: model.EmailValidationResponse{
				Email: "test@example.com",
				Validations: model.ValidationResults{
					Syntax:        true,
					DomainExists:  true,
					MXRecords:     true,
					IsDisposable:  false,
					IsRoleBased:   false,
					MailboxExists: true,
				},
				Score:  95,
				Status: model.ValidationStatusValid,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRuleValidator := new(mocks.MockEmailRuleValidator)
			mockDomainValidator := new(mocks.MockDomainValidator)
			mockDomainValidationSvc := new(mocks.MockDomainValidationService)
			mockMetricsCollector := new(mocks.MockMetricsCollector)

			tt.setup(mockRuleValidator, mockDomainValidationSvc, mockMetricsCollector)

			// Create a combined validator that implements both interfaces
			mockValidator := &MockEmailValidator{
				MockEmailRuleValidator: mockRuleValidator,
				MockDomainValidator:    mockDomainValidator,
			}

			// Create the service using the existing constructor
			svc := service.NewEmailServiceWithDeps(mockValidator)

			// Replace the dependencies with our mocks
			svc.SetDomainValidationService(mockDomainValidationSvc)
			svc.SetMetricsCollector(mockMetricsCollector)

			// Execute
			result := svc.ValidateEmail(tt.email)

			// Assert
			assert.Equal(t, tt.expected.Email, result.Email)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.Score, result.Score)
			assert.Equal(t, tt.expected.Validations, result.Validations)

			// Verify mocks
			mockRuleValidator.AssertExpectations(t)
			mockDomainValidationSvc.AssertExpectations(t)
			mockMetricsCollector.AssertExpectations(t)
		})
	}
}
