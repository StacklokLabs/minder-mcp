package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listRepositories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")
	provider := req.GetString("provider", "")
	cursor := req.GetString("cursor", "")
	limit := req.GetInt("limit", 0)

	// Single project mode - preserves pagination
	if projectID != "" {
		reqProto := &minderv1.ListRepositoriesRequest{
			Context: &minderv1.Context{
				Project: &projectID,
			},
		}
		if provider != "" {
			reqProto.Context.Provider = &provider
		}
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

		result := map[string]any{
			"results": resp.Results,
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
	repos, err := forEachProject(
		ctx, client, projectID,
		func(ctx context.Context, projID string) ([]*minderv1.Repository, error) {
			reqProto := &minderv1.ListRepositoriesRequest{
				Context: &minderv1.Context{
					Project: &projID,
				},
			}
			if provider != "" {
				reqProto.Context.Provider = &provider
			}
			resp, err := client.Repositories().ListRepositories(ctx, reqProto)
			if err != nil {
				return nil, err
			}
			return resp.Results, nil
		})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	result := map[string]any{
		"results":  repos,
		"has_more": false,
	}

	return marshalResult(result)
}

func (t *Tools) getRepository(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoID := req.GetString("repository_id", "")
	owner := req.GetString("owner", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")
	provider := req.GetString("provider", "")

	// Validate parameters
	if errMsg := ValidateRepositoryLookupParams(repoID, owner, name, map[string]string{
		"project_id": projectID,
		"provider":   provider,
	}); errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	var repository *minderv1.Repository

	if repoID != "" {
		// Lookup by ID - no project context needed
		resp, err := client.Repositories().GetRepositoryById(ctx, &minderv1.GetRepositoryByIdRequest{
			RepositoryId: repoID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		repository = resp.Repository
	} else {
		// Lookup by owner/name - search across projects if none specified
		fullName := owner + "/" + name
		repository, err = findInProjects(
			ctx, client, projectID,
			func(ctx context.Context, projID string) (*minderv1.Repository, error) {
				reqProto := &minderv1.GetRepositoryByNameRequest{
					Name: fullName,
					Context: &minderv1.Context{
						Project: &projID,
					},
				}
				if provider != "" {
					reqProto.Context.Provider = &provider
				}
				resp, err := client.Repositories().GetRepositoryByName(ctx, reqProto)
				if err != nil {
					return nil, err
				}
				return resp.Repository, nil
			})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
	}

	return marshalResult(repository)
}
