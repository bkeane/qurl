package openapi

import (
	"context"
	"net/http"
	"testing"
)

func TestViewerParamCompletions(t *testing.T) {
	tests := []struct {
		name           string
		openAPISpec    string
		path           string
		method         string
		expectedParams []string
		expectError    bool
	}{
		{
			name: "endpoint with query parameters",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"parameters": [
								{
									"name": "limit",
									"in": "query",
									"schema": {"type": "integer"}
								},
								{
									"name": "offset",
									"in": "query", 
									"schema": {"type": "integer"}
								},
								{
									"name": "filter",
									"in": "query",
									"schema": {"type": "string"}
								},
								{
									"name": "userId",
									"in": "path",
									"schema": {"type": "string"}
								},
								{
									"name": "authorization",
									"in": "header",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:           "/users",
			method:         "GET",
			expectedParams: []string{"limit", "offset", "filter"},
		},
		{
			name: "endpoint with no query parameters",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/health": {
						"get": {
							"parameters": [
								{
									"name": "userId",
									"in": "path",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:           "/health",
			method:         "GET",
			expectedParams: []string{}, // No query parameters
		},
		{
			name: "endpoint with mixed parameter types",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/items/{itemId}": {
						"get": {
							"parameters": [
								{
									"name": "itemId",
									"in": "path",
									"required": true,
									"schema": {"type": "string"}
								},
								{
									"name": "include",
									"in": "query",
									"schema": {"type": "string"}
								},
								{
									"name": "format",
									"in": "query",
									"schema": {"type": "string", "enum": ["json", "xml"]}
								},
								{
									"name": "x-api-key",
									"in": "header",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:           "/items/{itemId}",
			method:         "GET",
			expectedParams: []string{"include", "format"},
		},
		{
			name: "multiple methods same path different parameters",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"parameters": [
								{
									"name": "search",
									"in": "query",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						},
						"post": {
							"parameters": [
								{
									"name": "notify",
									"in": "query",
									"schema": {"type": "boolean"}
								}
							],
							"responses": {"201": {"description": "Created"}}
						}
					}
				}
			}`,
			path:           "/users",
			method:         "POST",
			expectedParams: []string{"notify"},
		},
		{
			name: "path with global parameters",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/orders/{orderId}": {
						"parameters": [
							{
								"name": "orderId",
								"in": "path",
								"required": true,
								"schema": {"type": "string"}
							},
							{
								"name": "version",
								"in": "query",
								"schema": {"type": "string"}
							}
						],
						"get": {
							"parameters": [
								{
									"name": "expand",
									"in": "query",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:           "/orders/{orderId}",
			method:         "GET",
			expectedParams: []string{"version", "expand"}, // Both global and operation parameters
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
			path:           "/nonexistent",
			method:         "GET",
			expectedParams: []string{}, // No matching path
		},
		{
			name: "duplicate parameters (should deduplicate)",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/test": {
						"parameters": [
							{
								"name": "shared",
								"in": "query",
								"schema": {"type": "string"}
							}
						],
						"get": {
							"parameters": [
								{
									"name": "shared",
									"in": "query",
									"schema": {"type": "string"}
								},
								{
									"name": "unique",
									"in": "query",
									"schema": {"type": "string"}
								}
							],
							"responses": {"200": {"description": "OK"}}
						}
					}
				}
			}`,
			path:           "/test",
			method:         "GET",
			expectedParams: []string{"shared", "unique"}, // Should deduplicate "shared"
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

			// Test ParamCompletions
			params, err := viewer.ParamCompletions(context.Background(), tt.path, tt.method)
			
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check if we got the expected parameters
			if len(params) != len(tt.expectedParams) {
				t.Errorf("Expected %d parameters, got %d: %v", len(tt.expectedParams), len(params), params)
			}

			// Convert to map for easier checking
			paramMap := make(map[string]bool)
			for _, param := range params {
				paramMap[param] = true
			}

			// Check each expected parameter is present
			for _, expected := range tt.expectedParams {
				if !paramMap[expected] {
					t.Errorf("Expected parameter %q not found in results: %v", expected, params)
				}
			}

			// Check no unexpected parameters
			expectedMap := make(map[string]bool)
			for _, expected := range tt.expectedParams {
				expectedMap[expected] = true
			}

			for _, param := range params {
				if !expectedMap[param] {
					t.Errorf("Unexpected parameter %q found in results", param)
				}
			}
		})
	}
}

func TestViewerParamCompletionsWithURL(t *testing.T) {
	// Test ParamCompletions with actual URL loading using our real 1Password Connect example
	// Use standard HTTP client for real URL
	viewer := NewViewer(
		&http.Client{}, 
		"https://raw.githubusercontent.com/konfig-sdks/openapi-examples/main/1-password/connect/openapi.yaml",
	)

	params, err := viewer.ParamCompletions(
		context.Background(),
		"/vaults/{vaultUuid}/items",
		"GET",
	)

	if err != nil {
		t.Fatalf("Failed to get parameter completions: %v", err)
	}

	// The 1Password Connect API should have a "filter" query parameter for the items endpoint
	found := false
	for _, param := range params {
		if param == "filter" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'filter' parameter for /vaults/{vaultUuid}/items GET endpoint, got: %v", params)
	}
}

func TestViewerParamCompletionsErrors(t *testing.T) {
	// Use real HTTP client with invalid URL to test error handling
	viewer := NewViewer(&http.Client{}, "invalid-url")

	// Test with invalid URL
	_, err := viewer.ParamCompletions(context.Background(), "/test", "GET")
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