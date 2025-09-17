package openapi

import (
	"context"
	"net/http"
	"testing"
)

func TestViewerMethodCompletions(t *testing.T) {
	tests := []struct {
		name            string
		openAPISpec     string
		path            string
		expectedMethods []string
		expectError     bool
	}{
		{
			name: "path with multiple methods",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						},
						"post": {
							"responses": {"201": {"description": "Created"}}
						},
						"delete": {
							"responses": {"204": {"description": "No Content"}}
						}
					}
				}
			}`,
			path:            "/users",
			expectedMethods: []string{"GET", "POST", "DELETE"},
		},
		{
			name: "path with single method",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/health": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:            "/health",
			expectedMethods: []string{"GET"},
		},
		{
			name: "nonexistent path",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:            "/nonexistent",
			expectedMethods: []string{}, // No matching path
		},
		{
			name: "path with path parameters",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users/{userId}": {
						"get": {
							"parameters": [
								{
									"name": "userId",
									"in": "path",
									"required": true,
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						},
						"put": {
							"parameters": [
								{
									"name": "userId",
									"in": "path",
									"required": true,
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "Updated"}}
						},
						"patch": {
							"parameters": [
								{
									"name": "userId",
									"in": "path",
									"required": true,
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "Patched"}}
						}
					}
				}
			}`,
			path:            "/users/{userId}",
			expectedMethods: []string{"GET", "PUT", "PATCH"},
		},
		{
			name: "multiple paths, only one matches",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						},
						"post": {
							"responses": {"201": {"description": "Created"}}
						}
					},
					"/users/{userId}": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						},
						"put": {
							"responses": {"200": {"description": "Updated"}}
						},
						"delete": {
							"responses": {"204": {"description": "No Content"}}
						}
					},
					"/orders": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:            "/users/{userId}",
			expectedMethods: []string{"GET", "PUT", "DELETE"},
		},
		{
			name: "path with all standard HTTP methods",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/api/resource": {
						"get": {
							"responses": {"200": {"description": "OK"}}
						},
						"post": {
							"responses": {"201": {"description": "Created"}}
						},
						"put": {
							"responses": {"200": {"description": "Updated"}}
						},
						"patch": {
							"responses": {"200": {"description": "Patched"}}
						},
						"delete": {
							"responses": {"204": {"description": "No Content"}}
						},
						"head": {
							"responses": {"200": {"description": "OK"}}
						},
						"options": {
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:            "/api/resource",
			expectedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viewer with mock HTTP client and empty URL
			viewer := NewViewer(&MockHTTPClient{}, "")

			// Load spec from bytes
			err := viewer.parser.LoadFromBytes([]byte(tt.openAPISpec))
			if err != nil {
				t.Fatalf("Failed to load OpenAPI spec: %v", err)
			}

			// Test MethodCompletions
			methods, err := viewer.MethodCompletions(context.Background(), tt.path)
			
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check if we got the expected methods
			if len(methods) != len(tt.expectedMethods) {
				t.Errorf("Expected %d methods, got %d: %v", len(tt.expectedMethods), len(methods), methods)
			}

			// Convert to map for easier checking
			methodMap := make(map[string]bool)
			for _, method := range methods {
				methodMap[method] = true
			}

			// Check each expected method is present
			for _, expected := range tt.expectedMethods {
				if !methodMap[expected] {
					t.Errorf("Expected method %q not found in results: %v", expected, methods)
				}
			}

			// Check no unexpected methods
			expectedMap := make(map[string]bool)
			for _, expected := range tt.expectedMethods {
				expectedMap[expected] = true
			}

			for _, method := range methods {
				if !expectedMap[method] {
					t.Errorf("Unexpected method %q found in results", method)
				}
			}
		})
	}
}

func TestViewerMethodCompletionsWithURL(t *testing.T) {
	// Test MethodCompletions with actual URL loading using our real 1Password Connect example
	// Use standard HTTP client for real URL
	viewer := NewViewer(
		&http.Client{}, 
		"https://raw.githubusercontent.com/konfig-sdks/openapi-examples/main/1-password/connect/openapi.yaml",
	)

	methods, err := viewer.MethodCompletions(
		context.Background(),
		"/vaults/{vaultUuid}/items",
	)

	if err != nil {
		t.Fatalf("Failed to get method completions: %v", err)
	}

	// The 1Password Connect API should have GET method for the items endpoint
	foundGet := false
	for _, method := range methods {
		if method == "GET" {
			foundGet = true
			break
		}
	}

	if !foundGet {
		t.Errorf("Expected 'GET' method for /vaults/{vaultUuid}/items endpoint, got: %v", methods)
	}

	// Should have at least one method
	if len(methods) == 0 {
		t.Error("Expected at least one method for /vaults/{vaultUuid}/items endpoint")
	}
}

func TestViewerMethodCompletionsErrors(t *testing.T) {
	// Use real HTTP client with invalid URL to test error handling
	viewer := NewViewer(&http.Client{}, "invalid-url")

	// Test with invalid URL
	_, err := viewer.MethodCompletions(context.Background(), "/test")
	if err == nil {
		t.Error("Expected error with invalid URL, got nil")
	}

	// Test with valid URL but invalid OpenAPI spec
	invalidSpec := `{invalid json}`
	err = viewer.parser.LoadFromBytes([]byte(invalidSpec))
	if err == nil {
		t.Error("Expected error with invalid OpenAPI spec, got nil")
	}
}