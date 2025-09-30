package errors

import (
	"fmt"

	"github.com/rs/zerolog/log"
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

	// Don't add suggestion formatting - let zerolog handle all formatting
	// The suggestion context can be logged as a separate field if needed
	return msg
}

func formatNetworkError(qErr *QUrlError) string {
	msg := qErr.Message
	if url, ok := qErr.Context["url"]; ok {
		msg = fmt.Sprintf("Network error accessing %s: %s", url, msg)
	}

	return msg
}

func formatAuthError(qErr *QUrlError) string {
	return qErr.Message
}

func formatConfigError(qErr *QUrlError) string {
	msg := qErr.Message

	if configType, ok := qErr.Context["config_type"]; ok {
		msg = fmt.Sprintf("Configuration error (%s): %s", configType, msg)
	}

	return msg
}

func formatOpenAPIError(qErr *QUrlError) string {
	return qErr.Message
}

// PresentError displays an error to the user through centralized zerolog system
func PresentError(err error) {
	if err == nil {
		return
	}

	// Use the global logger
	if qErr, ok := err.(*QUrlError); ok {
		event := log.Fatal()

		// Add context fields as structured data
		for key, value := range qErr.Context {
			event = event.Interface(key, value)
		}

		event.Msg(qErr.Message)
	} else {
		log.Fatal().Err(err).Msg("")
	}
}

func formatMCPError(qErr *QUrlError) string {
	return qErr.Message
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