// Package mocks contains mock implementations for service tests
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockEmailRuleValidator mocks the EmailRuleValidator interface
type MockEmailRuleValidator struct {
	mock.Mock
}

func (m *MockEmailRuleValidator) ValidateSyntax(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

func (m *MockEmailRuleValidator) IsRoleBased(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

func (m *MockEmailRuleValidator) CalculateScore(validations map[string]bool) int {
	args := m.Called(validations)
	return args.Int(0)
}

func (m *MockEmailRuleValidator) GetTypoSuggestions(email string) []string {
	args := m.Called(email)
	return args.Get(0).([]string)
}

// MockDomainValidator mocks the DomainValidator interface
type MockDomainValidator struct {
	mock.Mock
}

func (m *MockDomainValidator) ValidateDomain(domain string) bool {
	args := m.Called(domain)
	return args.Bool(0)
}

func (m *MockDomainValidator) ValidateMXRecords(domain string) bool {
	args := m.Called(domain)
	return args.Bool(0)
}

func (m *MockDomainValidator) IsDisposable(domain string) bool {
	args := m.Called(domain)
	return args.Bool(0)
}

// MockDomainValidationService mocks the DomainValidationService interface
type MockDomainValidationService struct {
	mock.Mock
}

func (m *MockDomainValidationService) ValidateDomainConcurrently(ctx context.Context, domain string) (exists, hasMX, isDisposable bool) {
	args := m.Called(ctx, domain)
	return args.Bool(0), args.Bool(1), args.Bool(2)
}

// MockMetricsCollector mocks the MetricsCollector interface
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordValidationScore(name string, score float64) {
	m.Called(name, score)
}

func (m *MockMetricsCollector) UpdateMemoryUsage(heapInUse, stackInUse float64) {
	m.Called(heapInUse, stackInUse)
}
