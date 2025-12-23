package validator

import (
	"context"
	"net"
	"time"
)

// SMTPVerificationResult contains the results of SMTP verification
type SMTPVerificationResult struct {
	Verified     bool   // Whether SMTP verification confirmed mailbox exists
	IsCatchAll   bool   // Whether domain is a catch-all (accepts all addresses)
	ResponseCode int    // SMTP response code (250, 550, etc.)
	ResponseMsg  string // SMTP response message
	Error        error  // Any error that occurred during verification
	Skipped      bool   // True if verification was skipped (e.g., no MX records)
	SkipReason   string // Reason for skipping verification
}

// SMTPVerifier interface for SMTP mailbox verification
type SMTPVerifier interface {
	// VerifyMailbox performs SMTP RCPT TO verification to check if mailbox exists
	VerifyMailbox(ctx context.Context, email, domain string, mxRecords []*net.MX) SMTPVerificationResult

	// DetectCatchAll tests if domain accepts all email addresses (catch-all configuration)
	DetectCatchAll(ctx context.Context, domain string, mxRecords []*net.MX) (bool, error)
}

// SMTPVerifierConfig holds configuration for SMTP verification
type SMTPVerifierConfig struct {
	Enabled      bool          // Enable/disable SMTP verification
	Timeout      time.Duration // Connection timeout per MX server
	HeloHostname string        // Hostname for HELO/EHLO command
	MailFrom     string        // MAIL FROM address
}

// DefaultSMTPVerifierConfig returns sensible defaults for SMTP verification
func DefaultSMTPVerifierConfig() SMTPVerifierConfig {
	return SMTPVerifierConfig{
		Enabled:      true, // Enabled by default per user preference
		Timeout:      10 * time.Second,
		HeloHostname: "localhost",
		MailFrom:     "verify@localhost",
	}
}
