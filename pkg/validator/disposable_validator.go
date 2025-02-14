package validator

// DisposableValidator handles disposable email validation
type DisposableValidator struct {
	disposableDomains map[string]struct{}
}

// NewDisposableValidator creates a new instance of DisposableValidator
func NewDisposableValidator() *DisposableValidator {
	return &DisposableValidator{
		disposableDomains: initDisposableDomains(),
	}
}

// Validate checks if the email domain is from a disposable email provider
func (v *DisposableValidator) Validate(domain string) bool {
	_, exists := v.disposableDomains[domain]
	return exists
}
