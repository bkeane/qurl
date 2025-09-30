package http

import (
	"net/http"

	"github.com/brendan.keane/qurl/internal/config"
	qurlhttp "github.com/brendan.keane/qurl/pkg/http"
	"github.com/rs/zerolog"
)

// AuthenticatedHTTPClient wraps an HTTP client and applies authentication
// This is specifically designed for OpenAPI spec fetching and other internal requests
type AuthenticatedHTTPClient struct {
	client *qurlhttp.Client
	config *config.Config
	logger zerolog.Logger
}

// NewAuthenticatedHTTPClient creates an HTTP client that applies authentication based on config
func NewAuthenticatedHTTPClient(config *config.Config, logger zerolog.Logger) *AuthenticatedHTTPClient {
	// Create lambda-capable client
	lambdaClient, err := qurlhttp.NewClient()
	if err != nil {
		// Fallback to a basic client that can still do HTTP requests
		logger.Warn().Err(err).Msg("failed to create lambda-capable client, falling back to basic client")
		lambdaClient = &qurlhttp.Client{Client: http.DefaultClient}
	}

	return &AuthenticatedHTTPClient{
		client: lambdaClient,
		config: config,
		logger: logger.With().Str("component", "auth_http_client").Logger(),
	}
}

// Do performs an HTTP request with authentication applied if configured
func (c *AuthenticatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	logger := c.logger.With().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Logger()

	logger.Debug().Msg("performing authenticated HTTP request")

	// Create a request builder to apply authentication
	builder := NewRequestBuilder(logger, c.config, nil)

	// Apply authentication if configured
	if err := builder.applyAuthentication(req.Context(), req, req.URL.String()); err != nil {
		logger.Error().Err(err).Msg("failed to apply authentication")
		return nil, err
	}

	logger.Debug().Msg("authentication applied, performing request")
	return c.client.Do(req)
}