// Package servicetest contains unit tests for the batch validation service
package servicetest

import (
	"testing"

	"emailvalidator/internal/model"
	"emailvalidator/internal/service"
	"emailvalidator/tests/unit/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBatchValidationService_ValidateEmails(t *testing.T) {
	tests := []struct {
		name     string
		emails   []string
		setup    func(*mocks.MockEmailRuleValidator, *mocks.MockDomainValidationService, *mocks.MockMetricsCollector)
		expected model.BatchValidationResponse
	}{
		{
			name:   "Empty email list",
			emails: []string{},
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
			},
			expected: model.BatchValidationResponse{
				Results: []model.EmailValidationResponse{},
			},
		},
		{
			name:   "Single valid email",
			emails: []string{"test@example.com"},
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
				rv.On("ValidateSyntax", "test@example.com").Return(true)
				rv.On("IsRoleBased", "test@example.com").Return(false)
				rv.On("DetectAlias", "test@example.com").Return("")
				rv.On("GetTypoSuggestions", "test@example.com").Return([]string{})
				dv.On("ValidateDomainConcurrently", mock.Anything, "example.com").Return(true, true, false)
				rv.On("CalculateScore", mock.Anything).Return(95)
				mc.On("RecordValidationScore", "overall", float64(95))
			},
			expected: model.BatchValidationResponse{
				Results: []model.EmailValidationResponse{
					{
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
			},
		},
		{
			name:   "Email with typo",
			emails: []string{"test@gmial.com"},
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
				rv.On("ValidateSyntax", "test@gmial.com").Return(true)
				rv.On("IsRoleBased", "test@gmial.com").Return(false)
				rv.On("DetectAlias", "test@gmial.com").Return("")
				rv.On("GetTypoSuggestions", "test@gmial.com").Return([]string{"test@gmail.com"})
				dv.On("ValidateDomainConcurrently", mock.Anything, "gmial.com").Return(true, true, false)
				rv.On("CalculateScore", mock.Anything).Return(95)
				mc.On("RecordValidationScore", "overall", float64(75)) // 95 - 20 (typo penalty)
			},
			expected: model.BatchValidationResponse{
				Results: []model.EmailValidationResponse{
					{
						Email: "test@gmial.com",
						Validations: model.ValidationResults{
							Syntax:        true,
							DomainExists:  true,
							MXRecords:     true,
							IsDisposable:  false,
							IsRoleBased:   false,
							MailboxExists: true,
						},
						Score:          75, // 95 - 20 (typo penalty)
						Status:         model.ValidationStatusProbablyValid,
						TypoSuggestion: "test@gmail.com",
					},
				},
			},
		},
		{
			name:   "Multiple emails with same domain",
			emails: []string{"test1@example.com", "test2@example.com"},
			setup: func(rv *mocks.MockEmailRuleValidator, dv *mocks.MockDomainValidationService, mc *mocks.MockMetricsCollector) {
				// First email
				rv.On("ValidateSyntax", "test1@example.com").Return(true)
				rv.On("IsRoleBased", "test1@example.com").Return(false)
				rv.On("DetectAlias", "test1@example.com").Return("")
				rv.On("GetTypoSuggestions", "test1@example.com").Return([]string{})

				// Second email
				rv.On("ValidateSyntax", "test2@example.com").Return(true)
				rv.On("IsRoleBased", "test2@example.com").Return(false)
				rv.On("DetectAlias", "test2@example.com").Return("")
				rv.On("GetTypoSuggestions", "test2@example.com").Return([]string{})

				// Domain validation (called once for the domain)
				dv.On("ValidateDomainConcurrently", mock.Anything, "example.com").Return(true, true, false)

				// Score calculations
				rv.On("CalculateScore", mock.Anything).Return(95).Times(2)
				mc.On("RecordValidationScore", "overall", float64(95)).Times(2)
			},
			expected: model.BatchValidationResponse{
				Results: []model.EmailValidationResponse{
					{
						Email: "test1@example.com",
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
					{
						Email: "test2@example.com",
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRuleValidator := new(mocks.MockEmailRuleValidator)
			mockDomainValidationSvc := new(mocks.MockDomainValidationService)
			mockMetricsCollector := new(mocks.MockMetricsCollector)

			tt.setup(mockRuleValidator, mockDomainValidationSvc, mockMetricsCollector)

			// Create service
			svc := service.NewBatchValidationService(
				mockRuleValidator,
				mockDomainValidationSvc,
				mockMetricsCollector,
			)

			// Execute
			result := svc.ValidateEmails(tt.emails)

			// Assert
			assert.Equal(t, len(tt.expected.Results), len(result.Results))
			for i := range tt.expected.Results {
				assert.Equal(t, tt.expected.Results[i].Email, result.Results[i].Email)
				assert.Equal(t, tt.expected.Results[i].Status, result.Results[i].Status)
				assert.Equal(t, tt.expected.Results[i].Score, result.Results[i].Score)
				assert.Equal(t, tt.expected.Results[i].Validations, result.Results[i].Validations)
			}

			// Verify mocks
			mockRuleValidator.AssertExpectations(t)
			mockDomainValidationSvc.AssertExpectations(t)
			mockMetricsCollector.AssertExpectations(t)
		})
	}
}
