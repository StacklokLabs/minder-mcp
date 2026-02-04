package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/stacklok/minder-mcp/internal/config"
)

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// getResultText extracts the text content from a tool result.
func getResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		return ""
	}
	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("content is not TextContent, got %T", result.Content[0])
	}
	return textContent.Text
}

// newTestTools creates a Tools instance with a mock client factory for testing.
func newTestTools(mockClient *mockMinderClient) *Tools {
	return NewWithClientFactory(&config.Config{}, func(_ context.Context) (MinderClient, error) {
		return mockClient, nil
	})
}
