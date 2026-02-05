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
	"github.com/rs/cors"

	"github.com/stacklok/minder-mcp/internal/config"
	"github.com/stacklok/minder-mcp/internal/logging"
	"github.com/stacklok/minder-mcp/internal/middleware"
	"github.com/stacklok/minder-mcp/internal/resources"
	"github.com/stacklok/minder-mcp/internal/tools"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Warn if insecure mode is enabled
	if cfg.Minder.Insecure {
		fmt.Fprintln(os.Stderr, "WARNING: Running in insecure mode - TLS is disabled")
	}

	// Setup logging
	logger := logging.Setup(cfg.LogLevel)
	slog.SetDefault(logger)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"minder-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false), // Enable resource listing
	)

	// Register tools
	t := tools.New(cfg, logger)
	defer t.Close() // Ensure cleanup of HTTP client resources
	t.Register(mcpServer)

	// Register resources (including compliance dashboard)
	resources.New(logger).Register(mcpServer)

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
	mcpHandler := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath(cfg.MCP.EndpointPath),
		server.WithHeartbeatInterval(30*time.Second),
		server.WithHTTPContextFunc(authContextFunc),
		server.WithStateLess(true), // Enable stateless mode for MCP Apps compatibility
	)

	// Wrap with CORS middleware for MCP Apps support
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(mcpHandler)

	addr := fmt.Sprintf(":%d", cfg.MCP.Port)
	slog.Info("Starting Minder MCP server", "addr", addr, "endpoint", cfg.MCP.EndpointPath)

	if err := http.ListenAndServe(addr, corsHandler); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
