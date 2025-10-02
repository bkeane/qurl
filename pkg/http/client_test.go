package http

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPRequestToLambdaEvent(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "GET request with query params",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "lambda://my-function/users?limit=10&offset=0", nil)
				req.Header.Set("Authorization", "Bearer token123")
				return req
			}(),
		},
		{
			name: "POST request with body",
			req: func() *http.Request {
				body := strings.NewReader(`{"name":"test"}`)
				req, _ := http.NewRequest("POST", "lambda://my-function/users", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := httpRequestToLambdaEvent(tt.req)
			if err != nil {
				t.Fatalf("httpRequestToLambdaEvent() error = %v", err)
			}

			// Check basic fields
			if event.Version != "2.0" {
				t.Errorf("Expected version 2.0, got %v", event.Version)
			}

			if event.RawPath != tt.req.URL.Path {
				t.Errorf("Expected path %s, got %v", tt.req.URL.Path, event.RawPath)
			}

			// Check method in request context
			if event.RequestContext.HTTP.Method != tt.req.Method {
				t.Errorf("Expected method %s, got %v", tt.req.Method, event.RequestContext.HTTP.Method)
			}

			// Check Host header is populated from req.Host
			if tt.req.Host != "" {
				if event.Headers["Host"] != tt.req.Host {
					t.Errorf("Expected Host header %s, got %v", tt.req.Host, event.Headers["Host"])
				}
			}
		})
	}
}

func TestLambdaResponseToHTTP(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    *http.Response
		wantErr bool
	}{
		{
			name: "successful response",
			payload: []byte(`{
				"statusCode": 200,
				"headers": {"Content-Type": "application/json"},
				"body": "{\"message\":\"success\"}",
				"isBase64Encoded": false
			}`),
			want: &http.Response{
				StatusCode: 200,
			},
		},
		{
			name: "error response",
			payload: []byte(`{
				"statusCode": 500,
				"headers": {"Content-Type": "text/plain"},
				"body": "Internal Server Error",
				"isBase64Encoded": false
			}`),
			want: &http.Response{
				StatusCode: 500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := lambdaResponseToHTTP(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Fatalf("lambdaResponseToHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if resp.StatusCode != tt.want.StatusCode {
				t.Errorf("StatusCode = %v, want %v", resp.StatusCode, tt.want.StatusCode)
			}

			// Read and check body
			if resp.Body != nil {
				bodyBytes, _ := io.ReadAll(resp.Body)
				if len(bodyBytes) == 0 && resp.ContentLength > 0 {
					t.Errorf("Body is empty but ContentLength = %d", resp.ContentLength)
				}
			}
		})
	}
}

func TestClientURLSchemeDetection(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantLambda bool
	}{
		{"HTTP URL", "http://example.com/path", false},
		{"HTTPS URL", "https://example.com/path", false},
		{"Lambda URL", "lambda://my-function/path", true},
		{"Lambda URL with query", "lambda://my-function/path?param=value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			isLambda := req.URL.Scheme == "lambda"
			if isLambda != tt.wantLambda {
				t.Errorf("URL %s: expected isLambda=%v, got %v", tt.url, tt.wantLambda, isLambda)
			}
		})
	}
}

func TestHTTPRequestToLambdaEvent_HostHeader(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedHost string
	}{
		{
			name:         "Lambda URL with function name",
			url:          "lambda://binnit-main-src/openapi.json",
			expectedHost: "binnit-main-src",
		},
		{
			name:         "Lambda URL with path",
			url:          "lambda://my-function/get",
			expectedHost: "my-function",
		},
		{
			name:         "Lambda URL with query params",
			url:          "lambda://test-function/path?param=value",
			expectedHost: "test-function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			event, err := httpRequestToLambdaEvent(req)
			if err != nil {
				t.Fatalf("httpRequestToLambdaEvent() error = %v", err)
			}

			// Verify Host header is set from req.Host
			if event.Headers["Host"] != tt.expectedHost {
				t.Errorf("Expected Host header %q, got %q", tt.expectedHost, event.Headers["Host"])
			}

			// Also verify req.Host is populated correctly
			if req.Host != tt.expectedHost {
				t.Errorf("Expected req.Host %q, got %q", tt.expectedHost, req.Host)
			}
		})
	}
}
