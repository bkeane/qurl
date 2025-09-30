package http

import (
	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	qurlhttp "github.com/brendan.keane/qurl/pkg/http"
	"github.com/brendan.keane/qurl/pkg/openapi"
	"github.com/rs/zerolog"
)

// ClientFactory centralizes HTTP client creation with dependency injection support
// This eliminates the 7 different client creation patterns found across the codebase
type ClientFactory struct {
	logger zerolog.Logger
}

// NewClientFactory creates a new client factory
func NewClientFactory(logger zerolog.Logger) *ClientFactory {
	return &ClientFactory{
		logger: logger,
	}
}

// CreateExecutor creates an HTTPExecutor with the given configuration
// This is the main entry point for all HTTP client creation in the application
func (f *ClientFactory) CreateExecutor(cfg *config.Config) (HTTPExecutor, error) {
	// Create the underlying HTTP client (with Lambda support)
	httpClient, err := qurlhttp.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to create HTTP client")
	}

	// Create OpenAPI viewer if URL is provided
	var viewer OpenAPIProvider
	if cfg.OpenAPIURL != "" {
		// Create authenticated HTTP client for OpenAPI spec fetching
		authClient := NewAuthenticatedHTTPClient(cfg, f.logger)
		openapiViewer := openapi.NewViewer(authClient, cfg.OpenAPIURL)
		viewer = NewOpenAPIAdapter(openapiViewer)
	}

	// Create URL resolver with the configuration
	resolver := NewURLResolver(cfg, viewer)

	// Create response handler
	responseHandler := NewResponseHandler(f.logger, cfg)

	// Create the main client with all dependencies
	return NewExecutorWithDependencies(
		f.logger.With().Str("component", "http_executor").Logger(),
		httpClient,
		viewer,
		resolver,
		responseHandler,
		cfg,
	), nil
}

// CreateExecutorWithCustomClient creates an HTTPExecutor with a custom HTTP client
// This is useful for testing with mock HTTP clients
func (f *ClientFactory) CreateExecutorWithCustomClient(
	cfg *config.Config,
	httpClient HTTPClientProvider,
	openapi OpenAPIProvider,
) HTTPExecutor {
	resolver := NewURLResolver(cfg, openapi)
	responseHandler := NewResponseHandler(f.logger, cfg)

	return NewExecutorWithDependencies(
		f.logger.With().Str("component", "http_executor").Logger(),
		httpClient,
		openapi,
		resolver,
		responseHandler,
		cfg,
	)
}

