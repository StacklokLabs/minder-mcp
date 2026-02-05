// Package resources implements MCP resource handlers for the Minder MCP server.
package resources

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// DashboardURI is the resource URI for the compliance dashboard.
	DashboardURI = "ui://minder/compliance-dashboard"
	// DashboardMIMEType is the MIME type for the compliance dashboard.
	DashboardMIMEType = "text/html"
)

// Resources holds the resource handlers and configuration.
type Resources struct {
	logger *slog.Logger
}

// New creates a new Resources instance.
func New(logger *slog.Logger) *Resources {
	return &Resources{
		logger: logger,
	}
}

// Register registers all MCP resources with the server.
func (r *Resources) Register(s *server.MCPServer) {
	// Compliance Dashboard resource
	s.AddResource(
		mcp.NewResource(
			DashboardURI,
			"Compliance Dashboard",
			mcp.WithResourceDescription(
				"Interactive compliance status dashboard showing real-time "+
					"compliance across repositories with drill-down capabilities"),
			mcp.WithMIMEType(DashboardMIMEType),
		),
		r.wrapHandler(DashboardURI, r.serveDashboardHTML),
	)
}

// wrapHandler wraps a resource handler with debug logging.
func (r *Resources) wrapHandler(uri string, handler server.ResourceHandlerFunc) server.ResourceHandlerFunc {
	return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		start := time.Now()
		r.logger.DebugContext(ctx, "resource requested", "uri", uri)
		result, err := handler(ctx, req)
		hasError := err != nil
		r.logger.DebugContext(ctx, "resource served",
			"uri", uri,
			"duration", time.Since(start),
			"error", hasError,
			"content_length", len(dashboardHTML),
		)
		return result, err
	}
}

// serveDashboardHTML serves the embedded compliance dashboard HTML.
func (*Resources) serveDashboardHTML(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	if dashboardHTML == "" {
		return nil, fmt.Errorf("dashboard HTML content is empty - ensure dist/index.html was built before compiling")
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      DashboardURI,
			MIMEType: DashboardMIMEType,
			Text:     dashboardHTML,
		},
	}, nil
}
