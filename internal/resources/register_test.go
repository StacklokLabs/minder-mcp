package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(_ *testing.T) {
	mcpServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithResourceCapabilities(true, false),
	)

	// Should not panic
	Register(mcpServer)
}

func TestServeDashboardHTML(t *testing.T) {
	ctx := context.Background()
	req := mcp.ReadResourceRequest{}

	contents, err := serveDashboardHTML(ctx, req)

	require.NoError(t, err)
	require.Len(t, contents, 1)

	textContent, ok := contents[0].(mcp.TextResourceContents)
	require.True(t, ok, "expected TextResourceContents")

	assert.Equal(t, DashboardURI, textContent.URI)
	assert.Equal(t, DashboardMIMEType, textContent.MIMEType)
	assert.Contains(t, textContent.Text, "<!DOCTYPE html>")
	assert.Contains(t, textContent.Text, "Minder Compliance Dashboard")
}

func TestDashboardHTMLContent(t *testing.T) {
	// Verify the dashboard HTML contains expected elements
	assert.NotEmpty(t, dashboardHTML)
	assert.Contains(t, dashboardHTML, "<!DOCTYPE html>")
	assert.Contains(t, dashboardHTML, "<title>Minder Compliance Dashboard</title>")

	// Check for key UI elements (these are in the HTML template, not minified JS)
	assert.Contains(t, dashboardHTML, "summary-cards")
	assert.Contains(t, dashboardHTML, "profiles-list")
	assert.Contains(t, dashboardHTML, "repositories-list")

	// Check for MCP tool names (these appear as string literals in the bundled JS)
	assert.Contains(t, dashboardHTML, "minder_list_profiles")
	assert.Contains(t, dashboardHTML, "minder_get_profile_status")
	assert.Contains(t, dashboardHTML, "minder_list_repositories")
}

func TestDashboardURIFormat(t *testing.T) {
	// Verify the URI follows the expected ui:// scheme
	assert.True(t, strings.HasPrefix(DashboardURI, "ui://"), "Dashboard URI should use ui:// scheme")
}
