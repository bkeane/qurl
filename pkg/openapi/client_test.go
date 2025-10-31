package openapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	Response *http.Response
	Error    error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Response, nil
}

func TestParserWithCustomClient(t *testing.T) {
	// Mock OpenAPI spec
	openAPISpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(openAPISpec))),
		},
	}

	// Test parser with custom client
	parser := NewParserWithClient(mockClient)

	err := parser.LoadFromURL(context.Background(), "lambda://api-docs/openapi.json")
	if err != nil {
		t.Fatalf("LoadFromURL failed: %v", err)
	}

	// Verify the spec was loaded correctly
	paths, err := parser.GetPaths("*", "*")
	if err != nil {
		t.Fatalf("GetPaths failed: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("Expected at least one path, got 0")
	}

	// Check that our test path exists
	found := false
	for _, path := range paths {
		if path.Path == "/test" && path.Method == "GET" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find GET /test endpoint")
	}
}

func TestViewerWithCustomClient(t *testing.T) {
	// Mock OpenAPI spec
	openAPISpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Lambda API", "version": "1.0.0"},
		"paths": {
			"/lambda-endpoint": {
				"post": {
					"summary": "Lambda endpoint",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(openAPISpec))),
		},
	}

	// Test viewer with custom client
	viewer := NewViewer(mockClient, "lambda://spec-provider/spec.json")

	output, err := viewer.View(context.Background(), "*", "*")
	if err != nil {
		t.Fatalf("View failed: %v", err)
	}

	// Check that the output contains our expected content
	if output == "" {
		t.Fatal("Expected non-empty output")
	}

	// Should contain the API title
	if !contains(output, "Lambda API") {
		t.Error("Expected output to contain 'Lambda API'")
	}

	// Should contain the endpoint
	if !contains(output, "/lambda-endpoint") {
		t.Error("Expected output to contain '/lambda-endpoint'")
	}
}

func TestClientError(t *testing.T) {
	// Create mock client that returns an error
	mockClient := &MockHTTPClient{
		Error: http.ErrServerClosed,
	}

	parser := NewParserWithClient(mockClient)

	err := parser.LoadFromURL(context.Background(), "lambda://failing-service/spec.json")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	// Error should be wrapped
	if !contains(err.Error(), "fetching OpenAPI spec") {
		t.Errorf("Expected error to mention 'fetching OpenAPI spec', got: %v", err)
	}
}

func TestHTTPError(t *testing.T) {
	// Create mock client that returns HTTP error
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader([]byte("Not Found"))),
		},
	}

	parser := NewParserWithClient(mockClient)

	err := parser.LoadFromURL(context.Background(), "lambda://missing-service/spec.json")
	if err == nil {
		t.Fatal("Expected an error for 404 response, got nil")
	}

	// Error should mention status code
	if !contains(err.Error(), "404") {
		t.Errorf("Expected error to mention '404', got: %v", err)
	}
}

func TestBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		specURL     string
		specContent string
		expected    string
		expectError bool
	}{
		{
			name:    "spec with servers defined",
			specURL: "https://api.example.com/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"servers": [{"url": "https://api.example.com/v1"}],
				"paths": {}
			}`,
			expected: "https://api.example.com/v1",
		},
		{
			name:    "spec with no servers - conservative fallback to host only",
			specURL: "https://prod.kaixo.io/binnit/main/binnit/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Binnit API", "version": "1.0.0"},
				"paths": {}
			}`,
			expected: "https://prod.kaixo.io", // Conservative fallback: just scheme + host
		},
		{
			name:    "spec with no servers - simple case",
			specURL: "https://api.example.com/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {}
			}`,
			expected: "https://api.example.com", // Conservative fallback: just scheme + host
		},
		{
			name:    "spec with relative server URL",
			specURL: "https://api.example.com/docs/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"servers": [{"url": "/api/v2"}],
				"paths": {}
			}`,
			expected: "https://api.example.com/api/v2",
		},
		{
			name:    "spec with empty server URL",
			specURL: "https://api.example.com/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"servers": [{"url": ""}],
				"paths": {}
			}`,
			expected: "https://api.example.com", // Falls back to spec URL host only
		},
		{
			name:    "spec with complex path fallback",
			specURL: "https://docs.company.com/api/v2/specs/service1.yaml",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Service API", "version": "2.0.0"},
				"paths": {}
			}`,
			expected: "https://docs.company.com", // Conservative: ignore complex paths
		},
		{
			name:    "spec with server containing full API path",
			specURL: "https://cdn.example.com/specs/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "API", "version": "1.0.0"},
				"servers": [{"url": "https://api.example.com/v1/services"}],
				"paths": {}
			}`,
			expected: "https://api.example.com/v1/services", // Use explicit server URL
		},
		{
			name:    "spec with relative server URL without leading slash",
			specURL: "https://api.example.com/docs/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "API", "version": "1.0.0"},
				"servers": [{"url": "api/v1"}],
				"paths": {}
			}`,
			expected: "https://api.example.com/api/v1", // Add leading slash to relative URL
		},
		{
			name:    "spec with lambda:// server URL",
			specURL: "lambda://binnit-main-src/openapi.json",
			specContent: `{
				"openapi": "3.0.0",
				"info": {"title": "Binnit API", "version": "1.0.0"},
				"servers": [{"url": "lambda://binnit-main-src"}],
				"paths": {}
			}`,
			expected: "lambda://binnit-main-src", // Use lambda:// URL as-is (absolute)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP client with the spec content
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader([]byte(tt.specContent))),
				},
			}

			// Create viewer with the mock client and spec URL
			viewer := NewViewer(mockClient, tt.specURL)

			// Call BaseURL
			result, err := viewer.BaseURL(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected BaseURL %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBaseURLErrors(t *testing.T) {
	tests := []struct {
		name        string
		specURL     string
		clientError error
		httpStatus  int
		expectError string
	}{
		{
			name:        "client error",
			specURL:     "https://api.example.com/openapi.json",
			clientError: fmt.Errorf("network error"),
			expectError: "network error",
		},
		{
			name:        "HTTP error",
			specURL:     "https://api.example.com/openapi.json",
			httpStatus:  404,
			expectError: "404",
		},
		{
			name:        "invalid spec URL",
			specURL:     "://invalid-url",
			expectError: "missing protocol scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient *MockHTTPClient

			if tt.clientError != nil {
				mockClient = &MockHTTPClient{Error: tt.clientError}
			} else if tt.httpStatus != 0 {
				mockClient = &MockHTTPClient{
					Response: &http.Response{
						StatusCode: tt.httpStatus,
						Body:       io.NopCloser(bytes.NewReader([]byte("Error"))),
					},
				}
			} else {
				mockClient = &MockHTTPClient{
					Response: &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
					},
				}
			}

			viewer := NewViewer(mockClient, tt.specURL)
			_, err := viewer.BaseURL(context.Background())

			if err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.expectError)
				return
			}

			if !contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
