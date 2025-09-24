package cli

import (
	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/brendan.keane/qurl/internal/mcp"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// MCPHandler handles MCP server commands
type MCPHandler struct {
	logger zerolog.Logger
}

// NewMCPHandler creates a new MCP command handler
func NewMCPHandler(logger zerolog.Logger) *MCPHandler {
	return &MCPHandler{
		logger: logger.With().Str("handler", "mcp").Logger(),
	}
}

// Execute handles the MCP server command
func (h *MCPHandler) Execute(cmd *cobra.Command, args []string) error {
	// Get config from context (already loaded in main.go)
	cfg, ok := config.FromContext(cmd.Context())
	if !ok {
		// Fallback to loading from flags
		var err error
		cfg, err = config.LoadFromFlags(cmd.Flags())
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to load configuration")
			return err
		}
	}

	// Methods are already propagated to MCP config in LoadFromFlags
	// If only default GET method and request flag not changed, allow all methods
	if len(cfg.Methods) == 1 && cfg.Methods[0] == "GET" && !cmd.Flags().Changed("request") {
		cfg.MCP.AllowedMethods = []string{} // Empty means all methods allowed
	}
	// Otherwise, cfg.MCP.AllowedMethods is already set to cfg.Methods

	// Validate that we have an OpenAPI URL
	if cfg.OpenAPIURL == "" {
		h.logger.Error().Msg("OpenAPI URL is required for MCP server")
		return errors.New(errors.ErrorTypeConfig, "OpenAPI URL is required for MCP server").
			WithContext("suggestion", "use --openapi flag or set QURL_OPENAPI environment variable")
	}

	h.logger.Debug().
		Str("openapi_url", cfg.OpenAPIURL).
		Str("path_prefix", cfg.MCP.PathPrefix).
		Strs("allowed_methods", cfg.MCP.AllowedMethods).
		Bool("sigv4", cfg.MCP.SigV4).
		Int("headers", len(cfg.MCP.Headers)).
		Msg("starting MCP server")

	// Create MCP server
	server, err := mcp.NewServer(h.logger, cfg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create MCP server")
		return err
	}

	h.logger.Debug().Msg("MCP server created, starting message loop")

	// Start the server
	return server.Start()
}