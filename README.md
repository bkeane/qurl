# qurl

curl for APIs. Point it at an OpenAPI spec, get instant API client.

## Quick Start

```bash
# Install
go install github.com/brendan.keane/qurl/cmd/qurl@latest

# Try it
export OPENAPI_URL=https://httpbin.org/openapi.json
qurl /get
```

That's it. qurl reads the OpenAPI spec, finds the server, and makes the request.

## Real Examples

```bash
# GitHub API
export OPENAPI_URL=https://api.github.com/openapi.json
qurl /users/octocat

# With auth
qurl -H "Authorization: token ghp_xxx" /user

# POST data
qurl -X POST -d '{"name":"test"}' /user/repos

# Explore what's available
qurl --docs
qurl --docs /users
```

## Why Use This

- **No manual URLs**: `qurl /users/me` instead of `curl https://api.example.com/v2/users/me`
- **Auto-discovery**: Finds servers, parameters, auth requirements from the spec
- **Tab completion**: Press tab to see available paths, methods, and parameters
- **Built-in docs**: `--docs` shows you what each endpoint does

## Advanced Features

### AWS Lambda
```bash
export QURL_OPENAPI=lambda://my-function/openapi.json
qurl /path
qurl --sig-v4 https://api.aws.com/endpoint
```

### LLM Integration (MCP)
Let Claude or other LLMs use your API:
```bash
export QURL_OPENAPI=https://api.example.com/openapi.json
qurl --mcp

# Restrict LLM access to specific paths and methods
qurl --mcp -X GET -X POST /users/
# Allows /users/{id}, /users/{id}/profile, etc.
```

### Configuration
- `OPENAPI_URL` or `QURL_OPENAPI`: Your OpenAPI spec URL
- `QURL_SERVER`: Override server from spec
- Standard curl flags work: `-X`, `-H`, `-d`, `-v`