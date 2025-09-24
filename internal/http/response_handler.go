package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/brendan.keane/qurl/internal/config"
	"github.com/brendan.keane/qurl/internal/errors"
	"github.com/rs/zerolog"
)

// responseHandler implements ResponseHandler interface
// Separates response processing logic for better testing
type responseHandler struct {
	logger zerolog.Logger
	config *config.Config
}

// NewResponseHandler creates a new response handler
func NewResponseHandler(logger zerolog.Logger, config *config.Config) ResponseHandler {
	return &responseHandler{
		logger: logger.With().Str("component", "response_handler").Logger(),
		config: config,
	}
}

// HandleResponse processes the HTTP response and displays output
// This consolidates the response handling logic for better testability
func (h *responseHandler) HandleResponse(resp *http.Response, method, targetURL string) error {
	logger := h.logger.With().
		Int("status", resp.StatusCode).
		Str("content_type", resp.Header.Get("Content-Type")).
		Logger()

	// Show request details if verbose
	if h.config.Verbose {
		h.showRequestDetails(method, targetURL, resp.Request)
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
	if h.config.Verbose {
		h.showResponseDetails(resp)
	} else if h.config.IncludeHeaders {
		h.showResponseHeaders(resp)
	}

	// Always print response body
	fmt.Print(string(body))

	// Log response summary
	logger.Debug().
		Int("body_length", len(body)).
		Bool("verbose", h.config.Verbose).
		Bool("include_headers", h.config.IncludeHeaders).
		Msg("response displayed")

	return nil
}

// HandleResponseForMCP processes the HTTP response and returns structured data
// This is used by the MCP server to capture response without printing to stdout
func (h *responseHandler) HandleResponseForMCP(resp *http.Response, method, targetURL string) (string, map[string][]string, int, error) {
	logger := h.logger.With().
		Int("status", resp.StatusCode).
		Str("content_type", resp.Header.Get("Content-Type")).
		Logger()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read response body")
		return "", nil, resp.StatusCode, errors.Wrap(err, errors.ErrorTypeNetwork, "failed to read response body")
	}

	logger.Debug().
		Int("body_length", len(body)).
		Msg("response body read for MCP")

	return string(body), resp.Header, resp.StatusCode, nil
}

// showRequestDetails displays request information for verbose mode
func (h *responseHandler) showRequestDetails(method, targetURL string, req *http.Request) {
	parsedURL, _ := url.Parse(targetURL)

	fmt.Fprintf(os.Stderr, "> %s %s\n", method, targetURL)
	if parsedURL != nil {
		fmt.Fprintf(os.Stderr, "> Host: %s\n", parsedURL.Host)
	}

	// Show headers
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "> %s: %s\n", key, value)
		}
	}
	fmt.Fprintf(os.Stderr, ">\n")
}

// showResponseDetails displays detailed response information for verbose mode
func (h *responseHandler) showResponseDetails(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "< %s %s\n", resp.Proto, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "< %s: %s\n", key, value)
		}
	}
	fmt.Fprintf(os.Stderr, "<\n")
}

// showResponseHeaders displays response headers for include mode
func (h *responseHandler) showResponseHeaders(resp *http.Response) {
	fmt.Printf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println() // Empty line between headers and body
}