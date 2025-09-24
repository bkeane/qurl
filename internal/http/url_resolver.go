package http

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
)

// urlResolver implements URLResolver interface
// Separates URL resolution logic for better testing
type urlResolver struct {
	config  *config.Config
	openapi OpenAPIProvider
}

// NewURLResolver creates a new URL resolver with the given configuration
func NewURLResolver(config *config.Config, openapi OpenAPIProvider) URLResolver {
	return &urlResolver{
		config:  config,
		openapi: openapi,
	}
}

// ResolveURL determines the final URL for the request
// This consolidates the complex URL resolution logic that was previously scattered
func (r *urlResolver) ResolveURL(ctx context.Context, path string) (string, error) {
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
	if r.config.Server != "" {
		baseURL, err = r.resolveServerURL(r.config.Server)
		if err != nil {
			return "", err
		}
	} else {
		// Priority 2: Get server URL from OpenAPI spec
		if r.openapi != nil {
			// Try to get base URL from OpenAPI viewer
			baseURL, err = r.openapi.BaseURL(ctx)
			if err != nil {
				return "", errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to get base URL from OpenAPI spec")
			}
		} else {
			return "", errors.New(errors.ErrorTypeConfig, "no server URL available").
				WithContext("path", path).
				WithContext("suggestion", "use --server flag or provide OpenAPI URL")
		}
	}

	// Parse the base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "invalid base URL").
			WithContext("base_url", baseURL)
	}

	// Ensure path starts with /
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
func (r *urlResolver) resolveServerURL(serverFlag string) (string, error) {
	// Check if it's a numeric index (0, 1, 2, etc.)
	if len(serverFlag) == 1 && serverFlag[0] >= '0' && serverFlag[0] <= '9' {
		// Parse as server index
		index := int(serverFlag[0] - '0')

		if r.openapi == nil {
			return "", errors.New(errors.ErrorTypeConfig, "server index requires OpenAPI specification").
				WithContext("index", index)
		}

		// Try to get servers from OpenAPI viewer
		servers, err := r.openapi.GetServers()
		if err != nil {
			return "", errors.Wrap(err, errors.ErrorTypeOpenAPI, "failed to get servers from OpenAPI spec")
		}

		if index >= len(servers) {
			return "", errors.New(errors.ErrorTypeValidation, "server index out of range").
				WithContext("index", index).
				WithContext("available_servers", len(servers))
		}

		return servers[index], nil
	}

	// Check if it looks like a relative URL that needs OpenAPI resolution
	if !strings.Contains(serverFlag, "://") {
		// Could be a relative server URL - need OpenAPI spec to resolve
		if !strings.HasPrefix(serverFlag, "http://") && !strings.HasPrefix(serverFlag, "https://") {
			if r.openapi == nil {
				return "", errors.New(errors.ErrorTypeConfig, "relative server URL requires OpenAPI specification")
			}
			// Use the viewer's logic to resolve relative URLs
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return r.openapi.BaseURL(ctx)
		}

		return serverFlag, nil
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