package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listProviders(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")
	cursor := req.GetString("cursor", "")
	limit := req.GetInt("limit", 0)

	// Single project mode - preserves pagination
	if projectID != "" {
		reqProto := &minderv1.ListProvidersRequest{
			Context: &minderv1.Context{
				Project: &projectID,
			},
		}
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

		result := map[string]any{
			"results": resp.Providers,
		}
		if resp.Cursor != "" {
			result["next_cursor"] = resp.Cursor
			result["has_more"] = true
		} else {
			result["has_more"] = false
		}
		return marshalResult(result)
	}

	// Multi-project aggregation - pagination not supported
	providers, err := forEachProject(
		ctx, client, projectID,
		func(ctx context.Context, projID string) ([]*minderv1.Provider, error) {
			reqProto := &minderv1.ListProvidersRequest{
				Context: &minderv1.Context{
					Project: &projID,
				},
			}
			resp, err := client.Providers().ListProviders(ctx, reqProto)
			if err != nil {
				return nil, err
			}
			return resp.Providers, nil
		})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	result := map[string]any{
		"results":  providers,
		"has_more": false,
	}

	return marshalResult(result)
}

func (t *Tools) getProvider(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	// Search across projects if none specified
	provider, err := findInProjects(ctx, client, projectID, func(ctx context.Context, projID string) (*minderv1.Provider, error) {
		resp, err := client.Providers().GetProvider(ctx, &minderv1.GetProviderRequest{
			Name: name,
			Context: &minderv1.Context{
				Project: &projID,
			},
		})
		if err != nil {
			return nil, err
		}
		return resp.Provider, nil
	})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(provider)
}
