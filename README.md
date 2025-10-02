# qurl

OpenAPI v3 REST client and MCP server.

## Install

**Linux/macOS:**
```bash
curl -sS https://raw.githubusercontent.com/bkeane/qurl/main/install.sh | bash
```

**Windows:**
```powershell
iwr -useb https://raw.githubusercontent.com/bkeane/qurl/main/install.ps1 | iex
```

**Manual:** Download from [releases](https://github.com/bkeane/qurl/releases).

## Quick Start

```bash
# Try it with the Swagger Petstore
export QURL_OPENAPI=https://petstore3.swagger.io/api/v3/openapi.json

qurl --docs
```

## 🔍 Explore

Use `--docs` to browse your API. The same options that configure requests also act as documentation filters:

```bash
qurl --docs                  # All endpoints
qurl --docs /pet/            # Endpoints under /pet
qurl --docs -X GET /pet      # Method documentation
qurl --docs -X GET -X DELETE # All GET and DELETE endpoints
qurl --docs -X POST /pet/    # POST endpoints under /pet
```

Tab completion knows your API:
```bash
qurl <TAB>                              # Complete paths: /pet, /store, /user
qurl -X <TAB>                           # Complete methods: GET, POST, PUT, DELETE
qurl --server <TAB>                     # Complete servers from OpenAPI spec
qurl /pet/findByStatus --query sta<TAB> # Complete params: status=
```

## 🚀 Execute

Make requests with curl-like syntax:

```bash
# Using QURL_OPENAPI environment variable (recommended)
export QURL_OPENAPI=https://petstore3.swagger.io/api/v3/openapi.json
qurl /pet/findByStatus --query status=available # GET with query param
qurl -X DELETE /pet/123                         # Delete pet by ID
qurl -v /store/inventory                        # Verbose output

# Direct URL (old fashioned way)
qurl https://api.example.com/users              # GET request
qurl -X POST https://api.example.com/users      # POST request
```

## 🔐 AWS

Native AWS service integration:

```bash
# HTTP compatible Lambda function invocation
qurl lambda://my-function/users

# HTTP compatible Lambda function ivocation via OpenAPI spec
export QURL_OPENAPI=lambda://my-function/openapi.json
qurl --docs

# API Gateway with SigV4
qurl --aws-sigv4 /users

# Any AWS service with SigV4
AWS_REGION=us-east-1 qurl --aws-sigv4 --aws-service sts \
  -X POST -d "Action=GetCallerIdentity&Version=2011-06-15" \
  https://sts.amazonaws.com/
```

## 🤖 MCP

Start an MCP server for LLM integration. Request filters act as safety constraints:

```bash
qurl --mcp                                   # Full API access
qurl --mcp -X GET                            # Read-only access
qurl --mcp /pet/                             # Only /pet endpoints
qurl --mcp -X GET -X POST /pet               # Only GET/POST on /pet
qurl --mcp -H "Authorization: Bearer $TOKEN" # Include header in all requests
```

Use with Claude Desktop, Cline, or any MCP client.

```json
{
   "mcpServers":{
      "petstore":{
         "command":"qurl",
         "args":[
            "--mcp",
            "-X", "GET",
            "-H", "Authorization: Bearer $TOKEN"
         ],
         "env":{
            "QURL_OPENAPI":"https://petstore3.swagger.io/api/v3/openapi.json",
            "QURL_MCP_DESCRIPTION":"Petstore API - manage pets, orders, and users in the pet store"
         }
      }
   }
}
```

## Configuration

```bash
# OpenAPI and Server
export QURL_OPENAPI=https://api.example.com/openapi.yaml # OpenAPI spec URL
export QURL_SERVER=https://staging.api.com               # Override server URL

# Logging
export QURL_LOG_LEVEL=debug                              # Log verbosity (debug, info, warn, error)
export QURL_LOG_FORMAT=json                              # Log format (json, pretty)

# MCP
export QURL_MCP_DESCRIPTION="API description and purpose" # Help LLM understand when to use this API
```