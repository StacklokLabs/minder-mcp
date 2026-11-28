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

	projectID := req.GetString(paramProjectID, "")
	provider := req.GetString(paramProvider, "")

	// Use multi-project aggregation when no project_id specified
	artifacts, err := forEachProject(ctx, client, projectID, func(ctx context.Context, projID string) ([]*minderv1.Artifact, error) {
		reqProto := &minderv1.ListArtifactsRequest{
			Context: &minderv1.Context{
				Project: &projID,
			},
		}
		if provider != "" {
			reqProto.Context.Provider = &provider
		}
		resp, err := client.Artifacts().ListArtifacts(ctx, reqProto)
		if err != nil {
			return nil, err
		}
		return resp.Results, nil
	})
	if err != nil {
		return mcp.NewToolResultError(MapGRPCError(err)), nil
	}

	return marshalResult(artifacts)
}

func (t *Tools) getArtifact(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artifactID := req.GetString("artifact_id", "")
	name := req.GetString(paramName, "")
	projectID := req.GetString(paramProjectID, "")
	provider := req.GetString(paramProvider, "")

	// Validate parameters
	if errMsg := ValidateLookupParams(artifactID, name, "artifact_id", paramName, map[string]string{
		paramProjectID: projectID,
		paramProvider:  provider,
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
		// Lookup by ID - no project context needed
		resp, err := client.Artifacts().GetArtifactById(ctx, &minderv1.GetArtifactByIdRequest{
			Id: artifactID,
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
		artifact = resp.Artifact
	} else {
		// Lookup by name - search across projects if none specified
		artifact, err = findInProjects(ctx, client, projectID, func(ctx context.Context, projID string) (*minderv1.Artifact, error) {
			reqProto := &minderv1.GetArtifactByNameRequest{
				Name: name,
				Context: &minderv1.Context{
					Project: &projID,
				},
			}
			if provider != "" {
				reqProto.Context.Provider = &provider
			}
			resp, err := client.Artifacts().GetArtifactByName(ctx, reqProto)
			if err != nil {
				return nil, err
			}
			return resp.Artifact, nil
		})
		if err != nil {
			return mcp.NewToolResultError(MapGRPCError(err)), nil
		}
	}

	return marshalResult(artifact)
}
