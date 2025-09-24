package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig should return non-nil config")
	}

	// Test default values
	if cfg.Logger.Level != "warn" {
		t.Errorf("default log level: got %q, expected %q", cfg.Logger.Level, "warn")
	}
	if cfg.Logger.Pretty != true {
		t.Errorf("default pretty logging: got %v, expected %v", cfg.Logger.Pretty, true)
	}
	if cfg.Logger.WithCaller != false {
		t.Errorf("default caller logging: got %v, expected %v", cfg.Logger.WithCaller, false)
	}
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