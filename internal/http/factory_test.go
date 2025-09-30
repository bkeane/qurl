package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/rs/zerolog"
)

func TestClientFactory_CreateExecutor(t *testing.T) {
	logger := zerolog.New(nil)
	factory := NewClientFactory(logger)

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config without OpenAPI",
			config: &config.Config{
				Methods: []string{"GET"},
				Server:  "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "valid config with OpenAPI",
			config: &config.Config{
				Methods:    []string{"POST"},
				Server:     "https://example.com",
				OpenAPIURL: "https://example.com/openapi.json",
			},
			wantErr: false,
		},
		{
			name: "empty config",
			config: &config.Config{
				Methods: []string{"GET"},
			},
			wantErr: false, // Should not error, just won't have server URL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := factory.CreateExecutor(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && executor == nil {
				t.Error("CreateExecutor() returned nil executor without error")
			}
		})
	}
}

func TestClientFactory_CreateExecutorWithCustomClient(t *testing.T) {
	logger := zerolog.New(nil)
	factory := NewClientFactory(logger)

	// Mock HTTP client for testing
	mockClient := &mockHTTPClient{}
	mockOpenAPI := &mockOpenAPIProvider{}

	config := &config.Config{
		Methods: []string{"GET"},
		Server:  "https://example.com",
	}

	executor := factory.CreateExecutorWithCustomClient(config, mockClient, mockOpenAPI)

	if executor == nil {
		t.Error("CreateExecutorWithCustomClient() returned nil executor")
	}
}

// Global factory tests removed - no longer using global factory pattern

// Mock implementations for testing

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

type mockOpenAPIProvider struct {
	headers map[string]string
	viewResult string
	viewError error
	baseURL string
	baseURLError error
	servers []string
	serversError error
}

func (m *mockOpenAPIProvider) SetHeaders(ctx context.Context, req *http.Request, path, method string) error {
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
	return nil
}

func (m *mockOpenAPIProvider) View(ctx context.Context, path, method string) (string, error) {
	return m.viewResult, m.viewError
}

func (m *mockOpenAPIProvider) BaseURL(ctx context.Context) (string, error) {
	return m.baseURL, m.baseURLError
}

func (m *mockOpenAPIProvider) GetServers() ([]string, error) {
	return m.servers, m.serversError
}