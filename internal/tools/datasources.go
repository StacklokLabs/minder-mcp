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

	reqProto := &minderv1.ListDataSourcesRequest{}
	if projectID != "" {
		reqProto.Context = &minderv1.ContextV2{
			ProjectId: projectID,
		}
	}

	resp, err := client.DataSources().ListDataSources(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(resp.DataSources)
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
		// Lookup by ID
		resp, err := client.DataSources().GetDataSourceById(ctx, &minderv1.GetDataSourceByIdRequest{
			Id: dataSourceID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		dataSource = resp.DataSource
	} else {
		// Lookup by name
		reqProto := &minderv1.GetDataSourceByNameRequest{
			Name: name,
		}
		if projectID != "" {
			reqProto.Context = &minderv1.ContextV2{
				ProjectId: projectID,
			}
		}
		resp, err := client.DataSources().GetDataSourceByName(ctx, reqProto)
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		dataSource = resp.DataSource
	}

	return marshalResult(dataSource)
}
