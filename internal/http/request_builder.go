package http

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/brendan.keane/qurl/internal/errors"
	internalconfig "github.com/brendan.keane/qurl/internal/config"
	"github.com/rs/zerolog"
)

// RequestBuilder builds HTTP requests with authentication and headers
type RequestBuilder struct {
	logger  zerolog.Logger
	config  *internalconfig.Config
	openapi OpenAPIProvider
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(logger zerolog.Logger, cfg *internalconfig.Config, openapi OpenAPIProvider) *RequestBuilder {
	return &RequestBuilder{
		logger:  logger.With().Str("component", "request_builder").Logger(),
		config:  cfg,
		openapi: openapi,
	}
}

// Build creates and configures an HTTP request with all headers, body, and authentication
func (b *RequestBuilder) Build(ctx context.Context, method, targetURL, originalPath string) (*http.Request, error) {
	logger := b.logger.With().
		Str("method", method).
		Str("target_url", targetURL).
		Logger()

	// Create the request with body if data is provided
	var requestBody io.Reader
	if b.config.Data != "" {
		requestBody = strings.NewReader(b.config.Data)
		logger.Debug().
			Int("body_length", len(b.config.Data)).
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
	if b.openapi != nil && originalPath != "" {
		headerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := b.openapi.SetHeaders(headerCtx, req, originalPath, method); err != nil {
			// Log warning but continue - headers are not critical
			logger.Warn().
				Err(err).
				Msg("could not set headers from OpenAPI spec")
		} else {
			logger.Debug().Msg("OpenAPI headers applied")
		}
	}

	// Apply authentication
	if err := b.applyAuthentication(ctx, req, targetURL); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeAuth, "failed to apply authentication")
	}

	// Set custom headers from -H flags (these override any headers set above)
	headerCount := 0
	for _, header := range b.config.Headers {
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
	if b.config.Data != "" && req.Header.Get("Content-Type") == "" {
		contentType := b.detectContentType(b.config.Data)
		req.Header.Set("Content-Type", contentType)
		logger.Debug().
			Str("content_type", contentType).
			Msg("content type auto-detected")
	}

	return req, nil
}

// applyAuthentication applies authentication to the request based on configuration
func (b *RequestBuilder) applyAuthentication(ctx context.Context, req *http.Request, targetURL string) error {
	logger := b.logger.With().Str("component", "auth").Logger()

	// Check if this is a lambda:// URL - skip SigV4 for direct invocation
	if strings.HasPrefix(targetURL, "lambda://") {
		logger.Debug().Msg("lambda URL detected, skipping SigV4")
		return nil
	}

	// Apply AWS SigV4 signing if requested
	if b.config.SigV4Enabled {
		logger.Debug().
			Str("service", b.config.SigV4Service).
			Msg("applying AWS SigV4 signature")

		if err := b.applySigV4(ctx, req); err != nil {
			return errors.Wrap(err, errors.ErrorTypeAuth, "SigV4 signing failed")
		}

		logger.Debug().Msg("SigV4 signature applied successfully")
	}

	return nil
}

// applySigV4 applies AWS SigV4 signing to the request
func (b *RequestBuilder) applySigV4(ctx context.Context, req *http.Request) error {
	service := b.config.SigV4Service

	// Load AWS config from default credential chain
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeAuth, "failed to load AWS configuration").
			WithContext("suggestion", "ensure AWS credentials are configured")
	}

	// Extract region from the URL or use default
	region := cfg.Region
	if region == "" {
		return errors.New(errors.ErrorTypeAuth, "AWS region not configured").
			WithContext("suggestion", "set AWS_REGION or AWS_DEFAULT_REGION environment variable")
	}

	// Retrieve credentials
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeAuth, "failed to retrieve AWS credentials").
			WithContext("suggestion", "check AWS credential configuration")
	}

	// Create signer
	signer := v4.NewSigner()

	// Calculate payload hash for the signature
	var payloadHash string
	if req.Body != nil {
		// Read body to calculate hash
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, "failed to read request body for signing")
		}

		// Calculate SHA256 hash
		hash := sha256.Sum256(bodyBytes)
		payloadHash = fmt.Sprintf("%x", hash)

		// Restore body for actual request
		req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
	} else {
		// Empty body hash
		hash := sha256.Sum256([]byte{})
		payloadHash = fmt.Sprintf("%x", hash)
	}

	// Sign the request
	err = signer.SignHTTP(ctx, creds, req, payloadHash, service, region, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeAuth, "failed to sign request with SigV4").
			WithContext("service", service).
			WithContext("region", region)
	}

	b.logger.Debug().
		Str("service", service).
		Str("region", region).
		Msg("SigV4 signature applied")

	return nil
}

// detectContentType attempts to detect the appropriate Content-Type for the request body
func (b *RequestBuilder) detectContentType(data string) string {
	trimmed := strings.TrimSpace(data)

	// Check if it looks like JSON
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return "application/json"
	}

	// Default to form-encoded
	return "application/x-www-form-urlencoded"
}