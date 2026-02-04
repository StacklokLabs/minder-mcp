package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestSetup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		level     string
		wantLevel slog.Level
	}{
		{
			name:      "debug level",
			level:     "debug",
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "info level",
			level:     "info",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "warn level",
			level:     "warn",
			wantLevel: slog.LevelWarn,
		},
		{
			name:      "error level",
			level:     "error",
			wantLevel: slog.LevelError,
		},
		{
			name:      "uppercase DEBUG",
			level:     "DEBUG",
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "mixed case Info",
			level:     "Info",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "invalid level defaults to info",
			level:     "invalid",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "empty string defaults to info",
			level:     "",
			wantLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := Setup(tt.level)
			if logger == nil {
				t.Fatal("Setup returned nil logger")
			}
			// Verify the logger is enabled at the expected level
			if !logger.Enabled(context.Background(), tt.wantLevel) {
				t.Errorf("Logger not enabled at %v level", tt.wantLevel)
			}
		})
	}
}
