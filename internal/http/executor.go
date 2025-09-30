package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/rs/zerolog"
)

// executor implements HTTPExecutor interface
// This is the new testable implementation that uses dependency injection
type executor struct {
	logger          zerolog.Logger
	httpClient      HTTPClientProvider
	openapi         OpenAPIProvider
	urlResolver     URLResolver
	responseHandler ResponseHandler
	requestBuilder  *RequestBuilder
	config          *config.Config
}

// NewExecutorWithDependencies creates a new HTTP executor with injected dependencies
// This enables comprehensive testing through dependency injection
func NewExecutorWithDependencies(
	logger zerolog.Logger,
	httpClient HTTPClientProvider,
	openapi OpenAPIProvider,
	urlResolver URLResolver,
	responseHandler ResponseHandler,
	config *config.Config,
) HTTPExecutor {
	requestBuilder := NewRequestBuilder(logger, config, openapi)
	return &executor{
		logger:          logger,
		httpClient:      httpClient,
		openapi:         openapi,
		urlResolver:     urlResolver,
		responseHandler: responseHandler,
		requestBuilder:  requestBuilder,
		config:          config,
	}
}

// Execute performs an HTTP request with the given path
// This is the CLI mode that prints response to stdout
func (e *executor) Execute(ctx context.Context, path string) error {
	logger := e.logger.With().
		Str("method", e.config.PrimaryMethod()).
		Str("path", path).
		Logger()

	logger.Debug().Msg("executing HTTP request")

	// Build and execute the request
	resp, targetURL, err := e.executeRequest(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle the response
	return e.responseHandler.HandleResponse(resp, e.config.PrimaryMethod(), targetURL)
}

// ExecuteForMCP performs an HTTP request and returns the response data
// This is used by the MCP server to capture response without printing to stdout
func (e *executor) ExecuteForMCP(ctx context.Context, path string) (string, map[string][]string, int, error) {
	logger := e.logger.With().
		Str("method", e.config.PrimaryMethod()).
		Str("path", path).
		Logger()

	logger.Debug().Msg("executing HTTP request for MCP")

	// Build and execute the request
	resp, _, err := e.executeRequest(ctx, path)
	if err != nil {
		return "", nil, 0, err
	}
	defer resp.Body.Close()

	// Handle the response for MCP (return structured data)
	return e.responseHandler.HandleResponseForMCP(resp, e.config.PrimaryMethod(), "")
}

// ShowDocs displays OpenAPI documentation
func (e *executor) ShowDocs(ctx context.Context, path, method string) error {
	if e.openapi == nil {
		return errors.New(errors.ErrorTypeConfig, "OpenAPI URL is required for documentation").
			WithContext("config_type", "openapi")
	}

	logger := e.logger.With().
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

	output, err := e.openapi.View(ctx, path, method)
	if err != nil {
		logger.Error().Err(err).Msg("failed to view OpenAPI documentation")
		return errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to view OpenAPI documentation")
	}

	// Print documentation to stdout
	logger.Debug().Int("doc_length", len(output)).Msg("documentation retrieved")
	fmt.Println(output)

	return nil
}

// executeRequest is a shared helper for building and executing HTTP requests
// This consolidates the common logic between Execute and ExecuteForMCP
func (e *executor) executeRequest(ctx context.Context, path string) (*http.Response, string, error) {
	// Resolve target URL
	targetURL, err := e.urlResolver.ResolveURL(ctx, path)
	if err != nil {
		e.logger.Error().Err(err).Msg("failed to resolve target URL")
		return nil, "", err
	}

	e.logger.Debug().Str("target_url", targetURL).Msg("URL resolved")

	// Add query parameters
	targetURL, err = ApplyQueryParameters(targetURL, e.config.QueryParams)
	if err != nil {
		e.logger.Error().Err(err).Msg("failed to apply query parameters")
		return nil, "", errors.Wrap(err, errors.ErrorTypeValidation, "invalid query parameters")
	}

	// Build HTTP request
	req, err := e.buildHTTPRequest(ctx, e.config.PrimaryMethod(), targetURL, path)
	if err != nil {
		e.logger.Error().Err(err).Msg("failed to build HTTP request")
		return nil, "", err
	}

	// Execute request
	startTime := time.Now()
	resp, err := e.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		e.logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("HTTP request failed")
		return nil, "", errors.Wrap(err, errors.ErrorTypeNetwork, "HTTP request failed").
			WithContext("url", targetURL).
			WithContext("duration", duration)
	}

	e.logger.Debug().
		Int("status", resp.StatusCode).
		Dur("duration", duration).
		Msg("HTTP request completed")

	return resp, targetURL, nil
}


// buildHTTPRequest creates an HTTP request with proper headers and body
func (e *executor) buildHTTPRequest(ctx context.Context, method, targetURL, originalPath string) (*http.Request, error) {
	return e.requestBuilder.Build(ctx, method, targetURL, originalPath)
}