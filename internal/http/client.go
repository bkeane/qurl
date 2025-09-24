package http

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	qurlhttp "github.com/brendan.keane/qurl/pkg/http"
	"github.com/brendan.keane/qurl/pkg/openapi"
	"github.com/rs/zerolog"
)

// Client wraps HTTP functionality with structured logging and error handling
type Client struct {
	logger     zerolog.Logger
	httpClient *qurlhttp.Client
	viewer     *openapi.Viewer
	config     *config.Config
}

// NewClient creates a new HTTP client wrapper
func NewClient(logger zerolog.Logger, cfg *config.Config) (*Client, error) {
	httpClient, err := qurlhttp.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to create HTTP client")
	}

	var viewer *openapi.Viewer
	if cfg.OpenAPIURL != "" {
		viewer = openapi.NewViewer(httpClient, cfg.OpenAPIURL)
	}

	return &Client{
		logger:     logger.With().Str("component", "http_client").Logger(),
		httpClient: httpClient,
		viewer:     viewer,
		config:     cfg,
	}, nil
}

// Execute performs an HTTP request with the given path
func (c *Client) Execute(ctx context.Context, path string) error {
	logger := c.logger.With().
		Str("method", c.config.PrimaryMethod()).
		Str("path", path).
		Logger()

	logger.Debug().Msg("executing HTTP request")

	// Resolve target URL
	targetURL, err := c.resolveTargetURL(ctx, path)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve target URL")
		return err
	}

	logger.Debug().Str("target_url", targetURL).Msg("URL resolved")

	// Add query parameters
	targetURL, err = c.applyQueryParameters(targetURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to apply query parameters")
		return errors.Wrap(err, errors.ErrorTypeValidation, "invalid query parameters")
	}

	// Build HTTP request
	req, err := c.buildHTTPRequest(ctx, c.config.PrimaryMethod(), targetURL, path)
	if err != nil {
		logger.Error().Err(err).Msg("failed to build HTTP request")
		return err
	}

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("HTTP request failed")
		return errors.Wrap(err, errors.ErrorTypeNetwork, "HTTP request failed").
			WithContext("url", targetURL).
			WithContext("duration", duration)
	}
	defer resp.Body.Close()

	logger.Debug().
		Int("status", resp.StatusCode).
		Dur("duration", duration).
		Msg("HTTP request completed")

	// Handle response
	return c.handleResponse(resp, c.config.PrimaryMethod(), targetURL)
}

// ExecuteForMCP performs an HTTP request and returns the response data
// This is used by the MCP server to capture response without printing to stdout
func (c *Client) ExecuteForMCP(ctx context.Context, path string) (string, map[string][]string, int, error) {
	logger := c.logger.With().
		Str("method", c.config.PrimaryMethod()).
		Str("path", path).
		Logger()
	logger.Debug().Msg("executing HTTP request for MCP")

	// Resolve target URL
	targetURL, err := c.resolveTargetURL(ctx, path)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve target URL")
		return "", nil, 0, err
	}
	logger.Debug().Str("target_url", targetURL).Msg("URL resolved")

	// Add query parameters
	targetURL, err = c.applyQueryParameters(targetURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to apply query parameters")
		return "", nil, 0, errors.Wrap(err, errors.ErrorTypeValidation, "invalid query parameters")
	}

	// Build HTTP request
	req, err := c.buildHTTPRequest(ctx, c.config.PrimaryMethod(), targetURL, path)
	if err != nil {
		logger.Error().Err(err).Msg("failed to build HTTP request")
		return "", nil, 0, err
	}

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("HTTP request failed")
		return "", nil, 0, errors.Wrap(err, errors.ErrorTypeNetwork, "HTTP request failed").
			WithContext("url", targetURL).
			WithContext("duration", duration)
	}
	defer resp.Body.Close()

	logger.Debug().
		Int("status", resp.StatusCode).
		Dur("duration", duration).
		Msg("HTTP request completed")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read response body")
		return "", nil, resp.StatusCode, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to read response body")
	}

	return string(body), resp.Header, resp.StatusCode, nil
}

// ShowDocs displays OpenAPI documentation
func (c *Client) ShowDocs(ctx context.Context, path, method string) error {
	if c.viewer == nil {
		return errors.New(errors.ErrorTypeConfig, "OpenAPI URL is required for documentation").
			WithContext("config_type", "openapi")
	}

	logger := c.logger.With().
		Str("doc_path", path).
		Str("doc_method", method).
		Logger()

	logger.Debug().Msg("showing OpenAPI documentation")

	if path == "" {
		path = "*"
	}
	if method == "" {
		method = "ANY"
	}

	output, err := c.viewer.View(ctx, path, method)
	if err != nil {
		logger.Error().Err(err).Msg("failed to view OpenAPI documentation")
		return errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to view OpenAPI documentation")
	}

	// Print documentation to stdout
	logger.Debug().Int("doc_length", len(output)).Msg("documentation retrieved")
	fmt.Println(output)

	return nil
}

// resolveTargetURL determines the final URL for the request
func (c *Client) resolveTargetURL(ctx context.Context, path string) (string, error) {
	// Parse the path to check if it's an absolute URL
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "invalid URL/path").
			WithContext("path", path)
	}

	// Check if we have a full URL (with scheme and host)
	if parsedURL.Scheme != "" && parsedURL.Host != "" {
		return path, nil
	}

	// Path only - need to get base URL
	var baseURL string

	// Priority 1: --server flag provided
	if c.config.Server != "" {
		baseURL, err = c.resolveServerURL(c.config.Server)
		if err != nil {
			return "", errors.Wrap(err, errors.ErrorTypeConfig, "failed to resolve server URL")
		}
	} else if c.config.OpenAPIURL != "" && c.viewer != nil {
		// Priority 2: Get base URL from OpenAPI spec
		baseURL, err = c.viewer.BaseURL(ctx)
		if err != nil {
			return "", errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to get base URL from OpenAPI spec")
		}
	} else {
		// Priority 3: Error - no way to construct URL
		return "", errors.New(errors.ErrorTypeConfig, "server URL required when using relative paths").
			WithContext("suggestion", "use --server flag, --openapi flag, or provide full URL")
	}

	// Combine base URL with path
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "invalid base URL from OpenAPI").
			WithContext("base_url", baseURL)
	}

	// Handle relative path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Properly append path to base URL
	if base.Path != "" && base.Path != "/" {
		base.Path = strings.TrimSuffix(base.Path, "/") + path
	} else {
		base.Path = path
	}

	return base.String(), nil
}

// resolveServerURL resolves the server URL based on the --server flag value
func (c *Client) resolveServerURL(serverFlag string) (string, error) {
	// Check if it's a numeric index (0, 1, 2, etc.)
	if len(serverFlag) == 1 && serverFlag[0] >= '0' && serverFlag[0] <= '9' {
		// Parse as server index
		index := int(serverFlag[0] - '0')

		if c.viewer == nil {
			return "", errors.New(errors.ErrorTypeConfig, "server index requires OpenAPI specification").
				WithContext("index", index)
		}

		servers, err := c.viewer.GetServers()
		if err != nil {
			return "", errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to get servers from OpenAPI spec")
		}

		if index >= len(servers) {
			return "", errors.Newf(errors.ErrorTypeValidation, "server index %d out of range (available: 0-%d)", index, len(servers)-1)
		}

		serverURL := servers[index].URL
		if serverURL == "" {
			return "", errors.Newf(errors.ErrorTypeValidation, "server at index %d has empty URL", index)
		}

		// Handle relative URLs
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			if c.viewer == nil {
				return "", errors.New(errors.ErrorTypeConfig, "relative server URL requires OpenAPI specification")
			}
			// Use the viewer's logic to resolve relative URLs
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return c.viewer.BaseURL(ctx)
		}

		return serverURL, nil
	}

	// Treat as full URL
	parsedURL, err := url.Parse(serverFlag)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "invalid server URL").
			WithContext("server_url", serverFlag)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", errors.New(errors.ErrorTypeValidation, "server URL must be complete (e.g., https://example.com)").
			WithContext("server_url", serverFlag)
	}

	return serverFlag, nil
}