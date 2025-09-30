#!/usr/bin/env sh

# petstore_spec.sh - User Experience & Documentation
# Purpose: Demonstrate qurl's capabilities through clear, readable examples
# Strategy: Use Petstore API to showcase real-world API interaction workflows

# Helper function to run qurl with Petstore API
qurl() {
    QURL_OPENAPI="https://petstore3.swagger.io/api/v3/openapi.json" \
    go run cmd/qurl/main.go "$@"
}

Describe "qurl: User Experience and Documentation"

    Describe "Feature: API Discovery and Documentation"

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
            The output should include "GET"
            The output should include "/pet/{petId}"
            The status should be success
        End

        It "shows me what parameters an endpoint accepts"
            When call qurl --docs /pet/findByStatus
            The output should include "status"
            The output should include "Parameters"
            The output should include "Query parameters"
            The output should include "Status values that need to be considered"
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

        It "groups endpoints by resource type"
            When call qurl --docs
            The output should include "Endpoints"
            The output should include "/pet"
            The output should include "/store"
            The output should include "/user"
            The status should be success
        End

    End

    Describe "Feature: Interactive Help System"

        # Helper for tests without OpenAPI
        qurl_direct() {
            go run cmd/qurl/main.go "$@"
        }

        It "shows comprehensive help with --help flag"
            When call qurl_direct --help
            The output should include "OpenAPI v3 REST client"
            The output should include "--docs"
            The output should include "--verbose"
            The output should include "--header"
            The status should be success
        End

        It "displays authentication options in help"
            When call qurl_direct --help
            The output should include "--aws-sigv4"
            The output should include "--aws-service"
            The output should include "--header"
            The status should be success
        End

        It "shows environment variable documentation"
            When call qurl_direct --help
            The output should include "QURL_OPENAPI"
            The output should include "QURL_SERVER"
            The status should be success
        End

        It "provides clear usage examples"
            When call qurl_direct --help
            The output should include "qurl [path]"
            The status should be success
        End

    End

    Describe "Feature: Real-world API Workflows"

        It "discovers Pet Store API structure"
            When call qurl --docs
            The output should include "Swagger Petstore"
            The output should include "Everything about your Pets"
            The status should be success
        End

        It "looks up specific resources with path parameters"
            When call qurl /pet/10
            The output should include '"id":10'
            The status should be success
        End

        It "checks store inventory with GET requests"
            When call qurl /store/inventory
            The output should include "{"
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

    Describe "Feature: Documentation Output Formatting"

        It "formats endpoint documentation clearly"
            When call qurl --docs /pet/{petId}
            The output should include "GET"
            The output should include "/pet/{petId}"
            The output should include "Find pet by ID"
            The status should be success
        End

        It "shows HTTP methods prominently"
            When call qurl --docs /pet/findByStatus
            The output should include "GET"
            The status should be success
        End

        It "displays parameter types and requirements"
            When call qurl --docs /pet/findByStatus
            The output should include "Query parameters"
            The output should include "status"
            The status should be success
        End

        It "shows server information from OpenAPI spec"
            When call qurl --docs
            The output should include "Servers"
            The output should include "/api/v3"
            The status should be success
        End

    End

    Describe "Feature: User-Friendly Error Handling"

        It "handles non-existent endpoints gracefully"
            When call qurl /this/does/not/exist
            The stdout should be present
            The status should be success
        End

        It "reports when no API documentation is found"
            When call qurl --docs /nonexistent/endpoint
            The output should include "No endpoints found"
            The status should be success
        End

        It "handles 404 responses appropriately"
            When call qurl /pet/999999999
            The stdout should be present
            The status should be success
        End

        It "provides helpful error messages for missing server/OpenAPI configuration"
            When call sh -c 'unset QURL_OPENAPI && unset OPENAPI_URL && go run cmd/qurl/main.go /relative/path'
            The status should not be success
            The stderr should include "ERR"
        End

    End

    Describe "Feature: Environment Variable Configuration"

        It "reads OpenAPI URL from environment variable"
            When call sh -c 'QURL_OPENAPI="https://petstore3.swagger.io/api/v3/openapi.json" go run cmd/qurl/main.go /store/inventory'
            The output should include "{"
            The status should be success
        End

        It "allows flag to override environment variable"
            When call sh -c 'QURL_OPENAPI="https://wrong.url" go run cmd/qurl/main.go --openapi "https://petstore3.swagger.io/api/v3/openapi.json" /store/inventory'
            The output should include "{"
            The status should be success
        End

        It "supports QURL_SERVER environment variable"
            When call sh -c 'QURL_SERVER="https://petstore3.swagger.io/api/v3" go run cmd/qurl/main.go /store/inventory'
            The output should include "{"
            The status should be success
        End

    End

    Describe "Feature: Shell Completion Support"

        It "generates bash completion scripts"
            When call go run cmd/qurl/main.go completion bash
            The output should include "# bash completion"
            The status should be success
        End

        It "provides completions for different shells"
            When call go run cmd/qurl/main.go completion zsh
            The output should include "#compdef"
            The status should be success
        End

        It "supports fish shell completion"
            When call go run cmd/qurl/main.go completion fish
            The output should include "complete"
            The status should be success
        End

        It "supports powershell completion"
            When call go run cmd/qurl/main.go completion powershell
            The output should include "Register-ArgumentCompleter"
            The status should be success
        End

    End

End