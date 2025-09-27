# qurl

OpenAPI v3 REST client and MCP server - a curl-like CLI for OpenAPI-documented APIs.

## Install

#### Manual: 
Download pre-built binaries from [releases](https://github.com/bkeane/qurl/releases).

#### Linux/macOS:
```bash
curl -sS https://raw.githubusercontent.com/bkeane/qurl/main/install.sh | bash
```

#### Windows:
```powershell
iwr -useb https://raw.githubusercontent.com/bkeane/qurl/main/install.ps1 | iex
```

## Quick Start

```bash
# Try it with httpbin
export QURL_OPENAPI=https://httpbin.org/openapi.json
qurl /get

# Explore what's available
qurl --docs

# Make a request
qurl /anything -d '{"hello":"world"}'
```

## Configuration

```bash
export QURL_OPENAPI=https://api.example.com/openapi.json
```

## Features

### Interactive Documentation
```bash
qurl --docs                   # Show all routes
qurl --docs /users            # Show /user method
qurl --docs /users/           # Show routes under /users/
qurl --docs -X GET -X DELETE  # Show all GET/DELETE routes
```

### Curl-like arguments
```bash
qurl -X POST -H "Auth: ${TOKEN}" -d '{"name":"test"}' /users
qurl -v /debug           # Verbose output
```

### Tab Completion
```bash
qurl <TAB>                # Complete paths
qurl -X <TAB>             # Complete methods
qurl /users --param <TAB> # Complete parameters
qurl --server <TAB>       # Complete servers
```

## Integrations

### AWS API Gateway / AWS Services
```bash
qurl --sig-v4 /users     # Sign request with AWS Signature V4
```

### AWS Lambda
```bash
export QURL_OPENAPI=lambda://my-function/openapi.json
qurl /path               # Invokes Lambda directly
```

### LLM/MCP Mode
```bash
qurl --mcp                                       # Full API access for LLMs
qurl --mcp -X GET /users/                        # Restrict to GET on /users/* paths
qurl --mcp -H "Authorization: Bearer ${TOKEN}"   # Include header in all LLM requests
```