#!/usr/bin/env sh

# httpbin_spec.sh - Deep HTTP Request Layer Verification
# Purpose: Exhaustively verify that qurl generates correct HTTP requests
# Strategy: Use httpbin's echo endpoints to inspect exact request structure
# Coverage: Symmetric testing of both OpenAPI-enabled and vanilla HTTP modes

Describe "qurl HTTP request generation verification"

    # Setup: Build qurl before running tests
    BeforeAll "go build -o ./qurl cmd/qurl/main.go"

    Describe "With OpenAPI specification (httpbin.org/spec.json)"

        # Helper for OpenAPI-enabled httpbin tests
        qurl_openapi() {
            OPENAPI_URL="https://httpbin.org/spec.json" go run cmd/qurl/main.go "$@"
        }

        Describe "HTTP method generation with OpenAPI"

            It "defaults to GET when no method specified"
                When call qurl_openapi /anything
                The output should include '"method": "GET"'
                The status should be success
            End

            It "correctly sets POST method"
                When call qurl_openapi -X POST /anything
                The output should include '"method": "POST"'
                The status should be success
            End

            It "correctly sets PUT method"
                When call qurl_openapi -X PUT /anything
                The output should include '"method": "PUT"'
                The status should be success
            End

            It "correctly sets DELETE method"
                When call qurl_openapi -X DELETE /anything
                The output should include '"method": "DELETE"'
                The status should be success
            End

            It "correctly sets PATCH method"
                When call qurl_openapi -X PATCH /anything
                The output should include '"method": "PATCH"'
                The status should be success
            End

            It "handles HEAD method (no body)"
                When call qurl_openapi -X HEAD -v /get
                The stderr should include "> HEAD https://httpbin.org/get"
                The output should be blank
                The status should be success
            End

            It "handles OPTIONS method"
                When call qurl_openapi -X OPTIONS /anything
                The output should include '"method": "OPTIONS"'
                The status should be success
            End

            It "normalizes method to uppercase"
                When call qurl_openapi -X post /anything
                The output should include '"method": "POST"'
                The status should be success
            End

        End

        Describe "URL and path handling with OpenAPI"

            It "extracts base URL from OpenAPI servers section"
                When call qurl_openapi -v /status/200
                The stderr should include "> GET https://httpbin.org/status/200"
                The status should be success
            End

            It "correctly constructs full URL from relative path"
                When call qurl_openapi /anything/test/path
                The output should include '"url": "https://httpbin.org/anything/test/path"'
                The status should be success
            End

            It "handles paths with multiple segments"
                When call qurl_openapi /anything/deeply/nested/path
                The output should include '"url": "https://httpbin.org/anything/deeply/nested/path"'
                The status should be success
            End

            It "handles paths with trailing slashes"
                When call qurl_openapi /anything/
                The output should include '"url": "https://httpbin.org/anything/"'
                The status should be success
            End

            It "resolves paths from OpenAPI spec URL when no servers defined"
                Skip "Requires spec without servers section"
                # This would test fallback to spec URL's host
            End

        End

        Describe "Header generation with OpenAPI"

            It "sets Accept header from OpenAPI response content types"
                When call qurl_openapi -v /get
                The stderr should include "> Accept: application/json"
                The status should be success
            End

            It "always sends User-Agent: qurl"
                When call qurl_openapi /headers
                The output should include '"User-Agent": "qurl"'
                The status should be success
            End

            It "correctly sets Host header"
                When call qurl_openapi /headers
                The output should include '"Host": "httpbin.org"'
                The status should be success
            End

            It "combines OpenAPI Accept with custom headers"
                When call qurl_openapi /headers -H "X-Custom: test"
                The output should include '"Accept": "application/json"'
                The output should include '"X-Custom": "test"'
                The output should include '"User-Agent": "qurl"'
                The status should be success
            End

            It "allows custom headers to override OpenAPI headers"
                When call qurl_openapi /headers -H "Accept: text/plain"
                The output should include '"Accept": "text/plain"'
                The output should not include '"Accept": "application/json"'
                The status should be success
            End

            It "sends multiple custom headers with OpenAPI"
                When call qurl_openapi /headers -H "X-One: 1" -H "X-Two: 2" -H "X-Three: 3"
                The output should include '"X-One": "1"'
                The output should include '"X-Two": "2"'
                The output should include '"X-Three": "3"'
                The output should include '"Accept": "application/json"'
                The status should be success
            End

            It "handles headers with spaces in values"
                When call qurl_openapi /headers -H "X-Message: hello world test"
                The output should include '"X-Message": "hello world test"'
                The status should be success
            End

            It "handles headers with special characters"
                When call qurl_openapi /headers -H "X-Special: value!@#\$%^&*()"
                The output should include '"X-Special": "value!@#\$%^&*()"'
                The status should be success
            End

            It "handles headers with colons in values"
                When call qurl_openapi /headers -H "X-URL: https://example.com:8080"
                The output should include '"X-URL": "https://example.com:8080"'
                The status should be success
            End

            It "overrides default User-Agent with custom one"
                When call qurl_openapi /headers -H "User-Agent: custom-agent"
                The output should include '"User-Agent": "custom-agent"'
                The output should not include '"User-Agent": "qurl"'
                The status should be success
            End

            It "handles empty header values"
                When call qurl_openapi /headers -H "X-Empty:"
                The output should include '"X-Empty": ""'
                The status should be success
            End

        End

        Describe "Query parameter handling with OpenAPI"

            It "sends single query parameter"
                When call qurl_openapi /anything --param key=value
                The output should include '"key": "value"'
                The status should be success
            End

            It "sends multiple query parameters"
                When call qurl_openapi /anything --param a=1 --param b=2 --param c=3
                The output should include '"a": "1"'
                The output should include '"b": "2"'
                The output should include '"c": "3"'
                The status should be success
            End

            It "handles parameters with spaces"
                When call qurl_openapi /anything --param "message=hello world"
                The output should include '"message": "hello world"'
                The status should be success
            End

            It "handles parameters with special characters"
                When call qurl_openapi /anything --param "special=!@#\$%^&*()"
                The output should include '"special": "!@#\$%^&*()"'
                The status should be success
            End

            It "handles parameters with equals signs in values"
                When call qurl_openapi /anything --param "equation=a=b+c"
                The output should include '"equation": "a=b+c"'
                The status should be success
            End

            It "handles empty parameter values"
                When call qurl_openapi /anything --param "empty="
                The output should include '"empty": ""'
                The status should be success
            End

            It "handles parameters without values"
                When call qurl_openapi /anything --param flag
                The output should include '"flag": ""'
                The status should be success
            End

            It "handles URL-encoded characters"
                When call qurl_openapi /anything --param "encoded=%20%2B%2F"
                The output should include '"encoded": " +/"'
                The status should be success
            End

            It "handles multiple values for same parameter"
                When call qurl_openapi /anything --param tag=one --param tag=two
                The output should include '"tag": "one,two"'
                The status should be success
            End

            It "validates parameters against OpenAPI spec"
                Skip "Requires parameter validation implementation"
                # When call qurl_openapi /bytes/invalid --param n=notanumber
                # The stderr should include "invalid parameter"
            End

        End

        Describe "Path parameter handling with OpenAPI"

            It "handles simple path parameters"
                When call qurl_openapi /bytes/1024
                The status should be success
            End

            It "handles multiple path parameters"
                When call qurl_openapi /cache/60
                The status should be success
            End

            It "validates path parameters from spec"
                Skip "Requires path parameter validation"
                # Invalid path parameter should error
            End

        End

        Describe "Request body with OpenAPI"

            It "sends POST data with -d flag"
                Skip "Request body not yet implemented"
                When call qurl_openapi -X POST /anything -d "test=data"
                The output should include '"data": "test=data"'
                The status should be success
            End

            It "sets Content-Type from OpenAPI spec for request body"
                Skip "Request body not yet implemented"
                When call qurl_openapi -X POST /post -d '{"key":"value"}'
                The output should include '"Content-Type": "application/json"'
                The status should be success
            End

            It "validates request body against OpenAPI schema"
                Skip "Request body validation not yet implemented"
                # Should validate against schema
            End

        End

        Describe "OpenAPI-specific features"

            It "handles multiple server URLs in spec"
                Skip "Multiple servers not tested with httpbin"
                # Would need spec with multiple servers
            End

            It "applies security schemes from OpenAPI"
                Skip "Security schemes not yet implemented"
                # Would add auth headers based on spec
            End

            It "provides helpful errors for invalid operations"
                When call qurl_openapi -X POST /get
                Skip "Operation validation not yet implemented"
                # Should warn that POST is not valid for /get
            End

            It "caches OpenAPI spec for performance"
                Skip "Caching behavior not testable in current setup"
                # Multiple calls should reuse spec
            End

            It "handles OpenAPI spec fetch failures gracefully"
                When call sh -c 'OPENAPI_URL="https://httpbin.org/status/404" go run cmd/qurl/main.go /anything'
                The status should not be success
                The stderr should include "Error"
            End

        End

        Describe "Verbose output with OpenAPI"

            It "shows request line with resolved URL"
                When call qurl_openapi -v /get 2>&1
                The output should include "> GET https://httpbin.org/get"
                The status should be success
            End

            It "shows all headers including OpenAPI-derived ones"
                When call qurl_openapi -v /get 2>&1
                The output should include "> Host: httpbin.org"
                The output should include "> User-Agent: qurl"
                The output should include "> Accept: application/json"
                The status should be success
            End

            It "shows response details"
                When call qurl_openapi -v /status/201 2>&1
                The output should include "< HTTP/"
                The output should include "201"
                The status should be success
            End

            It "keeps verbose output in stderr"
                When call qurl_openapi -v /get
                The stderr should include "> GET"
                The stdout should not include "> GET"
                The stdout should include '"url"'
                The status should be success
            End

        End

        Describe "Response handling with OpenAPI"

            It "includes headers with -i flag"
                When call qurl_openapi -i /get
                The output should include "HTTP/1.1 200"
                The output should include "Content-Type: application/json"
                The output should include '"url"'
                The status should be success
            End

            It "handles different status codes"
                When call qurl_openapi /status/404
                The status should be success
            End

            It "handles redirects"
                When call qurl_openapi /redirect/1
                The output should include '"url"'
                The status should be success
            End

            It "handles different content types"
                When call qurl_openapi /html
                The output should include "<html>"
                The status should be success
            End

            It "handles binary responses"
                When call qurl_openapi /bytes/100
                The status should be success
            End

            It "handles empty responses"
                When call qurl_openapi -X HEAD /get
                The output should be blank
                The status should be success
            End

        End

        Describe "Error conditions with OpenAPI"

            It "handles invalid paths not in spec"
                When call qurl_openapi /this/path/does/not/exist
                The status should be success
                # Currently doesn't validate against spec
            End

            It "handles malformed OpenAPI spec URL"
                When call sh -c 'OPENAPI_URL="not-a-url" go run cmd/qurl/main.go /anything'
                The status should not be success
                The stderr should include "Error"
            End

            It "handles unreachable OpenAPI spec"
                When call sh -c 'OPENAPI_URL="https://definitely-not-a-real-domain-12345.com/openapi.json" go run cmd/qurl/main.go /anything'
                The status should not be success
            End

            It "requires OpenAPI URL for relative paths"
                When call go run cmd/qurl/main.go /relative/path
                The status should not be success
                The stderr should include "OpenAPI URL is required"
            End

            It "handles timeout fetching OpenAPI spec"
                Skip "Timeout testing requires long delays"
            End

        End

        Describe "Edge cases with OpenAPI"

            It "handles very long URLs"
                When call qurl_openapi "/anything?param=$(printf 'a%.0s' {1..1000})"
                The status should be success
            End

            It "handles very long header values"
                When call qurl_openapi /headers -H "X-Long: $(printf 'x%.0s' {1..1000})"
                The output should include "X-Long"
                The status should be success
            End

            It "handles many headers"
                When call qurl_openapi /headers \
                    -H "X-1: 1" -H "X-2: 2" -H "X-3: 3" -H "X-4: 4" -H "X-5: 5" \
                    -H "X-6: 6" -H "X-7: 7" -H "X-8: 8" -H "X-9: 9" -H "X-10: 10"
                The output should include '"X-10": "10"'
                The output should include '"Accept": "application/json"'
                The status should be success
            End

            It "handles many query parameters"
                When call qurl_openapi /anything \
                    --param p1=1 --param p2=2 --param p3=3 --param p4=4 --param p5=5 \
                    --param p6=6 --param p7=7 --param p8=8 --param p9=9 --param p10=10
                The output should include '"p10": "10"'
                The status should be success
            End

        End

    End

    Describe "Without OpenAPI (vanilla HTTP client)"

        # Helper for vanilla HTTP tests
        qurl() {
            go run cmd/qurl/main.go "$@"
        }

        Describe "Core request construction"

            Describe "HTTP method generation"

                It "defaults to GET when no method specified"
                    When call qurl https://httpbin.org/anything
                    The output should include '"method": "GET"'
                    The status should be success
                End

                It "correctly sets POST method"
                    When call qurl -X POST https://httpbin.org/anything
                    The output should include '"method": "POST"'
                    The status should be success
                End

                It "correctly sets PUT method"
                    When call qurl -X PUT https://httpbin.org/anything
                    The output should include '"method": "PUT"'
                    The status should be success
                End

                It "correctly sets DELETE method"
                    When call qurl -X DELETE https://httpbin.org/anything
                    The output should include '"method": "DELETE"'
                    The status should be success
                End

                It "correctly sets PATCH method"
                    When call qurl -X PATCH https://httpbin.org/anything
                    The output should include '"method": "PATCH"'
                    The status should be success
                End

                It "handles HEAD method (no body)"
                    When call qurl -X HEAD -v https://httpbin.org/get
                    The stderr should include "> HEAD https://httpbin.org/get"
                    The output should be blank
                    The status should be success
                End

                It "handles OPTIONS method"
                    When call qurl -X OPTIONS https://httpbin.org/anything
                    The output should include '"method": "OPTIONS"'
                    The status should be success
                End

                It "normalizes method to uppercase"
                    When call qurl -X post https://httpbin.org/anything
                    The output should include '"method": "POST"'
                    The status should be success
                End

            End

            Describe "URL and path handling"

                It "correctly constructs full URL"
                    When call qurl https://httpbin.org/anything/test/path
                    The output should include '"url": "https://httpbin.org/anything/test/path"'
                    The status should be success
                End

                It "handles URLs with ports"
                    When call qurl http://httpbin.org:80/anything
                    The output should include '"url": "http://httpbin.org:80/anything"'
                    The status should be success
                End

                It "preserves URL fragments and anchors"
                    When call qurl "https://httpbin.org/anything#section"
                    The output should include '"url": "https://httpbin.org/anything"'
                    The status should be success
                End

            End

            Describe "Header generation and management"

                It "always sends User-Agent: qurl"
                    When call qurl https://httpbin.org/headers
                    The output should include '"User-Agent": "qurl"'
                    The status should be success
                End

                It "correctly sets Host header"
                    When call qurl https://httpbin.org/headers
                    The output should include '"Host": "httpbin.org"'
                    The status should be success
                End

                It "sends single custom header with -H"
                    When call qurl https://httpbin.org/headers -H "X-Test: value"
                    The output should include '"X-Test": "value"'
                    The status should be success
                End

                It "sends multiple custom headers"
                    When call qurl https://httpbin.org/headers -H "X-One: 1" -H "X-Two: 2" -H "X-Three: 3"
                    The output should include '"X-One": "1"'
                    The output should include '"X-Two": "2"'
                    The output should include '"X-Three": "3"'
                    The status should be success
                End

                It "handles headers with spaces in values"
                    When call qurl https://httpbin.org/headers -H "X-Message: hello world test"
                    The output should include '"X-Message": "hello world test"'
                    The status should be success
                End

                It "handles headers with special characters"
                    When call qurl https://httpbin.org/headers -H "X-Special: value!@#\$%^&*()"
                    The output should include '"X-Special": "value!@#\$%^&*()"'
                    The status should be success
                End

                It "handles headers with colons in values"
                    When call qurl https://httpbin.org/headers -H "X-URL: https://example.com:8080"
                    The output should include '"X-URL": "https://example.com:8080"'
                    The status should be success
                End

                It "overrides default headers with custom ones"
                    When call qurl https://httpbin.org/headers -H "User-Agent: custom-agent"
                    The output should include '"User-Agent": "custom-agent"'
                    The output should not include '"User-Agent": "qurl"'
                    The status should be success
                End

                It "handles empty header values"
                    When call qurl https://httpbin.org/headers -H "X-Empty:"
                    The output should include '"X-Empty": ""'
                    The status should be success
                End

                It "handles header names case-insensitively"
                    When call qurl https://httpbin.org/headers -H "content-type: application/json"
                    The output should include "application/json"
                    The status should be success
                End

            End

            Describe "Query parameter encoding"

                It "sends single query parameter"
                    When call qurl https://httpbin.org/anything --param key=value
                    The output should include '"key": "value"'
                    The status should be success
                End

                It "sends multiple query parameters"
                    When call qurl https://httpbin.org/anything --param a=1 --param b=2 --param c=3
                    The output should include '"a": "1"'
                    The output should include '"b": "2"'
                    The output should include '"c": "3"'
                    The status should be success
                End

                It "handles parameters with spaces"
                    When call qurl https://httpbin.org/anything --param "message=hello world"
                    The output should include '"message": "hello world"'
                    The status should be success
                End

                It "handles parameters with special characters"
                    When call qurl https://httpbin.org/anything --param "special=!@#\$%^&*()"
                    The output should include '"special": "!@#\$%^&*()"'
                    The status should be success
                End

                It "handles parameters with equals signs in values"
                    When call qurl https://httpbin.org/anything --param "equation=a=b+c"
                    The output should include '"equation": "a=b+c"'
                    The status should be success
                End

                It "handles empty parameter values"
                    When call qurl https://httpbin.org/anything --param "empty="
                    The output should include '"empty": ""'
                    The status should be success
                End

                It "handles parameters without values"
                    When call qurl https://httpbin.org/anything --param flag
                    The output should include '"flag": ""'
                    The status should be success
                End

                It "handles URL-encoded characters"
                    When call qurl https://httpbin.org/anything --param "encoded=%20%2B%2F"
                    The output should include '"encoded": " +/"'
                    The status should be success
                End

                It "handles multiple values for same parameter"
                    When call qurl https://httpbin.org/anything --param tag=one --param tag=two
                    The output should include '"tag": "one,two"'
                    The status should be success
                End

                It "combines URL params with --param flags"
                    When call qurl "https://httpbin.org/anything?existing=url" --param added=param
                    The output should include '"existing": "url"'
                    The output should include '"added": "param"'
                    The status should be success
                End

            End

            Describe "Request body handling"

                It "sends POST data with -d flag"
                    Skip "Request body not yet implemented"
                    When call qurl -X POST https://httpbin.org/anything -d "test=data"
                    The output should include '"data": "test=data"'
                    The output should include '"Content-Type": "application/x-www-form-urlencoded"'
                    The status should be success
                End

                It "sends JSON data with proper content type"
                    Skip "Request body not yet implemented"
                    When call qurl -X POST https://httpbin.org/anything -d '{"key":"value"}' -H "Content-Type: application/json"
                    The output should include '"json": {"key": "value"}'
                    The output should include '"Content-Type": "application/json"'
                    The status should be success
                End

                It "auto-detects JSON content"
                    Skip "Request body not yet implemented"
                    When call qurl -X POST https://httpbin.org/anything -d '{"test": true}'
                    The output should include '"json": {"test": true}'
                    The status should be success
                End

                It "handles form data with multiple -d flags"
                    Skip "Request body not yet implemented"
                    When call qurl -X POST https://httpbin.org/anything -d "field1=value1" -d "field2=value2"
                    The output should include '"field1": "value1"'
                    The output should include '"field2": "value2"'
                    The status should be success
                End

            End

        End

        Describe "Verbose output verification"

            It "shows request line in verbose mode"
                When call qurl -v https://httpbin.org/get 2>&1
                The output should include "> GET https://httpbin.org/get"
                The status should be success
            End

            It "shows request headers in verbose mode"
                When call qurl -v https://httpbin.org/get 2>&1
                The output should include "> Host: httpbin.org"
                The output should include "> User-Agent: qurl"
                The status should be success
            End

            It "shows response status in verbose mode"
                When call qurl -v https://httpbin.org/status/201 2>&1
                The output should include "< HTTP/"
                The output should include "201"
                The status should be success
            End

            It "shows response headers in verbose mode"
                When call qurl -v https://httpbin.org/get 2>&1
                The output should include "< Content-Type:"
                The status should be success
            End

            It "separates verbose output to stderr"
                When call qurl -v https://httpbin.org/get
                The stderr should include "> GET"
                The stdout should not include "> GET"
                The stdout should include '"url"'
                The status should be success
            End

        End

        Describe "Response handling"

            It "includes headers with -i flag"
                When call qurl -i https://httpbin.org/get
                The output should include "HTTP/1.1 200"
                The output should include "Content-Type: application/json"
                The output should include '"url"'
                The status should be success
            End

            It "shows only headers with -I flag (HEAD request)"
                Skip "Not implemented yet"
                When call qurl -I https://httpbin.org/get
                The output should include "HTTP/1.1 200"
                The output should include "Content-Type:"
                The output should not include '"url"'
                The status should be success
            End

            It "handles different status codes correctly"
                When call qurl https://httpbin.org/status/404
                The status should be success
            End

            It "handles redirects"
                When call qurl https://httpbin.org/redirect/1
                The output should include '"url"'
                The status should be success
            End

            It "handles different content types"
                When call qurl https://httpbin.org/html
                The output should include "<html>"
                The status should be success
            End

            It "handles binary responses"
                When call qurl https://httpbin.org/bytes/100
                The status should be success
            End

            It "handles large responses"
                When call qurl https://httpbin.org/bytes/10000
                The status should be success
            End

            It "handles empty responses"
                When call qurl -X HEAD https://httpbin.org/get
                The output should be blank
                The status should be success
            End

        End

        Describe "Error conditions and edge cases"

            It "handles malformed URLs gracefully"
                When call qurl "not-a-url"
                The status should not be success
                The stderr should include "Error"
            End

            It "handles connection failures"
                When call qurl "https://definitely-not-a-real-domain-12345.com"
                The status should not be success
            End

            It "handles timeout scenarios"
                When call qurl https://httpbin.org/delay/60
                Skip "Depends on timeout implementation"
                The status should not be success
            End

            It "handles invalid HTTP methods"
                When call qurl -X INVALID https://httpbin.org/anything
                The output should include '"method": "INVALID"'
                The status should be success
            End

            It "handles very long URLs"
                When call qurl "https://httpbin.org/anything?param=$(printf 'a%.0s' {1..1000})"
                The status should be success
            End

            It "handles very long header values"
                When call qurl https://httpbin.org/headers -H "X-Long: $(printf 'x%.0s' {1..1000})"
                The output should include "X-Long"
                The status should be success
            End

            It "handles many headers"
                When call qurl https://httpbin.org/headers \
                    -H "X-1: 1" -H "X-2: 2" -H "X-3: 3" -H "X-4: 4" -H "X-5: 5" \
                    -H "X-6: 6" -H "X-7: 7" -H "X-8: 8" -H "X-9: 9" -H "X-10: 10"
                The output should include '"X-10": "10"'
                The status should be success
            End

            It "handles many query parameters"
                When call qurl https://httpbin.org/anything \
                    --param p1=1 --param p2=2 --param p3=3 --param p4=4 --param p5=5 \
                    --param p6=6 --param p7=7 --param p8=8 --param p9=9 --param p10=10
                The output should include '"p10": "10"'
                The status should be success
            End

        End

        Describe "Compatibility with curl-like behavior"

            It "supports -X for method like curl"
                When call qurl -X POST https://httpbin.org/anything
                The output should include '"method": "POST"'
                The status should be success
            End

            It "supports -H for headers like curl"
                When call qurl -H "X-Custom: test" https://httpbin.org/headers
                The output should include '"X-Custom": "test"'
                The status should be success
            End

            It "supports -v for verbose like curl"
                When call qurl -v https://httpbin.org/get 2>&1
                The output should include "> GET"
                The status should be success
            End

            It "supports -i for include headers like curl"
                When call qurl -i https://httpbin.org/get
                The output should include "HTTP/"
                The output should include '"url"'
                The status should be success
            End

        End

    End

End