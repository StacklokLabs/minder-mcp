package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listDataSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

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

	data, err := json.MarshalIndent(resp.DataSources, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (t *Tools) getDataSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataSourceID := req.GetString("data_source_id", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")

	// Validate parameters
	if errMsg := ValidateLookupParams(dataSourceID, name, "data_source_id", "name"); errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

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

	data, err := json.MarshalIndent(dataSource, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
