package errors

import (
	"fmt"
	"testing"
)

func TestQUrlError(t *testing.T) {
	// Test basic error creation
	err := New(ErrorTypeValidation, "test error")
	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}
	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}

	// Test error wrapping
	cause := fmt.Errorf("underlying error")
	wrapped := Wrap(cause, ErrorTypeNetwork, "network issue")
	if wrapped.Cause != cause {
		t.Errorf("Expected cause to be preserved")
	}
	if wrapped.Type != ErrorTypeNetwork {
		t.Errorf("Expected type %s, got %s", ErrorTypeNetwork, wrapped.Type)
	}

	// Test context
	err.WithContext("field", "test_field")
	if err.Context["field"] != "test_field" {
		t.Errorf("Expected context to be set")
	}

	// Test error string
	errStr := wrapped.Error()
	expected := "network issue: underlying error"
	if errStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, errStr)
	}
}

func TestIsType(t *testing.T) {
	err := New(ErrorTypeAuth, "auth error")

	if !IsType(err, ErrorTypeAuth) {
		t.Errorf("Expected IsType to return true for correct type")
	}

	if IsType(err, ErrorTypeNetwork) {
		t.Errorf("Expected IsType to return false for incorrect type")
	}

	// Test with non-QUrlError
	stdErr := fmt.Errorf("standard error")
	if IsType(stdErr, ErrorTypeAuth) {
		t.Errorf("Expected IsType to return false for standard error")
	}
}

func TestGetType(t *testing.T) {
	err := New(ErrorTypeConfig, "config error")
	if GetType(err) != ErrorTypeConfig {
		t.Errorf("Expected type %s, got %s", ErrorTypeConfig, GetType(err))
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	if GetType(stdErr) != ErrorTypeInternal {
		t.Errorf("Expected type %s for standard error, got %s", ErrorTypeInternal, GetType(stdErr))
	}
}