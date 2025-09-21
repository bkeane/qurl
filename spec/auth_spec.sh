#!/usr/bin/env sh

# auth_spec.sh - Authentication functionality tests
# Purpose: Test bearer token, AWS SigV4, and auth error handling

Describe "qurl authentication features"

    # Setup: Build qurl before running tests
    BeforeAll "go build -o ./qurl cmd/qurl/main.go"

    Describe "Bearer token authentication"

        It "adds Authorization header with bearer token"
            When call ./qurl --bearer test-token https://httpbin.dmuth.org/headers
            The output should include '"authorization": "Bearer test-token"'
            The stdout should be present
            The status should be success
        End

        It "shows bearer token in verbose output"
            When call ./qurl --bearer secret-token -v https://httpbin.dmuth.org/headers
            The stderr should include "Authorization: Bearer secret-token"
            The stdout should be present
            The status should be success
        End

    End

    Describe "AWS SigV4 authentication"

        It "works with --sig-v4 flag (no arguments, defaults to execute-api)"
            When call ./qurl --sig-v4 https://httpbin.dmuth.org/get
            The output should include "AWS4-HMAC-SHA256"
            The output should include "x-amz-date"
            The output should include 'execute-api'
            The stdout should be present
            The status should be success
        End

        It "works with custom service using --sig-v4-service"
            When call ./qurl --sig-v4 --sig-v4-service s3 https://httpbin.dmuth.org/get
            The output should include "AWS4-HMAC-SHA256"
            The output should include "x-amz-date"
            The output should include '/s3/'
            The stdout should be present
            The status should be success
        End

        It "does NOT sign requests without --sig-v4 flag"
            When call ./qurl https://httpbin.dmuth.org/get
            The output should not include "AWS4-HMAC-SHA256"
            The output should not include "x-amz-date"
            The stdout should be present
            The status should be success
        End

        It "shows SigV4 verbose info when enabled"
            When call ./qurl --sig-v4 -v https://httpbin.dmuth.org/get
            The stderr should include "Applied AWS SigV4 signature for service: execute-api"
            The stdout should be present
            The status should be success
        End

    End

    Describe "Lambda protocol handling"

        It "skips SigV4 for lambda:// URLs even with --sig-v4 flag"
            # This would test lambda:// protocol when implemented
            Skip "Lambda invocation not fully implemented yet"
        End

    End

    Describe "Authentication priority and combinations"

        It "bearer token works alongside custom headers"
            When call ./qurl --bearer token123 -H "X-Custom: value" https://httpbin.dmuth.org/headers
            The output should include '"authorization": "Bearer token123"'
            The output should include '"x-custom": "value"'
            The stdout should be present
            The status should be success
        End

        It "custom Authorization header overrides bearer token"
            When call ./qurl --bearer token123 -H "Authorization: Basic dXNlcjpwYXNz" https://httpbin.dmuth.org/headers
            The output should include '"authorization": "Basic dXNlcjpwYXNz"'
            The output should not include "Bearer token123"
            The stdout should be present
            The status should be success
        End

        It "custom headers can override SigV4 authorization"
            When call ./qurl --sig-v4 -H "Authorization: Custom override" https://httpbin.dmuth.org/headers
            The output should include '"authorization": "Custom override"'
            The output should not include "AWS4-HMAC-SHA256"
            The stdout should be present
            The status should be success
        End

        It "SigV4 and bearer together - SigV4 wins (applied last)"
            When call ./qurl --sig-v4 --bearer token123 https://httpbin.dmuth.org/get
            The output should include "AWS4-HMAC-SHA256"
            The output should not include "Bearer token123"
            The stdout should be present
            The status should be success
        End

    End

    Describe "Help and flag validation"

        It "shows help for sig-v4 flags"
            When call ./qurl --help
            The output should include "--sig-v4"
            The output should include "--sig-v4-service"
            The output should include "--bearer"
            The status should be success
        End

        It "shows default service in help"
            When call ./qurl --help
            The output should include "execute-api"
            The status should be success
        End

    End

End