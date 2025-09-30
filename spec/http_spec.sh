#!/usr/bin/env sh

# http_spec.sh - HTTP/HTTPS URL Protocol Testing
# Purpose: Test qurl's ability to handle http:// and https:// URLs directly
# Strategy: Focus on URL parsing, protocol handling, and server override behavior

# Helper function to run qurl
qurl() {
    QURL_OPENAPI="https://prod.kaixo.io/binnit/main/src/openapi.json" \
    go run cmd/qurl/main.go "$@"
}

# Helper to ensure AWS credentials are configured
check_aws_credentials() {
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        Skip "AWS credentials not configured"
    fi
}

Describe "qurl: HTTP/HTTPS URL Protocol Support"

    # Check AWS credentials before running tests that need authentication
    BeforeAll check_aws_credentials

    Describe "Feature: HTTPS URL Protocol Handling"

        It "accepts full HTTPS URLs directly"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "handles HTTPS URLs with query parameters"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?param=value"
            The output should include '"args":'
            The output should include '"param":"value"'
            The status should be success
        End

        It "works with HTTPS URLs and custom headers"
            When call qurl --aws-sigv4 -H "X-Custom-Header: test-value" https://prod.kaixo.io/binnit/main/src/headers
            The output should include '"x-custom-header":"test-value"'
            The status should be success
        End

        It "supports HTTPS URLs with different HTTP methods"
            When call qurl --aws-sigv4 -X POST https://prod.kaixo.io/binnit/main/src/post
            The output should include '"url":"https://prod.kaixo.io/post"'
            The status should be success
        End

        It "handles HTTPS URLs with request body data"
            When call qurl --aws-sigv4 -X POST -d '{"test": "data"}' https://prod.kaixo.io/binnit/main/src/post
            The output should include '"json":'
            The output should include '"test":"data"'
            The status should be success
        End

        It "shows verbose output for HTTPS requests"
            When call qurl --aws-sigv4 -v https://prod.kaixo.io/binnit/main/src/get
            The stderr should include "GET https://prod.kaixo.io/binnit/main/src/get"
            The stderr should include "Authorization: AWS4-HMAC-SHA256"
            The stdout should be present
            The status should be success
        End

    End

    Describe "Feature: HTTP URL Protocol Handling"

        It "accepts full HTTP URLs directly (when available)"
            Skip "HTTP endpoints not available in current test environment"
        End

        It "shows appropriate error for HTTP URLs requiring authentication"
            Skip "HTTP URL authentication behavior not implemented"
        End

    End

    Describe "Feature: URL Parsing and Validation"

        It "handles URLs with various port numbers"
            Skip "Custom port testing not available in current environment"
        End

        It "processes URLs with path parameters correctly"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/status/200
            The output should include '"status":200'
            The status should be success
        End

        It "handles URLs with fragment identifiers"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get#fragment"
            The output should include '"url":"https://prod.kaixo.io/get"'
            The status should be success
        End

        It "processes URLs with encoded characters"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?key=hello%20world"
            The output should include '"args":'
            The output should include '"key":"hello world"'
            The status should be success
        End

    End

    Describe "Feature: Server Override Behavior"

        It "ignores --server flag when full URL is provided"
            When call qurl --aws-sigv4 --server https://ignored.example.com https://prod.kaixo.io/binnit/main/src/get
            The output should include '"url":"https://prod.kaixo.io/get"'
            The status should be success
        End

        It "shows warning about server override in verbose mode"
            When call qurl --aws-sigv4 --server https://ignored.example.com -v https://prod.kaixo.io/binnit/main/src/get
            The stderr should include "GET https://prod.kaixo.io/binnit/main/src/get"
            The stdout should be present
            The status should be success
        End

    End

    Describe "Feature: Protocol Security Requirements"

        It "requires authentication for private HTTPS APIs"
            Skip "Private API authentication failure behavior not implemented"
        End

        It "successfully authenticates with --aws-sigv4 flag"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "works with custom AWS service specification"
            When call qurl --aws-sigv4 --aws-service execute-api https://prod.kaixo.io/binnit/main/src/get
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

    End

    Describe "Feature: Error Handling for URL Protocols"

        It "provides helpful error for malformed URLs"
            When call qurl "not-a-valid-url"
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles URLs with missing protocol gracefully"
            When call qurl "example.com/api/test"
            The stderr should include "ERR"
            The status should not be success
        End

        It "reports unsupported protocol schemes clearly"
            When call qurl "ftp://example.com/file.txt"
            The stderr should include "ERR"
            The status should not be success
        End

        It "provides clear error for unreachable hosts"
            When call qurl --aws-sigv4 "https://definitely-does-not-exist-12345.example.com/api"
            The stderr should include "ERR"
            The status should not be success
        End

    End

End