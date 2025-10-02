package openapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// SetHeaders enriches an HTTP request with headers based on the OpenAPI specification.
// It sets the Accept header based on response content types defined in the spec.
// If no exact match is found or the spec is unavailable, it returns without error.
func (v *Viewer) SetHeaders(ctx context.Context, req *http.Request, path, method string) error {
	if v.specURL == "" {
		// No spec URL, nothing to do
		return nil
	}

	if err := v.ensureSpecLoaded(ctx); err != nil {
		return err
	}

	// Get all paths matching this pattern and method
	paths, err := v.parser.GetPaths(path, method)
	if err != nil {
		// If no document loaded, just return without error
		if err.Error() == "no OpenAPI document loaded" {
			return nil
		}
		return fmt.Errorf("getting paths: %w", err)
	}

	// Find exact match
	var exactMatch *PathInfo
	for _, p := range paths {
		if p.Path == path && strings.EqualFold(p.Method, method) {
			exactMatch = &p
			break
		}
	}

	if exactMatch == nil {
		// No exact match found, return without setting headers
		return nil
	}

	// Extract and set Accept header from response content types
	if exactMatch.Responses != nil && exactMatch.Responses.Codes != nil {
		acceptTypes := []string{}
		seenTypes := make(map[string]bool)

		// Check successful response codes first (2xx)
		for code, response := range exactMatch.Responses.Codes.FromOldest() {
			if strings.HasPrefix(code, "2") && response.Content != nil {
				for contentType := range response.Content.FromOldest() {
					if !seenTypes[contentType] {
						acceptTypes = append(acceptTypes, contentType)
						seenTypes[contentType] = true
					}
				}
			}
		}

		// If no 2xx responses, check all responses
		if len(acceptTypes) == 0 {
			for _, response := range exactMatch.Responses.Codes.FromOldest() {
				if response.Content != nil {
					for contentType := range response.Content.FromOldest() {
						if !seenTypes[contentType] {
							acceptTypes = append(acceptTypes, contentType)
							seenTypes[contentType] = true
						}
					}
				}
			}
		}

		// Set Accept header if we found content types
		if len(acceptTypes) > 0 {
			req.Header.Set("Accept", strings.Join(acceptTypes, ", "))
		}
	}

	// Future: Set Content-Type header from request body when body support is added
	// if exactMatch.RequestBody != nil && exactMatch.RequestBody.Content != nil && req.Body != nil {
	//     // Extract first content type as default
	//     for contentType := range exactMatch.RequestBody.Content.FromOldest() {
	//         req.Header.Set("Content-Type", contentType)
	//         break
	//     }
	// }

	return nil
}

// parseSpecURL helper function to extract scheme and host from the OpenAPI spec URL
func (v *Viewer) parseSpecURL() (scheme, host string, err error) {
	if v.specURL == "" {
		return "", "", fmt.Errorf("no spec URL available")
	}

	parsedURL, err := url.Parse(v.specURL)
	if err != nil {
		return "", "", fmt.Errorf("parsing OpenAPI URL: %w", err)
	}

	return parsedURL.Scheme, parsedURL.Host, nil
}

// BaseURL returns the base URL for API requests from the OpenAPI specification.
// It checks the servers section first, and falls back to just the spec URL's host if no servers are defined.
// This approach respects explicit server configurations while providing a conservative fallback.
func (v *Viewer) BaseURL(ctx context.Context) (string, error) {
	if err := v.ensureSpecLoaded(ctx); err != nil {
		return "", err
	}

	// Get servers from the spec
	servers, err := v.parser.GetServers()
	if err != nil {
		return "", fmt.Errorf("getting servers from spec: %w", err)
	}

	// Priority 1: Use explicit server URL from OpenAPI spec if available
	if len(servers) > 0 && servers[0].URL != "" {
		serverURL := servers[0].URL

		// Check if the server URL is absolute (has a supported scheme)
		isAbsolute := strings.HasPrefix(serverURL, "http://") ||
			strings.HasPrefix(serverURL, "https://") ||
			strings.HasPrefix(serverURL, "lambda://")

		if !isAbsolute {
			// It's a relative URL - combine with OpenAPI URL host
			scheme, host, err := v.parseSpecURL()
			if err != nil {
				return "", fmt.Errorf("server URL is relative but %w", err)
			}

			// Ensure relative URL starts with /
			if !strings.HasPrefix(serverURL, "/") {
				serverURL = "/" + serverURL
			}

			return fmt.Sprintf("%s://%s%s", scheme, host, serverURL), nil
		}

		// Absolute server URL - use as-is
		return serverURL, nil
	}

	// Priority 2: Conservative fallback to just the host from spec URL
	// No assumptions about paths - just use scheme + host
	scheme, host, err := v.parseSpecURL()
	if err != nil {
		return "", fmt.Errorf("no servers defined and %w", err)
	}

	return fmt.Sprintf("%s://%s", scheme, host), nil
}

// GetServers returns the servers from the OpenAPI specification
func (v *Viewer) GetServers() ([]*v3.Server, error) {
	if err := v.ensureSpecLoaded(context.Background()); err != nil {
		return nil, err
	}
	return v.parser.GetServers()
}
