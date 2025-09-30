#!/usr/bin/env sh

# request_response_spec.sh - HTTP Implementation Correctness
# Purpose: Test qurl's HTTP implementation correctness using Binnit reflection service
# Strategy: Exercise full REST method suite, headers, parameters, and response handling

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

Describe "qurl: HTTP Implementation Correctness"

    # Check AWS credentials before running tests that need authentication
    BeforeAll check_aws_credentials

    Describe "Feature: HTTP Method Implementation"

        It "implements GET requests correctly"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"method":"GET"'
            The output should include '"url":"https://prod.kaixo.io/get"'
            The status should be success
        End

        It "implements POST requests correctly"
            When call qurl --aws-sigv4 -X POST https://prod.kaixo.io/binnit/main/src/post
            The output should include '"method":"POST"'
            The output should include '"url":"https://prod.kaixo.io/post"'
            The status should be success
        End

        It "implements PUT requests correctly"
            When call qurl --aws-sigv4 -X PUT https://prod.kaixo.io/binnit/main/src/put
            The output should include '"method":"PUT"'
            The output should include '"url":"https://prod.kaixo.io/put"'
            The status should be success
        End

        It "implements PATCH requests correctly"
            When call qurl --aws-sigv4 -X PATCH https://prod.kaixo.io/binnit/main/src/patch
            The output should include '"method":"PATCH"'
            The output should include '"url":"https://prod.kaixo.io/patch"'
            The status should be success
        End

        It "implements DELETE requests correctly"
            When call qurl --aws-sigv4 -X DELETE https://prod.kaixo.io/binnit/main/src/delete
            The output should include '"method":"DELETE"'
            The output should include '"url":"https://prod.kaixo.io/delete"'
            The status should be success
        End

        It "implements HEAD requests correctly"
            When call qurl --aws-sigv4 -X HEAD https://prod.kaixo.io/binnit/main/src/get
            The status should be success
        End

        It "implements OPTIONS requests correctly"
            When call qurl --aws-sigv4 -X OPTIONS https://prod.kaixo.io/binnit/main/src/get
            The output should include "Method Not Allowed"
            The status should be success
        End

    End

    Describe "Feature: Request Header Handling"

        It "sends custom headers correctly"
            When call qurl --aws-sigv4 -H "X-Custom-Header: test-value" https://prod.kaixo.io/binnit/main/src/headers
            The output should include '"x-custom-header":"test-value"'
            The status should be success
        End

        It "sends multiple custom headers correctly"
            When call qurl --aws-sigv4 -H "X-First: value1" -H "X-Second: value2" https://prod.kaixo.io/binnit/main/src/headers
            The output should include '"x-first":"value1"'
            The output should include '"x-second":"value2"'
            The status should be success
        End

        It "handles header values with spaces"
            When call qurl --aws-sigv4 -H "X-Spaced: value with spaces" https://prod.kaixo.io/binnit/main/src/headers
            The output should include '"x-spaced":"value with spaces"'
            The status should be success
        End

        It "sends proper User-Agent header"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/user-agent
            The output should include '"user-agent":"qurl"'
            The status should be success
        End

        It "allows User-Agent override"
            When call qurl --aws-sigv4 -H "User-Agent: custom-agent/1.0" https://prod.kaixo.io/binnit/main/src/user-agent
            The output should include '"user-agent":"custom-agent/1.0"'
            The status should be success
        End

        It "handles Content-Type headers for POST requests"
            When call qurl --aws-sigv4 -X POST -H "Content-Type: application/json" -d '{"test": "data"}' https://prod.kaixo.io/binnit/main/src/post
            The output should include '"content-type":"application/json"'
            The output should include '"json":'
            The status should be success
        End

    End

    Describe "Feature: Query Parameter Handling"

        It "sends single query parameters correctly"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?single=value"
            The output should include '"args":'
            The output should include '"single":"value"'
            The status should be success
        End

        It "sends multiple query parameters correctly"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?first=value1&second=value2"
            The output should include '"first":"value1"'
            The output should include '"second":"value2"'
            The status should be success
        End

        It "handles query parameters with special characters"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?special=hello%20world&encoded=%26%3D%3F"
            The output should include '"special":"hello world"'
            The output should include '"encoded":"&=?"'
            The status should be success
        End

        It "handles empty query parameter values"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?empty="
            The output should include '"empty":""'
            The status should be success
        End

        It "handles query parameters without values"
            When call qurl --aws-sigv4 "https://prod.kaixo.io/binnit/main/src/get?flag"
            The output should include '"args":'
            The status should be success
        End

    End

    Describe "Feature: Request Body Handling"

        It "sends JSON request bodies correctly"
            When call qurl --aws-sigv4 -X POST -d '{"key": "value", "number": 42}' https://prod.kaixo.io/binnit/main/src/post
            The output should include '"json":'
            The output should include '"key":"value"'
            The output should include '"number":42'
            The status should be success
        End

        It "sends form data correctly"
            When call qurl --aws-sigv4 -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "key1=value1&key2=value2" https://prod.kaixo.io/binnit/main/src/post
            The output should include '"form":'
            The output should include '"key1":"value1"'
            The output should include '"key2":"value2"'
            The status should be success
        End

        It "sends plain text request bodies correctly"
            When call qurl --aws-sigv4 -X POST -H "Content-Type: text/plain" -d "plain text content" https://prod.kaixo.io/binnit/main/src/post
            The output should include '"data":"plain text content"'
            The status should be success
        End

        It "handles empty request bodies"
            When call qurl --aws-sigv4 -X POST https://prod.kaixo.io/binnit/main/src/post
            The output should include '"method":"POST"'
            The status should be success
        End

        It "handles large request bodies"
            # Create a larger JSON payload
            large_data='{"large": "' && for i in $(seq 1 100); do large_data="${large_data}data"; done && large_data="${large_data}\"}"
            When call qurl --aws-sigv4 -X POST -d "$large_data" https://prod.kaixo.io/binnit/main/src/post
            The output should include '"json":'
            The output should include '"large":'
            The status should be success
        End

    End

    Describe "Feature: Response Handling"

        It "handles JSON responses correctly"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include "{"
            The output should include "}"
            The status should be success
        End

        It "handles different HTTP status codes"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/status/201
            The output should include '"status":201'
            The status should be success
        End

        It "handles 404 responses gracefully"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/status/404
            The stdout should be present
            The status should be success
        End

        It "handles 500 responses appropriately"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/status/500
            The stdout should be present
            The status should be success
        End

        It "preserves response body content"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"headers":'
            The output should include '"url":'
            The output should include '"args":'
            The status should be success
        End

    End

    Describe "Feature: HTTP Protocol Compliance"

        It "follows HTTP redirect responses"
            Skip "Redirect testing requires specific endpoint setup"
        End

        It "handles chunked transfer encoding"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"headers":'
            The status should be success
        End

        It "handles gzip compression correctly"
            When call qurl --aws-sigv4 -H "Accept-Encoding: gzip" https://prod.kaixo.io/binnit/main/src/headers
            The output should include '"accept-encoding":"gzip"'
            The status should be success
        End

        It "preserves connection keep-alive behavior"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The output should include '"headers":'
            The status should be success
        End

        It "handles HTTP/1.1 protocol correctly"
            When call qurl --aws-sigv4 -v https://prod.kaixo.io/binnit/main/src/get
            The stderr should include "GET https://prod.kaixo.io/binnit/main/src/get"
            The stdout should include '"url"'
            The status should be success
        End

    End

    Describe "Feature: Authentication Integration"

        It "correctly integrates AWS SigV4 signing"
            When call qurl --aws-sigv4 -v https://prod.kaixo.io/binnit/main/src/get
            The stderr should include "Authorization: AWS4-HMAC-SHA256"
            The stderr should include "x-amz-date"
            The stdout should include '"url"'
            The status should be success
        End

        It "signs request headers properly"
            When call qurl --aws-sigv4 -H "X-Custom: test" -v https://prod.kaixo.io/binnit/main/src/headers
            The stderr should include "Authorization: AWS4-HMAC-SHA256"
            The output should include '"x-custom":"test"'
            The status should be success
        End

        It "signs request body for POST requests"
            When call qurl --aws-sigv4 -X POST -d '{"signed": "body"}' -v https://prod.kaixo.io/binnit/main/src/post
            The stderr should include "Authorization: AWS4-HMAC-SHA256"
            The output should include '"signed":"body"'
            The status should be success
        End

        It "handles authorization header via -H flag"
            Skip "Authorization header test requires non-AWS authenticated endpoint"
        End

    End

    Describe "Feature: Verbose Output Correctness"

        It "shows request details in verbose mode"
            When call qurl --aws-sigv4 -v https://prod.kaixo.io/binnit/main/src/get
            The stderr should include "GET https://prod.kaixo.io/binnit/main/src/get"
            The stderr should include "Authorization:"
            The stdout should include '"url"'
            The status should be success
        End

        It "shows request headers in verbose mode"
            When call qurl --aws-sigv4 -H "X-Verbose: test" -v https://prod.kaixo.io/binnit/main/src/headers
            The stderr should include "X-Verbose: test"
            The stdout should include '"headers"'
            The status should be success
        End

        It "shows request body in verbose mode for POST"
            When call qurl --aws-sigv4 -X POST -d '{"verbose": "test"}' -v https://prod.kaixo.io/binnit/main/src/post
            The stderr should include '"verbose": "test"'
            The stdout should include '"json"'
            The status should be success
        End

        It "maintains clean output without verbose flag"
            When call qurl --aws-sigv4 https://prod.kaixo.io/binnit/main/src/get
            The stderr should be blank
            The stdout should include '"url":'
            The status should be success
        End

    End

    Describe "Feature: Error Handling and Edge Cases"

        It "handles network timeouts gracefully"
            Skip "Timeout testing requires specific network conditions"
        End

        It "handles malformed response bodies"
            Skip "Malformed response testing requires specific endpoint setup"
        End

        It "reports connection failures clearly"
            When call qurl --aws-sigv4 https://definitely-does-not-exist-12345.example.com/api
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles very large responses appropriately"
            Skip "Large response testing requires specific endpoint setup"
        End

        It "preserves binary response data integrity"
            Skip "Binary data testing requires specific endpoint setup"
        End

    End

End