package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/rs/zerolog"
)

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		mockResponse   *http.Response
		mockError      error
		resolveURL     string
		resolveError   error
		responseError  error
		path           string
		expectedError  bool
	}{
		{
			name: "successful GET request",
			config: &config.Config{
				Methods: []string{"GET"},
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
				Header:     make(http.Header),
			},
			resolveURL: "https://api.example.com/users",
			path:       "/users",
		},
		{
			name: "URL resolution failure",
			config: &config.Config{
				Methods: []string{"GET"},
			},
			resolveError:  &mockError{msg: "failed to resolve URL"},
			path:          "/users",
			expectedError: true,
		},
		{
			name: "HTTP request failure",
			config: &config.Config{
				Methods: []string{"GET"},
			},
			mockError:     &mockError{msg: "network error"},
			resolveURL:    "https://api.example.com/users",
			path:          "/users",
			expectedError: true,
		},
		{
			name: "response handling failure",
			config: &config.Config{
				Methods: []string{"GET"},
			},
			mockResponse: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
				Header:     make(http.Header),
			},
			resolveURL:    "https://api.example.com/users",
			responseError: &mockError{msg: "response handling failed"},
			path:          "/users",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.New(nil)

			// Create mocks
			mockHTTPClient := &mockHTTPClient{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			mockResolver := &mockURLResolver{
				url: tt.resolveURL,
				err: tt.resolveError,
			}

			mockResponseHandler := &mockResponseHandler{
				err: tt.responseError,
			}

			// Create executor with mocked dependencies
			executor := NewExecutorWithDependencies(
				logger,
				mockHTTPClient,
				nil, // no OpenAPI needed for this test
				mockResolver,
				mockResponseHandler,
				tt.config,
			)

			// Execute the request
			err := executor.Execute(context.Background(), tt.path)

			// Check results
			if tt.expectedError {
				if err == nil {
					t.Error("Execute() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExecutor_ExecuteForMCP(t *testing.T) {
	logger := zerolog.New(nil)

	config := &config.Config{
		Methods: []string{"GET"},
	}

	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	mockHTTPClient := &mockHTTPClient{
		response: mockResponse,
	}

	mockResolver := &mockURLResolver{
		url: "https://api.example.com/users",
	}

	mockResponseHandler := &responseHandler{
		logger: logger,
		config: config,
	}

	executor := NewExecutorWithDependencies(
		logger,
		mockHTTPClient,
		nil,
		mockResolver,
		mockResponseHandler,
		config,
	)

	body, headers, statusCode, err := executor.ExecuteForMCP(context.Background(), "/users")

	if err != nil {
		t.Errorf("ExecuteForMCP() unexpected error: %v", err)
	}

	if body != `{"message": "success"}` {
		t.Errorf("ExecuteForMCP() body = %v, expected %v", body, `{"message": "success"}`)
	}

	if statusCode != 200 {
		t.Errorf("ExecuteForMCP() statusCode = %v, expected %v", statusCode, 200)
	}

	if headers == nil {
		t.Error("ExecuteForMCP() headers should not be nil")
	}
}

func TestExecutor_ShowDocs(t *testing.T) {
	tests := []struct {
		name          string
		openapi       OpenAPIProvider
		path          string
		method        string
		expectedError bool
	}{
		{
			name:          "no OpenAPI provider",
			path:          "/users",
			method:        "GET",
			expectedError: true,
		},
		{
			name: "successful docs retrieval",
			openapi: &mockOpenAPIProvider{
				viewResult: "# API Documentation\n\nGET /users",
			},
			path:   "/users",
			method: "GET",
		},
		{
			name: "OpenAPI view error",
			openapi: &mockOpenAPIProvider{
				viewError: &mockError{msg: "spec not found"},
			},
			path:          "/users",
			method:        "GET",
			expectedError: true,
		},
		{
			name: "empty path and method",
			openapi: &mockOpenAPIProvider{
				viewResult: "# API Documentation",
			},
			path:   "",
			method: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.New(nil)
			config := &config.Config{Methods: []string{"GET"}}

			executor := NewExecutorWithDependencies(
				logger,
				nil, // no HTTP client needed
				tt.openapi,
				nil, // no URL resolver needed
				nil, // no response handler needed
				config,
			)

			err := executor.ShowDocs(context.Background(), tt.path, tt.method)

			if tt.expectedError {
				if err == nil {
					t.Error("ShowDocs() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ShowDocs() unexpected error: %v", err)
				}
			}
		})
	}
}

// Mock implementations for testing

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

type mockURLResolver struct {
	url string
	err error
}

func (m *mockURLResolver) ResolveURL(ctx context.Context, path string) (string, error) {
	return m.url, m.err
}

type mockResponseHandler struct {
	err error
	mcpBody string
	mcpHeaders map[string][]string
	mcpStatusCode int
	mcpErr error
}

func (m *mockResponseHandler) HandleResponse(resp *http.Response, method, targetURL string) error {
	return m.err
}

func (m *mockResponseHandler) HandleResponseForMCP(resp *http.Response, method, targetURL string) (string, map[string][]string, int, error) {
	if m.mcpErr != nil {
		return "", nil, 0, m.mcpErr
	}
	return m.mcpBody, m.mcpHeaders, m.mcpStatusCode, nil
}