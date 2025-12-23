package validator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SMTPDialer interface for creating SMTP connections (mockable for testing)
type SMTPDialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

// DefaultSMTPDialer uses net.Dialer for connections
type DefaultSMTPDialer struct {
	Timeout time.Duration
}

// DialContext creates a new TCP connection with context and timeout
func (d *DefaultSMTPDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: d.Timeout}
	return dialer.DialContext(ctx, network, addr)
}

// DefaultSMTPVerifier implements SMTPVerifier interface
type DefaultSMTPVerifier struct {
	config SMTPVerifierConfig
	dialer SMTPDialer
}

// NewDefaultSMTPVerifier creates a new SMTP verifier with default dialer
func NewDefaultSMTPVerifier(config SMTPVerifierConfig) *DefaultSMTPVerifier {
	return &DefaultSMTPVerifier{
		config: config,
		dialer: &DefaultSMTPDialer{Timeout: config.Timeout},
	}
}

// NewSMTPVerifierWithDialer creates a verifier with custom dialer (for testing)
func NewSMTPVerifierWithDialer(config SMTPVerifierConfig, dialer SMTPDialer) *DefaultSMTPVerifier {
	return &DefaultSMTPVerifier{
		config: config,
		dialer: dialer,
	}
}

// VerifyMailbox performs SMTP RCPT TO verification
func (v *DefaultSMTPVerifier) VerifyMailbox(ctx context.Context, email, domain string, mxRecords []*net.MX) SMTPVerificationResult {
	result := SMTPVerificationResult{}

	// Skip if no MX records
	if len(mxRecords) == 0 {
		result.Skipped = true
		result.SkipReason = "no_mx_records"
		return result
	}

	// Sort MX by priority (lowest preference value first)
	sortedMX := sortMXByPriority(mxRecords)

	// Try each MX server in priority order
	var lastErr error
	for _, mx := range sortedMX {
		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		default:
		}

		host := strings.TrimSuffix(mx.Host, ".")
		addr := net.JoinHostPort(host, "25")

		code, msg, err := v.tryVerify(ctx, email, host, addr)
		if err != nil {
			lastErr = err
			continue // Try next MX server
		}

		result.ResponseCode = code
		result.ResponseMsg = msg

		// Handle response based on category
		switch CategorizeResponse(code) {
		case SMTPCategorySuccess:
			result.Verified = true
			return result
		case SMTPCategoryTemporary:
			// Greylisting or temporary error - try next MX
			lastErr = fmt.Errorf("temporary rejection: %d %s", code, msg)
			continue
		case SMTPCategoryPermanent:
			if IsMailboxNotFound(code) {
				result.Verified = false
				return result
			}
			// Some other permanent error, try next MX
			lastErr = fmt.Errorf("permanent error: %d %s", code, msg)
			continue
		}
	}

	// All MX servers failed
	result.Error = lastErr
	return result
}

// DetectCatchAll tests if domain accepts all addresses (catch-all configuration)
func (v *DefaultSMTPVerifier) DetectCatchAll(ctx context.Context, domain string, mxRecords []*net.MX) (bool, error) {
	// Generate a random address that definitely doesn't exist
	randomLocal := generateRandomLocalPart()
	randomEmail := randomLocal + "@" + domain

	result := v.VerifyMailbox(ctx, randomEmail, domain, mxRecords)

	if result.Error != nil {
		return false, result.Error
	}

	if result.Skipped {
		return false, nil
	}

	// If server accepts random address, it's a catch-all domain
	return result.Verified, nil
}

// tryVerify attempts SMTP verification against a single mail server
func (v *DefaultSMTPVerifier) tryVerify(ctx context.Context, email, host, addr string) (int, string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, v.config.Timeout)
	defer cancel()

	// Connect to mail server
	conn, err := v.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return 0, "", fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// Create SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return 0, "", fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	// Send EHLO (or fall back to HELO automatically)
	if err = client.Hello(v.config.HeloHostname); err != nil {
		return 0, "", fmt.Errorf("hello failed: %w", err)
	}

	// Send MAIL FROM
	if err = client.Mail(v.config.MailFrom); err != nil {
		return 0, "", fmt.Errorf("mail from failed: %w", err)
	}

	// Send RCPT TO - this is where we verify the mailbox
	err = client.Rcpt(email)

	// Parse SMTP response from error (or success)
	code, msg := parseSmtpError(err)

	// Clean up - send RSET and QUIT
	client.Reset()
	client.Quit()

	return code, msg, nil
}

// sortMXByPriority sorts MX records by preference (lowest first)
func sortMXByPriority(mxRecords []*net.MX) []*net.MX {
	sorted := make([]*net.MX, len(mxRecords))
	copy(sorted, mxRecords)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Pref < sorted[j].Pref
	})
	return sorted
}

// generateRandomLocalPart creates a random non-existent address for catch-all detection
func generateRandomLocalPart() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "verify-catchall-" + hex.EncodeToString(bytes)
}

// parseSmtpError extracts SMTP code and message from error
func parseSmtpError(err error) (int, string) {
	if err == nil {
		return SMTPCodeOK, "OK"
	}

	msg := err.Error()

	// Try to parse SMTP error format: "550 5.1.1 User unknown"
	// First, try to extract the 3-digit code at the start
	if len(msg) >= 3 {
		if code, parseErr := strconv.Atoi(msg[:3]); parseErr == nil && code >= 200 && code < 600 {
			return code, msg
		}
	}

	// If we can't parse a code, return 0 to indicate unknown
	return 0, msg
}
