package openapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestViewerView(t *testing.T) {
	// Use real HTTP client for real URL test
	viewer := NewViewer(&http.Client{}, "https://petstore3.swagger.io/api/v3/openapi.json")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name           string
		path           string
		method         string
		expectIndex    bool
		expectContains []string
	}{
		{
			name:        "View all endpoints",
			path:        "*",
			method:      "*",
			expectIndex: true,
			expectContains: []string{
				"Swagger Petstore",
				"/pet",
				"/store",
				"/user",
			},
		},
		{
			name:        "View specific endpoint",
			path:        "/pet/{petId}",
			method:      "GET",
			expectIndex: false,
			expectContains: []string{
				"GET",
				"/pet/{petId}",
				"Parameters",
				"Responses",
			},
		},
		{
			name:        "View path prefix",
			path:        "/pet*",
			method:      "*",
			expectIndex: true,
			expectContains: []string{
				"/pet",
				"GET",
				"POST",
				"PUT",
			},
		},
		{
			name:        "View all POST methods",
			path:        "*",
			method:      "POST",
			expectIndex: true,
			expectContains: []string{
				"POST",
				"/pet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := viewer.View(ctx, tt.path, tt.method)
			if err != nil {
				t.Fatalf("Failed to view: %v", err)
			}

			if output == "" {
				t.Error("Expected non-empty output")
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't", expected)
				}
			}

			if tt.expectIndex {
				if !strings.Contains(output, "Endpoints") {
					t.Error("Expected index view to contain 'Endpoints' section")
				}
			} else {
				if !strings.Contains(output, "Parameters") && !strings.Contains(output, "Responses") {
					t.Error("Expected detailed view to contain Parameters or Responses")
				}
			}
		})
	}
}

func TestViewerViewFromBytes(t *testing.T) {
	minimalSpec := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				},
				"post": {
					"summary": "Create test",
					"requestBody": {
						"required": true,
						"content": {
							"application/json": {
								"schema": {
									"type": "object"
								}
							}
						}
					},
					"responses": {
						"201": {
							"description": "Created"
						}
					}
				}
			},
			"/test/{id}": {
				"get": {
					"summary": "Get test by ID",
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": {
								"type": "string"
							}
						}
					],
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`)

	viewer := NewViewer(&MockHTTPClient{}, "")

	tests := []struct {
		name           string
		path           string
		method         string
		expectContains []string
	}{
		{
			name:   "View all from bytes",
			path:   "*",
			method: "*",
			expectContains: []string{
				"Test API",
				"/test",
				"GET",
				"POST",
			},
		},
		{
			name:   "View specific from bytes",
			path:   "/test/{id}",
			method: "GET",
			expectContains: []string{
				"GET",
				"/test/{id}",
				"Parameters",
				"id",
				"*required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := viewer.ViewFromBytes(minimalSpec, tt.path, tt.method)
			if err != nil {
				t.Fatalf("Failed to view from bytes: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't", expected)
				}
			}
		})
	}
}

func TestViewerNoMatchingPaths(t *testing.T) {
	minimalSpec := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`)

	viewer := NewViewer(&MockHTTPClient{}, "")
	
	output, err := viewer.ViewFromBytes(minimalSpec, "/nonexistent", "GET")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "No endpoints found") {
		t.Errorf("Expected 'No endpoints found' message, got: %s", output)
	}
}