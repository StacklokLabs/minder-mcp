package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listArtifacts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	projectID := req.GetString("project_id", "")
	provider := req.GetString("provider", "")

	reqProto := &minderv1.ListArtifactsRequest{}
	if projectID != "" || provider != "" {
		reqProto.Context = &minderv1.Context{}
		if projectID != "" {
			reqProto.Context.Project = &projectID
		}
		if provider != "" {
			reqProto.Context.Provider = &provider
		}
	}

	resp, err := client.Artifacts().ListArtifacts(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	data, err := json.MarshalIndent(resp.Results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (t *Tools) getArtifactByID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	artifactID := req.GetString("artifact_id", "")
	if artifactID == "" {
		return mcp.NewToolResultError("artifact_id is required"), nil
	}

	resp, err := client.Artifacts().GetArtifactById(ctx, &minderv1.GetArtifactByIdRequest{
		Id: artifactID,
	})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	data, err := json.MarshalIndent(resp.Artifact, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (t *Tools) getArtifactByName(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer client.Close()

	name := req.GetString("name", "")
	provider := req.GetString("provider", "")

	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	reqProto := &minderv1.GetArtifactByNameRequest{
		Name: name,
	}
	if provider != "" {
		reqProto.Context = &minderv1.Context{
			Provider: &provider,
		}
	}

	resp, err := client.Artifacts().GetArtifactByName(ctx, reqProto)
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	data, err := json.MarshalIndent(resp.Artifact, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
