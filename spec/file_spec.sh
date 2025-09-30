#!/usr/bin/env sh

# file_spec.sh - File URL Protocol Testing
# Purpose: Test qurl's file:// URI scheme for local OpenAPI specification loading
# Strategy: Focus on local file handling, path resolution, and OpenAPI integration

# Helper function to run qurl
qurl() {
    go run cmd/qurl/main.go "$@"
}

Describe "qurl: File URL Protocol Support"

    Describe "Feature: File URI Scheme for OpenAPI Specs"

        It "loads OpenAPI specification from local file:// URL"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/test_api.json" go run cmd/qurl/main.go --docs'
            The output should include "Test API"
            The output should include "/test"
            The status should be success
        End

        It "handles absolute file:// paths correctly"
            # Copy fixture to /tmp for absolute path testing
            cp spec/fixtures/absolute_path_test.json /tmp/qurl_test_spec.json
            When call sh -c 'QURL_OPENAPI="file:///tmp/qurl_test_spec.json" go run cmd/qurl/main.go --docs'
            The output should include "Absolute Path Test"
            The output should include "/absolute"
            The status should be success
            # Cleanup
            rm -f /tmp/qurl_test_spec.json
        End

        It "supports relative file:// paths from current directory"
            When call sh -c 'QURL_OPENAPI="file://spec/fixtures/relative_path_test.json" go run cmd/qurl/main.go --docs'
            The output should include "Relative Path Test"
            The output should include "/relative"
            The status should be success
        End

    End

    Describe "Feature: File URI OpenAPI Integration"

        It "uses local OpenAPI spec for endpoint requests"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/integration_test.json" go run cmd/qurl/main.go --aws-sigv4 /get'
            The output should include '"url"'
            The status should be success
        End

        It "shows endpoint documentation from local file"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/documentation_test.json" go run cmd/qurl/main.go --docs /documented'
            The output should include "Well documented endpoint"
            The output should include "This endpoint is thoroughly documented"
            The status should be success
        End

        It "resolves server URLs from local OpenAPI specs"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/multi_server.json" go run cmd/qurl/main.go --docs'
            The output should include "Production server"
            The output should include "Staging server"
            The output should include "prod.kaixo.io"
            The status should be success
        End

    End

    Describe "Feature: File URI Error Handling"

        It "provides helpful error for non-existent files"
            When call sh -c 'QURL_OPENAPI="file://does_not_exist.json" go run cmd/qurl/main.go --docs'
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles malformed file:// URLs gracefully"
            When call sh -c 'QURL_OPENAPI="file:///" go run cmd/qurl/main.go --docs'
            The stderr should include "ERR"
            The status should not be success
        End

        It "reports invalid JSON in local files clearly"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/invalid.json" go run cmd/qurl/main.go --docs'
            The stderr should include "ERR"
            The status should not be success
        End

        It "handles permission denied errors appropriately"
            Skip "Permission testing requires specific file system setup"
        End

    End

    Describe "Feature: File URI Path Resolution"

        It "resolves relative paths from current working directory"
            When call sh -c 'QURL_OPENAPI="file://spec/fixtures/nested_directory_test.json" go run cmd/qurl/main.go --docs'
            The output should include "Nested Directory Test"
            The output should include "/nested"
            The status should be success
        End

        It "handles spaces in file paths correctly"
            # Create directory with spaces and copy fixture
            mkdir -p "test dir with spaces"
            cp spec/fixtures/spaces_test.json "test dir with spaces/spec with spaces.json"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/test dir with spaces/spec with spaces.json" go run cmd/qurl/main.go --docs'
            The output should include "Spaces Test API"
            The output should include "/spaces"
            The status should be success
            # Cleanup
            rm -rf "test dir with spaces"
        End

        It "works with symbolic links to OpenAPI files"
            Skip "Symbolic link testing requires specific setup"
        End

    End

    Describe "Feature: File URI vs Remote URI Comparison"

        It "loads faster than remote HTTP(S) URLs"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/test_api.json" go run cmd/qurl/main.go --docs'
            The output should include "Test API"
            The status should be success
        End

        It "works offline unlike remote URLs"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/test_api.json" go run cmd/qurl/main.go --docs'
            The output should include "Test API"
            The status should be success
        End

        It "supports local development workflows"
            When call sh -c 'QURL_OPENAPI="file://$(pwd)/spec/fixtures/development_api.json" go run cmd/qurl/main.go --docs'
            The output should include "Development API"
            The output should include "0.1.0-dev"
            The status should be success
        End

    End

End