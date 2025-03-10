// Package servicetest contains unit tests for the domain validation service
package servicetest

import (
	"context"
	"testing"
	"time"

	"emailvalidator/internal/service"
	"emailvalidator/tests/unit/service/mocks"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentDomainValidationService_ValidateDomainConcurrently(t *testing.T) {
	tests := []struct {
		name            string
		domain          string
		timeout         time.Duration
		setup           func(*mocks.MockDomainValidator)
		expectedExists  bool
		expectedHasMX   bool
		expectedDispose bool
	}{
		{
			name:    "All validations pass",
			domain:  "example.com",
			timeout: 5 * time.Second,
			setup: func(mv *mocks.MockDomainValidator) {
				mv.On("ValidateDomain", "example.com").Return(true)
				mv.On("ValidateMXRecords", "example.com").Return(true)
				mv.On("IsDisposable", "example.com").Return(false)
			},
			expectedExists:  true,
			expectedHasMX:   true,
			expectedDispose: false,
		},
		{
			name:    "Domain does not exist",
			domain:  "nonexistent.com",
			timeout: 5 * time.Second,
			setup: func(mv *mocks.MockDomainValidator) {
				mv.On("ValidateDomain", "nonexistent.com").Return(false)
				mv.On("ValidateMXRecords", "nonexistent.com").Return(false)
				mv.On("IsDisposable", "nonexistent.com").Return(false)
			},
			expectedExists:  false,
			expectedHasMX:   false,
			expectedDispose: false,
		},
		{
			name:    "Disposable email domain",
			domain:  "temp.com",
			timeout: 5 * time.Second,
			setup: func(mv *mocks.MockDomainValidator) {
				mv.On("ValidateDomain", "temp.com").Return(true)
				mv.On("ValidateMXRecords", "temp.com").Return(true)
				mv.On("IsDisposable", "temp.com").Return(true)
			},
			expectedExists:  true,
			expectedHasMX:   true,
			expectedDispose: true,
		},
		{
			name:    "Context timeout",
			domain:  "slow.com",
			timeout: 1 * time.Millisecond,
			setup: func(mv *mocks.MockDomainValidator) {
				mv.On("ValidateDomain", "slow.com").After(10 * time.Millisecond).Return(true)
				mv.On("ValidateMXRecords", "slow.com").After(10 * time.Millisecond).Return(true)
				mv.On("IsDisposable", "slow.com").After(10 * time.Millisecond).Return(false)
			},
			expectedExists:  false,
			expectedHasMX:   false,
			expectedDispose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockValidator := new(mocks.MockDomainValidator)
			tt.setup(mockValidator)

			// Create service
			svc := service.NewConcurrentDomainValidationService(mockValidator)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Execute
			exists, hasMX, isDisposable := svc.ValidateDomainConcurrently(ctx, tt.domain)

			// Assert
			assert.Equal(t, tt.expectedExists, exists)
			assert.Equal(t, tt.expectedHasMX, hasMX)
			assert.Equal(t, tt.expectedDispose, isDisposable)

			// Verify mock expectations
			mockValidator.AssertExpectations(t)
		})
	}
}
