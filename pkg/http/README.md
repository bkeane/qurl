# HTTP Client with Lambda Support

Drop-in replacement for Go's `net/http.Client` that transparently supports direct AWS Lambda function invocations via `lambda://` URLs using the AWS SDK.

## Installation

```go
import "github.com/brendan.keane/qurl/pkg/http"
```

## Quick Start

```go
// Create client
client, err := http.NewClient()
if err != nil {
    log.Fatal(err)
}

// Regular HTTP requests work normally
resp, err := client.Get("https://api.example.com/users")

// Lambda function invocations via AWS SDK look like HTTP requests
resp, err := client.Get("lambda://my-function/users")
```

## Lambda URL Format

```
lambda://<function-name>/<path>?<query-params>
```

Examples:
- `lambda://user-service/users/123`
- `lambda://auth-service/login`
- `lambda://api/v1/data?format=json`

## API Reference

### Client

```go
// Create new client with Lambda support
client, err := http.NewClient()

// Create with custom HTTP client
client, err := http.NewClientWithHTTPClient(customHTTPClient)

// Standard HTTP methods work with both HTTP and Lambda URLs
resp, err := client.Get(url)
resp, err := client.Post(url, contentType, body)
resp, err := client.Do(request)
```

### Transport

```go
// Use as transport
transport, err := http.NewTransport()
client := &http.Client{Transport: transport}

// Use default Lambda-enabled client
resp, err := http.DefaultClient.Get("lambda://my-function/endpoint")
```

## AWS Configuration

Uses standard AWS SDK v2 configuration for direct Lambda invocation:
- Environment variables (`AWS_REGION`, `AWS_ACCESS_KEY_ID`, etc.)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (EC2/Lambda)
- Requires `lambda:InvokeFunction` permission

## Lambda Function Requirements

Your Lambda function should return API Gateway v2 format:

```json
{
  "statusCode": 200,
  "headers": {"Content-Type": "application/json"},
  "body": "{\"result\":\"success\"}",
  "isBase64Encoded": false
}
```

## Error Handling

```go
resp, err := client.Get("lambda://my-function/endpoint")
if err != nil {
    // Could be network error, Lambda error, or AWS permission error
    log.Printf("Request failed: %v", err)
}
```

## Testing

```go
// Use any HTTP client for testing
mockClient := &MockHTTPClient{...}
client := &Client{Client: mockClient}
```