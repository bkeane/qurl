package main

import (
	"context"
	"os"
	"time"

	"github.com/brendan.keane/qurl/internal/cli"
	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	qurlogger "github.com/brendan.keane/qurl/internal/logger"
	"github.com/brendan.keane/qurl/pkg/openapi"
	qurlhttp "github.com/brendan.keane/qurl/pkg/http"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func main() {
	if err := execute(); err != nil {
		errors.PresentError(err)
		os.Exit(1)
	}
}

// execute runs the root command
func execute() error {
	var cfg config.Config
	var mcpMode bool

	rootCmd := &cobra.Command{
		Use:   "qurl [path]",
		Short: "A modern CLI for testing OpenAPI endpoints",
		Long: `qurl is a command-line tool for interacting with OpenAPI-defined REST APIs.
It can automatically discover endpoints, validate requests, and provide
documentation - all from your OpenAPI specification.

Examples:
  # Make a GET request
  qurl /api/users

  # Show OpenAPI documentation
  qurl --docs /api/users

  # Make a POST request with data
  qurl -X POST -d '{"name":"Alice"}' /api/users

  # Show documentation for multiple methods
  qurl -X GET -X POST -X PUT --docs /api/users

  # Start MCP server for LLM integration (unconstrained)
  qurl --mcp

  # Start MCP server constrained to GET requests under /api/
  qurl -X GET /api/ --mcp

  # Start MCP server with multiple allowed methods
  qurl -X GET -X POST -X PUT /api/users --mcp`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only complete paths when we have 0 args (first positional argument)
			if len(args) == 0 {
				// Get OpenAPI URL - try multiple sources
				openAPIURL := os.Getenv("QURL_OPENAPI")
				if openAPIURL == "" {
					openAPIURL = os.Getenv("OPENAPI_URL")
				}
				if flagVal, _ := cmd.Flags().GetString("openapi"); flagVal != "" {
					openAPIURL = flagVal
				}

				// If no OpenAPI spec available, provide no completions (let shell handle files if needed)
				if openAPIURL == "" {
					return nil, cobra.ShellCompDirectiveDefault
				}

				// Simple, fast completion - just try to get paths, fail gracefully
				httpClient, err := qurlhttp.NewClient()
				if err != nil {
					return nil, cobra.ShellCompDirectiveDefault
				}

				viewer := openapi.NewViewer(httpClient, openAPIURL)

				// Use a short timeout to avoid hanging on slow networks
				ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Second)
				defer cancel()

				// Get method for filtering (use first specified method, default to GET)
				method := "GET"
				if methods, _ := cmd.Flags().GetStringSlice("request"); len(methods) > 0 {
					method = methods[0]
				}

				paths, err := viewer.PathCompletions(ctx, toComplete, method)
				if err != nil {
					// Fail gracefully - no completion rather than error
					return nil, cobra.ShellCompDirectiveDefault
				}

				return paths, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveDefault
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration from flags
			cfgPtr, err := config.LoadFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			cfg = *cfgPtr

			// Set MCP mode
			cfg.MCP.Enabled = mcpMode

			// If MCP mode and we have a path argument, use it as path prefix
			if mcpMode && len(args) > 0 {
				cfg.MCP.PathPrefix = args[0]
			}

			// Initialize logger
			loggerConfig := &qurlogger.Config{
				Level:      cfg.Logger.Level,
				Pretty:     cfg.Logger.Pretty,
				WithCaller: false,
				Output:     os.Stderr,
				TimeFormat: time.RFC3339,
			}
			logger := qurlogger.InitLogger(loggerConfig)

			// Store logger and config in context
			ctx := logger.WithContext(cmd.Context())
			ctx = config.WithConfig(ctx, &cfg)
			cmd.SetContext(ctx)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := zerolog.Ctx(cmd.Context())

			// Check if MCP mode is enabled
			if mcpMode {
				// Validate that incompatible flags aren't set
				if cfg.Data != "" {
					return errors.New(errors.ErrorTypeValidation, "cannot use --data flag with --mcp mode")
				}
				if cfg.ShowDocs {
					return errors.New(errors.ErrorTypeValidation, "cannot use --docs flag with --mcp mode")
				}
				if cfg.IncludeHeaders {
					return errors.New(errors.ErrorTypeValidation, "cannot use --include flag with --mcp mode")
				}

				// Start MCP server
				handler := cli.NewMCPHandler(*logger)
				return handler.Execute(cmd, args)
			}

			// Default: HTTP request mode
			handler := cli.NewHTTPHandler(*logger)
			return handler.Execute(cmd, args)
		},
	}

	// Add flags
	flags := rootCmd.PersistentFlags()

	// MCP mode flag
	flags.BoolVar(&mcpMode, "mcp", false, "Start MCP server for LLM integration")

	// OpenAPI and server configuration
	flags.StringVar(&cfg.OpenAPIURL, "openapi", "", "OpenAPI specification URL")
	flags.StringVar(&cfg.Server, "server", "", "Server URL or index (0,1,2...) from OpenAPI spec")

	// HTTP configuration
	flags.StringSliceVarP(&cfg.Methods, "request", "X", []string{"GET"}, "HTTP method to use (can be used multiple times)")
	flags.StringSliceVarP(&cfg.Headers, "header", "H", nil, "Custom headers (format: 'Name: Value')")
	flags.StringSliceVarP(&cfg.QueryParams, "query", "q", nil, "Query parameters (format: 'key=value')")
	flags.StringVarP(&cfg.Data, "data", "d", "", "Request body data")

	// Output configuration
	flags.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	flags.BoolVarP(&cfg.IncludeHeaders, "include", "i", false, "Include response headers in output")
	flags.BoolVar(&cfg.ShowDocs, "docs", false, "Show OpenAPI documentation for the endpoint")

	// Authentication
	flags.BoolVar(&cfg.SigV4Enabled, "aws-sigv4", false, "Sign requests with AWS SigV4")
	flags.StringVar(&cfg.SigV4Service, "aws-service", "execute-api", "AWS service name for SigV4 signing")

	// Logging configuration
	flags.StringVar(&cfg.Logger.Level, "log-level", "warn", "Log level (trace, debug, info, warn, error)")
	flags.BoolVar(&cfg.Logger.Pretty, "log-pretty", true, "Enable pretty logging output")

	// Environment variable bindings
	rootCmd.MarkPersistentFlagFilename("openapi")

	// Register completion functions for flags
	rootCmd.RegisterFlagCompletionFunc("request", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Always provide standard HTTP methods - simple and reliable
		commonMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

		// Try to enhance with OpenAPI-specific methods, but don't fail if we can't
		openAPIURL := os.Getenv("QURL_OPENAPI")
		if openAPIURL == "" {
			openAPIURL = os.Getenv("OPENAPI_URL")
		}
		if flagVal, _ := cmd.Flags().GetString("openapi"); flagVal != "" {
			openAPIURL = flagVal
		}

		if openAPIURL != "" {
			// Quick attempt to get OpenAPI-specific methods
			if httpClient, err := qurlhttp.NewClient(); err == nil {
				viewer := openapi.NewViewer(httpClient, openAPIURL)
				ctx, cancel := context.WithTimeout(cmd.Context(), 1*time.Second)
				defer cancel()

				// Get path from args if available
				path := "*"
				if len(args) > 0 {
					path = args[0]
				}

				if methods, err := viewer.MethodCompletions(ctx, path); err == nil && len(methods) > 0 {
					return methods, cobra.ShellCompDirectiveNoFileComp
				}
			}
		}

		// Fall back to common methods
		return commonMethods, cobra.ShellCompDirectiveNoFileComp
	})

	// Add completion command (keeping this as the only subcommand for shell completions)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(qurl completion bash)
  $ echo 'source <(qurl completion bash)' >> ~/.bashrc

Zsh:
  $ source <(qurl completion zsh)
  $ echo 'source <(qurl completion zsh)' >> ~/.zshrc

Fish:
  $ qurl completion fish | source
  $ qurl completion fish > ~/.config/fish/completions/qurl.fish

PowerShell:
  PS> qurl completion powershell | Out-String | Invoke-Expression`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})

	return rootCmd.Execute()
}