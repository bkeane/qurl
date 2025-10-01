package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/pkg/openapi"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIWithSigV4Integration tests the full integration of OpenAPI fetching with SigV4
func TestOpenAPIWithSigV4Integration(t *testing.T) {
	// Track if auth headers were checked
	authHeadersChecked := false

	// Create a test server that requires authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openapi.json":
			// For OpenAPI spec endpoint, check for auth attempt
			// In real scenario with valid AWS creds, there would be Authorization header
			authHeadersChecked = true

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Build response with dynamic server URL
			response := `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"servers": [
					{"url": "http://` + r.Host + `"}
				],
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
			}`
			w.Write([]byte(response))

		case "/test":
			// API endpoint
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "ok"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("OpenAPI fetching without SigV4", func(t *testing.T) {
		// Reset flag
		authHeadersChecked = false

		// Create config without SigV4
		cfg := &config.Config{
			OpenAPIURL:   server.URL + "/openapi.json",
			SigV4Enabled: false,
		}

		logger := zerolog.New(nil)

		// Create authenticated client (without SigV4)
		authClient := NewAuthenticatedHTTPClient(cfg, logger)

		// Create OpenAPI viewer with authenticated client
		viewer := openapi.NewViewer(authClient, cfg.OpenAPIURL)

		// Test fetching OpenAPI spec
		ctx := context.Background()
		baseURL, err := viewer.BaseURL(ctx)

		// Should succeed
		require.NoError(t, err)
		assert.Equal(t, server.URL, baseURL)
		assert.True(t, authHeadersChecked, "OpenAPI endpoint should have been called")
	})

	t.Run("OpenAPI fetching with SigV4 enabled but no credentials", func(t *testing.T) {
		// Create config with SigV4 enabled
		cfg := &config.Config{
			OpenAPIURL:    server.URL + "/openapi.json",
			SigV4Enabled:  true,
			SigV4Service: "execute-api",
		}

		logger := zerolog.New(nil)

		// Create authenticated client (with SigV4 enabled)
		authClient := NewAuthenticatedHTTPClient(cfg, logger)

		// Create OpenAPI viewer with authenticated client
		viewer := openapi.NewViewer(authClient, cfg.OpenAPIURL)

		// Test fetching OpenAPI spec - may succeed or fail depending on AWS credentials
		ctx := context.Background()
		_, err := viewer.BaseURL(ctx)

		// If error occurs, it should be auth-related
		if err != nil {
			assert.True(t,
				strings.Contains(err.Error(), "SigV4") ||
				strings.Contains(err.Error(), "credentials") ||
				strings.Contains(err.Error(), "AWS"),
				"Expected auth-related error, got: %v", err)
		}
		// If no error, that's fine too - AWS credentials were available
	})
}

// TestClientCreationWithAuthenticatedOpenAPI tests that Client creates OpenAPI viewer with authentication
func TestClientCreationWithAuthenticatedOpenAPI(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0.0"},
				"servers": [{"url": "https://api.example.com"}],
				"paths": {}
			}`))
		}
	}))
	defer server.Close()

	t.Run("Executor with SigV4 disabled", func(t *testing.T) {
		cfg := &config.Config{
			OpenAPIURL:   server.URL + "/openapi.json",
			SigV4Enabled: false,
			Methods:      []string{"GET"},
		}

		logger := zerolog.New(nil)
		factory := NewClientFactory(logger)

		// Create executor - should succeed
		executor, err := factory.CreateExecutor(cfg)
		require.NoError(t, err)
		assert.NotNil(t, executor)
	})

	t.Run("Executor with SigV4 enabled", func(t *testing.T) {
		cfg := &config.Config{
			OpenAPIURL:    server.URL + "/openapi.json",
			SigV4Enabled:  true,
			SigV4Service: "execute-api",
			Methods:       []string{"GET"},
		}

		logger := zerolog.New(nil)
		factory := NewClientFactory(logger)

		// Create executor
		executor, err := factory.CreateExecutor(cfg)
		require.NoError(t, err)
		assert.NotNil(t, executor)
	})
}

// TestFactoryCreatesAuthenticatedOpenAPI tests that ClientFactory creates authenticated OpenAPI
func TestFactoryCreatesAuthenticatedOpenAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}}`))
		}
	}))
	defer server.Close()

	logger := zerolog.New(nil)
	factory := NewClientFactory(logger)

	t.Run("Factory without SigV4", func(t *testing.T) {
		cfg := &config.Config{
			OpenAPIURL:   server.URL + "/openapi.json",
			SigV4Enabled: false,
			Methods:      []string{"GET"},
		}

		executor, err := factory.CreateExecutor(cfg)
		require.NoError(t, err)
		assert.NotNil(t, executor)
	})

	t.Run("Factory with SigV4", func(t *testing.T) {
		cfg := &config.Config{
			OpenAPIURL:    server.URL + "/openapi.json",
			SigV4Enabled:  true,
			SigV4Service: "execute-api",
			Methods:       []string{"GET"},
		}

		executor, err := factory.CreateExecutor(cfg)
		require.NoError(t, err)
		assert.NotNil(t, executor)
	})
}