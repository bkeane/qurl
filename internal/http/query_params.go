package http

import (
	"net/url"
	"strings"

	"github.com/brendan.keane/qurl/internal/errors"
)

// ApplyQueryParameters adds query parameters to a target URL
// This is a shared utility used by both executor and response handler
func ApplyQueryParameters(targetURL string, queryParams []string) (string, error) {
	if len(queryParams) == 0 {
		return targetURL, nil
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "failed to parse target URL for query parameters").
			WithContext("url", targetURL)
	}

	query := parsedURL.Query()

	for _, param := range queryParams {
		// Split on first = to get key and value
		parts := strings.SplitN(param, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			query.Add(key, value)
		} else if len(parts) == 1 && parts[0] != "" {
			// Handle param without value (just key)
			query.Add(parts[0], "")
		}
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}