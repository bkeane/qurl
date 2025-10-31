package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaConfig holds default configuration for Lambda event conversion
type LambdaConfig struct {
	AccountID    string
	APIID        string
	DomainName   string
	DomainPrefix string
	UserAgent    string
}

// DefaultLambdaConfig returns the default Lambda configuration
func DefaultLambdaConfig() LambdaConfig {
	return LambdaConfig{
		AccountID:    "123456789012",   // Dummy account ID
		APIID:        "lambda-adapter", // Identifier for Lambda adapter
		DomainName:   "lambda.local",   // Dummy domain
		DomainPrefix: "lambda",         // Dummy prefix
		UserAgent:    "qurl",
	}
}

// Client wraps the standard http.Client and adds Lambda invocation support
// Lambda client is initialized lazily on first lambda:// request
type Client struct {
	*http.Client
	lambdaClient *lambda.Client
	awsConfig    *aws.Config
	initOnce     sync.Once
	initErr      error
}

// NewClient creates a new HTTP client with lazy Lambda support
// AWS config is only loaded when a lambda:// URL is actually invoked
func NewClient() (*Client, error) {
	return &Client{
		Client: http.DefaultClient,
	}, nil
}

// NewClientWithHTTPClient creates a new client with a custom HTTP client
// AWS config is only loaded when a lambda:// URL is actually invoked
func NewClientWithHTTPClient(httpClient *http.Client) (*Client, error) {
	return &Client{
		Client: httpClient,
	}, nil
}

// initLambdaClient lazily loads AWS config and creates Lambda client
// This is only called when a lambda:// URL is actually invoked
func (c *Client) initLambdaClient(ctx context.Context) error {
	c.initOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			c.initErr = fmt.Errorf("loading AWS config: %w", err)
			return
		}
		c.awsConfig = &cfg
		c.lambdaClient = lambda.NewFromConfig(cfg)
	})
	return c.initErr
}

// Do performs the request, routing to Lambda or HTTP based on the URL scheme
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "lambda" {
		return c.doLambda(req)
	}
	return c.Client.Do(req)
}

// Get performs a GET request
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post performs a POST request
func (c *Client) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// doLambda handles Lambda invocations
func (c *Client) doLambda(req *http.Request) (*http.Response, error) {
	// Initialize Lambda client on first use
	if err := c.initLambdaClient(req.Context()); err != nil {
		return nil, err
	}

	// Extract Lambda function name from hostname
	functionName := req.URL.Host
	if functionName == "" {
		return nil, fmt.Errorf("lambda URL missing function name")
	}

	// Convert HTTP request to Lambda proxy event
	event, err := httpRequestToLambdaEvent(req)
	if err != nil {
		return nil, fmt.Errorf("converting request to Lambda event: %w", err)
	}

	// Marshal the event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshaling Lambda event: %w", err)
	}

	// Invoke the Lambda function
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	output, err := c.lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: types.InvocationTypeRequestResponse,
		Payload:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("invoking Lambda function: %w", err)
	}

	// Check for Lambda errors
	if output.FunctionError != nil {
		return nil, fmt.Errorf("Lambda function error: %s", *output.FunctionError)
	}

	// Convert Lambda response to HTTP response
	return lambdaResponseToHTTP(output.Payload)
}

// httpRequestToLambdaEvent converts an http.Request to an API Gateway v2 HTTP proxy event
func httpRequestToLambdaEvent(req *http.Request) (*events.APIGatewayV2HTTPRequest, error) {
	// Read body if present
	var bodyString string
	var isBase64Encoded bool

	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
		// Restore body for potential retries
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// For now, we'll send everything as plain text
		bodyString = string(bodyBytes)
		isBase64Encoded = false
	}

	// Build headers map
	headers := make(map[string]string)
	for key, values := range req.Header {
		headers[key] = strings.Join(values, ",")
	}

	// Add Host header from req.Host (it's not in req.Header in Go's http package)
	if req.Host != "" {
		headers["Host"] = req.Host
	}

	// Build query string parameters
	queryParams := make(map[string]string)
	for key, values := range req.URL.Query() {
		queryParams[key] = strings.Join(values, ",")
	}

	// Get default Lambda configuration
	config := DefaultLambdaConfig()

	// Construct API Gateway v2 event structure using the official type
	event := &events.APIGatewayV2HTTPRequest{
		Version:               "2.0",
		RouteKey:              fmt.Sprintf("%s %s", req.Method, req.URL.Path),
		RawPath:               req.URL.Path,
		RawQueryString:        req.URL.RawQuery,
		Headers:               headers,
		QueryStringParameters: queryParams,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			AccountID:    config.AccountID,
			APIID:        config.APIID,
			DomainName:   config.DomainName,
			DomainPrefix: config.DomainPrefix,
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:    req.Method,
				Path:      req.URL.Path,
				Protocol:  "HTTP/1.1",
				SourceIP:  "127.0.0.1",
				UserAgent: config.UserAgent,
			},
			RequestID: fmt.Sprintf("qurl-%d", time.Now().UnixNano()),
			RouteKey:  fmt.Sprintf("%s %s", req.Method, req.URL.Path),
			Stage:     "$default",
			Time:      time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			TimeEpoch: time.Now().UnixMilli(),
		},
		Body:            bodyString,
		IsBase64Encoded: isBase64Encoded,
	}

	return event, nil
}

// lambdaResponseToHTTP converts a Lambda response to an http.Response
func lambdaResponseToHTTP(payload []byte) (*http.Response, error) {
	// Parse the Lambda response using the official type
	var lambdaResp events.APIGatewayV2HTTPResponse

	if err := json.Unmarshal(payload, &lambdaResp); err != nil {
		return nil, fmt.Errorf("parsing Lambda response: %w", err)
	}

	// Create HTTP response
	resp := &http.Response{
		StatusCode: lambdaResp.StatusCode,
		Status:     fmt.Sprintf("%d %s", lambdaResp.StatusCode, http.StatusText(lambdaResp.StatusCode)),
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// Set headers
	for key, value := range lambdaResp.Headers {
		resp.Header.Set(key, value)
	}

	// Set body
	if lambdaResp.Body != "" {
		var bodyBytes []byte
		if lambdaResp.IsBase64Encoded {
			// Decode base64 if needed
			// For now, treating as plain text
			bodyBytes = []byte(lambdaResp.Body)
		} else {
			bodyBytes = []byte(lambdaResp.Body)
		}
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		resp.ContentLength = int64(len(bodyBytes))
	} else {
		resp.Body = io.NopCloser(bytes.NewReader([]byte{}))
		resp.ContentLength = 0
	}

	return resp, nil
}
