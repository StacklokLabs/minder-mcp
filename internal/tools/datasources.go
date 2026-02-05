package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listDataSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer func() { _ = client.Close() }()

	if errResult := checkHealth(ctx, client); errResult != nil {
		return errResult, nil
	}

	projectID := req.GetString("project_id", "")

	// Use multi-project aggregation when no project_id specified
	dataSources, err := forEachProject(
		ctx, client, projectID,
		func(ctx context.Context, projID string) ([]*minderv1.DataSource, error) {
			resp, err := client.DataSources().ListDataSources(ctx, &minderv1.ListDataSourcesRequest{
				Context: &minderv1.ContextV2{
					ProjectId: projID,
				},
			})
			if err != nil {
				return nil, err
			}
			return resp.DataSources, nil
		})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(dataSources)
}

func (t *Tools) getDataSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataSourceID := req.GetString("data_source_id", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	// Validate parameters
	if errMsg := ValidateLookupParams(dataSourceID, name, "data_source_id", "name", map[string]string{
		"project_id": projectID,
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

	var dataSource *minderv1.DataSource

	if dataSourceID != "" {
		// Lookup by ID - no project context needed
		resp, err := client.DataSources().GetDataSourceById(ctx, &minderv1.GetDataSourceByIdRequest{
			Id: dataSourceID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		dataSource = resp.DataSource
	} else {
		// Lookup by name - search across projects if none specified
		dataSource, err = findInProjects(
			ctx, client, projectID,
			func(ctx context.Context, projID string) (*minderv1.DataSource, error) {
				resp, err := client.DataSources().GetDataSourceByName(ctx, &minderv1.GetDataSourceByNameRequest{
					Name: name,
					Context: &minderv1.ContextV2{
						ProjectId: projID,
					},
				})
				if err != nil {
					return nil, err
				}
				return resp.DataSource, nil
			})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
	}

	return marshalResult(dataSource)
}
