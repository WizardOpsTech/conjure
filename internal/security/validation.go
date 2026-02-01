package security

import (
	"crypto/subtle"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// Ensures the path does not contain directory traversal sequences
func ValidatePathSafety(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)

	if filepath.IsAbs(cleanPath) && !filepath.IsAbs(path) {
		return fmt.Errorf("path traversal detected: path becomes absolute after cleaning")
	}

	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: path contains parent directory references")
	}

	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	suspiciousPatterns := []string{
		"/../",
		"/./",
		"\\..\\",
		"\\.\\",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(path, pattern) {
			return fmt.Errorf("path contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateURL validates a URL format
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("unsupported URL scheme: %s (only http and https are allowed)", parsedURL.Scheme)
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL must include a hostname")
	}

	return nil
}

func ValidateVariableName(name string) error {
	if name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("variable name '%s' contains invalid characters (only alphanumeric, underscore, and hyphen allowed)", name)
	}

	return nil
}

func CompareHashes(hash1, hash2 string) bool {
	if len(hash1) != len(hash2) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) == 1
}

func ValidateFileSize(size int64, maxSize int64) error {
	if size < 0 {
		return fmt.Errorf("file size cannot be negative")
	}

	if size > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes", size, maxSize)
	}

	return nil
}
