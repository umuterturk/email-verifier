package config

import (
	"os"
	"time"

	"emailvalidator/pkg/validator"
)

// LoadSMTPVerifierConfig loads SMTP verifier configuration from environment variables
func LoadSMTPVerifierConfig() validator.SMTPVerifierConfig {
	config := validator.DefaultSMTPVerifierConfig()

	// SMTP_VERIFY_ENABLED - defaults to true (enabled by default)
	if os.Getenv("SMTP_VERIFY_ENABLED") == "false" {
		config.Enabled = false
	}

	// SMTP_VERIFY_TIMEOUT - defaults to 10s
	if timeout := os.Getenv("SMTP_VERIFY_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.Timeout = d
		}
	}

	// SMTP_VERIFY_HELO_HOST - defaults to "localhost"
	if host := os.Getenv("SMTP_VERIFY_HELO_HOST"); host != "" {
		config.HeloHostname = host
	}

	// SMTP_VERIFY_MAIL_FROM - defaults to "verify@localhost"
	if mailFrom := os.Getenv("SMTP_VERIFY_MAIL_FROM"); mailFrom != "" {
		config.MailFrom = mailFrom
	}

	return config
}
