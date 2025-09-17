# OpenAPI Viewer and Parser

Parse and display OpenAPI v3 specifications with support for documentation viewing, path completion, and parameter discovery.

## Installation

```go
import "github.com/brendan.keane/qurl/pkg/openapi"
```

## Quick Start

```go
// Create HTTP client
client, err := http.NewClient()

// Create viewer with client and spec URL
viewer := openapi.NewViewer(client, "https://api.example.com/openapi.yaml")

// View documentation
docs, err := viewer.View(ctx, "/users", "GET")
fmt.Println(docs)

// Get path completions for shell completion
paths, err := viewer.PathCompletions(ctx, "", "GET")

// Get parameter completions
params, err := viewer.ParamCompletions(ctx, "/users", "GET")
```

## API Reference

### Viewer

```go
// Create viewer with HTTP client and spec URL
viewer := openapi.NewViewer(httpClient, "https://api.example.com/openapi.yaml")

// Create viewer with Lambda-enabled client
lambdaClient, err := http.NewClient()
viewer := openapi.NewViewer(lambdaClient, "lambda://spec-service/openapi.yaml")
```

### Documentation Viewing

```go
// View all endpoints
docs, err := viewer.View(ctx, "*", "*")

// View specific endpoint
docs, err := viewer.View(ctx, "/users/{id}", "GET")

// View with trailing slash shows index of sub-paths
docs, err := viewer.View(ctx, "/users/", "*")
```

### Completions

```go
// Get all paths for a method
paths, err := viewer.PathCompletions(ctx, "", "GET")
// Returns: ["/users", "/users/{id}", "/orders", ...]

// Get query parameters for an endpoint
params, err := viewer.ParamCompletions(ctx, "/users", "GET")
// Returns: ["limit", "offset", "search", ...]

// Get HTTP methods for a path
methods, err := viewer.MethodCompletions(ctx, "/users")
// Returns: ["GET", "POST", "DELETE", ...]
```

### Parser (Advanced)

```go
// Create parser
parser := openapi.NewParser()

// Load from URL
err := parser.LoadFromURL(ctx, "https://api.example.com/openapi.yaml")

// Load from bytes
err := parser.LoadFromBytes(specData)

// Get structured path information
paths, err := parser.GetPaths("/users*", "GET")
```

## URL Sources

Works with any URL scheme supported by your HTTP client:

```go
// Standard HTTP
viewer := openapi.NewViewer(httpClient, "https://api.example.com/spec.yaml")
docs, err := viewer.View(ctx, "/users", "GET")

// Local files
viewer := openapi.NewViewer(httpClient, "file:///path/to/spec.yaml")
docs, err := viewer.View(ctx, "/users", "GET")

// Lambda functions (with Lambda-enabled HTTP client)
lambdaClient, _ := http.NewClient()
viewer := openapi.NewViewer(lambdaClient, "lambda://spec-service/openapi.yaml")
docs, err := viewer.View(ctx, "/users", "GET")
```

## Path Filtering

```go
// Exact match
paths, err := parser.GetPaths("/users/{id}", "GET")

// Prefix match with wildcard
paths, err := parser.GetPaths("/users*", "GET")

// All paths
paths, err := parser.GetPaths("*", "GET")
```

## Method Filtering

```go
// Specific method
paths, err := parser.GetPaths("/users", "GET")

// All methods
paths, err := parser.GetPaths("/users", "*")
paths, err := parser.GetPaths("/users", "ANY")
```

## Error Handling

```go
docs, err := viewer.View(ctx, path, method)
if err != nil {
    log.Printf("Failed to load OpenAPI spec: %v", err)
}

if docs == "No endpoints found matching the specified path and method" {
    log.Println("Path not found in API specification")
}
```

## Testing

```go
// Test with mock HTTP client
mockClient := &MockHTTPClient{Response: mockResponse}
viewer := openapi.NewViewer(mockClient, "https://api.example.com/spec.yaml")

// Test with pre-loaded spec
viewer := openapi.NewViewer(mockClient, "")
err := viewer.parser.LoadFromBytes(specBytes)
paths, err := viewer.ParamCompletions(ctx, "/users", "GET")
```