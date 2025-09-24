package http

import (
	"context"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
)

func TestURLResolver_ResolveURL(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		openapi   OpenAPIProvider
		path      string
		expected  string
		wantError bool
	}{
		{
			name:     "absolute URL should be returned as-is",
			config:   &config.Config{},
			path:     "https://api.example.com/users",
			expected: "https://api.example.com/users",
		},
		{
			name:   "relative path with server config",
			config: &config.Config{Server: "https://api.example.com"},
			path:   "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:   "relative path without leading slash",
			config: &config.Config{Server: "https://api.example.com"},
			path:   "users",
			expected: "https://api.example.com/users",
		},
		{
			name:   "server with path prefix",
			config: &config.Config{Server: "https://api.example.com/v1"},
			path:   "/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:   "server with trailing slash",
			config: &config.Config{Server: "https://api.example.com/"},
			path:   "/users",
			expected: "https://api.example.com/users",
		},
		{
			name: "openapi base URL fallback",
			config: &config.Config{},
			openapi: &mockOpenAPIWithBaseURL{
				mockOpenAPIProvider{baseURL: "https://api.example.com"},
			},
			path:     "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:      "no server URL available",
			config:    &config.Config{},
			path:      "/users",
			wantError: true,
		},
		{
			name:      "invalid path URL",
			config:    &config.Config{Server: "https://api.example.com"},
			path:      "://invalid",
			wantError: true,
		},
		{
			name:      "invalid server URL",
			config:    &config.Config{Server: "://invalid"},
			path:      "/users",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewURLResolver(tt.config, tt.openapi)

			result, err := resolver.ResolveURL(context.Background(), tt.path)

			if tt.wantError {
				if err == nil {
					t.Errorf("ResolveURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveURL() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ResolveURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestURLResolver_ServerIndex(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		openapi   OpenAPIProvider
		path      string
		expected  string
		wantError bool
	}{
		{
			name:     "server index 0",
			config:   &config.Config{Server: "0"},
			openapi:  &mockOpenAPIWithServers{
				mockOpenAPIProvider{servers: []string{"https://api.example.com", "https://staging.example.com"}},
			},
			path:     "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "server index 1",
			config:   &config.Config{Server: "1"},
			openapi:  &mockOpenAPIWithServers{
				mockOpenAPIProvider{servers: []string{"https://api.example.com", "https://staging.example.com"}},
			},
			path:     "/users",
			expected: "https://staging.example.com/users",
		},
		{
			name:      "server index out of range",
			config:    &config.Config{Server: "2"},
			openapi:   &mockOpenAPIWithServers{
				mockOpenAPIProvider{servers: []string{"https://api.example.com"}},
			},
			path:      "/users",
			wantError: true,
		},
		{
			name:      "server index without OpenAPI",
			config:    &config.Config{Server: "0"},
			path:      "/users",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewURLResolver(tt.config, tt.openapi)

			result, err := resolver.ResolveURL(context.Background(), tt.path)

			if tt.wantError {
				if err == nil {
					t.Errorf("ResolveURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveURL() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ResolveURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Mock OpenAPI provider that implements BaseURL method
type mockOpenAPIWithBaseURL struct {
	mockOpenAPIProvider
}

// Mock OpenAPI provider that implements GetServers method
type mockOpenAPIWithServers struct {
	mockOpenAPIProvider
}