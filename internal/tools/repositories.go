package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listRepositories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	projectID := req.GetString("project_id", "")
	provider := req.GetString("provider", "")
	cursor := req.GetString("cursor", "")
	limit := req.GetInt("limit", 0)

	reqProto := &minderv1.ListRepositoriesRequest{}
	if projectID != "" || provider != "" {
		reqProto.Context = &minderv1.Context{}
		if projectID != "" {
			reqProto.Context.Project = &projectID
		}
		if provider != "" {
			reqProto.Context.Provider = &provider
		}
	}

	// Add pagination parameters
	if cursor != "" {
		reqProto.Cursor = cursor
	}
	if limit > 0 && limit <= 100 {
		reqProto.Limit = int64(limit) //nolint:gosec // limit is bounded by schema validation (1-100)
	}

	resp, err := client.Repositories().ListRepositories(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	// Build paginated response
	result := map[string]any{
		"results": resp.Results,
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

func (t *Tools) getRepository(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoID := req.GetString("repository_id", "")
	owner := req.GetString("owner", "")
	name := req.GetString("name", "")
	provider := req.GetString("provider", "")

	// Validate parameters
	if errMsg := ValidateRepositoryLookupParams(repoID, owner, name); errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	var repository *minderv1.Repository

	if repoID != "" {
		// Lookup by ID
		resp, err := client.Repositories().GetRepositoryById(ctx, &minderv1.GetRepositoryByIdRequest{
			RepositoryId: repoID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		repository = resp.Repository
	} else {
		// Lookup by owner/name
		fullName := owner + "/" + name
		reqProto := &minderv1.GetRepositoryByNameRequest{
			Name: fullName,
		}
		if provider != "" {
			reqProto.Context = &minderv1.Context{
				Provider: &provider,
			}
		}
		resp, err := client.Repositories().GetRepositoryByName(ctx, reqProto)
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		repository = resp.Repository
	}

	data, err := json.MarshalIndent(repository, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
