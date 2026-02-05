package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stacklok/minder-mcp/internal/resources"
)

func TestShowComplianceDashboard(t *testing.T) {
	tools := &Tools{}

	result, err := tools.showComplianceDashboard(context.Background(), mcp.CallToolRequest{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify content
	require.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Dashboard")

	// Verify metadata contains UI resource URI
	require.NotNil(t, result.Meta)
	require.NotNil(t, result.Meta.AdditionalFields)

	ui, ok := result.Meta.AdditionalFields["ui"].(map[string]any)
	require.True(t, ok, "expected ui map in metadata")

	resourceUri, ok := ui["resourceUri"].(string)
	require.True(t, ok, "expected resourceUri string")
	assert.Equal(t, resources.DashboardURI, resourceUri)
}
