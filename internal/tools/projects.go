package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")

	var projects any
	if projectID != "" {
		// List child projects of the specified parent
		resp, err := client.Projects().ListChildProjects(ctx, &minderv1.ListChildProjectsRequest{
			Context: &minderv1.ContextV2{
				ProjectId: projectID,
			},
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		projects = resp.Projects
	} else {
		// List all accessible projects
		resp, err := client.Projects().ListProjects(ctx, &minderv1.ListProjectsRequest{})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		projects = resp.Projects
	}

	return marshalResult(projects)
}
