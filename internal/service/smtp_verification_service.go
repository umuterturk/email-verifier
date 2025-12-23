package service

import (
	"context"
	"net"
	"time"

	"emailvalidator/pkg/validator"
)

// DefaultSMTPVerificationService implements SMTPVerificationService
type DefaultSMTPVerificationService struct {
	verifier     validator.SMTPVerifier
	cacheManager *validator.SMTPCacheManager
	config       validator.SMTPVerifierConfig
}

// NewSMTPVerificationService creates a new SMTP verification service
func NewSMTPVerificationService(config validator.SMTPVerifierConfig) *DefaultSMTPVerificationService {
	return &DefaultSMTPVerificationService{
		verifier:     validator.NewDefaultSMTPVerifier(config),
		cacheManager: validator.NewSMTPCacheManager(24 * time.Hour),
		config:       config,
	}
}

// NewSMTPVerificationServiceWithVerifier creates a service with a custom verifier (for testing)
func NewSMTPVerificationServiceWithVerifier(config validator.SMTPVerifierConfig, verifier validator.SMTPVerifier) *DefaultSMTPVerificationService {
	return &DefaultSMTPVerificationService{
		verifier:     verifier,
		cacheManager: validator.NewSMTPCacheManager(24 * time.Hour),
		config:       config,
	}
}

// IsEnabled returns whether SMTP verification is enabled
func (s *DefaultSMTPVerificationService) IsEnabled() bool {
	return s.config.Enabled
}

// VerifyMailbox performs SMTP RCPT TO verification with catch-all detection
func (s *DefaultSMTPVerificationService) VerifyMailbox(ctx context.Context, email, domain string, mxRecords []*net.MX) validator.SMTPVerificationResult {
	if !s.config.Enabled {
		return validator.SMTPVerificationResult{Skipped: true, SkipReason: "disabled"}
	}

	if len(mxRecords) == 0 {
		return validator.SMTPVerificationResult{Skipped: true, SkipReason: "no_mx_records"}
	}

	// Check if domain is known to be catch-all (cached)
	if isCatchAll, found := s.cacheManager.GetCatchAll(domain); found {
		if isCatchAll {
			return validator.SMTPVerificationResult{
				Verified:   true, // Can't verify individual mailboxes on catch-all
				IsCatchAll: true,
			}
		}
		// Domain is not catch-all, proceed with verification
	} else {
		// Not in cache - detect catch-all first
		isCatchAll, err := s.verifier.DetectCatchAll(ctx, domain, mxRecords)
		if err == nil {
			s.cacheManager.SetCatchAll(domain, isCatchAll)
		}

		if isCatchAll {
			return validator.SMTPVerificationResult{
				Verified:   true,
				IsCatchAll: true,
			}
		}
	}

	// Domain is not catch-all, verify the specific mailbox
	result := s.verifier.VerifyMailbox(ctx, email, domain, mxRecords)

	return result
}

// IsCatchAll checks if domain is a catch-all (accepts all addresses)
func (s *DefaultSMTPVerificationService) IsCatchAll(ctx context.Context, domain string, mxRecords []*net.MX) (bool, error) {
	// Check cache first
	if isCatchAll, found := s.cacheManager.GetCatchAll(domain); found {
		return isCatchAll, nil
	}

	// Detect catch-all
	isCatchAll, err := s.verifier.DetectCatchAll(ctx, domain, mxRecords)
	if err == nil {
		s.cacheManager.SetCatchAll(domain, isCatchAll)
	}

	return isCatchAll, err
}

// ClearCache clears the catch-all cache (useful for testing)
func (s *DefaultSMTPVerificationService) ClearCache() {
	s.cacheManager.Clear()
}
