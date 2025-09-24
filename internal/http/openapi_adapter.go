package http

import (
	"context"
	"net/http"

	"github.com/brendan.keane/qurl/pkg/openapi"
)

// openAPIAdapter adapts the existing openapi.Viewer to our OpenAPIProvider interface
// This maintains backward compatibility while enabling testable interfaces
type openAPIAdapter struct {
	viewer *openapi.Viewer
}

// NewOpenAPIAdapter wraps an existing openapi.Viewer to implement OpenAPIProvider
func NewOpenAPIAdapter(viewer *openapi.Viewer) OpenAPIProvider {
	return &openAPIAdapter{viewer: viewer}
}

// SetHeaders sets headers based on the OpenAPI specification
func (a *openAPIAdapter) SetHeaders(ctx context.Context, req *http.Request, path, method string) error {
	return a.viewer.SetHeaders(ctx, req, path, method)
}

// View returns the OpenAPI documentation for the given path and method
func (a *openAPIAdapter) View(ctx context.Context, path, method string) (string, error) {
	return a.viewer.View(ctx, path, method)
}

// BaseURL returns the base URL from the OpenAPI specification
func (a *openAPIAdapter) BaseURL(ctx context.Context) (string, error) {
	return a.viewer.BaseURL(ctx)
}

// GetServers returns the server URLs as strings from the OpenAPI specification
func (a *openAPIAdapter) GetServers() ([]string, error) {
	servers, err := a.viewer.GetServers()
	if err != nil {
		return nil, err
	}

	// Convert server objects to string URLs
	serverURLs := make([]string, len(servers))
	for i, server := range servers {
		serverURLs[i] = server.URL
	}

	return serverURLs, nil
}