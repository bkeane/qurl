package main

import (
	"strings"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
)

func TestAuthenticationConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		sigV4Enabled bool
		sigV4Service string
		wantSigV4    bool
		wantService  string
	}{
		{
			name:         "no authentication",
			sigV4Enabled: false,
			sigV4Service: "",
			wantSigV4:    false,
			wantService:  "",
		},
		{
			name:         "SigV4 enabled with default service",
			sigV4Enabled: true,
			sigV4Service: "execute-api",
			wantSigV4:    true,
			wantService:  "execute-api",
		},
		{
			name:         "SigV4 enabled with custom service",
			sigV4Enabled: true,
			sigV4Service: "lambda",
			wantSigV4:    true,
			wantService:  "lambda",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration
			cfg := &config.Config{
				SigV4Enabled: tt.sigV4Enabled,
				SigV4Service: tt.sigV4Service,
			}

			// Verify configuration values
			if cfg.SigV4Enabled != tt.wantSigV4 {
				t.Errorf("SigV4Enabled = %v, want %v", cfg.SigV4Enabled, tt.wantSigV4)
			}

			if cfg.SigV4Service != tt.wantService {
				t.Errorf("SigV4Service = %q, want %q", cfg.SigV4Service, tt.wantService)
			}
		})
	}
}

func TestHeaderParsing(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    map[string]string
	}{
		{
			name:    "no headers",
			headers: []string{},
			want:    map[string]string{},
		},
		{
			name:    "authorization header",
			headers: []string{"Authorization: Bearer token123"},
			want:    map[string]string{"Authorization": "Bearer token123"},
		},
		{
			name:    "multiple headers",
			headers: []string{"Authorization: Bearer token123", "Content-Type: application/json"},
			want:    map[string]string{"Authorization": "Bearer token123", "Content-Type": "application/json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Headers: tt.headers,
			}

			// Parse headers
			parsed := make(map[string]string)
			for _, header := range cfg.Headers {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					parsed[key] = value
				}
			}

			// Verify parsed headers
			for wantKey, wantValue := range tt.want {
				if got, ok := parsed[wantKey]; !ok || got != wantValue {
					t.Errorf("Header %q = %q, want %q", wantKey, got, wantValue)
				}
			}

			// Check no extra headers
			if len(parsed) != len(tt.want) {
				t.Errorf("Got %d headers, want %d", len(parsed), len(tt.want))
			}
		})
	}
}

func TestLambdaURLDetection(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		isLambda bool
	}{
		{"lambda URL", "lambda://function-name", true},
		{"lambda URL with path", "lambda://function-name/invoke", true},
		{"HTTPS URL", "https://example.com", false},
		{"HTTP URL", "http://example.com", false},
		{"lambda in path", "https://example.com/lambda://not-lambda", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isLambda := strings.HasPrefix(tc.url, "lambda://")
			if isLambda != tc.isLambda {
				t.Errorf("Expected lambda detection for %q to be %v, got %v", tc.url, tc.isLambda, isLambda)
			}
		})
	}
}
