package service

import "emailvalidator/pkg/monitoring"

// MetricsAdapter adapts the monitoring package to implement MetricsCollector interface
type MetricsAdapter struct{}

// NewMetricsAdapter creates a new instance of MetricsAdapter
func NewMetricsAdapter() *MetricsAdapter {
	return &MetricsAdapter{}
}

// RecordValidationScore records a validation score metric
func (m *MetricsAdapter) RecordValidationScore(name string, score float64) {
	monitoring.RecordValidationScore(name, score)
}

// UpdateMemoryUsage updates memory usage metrics
func (m *MetricsAdapter) UpdateMemoryUsage(heapInUse, stackInUse float64) {
	monitoring.UpdateMemoryUsage(heapInUse, stackInUse)
}
