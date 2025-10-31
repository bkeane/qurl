package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig should return non-nil config")
	}

	// Test default values
	if cfg.Verbose != false {
		t.Errorf("default verbose: got %v, expected %v", cfg.Verbose, false)
	}
	if cfg.IncludeHeaders != false {
		t.Errorf("default include headers: got %v, expected %v", cfg.IncludeHeaders, false)
	}
	if cfg.ShowDocs != false {
		t.Errorf("default show docs: got %v, expected %v", cfg.ShowDocs, false)
	}
	if cfg.SigV4Enabled != false {
		t.Errorf("default SigV4: got %v, expected %v", cfg.SigV4Enabled, false)
	}
	if cfg.MCP.Enabled != false {
		t.Errorf("default MCP enabled: got %v, expected %v", cfg.MCP.Enabled, false)
	}
	if len(cfg.Methods) != 1 || cfg.Methods[0] != "GET" {
		t.Errorf("default methods: got %v, expected %v", cfg.Methods, []string{"GET"})
	}
}


func TestConfig_Validation_Valid(t *testing.T) {
	cfg := NewConfig()
	cfg.Path = "/users"
	cfg.Methods = []string{"GET"}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Config validation should pass: %v", err)
	}
}

func TestConfig_Validation_ValidWithoutPath(t *testing.T) {
	cfg := NewConfig()
	cfg.Methods = []string{"GET"}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Config validation should pass without path requirement: %v", err)
	}
}

func TestConfig_Validation_EmptyMethods(t *testing.T) {
	cfg := NewConfig()
	cfg.Path = "/users"
	cfg.Methods = []string{}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Config validation should pass with empty methods (defaults to GET): %v", err)
	}
}

func TestConfig_PrimaryMethod(t *testing.T) {
	tests := []struct {
		name     string
		methods  []string
		expected string
	}{
		{
			name:     "single GET method",
			methods:  []string{"GET"},
			expected: "GET",
		},
		{
			name:     "single POST method",
			methods:  []string{"POST"},
			expected: "POST",
		},
		{
			name:     "multiple methods",
			methods:  []string{"GET", "POST", "PUT"},
			expected: "GET",
		},
		{
			name:     "empty methods returns GET default",
			methods:  []string{},
			expected: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Methods = tt.methods

			result := cfg.PrimaryMethod()
			if result != tt.expected {
				t.Errorf("PrimaryMethod(): got %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMCPConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *MCPConfig
		expectErr bool
	}{
		{
			name: "valid MCP config with OpenAPI URL",
			setupFunc: func() *MCPConfig {
				return &MCPConfig{
					Enabled:        true,
					AllowedMethods: []string{"GET", "POST"},
					OpenAPIURL:     "https://api.example.com/openapi.json",
				}
			},
			expectErr: false,
		},
		{
			name: "MCP enabled without OpenAPI URL",
			setupFunc: func() *MCPConfig {
				return &MCPConfig{
					Enabled:        true,
					AllowedMethods: []string{"GET", "POST"},
				}
			},
			expectErr: true, // OpenAPI URL is required for MCP
		},
		{
			name: "MCP enabled with no methods but with OpenAPI URL",
			setupFunc: func() *MCPConfig {
				return &MCPConfig{
					Enabled:        true,
					AllowedMethods: []string{},
					OpenAPIURL:     "https://api.example.com/openapi.json",
				}
			},
			expectErr: false, // Should default to allowed methods
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcpCfg := tt.setupFunc()
			err := mcpCfg.Validate()

			if tt.expectErr {
				if err == nil {
					t.Error("MCPConfig validation should fail")
				}
			} else {
				if err != nil {
					t.Errorf("MCPConfig validation should pass: %v", err)
				}
			}
		})
	}
}

func TestLoadFromFlags_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		flagValues     map[string]string
		expectedConfig func(*Config)
	}{
		{
			name: "QURL_OPENAPI environment variable",
			envVars: map[string]string{
				"QURL_OPENAPI": "https://api.example.com/openapi.json",
			},
			expectedConfig: func(c *Config) {
				if c.OpenAPIURL != "https://api.example.com/openapi.json" {
					t.Errorf("OpenAPIURL: got %q, expected %q", c.OpenAPIURL, "https://api.example.com/openapi.json")
				}
			},
		},
		{
			name: "QURL_SERVER environment variable",
			envVars: map[string]string{
				"QURL_SERVER": "https://api.example.com",
			},
			expectedConfig: func(c *Config) {
				if c.Server != "https://api.example.com" {
					t.Errorf("Server: got %q, expected %q", c.Server, "https://api.example.com")
				}
			},
		},
		{
			name: "QURL_LOG_LEVEL environment variable",
			envVars: map[string]string{
				"QURL_LOG_LEVEL": "debug",
			},
			expectedConfig: func(c *Config) {
				// No specific checks for this test
			},
		},
		{
			name: "flag overrides QURL_OPENAPI",
			envVars: map[string]string{
				"QURL_OPENAPI": "https://env.example.com/openapi.json",
			},
			flagValues: map[string]string{
				"openapi": "https://flag.example.com/openapi.json",
			},
			expectedConfig: func(c *Config) {
				if c.OpenAPIURL != "https://flag.example.com/openapi.json" {
					t.Errorf("OpenAPIURL: got %q, expected flag value %q", c.OpenAPIURL, "https://flag.example.com/openapi.json")
				}
			},
		},
		{
			name: "flag overrides QURL_SERVER",
			envVars: map[string]string{
				"QURL_SERVER": "https://env.example.com",
			},
			flagValues: map[string]string{
				"server": "https://flag.example.com",
			},
			expectedConfig: func(c *Config) {
				if c.Server != "https://flag.example.com" {
					t.Errorf("Server: got %q, expected flag value %q", c.Server, "https://flag.example.com")
				}
			},
		},
		{
			name: "all environment variables together",
			envVars: map[string]string{
				"QURL_OPENAPI": "https://api.example.com/openapi.json",
				"QURL_SERVER": "https://api.example.com",
				"QURL_LOG_LEVEL": "info",
			},
			expectedConfig: func(c *Config) {
				if c.OpenAPIURL != "https://api.example.com/openapi.json" {
					t.Errorf("OpenAPIURL: got %q, expected %q", c.OpenAPIURL, "https://api.example.com/openapi.json")
				}
				if c.Server != "https://api.example.com" {
					t.Errorf("Server: got %q, expected %q", c.Server, "https://api.example.com")
				}
			},
		},
		{
			name: "QURL_LOG_FORMAT json",
			envVars: map[string]string{
				"QURL_LOG_FORMAT": "json",
			},
			expectedConfig: func(c *Config) {
				// No specific checks for this test
			},
		},
		{
			name: "QURL_LOG_FORMAT pretty",
			envVars: map[string]string{
				"QURL_LOG_FORMAT": "pretty",
			},
			expectedConfig: func(c *Config) {
				// No specific checks for this test
			},
		},
		{
			name: "QURL_LOG_FORMAT invalid value ignored",
			envVars: map[string]string{
				"QURL_LOG_FORMAT": "invalid",
			},
			expectedConfig: func(c *Config) {
				// No specific checks for this test
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("QURL_OPENAPI")
			os.Unsetenv("OPENAPI_URL")
			os.Unsetenv("QURL_SERVER")
			os.Unsetenv("QURL_LOG_LEVEL")
			os.Unsetenv("QURL_LOG_FORMAT")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create flag set with defaults
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			var cfg Config

			// Define flags as they are in main.go
			flags.StringVar(&cfg.OpenAPIURL, "openapi", "", "OpenAPI specification URL")
			flags.StringVar(&cfg.Server, "server", "", "Server URL or index")
			flags.StringSliceVar(&cfg.Methods, "request", []string{"GET"}, "HTTP method")
			flags.StringSliceVar(&cfg.Headers, "header", nil, "Custom headers")
			flags.StringSliceVar(&cfg.QueryParams, "query", nil, "Query parameters")
			flags.StringVar(&cfg.Data, "data", "", "Request body data")
			flags.BoolVar(&cfg.Verbose, "verbose", false, "Verbose output")
			flags.BoolVar(&cfg.IncludeHeaders, "include", false, "Include headers")
			flags.BoolVar(&cfg.ShowDocs, "docs", false, "Show docs")
			flags.BoolVar(&cfg.SigV4Enabled, "aws-sigv4", false, "Sign with SigV4")
			flags.StringVar(&cfg.SigV4Service, "aws-service", "execute-api", "AWS service")
			flags.StringVar(&cfg.MCP.Description, "mcp-desc", "", "MCP server description")
			// log-pretty flag removed - now controlled by QURL_LOG_FORMAT env var

			// Set flag values from test
			for flag, value := range tt.flagValues {
				flags.Set(flag, value)
			}

			// Load config
			config, err := LoadFromFlags(flags)
			if err != nil {
				t.Fatalf("LoadFromFlags failed: %v", err)
			}

			// Check expectations
			tt.expectedConfig(config)

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewConfig()
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	cfg := NewConfig()
	cfg.Path = "/test"
	cfg.Methods = []string{"GET"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}