package testutil

import (
	"net/http"
	"strings"
	"testing"
)

// Custom assertion helpers to reduce boilerplate in tests

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: got error %v, expected none", msg, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got none", msg)
	}
}

// AssertErrorContains fails the test if err is nil or doesn't contain the expected substring
func AssertErrorContains(t *testing.T, err error, expected string, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error containing %q, got none", msg, expected)
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("%s: expected error containing %q, got %q", msg, expected, err.Error())
	}
}

// AssertEqual fails the test if got != expected
func AssertEqual(t *testing.T, got, expected interface{}, msg string) {
	t.Helper()
	if got != expected {
		t.Fatalf("%s: got %v, expected %v", msg, got, expected)
	}
}

// AssertStringEqual fails the test if got != expected (string-specific for cleaner output)
func AssertStringEqual(t *testing.T, got, expected string, msg string) {
	t.Helper()
	if got != expected {
		t.Fatalf("%s: got %q, expected %q", msg, got, expected)
	}
}

// AssertStringContains fails the test if str doesn't contain substring
func AssertStringContains(t *testing.T, str, substring string, msg string) {
	t.Helper()
	if !strings.Contains(str, substring) {
		t.Fatalf("%s: expected %q to contain %q", msg, str, substring)
	}
}

// AssertStringNotContains fails the test if str contains substring
func AssertStringNotContains(t *testing.T, str, substring string, msg string) {
	t.Helper()
	if strings.Contains(str, substring) {
		t.Fatalf("%s: expected %q to not contain %q", msg, str, substring)
	}
}

// AssertSliceEqual fails the test if slices don't have the same elements in the same order
func AssertSliceEqual(t *testing.T, got, expected []string, msg string) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("%s: got %d elements, expected %d\ngot: %v\nexpected: %v", msg, len(got), len(expected), got, expected)
	}

	for i, g := range got {
		if g != expected[i] {
			t.Fatalf("%s: element %d: got %q, expected %q\ngot: %v\nexpected: %v", msg, i, g, expected[i], got, expected)
		}
	}
}

// AssertSliceContains fails the test if slice doesn't contain element
func AssertSliceContains(t *testing.T, slice []string, element string, msg string) {
	t.Helper()
	for _, item := range slice {
		if item == element {
			return
		}
	}
	t.Fatalf("%s: expected slice %v to contain %q", msg, slice, element)
}

// AssertHeaderSet fails the test if the request doesn't have the expected header value
func AssertHeaderSet(t *testing.T, req *http.Request, header, expectedValue string, msg string) {
	t.Helper()
	actualValue := req.Header.Get(header)
	if actualValue != expectedValue {
		t.Fatalf("%s: header %q: got %q, expected %q", msg, header, actualValue, expectedValue)
	}
}

// AssertHeaderContains fails the test if the request header doesn't contain the expected substring
func AssertHeaderContains(t *testing.T, req *http.Request, header, expectedSubstring string, msg string) {
	t.Helper()
	actualValue := req.Header.Get(header)
	if !strings.Contains(actualValue, expectedSubstring) {
		t.Fatalf("%s: header %q: got %q, expected to contain %q", msg, header, actualValue, expectedSubstring)
	}
}

// AssertHeaderNotSet fails the test if the request has the specified header
func AssertHeaderNotSet(t *testing.T, req *http.Request, header string, msg string) {
	t.Helper()
	if req.Header.Get(header) != "" {
		t.Fatalf("%s: expected header %q to not be set, but got %q", msg, header, req.Header.Get(header))
	}
}

// AssertMethodEqual fails the test if the request method doesn't match expected
func AssertMethodEqual(t *testing.T, req *http.Request, expectedMethod string, msg string) {
	t.Helper()
	if req.Method != expectedMethod {
		t.Fatalf("%s: got method %q, expected %q", msg, req.Method, expectedMethod)
	}
}

// AssertPathEqual fails the test if the request path doesn't match expected
func AssertPathEqual(t *testing.T, req *http.Request, expectedPath string, msg string) {
	t.Helper()
	if req.URL.Path != expectedPath {
		t.Fatalf("%s: got path %q, expected %q", msg, req.URL.Path, expectedPath)
	}
}

// AssertQueryParam fails the test if the request doesn't have the expected query parameter
func AssertQueryParam(t *testing.T, req *http.Request, param, expectedValue string, msg string) {
	t.Helper()
	actualValue := req.URL.Query().Get(param)
	if actualValue != expectedValue {
		t.Fatalf("%s: query param %q: got %q, expected %q", msg, param, actualValue, expectedValue)
	}
}

// AssertMockCalled fails the test if the mock wasn't called the expected number of times
func AssertMockCalled(t *testing.T, actualCalls, expectedCalls int, mockName string) {
	t.Helper()
	if actualCalls != expectedCalls {
		t.Fatalf("Mock %s: expected %d calls, got %d", mockName, expectedCalls, actualCalls)
	}
}

// AssertMockCalledWith fails the test if the mock wasn't called with expected parameters
func AssertMockCalledWith(t *testing.T, calls []string, expectedCall string, mockName string) {
	t.Helper()
	for _, call := range calls {
		if call == expectedCall {
			return
		}
	}
	t.Fatalf("Mock %s: expected call with %q, got calls: %v", mockName, expectedCall, calls)
}

// Helper functions for common test patterns

// SkipIfShort skips the test if running with -short flag (for integration tests)
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("Skipping in short mode: %s", reason)
	}
}