package validator

import (
	"strings"
	"time"
)

// EmailValidator provides methods for validating email addresses
type EmailValidator struct {
	syntaxValidator     *SyntaxValidator
	domainValidator     *DomainValidator
	roleValidator       *RoleValidator
	disposableValidator *DisposableValidator
}

// NewEmailValidator creates a new instance of EmailValidator
func NewEmailValidator() (*EmailValidator, error) {
	cacheManager := NewDomainCacheManager(time.Hour)
	resolver := &DefaultResolver{timeout: 2 * time.Second}

	disposableValidator, err := NewDisposableValidator()
	if err != nil {
		return nil, err
	}

	return &EmailValidator{
		syntaxValidator:     NewSyntaxValidator(),
		domainValidator:     NewDomainValidator(resolver, cacheManager),
		roleValidator:       NewRoleValidator(),
		disposableValidator: disposableValidator,
	}, nil
}

// NewEmailValidatorWithResolver creates a new instance of EmailValidator with a custom resolver
func NewEmailValidatorWithResolver(resolver DNSResolver) (*EmailValidator, error) {
	cacheManager := NewDomainCacheManager(time.Hour)

	disposableValidator, err := NewDisposableValidator()
	if err != nil {
		return nil, err
	}

	return &EmailValidator{
		syntaxValidator:     NewSyntaxValidator(),
		domainValidator:     NewDomainValidator(resolver, cacheManager),
		roleValidator:       NewRoleValidator(),
		disposableValidator: disposableValidator,
	}, nil
}

// SetResolver allows changing the DNS resolver
func (v *EmailValidator) SetResolver(resolver DNSResolver) {
	v.domainValidator = NewDomainValidator(resolver, v.domainValidator.cacheManager)
}

// SetCacheDuration sets how long domain lookup results are cached
func (v *EmailValidator) SetCacheDuration(duration time.Duration) {
	v.domainValidator.cacheManager.SetDuration(duration)
}

// ValidateSyntax checks if the email address format is valid
func (v *EmailValidator) ValidateSyntax(email string) bool {
	// Check maximum length (RFC 5321)
	if len(email) > 254 {
		return false
	}

	// Split email into local part and domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	// Check local part and domain lengths
	localPart, domain := parts[0], parts[1]
	if len(localPart) > 64 || len(domain) > 255 {
		return false
	}

	return v.syntaxValidator.Validate(email)
}

// ValidateDomain checks if the domain exists
func (v *EmailValidator) ValidateDomain(domain string) bool {
	return v.domainValidator.Validate(domain)
}

// ValidateMXRecords checks if the domain has valid MX records
func (v *EmailValidator) ValidateMXRecords(domain string) bool {
	return v.domainValidator.ValidateMX(domain)
}

// IsDisposable checks if the email domain is from a disposable email provider
func (v *EmailValidator) IsDisposable(domain string) bool {
	return v.disposableValidator.Validate(domain)
}

// IsRoleBased checks if the email address is role-based
func (v *EmailValidator) IsRoleBased(email string) bool {
	return v.roleValidator.Validate(email)
}

// CalculateScore calculates a score based on validation results
func (v *EmailValidator) CalculateScore(validations map[string]bool) int {
	score := 0
	weights := map[string]int{
		"syntax":         20,
		"domain_exists":  20,
		"mx_records":     20,
		"mailbox_exists": 20,
		"is_disposable":  10,
		"is_role_based":  10,
	}

	for check, weight := range weights {
		passed, exists := validations[check]
		if !exists {
			continue
		}

		switch check {
		case "is_disposable", "is_role_based":
			// For negative checks, add points when false
			if !passed {
				score += weight
			}
		default:
			// For positive checks, add points when true
			if passed {
				score += weight
			}
		}
	}

	return score
}

// GetTypoSuggestions returns possible corrections for common email typos
func (v *EmailValidator) GetTypoSuggestions(email string) []string {
	// Common domain corrections
	commonDomains := map[string]string{
		"gmial.com":  "gmail.com",
		"gmal.com":   "gmail.com",
		"gamil.com":  "gmail.com",
		"yaho.com":   "yahoo.com",
		"yahooo.com": "yahoo.com",
		"hotmai.com": "hotmail.com",
		"hotmal.com": "hotmail.com",
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil
	}

	localPart, domain := parts[0], parts[1]
	var suggestions []string

	// Check for common domain typos
	if correctedDomain, exists := commonDomains[domain]; exists {
		suggestions = append(suggestions, localPart+"@"+correctedDomain)
	}

	return suggestions
}
