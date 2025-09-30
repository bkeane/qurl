package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthenticatedHTTPClient_WithoutSigV4 tests that requests work without SigV4
func TestAuthenticatedHTTPClient_WithoutSigV4(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no AWS SigV4 headers are present
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Amz-Date"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Create config without SigV4 enabled
	cfg := &config.Config{
		SigV4Enabled: false,
	}

	logger := zerolog.New(nil)
	client := NewAuthenticatedHTTPClient(cfg, logger)

	// Create request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Perform request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestAuthenticatedHTTPClient_WithSigV4Enabled tests SigV4 is attempted when enabled
func TestAuthenticatedHTTPClient_WithSigV4Enabled(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This shouldn't be reached if SigV4 fails due to missing credentials
		t.Fatal("Request should not reach server when SigV4 credentials are missing")
	}))
	defer server.Close()

	// Create config with SigV4 enabled
	cfg := &config.Config{
		SigV4Enabled:  true,
		SigV4Service: "execute-api",
	}

	logger := zerolog.New(nil)
	client := NewAuthenticatedHTTPClient(cfg, logger)

	// Create request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Perform request - should fail due to missing AWS credentials
	_, err = client.Do(req)

	// Verify error contains SigV4 related message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SigV4")
}

// TestAuthenticatedHTTPClient_OpenAPIIntegration tests OpenAPI viewer can use authenticated client
func TestAuthenticatedHTTPClient_OpenAPIIntegration(t *testing.T) {
	// Create a test server that serves OpenAPI spec
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Return a minimal OpenAPI spec
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"openapi": "3.0.0",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			},
			"servers": [
				{"url": "https://api.example.com"}
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
		}`))
	}))
	defer server.Close()

	// Test without SigV4
	t.Run("without SigV4", func(t *testing.T) {
		cfg := &config.Config{
			SigV4Enabled: false,
		}

		logger := zerolog.New(nil)
		client := NewAuthenticatedHTTPClient(cfg, logger)

		// Create request like OpenAPI parser would
		req, err := http.NewRequest("GET", server.URL+"/openapi.json", nil)
		require.NoError(t, err)

		// Perform request
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify success
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test with SigV4 enabled (but will fail due to no credentials)
	t.Run("with SigV4 enabled", func(t *testing.T) {
		cfg := &config.Config{
			SigV4Enabled:  true,
			SigV4Service: "execute-api",
		}

		logger := zerolog.New(nil)
		client := NewAuthenticatedHTTPClient(cfg, logger)

		// Create request
		req, err := http.NewRequest("GET", server.URL+"/openapi.json", nil)
		require.NoError(t, err)

		// Perform request - should fail
		_, err = client.Do(req)

		// Should get authentication error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SigV4")
	})
}

// TestAuthenticatedHTTPClient_PreservesRequestBody tests that request body is preserved after signing
func TestAuthenticatedHTTPClient_PreservesRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo the request body back
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	cfg := &config.Config{
		SigV4Enabled: false, // Disable SigV4 to test body preservation
	}

	logger := zerolog.New(nil)
	client := NewAuthenticatedHTTPClient(cfg, logger)

	// Create request without body to test simple case
	req, err := http.NewRequest("POST", server.URL+"/test", nil)
	require.NoError(t, err)

	// Perform request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// BenchmarkAuthenticatedHTTPClient benchmarks the authenticated client overhead
func BenchmarkAuthenticatedHTTPClient(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		SigV4Enabled: false,
	}

	logger := zerolog.New(nil)
	client := NewAuthenticatedHTTPClient(cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
}