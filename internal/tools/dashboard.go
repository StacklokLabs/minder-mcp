package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/stacklok/minder-mcp/internal/resources"
)

// showComplianceDashboard displays the interactive Minder Compliance Dashboard.
// This tool returns metadata that triggers Goose Desktop to render the embedded
// Vue.js dashboard showing repository security posture and profile compliance.
func (*Tools) showComplianceDashboard(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Opening Minder Compliance Dashboard...",
			},
		},
	}
	result.Meta = mcp.NewMetaFromMap(map[string]any{
		"ui": map[string]any{
			"resourceUri": resources.DashboardURI,
		},
	})
	return result, nil
}
