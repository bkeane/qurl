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
)

// applyAuthentication applies authentication to the request based on configuration
func (c *Client) applyAuthentication(req *http.Request, targetURL string) error {
	logger := c.logger.With().Str("component", "auth").Logger()

	// Check if this is a lambda:// URL - skip SigV4 for direct invocation
	if strings.HasPrefix(targetURL, "lambda://") {
		logger.Debug().Msg("lambda URL detected, skipping SigV4")
		return nil
	}

	// Apply AWS SigV4 signing if requested
	if c.config.SigV4Enabled {
		logger.Debug().
			Str("service", c.config.SigV4Service).
			Msg("applying AWS SigV4 signature")

		if err := c.applySigV4(req); err != nil {
			return errors.Wrap(err, errors.ErrorTypeAuth, "SigV4 signing failed")
		}

		logger.Debug().Msg("SigV4 signature applied successfully")
	}

	return nil
}

// applySigV4 applies AWS SigV4 signing to the request
func (c *Client) applySigV4(req *http.Request) error {
	service := c.config.SigV4Service

	// Load AWS config from default credential chain
	ctx := context.Background()
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

	c.logger.Debug().
		Str("service", service).
		Str("region", region).
		Msg("SigV4 signature applied")

	return nil
}