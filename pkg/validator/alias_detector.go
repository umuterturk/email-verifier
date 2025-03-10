package validator

import (
	"regexp"
	"strings"
)

// AliasDetector provides methods for detecting email aliases
type AliasDetector struct {
	providers map[string]AliasProvider
}

// AliasProvider defines behavior for detecting aliases for a specific provider
type AliasProvider interface {
	// IsAlias checks if the email is an alias
	IsAlias(localPart string) bool
	// GetCanonicalEmail returns the canonical email from an alias
	GetCanonicalEmail(localPart, domain string) string
}

// NewAliasDetector creates a new instance of AliasDetector
func NewAliasDetector() *AliasDetector {
	detector := &AliasDetector{
		providers: make(map[string]AliasProvider),
	}

	// Register providers
	detector.providers["gmail.com"] = NewGmailAliasProvider()
	detector.providers["googlemail.com"] = NewGmailAliasProvider()
	detector.providers["yahoo.com"] = NewYahooAliasProvider()
	detector.providers["outlook.com"] = NewOutlookAliasProvider()
	detector.providers["hotmail.com"] = NewOutlookAliasProvider()
	detector.providers["live.com"] = NewOutlookAliasProvider()

	return detector
}

// DetectAlias checks if the email is an alias and returns the canonical email if it is
func (d *AliasDetector) DetectAlias(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localPart, domain := parts[0], parts[1]

	// Get provider handler
	provider, exists := d.providers[strings.ToLower(domain)]
	if !exists {
		return ""
	}

	// Check if it's an alias
	if provider.IsAlias(localPart) {
		return provider.GetCanonicalEmail(localPart, domain)
	}

	return ""
}

// --------------------------------------------------------
// Gmail Alias Provider Implementation
// --------------------------------------------------------

// GmailAliasProvider handles Gmail-specific alias detection
type GmailAliasProvider struct{}

// NewGmailAliasProvider creates a new Gmail alias provider
func NewGmailAliasProvider() *GmailAliasProvider {
	return &GmailAliasProvider{}
}

// IsAlias checks if the Gmail email is an alias
func (p *GmailAliasProvider) IsAlias(localPart string) bool {
	// Gmail aliases can have dots or a plus sign followed by anything
	return strings.Contains(localPart, ".") || strings.Contains(localPart, "+")
}

// GetCanonicalEmail returns the canonical Gmail email
func (p *GmailAliasProvider) GetCanonicalEmail(localPart, domain string) string {
	// For Gmail, remove all dots and anything after plus
	canonical := localPart

	// Remove anything after plus
	if idx := strings.Index(canonical, "+"); idx != -1 {
		canonical = canonical[:idx]
	}

	// Remove all dots
	canonical = strings.ReplaceAll(canonical, ".", "")

	// Always use gmail.com as canonical domain
	return canonical + "@gmail.com"
}

// --------------------------------------------------------
// Yahoo Alias Provider Implementation
// --------------------------------------------------------

// YahooAliasProvider handles Yahoo-specific alias detection
type YahooAliasProvider struct {
	// Yahoo uses the format localPart-aliasText@yahoo.com
	aliasRegex *regexp.Regexp
}

// NewYahooAliasProvider creates a new Yahoo alias provider
func NewYahooAliasProvider() *YahooAliasProvider {
	return &YahooAliasProvider{
		aliasRegex: regexp.MustCompile(`^([^-]+)-([^@]+)$`),
	}
}

// IsAlias checks if the Yahoo email is an alias
func (p *YahooAliasProvider) IsAlias(localPart string) bool {
	// Yahoo aliases have a hyphen
	return p.aliasRegex.MatchString(localPart)
}

// GetCanonicalEmail returns the canonical Yahoo email
func (p *YahooAliasProvider) GetCanonicalEmail(localPart, domain string) string {
	matches := p.aliasRegex.FindStringSubmatch(localPart)
	if len(matches) > 1 {
		// Extract the base email address
		return matches[1] + "@" + domain
	}
	return localPart + "@" + domain
}

// --------------------------------------------------------
// Outlook/Hotmail Alias Provider Implementation
// --------------------------------------------------------

// OutlookAliasProvider handles Outlook/Hotmail-specific alias detection
type OutlookAliasProvider struct {
	// Outlook uses the format localPart+aliasText@outlook.com
}

// NewOutlookAliasProvider creates a new Outlook alias provider
func NewOutlookAliasProvider() *OutlookAliasProvider {
	return &OutlookAliasProvider{}
}

// IsAlias checks if the Outlook email is an alias
func (p *OutlookAliasProvider) IsAlias(localPart string) bool {
	// Outlook aliases have a plus sign
	return strings.Contains(localPart, "+")
}

// GetCanonicalEmail returns the canonical Outlook email
func (p *OutlookAliasProvider) GetCanonicalEmail(localPart, domain string) string {
	// For Outlook, remove anything after plus
	if idx := strings.Index(localPart, "+"); idx != -1 {
		return localPart[:idx] + "@" + domain
	}
	return localPart + "@" + domain
}
