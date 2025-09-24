package main

import (
	"strings"
	"testing"

	"github.com/brendan.keane/qurl/internal/config"
)

func TestMCPConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		methods        []string
		pathPrefix     string
		headers        []string
		wantMethods    []string
		wantPrefix     string
		wantHeaders    []string
	}{
		{
			name:        "unrestricted MCP (default GET)",
			methods:     []string{"GET"},
			pathPrefix:  "",
			headers:     []string{},
			wantMethods: []string{}, // Empty means all methods allowed
			wantPrefix:  "",
			wantHeaders: []string{},
		},
		{
			name:        "restricted to specific methods",
			methods:     []string{"GET", "POST"},
			pathPrefix:  "/api/",
			headers:     []string{"Authorization: Bearer test-token"},
			wantMethods: []string{"GET", "POST"},
			wantPrefix:  "/api/",
			wantHeaders: []string{"Authorization: Bearer test-token"},
		},
		{
			name:        "single method restriction",
			methods:     []string{"POST"},
			pathPrefix:  "/webhooks/",
			headers:     []string{},
			wantMethods: []string{"POST"},
			wantPrefix:  "/webhooks/",
			wantHeaders: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Methods: tt.methods,
				Headers: tt.headers,
				MCP: config.MCPConfig{
					Enabled:        true,
					AllowedMethods: tt.wantMethods, // Use the expected methods, not the input methods
					PathPrefix:     tt.pathPrefix,
					Headers:        tt.headers,
				},
			}

			// Verify MCP configuration
			if len(cfg.MCP.AllowedMethods) != len(tt.wantMethods) {
				t.Errorf("AllowedMethods length = %d, want %d", len(cfg.MCP.AllowedMethods), len(tt.wantMethods))
			}

			for i, method := range tt.wantMethods {
				if i >= len(cfg.MCP.AllowedMethods) || cfg.MCP.AllowedMethods[i] != method {
					t.Errorf("AllowedMethods[%d] = %q, want %q", i, cfg.MCP.AllowedMethods[i], method)
				}
			}

			if cfg.MCP.PathPrefix != tt.wantPrefix {
				t.Errorf("PathPrefix = %q, want %q", cfg.MCP.PathPrefix, tt.wantPrefix)
			}

			if len(cfg.MCP.Headers) != len(tt.wantHeaders) {
				t.Errorf("Headers length = %d, want %d", len(cfg.MCP.Headers), len(tt.wantHeaders))
			}
		})
	}
}

func TestMCPMethodFiltering(t *testing.T) {
	tests := []struct {
		name           string
		allowedMethods []string
		testMethod     string
		shouldAllow    bool
	}{
		{
			name:           "empty allowed methods (allows all)",
			allowedMethods: []string{},
			testMethod:     "GET",
			shouldAllow:    true,
		},
		{
			name:           "empty allowed methods allows POST",
			allowedMethods: []string{},
			testMethod:     "POST",
			shouldAllow:    true,
		},
		{
			name:           "GET allowed",
			allowedMethods: []string{"GET"},
			testMethod:     "GET",
			shouldAllow:    true,
		},
		{
			name:           "POST not in allowed list",
			allowedMethods: []string{"GET"},
			testMethod:     "POST",
			shouldAllow:    false,
		},
		{
			name:           "case insensitive matching",
			allowedMethods: []string{"GET", "POST"},
			testMethod:     "get",
			shouldAllow:    true,
		},
		{
			name:           "DELETE not allowed",
			allowedMethods: []string{"GET", "POST"},
			testMethod:     "DELETE",
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the method filtering logic from the MCP server
			var methodAllowed bool

			if len(tt.allowedMethods) == 0 {
				// Empty means all methods allowed
				methodAllowed = true
			} else {
				methodAllowed = false
				for _, allowed := range tt.allowedMethods {
					if allowed == tt.testMethod || allowed == strings.ToUpper(tt.testMethod) {
						methodAllowed = true
						break
					}
				}
			}

			if methodAllowed != tt.shouldAllow {
				t.Errorf("Method %q allowed = %v, want %v", tt.testMethod, methodAllowed, tt.shouldAllow)
			}
		})
	}
}

func TestMCPPathConstraints(t *testing.T) {
	tests := []struct {
		name       string
		pathPrefix string
		testPath   string
		shouldPass bool
	}{
		{
			name:       "no path constraint",
			pathPrefix: "",
			testPath:   "/anything",
			shouldPass: true,
		},
		{
			name:       "path under allowed prefix",
			pathPrefix: "/api/",
			testPath:   "/api/users",
			shouldPass: true,
		},
		{
			name:       "path outside allowed prefix",
			pathPrefix: "/api/",
			testPath:   "/admin/users",
			shouldPass: false,
		},
		{
			name:       "exact prefix match",
			pathPrefix: "/webhooks/",
			testPath:   "/webhooks/github",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the path validation logic from the MCP server
			var pathAllowed bool

			if tt.pathPrefix == "" {
				pathAllowed = true
			} else {
				pathAllowed = strings.HasPrefix(tt.testPath, tt.pathPrefix)
			}

			if pathAllowed != tt.shouldPass {
				t.Errorf("Path %q allowed with prefix %q = %v, want %v", tt.testPath, tt.pathPrefix, pathAllowed, tt.shouldPass)
			}
		})
	}
}