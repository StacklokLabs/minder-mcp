// Package main is the entry point for the Minder MCP server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/stacklok/minder-mcp/internal/config"
	"github.com/stacklok/minder-mcp/internal/middleware"
	"github.com/stacklok/minder-mcp/internal/tools"
)

func main() {
	cfg := config.Load()

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"minder-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	t := tools.New(cfg)
	t.Register(mcpServer)

	// Create HTTP context function that extracts auth token
	authContextFunc := func(ctx context.Context, r *http.Request) context.Context {
		token := ""
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			token = cfg.Minder.AuthToken
		}
		return context.WithValue(ctx, middleware.AuthTokenKey, token)
	}

	// Create streamable HTTP server with auth context
	httpServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath(cfg.MCP.EndpointPath),
		server.WithHeartbeatInterval(30*time.Second),
		server.WithHTTPContextFunc(authContextFunc),
	)

	addr := fmt.Sprintf(":%d", cfg.MCP.Port)
	log.Printf("Starting Minder MCP server on %s%s", addr, cfg.MCP.EndpointPath)

	if err := httpServer.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
