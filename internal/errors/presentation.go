package errors

import (
	"fmt"
	"os"
	"strings"
)

// UserMessage returns a user-friendly error message
func UserMessage(err error) string {
	if qErr, ok := err.(*QUrlError); ok {
		return formatUserError(qErr)
	}
	return err.Error()
}

// formatUserError creates user-friendly error messages based on error type
func formatUserError(qErr *QUrlError) string {
	switch qErr.Type {
	case ErrorTypeValidation:
		return formatValidationError(qErr)
	case ErrorTypeNetwork:
		return formatNetworkError(qErr)
	case ErrorTypeAuth:
		return formatAuthError(qErr)
	case ErrorTypeConfig:
		return formatConfigError(qErr)
	case ErrorTypeOpenAPI:
		return formatOpenAPIError(qErr)
	case ErrorTypeMCP:
		return formatMCPError(qErr)
	default:
		return qErr.Message
	}
}

func formatValidationError(qErr *QUrlError) string {
	msg := qErr.Message
	if field, ok := qErr.Context["field"]; ok {
		msg = fmt.Sprintf("Invalid %s: %s", field, msg)
	}

	if suggestion, ok := qErr.Context["suggestion"]; ok {
		msg = fmt.Sprintf("%s\nSuggestion: %s", msg, suggestion)
	}

	return msg
}

func formatNetworkError(qErr *QUrlError) string {
	msg := qErr.Message
	if url, ok := qErr.Context["url"]; ok {
		msg = fmt.Sprintf("Network error accessing %s: %s", url, msg)
	}

	// Add common troubleshooting suggestions
	suggestions := []string{}
	if strings.Contains(strings.ToLower(qErr.Message), "timeout") {
		suggestions = append(suggestions, "Check your internet connection")
		suggestions = append(suggestions, "Try increasing timeout with --timeout flag")
	}
	if strings.Contains(strings.ToLower(qErr.Message), "connection refused") {
		suggestions = append(suggestions, "Verify the server is running")
		suggestions = append(suggestions, "Check the URL and port")
	}

	if len(suggestions) > 0 {
		msg = fmt.Sprintf("%s\nTroubleshooting:\n  - %s", msg, strings.Join(suggestions, "\n  - "))
	}

	return msg
}

func formatAuthError(qErr *QUrlError) string {
	msg := qErr.Message

	suggestions := []string{}
	if strings.Contains(strings.ToLower(qErr.Message), "unauthorized") {
		suggestions = append(suggestions, "Check your authentication credentials")
		suggestions = append(suggestions, "Verify API key or token is valid")
		suggestions = append(suggestions, "Use -H \"Authorization: Bearer YOUR_TOKEN\"")
	}
	if strings.Contains(strings.ToLower(qErr.Message), "forbidden") {
		suggestions = append(suggestions, "Check if you have permission for this operation")
		suggestions = append(suggestions, "Verify your account has required access")
	}

	if len(suggestions) > 0 {
		msg = fmt.Sprintf("%s\nAuth help:\n  - %s", msg, strings.Join(suggestions, "\n  - "))
	}

	return msg
}

func formatConfigError(qErr *QUrlError) string {
	msg := qErr.Message

	if configType, ok := qErr.Context["config_type"]; ok {
		msg = fmt.Sprintf("Configuration error (%s): %s", configType, msg)
	}

	suggestions := []string{}
	if strings.Contains(strings.ToLower(qErr.Message), "openapi") {
		suggestions = append(suggestions, "Set QURL_OPENAPI environment variable")
		suggestions = append(suggestions, "Use --openapi flag to specify OpenAPI URL")
		suggestions = append(suggestions, "Verify OpenAPI spec is accessible")
	}
	if strings.Contains(strings.ToLower(qErr.Message), "server") {
		suggestions = append(suggestions, "Use --server flag to specify server URL")
		suggestions = append(suggestions, "Check if server URL is correct")
	}

	if len(suggestions) > 0 {
		msg = fmt.Sprintf("%s\nConfiguration help:\n  - %s", msg, strings.Join(suggestions, "\n  - "))
	}

	return msg
}

func formatOpenAPIError(qErr *QUrlError) string {
	msg := qErr.Message

	suggestions := []string{}
	suggestions = append(suggestions, "Verify OpenAPI specification is valid")
	suggestions = append(suggestions, "Check if the OpenAPI URL is accessible")
	suggestions = append(suggestions, "Try using --docs to explore the API")

	return fmt.Sprintf("%s\nOpenAPI help:\n  - %s", msg, strings.Join(suggestions, "\n  - "))
}

// PresentError displays an error to the user with appropriate formatting and context
func PresentError(err error) {
	if err == nil {
		return
	}

	if qErr, ok := err.(*QUrlError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", UserMessage(qErr))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func formatMCPError(qErr *QUrlError) string {
	msg := qErr.Message

	suggestions := []string{}
	if strings.Contains(strings.ToLower(qErr.Message), "method") {
		suggestions = append(suggestions, "Check --allow-methods flag")
		suggestions = append(suggestions, "Verify the HTTP method is supported")
	}
	if strings.Contains(strings.ToLower(qErr.Message), "tool") {
		suggestions = append(suggestions, "Use 'discover' tool to explore the API")
		suggestions = append(suggestions, "Check tool parameters are correct")
	}

	if len(suggestions) > 0 {
		msg = fmt.Sprintf("%s\nMCP help:\n  - %s", msg, strings.Join(suggestions, "\n  - "))
	}

	return msg
}

// DebugInfo returns detailed error information for debugging
func DebugInfo(err error) map[string]interface{} {
	info := map[string]interface{}{
		"error":   err.Error(),
		"type":    "unknown",
		"context": map[string]interface{}{},
	}

	if qErr, ok := err.(*QUrlError); ok {
		info["type"] = string(qErr.Type)
		info["message"] = qErr.Message
		info["context"] = qErr.Context

		if qErr.Cause != nil {
			info["cause"] = qErr.Cause.Error()
		}
	}

	return info
}