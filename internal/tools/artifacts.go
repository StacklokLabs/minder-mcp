package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	minderv1 "github.com/mindersec/minder/pkg/api/protobuf/go/minder/v1"
)

func (t *Tools) listArtifacts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	return marshalResult(resp.Results)
}

func (t *Tools) getArtifact(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artifactID := req.GetString("artifact_id", "")
	name := req.GetString("name", "")
	projectID := req.GetString("project_id", "")
	provider := req.GetString("provider", "")

	// Validate parameters
	if errMsg := ValidateLookupParams(artifactID, name, "artifact_id", "name", map[string]string{
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

	var artifact *minderv1.Artifact

	if artifactID != "" {
		// Lookup by ID
		resp, err := client.Artifacts().GetArtifactById(ctx, &minderv1.GetArtifactByIdRequest{
			Id: artifactID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		artifact = resp.Artifact
	} else {
		// Lookup by name
		reqProto := &minderv1.GetArtifactByNameRequest{
			Name: name,
		}
		if projectID != "" || provider != "" {
			reqProto.Context = &minderv1.Context{}
			if projectID != "" {
				reqProto.Context.Project = &projectID
			}
			if provider != "" {
				reqProto.Context.Provider = &provider
			}
		}
		resp, err := client.Artifacts().GetArtifactByName(ctx, reqProto)
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		artifact = resp.Artifact
	}

	return marshalResult(artifact)
}
