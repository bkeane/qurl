package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/brendan.keane/qurl/internal/errors"
)

// handleResponse processes the HTTP response and displays output
func (c *Client) handleResponse(resp *http.Response, method, targetURL string) error {
	logger := c.logger.With().
		Int("status", resp.StatusCode).
		Str("content_type", resp.Header.Get("Content-Type")).
		Logger()

	// Show request details if verbose
	if c.config.Verbose {
		c.showRequestDetails(method, targetURL, resp.Request)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read response body")
		return errors.Wrap(err, errors.ErrorTypeNetwork, "failed to read response body")
	}

	logger.Debug().
		Int("body_length", len(body)).
		Msg("response body read")

	// Print response details based on flags
	if c.config.Verbose {
		c.showResponseDetails(resp)
	} else if c.config.IncludeHeaders {
		c.showResponseHeaders(resp)
	}

	// Always print response body
	fmt.Print(string(body))

	// Log response summary
	logger.Debug().
		Int("body_length", len(body)).
		Bool("verbose", c.config.Verbose).
		Bool("include_headers", c.config.IncludeHeaders).
		Msg("response displayed")

	return nil
}

// showRequestDetails displays request information for verbose output
func (c *Client) showRequestDetails(method, targetURL string, req *http.Request) {
	fmt.Fprintf(os.Stderr, "> %s %s\n", method, targetURL)
	fmt.Fprintf(os.Stderr, "> Host: %s\n", req.URL.Host)

	// Show all headers
	for name, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "> %s: %s\n", name, value)
		}
	}
	fmt.Fprintf(os.Stderr, ">\n")
}

// showResponseDetails displays response information for verbose output
func (c *Client) showResponseDetails(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "< HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "< %s: %s\n", key, value)
		}
	}
	fmt.Fprintf(os.Stderr, "<\n")
}

// showResponseHeaders displays response headers for include mode
func (c *Client) showResponseHeaders(resp *http.Response) {
	fmt.Printf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println() // Empty line between headers and body
}

// applyQueryParameters adds query parameters from flags to the target URL
func (c *Client) applyQueryParameters(targetURL string) (string, error) {
	if len(c.config.QueryParams) == 0 {
		return targetURL, nil
	}

	logger := c.logger.With().
		Int("param_count", len(c.config.QueryParams)).
		Logger()

	parsedTarget, err := url.Parse(targetURL)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "failed to parse target URL for query parameters").
			WithContext("url", targetURL)
	}

	query := parsedTarget.Query()
	paramCount := 0

	for _, param := range c.config.QueryParams {
		// Split on first = to get key and value
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			query.Add(key, value)
			paramCount++
		} else if len(parts) == 1 && parts[0] != "" {
			// Handle param without value (just key)
			query.Add(parts[0], "")
			paramCount++
		}
	}

	parsedTarget.RawQuery = query.Encode()

	logger.Debug().
		Int("params_applied", paramCount).
		Str("final_url", parsedTarget.String()).
		Msg("query parameters applied")

	return parsedTarget.String(), nil
}