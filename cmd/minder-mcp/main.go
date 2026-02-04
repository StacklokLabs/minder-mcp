// Package main is the entry point for the Minder MCP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/stacklok/minder-mcp/internal/config"
	"github.com/stacklok/minder-mcp/internal/logging"
	"github.com/stacklok/minder-mcp/internal/middleware"
	"github.com/stacklok/minder-mcp/internal/tools"
)

func main() {
	cfg := config.Load()

	// Setup logging
	logger := logging.Setup(cfg.LogLevel)
	slog.SetDefault(logger)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"minder-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	t := tools.New(cfg, logger)
	t.Register(mcpServer)

	// Create HTTP context function that extracts auth token
	authContextFunc := func(ctx context.Context, r *http.Request) context.Context {
		var token, source string
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
			source = "header"
		}
		if token == "" {
			token = cfg.Minder.AuthToken
			source = "config"
		}
		slog.Debug("auth context", "has_token", token != "", "source", source)
		return middleware.ContextWithToken(ctx, token)
	}

	// Create streamable HTTP server with auth context
	httpServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath(cfg.MCP.EndpointPath),
		server.WithHeartbeatInterval(30*time.Second),
		server.WithHTTPContextFunc(authContextFunc),
	)

	addr := fmt.Sprintf(":%d", cfg.MCP.Port)
	slog.Info("Starting Minder MCP server", "addr", addr, "endpoint", cfg.MCP.EndpointPath)

	if err := httpServer.Start(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
