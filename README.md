# qurl

HTTP client that auto-configures from OpenAPI specs. Like curl, but it understands your API.

## What it does

Point qurl at an OpenAPI spec, and it becomes a custom client for that API. It automatically discovers the server URL, available endpoints, parameters, and request formats. Browse API documentation with `--docs`, make requests with familiar curl-like syntax, and enjoy intelligent tab completion for paths, methods, and parameters - all powered by the OpenAPI specification.

## Installation

```bash
go install github.com/brendan.keane/qurl/cmd/qurl@latest
```

## Usage

Works like curl for regular HTTP requests, but becomes API-aware when you set `OPENAPI_URL` to point to an OpenAPI specification. Once configured, use relative paths that automatically resolve to the correct server, explore endpoints with `--docs`, and let tab completion guide you through available operations.

## Lambda Support

Invoke AWS Lambda functions directly using `lambda://` URLs if they implement an HTTP interface or use the [AWS Lambda Web Adapter](https://github.com/awslabs/aws-lambda-web-adapter). Just use `lambda://function-name/` like any other URL.

## Features

- **Smart Configuration**: Discovers servers, paths, and parameters from OpenAPI specs
- **API Documentation**: Built-in `--docs` flag to explore APIs interactively
- **Familiar Interface**: curl-like flags (`-X`, `-H`, `-d`, `-v`)
- **AWS Support**: SigV4 signing (`--sig-v4`) and direct Lambda invocation
- **MCP Server**: Expose any OpenAPI-documented API to LLMs via Model Context Protocol
- **Shell Completion**: Tab completion for API paths, methods, and parameters

## MCP Server Mode

qurl can act as a Model Context Protocol (MCP) server, exposing any OpenAPI-documented API to LLMs like Claude. This allows LLMs to explore API documentation and make HTTP requests directly.

### Starting the MCP Server

```bash
# Basic usage
OPENAPI_URL=https://api.example.com/openapi.json qurl mcp

# With authentication
qurl mcp --openapi https://api.example.com/openapi.json \
  -H "Authorization: Bearer YOUR_TOKEN" \
  --allow-methods GET,POST

# With AWS SigV4 signing
qurl mcp --openapi https://api.aws.com/openapi.json \
  --sig-v4 --sig-v4-service execute-api
```

### MCP Tools Provided

- **discover**: Explore API documentation, optionally filtered by path or method
- **execute**: Make HTTP requests with full control over method, path, headers, query parameters, and body

### Safety Features

- Method restrictions via `--allow-methods` (default: GET, POST, PUT, PATCH)
- All server-level headers and authentication are applied to every request
- Comprehensive error handling and response formatting

### Claude Code Integration

qurl can be configured as an MCP server to give Claude access to any OpenAPI-documented API. You can restrict access by specifying allowed HTTP methods (using multiple `-X` flags) and include authentication headers with `-H`.

#### Quick Setup

```bash
claude mcp add qurl qurl --args "--mcp" --env QURL_OPENAPI=https://api.example.com/openapi.json
```

#### Manual Configuration

```json
{
  "mcpServers": {
    "qurl": {
      "command": "qurl",
      "args": ["--mcp"],
      "env": {
        "QURL_OPENAPI": "https://api.example.com/openapi.json"
      }
    }
  }
}
```

Add authentication with `-H "Authorization: Bearer TOKEN"` or restrict methods with `-X GET -X POST` as needed. Restart Claude Code after configuration changes.

## Configuration

Set `OPENAPI_URL` or `QURL_OPENAPI` to your OpenAPI specification URL. Override the server with `QURL_SERVER` or the `--server` flag. AWS credentials work through the standard AWS credential chain.