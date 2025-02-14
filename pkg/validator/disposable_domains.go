package validator

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// initDisposableDomains initializes the map of disposable email domains by reading from a file
func initDisposableDomains() map[string]struct{} {
	disposableDomains := make(map[string]struct{})

	// Try multiple possible locations for the config file
	possiblePaths := []string{
		"../../config/disposable_domains.txt",
	}

	var file *os.File
	var err error

	// Try to open the file from possible locations
	for _, path := range possiblePaths {
		file, err = os.Open(filepath.Clean(path))
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Printf("Warning: Could not open disposable domains file from any location: %v", err)
		return disposableDomains
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			disposableDomains[domain] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Warning: Error reading disposable domains file: %v", err)
	}

	return disposableDomains
}
