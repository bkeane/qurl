package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Config holds logger configuration
type Config struct {
	Level       string
	Format      string // "pretty" or "json"
	WithCaller  bool
	Output      io.Writer
	TimeFormat  string
}

// DefaultConfig returns sensible defaults for logging
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "pretty",
		WithCaller: false,
		Output:     os.Stderr,
		TimeFormat: time.RFC3339,
	}
}

// InitLogger creates and configures a new zerolog logger
func InitLogger(config *Config) zerolog.Logger {
	if config == nil {
		config = DefaultConfig()
	}

	// Set global log level
	level := parseLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = config.Output
	if config.Format == "pretty" {
		output = &zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
	}

	// Create logger with timestamp
	logger := zerolog.New(output).With().
		Timestamp().
		Str("app", "qurl").
		Logger()

	// Add caller info if requested
	if config.WithCaller {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// SetupFromFlags configures logger based on command flags
func SetupFromFlags(verbose bool, debug bool) zerolog.Logger {
	config := DefaultConfig()

	// Determine log level
	if debug {
		config.Level = "debug"
		config.WithCaller = true
	} else if verbose {
		config.Level = "info"
	} else {
		config.Level = "warn"
	}

	return InitLogger(config)
}

// ForComponent creates a logger with component context
func ForComponent(logger zerolog.Logger, component string) zerolog.Logger {
	return logger.With().Str("component", component).Logger()
}

// ForRequest creates a logger with request context
func ForRequest(logger zerolog.Logger, method, path string) zerolog.Logger {
	return logger.With().
		Str("method", method).
		Str("path", path).
		Logger()
}

// ForMCP creates a logger with MCP context
func ForMCP(logger zerolog.Logger, tool string) zerolog.Logger {
	return logger.With().
		Str("mcp_tool", tool).
		Str("component", "mcp").
		Logger()
}