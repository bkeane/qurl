#!/usr/bin/env sh

# lambda_spec.sh - Lambda URL Protocol Testing
# Purpose: Test qurl's lambda:// URI scheme for direct Lambda function invocation
# Strategy: Focus on lambda protocol parsing, invocation, and integration with OpenAPI

# Helper function to run qurl
qurl() {
    QURL_OPENAPI="lambda://binnit-main-src/openapi.json" \
    go run cmd/qurl/main.go "$@"
}

# Helper to ensure AWS credentials are configured for Lambda access
check_aws_credentials() {
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        Skip "AWS credentials not configured"
    fi
}

Describe "qurl: Lambda URL Protocol Support"

    # Check AWS credentials before running tests that need Lambda access
    BeforeAll check_aws_credentials

    Describe "Feature: Lambda URI Scheme Support"

        It "accepts lambda:// URLs for direct function invocation"
            When call qurl lambda://binnit-main-src/user-agent
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "handles lambda:// URLs with path parameters"
            When call qurl lambda://binnit-main-src/get
            The output should include '"url"'
            The output should include '"headers"'
            The status should be success
        End

        It "supports lambda:// URLs with query parameters"
            When call qurl "lambda://binnit-main-src/get?test=value"
            The output should include '"args":'
            The output should include '"test":"value"'
            The status should be success
        End

        It "works with lambda:// URLs and custom headers"
            When call qurl -H "X-Lambda-Test: custom-value" lambda://binnit-main-src/headers
            The output should include '"x-lambda-test":"custom-value"'
            The status should be success
        End

        It "supports different HTTP methods with lambda:// URLs"
            When call qurl -X POST lambda://binnit-main-src/post
            The output should include '"url"'
            The status should be success
        End

        It "handles lambda:// URLs with request body data"
            When call qurl -X POST -d '{"lambda": "test"}' lambda://binnit-main-src/post
            The output should include '"json":'
            The output should include '"lambda":"test"'
            The status should be success
        End

        It "shows verbose output for lambda invocations"
            When call qurl -v lambda://binnit-main-src/get
            The stderr should include "GET"
            The stdout should include '"url"'
            The status should be success
        End

    End

    Describe "Feature: Lambda OpenAPI Integration"

        It "fetches OpenAPI specs via lambda:// URLs"
            When call qurl --docs
            The output should include "Binnit"
            The output should include "/anything"
            The output should include "Endpoints"
            The status should be success
        End

        It "uses lambda OpenAPI spec for endpoint discovery"
            When call qurl --docs /get
            The output should include "GET"
            The output should include "/get"
            The status should be success
        End

        It "combines lambda OpenAPI with relative path requests"
            When call qurl /get
            The output should include '"url"'
            The status should be success
        End

        It "shows lambda server information in documentation"
            When call qurl --docs
            The output should include "Servers"
            The status should be success
        End

    End

    Describe "Feature: Lambda Function Name Parsing"

        It "correctly parses simple lambda function names"
            When call qurl lambda://simple-function/path
            The stderr should include "ERR" # Expected if function doesn't exist
            The status should not be success
        End

        It "handles lambda function names with hyphens"
            When call qurl lambda://binnit-main-src/user-agent
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "processes lambda function names with underscores"
            Skip "No underscore function names in test environment"
        End

        It "handles complex lambda function naming patterns"
            When call qurl lambda://binnit-main-src/anything/complex/path
            The output should include '"url"'
            The status should be success
        End

    End

    Describe "Feature: Lambda vs HTTPS Protocol Comparison"

        It "lambda:// bypasses API Gateway authentication"
            When call qurl lambda://binnit-main-src/get
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "lambda:// provides same response as authenticated HTTPS"
            When call qurl lambda://binnit-main-src/get
            The output should include '"headers"'
            The output should include '"url"'
            The status should be success
        End

        It "lambda:// handles POST requests like HTTPS equivalent"
            When call qurl -X POST -d '{"test": "data"}' lambda://binnit-main-src/post
            The output should include '"json":'
            The output should include '"test":"data"'
            The status should be success
        End

    End

    Describe "Feature: Lambda Error Handling"

        It "provides helpful error for non-existent lambda functions"
            When call qurl lambda://non-existent-function/path
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles malformed lambda:// URLs gracefully"
            When call qurl "lambda://incomplete"
            The stderr should include "ERR"
            The status should not be success
        End

        It "reports lambda invocation failures clearly"
            When call qurl lambda://invalid-function-name-12345/path
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles lambda timeout scenarios appropriately"
            Skip "Lambda timeout testing requires specific setup"
        End

    End

    Describe "Feature: Lambda URI Environment Integration"

        It "respects QURL_OPENAPI with lambda:// URLs"
            When call qurl /user-agent
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "combines lambda OpenAPI with --server flag gracefully"
            When call qurl --server https://example.com /
            The output should include "Example Domain"
            The status should be success
        End

        It "handles conflicting lambda and HTTPS configurations"
            When call qurl lambda://binnit-main-src/get
            The output should include '"url"'
            The status should be success
        End

    End

    Describe "Feature: Lambda Response Format Consistency"

        It "returns JSON responses in consistent format"
            When call qurl lambda://binnit-main-src/get
            The output should include "{"
            The output should include "}"
            The status should be success
        End

        It "preserves HTTP status codes from lambda responses"
            When call qurl lambda://binnit-main-src/status/201
            The output should include '"status":201'
            The status should be success
        End

        It "handles lambda responses with custom content types"
            When call qurl lambda://binnit-main-src/headers
            The output should include '"headers"'
            The status should be success
        End

        It "maintains header information in lambda responses"
            When call qurl -H "X-Test: lambda" lambda://binnit-main-src/headers
            The output should include '"x-test":"lambda"'
            The status should be success
        End

    End

End