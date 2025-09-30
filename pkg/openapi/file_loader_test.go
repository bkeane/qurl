package openapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadFromFileURL tests loading OpenAPI specs from file:// URLs
func TestLoadFromFileURL(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "qurl_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test OpenAPI spec content
	validSpec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"servers": [
			{
				"url": "https://api.example.com",
				"description": "Production server"
			},
			{
				"url": "http://localhost:8080",
				"description": "Local development"
			}
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

	t.Run("Load valid OpenAPI spec from file:// URL", func(t *testing.T) {
		// Write test spec to file
		specFile := filepath.Join(tempDir, "test-spec.json")
		err := os.WriteFile(specFile, []byte(validSpec), 0644)
		require.NoError(t, err)

		// Create parser and load from file:// URL
		parser := NewParser()
		fileURL := "file://" + specFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		// Verify the spec was loaded correctly
		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "Test API", info.Title)
		assert.Equal(t, "1.0.0", info.Version)

		// Verify servers were loaded
		servers, err := parser.GetServers()
		require.NoError(t, err)
		require.Len(t, servers, 2)
		assert.Equal(t, "https://api.example.com", servers[0].URL)
		assert.Equal(t, "Production server", servers[0].Description)
		assert.Equal(t, "http://localhost:8080", servers[1].URL)
		assert.Equal(t, "Local development", servers[1].Description)

		// Verify paths were loaded
		paths, err := parser.GetPaths("*", "*")
		require.NoError(t, err)
		require.Len(t, paths, 1)
		assert.Equal(t, "/test", paths[0].Path)
		assert.Equal(t, "GET", paths[0].Method)
		assert.Equal(t, "Test endpoint", paths[0].Summary)
	})

	t.Run("Load OpenAPI spec with absolute file path", func(t *testing.T) {
		// Write test spec to file
		specFile := filepath.Join(tempDir, "absolute-spec.json")
		err := os.WriteFile(specFile, []byte(validSpec), 0644)
		require.NoError(t, err)

		// Create parser and load with absolute path in file:// URL
		parser := NewParser()
		// Use absolute path
		absPath, err := filepath.Abs(specFile)
		require.NoError(t, err)

		fileURL := "file://" + absPath

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		// Verify the spec was loaded correctly
		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "Test API", info.Title)
	})

	t.Run("Error on non-existent file", func(t *testing.T) {
		parser := NewParser()
		nonExistentFile := "file://" + filepath.Join(tempDir, "does-not-exist.json")

		ctx := context.Background()
		err := parser.LoadFromURL(ctx, nonExistentFile)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reading file")
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("Error on invalid JSON", func(t *testing.T) {
		// Write invalid JSON to file
		invalidFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidFile, []byte(`{invalid json`), 0644)
		require.NoError(t, err)

		parser := NewParser()
		fileURL := "file://" + invalidFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing OpenAPI document")
	})

	t.Run("Error on permission denied", func(t *testing.T) {
		// Write test spec to file with no read permissions
		noPermFile := filepath.Join(tempDir, "no-perm.json")
		err := os.WriteFile(noPermFile, []byte(validSpec), 0000) // No permissions
		require.NoError(t, err)
		defer os.Chmod(noPermFile, 0644) // Restore permissions for cleanup

		parser := NewParser()
		fileURL := "file://" + noPermFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reading file")
		assert.Contains(t, err.Error(), "permission denied")
	})
}

// TestLoadFromFileURLWithYAML tests loading YAML OpenAPI specs
func TestLoadFromFileURLWithYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "qurl_yaml_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	yamlSpec := `openapi: 3.0.0
info:
  title: YAML Test API
  version: 2.0.0
  description: Testing YAML OpenAPI specs
servers:
  - url: https://yaml.example.com
    description: YAML server
paths:
  /yaml-test:
    get:
      summary: YAML test endpoint
      responses:
        '200':
          description: Success
`

	t.Run("Load YAML OpenAPI spec from file:// URL", func(t *testing.T) {
		// Write YAML spec to file
		specFile := filepath.Join(tempDir, "test-spec.yaml")
		err := os.WriteFile(specFile, []byte(yamlSpec), 0644)
		require.NoError(t, err)

		// Create parser and load from file:// URL
		parser := NewParser()
		fileURL := "file://" + specFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		// Verify the YAML spec was loaded correctly
		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "YAML Test API", info.Title)
		assert.Equal(t, "2.0.0", info.Version)
		assert.Equal(t, "Testing YAML OpenAPI specs", info.Description)

		// Verify servers were loaded from YAML
		servers, err := parser.GetServers()
		require.NoError(t, err)
		require.Len(t, servers, 1)
		assert.Equal(t, "https://yaml.example.com", servers[0].URL)
		assert.Equal(t, "YAML server", servers[0].Description)

		// Verify paths were loaded from YAML
		paths, err := parser.GetPaths("*", "*")
		require.NoError(t, err)
		require.Len(t, paths, 1)
		assert.Equal(t, "/yaml-test", paths[0].Path)
		assert.Equal(t, "GET", paths[0].Method)
		assert.Equal(t, "YAML test endpoint", paths[0].Summary)
	})
}

// TestLoadFromURLSchemeDetection tests that the correct loading method is used based on URL scheme
func TestLoadFromURLSchemeDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "qurl_scheme_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	validSpec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Scheme Test API",
			"version": "1.0.0"
		},
		"paths": {}
	}`

	t.Run("file:// scheme uses file loading", func(t *testing.T) {
		specFile := filepath.Join(tempDir, "scheme-test.json")
		err := os.WriteFile(specFile, []byte(validSpec), 0644)
		require.NoError(t, err)

		parser := NewParser()
		fileURL := "file://" + specFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "Scheme Test API", info.Title)
	})

	t.Run("http:// scheme attempts HTTP loading", func(t *testing.T) {
		parser := NewParser()
		httpURL := "http://example.com/nonexistent.json"

		ctx := context.Background()
		err := parser.LoadFromURL(ctx, httpURL)

		// Should fail with HTTP-related error, not file error
		assert.Error(t, err)
		// Could be "unexpected status code" or "fetching OpenAPI spec"
		assert.True(t, strings.Contains(err.Error(), "unexpected status code") || strings.Contains(err.Error(), "fetching OpenAPI spec"))
		assert.NotContains(t, err.Error(), "reading file")
	})

	t.Run("https:// scheme attempts HTTPS loading", func(t *testing.T) {
		parser := NewParser()
		httpsURL := "https://example.com/nonexistent.json"

		ctx := context.Background()
		err := parser.LoadFromURL(ctx, httpsURL)

		// Should fail with HTTP-related error, not file error
		assert.Error(t, err)
		// Could be "unexpected status code" or "fetching OpenAPI spec"
		assert.True(t, strings.Contains(err.Error(), "unexpected status code") || strings.Contains(err.Error(), "fetching OpenAPI spec"))
		assert.NotContains(t, err.Error(), "reading file")
	})

	t.Run("invalid URL scheme", func(t *testing.T) {
		parser := NewParser()
		invalidURL := "ftp://example.com/spec.json"

		ctx := context.Background()
		err := parser.LoadFromURL(ctx, invalidURL)

		// Should fail when HTTP client tries to handle unsupported scheme
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fetching OpenAPI spec")
	})
}

// TestLoadFromFileURLEdgeCases tests edge cases and special scenarios
func TestLoadFromFileURLEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "qurl_edge_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("file:// URL with spaces in path", func(t *testing.T) {
		// Create directory and file with spaces
		spacedDir := filepath.Join(tempDir, "dir with spaces")
		err := os.MkdirAll(spacedDir, 0755)
		require.NoError(t, err)

		specFile := filepath.Join(spacedDir, "spec with spaces.json")
		validSpec := `{"openapi": "3.0.0", "info": {"title": "Spaced API", "version": "1.0.0"}, "paths": {}}`
		err = os.WriteFile(specFile, []byte(validSpec), 0644)
		require.NoError(t, err)

		parser := NewParser()
		fileURL := "file://" + specFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "Spaced API", info.Title)
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(emptyFile, []byte(""), 0644)
		require.NoError(t, err)

		parser := NewParser()
		fileURL := "file://" + emptyFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing OpenAPI document")
	})

	t.Run("large file", func(t *testing.T) {
		// Create a large but valid OpenAPI spec
		largeSpec := `{
			"openapi": "3.0.0",
			"info": {
				"title": "Large API",
				"version": "1.0.0",
				"description": "` + strings.Repeat("A very long description. ", 1000) + `"
			},
			"paths": {`

		// Add many paths
		for i := 0; i < 100; i++ {
			if i > 0 {
				largeSpec += ","
			}
			largeSpec += `"/path` + fmt.Sprintf("%d", i) + `": {
				"get": {
					"summary": "Path ` + fmt.Sprintf("%d", i) + `",
					"responses": {"200": {"description": "Success"}}
				}
			}`
		}
		largeSpec += "}}"

		largeFile := filepath.Join(tempDir, "large.json")
		err := os.WriteFile(largeFile, []byte(largeSpec), 0644)
		require.NoError(t, err)

		parser := NewParser()
		fileURL := "file://" + largeFile

		ctx := context.Background()
		err = parser.LoadFromURL(ctx, fileURL)
		require.NoError(t, err)

		info, err := parser.GetInfo()
		require.NoError(t, err)
		assert.Equal(t, "Large API", info.Title)

		// Verify all paths were loaded
		paths, err := parser.GetPaths("*", "*")
		require.NoError(t, err)
		assert.Len(t, paths, 100)
	})
}

// BenchmarkLoadFromFileURL benchmarks file:// URL loading performance
func BenchmarkLoadFromFileURL(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "qurl_bench_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	validSpec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Benchmark API",
			"version": "1.0.0"
		},
		"servers": [
			{"url": "https://api.example.com"}
		],
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`

	specFile := filepath.Join(tempDir, "benchmark.json")
	err = os.WriteFile(specFile, []byte(validSpec), 0644)
	require.NoError(b, err)

	fileURL := "file://" + specFile

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewParser()
		ctx := context.Background()
		err := parser.LoadFromURL(ctx, fileURL)
		if err != nil {
			b.Fatalf("Failed to load spec: %v", err)
		}
	}
}