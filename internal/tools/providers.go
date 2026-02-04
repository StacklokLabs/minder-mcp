package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listProviders(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	projectID := req.GetString("project_id", "")
	cursor := req.GetString("cursor", "")
	limit := req.GetInt("limit", 0)

	reqProto := &minderv1.ListProvidersRequest{}
	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}

	// Add pagination parameters
	if cursor != "" {
		reqProto.Cursor = cursor
	}
	if limit > 0 && limit <= 100 {
		reqProto.Limit = int32(limit) //nolint:gosec // limit is bounded by schema validation (1-100)
	}

	resp, err := client.Providers().ListProviders(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	// Build paginated response
	result := map[string]any{
		"results": resp.Providers,
	}
	if resp.Cursor != "" {
		result["next_cursor"] = resp.Cursor
		result["has_more"] = true
	} else {
		result["has_more"] = false
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (t *Tools) getProvider(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	reqProto := &minderv1.GetProviderRequest{
		Name: name,
	}
	if projectID != "" {
		reqProto.Context = &minderv1.Context{
			Project: &projectID,
		}
	}

	resp, err := client.Providers().GetProvider(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	data, err := json.MarshalIndent(resp.Provider, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
