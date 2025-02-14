package validator

import "regexp"

// SyntaxValidator handles email syntax validation
type SyntaxValidator struct {
	emailRegex *regexp.Regexp
}

// NewSyntaxValidator creates a new instance of SyntaxValidator
func NewSyntaxValidator() *SyntaxValidator {
	return &SyntaxValidator{
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}
}

// Validate checks if the email address format is valid
func (v *SyntaxValidator) Validate(email string) bool {
	return v.emailRegex.MatchString(email)
}
