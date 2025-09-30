package testutil

import (
	"github.com/brendan.keane/qurl/internal/config"
)

// ConfigBuilder provides a fluent interface for building test configurations
// This eliminates repetitive config setup across test files
type ConfigBuilder struct {
	config *config.Config
}

// NewConfigBuilder creates a new config builder with sensible defaults
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &config.Config{
			Methods:        []string{"GET"},
			Headers:        []string{},
			QueryParams:    []string{},
			Data:           "",
			Server:         "",
			Verbose:        false,
			IncludeHeaders: false,
			ShowDocs:       false,
			SigV4Enabled:   false,
			SigV4Service:   "",
			OpenAPIURL:     "",
			MCP: config.MCPConfig{
				Enabled:        false,
				AllowedMethods: []string{},
				PathPrefix:     "",
				Headers:        []string{},
				SigV4:          false,
				SigV4Service:   "",
				ServerURL:      "",
				OpenAPIURL:     "",
			},
		},
	}
}

// WithMethods sets the HTTP methods
func (b *ConfigBuilder) WithMethods(methods ...string) *ConfigBuilder {
	b.config.Methods = methods
	return b
}

// WithMethod sets a single HTTP method (convenience)
func (b *ConfigBuilder) WithMethod(method string) *ConfigBuilder {
	b.config.Methods = []string{method}
	return b
}

// WithServer sets the server URL
func (b *ConfigBuilder) WithServer(server string) *ConfigBuilder {
	b.config.Server = server
	return b
}

// WithOpenAPIURL sets the OpenAPI URL
func (b *ConfigBuilder) WithOpenAPIURL(url string) *ConfigBuilder {
	b.config.OpenAPIURL = url
	return b
}

// WithHeaders adds HTTP headers
func (b *ConfigBuilder) WithHeaders(headers ...string) *ConfigBuilder {
	b.config.Headers = append(b.config.Headers, headers...)
	return b
}

// WithHeader adds a single HTTP header (convenience)
func (b *ConfigBuilder) WithHeader(header string) *ConfigBuilder {
	b.config.Headers = append(b.config.Headers, header)
	return b
}

// WithQueryParams adds query parameters
func (b *ConfigBuilder) WithQueryParams(params ...string) *ConfigBuilder {
	b.config.QueryParams = append(b.config.QueryParams, params...)
	return b
}

// WithQueryParam adds a single query parameter (convenience)
func (b *ConfigBuilder) WithQueryParam(param string) *ConfigBuilder {
	b.config.QueryParams = append(b.config.QueryParams, param)
	return b
}

// WithData sets the request body data
func (b *ConfigBuilder) WithData(data string) *ConfigBuilder {
	b.config.Data = data
	return b
}

// WithVerbose enables verbose output
func (b *ConfigBuilder) WithVerbose() *ConfigBuilder {
	b.config.Verbose = true
	return b
}

// WithIncludeHeaders enables header output
func (b *ConfigBuilder) WithIncludeHeaders() *ConfigBuilder {
	b.config.IncludeHeaders = true
	return b
}

// WithShowDocs enables documentation mode
func (b *ConfigBuilder) WithShowDocs() *ConfigBuilder {
	b.config.ShowDocs = true
	return b
}

// WithSigV4 enables AWS SigV4 authentication
func (b *ConfigBuilder) WithSigV4(service string) *ConfigBuilder {
	b.config.SigV4Enabled = true
	b.config.SigV4Service = service
	return b
}

// WithMCP configures for MCP mode
func (b *ConfigBuilder) WithMCP() *ConfigBuilder {
	b.config.MCP.Enabled = true
	return b
}

// WithMCPMethods sets allowed methods for MCP mode
func (b *ConfigBuilder) WithMCPMethods(methods ...string) *ConfigBuilder {
	b.config.MCP.AllowedMethods = methods
	return b
}

// WithMCPPathPrefix sets the path prefix constraint for MCP mode
func (b *ConfigBuilder) WithMCPPathPrefix(prefix string) *ConfigBuilder {
	b.config.MCP.PathPrefix = prefix
	return b
}

// WithLogLevel sets the logging level
func (b *ConfigBuilder) WithLogLevel(level string) *ConfigBuilder {
	return b
}

// Build returns the configured Config
func (b *ConfigBuilder) Build() *config.Config {
	// Create a deep copy to prevent test interference
	config := &config.Config{
		Methods:        make([]string, len(b.config.Methods)),
		Headers:        make([]string, len(b.config.Headers)),
		QueryParams:    make([]string, len(b.config.QueryParams)),
		Data:           b.config.Data,
		Server:         b.config.Server,
		Verbose:        b.config.Verbose,
		IncludeHeaders: b.config.IncludeHeaders,
		ShowDocs:       b.config.ShowDocs,
		SigV4Enabled:   b.config.SigV4Enabled,
		SigV4Service:   b.config.SigV4Service,
		OpenAPIURL:     b.config.OpenAPIURL,
		MCP: config.MCPConfig{
			Enabled:        b.config.MCP.Enabled,
			AllowedMethods: make([]string, len(b.config.MCP.AllowedMethods)),
			PathPrefix:     b.config.MCP.PathPrefix,
			Headers:        make([]string, len(b.config.MCP.Headers)),
			SigV4:          b.config.MCP.SigV4,
			SigV4Service:   b.config.MCP.SigV4Service,
			ServerURL:      b.config.MCP.ServerURL,
			OpenAPIURL:     b.config.MCP.OpenAPIURL,
		},
	}

	copy(config.Methods, b.config.Methods)
	copy(config.Headers, b.config.Headers)
	copy(config.QueryParams, b.config.QueryParams)
	copy(config.MCP.AllowedMethods, b.config.MCP.AllowedMethods)
	copy(config.MCP.Headers, b.config.MCP.Headers)

	return config
}

// Common pre-built configs for frequent test scenarios

// BasicGETConfig returns a basic GET request configuration
func BasicGETConfig() *config.Config {
	return NewConfigBuilder().WithMethod("GET").Build()
}

// BasicPOSTConfig returns a basic POST request configuration with JSON data
func BasicPOSTConfig(data string) *config.Config {
	return NewConfigBuilder().
		WithMethod("POST").
		WithData(data).
		WithHeader("Content-Type: application/json").
		Build()
}

// VerboseConfig returns a configuration with verbose output enabled
func VerboseConfig() *config.Config {
	return NewConfigBuilder().
		WithMethod("GET").
		WithVerbose().
		WithIncludeHeaders().
		Build()
}

// MCPConfig returns a basic MCP server configuration
func MCPConfig(allowedMethods []string, pathPrefix string) *config.Config {
	return NewConfigBuilder().
		WithMCP().
		WithMCPMethods(allowedMethods...).
		WithMCPPathPrefix(pathPrefix).
		Build()
}

// AuthenticatedConfig returns a configuration with SigV4 authentication
func AuthenticatedConfig(service string) *config.Config {
	return NewConfigBuilder().
		WithMethod("GET").
		WithSigV4(service).
		Build()
}