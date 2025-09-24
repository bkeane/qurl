package cli

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func TestNewHTTPHandler(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	handler := NewHTTPHandler(logger)
	if handler == nil {
		t.Fatal("NewHTTPHandler should return non-nil handler")
	}
}

func TestHTTPHandler_Execute_InvalidConfig(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	handler := NewHTTPHandler(logger)

	// Create a command with no properly configured flags (should fail)
	cmd := &cobra.Command{}

	err := handler.Execute(cmd, []string{})
	if err == nil {
		t.Error("Execute should error with invalid config")
	}
}

func TestHTTPHandler_LoggerInjection(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Str("test", "handler").Logger()
	handler := NewHTTPHandler(logger)

	if handler == nil {
		t.Fatal("Handler should be created successfully")
	}
}

// Benchmark the handler creation
func BenchmarkNewHTTPHandler(b *testing.B) {
	logger := zerolog.New(os.Stderr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewHTTPHandler(logger)
	}
}