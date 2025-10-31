package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

// TestClientCreationWithoutAWSConfig verifies that creating a client and making
// regular HTTP/HTTPS requests does NOT require AWS credentials to be configured.
// This is a regression test for the issue where AWS config was loaded unconditionally,
// causing failures for users with MFA-enabled AWS profiles when making non-AWS requests.
func TestClientCreationWithoutAWSConfig(t *testing.T) {
	// Set up environment to simulate missing/broken AWS credentials
	// This ensures AWS config loading would fail if attempted
	originalHome := os.Getenv("HOME")
	originalAWSProfile := os.Getenv("AWS_PROFILE")
	originalAWSRegion := os.Getenv("AWS_REGION")
	originalAWSAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	originalAWSSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	defer func() {
		// Restore original environment
		os.Setenv("HOME", originalHome)
		os.Setenv("AWS_PROFILE", originalAWSProfile)
		os.Setenv("AWS_REGION", originalAWSRegion)
		os.Setenv("AWS_ACCESS_KEY_ID", originalAWSAccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", originalAWSSecretKey)
	}()

	// Clear AWS environment variables to ensure no valid AWS config exists
	os.Setenv("HOME", t.TempDir())
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_SESSION_TOKEN")

	// Test 1: Client creation should succeed without AWS config
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed without AWS config: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}

	// Test 2: Regular HTTPS requests should work without AWS config
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() failed for regular HTTPS request without AWS config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "success") {
		t.Errorf("Expected response to contain 'success', got: %s", string(body))
	}
}

// TestClientLazyAWSInitialization verifies that AWS config is only loaded
// when actually needed (i.e., when making a lambda:// request), not at client creation time.
func TestClientLazyAWSInitialization(t *testing.T) {
	// Clear AWS environment to ensure config loading would fail
	originalAWSRegion := os.Getenv("AWS_REGION")
	defer os.Setenv("AWS_REGION", originalAWSRegion)
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")

	// Client creation should succeed even without AWS config
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() should succeed without AWS config: %v", err)
	}

	// Verify that Lambda client is not initialized yet
	if client.lambdaClient != nil {
		t.Error("Lambda client should not be initialized at client creation time")
	}

	if client.awsConfig != nil {
		t.Error("AWS config should not be loaded at client creation time")
	}

	// Note: We can't test actual lambda:// invocation here without valid AWS credentials,
	// but we've verified that initialization is deferred and regular HTTP works
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
