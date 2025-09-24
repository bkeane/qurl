package testutil

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/brendan.keane/qurl/internal/config"
	httpinternal "github.com/brendan.keane/qurl/internal/http"
	"github.com/rs/zerolog"
)

// MockHTTPClient provides a unified mock HTTP client implementation
// This replaces the 4+ different mock implementations scattered across test files
type MockHTTPClient struct {
	Response *http.Response
	Error    error
	Requests []*http.Request // Track all requests made
}

// Do implements the HTTPClientProvider interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Track the request for assertion purposes
	m.Requests = append(m.Requests, req)
	return m.Response, m.Error
}

// NewMockHTTPClient creates a mock HTTP client with the given response and error
func NewMockHTTPClient(body string, statusCode int, headers map[string]string, err error) *MockHTTPClient {
	var resp *http.Response
	if err == nil {
		resp = &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}

		// Set headers
		for key, value := range headers {
			resp.Header.Set(key, value)
		}
	}

	return &MockHTTPClient{
		Response: resp,
		Error:    err,
		Requests: make([]*http.Request, 0),
	}
}

// MockOpenAPIProvider provides a unified OpenAPI provider mock
type MockOpenAPIProvider struct {
	Headers         map[string]string
	HeadersError    error
	ViewResult      string
	ViewError       error
	BaseURLResult   string
	BaseURLError    error
	Servers         []string
	ServersError    error
	SetHeadersCalls []SetHeadersCall // Track calls for assertions
	ViewCalls       []ViewCall
}

type SetHeadersCall struct {
	Path   string
	Method string
}

type ViewCall struct {
	Path   string
	Method string
}

func (m *MockOpenAPIProvider) SetHeaders(ctx context.Context, req *http.Request, path, method string) error {
	m.SetHeadersCalls = append(m.SetHeadersCalls, SetHeadersCall{Path: path, Method: method})

	for key, value := range m.Headers {
		req.Header.Set(key, value)
	}
	return m.HeadersError
}

func (m *MockOpenAPIProvider) View(ctx context.Context, path, method string) (string, error) {
	m.ViewCalls = append(m.ViewCalls, ViewCall{Path: path, Method: method})
	return m.ViewResult, m.ViewError
}

func (m *MockOpenAPIProvider) BaseURL(ctx context.Context) (string, error) {
	return m.BaseURLResult, m.BaseURLError
}

func (m *MockOpenAPIProvider) GetServers() ([]string, error) {
	return m.Servers, m.ServersError
}

// NewMockOpenAPIProvider creates a mock OpenAPI provider with common defaults
func NewMockOpenAPIProvider() *MockOpenAPIProvider {
	return &MockOpenAPIProvider{
		Headers:         make(map[string]string),
		ViewResult:      "# API Documentation",
		BaseURLResult:   "https://api.example.com",
		Servers:         []string{"https://api.example.com"},
		SetHeadersCalls: make([]SetHeadersCall, 0),
		ViewCalls:       make([]ViewCall, 0),
	}
}

// MockURLResolver provides a mock URL resolver
type MockURLResolver struct {
	URL   string
	Error error
	Calls []string // Track all paths resolved
}

func (m *MockURLResolver) ResolveURL(ctx context.Context, path string) (string, error) {
	m.Calls = append(m.Calls, path)
	return m.URL, m.Error
}

// NewMockURLResolver creates a mock URL resolver
func NewMockURLResolver(url string, err error) *MockURLResolver {
	return &MockURLResolver{
		URL:   url,
		Error: err,
		Calls: make([]string, 0),
	}
}

// MockResponseHandler provides a mock response handler
type MockResponseHandler struct {
	Error             error
	MCPBody          string
	MCPHeaders       map[string][]string
	MCPStatusCode    int
	MCPError         error
	HandleCalls      []HandleCall
	HandleMCPCalls   []HandleCall
}

type HandleCall struct {
	Method    string
	TargetURL string
}

func (m *MockResponseHandler) HandleResponse(resp *http.Response, method, targetURL string) error {
	m.HandleCalls = append(m.HandleCalls, HandleCall{Method: method, TargetURL: targetURL})
	return m.Error
}

func (m *MockResponseHandler) HandleResponseForMCP(resp *http.Response, method, targetURL string) (string, map[string][]string, int, error) {
	m.HandleMCPCalls = append(m.HandleMCPCalls, HandleCall{Method: method, TargetURL: targetURL})
	if m.MCPError != nil {
		return "", nil, 0, m.MCPError
	}
	return m.MCPBody, m.MCPHeaders, m.MCPStatusCode, nil
}

// NewMockResponseHandler creates a mock response handler with defaults
func NewMockResponseHandler() *MockResponseHandler {
	return &MockResponseHandler{
		MCPBody:        `{"message": "success"}`,
		MCPHeaders:     map[string][]string{"Content-Type": {"application/json"}},
		MCPStatusCode:  200,
		HandleCalls:    make([]HandleCall, 0),
		HandleMCPCalls: make([]HandleCall, 0),
	}
}

// MockError provides a simple mock error implementation
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// NewMockError creates a mock error
func NewMockError(message string) *MockError {
	return &MockError{Message: message}
}

// Mock factory functions for common scenarios

// NewSuccessfulHTTPExecutor creates a mock HTTP executor that always succeeds
func NewSuccessfulHTTPExecutor(responseBody string) httpinternal.HTTPExecutor {
	// Use a discarded logger for tests
	logger := zerolog.New(io.Discard)
	factory := httpinternal.NewClientFactory(logger)

	mockHTTPClient := NewMockHTTPClient(responseBody, 200, map[string]string{"Content-Type": "application/json"}, nil)
	mockOpenAPI := NewMockOpenAPIProvider()

	return factory.CreateExecutorWithCustomClient(&config.Config{Methods: []string{"GET"}}, mockHTTPClient, mockOpenAPI)
}

// NewFailingHTTPExecutor creates a mock HTTP executor that always fails
func NewFailingHTTPExecutor(errorMessage string) httpinternal.HTTPExecutor {
	logger := zerolog.New(io.Discard)
	factory := httpinternal.NewClientFactory(logger)

	mockHTTPClient := NewMockHTTPClient("", 0, nil, NewMockError(errorMessage))
	mockOpenAPI := NewMockOpenAPIProvider()

	return factory.CreateExecutorWithCustomClient(&config.Config{Methods: []string{"GET"}}, mockHTTPClient, mockOpenAPI)
}