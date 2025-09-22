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
- **Bearer Auth**: `--bearer token` for JWT/OAuth tokens
- **Shell Completion**: Tab completion for API paths, methods, and parameters

## Configuration

Set `OPENAPI_URL` or `QURL_OPENAPI` to your OpenAPI specification URL. Override the server with `QURL_SERVER` or the `--server` flag. AWS credentials work through the standard AWS credential chain.