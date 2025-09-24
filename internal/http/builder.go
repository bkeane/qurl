package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/brendan.keane/qurl/internal/errors"
)

// buildHTTPRequest creates and configures the HTTP request with all headers and body
func (c *Client) buildHTTPRequest(ctx context.Context, method, targetURL, path string) (*http.Request, error) {
	logger := c.logger.With().
		Str("method", method).
		Str("target_url", targetURL).
		Logger()

	// Create the request with body if data is provided
	var requestBody io.Reader
	if c.config.Data != "" {
		requestBody = strings.NewReader(c.config.Data)
		logger.Debug().
			Int("body_length", len(c.config.Data)).
			Msg("request body added")
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "failed to create HTTP request").
			WithContext("method", method).
			WithContext("url", targetURL)
	}

	// Set standard headers
	req.Header.Set("User-Agent", "qurl")

	// Set headers based on OpenAPI spec if available
	if c.viewer != nil && path != "" {
		headerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := c.viewer.SetHeaders(headerCtx, req, path, method); err != nil {
			// Log warning but continue - headers are not critical
			logger.Warn().
				Err(err).
				Msg("could not set headers from OpenAPI spec")
		} else {
			logger.Debug().Msg("OpenAPI headers applied")
		}
	}

	// Apply authentication
	if err := c.applyAuthentication(req, targetURL); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeAuth, "failed to apply authentication")
	}

	// Set custom headers from -H flags (these override any headers set above)
	headerCount := 0
	for _, header := range c.config.Headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			headerName := strings.TrimSpace(parts[0])
			headerValue := strings.TrimSpace(parts[1])
			if headerName != "" {
				req.Header.Set(headerName, headerValue)
				headerCount++
			}
		} else if len(parts) == 1 && parts[0] != "" {
			// Handle header without value (just name)
			headerName := strings.TrimSpace(parts[0])
			if headerName != "" {
				req.Header.Set(headerName, "")
				headerCount++
			}
		}
	}

	if headerCount > 0 {
		logger.Debug().
			Int("custom_headers", headerCount).
			Msg("custom headers applied")
	}

	// Set Content-Type header if data is provided and no custom Content-Type was set
	if c.config.Data != "" && req.Header.Get("Content-Type") == "" {
		contentType := c.detectContentType(c.config.Data)
		req.Header.Set("Content-Type", contentType)
		logger.Debug().
			Str("content_type", contentType).
			Msg("content type auto-detected")
	}

	return req, nil
}

// detectContentType attempts to detect the appropriate Content-Type for the request body
func (c *Client) detectContentType(data string) string {
	trimmed := strings.TrimSpace(data)

	// Check if it looks like JSON
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return "application/json"
	}

	// Default to form-encoded
	return "application/x-www-form-urlencoded"
}