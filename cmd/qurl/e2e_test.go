package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestE2EServerURLResolution(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "qurl_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build qurl: %v", err)
	}
	defer os.Remove("qurl_test")

	tests := []struct {
		name        string
		env         map[string]string
		args        []string
		expectURL   string
		expectError bool
	}{
		{
			name: "Binnit API with no servers in spec - conservative fallback",
			env: map[string]string{
				"QURL_OPENAPI": "https://prod.kaixo.io/binnit/main/binnit/openapi.json",
			},
			args:      []string{"-v", "/anything"},
			expectURL: "> GET https://prod.kaixo.io/anything", // Conservative: just host, no path assumptions
		},
		{
			name: "Override with server flag",
			env: map[string]string{
				"QURL_OPENAPI": "https://prod.kaixo.io/binnit/main/binnit/openapi.json",
			},
			args:      []string{"--server", "https://prod.kaixo.io/binnit/main/binnit", "-v", "/get"},
			expectURL: "> GET https://prod.kaixo.io/binnit/main/binnit/get",
		},
		{
			name: "Absolute URL ignores OpenAPI",
			env: map[string]string{
				"QURL_OPENAPI": "https://prod.kaixo.io/binnit/main/binnit/openapi.json",
			},
			args:      []string{"-v", "https://httpbin.dmuth.org/anything"},
			expectURL: "> GET https://httpbin.dmuth.org/anything",
		},
		{
			name:        "Relative path without OpenAPI or server requires explicit configuration",
			args:        []string{"-v", "/anything"},
			expectError: true, // Now properly requires explicit server configuration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./qurl_test", tt.args...)

			// Set environment variables
			cmd.Env = os.Environ()
			for key, value := range tt.env {
				cmd.Env = append(cmd.Env, key+"="+value)
			}

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but command succeeded")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v\nStderr: %s", err, stderr.String())
				return
			}

			// Check if the expected URL appears in stderr (verbose output)
			stderrOutput := stderr.String()
			if !strings.Contains(stderrOutput, tt.expectURL) {
				t.Errorf("Expected stderr to contain %q, got:\n%s", tt.expectURL, stderrOutput)
			}
		})
	}
}

func TestE2EResponseFormat(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "qurl_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build qurl: %v", err)
	}
	defer os.Remove("qurl_test")

	t.Run("Binnit API response format", func(t *testing.T) {
		// Since Binnit API has no servers section, we need to explicitly specify the server
		cmd := exec.Command("./qurl_test", "--server", "https://prod.kaixo.io/binnit/main/binnit", "/anything")
		cmd.Env = append(os.Environ(), "QURL_OPENAPI=https://prod.kaixo.io/binnit/main/binnit/openapi.json")

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		output := stdout.String()

		// Verify it's JSON and contains expected Binnit response fields
		if !strings.Contains(output, `"method"`) {
			t.Errorf("Expected response to contain '\"method\"' field")
		}

		if !strings.Contains(output, `"url"`) {
			t.Errorf("Expected response to contain '\"url\"' field")
		}

		if !strings.Contains(output, `"headers"`) {
			t.Errorf("Expected response to contain '\"headers\"' field")
		}

		// Should NOT contain the old httpbin format
		if strings.Contains(output, `"verb"`) {
			t.Errorf("Response should not contain legacy '\"verb\"' field")
		}
	})
}

func TestE2EMultipleMethodsDocumentation(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "qurl_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build qurl: %v", err)
	}
	defer os.Remove("qurl_test")

	tests := []struct {
		name        string
		args        []string
		expectText  []string
		notExpected []string
	}{
		{
			name: "Single method documentation",
			args: []string{"-X", "POST", "--docs", "/pet"},
			expectText: []string{
				"POST",
				"Add a new pet to the store",
			},
			notExpected: []string{
				"PUT", // Should not show PUT method
			},
		},
		{
			name: "Multiple methods documentation",
			args: []string{"-X", "POST", "-X", "PUT", "--docs", "/pet"},
			expectText: []string{
				"POST",
				"PUT",
				"Add a new pet to the store",
				"Update an existing pet",
			},
			notExpected: []string{
				"GET",    // Should not show other methods
				"DELETE", // Should not show other methods
			},
		},
		{
			name: "Multiple methods with no match",
			args: []string{"-X", "GET", "-X", "DELETE", "--docs", "/pet"},
			expectText: []string{
				"No endpoints found matching the specified path and method",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./qurl_test")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Env = append(os.Environ(), "QURL_OPENAPI=https://petstore3.swagger.io/api/v3/openapi.json")

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				// Check if it's expected to fail
				if len(tt.expectText) > 0 && tt.expectText[0] == "No endpoints found matching the specified path and method" {
					// This is expected for the "no match" test case
				} else {
					t.Fatalf("Command failed: %v, stderr: %s", err, stderr.String())
				}
			}

			output := stdout.String()

			// Check expected text is present
			for _, expected := range tt.expectText {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output: %s", expected, output)
				}
			}

			// Check unwanted text is not present
			for _, notExpected := range tt.notExpected {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output NOT to contain %q, but it did. Output: %s", notExpected, output)
				}
			}
		})
	}
}
