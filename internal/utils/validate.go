package utils

import (
	"errors"
	"net/url"
)

// ValidateURL checks if the input string is a valid HTTP/HTTPS URL
func ValidateURL(raw string) error {
	if raw == "" {
		return errors.New("URL is empty")
	}

	// Use net/url to parse and further validate
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return errors.New("Invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("Only HTTP and HTTPS URLs are supported")
	}

	if parsed.Host == "" {
		return errors.New("URL must contain a host")
	}

	return nil
}
