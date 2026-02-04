// Package logging provides structured logging setup for the MCP server.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Setup creates and returns a configured slog.Logger based on the log level string.
// Valid levels are: debug, info, warn, error.
// If an invalid level is provided, it defaults to info.
func Setup(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler)
}
