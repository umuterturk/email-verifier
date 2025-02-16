// Package validator provides email validation functionality including syntax checking,
// DNS validation, disposable email detection, and typo suggestion generation.
package validator

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadDisposableDomainsFromFile loads disposable email domains from a file
func LoadDisposableDomainsFromFile(path string) ([]string, error) {
	var domains []string

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Error closing disposable domains file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			domains = append(domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

// NewDisposableValidatorFromFile creates a new instance of DisposableValidator using domains from a file
func NewDisposableValidatorFromFile(path string) *DisposableValidator {
	domains, err := LoadDisposableDomainsFromFile(path)
	if err != nil {
		log.Printf("Warning: Could not load disposable domains from file: %v", err)
		return NewDisposableValidator() // Fall back to default domains
	}
	return NewDisposableValidatorWithDomains(domains)
}
