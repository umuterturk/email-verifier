package validator

// DisposableValidator handles disposable email validation
type DisposableValidator struct {
	disposableDomains map[string]struct{}
}

// defaultDisposableDomains returns a default list of disposable email domains
func defaultDisposableDomains() []string {
	return []string{
		"10minutemail.com",
		"mytempmail.com",
		"tempmail.com",
		"throwawaymail.com",
		"guerrillamail.com",
		"mailinator.com",
		"yopmail.com",
		"getairmail.com",
		"temp-mail.org",
		"fakeinbox.com",
		"trashmail.com",
		"sharklasers.com",
		"guerrillamail.info",
		"grr.la",
		"maildrop.cc",
		"dispostable.com",
		"tempmailaddress.com",
		"tempmail.net",
		"emailondeck.com",
		"spamgourmet.com",
		"tempmailer.com",
		"tempmail.de",
		"disposablemail.com",
		"mailnesia.com",
		"tempmailbox.com",
		"tempinbox.com",
		"throwawayemail.com",
		"mailcatch.com",
		"mintemail.com",
		"mytrashmail.com",
	}
}

// NewDisposableValidator creates a new instance of DisposableValidator
func NewDisposableValidator() *DisposableValidator {
	return NewDisposableValidatorWithDomains(defaultDisposableDomains())
}

// NewDisposableValidatorWithDomains creates a new instance of DisposableValidator with a custom list of domains
func NewDisposableValidatorWithDomains(domains []string) *DisposableValidator {
	disposableDomains := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		disposableDomains[domain] = struct{}{}
	}
	return &DisposableValidator{
		disposableDomains: disposableDomains,
	}
}

// Validate checks if the email domain is from a disposable email provider
func (v *DisposableValidator) Validate(domain string) bool {
	_, exists := v.disposableDomains[domain]
	return exists
}
