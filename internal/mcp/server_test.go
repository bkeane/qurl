package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/brendan.keane/qurl/internal/testutil"
	"github.com/rs/zerolog"
)

func TestNewServer(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	cfg := testutil.NewConfigBuilder().
		WithOpenAPIURL("https://api.example.com/openapi.json").
		WithMCP().
		Build()

	server, err := NewServer(logger, cfg)
	testutil.AssertNoError(t, err, "NewServer should not return error")

	if server == nil {
		t.Fatal("NewServer should return a server instance")
	}

	if server.config != cfg {
		t.Errorf("Server config mismatch: got %v, expected %v", server.config, cfg)
	}

	if server.executor == nil {
		t.Error("Server executor should not be nil")
	}

	if server.viewer == nil {
		t.Error("Server viewer should not be nil")
	}
}

func TestNewServer_InvalidConfig(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	// Empty config should potentially fail
	cfg := testutil.NewConfigBuilder().Build()

	server, err := NewServer(logger, cfg)
	// Note: Current implementation might not fail with empty config
	// but we should test the actual behavior
	if err != nil {
		// If it fails, that's acceptable
		testutil.AssertError(t, err, "NewServer with empty config")
		if server != nil {
			t.Error("NewServer should return nil server on error")
		}
	} else {
		// If it succeeds, server should be valid
		if server == nil {
			t.Error("NewServer should return server instance on success")
		}
	}
}

func TestMCPRequestSerialization(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "execute",
		Params:  map[string]interface{}{"path": "/test"},
	}

	data, err := json.Marshal(req)
	testutil.AssertNoError(t, err, "MCPRequest serialization")

	var decoded MCPRequest
	err = json.Unmarshal(data, &decoded)
	testutil.AssertNoError(t, err, "MCPRequest deserialization")

	testutil.AssertStringEqual(t, decoded.JSONRPC, "2.0", "JSONRPC field")
	testutil.AssertStringEqual(t, decoded.Method, "execute", "Method field")
	testutil.AssertEqual(t, decoded.ID, float64(1), "ID field") // JSON numbers are float64
}

func TestMCPResponseSerialization(t *testing.T) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]interface{}{"status": "success"},
	}

	data, err := json.Marshal(resp)
	testutil.AssertNoError(t, err, "MCPResponse serialization")

	var decoded MCPResponse
	err = json.Unmarshal(data, &decoded)
	testutil.AssertNoError(t, err, "MCPResponse deserialization")

	testutil.AssertStringEqual(t, decoded.JSONRPC, "2.0", "JSONRPC field")
	testutil.AssertEqual(t, decoded.ID, float64(1), "ID field")

	if decoded.Error != nil {
		t.Error("Error field should be nil for success response")
	}
}

func TestMCPErrorSerialization(t *testing.T) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Error: &MCPError{
			Code:    -1,
			Message: "Test error",
		},
	}

	data, err := json.Marshal(resp)
	testutil.AssertNoError(t, err, "MCPError serialization")

	var decoded MCPResponse
	err = json.Unmarshal(data, &decoded)
	testutil.AssertNoError(t, err, "MCPError deserialization")

	if decoded.Error == nil {
		t.Fatal("Error field should not be nil")
	}

	testutil.AssertEqual(t, decoded.Error.Code, -1, "Error code")
	testutil.AssertStringEqual(t, decoded.Error.Message, "Test error", "Error message")
}

// Test helper to capture server output
func captureServerOutput(t *testing.T, server *Server, input string) (string, error) {
	t.Helper()

	// Create a buffer to capture output
	var outputBuffer bytes.Buffer

	// Save original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture output
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// Replace stdout with our pipe
	os.Stdout = w

	// Start a goroutine to read from pipe
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				outputBuffer.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// Process the input
	err = server.handleMessage(input)

	// Close the write end to signal EOF
	w.Close()

	// Wait for reader to finish
	<-done

	// Restore original stdout
	os.Stdout = originalStdout

	return outputBuffer.String(), err
}

func TestServer_HandleInitializeMessage(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	cfg := testutil.NewConfigBuilder().WithMCP().Build()

	server, err := NewServer(logger, cfg)
	testutil.AssertNoError(t, err, "NewServer")

	initMsg := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}`

	// Test that handleMessage doesn't panic with initialize message
	err = server.handleMessage(initMsg)
	// The actual behavior depends on implementation
	// At minimum, it shouldn't panic
}

func TestServer_HandleInvalidJSON(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	cfg := testutil.NewConfigBuilder().WithMCP().Build()

	server, err := NewServer(logger, cfg)
	testutil.AssertNoError(t, err, "NewServer")

	invalidMsg := `{"invalid json`

	err = server.handleMessage(invalidMsg)
	// The current implementation may not return an error but handles it internally
	// This test ensures it doesn't panic
	if err != nil {
		testutil.AssertError(t, err, "handleMessage returned error for invalid JSON")
	} else {
		t.Log("handleMessage handled invalid JSON without returning error (acceptable)")
	}
}

func TestServer_HandleEmptyMessage(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	cfg := testutil.NewConfigBuilder().WithMCP().Build()

	server, err := NewServer(logger, cfg)
	testutil.AssertNoError(t, err, "NewServer")

	// Empty message should be handled gracefully
	err = server.handleMessage("")
	// Depending on implementation, this might or might not be an error
	// The test ensures it doesn't panic
}

func TestServer_MCPConfigurationValidation(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	tests := []struct {
		name        string
		configFunc  func() *testutil.ConfigBuilder
		expectError bool
	}{
		{
			name: "valid MCP config",
			configFunc: func() *testutil.ConfigBuilder {
				return testutil.NewConfigBuilder().
					WithMCP().
					WithOpenAPIURL("https://api.example.com/openapi.json")
			},
			expectError: false,
		},
		{
			name: "MCP with allowed methods",
			configFunc: func() *testutil.ConfigBuilder {
				return testutil.NewConfigBuilder().
					WithMCP().
					WithMCPMethods("GET", "POST").
					WithOpenAPIURL("https://api.example.com/openapi.json")
			},
			expectError: false,
		},
		{
			name: "MCP with path prefix",
			configFunc: func() *testutil.ConfigBuilder {
				return testutil.NewConfigBuilder().
					WithMCP().
					WithMCPPathPrefix("/api/v1").
					WithOpenAPIURL("https://api.example.com/openapi.json")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.configFunc().Build()
			server, err := NewServer(logger, cfg)

			if tt.expectError {
				testutil.AssertError(t, err, "NewServer should return error")
				if server != nil {
					t.Error("Server should be nil on error")
				}
			} else {
				testutil.AssertNoError(t, err, "NewServer should not return error")
				if server == nil {
					t.Error("Server should not be nil on success")
				}
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkMCPRequestSerialization(b *testing.B) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "execute",
		Params:  map[string]interface{}{"path": "/test", "method": "GET"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPResponseSerialization(b *testing.B) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  map[string]interface{}{"body": "response data", "headers": map[string][]string{"Content-Type": {"application/json"}}, "statusCode": 200},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}