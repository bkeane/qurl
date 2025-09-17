package http

import (
	"net/http"
)

// Transport implements http.RoundTripper with Lambda support
type Transport struct {
	*Client
}

// NewTransport creates a new transport with Lambda support
func NewTransport() (*Transport, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Transport{Client: client}, nil
}

// RoundTrip implements the http.RoundTripper interface
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Do(req)
}

// DefaultTransport is a transport with Lambda support using default configuration
var DefaultTransport http.RoundTripper = func() http.RoundTripper {
	transport, err := NewTransport()
	if err != nil {
		// Fall back to standard HTTP transport if AWS config fails
		return http.DefaultTransport
	}
	return transport
}()

// DefaultClient is an HTTP client with Lambda support using default configuration
var DefaultClient = &http.Client{
	Transport: DefaultTransport,
}