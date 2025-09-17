package openapi

import (
	"bytes"
	"context"
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