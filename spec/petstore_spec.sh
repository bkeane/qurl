#!/usr/bin/env sh

# petstore_spec.sh - User-Facing Feature Documentation
# Purpose: Demonstrate qurl's capabilities through clear, readable examples
# Strategy: Use Petstore API to showcase real-world API interaction workflows

# Helper function to run qurl with Petstore API
qurl() {
    OPENAPI_URL="https://petstore3.swagger.io/api/v3/openapi.json" \
    go run cmd/qurl/main.go "$@"
}

Describe "qurl: A smart HTTP client for OpenAPI"

    # Setup: Build qurl before running tests
    BeforeAll "go build -o ./qurl cmd/qurl/main.go"

    Describe "Feature: Exploring APIs with documentation"

        It "shows me all available API endpoints"
            When call qurl --docs
            The output should include "Swagger Petstore"
            The output should include "/pet"
            The output should include "/store"
            The output should include "/user"
            The status should be success
        End

        It "lets me explore a specific endpoint's documentation"
            When call qurl --docs /pet/{petId}
            The output should include "Find pet by ID"
            The output should include "Parameters"
            The output should include "petId"
            The status should be success
        End

        It "shows me what parameters an endpoint accepts"
            When call qurl --docs /pet/findByStatus
            The output should include "status"
            The output should include "available"
            The output should include "pending"
            The output should include "sold"
            The status should be success
        End

        It "displays request body requirements for POST endpoints"
            When call qurl --docs /pet -X POST
            The output should include "Request Body"
            The output should include "application/json"
            The output should include "name"
            The status should be success
        End

        It "shows expected response formats"
            When call qurl --docs /store/inventory
            The output should include "200"
            The output should include "application/json"
            The status should be success
        End

    End

    Describe "Feature: Making API calls with OpenAPI awareness"

        It "automatically uses the correct base URL from the spec"
            When call qurl -v /pet/findByStatus --param status=available
            The stderr should include "https://petstore3.swagger.io/api/v3/pet/findByStatus"
            The output should include '"id"'
            The status should be success
        End

        It "sets appropriate Accept headers based on the API spec"
            When call qurl -v /store/inventory
            The stderr should include "> Accept: application/json"
            The output should include "{"
            The status should be success
        End

        It "works with path parameters"
            # Pet ID 10 is a test pet that should exist
            When call qurl /pet/10
            The output should include '"id": 10'
            The status should be success
        End

        It "handles query parameters elegantly"
            When call qurl /pet/findByStatus --param status=available
            The output should include '['
            The output should include '"name"'
            The status should be success
        End

        It "supports multiple values for the same parameter"
            When call qurl /pet/findByStatus --param status=available --param status=pending
            The output should include '['
            The status should be success
        End

    End

    Describe "Feature: Different HTTP methods"

        It "performs GET requests by default"
            When call qurl /store/inventory
            The output should include "{"
            The status should be success
        End

        It "can make HEAD requests to check endpoints"
            When call qurl -X HEAD -v /store/inventory
            The stderr should include "> HEAD"
            The stderr should include "< HTTP/"
            The output should be blank
            The status should be success
        End

        It "supports OPTIONS for checking allowed methods"
            When call qurl -X OPTIONS -v /pet/10
            The stderr should include "> OPTIONS"
            The status should be success
        End

        It "handles POST requests (when implemented)"
            Skip "Request body support not yet implemented"
            When call qurl -X POST /pet -d '{"name":"fluffy","status":"available"}'
            The output should include '"name": "fluffy"'
            The status should be success
        End

    End

    Describe "Feature: Verbose mode for debugging"

        It "shows the full request being made"
            When call qurl -v /store/inventory 2>&1
            The output should include "> GET https://petstore3.swagger.io/api/v3/store/inventory"
            The output should include "> Host: petstore3.swagger.io"
            The output should include "> User-Agent: qurl"
            The output should include "> Accept: application/json"
            The status should be success
        End

        It "displays response headers and status"
            When call qurl -v /store/inventory 2>&1
            The output should include "< HTTP/"
            The output should include "< Content-Type: application/json"
            The status should be success
        End

        It "keeps verbose output separate from response data"
            When call qurl -v /store/inventory
            The stderr should include "> GET"
            The stdout should include "{"
            The stdout should not include "> GET"
            The status should be success
        End

    End

    Describe "Feature: Response formatting options"

        It "can include response headers in output"
            When call qurl -i /store/inventory
            The output should include "HTTP/"
            The output should include "Content-Type: application/json"
            The output should include "{"
            The status should be success
        End

        It "handles different response content types"
            When call qurl /pet/10
            The output should include '"id"'
            The output should include '"name"'
            The status should be success
        End

        It "properly displays array responses"
            When call qurl /pet/findByStatus --param status=sold
            The output should include "["
            The status should be success
        End

    End

    Describe "Feature: Error handling and edge cases"

        It "handles non-existent endpoints gracefully"
            When call qurl /this/does/not/exist
            The status should be success
        End

        It "reports when no API documentation is found"
            When call qurl --docs /nonexistent/endpoint
            The output should include "No endpoints found"
            The status should be success
        End

        It "handles 404 responses appropriately"
            When call qurl /pet/999999999
            The status should be success
        End

        It "works with empty query results"
            When call qurl /pet/findByStatus --param status=nonexistent
            The output should include "[]"
            The status should be success
        End

    End

    Describe "Feature: Custom headers and overrides"

        It "allows custom headers to be added"
            When call qurl -v /pet/10 -H "X-Custom-Header: test-value" 2>&1
            The output should include "> X-Custom-Header: test-value"
            The status should be success
        End

        It "lets users override default headers"
            When call qurl -v /pet/10 -H "User-Agent: my-custom-agent" 2>&1
            The output should include "> User-Agent: my-custom-agent"
            The output should not include "> User-Agent: qurl"
            The status should be success
        End

        It "can override OpenAPI-derived Accept headers"
            When call qurl -v /pet/10 -H "Accept: text/plain" 2>&1
            The output should include "> Accept: text/plain"
            The output should not include "> Accept: application/json"
            The status should be success
        End

    End

    Describe "Feature: Working without OpenAPI"

        # Helper for tests without OpenAPI
        qurl_direct() {
            go run cmd/qurl/main.go "$@"
        }

        It "can make direct HTTP requests to any URL"
            When call qurl_direct https://petstore3.swagger.io/api/v3/store/inventory
            The output should include "{"
            The status should be success
        End

        It "works like curl for simple requests"
            When call qurl_direct -X GET https://petstore3.swagger.io/api/v3/pet/10
            The output should include '"id"'
            The status should be success
        End

        It "requires full URLs when no OpenAPI spec is available"
            When call qurl_direct /pet/10
            The status should not be success
            The stderr should include "Error"
        End

    End

    Describe "Feature: Shell completion support"

        It "can generate bash completion scripts"
            When call qurl completion bash
            The output should include "# bash completion"
            The status should be success
        End

        It "provides completions for different shells"
            When call qurl completion zsh
            The output should include "#compdef"
            The status should be success
        End

        It "autocompletes API paths (when integrated)"
            Skip "Testing shell completion requires shell integration"
            # This would be tested in an integrated shell environment
        End

        It "autocompletes query parameters (when integrated)"
            Skip "Testing shell completion requires shell integration"
            # This would be tested in an integrated shell environment
        End

    End

    Describe "Feature: Environment variable configuration"

        # Test with explicit environment variable
        It "reads OpenAPI URL from environment variable"
            When call sh -c 'OPENAPI_URL="https://petstore3.swagger.io/api/v3/openapi.json" go run cmd/qurl/main.go /store/inventory'
            The output should include "{"
            The status should be success
        End

        It "allows flag to override environment variable"
            When call sh -c 'OPENAPI_URL="https://wrong.url" go run cmd/qurl/main.go --openapi "https://petstore3.swagger.io/api/v3/openapi.json" /store/inventory'
            The output should include "{"
            The status should be success
        End

    End

    Describe "Feature: Practical API workflows"

        It "helps me find available pets"
            When call qurl /pet/findByStatus --param status=available
            The output should include '"status": "available"'
            The status should be success
        End

        It "lets me check store inventory"
            When call qurl /store/inventory
            The output should include "{"
            The status should be success
        End

        It "allows me to look up specific resources"
            When call qurl /pet/10
            The output should include '"id": 10'
            The status should be success
        End

        It "shows me API documentation when I need it"
            When call qurl --docs /user/login
            The output should include "Logs user into the system"
            The output should include "username"
            The output should include "password"
            The status should be success
        End

    End

End