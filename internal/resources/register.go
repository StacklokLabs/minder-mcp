// Package resources implements MCP resource handlers for the Minder MCP server.
package resources

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// DashboardURI is the resource URI for the compliance dashboard.
	DashboardURI = "ui://minder/compliance-dashboard"
	// DashboardMIMEType is the MIME type for the compliance dashboard.
	DashboardMIMEType = "text/html"
)

// Register registers all MCP resources with the server.
func Register(s *server.MCPServer) {
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
		serveDashboardHTML,
	)
}

// serveDashboardHTML serves the embedded compliance dashboard HTML.
func serveDashboardHTML(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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
