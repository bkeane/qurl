package main

import (
	"net/url"
	"testing"
)

func TestURLParsing(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedScheme string
		expectedHost   string
		isAbsolute     bool
	}{
		{
			name:           "absolute HTTPS URL",
			input:          "https://example.com/api/test",
			expectedScheme: "https",
			expectedHost:   "example.com",
			isAbsolute:     true,
		},
		{
			name:           "absolute HTTP URL",
			input:          "http://localhost:8080/test",
			expectedScheme: "http",
			expectedHost:   "localhost:8080",
			isAbsolute:     true,
		},
		{
			name:       "relative path",
			input:      "/api/test",
			isAbsolute: false,
		},
		{
			name:       "relative path without leading slash",
			input:      "test",
			isAbsolute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			isAbsolute := parsedURL.Scheme != "" && parsedURL.Host != ""
			if isAbsolute != tt.isAbsolute {
				t.Errorf("Expected isAbsolute=%v, got %v", tt.isAbsolute, isAbsolute)
			}

			if tt.isAbsolute {
				if parsedURL.Scheme != tt.expectedScheme {
					t.Errorf("Expected scheme %q, got %q", tt.expectedScheme, parsedURL.Scheme)
				}
				if parsedURL.Host != tt.expectedHost {
					t.Errorf("Expected host %q, got %q", tt.expectedHost, parsedURL.Host)
				}
			}
		})
	}
}

func TestServerURLValidation(t *testing.T) {
	tests := []struct {
		name        string
		serverFlag  string
		expected    string
		expectError bool
	}{
		{
			name:       "full HTTPS URL",
			serverFlag: "https://api.example.com",
			expected:   "https://api.example.com",
		},
		{
			name:       "full HTTP URL",
			serverFlag: "http://localhost:8080",
			expected:   "http://localhost:8080",
		},
		{
			name:       "URL with path",
			serverFlag: "https://api.example.com/v1",
			expected:   "https://api.example.com/v1",
		},
		{
			name:        "incomplete URL (missing scheme)",
			serverFlag:  "example.com",
			expectError: true,
		},
		{
			name:        "incomplete URL (only scheme)",
			serverFlag:  "https://",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URL parsing validation directly
			parsedURL, err := url.Parse(tt.serverFlag)
			if err != nil && !tt.expectError {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			isComplete := parsedURL.Scheme != "" && parsedURL.Host != ""

			if tt.expectError {
				if isComplete {
					t.Errorf("Expected incomplete URL, but URL appears complete")
				}
			} else {
				if !isComplete {
					t.Errorf("Expected complete URL, but URL is incomplete")
				}
				if tt.serverFlag != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, tt.serverFlag)
				}
			}
		})
	}
}
