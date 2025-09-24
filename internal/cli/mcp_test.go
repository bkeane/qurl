package cli

import (
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func TestNewMCPHandler(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	handler := NewMCPHandler(logger)
	if handler == nil {
		t.Fatal("NewMCPHandler should return non-nil handler")
	}
}

func TestMCPHandler_Execute_InvalidConfig(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	handler := NewMCPHandler(logger)

	// Create a command with no context or properly configured flags (should fail)
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background()) // Provide a context to avoid panic

	err := handler.Execute(cmd, []string{})
	if err == nil {
		t.Error("Execute should error with invalid config")
	}
}

func TestMCPHandler_ContextHandling(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	handler := NewMCPHandler(logger)

	// Test with context
	ctx := context.Background()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	// Test that context is properly handled (shouldn't panic)
	err := handler.Execute(cmd, []string{})

	// We expect an error due to missing configuration, but no panic
	if err == nil {
		t.Error("Execute should error with incomplete config")
	}
}

func TestMCPHandler_LoggerInjection(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Str("test", "mcp_handler").Logger()
	handler := NewMCPHandler(logger)

	if handler == nil {
		t.Fatal("Handler should be created successfully")
	}
}

// Benchmark the MCP handler creation
func BenchmarkNewMCPHandler(b *testing.B) {
	logger := zerolog.New(os.Stderr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMCPHandler(logger)
	}
}