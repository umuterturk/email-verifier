package validator

import (
	"bufio"
	"os"
	"strings"
)

// DomainReader defines the interface for reading disposable domains
type DomainReader interface {
	ReadDomains() ([]string, error)
}

// FileDomainReader implements DomainReader interface for reading from a file
type FileDomainReader struct {
	filePath string
}

// NewFileDomainReader creates a new FileDomainReader instance
func NewFileDomainReader(filePath string) *FileDomainReader {
	return &FileDomainReader{
		filePath: filePath,
	}
}

// ReadDomains reads domains from a file, skipping empty lines and comments
func (r *FileDomainReader) ReadDomains() ([]string, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// If we already have an error, keep it as it's likely more important
			if err == nil {
				err = closeErr
			}
		}
	}()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

// StaticDomainReader implements DomainReader interface for static domain list
type StaticDomainReader struct {
	domains []string
}

// NewStaticDomainReader creates a new StaticDomainReader instance
func NewStaticDomainReader(domains []string) *StaticDomainReader {
	return &StaticDomainReader{
		domains: domains,
	}
}

// ReadDomains returns the static list of domains
func (r *StaticDomainReader) ReadDomains() ([]string, error) {
	return r.domains, nil
}
