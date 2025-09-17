package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSetHeaders(t *testing.T) {
	tests := []struct {
		name           string
		openAPISpec    string
		path           string
		method         string
		expectedAccept string
		expectError    bool
	}{
		{
			name: "sets Accept header for JSON response",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {
											"schema": {"type": "object"}
										}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/users",
			method:         "GET",
			expectedAccept: "application/json",
			expectError:    false,
		},
		{
			name: "sets multiple content types in Accept header",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/files": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {
											"schema": {"type": "object"}
										},
										"text/plain": {
											"schema": {"type": "string"}
										}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/files",
			method:         "GET",
			expectedAccept: "application/json, text/plain",
			expectError:    false,
		},
		{
			name: "prioritizes 2xx response codes",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/items": {
						"post": {
							"responses": {
								"201": {
									"description": "Created",
									"content": {
										"application/json": {
											"schema": {"type": "object"}
										}
									}
								},
								"400": {
									"description": "Bad Request",
									"content": {
										"application/problem+json": {
											"schema": {"type": "object"}
										}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/items",
			method:         "POST",
			expectedAccept: "application/json",
			expectError:    false,
		},
		{
			name: "uses non-2xx responses when no 2xx defined",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/error": {
						"delete": {
							"responses": {
								"404": {
									"description": "Not Found",
									"content": {
										"application/problem+json": {
											"schema": {"type": "object"}
										}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/error",
			method:         "DELETE",
			expectedAccept: "application/problem+json",
			expectError:    false,
		},
		{
			name: "no header when path not found",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/nonexistent",
			method:         "GET",
			expectedAccept: "",
			expectError:    false,
		},
		{
			name: "no header when method not found",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/users",
			method:         "POST",
			expectedAccept: "",
			expectError:    false,
		},
		{
			name: "handles XML content type",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/data": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/xml": {
											"schema": {"type": "object"}
										}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/data",
			method:         "GET",
			expectedAccept: "application/xml",
			expectError:    false,
		},
		{
			name: "case-insensitive method matching",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/users": {
						"get": {
							"responses": {
								"200": {
									"description": "Success",
									"content": {
										"application/json": {}
									}
								}
							}
						}
					}
				}
			}`,
			path:           "/users",
			method:         "get",
			expectedAccept: "application/json",
			expectError:    false,
		},
		{
			name: "no header when response has no content",
			openAPISpec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test API", "version": "1.0.0"},
				"paths": {
					"/status": {
						"head": {
							"responses": {
								"204": {
									"description": "No Content"
								}
							}
						}
					}
				}
			}`,
			path:           "/status",
			method:         "HEAD",
			expectedAccept: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server to serve the OpenAPI spec
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.openAPISpec))
			}))
			defer server.Close()

			// Create HTTP client
			httpClient := &http.Client{Timeout: 10 * time.Second}

			// Create viewer with test server URL
			viewer := NewViewer(httpClient, server.URL)

			// Create a test request
			req, _ := http.NewRequest(tt.method, "http://example.com"+tt.path, nil)

			// Call SetHeaders
			ctx := context.Background()
			err := viewer.SetHeaders(ctx, req, tt.path, tt.method)

			// Check error
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check Accept header
			actualAccept := req.Header.Get("Accept")
			if actualAccept != tt.expectedAccept {
				t.Errorf("Expected Accept header %q, got %q", tt.expectedAccept, actualAccept)
			}
		})
	}
}

func TestSetHeadersWithoutSpec(t *testing.T) {
	// Test that SetHeaders works when no spec URL is provided
	viewer := &Viewer{
		parser:    NewParser(),
		displayer: nil,
		specURL:   "", // No spec URL
	}

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)
	ctx := context.Background()

	err := viewer.SetHeaders(ctx, req, "/test", "GET")
	if err != nil {
		t.Errorf("Unexpected error when no spec URL: %v", err)
	}

	// Should not set any headers
	if req.Header.Get("Accept") != "" {
		t.Errorf("Expected no Accept header when no spec, got %q", req.Header.Get("Accept"))
	}
}

func TestSetHeadersPreservesExistingHeaders(t *testing.T) {
	openAPISpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {}
							}
						}
					}
				}
			}
		}
	}`

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPISpec))
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	viewer := NewViewer(httpClient, server.URL)

	// Create request with existing headers
	req, _ := http.NewRequest("GET", "http://example.com/users", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Custom-Header", "custom-value")

	ctx := context.Background()
	err := viewer.SetHeaders(ctx, req, "/users", "GET")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that existing headers are preserved
	if req.Header.Get("Authorization") != "Bearer token123" {
		t.Errorf("Authorization header was modified")
	}
	if req.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header was modified")
	}
	// And new header is added
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Accept header not set correctly")
	}
}

func TestSetHeadersWithMultipleResponseCodes(t *testing.T) {
	openAPISpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"paths": {
			"/resource": {
				"get": {
					"responses": {
						"200": {
							"description": "Success",
							"content": {
								"application/json": {},
								"application/hal+json": {}
							}
						},
						"206": {
							"description": "Partial Content",
							"content": {
								"application/json": {}
							}
						},
						"400": {
							"description": "Bad Request",
							"content": {
								"application/problem+json": {}
							}
						}
					}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPISpec))
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	viewer := NewViewer(httpClient, server.URL)

	req, _ := http.NewRequest("GET", "http://example.com/resource", nil)
	ctx := context.Background()

	err := viewer.SetHeaders(ctx, req, "/resource", "GET")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should only include 2xx response content types, not 400
	acceptHeader := req.Header.Get("Accept")
	if !strings.Contains(acceptHeader, "application/json") {
		t.Errorf("Expected application/json in Accept header, got %q", acceptHeader)
	}
	if !strings.Contains(acceptHeader, "application/hal+json") {
		t.Errorf("Expected application/hal+json in Accept header, got %q", acceptHeader)
	}
	if strings.Contains(acceptHeader, "application/problem+json") {
		t.Errorf("Should not include error response content type in Accept header, got %q", acceptHeader)
	}
}