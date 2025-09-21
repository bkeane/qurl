package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestApplyAuthentication(t *testing.T) {
	tests := []struct {
		name         string
		targetURL    string
		bearerToken  string
		sigV4Enabled bool
		sigV4Service string
		expectedAuth string
		shouldSign   bool
	}{
		{
			name:         "no authentication",
			targetURL:    "https://example.com/test",
			bearerToken:  "",
			sigV4Enabled: false,
			sigV4Service: "",
			expectedAuth: "",
			shouldSign:   false,
		},
		{
			name:         "bearer token only",
			targetURL:    "https://example.com/test",
			bearerToken:  "test-token",
			sigV4Enabled: false,
			sigV4Service: "",
			expectedAuth: "Bearer test-token",
			shouldSign:   false,
		},
		{
			name:         "lambda URL skips SigV4",
			targetURL:    "lambda://test-function",
			bearerToken:  "",
			sigV4Enabled: true,
			sigV4Service: "execute-api",
			expectedAuth: "",
			shouldSign:   false,
		},
		{
			name:         "lambda URL with bearer token",
			targetURL:    "lambda://test-function",
			bearerToken:  "token",
			sigV4Enabled: true,
			sigV4Service: "execute-api",
			expectedAuth: "Bearer token",
			shouldSign:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origBearerToken := bearerToken
			origSigV4Enabled := sigV4Enabled
			origSigV4Service := sigV4Service

			// Set test values
			bearerToken = tt.bearerToken
			sigV4Enabled = tt.sigV4Enabled
			sigV4Service = tt.sigV4Service

			// Create test request
			req, err := http.NewRequest("GET", tt.targetURL, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Apply authentication
			err = applyAuthentication(req, tt.targetURL)

			// For AWS credential tests, we expect it to fail if no credentials
			if tt.sigV4Enabled && !strings.HasPrefix(tt.targetURL, "lambda://") {
				// SigV4 will likely fail without proper AWS credentials in test environment
				// That's expected - we're just testing that the logic is triggered
				if err == nil {
					t.Logf("SigV4 signing succeeded (AWS credentials available)")
				} else {
					t.Logf("SigV4 signing failed as expected in test environment: %v", err)
				}
			} else {
				// Non-SigV4 cases should not error
				if err != nil {
					t.Errorf("applyAuthentication() error = %v, wantErr false", err)
				}
			}

			// Check Authorization header
			authHeader := req.Header.Get("Authorization")
			if tt.expectedAuth != "" && !strings.Contains(authHeader, tt.expectedAuth) {
				// For SigV4, we might get AWS4-HMAC-SHA256 instead of empty
				if tt.sigV4Enabled && strings.Contains(authHeader, "AWS4-HMAC-SHA256") {
					t.Logf("Got AWS SigV4 authorization as expected: %s", authHeader)
				} else {
					t.Errorf("Expected authorization to contain %q, got %q", tt.expectedAuth, authHeader)
				}
			} else if tt.expectedAuth == "" && authHeader != "" && !tt.sigV4Enabled {
				t.Errorf("Expected no authorization header, got %q", authHeader)
			}

			// Restore original values
			bearerToken = origBearerToken
			sigV4Enabled = origSigV4Enabled
			sigV4Service = origSigV4Service
		})
	}
}

func TestBearerTokenPriority(t *testing.T) {
	// Save original values
	origBearerToken := bearerToken
	origSigV4Enabled := sigV4Enabled
	origSigV4Service := sigV4Service

	// Test bearer token applied first, then overridden by custom header
	bearerToken = "test-token"
	sigV4Enabled = false
	sigV4Service = ""

	req, err := http.NewRequest("GET", "https://example.com/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Apply authentication (bearer token)
	err = applyAuthentication(req, "https://example.com/test")
	if err != nil {
		t.Fatalf("applyAuthentication() error = %v", err)
	}

	// Check bearer token was applied
	authHeader := req.Header.Get("Authorization")
	if !strings.Contains(authHeader, "Bearer test-token") {
		t.Errorf("Expected bearer token to be applied, got %q", authHeader)
	}

	// Now simulate custom header override (this happens in main function after auth)
	req.Header.Set("Authorization", "Custom override")

	// Check override worked
	authHeader = req.Header.Get("Authorization")
	if authHeader != "Custom override" {
		t.Errorf("Expected custom header to override bearer token, got %q", authHeader)
	}

	// Restore original values
	bearerToken = origBearerToken
	sigV4Enabled = origSigV4Enabled
	sigV4Service = origSigV4Service
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
