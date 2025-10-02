package config

import (
	"context"
	"os"
	"strings"

	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

// Config holds all application configuration
type Config struct {
	// Core HTTP settings
	OpenAPIURL    string
	Methods       []string // Changed from Method to Methods (slice)
	Path          string
	Headers       []string
	QueryParams   []string
	Data          string
	Server        string
	Verbose       bool
	IncludeHeaders bool
	ShowDocs      bool

	// Authentication
	SigV4Enabled bool
	SigV4Service string

	// MCP settings
	MCP MCPConfig
}

// MCPConfig holds MCP-specific configuration
type MCPConfig struct {
	Enabled        bool
	Description    string    // Server description for LLM context
	AllowedMethods []string // Parsed from Method flag in MCP mode
	PathPrefix     string    // Path constraint from positional argument
	Headers        []string  // Inherited from -H flags
	SigV4          bool      // Inherited from --aws-sigv4
	SigV4Service   string    // Inherited from --aws-service
	ServerURL      string    // Inherited from --server
	OpenAPIURL     string    // Inherited from --openapi
}


// contextKey is a custom type for context keys
type contextKey string

// configKey is the context key for storing config
const configKey contextKey = "config"

// WithConfig adds config to context
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

// FromContext retrieves config from context
func FromContext(ctx context.Context) (*Config, bool) {
	cfg, ok := ctx.Value(configKey).(*Config)
	return cfg, ok
}

// NewConfig creates a Config with default values
func NewConfig() *Config {
	return &Config{
		Methods: []string{"GET"},
	}
}

// LoadFromFlags creates a Config from command line flags
func LoadFromFlags(flags *pflag.FlagSet) (*Config, error) {
	config := NewConfig()

	var err error

	// Core flags
	if config.Methods, err = flags.GetStringSlice("request"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get methods flag")
	}
	// Normalize methods to uppercase
	for i, method := range config.Methods {
		config.Methods[i] = strings.ToUpper(strings.TrimSpace(method))
	}
	// Default to GET if no methods specified
	if len(config.Methods) == 0 {
		config.Methods = []string{"GET"}
	}

	if config.Headers, err = flags.GetStringSlice("header"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get headers flag")
	}

	if config.QueryParams, err = flags.GetStringSlice("query"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get query flag")
	}

	if config.Data, err = flags.GetString("data"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get data flag")
	}

	if config.Server, err = flags.GetString("server"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get server flag")
	}

	if config.Verbose, err = flags.GetBool("verbose"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get verbose flag")
	}

	if config.IncludeHeaders, err = flags.GetBool("include"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get include flag")
	}

	if config.ShowDocs, err = flags.GetBool("docs"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get docs flag")
	}

	// Authentication flags
	if config.SigV4Enabled, err = flags.GetBool("aws-sigv4"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get aws-sigv4 flag")
	}

	if config.SigV4Service, err = flags.GetString("aws-service"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get aws-service flag")
	}

	// OpenAPI URL from flag or environment
	if config.OpenAPIURL, err = flags.GetString("openapi"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get openapi flag")
	}

	// If not set via flag, try environment variables
	if config.OpenAPIURL == "" {
		config.OpenAPIURL = getOpenAPIURL()
	}

	// Get server from environment if not set via flag
	if config.Server == "" {
		if server := os.Getenv("QURL_SERVER"); server != "" {
			config.Server = server
		}
	}

	// Configure debug mode for verbose flag
	if config.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// MCP-specific flags
	if config.MCP.Description, err = flags.GetString("mcp-desc"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get mcp-desc flag")
	}

	// If not set via flag, try environment variable
	if config.MCP.Description == "" {
		if desc := os.Getenv("QURL_MCP_DESCRIPTION"); desc != "" {
			config.MCP.Description = desc
		}
	}

	// Propagate settings to MCP config
	config.MCP.Headers = config.Headers
	config.MCP.SigV4 = config.SigV4Enabled
	config.MCP.SigV4Service = config.SigV4Service
	config.MCP.ServerURL = config.Server
	config.MCP.OpenAPIURL = config.OpenAPIURL

	// Propagate methods to MCP config
	config.MCP.AllowedMethods = config.Methods

	return config, nil
}

// PrimaryMethod returns the first method for HTTP requests
func (c *Config) PrimaryMethod() string {
	if len(c.Methods) > 0 {
		return c.Methods[0]
	}
	return "GET"
}

// LoadMCPFromFlags creates MCPConfig from MCP-specific flags
func LoadMCPFromFlags(flags *pflag.FlagSet) (*MCPConfig, error) {
	config := &MCPConfig{}

	var err error

	if config.AllowedMethods, err = flags.GetStringSlice("allow-methods"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get allow-methods flag")
	}

	if config.Headers, err = flags.GetStringSlice("header"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get headers flag")
	}

	if config.SigV4, err = flags.GetBool("sig-v4"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get sig-v4 flag")
	}

	if config.SigV4Service, err = flags.GetString("sig-v4-service"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get sig-v4-service flag")
	}

	if config.ServerURL, err = flags.GetString("server"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get server flag")
	}

	if config.OpenAPIURL, err = flags.GetString("openapi"); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeConfig, "failed to get openapi flag")
	}

	// If not set via flag, try environment variables
	if config.OpenAPIURL == "" {
		config.OpenAPIURL = getOpenAPIURL()
	}

	return config, nil
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	if c.ShowDocs && c.OpenAPIURL == "" {
		return errors.New(errors.ErrorTypeValidation, "OpenAPI URL is required when using --docs flag").
			WithContext("suggestion", "set QURL_OPENAPI environment variable or use --openapi flag")
	}

	// Validate HTTP method(s)
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range c.Methods {
		methodValid := false
		for _, valid := range validMethods {
			if method == valid {
				methodValid = true
				break
			}
		}
		if !methodValid {
			return errors.New(errors.ErrorTypeValidation, "invalid HTTP method").
				WithContext("method", method).
				WithContext("valid_methods", validMethods)
		}
	}

	return nil
}

// ValidateMCP ensures MCP configuration is valid
func (c *MCPConfig) Validate() error {
	if c.OpenAPIURL == "" {
		return errors.New(errors.ErrorTypeConfig, "OpenAPI URL is required for MCP server").
			WithContext("config_type", "mcp").
			WithContext("suggestion", "set QURL_OPENAPI environment variable or use --openapi flag")
	}

	// Validate allowed methods
	if len(c.AllowedMethods) == 0 {
		c.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH"} // Default
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range c.AllowedMethods {
		methodValid := false
		upperMethod := strings.ToUpper(method)
		for _, valid := range validMethods {
			if upperMethod == valid {
				methodValid = true
				break
			}
		}
		if !methodValid {
			return errors.New(errors.ErrorTypeValidation, "invalid HTTP method in allow-methods").
				WithContext("method", method).
				WithContext("valid_methods", validMethods)
		}
	}

	return nil
}

// getOpenAPIURL retrieves OpenAPI URL from environment variables
func getOpenAPIURL() string {
	if url := os.Getenv("QURL_OPENAPI"); url != "" {
		return url
	}
	return ""
}