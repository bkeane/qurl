package http

import (
	"context"
	"net/http"
)

// HTTPExecutor defines the core HTTP execution interface
// This enables easy mocking and testing of HTTP operations
type HTTPExecutor interface {
	// Execute performs an HTTP request and prints response to stdout (CLI mode)
	Execute(ctx context.Context, path string) error

	// ExecuteForMCP performs an HTTP request and returns structured response (MCP mode)
	ExecuteForMCP(ctx context.Context, path string) (body string, headers map[string][]string, statusCode int, err error)

	// ShowDocs displays OpenAPI documentation for the given path and method
	ShowDocs(ctx context.Context, path, method string) error
}

// URLResolver defines interface for resolving target URLs
// Separates URL resolution logic for better testing
type URLResolver interface {
	ResolveURL(ctx context.Context, path string) (string, error)
}

// ResponseHandler defines interface for handling HTTP responses
// Allows testing response processing separately from request execution
type ResponseHandler interface {
	HandleResponse(resp *http.Response, method, targetURL string) error
	HandleResponseForMCP(resp *http.Response, method, targetURL string) (string, map[string][]string, int, error)
}

// HTTPClientProvider defines interface for the underlying HTTP client
// Enables testing with mock HTTP clients
type HTTPClientProvider interface {
	Do(req *http.Request) (*http.Response, error)
}

// OpenAPIProvider defines interface for OpenAPI operations
// Allows testing without real OpenAPI specs
type OpenAPIProvider interface {
	SetHeaders(ctx context.Context, req *http.Request, path, method string) error
	View(ctx context.Context, path, method string) (string, error)
	BaseURL(ctx context.Context) (string, error)
	GetServers() ([]string, error)
}