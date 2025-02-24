package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

// SyntaxValidator handles email syntax validation
type SyntaxValidator struct {
	// Additional validation for cases not covered by net/mail
	additionalChecks *regexp.Regexp
	// Regex to detect quoted strings
	quotedStringCheck *regexp.Regexp
}

// NewSyntaxValidator creates a new instance of SyntaxValidator
func NewSyntaxValidator() *SyntaxValidator {
	return &SyntaxValidator{
		// Additional regex for cases not handled by net/mail
		additionalChecks: regexp.MustCompile(`^[^.][^@]*[^.]@[^.][^@]*[^.]$`),
		// Regex to detect quoted strings in local part
		quotedStringCheck: regexp.MustCompile(`"[^"]*"`),
	}
}

// Validate checks if the email address format is valid
func (v *SyntaxValidator) Validate(email string) bool {
	if email == "" {
		return false
	}

	// Check maximum length (RFC 5321)
	if len(email) > 254 {
		return false
	}

	// Check for quoted strings
	if v.quotedStringCheck.MatchString(email) {
		return false
	}

	// Parse with net/mail
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// Additional validations not covered by net/mail
	if !v.additionalChecks.MatchString(addr.Address) {
		return false
	}

	// Check for consecutive dots
	if strings.Contains(addr.Address, "..") {
		return false
	}

	// Split email into local part and domain
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return false
	}

	localPart, domain := parts[0], parts[1]

	// Check local part and domain lengths
	if len(localPart) > 64 || len(domain) > 255 {
		return false
	}

	return true
}
