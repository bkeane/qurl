package cli

import (
	"context"
	"strings"
	"time"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/brendan.keane/qurl/internal/http"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// HTTPHandler handles HTTP request commands
type HTTPHandler struct {
	logger zerolog.Logger
}

// NewHTTPHandler creates a new HTTP command handler
func NewHTTPHandler(logger zerolog.Logger) *HTTPHandler {
	return &HTTPHandler{
		logger: logger.With().Str("handler", "http").Logger(),
	}
}

// Execute handles the HTTP request command
func (h *HTTPHandler) Execute(cmd *cobra.Command, args []string) error {
	// Load configuration from flags
	cfg, err := config.LoadFromFlags(cmd.Flags())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to load configuration")
		return err
	}

	// Get path from arguments
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	cfg.Path = path

	// Validate that only one method is used for HTTP requests (not docs or MCP)
	if len(cfg.Methods) > 1 && !cfg.ShowDocs {
		h.logger.Error().Strs("methods", cfg.Methods).Msg("multiple methods not allowed for HTTP requests")
		return errors.New(errors.ErrorTypeValidation, "cannot specify multiple HTTP methods for a single request").
			WithContext("methods", cfg.Methods).
			WithContext("suggestion", "use -X with a single method (e.g., -X POST) or add --docs flag to view documentation for multiple methods")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		h.logger.Error().Err(err).Msg("configuration validation failed")
		return err
	}

	h.logger.Debug().
		Strs("methods", cfg.Methods).
		Str("path", path).
		Bool("docs", cfg.ShowDocs).
		Msg("processing HTTP command")

	// Create HTTP client
	client, err := http.NewClient(h.logger, cfg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create HTTP client")
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle documentation request
	if cfg.ShowDocs {
		h.logger.Debug().Msg("showing documentation")

		// For documentation, always show unified view
		// When multiple methods specified, pass them as comma-separated string
		// When single method specified, show that specific method (unless default GET)
		var docsMethod string
		if len(cfg.Methods) > 1 {
			// Multiple methods: pass as comma-separated string for filtering
			docsMethod = strings.Join(cfg.Methods, ",")
		} else {
			// Single method
			docsMethod = cfg.Methods[0]
			if cfg.Methods[0] == "GET" && !cmd.Flags().Changed("request") {
				// Default GET case: show all methods
				docsMethod = "ANY"
			}
		}

		return client.ShowDocs(ctx, path, docsMethod)
	}

	// Validate that we have a path for HTTP requests
	if path == "" {
		h.logger.Warn().Msg("no path provided for HTTP request")
		return errors.New(errors.ErrorTypeValidation, "path is required for HTTP requests").
			WithContext("suggestion", "provide a URL or path as an argument")
	}

	// Execute HTTP request
	h.logger.Debug().Msg("executing HTTP request")
	return client.Execute(ctx, path)
}