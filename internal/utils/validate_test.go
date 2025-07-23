package utils

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"valid http URL", "http://example.com", false},
		{"valid https URL", "https://example.com", false},
		{"empty URL", "", true},
		{"invalid format", "://no-scheme.com", true},
		{"unsupported scheme", "ftp://example.com", true},
		{"missing host", "http://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input)
			if (err != nil) != tt.error {
				t.Errorf("ValidateURL(%q) error = %v, error = %v", tt.input, err, tt.error)
			}
		})
	}
}
